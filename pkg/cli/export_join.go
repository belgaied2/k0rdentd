package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/belgaied2/k0rdentd/internal/airgap"
	"github.com/belgaied2/k0rdentd/pkg/config"
	checker "github.com/belgaied2/k0rdentd/pkg/k0s"
	"github.com/belgaied2/k0rdentd/pkg/network"
	"github.com/belgaied2/k0rdentd/pkg/token"
	"github.com/belgaied2/k0rdentd/pkg/utils"
	"github.com/urfave/cli/v2"
)

var ExportJoinConfigCommand = &cli.Command{
	Name:      "export-join-config",
	Usage:     "Export join configurations for additional nodes",
	UsageText: "k0rdentd export-join-config [options]",
	Action:    exportJoinConfigAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Value:   "./join-configs",
			Usage:   "Output directory for join config files",
		},
		&cli.StringFlag{
			Name:  "controller-ip",
			Usage: "Override auto-detected controller IP address",
		},
		&cli.DurationFlag{
			Name:    "expiry",
			Aliases: []string{"e"},
			Value:   24 * time.Hour,
			Usage:   "Token expiry time (e.g., 24h, 168h for 7 days)",
		},
		&cli.IntFlag{
			Name:  "registry-port",
			Value: 5000,
			Usage: "Registry port for airgap mode",
		},
		&cli.BoolFlag{
			Name:    "overwrite",
			Aliases: []string{"f"},
			Usage:   "Overwrite existing files",
		},
	},
}

func exportJoinConfigAction(c *cli.Context) error {
	logger := utils.GetLogger()

	// Load current config to get k0s version and airgap settings
	cfg, err := config.LoadConfigWithFallback(
		c.String("config-file"),
		"/etc/k0rdentd/k0rdentd.yaml",
		c.IsSet("config-file"),
	)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get controller IP
	controllerIP, err := network.GetControllerIP(c.String("controller-ip"))
	if err != nil {
		return fmt.Errorf("failed to get controller IP: %w", err)
	}
	logger.Infof("Using controller IP: %s", controllerIP)

	// Create token manager
	tokenManager := token.NewManager("/usr/local/bin/k0s", c.Bool("debug"))

	// Create tokens
	logger.Info("Creating controller join token...")
	controllerToken, err := tokenManager.CreateControllerToken(c.Duration("expiry"))
	if err != nil {
		return fmt.Errorf("failed to create controller token: %w", err)
	}

	logger.Info("Creating worker join token...")
	workerToken, err := tokenManager.CreateWorkerToken(c.Duration("expiry"))
	if err != nil {
		return fmt.Errorf("failed to create worker token: %w", err)
	}

	// Create output directory with 0777 permissions for easy user access
	outputDir := c.String("output")
	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate controller join config
	controllerConfig := generateJoinConfig(cfg, "controller", controllerIP, controllerToken, c.Int("registry-port"))
	controllerPath := filepath.Join(outputDir, "controller-join.yaml")

	if err := writeJoinConfig(controllerPath, controllerConfig, c.Bool("overwrite")); err != nil {
		return fmt.Errorf("failed to write controller join config: %w", err)
	}
	logger.Infof("✅ Created: %s", controllerPath)

	// Generate worker join config
	workerConfig := generateJoinConfig(cfg, "worker", controllerIP, workerToken, c.Int("registry-port"))
	workerPath := filepath.Join(outputDir, "worker-join.yaml")

	if err := writeJoinConfig(workerPath, workerConfig, c.Bool("overwrite")); err != nil {
		return fmt.Errorf("failed to write worker join config: %w", err)
	}
	logger.Infof("✅ Created: %s", workerPath)

	logger.Info("")
	logger.Info("Join configuration files created successfully!")
	logger.Info("")
	logger.Info("To add a controller node:")
	logger.Infof("  1. Copy %s to the new node: /etc/k0rdentd/k0rdentd.yaml", controllerPath)
	logger.Info("  2. Run: k0rdentd install")
	logger.Info("")
	logger.Info("To add a worker node:")
	logger.Infof("  1. Copy %s to the new node: /etc/k0rdentd/k0rdentd.yaml", workerPath)
	logger.Info("  2. Run: k0rdentd install")

	return nil
}

// generateJoinConfig creates a join configuration file content
func generateJoinConfig(baseCfg *config.K0rdentdConfig, mode, controllerIP, token string, registryPort int) *config.K0rdentdConfig {
	version, err := checker.GetK0sVersion()
	if err != nil {
		version = baseCfg.K0s.Version
	}
	cfg := &config.K0rdentdConfig{
		K0s: config.K0sConfig{
			Version: version,
		},
		Join: config.JoinConfig{
			Mode:   mode,
			Server: controllerIP,
			Token:  token,
		},
	}

	// Only include airgap settings if running in airgap mode
	if airgap.IsAirGap() {
		if baseCfg.Airgap.Registry.Address != "" {
			cfg.Airgap = config.AirgapConfig{
				Registry: baseCfg.Airgap.Registry,
			}
		} else if baseCfg.Airgap.BundlePath != "" {
			// Generate registry address from controller IP
			cfg.Airgap = config.AirgapConfig{
				Registry: config.RegistryConfig{
					Address:  fmt.Sprintf("%s:%d", controllerIP, registryPort),
					Insecure: true,
				},
			}
		}
	}

	return cfg
}

// writeJoinConfig writes a join configuration to file
func writeJoinConfig(path string, cfg *config.K0rdentdConfig, overwrite bool) error {
	// Check if file exists
	if _, err := os.Stat(path); err == nil && !overwrite {
		return fmt.Errorf("file %s already exists (use --overwrite to replace)", path)
	}

	data, err := config.MarshalConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := fmt.Sprintf("# K0rdentd Join Configuration\n# Generated: %s\n# Copy this file to /etc/k0rdentd/k0rdentd.yaml on the joining node\n\n", time.Now().Format(time.RFC3339))
	data = append([]byte(header), data...)

	// Write file with 0666 permissions for easy user access
	if err := os.WriteFile(path, data, 0666); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
