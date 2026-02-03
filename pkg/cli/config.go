package cli

import (
	"fmt"

	"github.com/belgaied2/k0rdentd/pkg/config"
	"github.com/belgaied2/k0rdentd/pkg/utils"
	"github.com/urfave/cli/v2"
)

var ConfigCommand = &cli.Command{
	Name:      "config",
	Aliases:   []string{"c"},
	Usage:     "Manage configuration",
	UsageText: "k0rdentd config [subcommand] [options]",
	Subcommands: []*cli.Command{
		{
			Name:   "validate",
			Usage:  "Validate configuration file",
			Action: validateConfigAction,
		},
		{
			Name:   "show",
			Usage:  "Show current configuration",
			Action: showConfigAction,
		},
		{
			Name:   "init",
			Usage:  "Create default configuration file",
			Action: initConfigAction,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
					Value:   "/etc/k0rdentd/k0rdentd.yaml",
					Usage:   "Output file path",
				},
			},
		},
	},
}

func validateConfigAction(c *cli.Context) error {
	_, err := config.LoadConfigWithFallback(
		c.String("config-file"),
		"/etc/k0rdentd/k0rdentd.yaml",
		c.IsSet("config-file"),
	)
	if err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}
	utils.GetLogger().Info("✅ Configuration is valid!")
	return nil
}

func showConfigAction(c *cli.Context) error {
	cfg, err := config.LoadConfigWithFallback(
		c.String("config-file"),
		"/etc/k0rdentd/k0rdentd.yaml",
		c.IsSet("config-file"),
	)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	configData, err := config.MarshalConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	utils.GetLogger().Info(string(configData))
	return nil
}

func initConfigAction(c *cli.Context) error {
	cfg := config.DefaultConfig()
	configData, err := config.MarshalConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	outputPath := c.String("output")
	if err := config.WriteConfigFile(outputPath, configData); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	utils.GetLogger().Infof("✅ Default configuration written to %s", outputPath)
	return nil
}
