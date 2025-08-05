package actions

import (
	"github.com/Nevermore12321/dockergsh/external/registry/libs/dcontext"
	"github.com/Nevermore12321/dockergsh/external/registry/version"
	"github.com/urfave/cli/v2"
)

func CmdServeDescription() string {
	return "`serve` stores and distributes Docker image."
}

func CmdRegisgtryServe(context *cli.Context) error {
	// 创建 context
	dcontext.WithVersion(dcontext.Background(), version.Version())

}
