package cli

import (
	"fmt"

	"github.com/belgaied2/k0rdentd/pkg/installer"
	"github.com/belgaied2/k0rdentd/pkg/k0s"
	"github.com/belgaied2/k0rdentd/pkg/utils"
	"github.com/urfave/cli/v2"
)

var UninstallCommand = &cli.Command{
	Name:      "uninstall",
	Aliases:   []string{"u"},
	Usage:     "Uninstall K0s and K0rdent",
	UsageText: "k0rdentd uninstall [options]",
	Action:    uninstallAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage:   "Force uninstall without confirmation",
		},
	},
}

func uninstallAction(c *cli.Context) error {

	// Check if k0s binary exists
	k0sCheck, err := k0s.CheckK0s()
	if err != nil {
		return fmt.Errorf("failed to check k0s: %w", err)
	}

	// If k0s is not installed, uninstall cannot proceed
	if !k0sCheck.Installed {
		return fmt.Errorf("k0s binary not found, cannot uninstall")
	}

	if !c.Bool("force") {
		utils.GetLogger().Info("🚨 This will uninstall K0s and K0rdent from this system.")
		utils.GetLogger().Info("Are you sure you want to continue? (y/N): ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil || (response != "y" && response != "Y") {
			utils.GetLogger().Info("Cancelled.")
			return nil
		}
	}

	installer := installer.NewInstaller(
		c.Bool("debug"),
		c.Bool("dry-run"),
	)

	if err := installer.Uninstall(); err != nil {
		return fmt.Errorf("uninstallation failed: %w", err)
	}

	utils.GetLogger().Info("✅ K0s and K0rdent uninstalled successfully!")
	return nil
}
