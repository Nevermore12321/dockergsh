package selinux

var (
	selinuxEnabled        = false // 当前是否启用selinux
	selinuxEnabledChecked = false // 是否已检查过 selinux 的开关
)

// SetDisabled 禁用 selinux
func SetDisabled() {
	selinuxEnabled = false
	selinuxEnabledChecked = true // 已经检查，已禁用
}
