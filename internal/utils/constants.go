package utils

import "os"

// dockergsh dockergsh 相关默认值
const (
	DEFAULTHTTPHOST   = "127.0.0.1"
	DEFAULTUNIXSOCKET = "/var/run/docker.sock"
)

// dockergsh dockergsh 相关的环境变量名称
const (
	DOCKERGSH_DEBUG       = "DOCKERGSH_DEBUG"
	DOCKERGSH_CONFIG_HOST = "DOCKERGSH_CONFIG_HOST"
	DOCKERGSH_SERVER_HOST = "DOCKERGSH_SERVER_HOST"
)

var (
	DockerCertPath  = os.Getenv("DOCKERGSH_CERT_PATH")
	DefaultCaFile   = "ca.pem"
	DefaultKeyFile  = "key.pem"
	DefaultCertFile = "cert.pem"
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
