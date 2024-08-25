//go:build linux

package cgroups

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/external/libcontainer/cgroups/v1"
	"github.com/Nevermore12321/dockergsh/external/libcontainer/cgroups/v2"
)

func FindCgroupMountPoint(name string) (string, error) {
	switch Version {
	case "CgroupV1":
		return v1.FindCgroupMountPoint(name)
	case "CgroupV2":
		return v2.FindCgroupMountPoint(name)
	default:
		return "", fmt.Errorf("unsupported cgroup version %q", Version)
	}
}
