package iptables

import (
	"errors"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/client"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//
/*
 * iptables -t 表名 <-A/I/D/R> 规则链名 [规则号] <-i/o 网卡名> -p 协议名 <-s 源IP/源子网> --sport 源端口 <-d 目标IP/目标子网> --dport 目标端口 -j 动作
 * 选项包括：
 * 	- -t<表>：指定要操纵的表；
 * 	- -A：向规则链中添加条目；
 * 	- -D：从规则链中删除条目；
 * 	- -I：向规则链中插入条目；
 * 	- -R：替换规则链中的条目；
 * 	- -L：显示规则链中已有的条目；
 * 	- -F：清楚规则链中已有的条目；
 * 	- -Z：清空规则链中的数据包计算器和字节计数器；
 * 	- -N：创建新的用户自定义规则链；
 * 	- -P：定义规则链中的默认目标；
 * 	- -h：显示帮助信息；
 * 	- -p：指定要匹配的数据包协议类型；
 * 	- -s：指定要匹配的数据包源ip地址；
 * 	- -j<目标>：指定要跳转的目标；
 * 	- -i<网络接口>：指定数据包进入本机的网络接口；
 * 	- -o<网络接口>：指定数据包要离开本机所使用的网络接口。
 *
 * 表名包括：
 *  - raw：高级功能，如：网址过滤。
 *  - mangle：数据包修改（QOS），用于实现服务质量。
 *  - net：地址转换，用于网关路由器。
 *  - filter：包过滤，用于防火墙规则。
 *
 * 规则链名包括：
 *  - INPUT链：处理输入数据包。
 *  - OUTPUT链：处理输出数据包。
 *  - PORWARD链：处理转发数据包。
 *  - PREROUTING链：用于目标地址转换（DNAT）。
 *  - POSTOUTING链：用于源地址转换（SNAT）。
 *
 * 动作包括：
 * 	- accept：接收数据包。
 *  - DROP：丢弃数据包。
 *  - REDIRECT：重定向、映射、透明代理。
 *  - SNAT：源地址转换。
 *  - DNAT：目标地址转换。
 *  - MASQUERADE：IP伪装（NAT），用于ADSL。
 *  - LOG：日志记录。
 *  - 还可以指定其他链（Chain）作为目标
 */

var (
	ErrIptablesNotFound = errors.New("iptables not found")
	supportsXlock       = false
	nat                 = []string{"-t", "nat"}
)

type Action string

const (
	Add    Action = "-A" // iptables 添加 chain 操作
	Delete Action = "-D" // iptables 删除 chain 操作
)

// Chain iptables chain 的封装
type Chain struct {
	Name   string
	Bridge string
}

func init() {
	// 检查 iptbales 是否安装
	supportsXlock = exec.Command("iptables", "--wait", "-L", "-n").Run() == nil
}

func NewChain(name, bridge string) (*Chain, error) {
	// 在NAT表中建立用户自定义链，-N 建立用户定义链
	if output, err := Raw("-t", "nat", "-N", name); err != nil {
		return nil, err
	} else if len(output) != 0 {
		return nil, fmt.Errorf("error creating new iptables chain: %s", output)
	}
	chain := &Chain{
		Name:   name,
		Bridge: bridge,
	}

	if err := chain.PreRouting(Add, "-m", "addrtype", "--dst-type", "LOCAL"); err != nil {
		return nil, fmt.Errorf("failed to inject dockergsh in PREROUTING chain: %s", err)
	}
	if err := chain.Output(Add, "-m", "addrtype", "--dst-type", "LOCAL", "!", "--dst", "127.0.0.0/8"); err != nil {
		return nil, fmt.Errorf("failed to inject dockergsh in OUTPUT chain: %s", err)
	}
	return chain, nil
}

// Forward 向 Forward 规则链中添加/删除规则，配置 DNAT 转发，转发数据包时应用的规则
// 所有发往 ip:port 的数据包，通过 DNAT 转换，通过 chain.Bridge 设备，修改目的地址为 destAddr:destPort
func (c *Chain) Forward(action Action, ip net.IP, port int, proto, destAddr string, destPort int) error {
	daddr := ip.String()
	if ip.IsUnspecified() { // 如果 ip 没有指定地址
		daddr = "0/0"
	}

	// 执行 iptables 命令：iptables -t nat -A [chain.Name] -p [proto] -d [daddr] --dport [dport] ! -i [eth] -j DNAT --to-destination [daddr:dport]
	if output, err := Raw(append(nat, fmt.Sprint(action), c.Name,
		"-p", proto,
		"-d", daddr,
		"--dport", strconv.Itoa(port),
		"!", "-i", c.Bridge,
		"-j", "DNAT",
		"--to-destination", net.JoinHostPort(destAddr, strconv.Itoa(destPort)))...); err != nil {
		return err
	} else if len(output) != 0 {
		return fmt.Errorf("error iptables forward: %s", output)
	}

	fAction := action
	if fAction == Add {
		fAction = "-I"
	}

	// 执行 iptables 命令：ipbtales -I FORWARD ! -i [chain.Bridge] -o [chain.Bridge] -p proto -d [destAddr] --dport [destPort] -j ACCEPT
	// 允许转发
	if output, err := Raw(string(fAction), "FORWARD",
		"!", "-i", c.Bridge,
		"-o", c.Bridge,
		"-p", proto,
		"-d", destAddr,
		"--dport", strconv.Itoa(destPort),
		"-j", "ACCEPT"); err != nil {
		return err
	} else if len(output) != 0 {
		return fmt.Errorf("error iptables forward: %s", output)
	}
	return nil
}

// PreRouting 向 REROUTING 规则链中添加/删除规则，默认走 chain.Name，对数据包作路由选择时先应用的规则
func (c *Chain) PreRouting(action Action, args ...string) error {
	a := append(nat, fmt.Sprint(action), "PREROUTING")
	if len(args) > 0 {
		a = append(a, args...)
	}

	// 执行命令 iptables -t nat -A PREROUTING [args...] -j [chain.Name]
	// -j 代表 "jump to target",指定了当与规则(Rule)匹配时如何处理数据包,可能的值是:
	//		- ACCEPT, DROP, QUEUE, RETURN，MASQUERADE
	//		- 还可以指定其他链（Chain）作为目标
	if output, err := Raw(append(a, "-j", c.Name)...); err != nil {
		return err
	} else if len(output) != 0 {
		return fmt.Errorf("error prerouting iptables: %s", output)
	}

	return nil
}

// Output 向 OUTPUT 规则链中添加/删除 chain，外出的数据包应用的规则
func (c *Chain) Output(action Action, args ...string) error {
	a := append(nat, fmt.Sprint(action), "OUTPUT")
	if len(args) > 0 {
		a = append(a, args...)
	}

	// 执行 iptables 命令：iptables -t nat -A/-D OUTPUT -j [chaim.Name]
	// 在 nat 表中的 output 链添加规则
	if output, err := Raw(append(a, "-j", c.Name)...); err != nil {
		return err
	} else if len(output) != 0 {
		return fmt.Errorf("error output iptables: %s", output)
	}
	return nil
}

// Remove 删除 chain
func (c *Chain) Remove() error {
	// Ignore errors - This could mean the chains were never set up
	c.PreRouting(Delete, "-m", "addrtype", "--dst-type", "LOCAL")
	c.Output(Delete, "-m", "addrtype", "--dst-type", "LOCAL", "!", "--dst", "127.0.0.0/8")
	c.Output(Delete, "-m", "addrtype", "--dst-type", "LOCAL") // Created in versions <= 0.1.6

	c.PreRouting(Delete)
	c.Output(Delete)

	Raw("-t", "nat", "-F", c.Name)
	Raw("-t", "nat", "-X", c.Name)

	return nil
}

// RemoveExistingChain 删除特定名字的 iptables 链
func RemoveExistingChain(name string) error {
	chain := &Chain{Name: name}
	return chain.Remove()
}

// Exists 判断当前 itables 规则是否已经存在，args 表示 iptables 后面跟的参数
func Exists(args ...string) bool {
	// -C 选项用于检查一条规则是否已经存在。它不会添加或删除规则，只是查看指定的规则是否已经在防火墙中配置。
	// 如果规则存在，命令返回 0（成功）；如果规则不存在，返回 1（失败）。
	if _, err := Raw(append([]string{"-C"}, args...)...); err != nil {
		return false
	}
	return true
}

// Raw 执行 iptables 命令，args 表示 iptables 后面跟的参数
func Raw(args ...string) ([]byte, error) {
	// 判断是否存在 iptables 命令
	path, err := exec.LookPath("iptables")
	if err != nil {
		return nil, ErrIptablesNotFound
	}

	// --wait 选项用于防止 iptables 在规则集被其他命令修改时退出。在规则集被锁定时，iptables 将等待锁定解除，而不是立即退出。
	// 这对于防止多线程或多进程修改 iptables 时发生竞争条件很有用
	if supportsXlock {
		args = append([]string{"--wait"}, args...)
	}

	// debug 信息
	if os.Getenv(client.DOCKERGSH_DEBUG) != "" {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("[debug] %s, %v \n", path, args))
	}

	// 执行 iptables 命令
	fmt.Println(args)
	output, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iptables failed: iptables %v: %s (%s)", strings.Join(args, " "), output, err)
	}

	// 忽略 iptables 关于 xtables 锁的消息
	if strings.Contains(string(output), "waiting for it to exit") {
		output = []byte("")
	}
	return output, nil
}
