package network

import (
	"encoding/json"
	"fmt"
	"github.com/Nevermore12321/dockergsh/container"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"io/fs"
	"net"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"text/tabwriter"
)

var (
	networkDefaultPath = "/var/lib/dockergsh/network/network"
	// drivers 字典，是各个网络驱动的实例字典
	drivers = map[string]NetworkDriver{}
	// networks 字典，是所有网络的实例字典
	networks = map[string]*Network{}
)

// Network 网络驱动 driver
type Network struct {
	Name    string     // 网络名
	IpRange *net.IPNet // 网络地址段
	Driver  string     // 网络驱动名
}

// Endpoint 网络端点
type Endpoint struct {
	Id          string           `json:"id"`          // ID 号
	Device      netlink.Veth     `json:"dev"`         // Veth 设备
	IpAddress   net.IP           `json:"ip"`          // ip 地址
	MacAddress  net.HardwareAddr `json:"mac"`         // mac 地址
	PortMapping []string         `json:"portMapping"` // 端口映射
	Network     *Network         // 对应网络
}

// NetworkDriver 网络驱动
// 一个网络功能中的组件，不同的驱动对网络的创建、连接、销毁的策略不同
// 通过在创建网络时指定不同的网络驱动来定义使用那个驱动做网络的配置
type NetworkDriver interface {
	// Name 驱动名称
	Name() string
	// Create 创建网络
	Create(subnet string, name string) (*Network, error)
	// Delete 删除网路
	Delete(network *Network) error
	// Connect 连接容器网络端点到网络
	Connect(network *Network, endpoint *Endpoint) error
	// Disconnect 从网络上移除容器网络端点
	Disconnect(network *Network, endpoint *Endpoint) error
}

// 将网络配置信息保存在文件中
func (nw *Network) dump(dumpPath string) error {
	// 判断配置文件是否存在，不存在则创建
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}

	}

	// 保存的网络配置文件名称，就是网络的名字
	nwPath := path.Join(dumpPath, nw.Name)

	// 打开网络配置文件，打开模式为：
	// os.O_TRUNC 表示如果文件有内容，则清空
	// os.O_CREATE 表示文件不存在创建
	// os.O_WRONLY 表示只写
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Errorf("open network file error：%v", err)
		return err
	}

	defer nwFile.Close()

	// 通过 json 库，将 Network 对象序列化成 json 字符串
	nwJson, err := json.Marshal(nw)
	if err != nil {
		log.Errorf("Marshal Network object error：%v", err)
		return err
	}

	// 把序列化后的 json 字符串写入到网络配置文件中
	_, err = nwFile.Write(nwJson)
	if err != nil {
		log.Errorf("write network file error：%v", err)
		return err
	}

	return nil
}

// 从网络的配置目录中的文件读取到网络的配置
func (nw *Network) load(dumpPath string) error {
	// 打开配置文件
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		log.Errorf("load nw info err: %v", err)
		return err
	}
	defer nwConfigFile.Close()

	// 读取配置文件中的内容
	nwJson := make([]byte, 4096)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		log.Errorf("Read network config file err: %v", err)
		return err
	}

	// 通过 json 库将配置文件中的 json 字符串反序列化成 Network 对象
	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		log.Errorf("Unmarshal network config file err: %v", err)
		return err
	}
	return nil
}

// 从网络的配置目录中的删除配置文件
func (nw *Network) remove(dumpPath string) error {
	nwFile := path.Join(dumpPath, nw.Name)
	// 判断配置文件是否存在，存在则删除
	if _, err := os.Stat(nwFile); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		err := os.Remove(nwFile)
		if err != nil {
			log.Errorf("write network file error：%v", err)
		}
		return err
	}
}

// Init 初始化，从网络配置文件中加载所有网络到 networks 全局字典中
func Init() error {
	// 注册 bridge 网络驱动
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	// 判断网络配置文件的目录是否存在，不存在则创建
	if _, err := os.Stat(networkDefaultPath); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(networkDefaultPath, 0644)
		} else {
			return err
		}
	}

	// 检查网络配置目录下的所有网络配置文件，
	return filepath.Walk(networkDefaultPath, func(nwPath string, info fs.FileInfo, err error) error {
		// 如果是目录，直接跳过
		if info.IsDir() {
			return nil
		}

		// 加载的文件名，就是网络名称
		_, nwName := path.Split(nwPath)
		network := Network{
			Name: nwName,
		}

		// 加载配置文件，序列化到 network 对象中
		err = network.load(nwPath)
		if err != nil {
			return err
		}

		// 将网络信息，加入到全局 Networks 字典中
		networks[nwName] = &network
		return nil
	})
}

// CreateNetwork 创建网络
func CreateNetwork(driver, subnet, name string) error {
	// ParseCIDR 将 网段的 ip 地址转换成 net.IpNet 对象，例如 ParseCIDR("192.0.2.1/24")
	_, cidr, _ := net.ParseCIDR(subnet)
	// 通过 IPAM 组件，分配网关 IP 地址，获取网络的 第一个 IP 地址作为 网关 IP
	gatewayIp, err := IpAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = gatewayIp

	// 调用指定的网络驱动创建网络
	network, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}

	// 保存网络信息，将网络信息保存在网络的配置文件中，以便查询和在网络上连接网络端点。
	return network.dump(networkDefaultPath)

}

// 展示所有的网络
func ListNetwork() error {
	// 通过 tabwrite 格式化输出
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, err := fmt.Fprintf(w, "Name\tIpRange\tDriver\n")
	if err != nil {
		return err
	}

	// 遍历 init 中读取到的所有网络的全局变量，显示
	for _, nw := range networks {
		_, err := fmt.Fprintf(w, "%s\t%s\t%s\n", nw.Name, nw.IpRange.String(), nw.Driver)
		if err != nil {
			return err
		}
	}

	// 刷新
	if err := w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
		return err
	}
	return nil
}

// ConnectNetwork 创建容器时，指定连接的网络
func ConnectNetwork(networkName string, containerInfo *container.ContainerInfo) error {
	// networks 字典中保存了当前已经创建的所有网络
	// 从 networks 字典中获取 docker run --net 指定的网络信息
	network, ok := networks[networkName]
	if !ok { // 如果不存在，返回错误
		return fmt.Errorf("No Such Network: %s", networkName)
	}

	// 通过 IPAM 从网络的网段中，分配一个可用的 IP 地址
	ip, err := IpAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}

	// 创建网络端点 endpoint
	// id 就是 containerID-networkName
	endpoint := &Endpoint{
		Id:        fmt.Sprintf("%s-%s", containerInfo.Id, networkName),
		IpAddress: ip,
		Network:   network,
		/*todo portMapping*/
	}

	// 通过调用 network driver 的 connect 方法，将网络端点与网络进行连接，这里是以  linux-bridge 为例
	// 第一步，将endpoint 的一端连接到
	if err = drivers[network.Driver].Connect(network, endpoint); err != nil {
		return err
	}

	// 第二步，将 endpoint Veth 的另一端连接到 容器的 namespace
	if err = configEndpointIpAddressAndRoute(endpoint, containerInfo); err != nil {
		return err
	}

	return nil
	// todo portmapping
}

// 容器有自己的 network namespace，因此需要将 上一步创建的 veth 设备的一端，添加到容器的 namespace 中
// 才能将该 容器 插上网线，连接到此网络
func configEndpointIpAddressAndRoute(endpoint *Endpoint, containerInfo *container.ContainerInfo) error {
	// 通过 endpoint 的 Device veth 设备，找到 peerName 的设备
	peerLink, err := netlink.LinkByName(endpoint.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}

	// 将容器的网络端点，加入到容器的 network namespace 中
	// 通过 defer 再次函数执行结束后，进入容器网络空间，添加网络后，在恢复到默认网络空间
	// 因为 enterContainerNetns 函数返回的是一个函数，限制性 enterContainerNetns 函数，而 defer 执行的是返回值函数
	// 因此，defer 后的所有代码都是再容器内的 网络空间操作
	defer enterContainerNetns(&peerLink, containerInfo)()

	// 获取容器 Ip 和 地址段，用于配置容器内 veth 设备的接口地址
	// 例如，容器ip是 192.168.11.2，而网络的网段是 192.168.11.0/24，那么这里 interfaceIp 的字符串格式就是 192.168.11.2/24
	interfaceIp := *endpoint.Network.IpRange
	interfaceIp.IP = endpoint.IpAddress

	// 设置 容器端的 veth 设备 ip 地址
	if err = setInterfaceIP(endpoint.Device.PeerName, interfaceIp.String()); err != nil {
		return err
	}

	// 将容器段的 veth 设备启动，set up
	if err = setInterfaceUp(endpoint.Device.PeerName); err != nil {
		return err
	}

	// 默认新建的 network namespace 中的 lo 设备是关闭状态，将其启动
	if err = setInterfaceUp("lo"); err != nil {
		return err
	}

	// 添加路由，设置容器内的 外部请求，都通过容器内的 veth 接口路由
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	// 添加路由时需要的参数，包括网络设备，网关ip，目的网段
	// 相当于命令 route add -net 0.0.0.0/0 gw {bridge 网桥地址} dev {容器内的veth端点 设备}
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        endpoint.Network.IpRange.IP,
		Dst:       cidr,
	}

	// 通过 netlink.RouteAdd 命令添加默认路由
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil
}

// 将容器的 veth 端点加入到容器的 network namespace，锁定当前线程，将当前线程进入到容器的 network namespace 中
// 返回一个函数指针，执行完返回的函数后，才会从容器的网络空间退出，回归到宿主机的默认网络空间中
func enterContainerNetns(enLink *netlink.Link, containerInfo *container.ContainerInfo) func() {
	// 找到容器所在的 network namespace
	// 目录对应 /proc/[pid]/ns/net ，打开该文件的文件描述符，就可以操作 net namespace
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", containerInfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("error set link netns , %v", err)
	}

	// 获取到 network namespace 的文件描述符
	nsFd := f.Fd()

	// 锁定当前的线程，如果不锁定操作系统线程，Goroutine 有可能会调度到别的线程上
	// 从而不能保证在所需要的网络空间中
	runtime.LockOSThread()

	// 将 veth 的另一端，添加到容器的 网络空间中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFd)); err != nil {
		log.Errorf("error set link netns , %v", err)
	}

	// 获取当前的 网络空间，以便后面退出容器网络空间后，能再回到原来的网络空间
	originNs, err := netns.Get()
	if err != nil {
		log.Errorf("error get current netns, %v", err)
	}

	// 设置当前进程到新的网络namespace，并在函数执行完成之后再恢复到之前的namespace
	if err = netns.Set(netns.NsHandle(nsFd)); err != nil {
		log.Errorf("error set netns, %v", err)
	}

	// 返回 defer 执行函数,恢复到原先的网络空间中
	return func() {
		// 恢复到原来的默认网络空间中
		netns.Set(originNs)
		// 关闭 network namespace 文件
		originNs.Close()
		// 取消对线程的锁定
		runtime.UnlockOSThread()
		// 关闭 network namespace 文件描述
		f.Close()

	}
}
