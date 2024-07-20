//go:build !arm && !ppc64 && !ppc64le
// +build !arm,!ppc64,!ppc64le

package netlink

// byte to int8
func ifrDataByte(b byte) int8 {
	return int8(b)
}
