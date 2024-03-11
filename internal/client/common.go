package client

// dockergsh client 相关默认值
const (
	DEFAULTHTTPHOST   = "127.0.0.1"
	DEFAULTUNIXSOCKET = "/var/run/docker.sock"
)

var (
	GITCOMMIT string
)

// dockergsh client 相关的环境变量名称
const (
	DOCKERGSH_DEBUG       = "DOCKERGSH_DEBUG"
	DOCKERGSH_CONFIG_HOST = "DOCKERGSH_CONFIG_HOST"
	DOCKERGSH_SERVER_HOST = "DOCKERGSH_SERVER_HOST"
)
