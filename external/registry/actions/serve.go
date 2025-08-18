package actions

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/external/registry/libs/dcontext"
	"github.com/Nevermore12321/dockergsh/external/registry/version"
	"github.com/urfave/cli/v2"
)

func CmdServeDescription() string {
	return "`serve` stores and distributes Docker image."
}

func CmdRegistryServe(context *cli.Context) error {
	// 创建 context
	dcontext.WithVersion(dcontext.Background(), version.Version())

	if context.NArg() < 1 {
		return fmt.Errorf("Error: Missing configuration file parameters\nUsage: serve <config>")
	}
	filePath := context.Args().Get(0)

	config, err := resolveConfiguration(filePath)
	if err != nil {
		return fmt.Errorf("configuration error: %v\n", err)
	}

}
