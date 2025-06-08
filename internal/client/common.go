package client

import (
	log "github.com/sirupsen/logrus"
	"mime"
)

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
	VERSION               = "0.0.1"
	DOCKERGSH_DEBUG       = "DOCKERGSH_DEBUG"
	DOCKERGSH_CONFIG_HOST = "DOCKERGSH_CONFIG_HOST"
	DOCKERGSH_SERVER_HOST = "DOCKERGSH_SERVER_HOST"
)

func MatchesContentType(contentType, expectedType string) bool {
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Errorf("Error parsing media type: %s error: %s", contentType, err.Error())
	}

	return err == nil && mediatype == expectedType
}
