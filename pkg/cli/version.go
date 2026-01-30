package cli

import (
	"github.com/belgaied2/k0rdentd/pkg/utils"
	"github.com/urfave/cli/v2"
)

// Version is set at build time via ldflags
var Version = "dev"

var VersionCommand = &cli.Command{
	Name:      "version",
	Aliases:   []string{"v"},
	Usage:     "Show version information",
	UsageText: "k0rdentd version",
	Action:    versionAction,
}

func versionAction(c *cli.Context) error {
	utils.GetLogger().Infof("k0rdentd version %s", Version)
	utils.GetLogger().Info("A CLI tool to deploy K0s and K0rdent")
	utils.GetLogger().Info("Copyright © 2024 belgaied2")
	return nil
}
