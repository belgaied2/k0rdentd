package cli

import (
	"fmt"

	"github.com/belgaied2/k0rdentd/pkg/config"
	"github.com/belgaied2/k0rdentd/pkg/generator"
	"github.com/belgaied2/k0rdentd/pkg/installer"
	"github.com/belgaied2/k0rdentd/pkg/k0s"
	"github.com/belgaied2/k0rdentd/pkg/utils"
	"github.com/urfave/cli/v2"
)

var InstallCommand = &cli.Command{
	Name:      "install",
	Aliases:   []string{"i"},
	Usage:     "Install K0s and K0rdent",
	UsageText: "k0rdentd install [options]",
	Action:    installAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "k0s-version",
			Aliases: []string{"k"},
			Usage:   "Override K0s version from config",
		},
		&cli.StringFlag{
			Name:    "k0rdent-version",
			Aliases: []string{"r"},
			Usage:   "Override K0rdent version from config",
		},
	},
}

func installAction(c *cli.Context) error {
	// Load configuration
	cfg, err := config.LoadConfig(c.String("config-file"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Apply CLI overrides
	if c.IsSet("k0s-version") {
		cfg.K0s.Version = c.String("k0s-version")
	}
	if c.IsSet("k0rdent-version") {
		cfg.K0rdent.Version = c.String("k0rdent-version")
	}

	// Check if k0s binary exists
	k0sCheck, err := k0s.CheckK0s(cfg)
	if err != nil {
		return fmt.Errorf("failed to check k0s: %w", err)
	}

	// If k0s is not installed, install it
	if !k0sCheck.Installed {
		utils.GetLogger().Info("k0s binary not found, installing...")
		if err := k0s.InstallK0s(); err != nil {
			return fmt.Errorf("failed to install k0s: %w", err)
		}
	}

	// Generate K0s configuration
	k0sConfig, err := generator.GenerateK0sConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate K0s config: %w", err)
	}

	// Install K0s and K0rdent
	installer := installer.NewInstaller(
		c.Bool("debug"),
		c.Bool("dry-run"),
	)

	if err := installer.Install(k0sConfig); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Println("✅ K0s and K0rdent installed successfully!")
	return nil
}
