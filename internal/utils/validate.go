package utils

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/pkg/parse"
	log "github.com/sirupsen/logrus"
	"net"
	"regexp"
	"strings"
)

var (
	ErrFormatString = "the format of the %s you entered %s is incorrect"
)

// Validates 校验 hosts 参数是否正确
func Validates(vals []string, validate func(string) (string, error)) []string {
	var parsed []string
	for _, val := range vals {
		parsedVal, err := validate(val)
		if err == nil {
			log.Warningf("%s is not a correct format", val)
			continue
		}
		parsed = append(parsed, parsedVal)
	}
	return parsed
}

func ValidateHost(val string) (string, error) {
	parsedHost, err := parse.ParseHost(val, DefaultHttpHost, DefaultUnixSocket)
	if err != nil {
		return "", err
	}
	return parsedHost, nil
}

// ValidateIPAddress 校验 IP 地址格式是否正确，返回字符串
func ValidateIPAddress(val string) (string, error) {
	var ip = net.ParseIP(strings.TrimSpace(val))
	if ip != nil { // 转换 ip 成功
		return ip.String(), nil
	}
	return "", fmt.Errorf(ErrFormatString, "Ip address", val)
}

// ValidateDnsSearch 验证 resolvconf 搜索配置的域。
// /etc/resolv.conf是DNS客户机的配置文件，用于设置DNS服务器的IP地址及DNS域名，还包含了主机的域名搜索顺序。
// resolv.conf的关键字主要有4个，分别为：
//   - nameserver：定义DNS服务器的IP地址，例如 `nameserver 8.8.8.8`
//   - domain：定义本地域名，例如 `domain  xxx.com`
//   - search：定义域名的搜索列表，例如 `search  www.xxx.com  xxx.com`
//   - sortlist：对返回的域名进行排序
func ValidateDnsSearch(val string) (string, error) {
	// 如果配置的是  `.` 直接返回
	if val = strings.Trim(val, " "); val == "." {
		return val, nil
	}
	return validateDomain(val)
}

func validateDomain(val string) (string, error) {
	// 正则表达式，大小写字母，如果没有字母，直接返回错误
	alphaReg := regexp.MustCompile(`[a-zA-z]`)
	if alphaReg.FindString(val) == "" {
		return "", fmt.Errorf(ErrFormatString, "domain", val)
	}

	// 可以是 ip 地址，也可以是域名
	reg := regexp.MustCompile(`^(:?(:?[a-zA-Z0-9]|(:?[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]))(:?\.(:?[a-zA-Z0-9]|(:?[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])))*)\.?\s*$`)
	nameServer := reg.FindSubmatch([]byte(val)) // nameserver 8.8.8.8
	if len(nameServer) > 0 {
		return string(nameServer[1]), nil
	}
	return "", fmt.Errorf(ErrFormatString, "domain", val)
}
