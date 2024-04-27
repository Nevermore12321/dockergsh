package daemongsh

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/graphdriver"
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/internal/graph"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/Nevermore12321/dockergsh/pkg/parse/kernel"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"runtime"
)

// Daemongsh 在 Dockergsh架构中 Daemongsh 支撑着整个后台进程的运行，
// 同时也统一化管理着 Docker 架构中 graph、graphdriver、execdriver、volumes、Docker 容器等众多资源。
// 可以说，Dockergsh Daemongsh复杂的运作均由daemongsh对象来调度
// todo add elements
type Daemongsh struct {
}

// NewDaemongsh 创建 Daemongsh 对象实例
func NewDaemongsh(config *Config, eng *engine.Engine) (*Daemongsh, error) {
	daemongsh, err := NewDaemongshFromDirectory(config, eng)
	if err != nil {
		return nil, err
	}
	return daemongsh, nil
}

// NewDaemongshFromDirectory 具体通过 Config 配置和 engine 对象创建 Daemongsh 对象实例
func NewDaemongshFromDirectory(config *Config, eng *engine.Engine) (*Daemongsh, error) {
	// 1. 应用配置信息
	// 1.1 配置 Dockergsh 容器的 MTU。容器网络接口的最大传输单元（MTU）
	if config.Mtu == 0 { // 表示没有设置
		// 设置一个默认值
		config.Mtu = GetDefaultNetworkMtu()
	}
	// 1.2 检测网桥配置信息，即 BridgeIface 和 BridgeIP
	// BridgeIface 和 BridgeIP 的作用是为创建网桥的任务 init_networkdriver 提供参数
	if config.BridgeIp != "" && config.BridgeIface != "" {
		// 若config中BridgeIface和BridgeIP两个属性均不为空，则返回nil对象，并返回错误信息
		/*
			原因：
				1. 当用户为Docker网桥选定已经存在的网桥接口时，应该沿用已有网桥的当前IP地址，不应再提供IP地址；
				2. 当用户不选已经存在的网桥接口作为Docker网桥时，Docker会另行创建一个全新的网桥接口作为Docker网桥，此时用户可以为此新创建的网桥接口设定自定义IP地址；
				3. 当然两者都不选的话，Docker会为用户接管完整的Docker网桥创建流程，从创建默认的网桥接口，到尾网桥接口设置默认的IP地址
		*/
		return nil, fmt.Errorf("you specified -b & --bip, mutually exclusive options. Please specify only one")
	}

	// 1.3 查验容器间的通信配置，即 EnableIptables 和 Inter-Container Communication 属性
	if !config.EnableIptables && !config.InterContainerCommunication {
		/*
			EnableIptables属性主要来源于flag参数--iptables
				它的作用是：在DockerDaemon启动时，是否对宿主机的iptables规则作修改
			InterContainerCommunication属性来源于flag参数--icc
				它的作用是：在Docker Daemon启动时，是否开启Docker容器之间互相通信的功能

			容器间通信依赖与 iptables：
				1. 若 InterContainerCommunication 的值为false，则Docker Daemon会在宿主机iptables的FORWARD链中添加一条Docker容器间流量均DROP的规则；
				2. 若 InterContainerCommunication 为true，则Docker Daemon会在宿主机iptables的FORWARD链中添加一条Docker容器间流量均ACCEPT的规则。
		*/
		return nil, fmt.Errorf("you specified --iptables=false with --icc=false. ICC uses iptables to function. Please set --icc or --iptables to true")
	}

	// 1.4 网络相关配置，即 DisableNetwork 属性
	/*
		后续创建并执行创建Docker Daemon网络环境时会使用此属性，即在名为init_networkdriver的Job创建并运行中体现

	*/
	config.DisableNetwork = DisableNetworkBridge == config.BridgeIface

	// 1.5 处理 PID 文件配置
	if config.PidFile != "" {
		/*
			1. 为运行时 Daemongsh 进程的 PID 号创建一个 PID 文件，文件的路径即为 config中的P idfile 属性，
			2. 并且为 Daemongsh 的shutdown操作添加一个删除此Pidfile的函数，以便在 Daemongsh 退出的时候，可以在第一时间删除 Pidfile
		*/
		if err := utils.CreatePidFile(config.PidFile); err != nil {
			return nil, err
		}
		eng.OnShutdown(func() {
			// 始终在结束时，删除 pid 文件
			utils.RemovePidFile(config.PidFile)
		})
	}

	/* -------------- */

	// 2. 检测系统支持及用户权限
	// 2.1 操作系统类型对 Daemongsh 的支持；runtime.GOOS返回运行程序所在操作系统的类型
	if runtime.GOOS != "linux" {
		// 目前只能支持 linux 系统
		log.Fatalf("The Dockergsh daemon is only supported on linux")
	}
	// 2.2 用户权限的级别；os.Geteuid() 返回调用者所在组的组id
	if os.Geteuid() != 0 {
		// 需要使用 root 用户执行
		log.Fatalf("The Dockergsh daemon needs to be run as root")
	}
	// 2.3 检测内核的版本以及主机处理器类型
	if err := checkKernelAndArch(); err != nil {
		log.Fatalf(err.Error())
	}

	/* -------------- */
	// 3.配置工作目录，包括：所有的Docker镜像内容、所有Docker容器的文件系统、所有Docker容器的元数据、所有容器的数据卷内容等
	// 3.1 使用规范路径创建一个TempDir，路径名为 tmp
	tmp, err := utils.TempDir(config.Root)
	if err != nil {
		log.Fatalf("Unable to get the TempDir under %s: %s", config.Root, err)
	}
	// 3.2 通过 tmp（config.root），找到一个指向 tmp 的实际的绝对目录 realTmp
	realTmp, err := utils.ReadSymlinkedDirectory(tmp)
	if err != nil {
		log.Fatalf("Unable to get the full path to the TempDir (%s): %s", tmp, err)
	}
	// 3.3 使用realTemp的值，创建并赋值给环境变量TMPDIR
	_ = os.Setenv(utils.DaemongshTempdir, realTmp)
	// 3.4 处理config的属性EnableSelinuxSupport,如果不开启 selinux，将其关闭
	if !config.EnableSelinuxSupport {
		selinuxSetDisabled()
	}
	// 3.5 将realRoot重新赋值于config.Root，并创建Docker Daemon的工作根目录。
	var realRoot string
	if _, err := os.Stat(config.Root); err != nil && os.IsNotExist(err) { // config.Root 路径不存在
		realRoot = config.Root
	} else { // config.Root 路径存在
		realRoot, err = utils.ReadSymlinkedDirectory(config.Root)
		if err != nil {
			log.Fatalf("Unable to get the full path to root (%s): %s", config.Root, err)
		}
	}
	config.Root = realRoot

	if err := os.MkdirAll(config.Root, 0700); err != nil && !os.IsExist(err) { // 如果 root 不存在创建目录
		return nil, err
	}

	/* -------------- */
	// 4. 加载并配置 graphdriver，镜像管理所需的驱动环境
	// 4.1 配置默认的 graphdriver
	graphdriver.DefaultDriver = config.GraphDriver

	// 4.2 创建 graphdriver
	driver, err := graphdriver.New(config.Root, config.GraphOptions)
	if err != nil {
		return nil, err
	}
	log.Debugf("Using Graph Driver %s", driver)

	// 4.3 由于在btrfs文件系统上运行的Docker不兼容SELinux，因此启用SELinux的支持，并且驱动的类型为btrfs不能同时设置
	if config.EnableSelinuxSupport && driver.String() == "btrfs" {
		return nil, fmt.Errorf("SELinux is not supported with the BTRFS graph driver")
	}

	// 4.4 创建容器仓库目录，存储容器的元数据信息，/var/lib/dockergsh/containers
	daemonContainerRepo := path.Join(config.Root, "containers")
	if err := os.MkdirAll(daemonContainerRepo, 0700); err != nil && !os.IsExist(err) {
		return nil, err
	}

	// 4.5 创建镜像graph，通过Docker的root目录以及graphdriver实例，实例化一个全新的graph对象，用以管理在文件系统中Docker的root路径下graph目录的内容
	log.Debugf("Creating images graph")
	g, err := graph.NewGraph(path.Join(config.Root, "graph"), driver)
	fmt.Println(g)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Install 向 engine 中注册所有可执行的任务
func (daemon *Daemongsh) Install(eng *engine.Engine) error {
	// 所有任务
	jobs := map[string]engine.Handler{
		//"attach":            daemon.ContainerAttach,
		//"build":             daemon.CmdBuild,
		//"commit":            daemon.ContainerCommit,
		//"container_changes": daemon.ContainerChanges,
		//"container_copy":    daemon.ContainerCopy,
		//"container_inspect": daemon.ContainerInspect,
		//"containers":        daemon.Containers,
		//"create":            daemon.ContainerCreate,
		//"delete":            daemon.ContainerDestroy,
		//"export":            daemon.ContainerExport,
		//"info":              daemon.CmdInfo,
		//"kill":              daemon.ContainerKill,
		"logs": daemon.ContainerLogs,
		//"pause":             daemon.ContainerPause,
		//"resize":            daemon.ContainerResize,
		//"restart":           daemon.ContainerRestart,
		//"start":             daemon.ContainerStart,
		//"stop":              daemon.ContainerStop,
		//"top":               daemon.ContainerTop,
		//"unpause":           daemon.ContainerUnpause,
		//"wait":              daemon.ContainerWait,
		//"image_delete":      daemon.ImageDelete,
	}

	// 将所有 jobs 注册到 engine 中
	for name, method := range jobs {
		err := eng.Register(name, method)
		if err != nil {
			return err
		}
	}
	// todo
	return nil
}

// 检测内核的版本以及主机处理器类型
func checkKernelAndArch() error {
	// 检测内核的版本以及主机处理器类型, 目前支持 amd64
	if runtime.GOARCH != "amd64" {
		return fmt.Errorf("the Dockergsh runtime currently only supports amd64 (not %s). This will change in the future. Aborting", runtime.GOARCH)
	}

	// 检测Linux内核版本是否满足要求，建议用户升级内核版本至3.8.0或以上版本
	// 先获取当前系统的版本信息
	if kv, err := kernel.GetKernelVersion(); err != nil {
		log.Warning(err)
	} else {
		// 如果版本小于 3.8.0
		if kernel.CompareKernelVersion(kv, &kernel.KernelVersionInfo{Kernel: 3, Major: 8, Minor: 0}) < 0 {
			// 如果不可以容忍低版本运行 daemon
			if os.Getenv(utils.NowarnKernelVersion) == "" {
				log.Warnf("You are running linux kernel version %s, which might be unstable running docker. Please upgrade your kernel to 3.8.0.", kv.String())
			}
		}
	}
	return nil
}
