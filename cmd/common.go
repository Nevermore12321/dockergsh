package service

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/Nevermore12321/dockergsh/pkg/parse"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrMultiHosts = errors.New("Please specify only one -H")
)

func CmdFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"D"},
			Usage:   "Enable debug mode",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:  "tls-verify",
			Usage: "Use TLS and verify the remote",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "tls-cacert",
			Usage: "Trust only remotes providing a certificate signed by the CA given here",
			Value: filepath.Join(utils.DockerCertPath, utils.DefaultCaFile),
		},
		&cli.StringFlag{
			Name:  "tls-cert",
			Usage: "Path to TLS certificate file",
			Value: filepath.Join(utils.DockerCertPath, utils.DefaultCertFile),
		},
		&cli.StringFlag{
			Name:  "tls-key",
			Usage: "Path to TLS key file",
			Value: filepath.Join(utils.DockerCertPath, utils.DefaultKeyFile),
		},
		&cli.StringSliceFlag{
			Name:  "hosts",
			Usage: "The socket(s) to bind to in daemongsh mode\nspecified using one or more tcp://host:port, unix:///path/to/socket, fd://* or fd://socketfd.",
		},
	}
}

func PreCheckConfDebug(context *cli.Context) error {
	// debug为真，设置 DOCKERGSH_DEBUG 环境变量为 1
	if context.Bool("debug") {
		if err := os.Setenv(utils.DOCKERGSH_DEBUG, "1"); err != nil {
			return err
		}
	}
	return nil
}

func PreCheckConfHost(context *cli.Context) ([]string, error) {
	// hosts的作用 dockergsh 要连接的目的地址，也就是 dockergsh daemongsh 的地址
	hosts := context.StringSlice("hosts")
	if len(hosts) == 0 { // 如果没有传入 hosts 地址
		// 首先获取 DOCKERGSH_HOST 环境变量的值
		defaultHost := os.Getenv(utils.DOCKERGSH_CONFIG_HOST)
		if defaultHost == "" {
			defaultHost = fmt.Sprintf("unix://%s", utils.DEFAULTUNIXSOCKET)
		}

		hosts = append(hosts, defaultHost)
	}

	if len(hosts) > 1 {
		return nil, ErrMultiHosts
	}

	// 验证该 hosts 的合法性
	host, err := ValidateHost(hosts[0])

	if err != nil {
		return nil, err
	}

	if err := os.Setenv(utils.DOCKERGSH_SERVER_HOST, host); err != nil {
		return nil, err
	}

	protohost := strings.SplitN(host, "://", 2)
	return protohost, nil
}

func PreCheckConfTLS(context *cli.Context) (*tls.Config, error) {
	var tlsConfig tls.Config
	tlsConfig.InsecureSkipVerify = true
	// tlsConfig 对象需要加载一个受信的 ca 文件
	// 如果flTlsVerify为true，Docker Client连接Docker Server需要验证安全性
	if context.Bool("tls-verify") {
		// 读取 ca 证书
		filePath := context.String("tls-cacert")
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("Couldn't read ca cert %s: %s", filePath, err)
		}

		// 证书池是用于存储根证书和中间证书的集合，用于在验证证书链时提供验证所需的信任锚点。
		certPool := x509.NewCertPool()

		// 向证书池添加根证书
		certPool.AppendCertsFromPEM(file)

		tlsConfig.RootCAs = certPool
		tlsConfig.InsecureSkipVerify = false

		// 如果flTlsVerify有一个为真，那么需要加载证书发送给客户端。
		_, errCert := os.Stat(context.String("tls-cert"))
		_, errKey := os.Stat(context.String("tls-key"))
		if errCert == nil && errKey == nil {
			cert, err := tls.LoadX509KeyPair(context.String("tls-cert"), context.String("tls-key"))
			if err != nil {
				return nil, fmt.Errorf("Couldn't load X509 key pair: %s. Key encrypted?", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}
	return &tlsConfig, nil
}

func RootBefore(context *cli.Context) error {
	// 命令运行前的初始化 logrus 的日志配置
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(context.App.Writer)
	return nil
}

func ValidateHost(host string) (string, error) {
	parsedHost, err := parse.ParseHost(host, utils.DEFAULTHTTPHOST, utils.DEFAULTUNIXSOCKET)
	if err != nil {
		return host, err
	}
	return parsedHost, nil
}
