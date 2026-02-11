// Package cli provides CLI commands for k0rdentd
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/belgaied2/k0rdentd/internal/airgap/registry"
	"github.com/belgaied2/k0rdentd/pkg/utils"
	"github.com/urfave/cli/v2"
)

// RegistryCommand runs an OCI registry daemon for airgap installations
var RegistryCommand = &cli.Command{
	Name:      "registry",
	Usage:     "Run OCI registry daemon for airgap installations",
	UsageText: "k0rdentd registry [options]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Value:   "5000",
			Usage:   "Port for the registry server",
			EnvVars: []string{"K0RDENTD_REGISTRY_PORT"},
		},
		&cli.StringFlag{
			Name:    "host",
			Aliases: []string{"H"},
			Value:   "0.0.0.0",
			Usage:   "Host address to bind to (default: 0.0.0.0 for all interfaces)",
			EnvVars: []string{"K0RDENTD_REGISTRY_HOST"},
		},
		&cli.StringFlag{
			Name:    "storage",
			Aliases: []string{"s"},
			Value:   "/var/lib/k0rdentd/registry",
			Usage:   "Storage directory for registry data",
			EnvVars: []string{"K0RDENTD_REGISTRY_STORAGE"},
		},
		&cli.StringFlag{
			Name:    "bundle-path",
			Aliases: []string{"b"},
			Usage:   "Path to k0rdent airgap bundle (tar.gz or extracted directory)",
			EnvVars: []string{"K0RDENTD_AIRGAP_BUNDLE_PATH"},
			Required: true,
		},
		&cli.BoolFlag{
			Name:    "verify",
			Usage:   "Verify bundle signature with cosign",
			EnvVars: []string{"K0RDENTD_VERIFY_SIGNATURE"},
			Value:   true,
		},
		&cli.StringFlag{
			Name:    "cosignKey",
			Usage:   "Cosign public key URL or local path",
			Value:   "https://get.mirantis.com/cosign.pub",
			EnvVars: []string{"K0RDENTD_COSIGN_KEY"},
		},
		&cli.BoolFlag{
			Name:    "background",
			Aliases: []string{"d"},
			Usage:   "Run as daemon in background (not recommended, use systemd/supervise instead)",
			Value:   false,
		},
	},
	Action: registryAction,
}

func registryAction(c *cli.Context) error {
	logger := utils.GetLogger()

	port := c.String("port")
	host := c.String("host")
	storage := c.String("storage")
	bundlePath := c.String("bundle-path")
	verify := c.Bool("verify")
	cosignKey := c.String("cosignKey")
	background := c.Bool("background")

	// Validate bundle path exists
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		return fmt.Errorf("bundle not found at %s", bundlePath)
	}

	// Create registry daemon
	daemon := registry.NewRegistryDaemon(port, storage, bundlePath, verify, cosignKey)

	// Check if port is already in use
	if daemon.IsRunning() {
		return fmt.Errorf("registry port %s is already in use", port)
	}

	// Warn about background mode
	if background {
		logger.Warn("Running in background mode is not recommended")
		logger.Warn("Consider using systemd, supervisord, or similar for process management")
	}

	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handlers
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		logger.Infof("Received signal: %v", sig)
		cancel()
	}()

	// Start registry daemon
	logger.Info("Starting k0rdentd registry daemon...")
	logger.Infof("Configuration:")
	logger.Infof("  Bundle: %s", bundlePath)
	logger.Infof("  Storage: %s", storage)
	logger.Infof("  Address: %s:%s", host, port)
	logger.Infof("  Verify signature: %t", verify)

	if err := daemon.Start(ctx); err != nil {
		return fmt.Errorf("registry daemon failed: %w", err)
	}

	return nil
}
