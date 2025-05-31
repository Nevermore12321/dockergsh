package parse

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseHost 解析 addr 参数
// addr: 需要解析的 addr
// defaultHost: 默认的 addr 设置
// defaultUnix: 默认的 Unix Socks 设置
func ParseHost(addr, defaultHost, defaultUnix string) (string, error) {
	var (
		proto string
		host  string
		port  int
	)
	addr = strings.TrimSpace(addr)

	// 解析 addr 中的 协议前缀
	switch {
	case addr == "tcp://":
		return "", fmt.Errorf("Invalid bind address format: %s", addr)
	case strings.HasPrefix(addr, "unix://"):
		proto = "unix"
		addr = strings.TrimPrefix(addr, "unix://")
		if addr == "" {
			addr = defaultUnix
		}
	case strings.HasPrefix(addr, "tcp://"):
		proto = "tcp"
		addr = strings.TrimPrefix(addr, "tcp://")
	case strings.HasPrefix(addr, "fd://"):
		return addr, nil
	case addr == "":
		addr = defaultUnix
		proto = "unix"
	default:
		if strings.Contains(addr, "://") {
			return "", fmt.Errorf("Invalid bind address protocol: %s", addr)
		}
		proto = "tcp"
	}

	// unix，并且格式为 tcp://ip:port
	if proto != "unix" && strings.Contains(addr, ":") {
		hostParts := strings.Split(addr, ":")
		if len(hostParts) != 2 {
			return "", fmt.Errorf("Invalid bind address format: %s", addr)
		}
		// host
		if hostParts[0] != "" {
			host = hostParts[0]
		} else {
			host = defaultHost
		}

		// port
		if p, err := strconv.Atoi(hostParts[1]); err == nil && p != 0 {
			port = p
		} else {
			return "", fmt.Errorf("Invalid bind address format: %s", addr)
		}
	} else if proto == "tcp" && !strings.Contains(addr, ":") { // tcp://ip:port, 如果没有 : , 格式错误
		return "", fmt.Errorf("Invalid bind address format: %s", addr)
	} else { // unix:///var/run/dockergsh.sock
		host = addr
	}

	if proto == "unix" {
		return fmt.Sprintf("%s://%s", proto, host), nil
	}
	return fmt.Sprintf("%s://%s:%d", proto, host, port), nil
}

// ParseRepositoryTag 通过镜像url解析出镜像的仓库地址和 tag
func ParseRepositoryTag(repos string) (string, string) {
	// 找到最后一个冒号的位置，冒号后就是 tag，如果没有设置，那么默认 latest
	n := strings.LastIndex(repos, ":")
	if n < 0 {
		return repos, "latest"
	}

	if tag := repos[n+1:]; !strings.Contains(tag, "/") {
		return repos[:n], tag
	}
	return repos, "latest"
}
