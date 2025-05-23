//go:build linux

package cgroups

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/external/libcontainer/cgroups/v1"
	"github.com/Nevermore12321/dockergsh/external/libcontainer/cgroups/v2"
)

func GetCgroupPath(name string) (string, error) {
	switch Version {
	case "CgroupV1":
		return v1.GetCgroupPath(name, true)
	case "CgroupV2":
		return v2.GetCgroupPath(name, true)
	default:
		return "", fmt.Errorf("unsupported cgroup version %q", Version)
	}
}
