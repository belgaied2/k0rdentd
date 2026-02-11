package main

import (
	"os"

	"github.com/belgaied2/k0rdentd/pkg/cli"
	"github.com/belgaied2/k0rdentd/pkg/utils"
	urfavecli "github.com/urfave/cli/v2"
)

func main() {
	app := &urfavecli.App{
		Name:                 "k0rdentd",
		Usage:                "Deploy K0s and K0rdent on the VM it runs on",
		Version:              cli.Version,
		EnableBashCompletion: true,
		Commands: []*urfavecli.Command{
			cli.InstallCommand,
			cli.UninstallCommand,
			cli.RegistryCommand,
			cli.VersionCommand,
			cli.ConfigCommand,
			cli.ExposeUICommand,
			cli.ExportWorkerArtifactsCommand,
			cli.ShowFlavorCommand,
		},
		Flags: []urfavecli.Flag{
			&urfavecli.StringFlag{
				Name:    "config-file",
				Aliases: []string{"c"},
				Value:   "/etc/k0rdentd/k0rdentd.yaml",
				Usage:   "Path to k0rdentd configuration file",
				EnvVars: []string{"K0RDENTD_CONFIG_FILE"},
			},
			&urfavecli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Value:   false,
				Usage:   "Enable debug logging",
				EnvVars: []string{"K0RDENTD_DEBUG"},
			},
			&urfavecli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Value:   false,
				Usage:   "Show what would be done without making changes",
				EnvVars: []string{"K0RDENTD_DRY_RUN"},
			},
		},
	}

	// Initialize logging based on --debug flag
	utils.SetupLogging(app.Flags[1].(*urfavecli.BoolFlag).Value)

	if err := app.Run(os.Args); err != nil {
		utils.GetLogger().Fatal(err)
	}
}
