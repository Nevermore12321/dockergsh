package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	RootCmd        = cli.NewApp()
	dockerCertPath = os.Getenv("DOCKERGSH_CERT_PATH")
)

const (
	defaultCaFile   = "ca.pem"
	defaultKeyFile  = "key.pem"
	defaultCertFile = "cert.pem"
)

var (
	ErrMultiHosts = errors.New("Please specify only one -H")
)

func RootCmdInitial(name string, in io.Reader, out, err io.Writer) {
	RootCmd.Name = name
	if in != nil {
		RootCmd.Reader = in
	}
	if out != nil {
		RootCmd.Writer = out
	}
	if err != nil {
		RootCmd.ErrWriter = err
	}

	// help 信息
	RootCmd.Usage = GetHelpUsage("")

	// 初始化子命令
	RootCmd.Commands = []*cli.Command{}

	// 初始化版本
	RootCmd.Version = VERSION

	// 初始化 RootCmd 的 flags
	RootCmd.Flags = []cli.Flag{
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
			Value: filepath.Join(dockerCertPath, defaultCaFile),
		},
		&cli.StringFlag{
			Name:  "tls-cert",
			Usage: "Path to TLS certificate file",
			Value: filepath.Join(dockerCertPath, defaultCertFile),
		},
		&cli.StringFlag{
			Name:  "tls-key",
			Usage: "Path to TLS key file",
			Value: filepath.Join(dockerCertPath, defaultKeyFile),
		},
		&cli.StringSliceFlag{
			Name:  "hosts",
			Usage: "The socket(s) to bind to in daemon mode\nspecified using one or more tcp://host:port, unix:///path/to/socket, fd://* or fd://socketfd.",
		},
	}

	RootCmd.Action = rootAction
	RootCmd.Before = rootBefore
	RootCmd.After = rootAfter
}

func rootAction(context *cli.Context) error {
	// debug为真，设置 DOCKERGSH_DEBUG 环境变量为 1
	if context.Bool("debug") {
		if err := os.Setenv(DOCKERGSH_DEBUG, "1"); err != nil {
			return err
		}
	}

	// hosts的作用 client 要连接的目的地址，也就是 dockergsh daemon 的地址
	hosts := context.StringSlice("hosts")
	if len(hosts) == 0 { // 如果没有传入 hosts 地址
		// 首先获取 DOCKERGSH_HOST 环境变量的值
		defaultHost := os.Getenv(DOCKERGSH_HOST)
		if defaultHost == "" {
			defaultHost = fmt.Sprintf("unix://%s", DEFAULTUNIXSOCKET)
		}
		hosts = append(hosts, defaultHost)
	}

	if len(hosts) > 1 {
		return ErrMultiHosts
	}

	// 验证该 hosts 的合法性
	host, err := ValidateHost(hosts[0])
	if err != nil {
		return err
	}

	if err = context.Set("dockergsh_host", host); err != nil {
		return err
	}

	protohost := strings.SplitN(host, "://", 2)

	var tlsConfig tls.Config
	tlsConfig.InsecureSkipVerify = true
	// todo tls config

	// 初始化 dockergshclient
	// 创建Docker Client实例。
	DockerGshCliInitial(context.App.Reader, context.App.Writer, context.App.ErrWriter, protohost[0], protohost[1], &tlsConfig)

	return nil
}

func rootBefore(context *cli.Context) error {
	// 命令运行前的初始化 logrus 的日志配置
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(context.App.Writer)
	return nil
}

func rootAfter(context *cli.Context) error {
	err := context.Err()
	logrus.Info(err)
	// todo http status error
	//if err != nil {
	//	if sterr, ok := err.(*StatusError); ok {
	//		if sterr.Status != "" {
	//			log.Println(sterr.Status)
	//		}
	//		os.Exit(sterr.StatusCode)
	//	}
	//}
}
