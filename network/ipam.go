package network

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"path"
)

// 网络信息的存放文件
// 通过将分配的信息序列化成 json 文件，后者将 json 文件反序列化成结构体
// subnet.json 文件存储了网段对应的分配了的 ip 信息，例如：(0 1 表示对应的 ip 是否被分配)
// 192.168.1.0/24: [0,1,1,1,1,0]
const ipamDefaultAllocatorPath = "/var/lib/dockergsh/network/ipam/subnet.json"

// 存放 ip 地址的分配信息
type IPAM struct {
	// 分配文件存放地址
	SubnetAllocatorPath string

	// 网段和位图算法的数组map key 是网段 value是分配的位图数组
	Subnets *map[string][]byte
}

var IpAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

// 加载网段分配信息
// 也就是将分配的ip信息存入到json文件中，用于判断ip的唯一性
func (ipam *IPAM) load() error {
	// 通过配置的存储网络信息的 json 文件地址，判断文件是否存在，如果不存在表示还没有分配过 ip，不需要加载
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	// 如果 json 文件存在，那么就读取文件内容，并且 序列化成  Subnets
	// 打开文件，并读取内容
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	if err != nil {
		log.Errorf("Open Network subnet.json file err: %v", err)
		return err
	}
	defer subnetConfigFile.Close()

	// 设置读取 buffer
	subnetBytes := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetBytes)
	if err != nil {
		log.Errorf("Read Network subnet.json file err: %v", err)
		return err
	}

	// 将读取的内容反序列化成 ipam.Subnet, 也就是 map，存储的网段 ip 信息
	// 注意，这里只转化读取到的字节数，也就是前 n 个
	err = json.Unmarshal(subnetBytes[:n], ipam.Subnets)
	if err != nil {
		log.Errorf("Unmarshal Network subnet.json file err: %v", err)
		return err
	}
	return nil
}

// 将 ipam.Subnets 序列化成 json 存入到 subnet.json 文件中
func (ipam *IPAM) dump() error {
	// 判断文件 subnet.json 文件是否存在，如果不存在，就创建，并且添加内容
	// path.Split 将文件路径，分成目录和文件名
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	// 判断文件夹是否存在，不存在则创建(第一次需要创建文件夹）
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			// 如果不存在，创建文件夹，os.MkdirAll 相当于 mkdir -p
			err := os.MkdirAll(ipamConfigFileDir, 0644)
			if err != nil {
				log.Errorf("Create Network subnet.json folder err: %v", err)
				return err
			}
		} else {
			log.Errorf("State Network folder err: %v", err)
			return err
		}
	}

	// 打开 subnet.json 文件，打开的模式为：O_TRUNC 表示如果存在就清空，O_CREATE 表示不存在就创建
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Errorf("Open Network subnet.json file err: %v", err)
		return err
	}
	defer subnetConfigFile.Close()

	// 将 ipam.Subnets 序列化成 json 格式的 bytes 数组
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		log.Errorf("Marshal Network subnet.json file err: %v", err)
		return err
	}

	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		log.Errorf("Write to Network subnet.json file err: %v", err)
		return err
	}
	return nil
}

// Allocator 从指定的子网中，分配 ip 地址
// subnet: 子网
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 初始化存放网段中 地址分配信息的数组
	ipam.Subnets = &map[string][]byte{}
	// 从 subnet.json 配置文件中读取相应ip分配信息
	if err = ipam.load(); err != nil {
		// 这里报错不会退出，按照 subnet.json 文件不存在处理
		log.Errorf("Error load allocation info, %v", err)
	}

	if _, subnet, err = net.ParseCIDR(subnet.String()); err != nil {
		log.Errorf("Subnet convert to IPNet err: %v", err)
		return nil, err
	}

	// Size 返回掩码中前导个数和主机位总位数，其实就是返回子网掩码的总长度和网段前面的固定位的长度
	// 例如：127.0.0.1/8 网段的子网掩码是 127.0.0.0, 返回的就是 8,32
	one, size := subnet.Mask.Size()

	// 如果之前没有分配过这个网段的 ip 地址，那么就初始化网段的分配配置
	// 也就是初始化 bitmap 的 位数组 都为0
	if _, ok := (*ipam.Subnets)[subnet.String()]; !ok {
		// 例如 192.168.0.0/24 , 上面的 size 函数返回的是 24,32，也就是最后有 8 位可以分配，一共有 2^8 也就是 256
		// 通常计算 2 的 n 次方，可以使用 1 向左移位，2^8 相当于 1<<8
		// 初始化 2^(32-mask) 长度的数组，默认值是 0
		(*ipam.Subnets)[subnet.String()] = make([]byte, 1<<(size-one))
	}

	// 遍历位图数组，找到第一个 位数位 0 的 ip 地址，分配出去
	for index := range (*ipam.Subnets)[subnet.String()] {
		// 如果数组的 位数为0，可以分配
		if (*ipam.Subnets)[subnet.String()][index] == 0 {
			ipalloc := (*ipam.Subnets)[subnet.String()]
			// 对应位 置为 1，表示当前 ip 已被分配
			ipalloc[index] = 1
			// 修改原始的 ipam.Subnets
			(*ipam.Subnets)[subnet.String()] = ipalloc

			// 取出网段 ip，例如上面的 192.168.0.0/24 网段，ip 为 192.168.0.0
			ip = subnet.IP
			// ip 加上位数组的 偏移量，也就是当前位数组的 索引位置
			// IPNet.IP 其实就是 []byte 类型, 转换的方法是：192.168.0.0 -> [192, 168, 0, 0]
			// 怎么找到位数组对应索引位置的 ip 地址呢？这就需要将 ip 的四个部分，每个 部分都要加上偏移量
			// 注意，这里使用了 uint8，也就是 8 位来截断二进制
			// 例如 172.16.0.0/12，数组的序号是65555，转换成二进制就是分成四组后 (0000 0000) (0000 0001) (0000 0000) (0001 0011)，这里使用移位+低8位截断的方法，获取对应的偏移量
			// 最终的偏移量为：[0, 1, 0, 19]，对应[ 65555 >> 24(从右截断8位), 65555 >> 16(从右截断8位), 65555 >> 8(从右截断8位), 65555 >> 0(从右截断8位)]
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(index >> ((t - 1) * 8))
			}

			// ip 是从 1 开始分配的，也就是 0 位置存放的是 1
			ip[3] += 1
			break
		}
	}

	// 将分配后的信息写入到配置文件
	err = ipam.dump()
	if err != nil {
		log.Errorf("Write to Network subnet.json file err after allocate: %v", err)
	}
	return

}

// Release 释放 IP 地址，并保存到配置文件 subnet.json
// subnet: 子网
// ipaddr: 从子网中释放的ip地址
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) (err error) {
	//	初始化 ipam.Subnets
	ipam.Subnets = &map[string][]byte{}

	if _, subnet, err = net.ParseCIDR(subnet.String()); err != nil {
		log.Errorf("Subnet convert to IPNet err: %v", err)
	}

	// 从 配置文件中 加载已分配的 ip 信息
	if err = ipam.load(); err != nil {
		log.Errorf("Error load allocation info, %v", err)
	}

	// 计算该 ip 地址在 位图数组中的位置
	index := 0
	// 将 ipaddr 转换成 4 个字节的表示方式，也就是 [192, 168, 0, 0]
	releaseIP := ipaddr.To4()
	// 位图数组从ip为 1 开始存储，因此 -1
	releaseIP[3] -= 1

	// 计算偏移量
	// 例如要释放的 IP 172.17.0.20, 减一为 172.17.0.19，子网段为 172.16.0.0，那么对应部分相减，得到 [0, 1, 0, 19]，转换成二进制为[(0000 0000), (0000 0001), (0000 0000), (0001 0011)]
	// [0, 1, 0, 18] 转成二进制，就通过 移位实现，最终结果就是 0001 0000 0000 0001 0011，十进制就是 65555
	for t := uint(4); t > 0; t -= 1 {
		subnetIP := subnet.IP.To4()
		index += int(releaseIP[t-1]-subnetIP[t-1]) << ((4 - t) * 8)
	}

	// 修改对应索引的位图数组为0
	ipalloc := (*ipam.Subnets)[subnet.String()]
	ipalloc[index] = 0
	(*ipam.Subnets)[subnet.String()] = ipalloc

	// 修改完后，存入配置文件
	err = ipam.dump()
	if err != nil {
		log.Errorf("Write to Network subnet.json file err after release: %v", err)
	}

	return
}
