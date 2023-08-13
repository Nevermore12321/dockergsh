package client

const (
	DEFAULTHTTPHOST   = "127.0.0.1"
	DEFAULTUNIXSOCKET = "/var/run/docker.sock"
)

var (
	GITCOMMIT string
	VERSION   string
)

// dockergsh client 相关的环境变量名称
const (
	DOCKERGSH_DEBUG          = "DOCKERGSH_DEBUG"
	DOCKERGSH_HOST           = "DOCKERGSH_HOST"
	DOCKERGSH_DEFAULT_SOCKET = "DOCKERGSH_DEFAULT_SOCKET"
)

// dockergsh client 相关默认值
const (
	DEFAULTHTTPHOST   = "127.0.0.1"
	DEFAULTUNIXSOCKET = "/var/run/docker.sock"
)
