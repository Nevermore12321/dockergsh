package root

import (
	"fmt"
	clientActions "github.com/Nevermore12321/dockergsh/internal/client/actions"
	daemonActions "github.com/Nevermore12321/dockergsh/internal/daemongsh/actions"
	"github.com/urfave/cli/v2"
)

func initSubCmd(root *cli.App) {
	root.Commands = cli.Commands{
		&cli.Command{
			Name:   "version",
			Usage:  GetHelpUsage("version"),
			Action: CmdShowVersion,
		},
		&cli.Command{
			Name:   "client",
			Usage:  GetHelpUsage("client"),
			Before: clientActions.CmdClientInitial,
			Subcommands: cli.Commands{
				{
					Name:        "pull",
					Usage:       GetHelpUsage("client"),
					Action:      clientActions.CmdPull,
					Description: clientActions.CmdPullDescription(),
				},
				{
					Name:        "login",
					Usage:       GetHelpUsage("client"),
					Flags:       clientActions.CmdLoginFlags(),
					Action:      clientActions.CmdLogin,
					Description: clientActions.CmdLoginDescription(),
				},
			},
		},
		&cli.Command{
			Name:  "daemon",
			Usage: GetHelpUsage("daemon"),
			Flags: cmdDaemonFlags(),
			Subcommands: cli.Commands{
				{
					Name:   "start",
					Usage:  GetHelpUsage("daemon_start"),
					Action: daemonActions.CmdStart,
				},
			},
		},
	}
}

func cmdDaemonFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "pidfile",
			Aliases: []string{"p"},
			Value:   "/var/run/dockergsh.pid",
			Usage:   "Path to use for daemon PID file",
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
			Usage:   "--restart on the daemon has been deprecated infavor of --restart policies on dockergsh run",
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
		&cli.StringFlag{
			Name:  "ip",
			Value: "0.0.0.0",
			Usage: "Default IP address to use when binding container port",
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

func CmdShowVersion(context *cli.Context) error {
	fmt.Printf("Docker version %s, build %s\n", VERSION, GITCOMMIT)
	return nil
}
