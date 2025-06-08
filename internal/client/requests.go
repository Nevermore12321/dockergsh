package client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/api_server"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"strings"
)

func (cli *DockerGshClient) HttpClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: cli.tlsConfig,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(cli.proto, cli.addr)
		},
	}

	return &http.Client{Transport: transport}
}

func (cli *DockerGshClient) Stream(method, path string, in io.Reader, out io.Writer, headers map[string][]string) error {
	return cli.streamHelper(method, path, true, in, out, nil, headers)
}

// 发送请求
func (cli *DockerGshClient) streamHelper(method, path string, setRawTerminal bool, in io.Reader, stdout, stderr io.Writer, headers map[string][]string) error {
	// 如果 post 和 put 请求，没有请求体时，初始化为空
	if (method == "POST" || method == "PUT") && in == nil {
		in = bytes.NewReader([]byte{})
	}

	// 初始化请求
	req, err := http.NewRequest(method, fmt.Sprintf("http://v%s%s", api_server.APIVERSION, path), in)
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Docker-Client/"+VERSION)
	req.URL.Host = cli.addr
	req.URL.Scheme = cli.scheme
	if method == "POST" {
		req.Header.Set("Content-Type", "plain/text")
	}
	if headers != nil {
		for k, v := range headers {
			req.Header[k] = v
		}
	}

	// 执行请求
	resp, err := cli.HttpClient().Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf("cannot connect to the Docker daemon. Is 'docker -d' running on this host")
		}
		return err
	}

	defer resp.Body.Close()

	// 响应码判断
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if len(body) == 0 {
			return fmt.Errorf("error :%s", http.StatusText(resp.StatusCode))
		}

		return fmt.Errorf("error: %s", bytes.TrimSpace(body))
	}

	// 解析响应，如果时 json 格式，记录日志
	if MatchesContentType(resp.Header.Get("Content-Type"), "application/json") {
		return utils.DisplayJSONMessagesStream(resp.Body, stdout, cli.terminalFd, cli.isTerminal)
	}

	// 否则，如果 setRawTerminal 打开，则拷贝
	if stdout != nil || stderr != nil {
		if setRawTerminal {
			_, err = io.Copy(stdout, resp.Body)
		} else {
			_, err = utils.StdCopy(stdout, stderr, resp.Body)
		}
		log.Debugf("[Stream] End of stdout")
		return err
	}

	return nil
}
