package root

import (
	"errors"
	"fmt"
	"github.com/Nevermore12321/dockergsh/internal/utils"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"strings"
)

var (
	VERSION   string
	GITCOMMIT string
)

var (
	ErrMultiHosts     = errors.New("please specify only one -H")
	ErrNoCorrectHosts = errors.New("please specify correct hosts")
)

func cmdCommonFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"D"},
			Usage:   "Enable debug mode",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:  "api-enable-cors",
			Usage: "Enable CORS headers in the remote API",
			Value: false,
		},
		&cli.StringFlag{
			Name:    "socket-group",
			Aliases: []string{"G"},
			Usage:   "Group to assign the unix socket specified by -H when running in daemon mode\nuse '' (the empty string) to disable setting of a group",
			Value:   "docker",
		},
		&cli.BoolFlag{
			Name:  "tls",
			Usage: "Use TLS; implied by tls-verify flags",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "tls-verify",
			Usage: "Use TLS and verify the remote",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "tls-cacert",
			Usage: "Trust only remotes providing a certificate signed by the CA given here",
			Value: filepath.Join(utils.DockergshCertPath, utils.DefaultCaFile),
		},
		&cli.StringFlag{
			Name:  "tls-cert",
			Usage: "Path to TLS certificate file",
			Value: filepath.Join(utils.DockergshCertPath, utils.DefaultCertFile),
		},
		&cli.StringFlag{
			Name:  "tls-key",
			Usage: "Path to TLS key file",
			Value: filepath.Join(utils.DockergshCertPath, utils.DefaultKeyFile),
		},
		&cli.StringSliceFlag{
			Name:  "host",
			Usage: "The socket(s) to bind to in daemon mode\nspecified using one or more tcp://host:port, unix:///path/to/socket, fd://* or fd://socketfd.\n Or, the address to connection to daemon(only one addr).",
		},
	}
}

func PreCheckConfDebug(context *cli.Context) error {
	// debug为真，设置 DOCKERGSH_DEBUG 环境变量为 1
	if context.Bool("debug") {
		if err := os.Setenv(utils.DockergshDebug, "1"); err != nil {
			return err
		}
	}
	return nil
}

func PreCheckConfHost(context *cli.Context) error {
	// hosts的作用 dockergsh 要连接的目的地址，也就是 dockergsh daemon 的地址
	hosts := context.StringSlice("hosts")
	isDaemon := context.Args().Get(0) == "daemon"
	isClient := context.Args().Get(0) == "client"

	if len(hosts) == 0 { // 如果没有传入 hosts 地址
		// 首先获取 DOCKERGSH_HOST 环境变量的值
		defaultHost := os.Getenv(utils.DockergshConfigHost)

		// 若 defaultHost 为空或者 isDaemon 为真，说明目前还没有一个定义的 host对象，
		// 则将其默认设置为 unix socket ，值为 utils.DefaultUnixSocket 即 "/var/run/docker.sock"
		if defaultHost == "" || isDaemon {
			defaultHost = fmt.Sprintf("unix://%s", utils.DefaultUnixSocket)
		}

		// 验证该 defaultHost 的合法性之后，将 defaultHost 的值追加至 hosts 的末尾， 继续往下执行。
		hosts = append(hosts, defaultHost)
	}

	// 验证该 hosts 的合法性
	parsedHosts := utils.Validates(hosts, utils.ValidateHost)
	if parsedHosts == nil {
		return ErrNoCorrectHosts
	}

	if len(parsedHosts) > 1 && isClient { // dockergsh client 只能有一个 host
		return ErrMultiHosts
	}

	// 传递给后续 dockergsh 子命令
	return os.Setenv(utils.DockergshHosts, strings.Join(parsedHosts, ","))
}

func GetHelpUsage(method string) string {
	var usages = map[string]string{
		"attach":  "Attach to a running container",
		"build":   "Build an image from a Dockerfile",
		"commit":  "Create a new image from a container's changes",
		"cp":      "Copy files/folders from a container's filesystem to the host path",
		"diff":    "Insp:ct changes on a container's filesystem",
		"events":  "Get real time events from the server",
		"export":  "Stream the contents of a container as a tar archive",
		"history": "Show the history of an image",
		"images":  "List images",
		"import":  "Create a new filesystem image from the contents of a tarball",
		"info":    "Display system-wide information",
		"inspect": "Return low-level information on a container",
		"kill":    "Kill a running container",
		"load":    "Load an image from a tar archive",
		"login":   "Register or log in to a Docker registry server",
		"logout":  "Log out from a Docker registry server",
		"logs":    "Fetch the logs of a container",
		"port":    "Lookup the public-facing port that is NAT-ed to PRIVATE_PORT",
		"pause":   "Pause all processes within a container",
		"ps":      "List containers",
		"pull":    "Pull an image or a repository from a Docker registry server",
		"push":    "Push an image or a repository to a Docker registry server",
		"restart": "Restart a running container",
		"rm":      "Remove one or more containers",
		"rmi":     "Remove one or more images",
		"run":     "Run a command in a new container",
		"save":    "Save an image to a tar archive",
		"search":  "Search for an image on the Docker Hub",
		"start":   "Start a stopped container",
		"stop":    "Stop a running container",
		"tag":     "Tag an image into a repository",
		"top":     "Lookup the running processes of a container",
		"unpause": "Unpause a paused container",
		"version": "Show the Docker version information",
		"wait":    "Block until a container stops, then print its exit code",
		"daemon":  "Start daemongsh service",
		"client":  "Dockergsh client to use subcommands to communicate to daemon",
	}
	if method != "" {
		cmdHelp, exist := usages[method]
		if exist {
			return cmdHelp
		}
	}
	help := fmt.Sprintf("dockergsh [OPTIONS] COMMAND [arg...]\n -H=[unix://%s]: tcp://host:port to bind/connect to or unix://path/to/socket to use\n\nA self-sufficient runtime for linux containers.\n\nCommands:\n", utils.DefaultUnixSocket)
	return help
}
