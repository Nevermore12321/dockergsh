package client

import (
	"github.com/Nevermore12321/dockergsh/pkg/parse"
)

func ValidateHost(host string) (string, error) {
	parsedHost, err := parse.ParseHost(host, DEFAULTHTTPHOST, DEFAULTUNIXSOCKET)
	if err != nil {
		return host, err
	}
	return parsedHost, nil
}
