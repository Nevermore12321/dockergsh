//go:build linux

package daemongsh

import "github.com/Nevermore12321/dockergsh/external/libcontainer/selinux"

// 关闭 selinux
func selinuxSetDisabled() {
	selinux.SetDisabled()
}

func selinuxFreeLxcContexts(label string) {
	selinux.FreeLxcContexts(label)
}
