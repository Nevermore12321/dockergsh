package utils

import "os"

// dockergsh dockergsh 相关默认值
const (
	DefaultHttpHost   = "127.0.0.1"
	DefaultUnixSocket = "/var/run/dockergsh.sock"
)

// dockergsh dockergsh 相关的环境变量名称
const (
	DockergshDebug      = "DOCKERGSH_DEBUG"
	DockergshConfigHost = "DOCKERGSH_CONFIG_HOST"
	DockergshHosts      = "DOCKERGSH_HOSTS"
)

var (
	DockergshCertPath = os.Getenv("DOCKERGSH_CERT_PATH")
	DefaultCaFile     = "ca.pem"
	DefaultKeyFile    = "key.pem"
	DefaultCertFile   = "cert.pem"
)

/* ============ServeApi 环境变量名称================= */
const (
	Logging        = "DOCKERGSH_Logging"
	EnableCors     = "DOCKERGSH_EnableCors"
	Version        = "DOCKERGSH_Version"
	SocketGroup    = "DOCKERGSH_SocketGroup"
	Tls            = "DOCKERGSH_Tls"
	TlsVerify      = "DOCKERGSH_TlsVerify"
	TlsCa          = "DOCKERGSH_TlsCa"
	TlsCert        = "DOCKERGSH_TlsCert"
	TlsKey         = "DOCKERGSH_TlsKey"
	BufferRequests = "DOCKERGSH_BufferRequests"
)

/* ==============daemon 相关====================*/
const (
	NowarnKernelVersion = "DOCKERGSH_NOWARN_KERNEL_VERSION"
	ConfigTempdir       = "DOCKERGSH_TMPDIR"
	DaemongshTempdir    = "TEMPDIR"
	GraphDriver         = "DOCKDERGSH_GRAPHDRIVER"
)
