package cli

import (
	"github.com/belgaied2/k0rdentd/internal/airgap"
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
	logger := utils.GetLogger()

	metadata := airgap.GetBuildMetadata()
	logger.Infof("k0rdentd version %s", Version)
	logger.Infof("Build flavor: %s", metadata.Flavor)

	if metadata.Flavor == "airgap" {
		logger.Infof("K0s version: %s", metadata.K0sVersion)
		logger.Infof("K0rdent version: %s", metadata.K0rdentVersion)
	}

	logger.Info("A CLI tool to deploy K0s and K0rdent")
	logger.Info("Copyright © 2024 belgaied2")

	return nil
}
