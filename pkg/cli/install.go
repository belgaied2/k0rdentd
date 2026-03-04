package cli

import (
	"fmt"

	"github.com/belgaied2/k0rdentd/internal/airgap"
	"github.com/belgaied2/k0rdentd/pkg/config"
	"github.com/belgaied2/k0rdentd/pkg/generator"
	"github.com/belgaied2/k0rdentd/pkg/installer"
	"github.com/belgaied2/k0rdentd/pkg/k0s"
	"github.com/belgaied2/k0rdentd/pkg/ui"
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
		&cli.BoolFlag{
			Name:  "join",
			Usage: "Join an existing cluster (requires --mode)",
		},
		&cli.StringFlag{
			Name:  "mode",
			Usage: "Node mode: controller or worker (required if --join is set)",
		},
		&cli.BoolFlag{
			Name:    "replace-k0s",
			Aliases: []string{"R"},
			Usage:   "Replace existing k0s binary without prompting (only if not running)",
			EnvVars: []string{"K0RDENTD_REPLACE_K0S"},
		},
	},
}

func installAction(c *cli.Context) error {
	logger := utils.GetLogger()

	// Load configuration with fallback logic
	cfg, err := config.LoadConfigWithFallback(
		c.String("config-file"),
		"/etc/k0rdentd/k0rdentd.yaml",
		c.IsSet("config-file"),
	)
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

	// Determine if we're joining a cluster
	// Priority: CLI flags > config file
	joinMode := ""
	if c.IsSet("join") || c.IsSet("mode") {
		// CLI flags provided
		if c.IsSet("join") && !c.IsSet("mode") {
			return fmt.Errorf("--mode is required when --join is specified")
		}
		if c.IsSet("mode") && !c.IsSet("join") && !cfg.Join.IsJoin() {
			return fmt.Errorf("--join is required when --mode is specified without join config in file")
		}
		joinMode = c.String("mode")
	} else if cfg.Join.IsJoin() {
		// Join config from file
		joinMode = cfg.Join.Mode
		logger.Infof("Using join configuration from file (mode: %s)", joinMode)
	}

	// Validate join mode
	if joinMode != "" && joinMode != "controller" && joinMode != "worker" {
		return fmt.Errorf("invalid mode '%s': must be 'controller' or 'worker'", joinMode)
	}

	// Check if k0s binary exists
	k0sCheck, err := k0s.CheckK0s()
	if err != nil {
		return fmt.Errorf("failed to check k0s: %w", err)
	}

	// If k0s is not installed, install it (for both init and join modes)
	// In airgap mode, k0s binary should already be extracted from embedded assets
	if !airgap.IsAirGap() && !k0sCheck.Installed {
		if cfg.K0s.Version != "" {
			// Install specific version if configured
			logger.Infof("k0s binary not found, installing version %s...", cfg.K0s.Version)
			if err := k0s.InstallK0sVersion(cfg.K0s.Version); err != nil {
				return fmt.Errorf("failed to install k0s version %s: %w", cfg.K0s.Version, err)
			}
		} else {
			// Install latest version
			logger.Info("k0s binary not found, installing latest version...")
			if err := k0s.InstallK0s(); err != nil {
				return fmt.Errorf("failed to install k0s: %w", err)
			}
		}
	}

	// Create installer
	inst := installer.NewInstaller(
		c.Bool("debug"),
		c.Bool("dry-run"),
	)
	inst.SetConfig(cfg)

	// Set replace-k0s flag if specified
	if c.IsSet("replace-k0s") {
		inst.SetReplaceK0s(c.Bool("replace-k0s"))
	}

	// Execute installation based on mode
	if joinMode != "" {
		// Join existing cluster
		if err := inst.InstallJoin(&cfg.Join); err != nil {
			return fmt.Errorf("join installation failed: %w", err)
		}
		logger.Info("✅ Successfully joined the cluster!")
	} else {
		// Initialize new cluster (cluster-init is implicit)
		k0sConfig, err := generator.GenerateK0sConfig(cfg)
		if err != nil {
			return fmt.Errorf("failed to generate K0s config: %w", err)
		}

		if err := inst.Install(k0sConfig, &cfg.K0rdent); err != nil {
			return fmt.Errorf("installation failed: %w", err)
		}
		logger.Info("✅ K0s and K0rdent installed successfully!")

		// Expose k0rdent UI (only for controller init mode)
		if err := ui.ExposeUI(); err != nil {
			logger.Warnf("Failed to expose k0rdent UI: %v", err)
		}
	}

	return nil
}
