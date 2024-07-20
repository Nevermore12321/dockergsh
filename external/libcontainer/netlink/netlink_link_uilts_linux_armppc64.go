//go:build arm || ppc64 || ppc64le
// +build arm ppc64 ppc64le

package netlink

// byte to uint8
func ifrDataByte(b byte) uint8 {
	return uint8(b)
}
