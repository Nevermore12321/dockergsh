package command

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/network"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var NetworkCommand = &cli.Command{
	Name:  "network",
	Usage: "Manage networks",
	Subcommands: []*cli.Command{{
		Name:  "create",
		Usage: "Create a network",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "driver",
				Usage: "Driver to manage the Network (default \"bridge\")",
			},
			&cli.StringFlag{
				Name:  "subnet",
				Usage: "Subnet in CIDR format that represents a network segment",
			},
		},
		Action: func(context *cli.Context) error {
			if context.NArg() < 1 {
				return fmt.Errorf("missing network name")
			}
			if err := network.Init(); err != nil {
				return err
			}
			// 创建网络
			err := network.CreateNetwork(context.String("driver"), context.String("subnet"), context.Args().Get(0))
			if err != nil {
				log.Errorf("commit container err: %v", err)
				return err
			}
			return nil
		},
	},
		{
			Name:  "list",
			Usage: "List networks",
			Action: func(context *cli.Context) error {
				if err := network.Init(); err != nil {
					return err
				}
				if err := network.ListNetwork(); err != nil {
					return err
				}
				return nil
			},
		}},
}
