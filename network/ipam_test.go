package network

import (
	"fmt"
	"net"
	"testing"
)

func TestAllocate(t *testing.T) {
	// 在192.168.0.0/24网段内分配IP
	_, ipnet, _ := net.ParseCIDR("192.168.0.0/24")
	ip, _ := IpAllocator.Allocate(ipnet)
	t.Logf("alloc ip: %v", ip)
}

func TestRelease(t *testing.T) {
	// 在192.168.0.0/24网段中释放方才分配的IP
	_, ipnet, _ := net.ParseCIDR("192.168.0.2/24")
	ip := net.ParseIP("192.168.0.1")
	fmt.Println(ip)
	IpAllocator.Release(ipnet, &ip)
	fmt.Println(ip)
}
