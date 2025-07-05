package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/api_server"
	"github.com/Nevermore12321/dockergsh/internal/engine"
	"github.com/Nevermore12321/dockergsh/internal/registry"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"strings"
)

var (
	ErrConnectionRefused = errors.New("cannot connect to the Docker daemon. Is 'docker -d' running on this host")

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

// 发送 stream 请求
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

// Call 发送 api 请求
func (cli *DockerGshClient) Call(method, path string, data interface{}, passAuthInfo bool) (io.ReadCloser, int, error) {
	// request parameters
	params := bytes.NewBuffer(nil)
	if data != nil {
		if env, ok := data.(engine.Env); ok {
			if err := env.Encode(params); err != nil {
				return nil, -1, err
			}
		} else {
			buf, err := json.Marshal(data)
			if err != nil {
				return nil, -1, err
			}

			if _, err := params.Write(buf); err != nil {
				return nil, -1, err
			}
		}
	}

	// 初始化 request 实例
	req, err := http.NewRequest(method, fmt.Sprintf("/v%s%s", api_server.APIVERSION, path), params)
	if err != nil {
		return nil, -1, err
	}

	// 是否需要认证信息
	if passAuthInfo {
		_ = Client.LoadConfigFile()
		// 找到对应此服务器相关的auth配置
		authConfig := Client.ConfigFile.ResolveAuthConfig(registry.IndexServerAddress())
		getHeaders := func(authConfig registry.AuthConfig) (map[string][]string, error) {
			buf, err := json.Marshal(authConfig)
			if err != nil {
				return nil, err
			}
			registryAuthHeader := []string{
				base64.URLEncoding.EncodeToString(buf),
			}
			return map[string][]string{
				"X-Registry-Auth": registryAuthHeader,
			}, nil
		}

		if headers, err := getHeaders(authConfig); err == nil && headers != nil {
			for k, v := range headers {
				req.Header[k] = v
			}
		}
	}

	req.Header.Set("User-Agent", "Dockergsh-Client/"+VERSION)
	req.URL.Host = cli.addr
	req.URL.Scheme = cli.scheme
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	} else if method == "POST" {
		req.Header.Set("Content-Type", "plain/text")
	}

	// 发送请求
	resp, err := cli.HttpClient().Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return nil, -1, ErrConnectionRefused
		}
		return nil, -1, err
	}

	// 错误处理
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, -1, err
		}
		if len(body) == 0 {
			return nil, resp.StatusCode, fmt.Errorf("error: request returned %s for API route and version %s, check if the server supports the requested API version", http.StatusText(resp.StatusCode), req.URL)
		}
		return nil, resp.StatusCode, fmt.Errorf("error response from daemon: %s", bytes.TrimSpace(body))
	}
	return resp.Body, resp.StatusCode, nil
}
