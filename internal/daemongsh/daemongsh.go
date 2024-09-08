package daemongsh

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/execdriver"
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/execdriver/execdrivers"
	"github.com/Nevermore12321/dockergsh/internal/daemongsh/graphdriver"
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/internal/graph"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/Nevermore12321/dockergsh/pkg/graphdb"
	"github.com/Nevermore12321/dockergsh/pkg/networkfs/resolvconf"
	"github.com/Nevermore12321/dockergsh/pkg/parse/kernel"
	"github.com/Nevermore12321/dockergsh/pkg/sysinfo"
	"github.com/Nevermore12321/dockergsh/pkg/truncindex"
	utilsPackage "github.com/Nevermore12321/dockergsh/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	DefaultDns = []string{"8.8.8.8", "8.8.4.4"}
)

// 全局所有容器信息
type contStore struct {
	s          map[string]*Container // 所有容器全局信息，id-container 映射关系
	sync.Mutex                       // 全局变量修改锁
}

func (c *contStore) Add(id string, cont *Container) {
	c.Lock()
	defer c.Unlock()
	c.s[id] = cont
}

func (c *contStore) Get(id string) *Container {
	c.Lock()
	defer c.Unlock()
	res := c.s[id]
	return res
}

func (c *contStore) Delete(id string) {
	c.Lock()
	defer c.Unlock()
	delete(c.s, id)
}

func (c *contStore) List() []*Container {
	containers := new(History)
	c.Lock()
	for _, container := range c.s { // 将所有容器放入 History 对象中
		containers.Add(container)
	}
	c.Unlock()
	containers.Sort() // 按照时间排序
	return *containers
}

// Daemongsh 在 Dockergsh架构中 Daemongsh 支撑着整个后台进程的运行，
// 同时也统一化管理着 Docker 架构中 graph、graphdriver、execdriver、volumes、Docker 容器等众多资源。
// 可以说，Dockergsh Daemongsh复杂的运作均由daemongsh对象来调度
type Daemongsh struct {
	repository     string                 // 存储所有docker容器信息的路径，默认为：/var/lib/dockergsh/containers
	sysInitPath    string                 // 系统dockerinit二进制文件所在的路径
	containers     *contStore             // 用户存储docker容器信息的对象
	graph          *graph.Graph           // 存储docker镜像的graph对象
	repositories   *graph.TagStore        // 存储本级所有docker镜像repo信息的对象
	idIndex        *truncindex.TruncIndex // 镜像的短ID，通过简短有效的字符串前缀定位唯一的镜像
	sysInfo        *sysinfo.SysInfo       // 系统功能信息
	volumes        *graph.Graph           // 管理宿主机上 volume 内容的 graphdriver，默认为 vfs
	eng            *engine.Engine         // docker 执行引擎 Engine 类型
	config         *Config                // config.go 文件中的配置信息，以及执行后产生的配置 DisableNetwork
	containerGraph *graphdb.Database      // 存储docker镜像关系的 graphdb
	driver         graphdriver.Driver     // 管理docker镜像的驱动 graphdriver，默认为 aufs
	execDriver     execdriver.Driver      // docker daemon 的 exec 驱动，默认为 native
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

	// 4.2 创建 graphdriver，用来管理 graph
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
	// graph driver 与 graph 是一对的
	log.Debugf("Creating images graph")
	g, err := graph.NewGraph(path.Join(config.Root, "graph"), driver)
	if err != nil {
		return nil, err
	}

	// 4.6 创建 volumesdriver 以及 volumes graph
	// 		- 使用vfs这种类型的driver创建volumesDriver；
	//		- 在Docker的root路径下创建volumes目录，并返回volumes这个graph对象实例
	// 之前 4.5 创建的graphdriver 用来管理容器的文件系统，Docker需要使用volumedriver来管理数据卷。
	// 因为数据卷的管理不会像容器文件系统管理那么复杂，故Docker采用vfs驱动实现数据卷的管理
	volumeDriver, err := graphdriver.GetDriver("vfs", config.Root, config.GraphOptions)
	if err != nil {
		return nil, err
	}
	// volume driver 与 volume graph 是一对
	log.Debugf("Creating volumes graph")
	volumeGraph, err := graph.NewGraph(path.Join(config.Root, "volumes"), volumeDriver)
	if err != nil {
		return nil, err
	}

	// 4.7 创建 tagStore
	// TagStore 主要是用于管理存储镜像的仓库列表（repository list）
	// 这里总结一下 graph、graphdriver、tagstore 的关系：
	// 	- Graph：每个 Docker 镜像由多个层（layer）组成，这些层叠加在一起形成最终的镜像。Graph 组件管理这些层次结构以及它们之间的关系。
	// 	- graphdriver：不同的存储驱动（如 aufs、btrfs、overlay、devicemapper 等）实现了不同的 GraphDriver 接口，以适应各种存储后端。
	// 	- tagstore：管理镜像标签（tags）的组件。它维护了镜像名称与具体镜像 ID 之间的映射关系
	// TagStore 管理镜像tag ，graph 通过管理镜像的 id，graphdriver 调用底层驱动创建 layer
	log.Debugf("Creating repository list")
	repositories, err := graph.NewTagStore(path.Join(config.Root, "repositories-"+driver.String()), g)
	if err != nil {
		return nil, fmt.Errorf("couldn't create Tag store: %s", err)
	}

	/* -------------- */
	// 5. 配置Docker Daemon网络环境
	// 	通过运行 init_networkdriver 的Job来完成
	if !config.DisableNetwork {
		netJob := eng.Job("init_networkdriver")

		// 设置环境变量
		netJob.SetEnvBool("EnableIptables", config.EnableIptables)
		netJob.SetEnvBool("InterContainerCommunication", config.InterContainerCommunication)
		netJob.SetEnvBool("EnableIpForward", config.EnableIpForward)
		netJob.SetEnv("BridgeIface", config.BridgeIface)
		netJob.SetEnv("BridgeIp", config.BridgeIp)
		netJob.SetEnv("DefaultBindingIp", config.DefaultIp.String())

		// 运行 init_networkdriver job，其实就是用来创建 docker0 网桥
		if err := netJob.Run(); err != nil {
			return nil, err
		}
	}

	/* -------------- */
	// 6. 创建graphdb并初始化
	// graphdb是一个构建在SQLite之上的图形数据库，通常用来记录节点命名以及节点之间的关联
	// Daemon正是使用graphdb来记录容器间的关联信息(docker link)
	// graphdb 的目录为 /var/lib/dockergsh/linkgraph.db
	graphdbPath := path.Join(config.Root, "linkgraph.db")
	graphdbConn, err := graphdb.NewSqliteConn(graphdbPath)
	if err != nil {
		return nil, err
	}

	// 7. 寻找 dockergshinit 的二进制文件
	// 当我们找到合适的 dockergshinit 二进制文件（即使它是本地二进制文件）时，将其复制到 localCopy 的 config.Root 中以供将来使用（这样原始文件就可以消失而不会出现问题，例如在软件包升级期间）
	// 路径为/var/lib/dockergsh/init/dockergshinit-[VERSION]
	localPath := path.Join(config.Root, "init", fmt.Sprint("dockergsh-%s", utils.VERSION))
	sysInitPath := utilsPackage.DockergshInitPath(localPath)
	if sysInitPath == "" {
		return nil, fmt.Errorf("could not locate dockergshinit: This usually means docker was built incorrectly")
	}
	if sysInitPath != localPath {
		if err := os.Mkdir(path.Dir(localPath), 0700); err != nil && os.IsExist(err) {
			return nil, err
		}
		// 将 dockergshinit 文件拷贝 到 localPath 下
		if _, err := utilsPackage.CopyFile(sysInitPath, localPath); err != nil {
			return nil, err
		}
		if err := os.Chmod(localPath, 0700); err != nil {
			return nil, err
		}
		sysInitPath = localPath
	}

	// 8. 创建 execdriver
	// execdriver 是 Dockergsh 中用来执行容器任务的驱动
	sysInfo := sysinfo.New(false)
	ed, err := execdrivers.NewDriver(config.ExecDriver, config.Root, sysInitPath, sysInfo)
	if err != nil {
		return nil, err
	}

	// 9. 创建 daemongsh 实例
	// 对象实例daemon涉及的内容极多，比如：
	// - 对于Docker镜像的存储可以通过graph来管理
	// - 所有Docker容器的元数据信息都保存在containers对象中
	// - 整个Docker Daemon的任务执行位于eng属性中
	// - 等等
	daemongsh := &Daemongsh{
		repository:     daemonContainerRepo,
		sysInitPath:    sysInitPath,
		containers:     &contStore{s: make(map[string]*Container)},
		graph:          g,
		repositories:   repositories,
		idIndex:        truncindex.NewTruncIndex([]string{}),
		sysInfo:        sysInfo,
		volumes:        volumeGraph,
		eng:            eng,
		config:         config,
		containerGraph: graphdbConn,
		driver:         driver,
		execDriver:     ed,
	}

	// 10. 检测DNS配置
	// 检测Docker运行环境中DNS的配置
	if err := daemongsh.checkLocalDNS(); err != nil {
		return nil, err
	}

	// 11. 启动时，加载已有的容器
	// 在Docker Daemon重启时，有可能有遗留的容器，这部分容器存储在 daemon.repository 中（路径在 /var/lib/dockergsh/containers）
	// todo complete restore
	if err := daemongsh.restore(); err != nil {
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

// checkLocalDNS 检查本地 DNS 配置 /etc/resolv.conf 中是否有本地 127.0.0.1 的 dns
func (daemon *Daemongsh) checkLocalDNS() error {
	resolveConf, err := resolvconf.Get()
	if err != nil {
		return err
	}

	// 若宿主机上DNS文件resolv.conf中有127.0.0.1，而Docker容器在自身内部不能使用该地址，只能使用默认 dns 地址
	// 默认 dns 为 8.8.8.8，8.8.4.4
	if len(resolveConf) == 0 && utilsPackage.CheckLocalDNS(resolveConf) {
		log.Infof("Local (127.0.0.1) DNS resolver found in resolv.conf and containers can't use it. Using default external servers : %v", DefaultDns)
		daemon.config.Dns = DefaultDns
	}
	return nil
}

// restore docker daemon 容器时，加载之前已经运行的容器
func (daemon *Daemongsh) restore() error {
	var (
		DEBUG         = os.Getenv(utils.DockergshDebug) != "" || os.Getenv("TEST") != ""
		containers    = make(map[string]*Container)
		currentDriver = daemon.driver.String() // 当前 docker daemon 使用的 graph driver
	)

	if !DEBUG {
		log.Infof("Loading containers: ")
	}

	// 读取 daemongsh.repository 目录，即 /var/lib/dockergsh/containers
	//
	dir, err := os.ReadDir(daemon.repository)
	if err != nil {
		return err
	}

	// 遍历目录下所有文件夹，文件夹名就是容器的ID
	for _, file := range dir {
		containerID := file.Name()
		container, err := daemon.load(containerID) // 加载容器文件夹下的容器信息
		if !DEBUG {
			fmt.Print(".")
		}
		if err != nil {
			log.Errorf("Failed to load container %v: %v", containerID, err)
			continue
		}

		// 如果容器的 graph driver 不支持图形当前使用的 graph 驱动程序，则忽略该容器
		if (container.Driver == "" && currentDriver == "aufs") || container.Driver == currentDriver {
			log.Debugf("Loaded container %v", container.ID)
			containers[container.ID] = container
		} else { // 不支持
			log.Debugf("Cannot load container %s because it was created with another graph driver.", container.ID)
		}
	}

	// todo
}

// 拼出当前id容器的根目录，即 /var/lib/dockergsh/containers/[CONTAINER_ID]
func (daemon *Daemongsh) containerRoot(containerID string) string {
	return filepath.Join(daemon.repository, containerID)
}

// load 根据容器 ID 加载容器信息
func (daemon *Daemongsh) load(containerID string) (*Container, error) {
	// 初始化一个空的容器
	container := &Container{
		Root:  daemon.containerRoot(containerID),
		State: NewState(),
	}

	// 从配置文件Container.Root中加载容器的相关设置
	if err := container.LoadFromDisk(); err != nil {
		return nil, err
	}

	if container.ID != containerID {
		return container, fmt.Errorf("container %s is stored at %s", container.ID, containerID)
	}

	return container, nil
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
