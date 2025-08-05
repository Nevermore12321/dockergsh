package root

import (
	"fmt"
	regsitryActions "github.com/Nevermore12321/dockergsh/external/registry/actions"
	"github.com/Nevermore12321/dockergsh/external/registry/version"
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
					Usage:       GetHelpUsage("pull"),
					Action:      clientActions.CmdPull,
					Description: clientActions.CmdPullDescription(),
				},
				{
					Name:        "login",
					Usage:       GetHelpUsage("login"),
					Action:      clientActions.CmdLogin,
					Description: clientActions.CmdLoginDescription(),
				},
				{
					Name:        "logout",
					Usage:       GetHelpUsage("logout"),
					Action:      clientActions.CmdLogout,
					Description: clientActions.CmdLogoutDescription(),
				},
				{
					Name:        "run",
					Usage:       GetHelpUsage("run"),
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
		&cli.Command{
			Name:        "registry",
			Usage:       GetHelpUsage("registry"),
			Flags:       cmdRegistryFlag(),
			Description: "Start a dockergsh registry server for saving docker images",
			Action: func(context *cli.Context) error {
				showVersion := context.Bool("version")
				if showVersion {
					version.PrintVersion()
					return nil
				}
				GetHelpUsage("registry")
			},
			Subcommands: cli.Commands{
				{
					Name:        "serve",
					Usage:       GetHelpUsage("serve"),
					Action:      regsitryActions.CmdRegisgtryServe,
					Description: regsitryActions.CmdServeDescription(),
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

func cmdRegistryFlag() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "dry-run",
			Value: false,
			Usage: "do everything except remove the blobs",
		},
		&cli.BoolFlag{
			Name:  "delete-untagged",
			Value: false,
			Usage: "delete manifests that are not currently referenced via tag",
		},
		&cli.BoolFlag{
			Name:  "quiet",
			Value: false,
			Usage: "silence output",
		},
		&cli.BoolFlag{
			Name:  "version",
			Value: false,
			Usage: "show the version and exit",
		},
	}
}

func CmdShowVersion(context *cli.Context) error {
	fmt.Printf("Docker version %s, build %s\n", VERSION, GITCOMMIT)
	return nil
}
