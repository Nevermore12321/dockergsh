//go:build !linux

package daemongsh

func selinuxSetDisabled() {
}

func selinuxFreeLxcContexts(label string) {
}
