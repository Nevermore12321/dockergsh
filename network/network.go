package network

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"io/fs"
	"net"
	"os"
	"path"
	"path/filepath"
)

var (
	networkDefaultPath = "/var/lib/dockergsh/network/network"
	// drivers 字典，是各个网络驱动的实例字典
	drivers = map[string]NetworkDriver{}
	// networks 字典，是所有网络的实例字典
	networks = map[string]Network{}
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
			os.MkdirAll(networkDefaultPath, 0644)
		} else {
			return err
		}
	}

	// 检查网络配置目录下的所有网络配置文件，
	filepath.Walk(networkDefaultPath, func(nwPath string, info fs.FileInfo, err error) error {
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
		networks[nwName] = network
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
