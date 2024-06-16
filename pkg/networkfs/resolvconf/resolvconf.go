package resolvconf

import (
	"bytes"
	"os"
	"regexp"
)

// Get 获取 /etc/resolv.conf 配置的 dns server
func Get() ([]byte, error) {
	resolv, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	return resolv, nil
}

// getLines 将输入解析为行并去除注释，例如 "nameserver 10.10.10.10 # comment test" => "nameserver 10.10.10.10"
// commentMaker 表示注释标识符，这里是 #
func getLines(input []byte, commentMaker []byte) [][]byte {
	// 换行
	lines := bytes.Split(input, []byte("\n"))

	var output [][]byte
	for _, currentLine := range lines {
		// 判断有没有注释标识符
		var commentIndex = bytes.Index(currentLine, commentMaker)
		if commentIndex == -1 { // 当前行不是注释
			output = append(output, currentLine)
		} else { // 当前行是注释，则去掉注释标识符
			output = append(output, currentLine[:commentIndex])
		}
	}
	return output
}

// GetNameservers 将 /etc/resolv.conf 中的配置 nameserver 10.10.10.10，只保留 dns 地址
func GetNameservers(resolvConf []byte) []string {
	nameservers := []string{}
	// 格式必须正确，nameserver xx.xx.xx.xx
	re := regexp.MustCompile(`^\s*nameserver\s*(([0-9]+\.){3}([0-9]+))\s*$`)
	// 过滤掉注释信息，只保留配置
	for _, line := range getLines(resolvConf, []byte("#")) {
		ns := re.FindSubmatch(line)
		if len(ns) > 0 {
			nameservers = append(nameservers, string(ns[1]))
		}
	}
	return nameservers
}

// GetNameserversAsCIDR 将 nameserver 的配置 ip 地址转成 CIDR 例如 10.10.10.10/32
func GetNameserversAsCIDR(resolvConf []byte) []string {
	nameservers := []string{}

	// 将 resolv.conf 中配置的 dns 地址转成 cidr 格式
	for _, nameserver := range GetNameservers(resolvConf) {
		nameservers = append(nameservers, nameserver+"/32")
	}

	return nameservers
}
