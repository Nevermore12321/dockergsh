package daemon

import (
	"github.com/Nevermore12321/dockergsh/client"
	service "github.com/Nevermore12321/dockergsh/cmd"
	"github.com/Nevermore12321/dockergsh/daemongsh/daemon"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
)

var (
	RootCmd   = cli.NewApp()
	daemonCfg = &daemon.Config{}
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
	RootCmd.Usage = client.GetHelpUsage("")

	// 初始化版本
	RootCmd.Version = client.VERSION

	// 初始化 RootCmd 的 flags
	// 添加 daemongsh 的 flags
	RootCmd.Flags = append(service.CmdFlags(), daemongshFlags()...)

	RootCmd.Action = rootAction
	RootCmd.Before = service.RootBefore
	RootCmd.After = rootAfter
}

func rootAction(context *cli.Context) error {
	if err := service.PreCheckConfDebug(context); err != nil {
		return err
	}

	protohost, err := service.PreCheckConfHost(context)
	if err != nil {
		return err
	}

	tlsConfig, err := service.PreCheckConfTLS(context)
	if err != nil {
		return err
	}

	// todo delete docker client, add maindaemongsh
	// 初始化 dockergshclient
	// 创建Docker Client实例。
	client.DockerGshCliInitial(context.App.Reader, context.App.Writer, context.App.ErrWriter, protohost[0], protohost[1], tlsConfig)

	return nil
}

func rootAfter(context *cli.Context) error {
	err := context.Err()
	if err != nil {
		logrus.Info(err)
	}
	// todo http status error
	//if err != nil {
	//	if sterr, ok := err.(*StatusError); ok {
	//		if sterr.Status != "" {
	//			log.Println(sterr.Status)
	//		}
	//		os.Exit(sterr.StatusCode)
	//	}
	//}
	return nil
}

func daemongshFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "pidfile",
			Aliases: []string{"p"},
			Value:   "/var/run/dockergsh.pid",
			Usage:   "Path to use for daemongsh PID file",
		},
		&cli.StringFlag{
			Name:    "graph",
			Aliases: []string{"g"},
			Value:   "/var/lib/dockergsh",
			Usage:   "Path to use as the root of the Dockergsh runtime",
		},
		&cli.BoolFlag{
			Name:    "restart",
			Aliases: []string{"r"},
			Value:   true,
			Usage:   "--restart on the daemongsh has been deprecated infavor of --restart policies on dockergsh run",
		},
		&cli.StringSliceFlag{
			Name:    "dns",
			Aliases: []string{"d"},
			Usage:   "Force Dockergsh to use specific DNS servers",
		},
		&cli.StringSliceFlag{
			Name:  "dns-search",
			Usage: "Force Dockergsh to use specific DNS search domains",
		},
		&cli.BoolFlag{
			Name:  "iptables",
			Value: true,
			Usage: "Enable Dockergsh's addition of iptables rules",
		},
		&cli.BoolFlag{
			Name:  "ip-forward",
			Value: true,
			Usage: "Enable net.ipv4.ip_forward",
		},
		&cli.StringSliceFlag{
			Name:  "ip",
			Usage: "Default IP address to use when binding container ports",
		},
		&cli.StringFlag{
			Name:    "bridge-ip",
			Aliases: []string{"bip"},
			Usage:   "Use this CIDR notation address for the network bridge's IP, not compatible with -b",
		},
		&cli.StringFlag{
			Name:    "bridge",
			Aliases: []string{"b"},
			Usage:   "Attach containers to a pre-existing network bridge\nuse 'none' to disable container networking",
		},
		&cli.BoolFlag{
			Name:    "inter-container-communication",
			Aliases: []string{"icc"},
			Value:   true,
			Usage:   "Enable inter-container communication",
		},
		&cli.StringFlag{
			Name:    "storage-driver",
			Aliases: []string{"s"},
			Usage:   "Force the Dockergsh runtime to use a specific storage driver",
		},
		&cli.StringSliceFlag{
			Name:  "storage-opts",
			Usage: "Set storage driver options",
		},
		&cli.StringFlag{
			Name:    "exec-driver",
			Aliases: []string{"e"},
			Value:   "native",
			Usage:   "Force the Dockergsh runtime to use a specific exec driver",
		},
		&cli.IntFlag{
			Name:  "mtu",
			Value: 0,
			Usage: "Set the containers network MTU\nif no value is provided: default to the default route MTU or 1500 if no default route is available",
		},
		&cli.BoolFlag{
			Name:    "selinux-enabled",
			Aliases: []string{"se"},
			Value:   false,
			Usage:   "Enable selinux support. SELinux does not presently support the BTRFS storage driver",
		},
	}
}

// daemongsh 的启动流程
func mainDaemon(context *cli.Context) {
	// 1.daemon的配置初始化
	daemonCfg.InitialFlags(context)

	//2. 命令行flag参数个数检查。
	if context.NArg() != 0 {
		return
	}
	//3. 创建engine对象。

	//4）设置engine的信号捕获及处理方法。
	//5）加载builtins。
	//6）使用goroutine加载daemon对象并运行。
	//7）打印Docker版本及驱动信息。
	//8）serveapi的创建与运行。
}
