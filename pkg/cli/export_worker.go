package cli

import (
	"fmt"

	"github.com/belgaied2/k0rdentd/internal/airgap"
	"github.com/belgaied2/k0rdentd/pkg/utils"
	"github.com/urfave/cli/v2"
)

var ExportWorkerArtifactsCommand = &cli.Command{
	Name:      "export-worker-artifacts",
	Usage:     "Export worker artifacts for multi-worker airgap installations",
	UsageText: "k0rdentd export-worker-artifacts [options]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "output-dir",
			Aliases: []string{"o"},
			Value:   "./worker-bundle",
			Usage:   "Output directory for worker artifacts",
			EnvVars: []string{"K0RDENTD_WORKER_BUNDLE_DIR"},
		},
		&cli.StringFlag{
			Name:    "bundle-path",
			Aliases: []string{"b"},
			Value:   "./airgap-bundle.tar.gz",
			Usage:   "Path to k0rdent airgap bundle (tar.gz or extracted directory)",
			EnvVars: []string{"K0RDENTD_AIRGAP_BUNDLE_PATH"},
		},
	},
	Action: exportWorkerArtifactsAction,
}

func exportWorkerArtifactsAction(c *cli.Context) error {
	logger := utils.GetLogger()

	if !airgap.IsAirGap() {
		logger.Warn("export-worker-artifacts is only available in airgap builds")
		logger.Info("Build a 'k0rdentd-airgap' binary and use that instead")
		return fmt.Errorf("not an airgap build")
	}

	outputDir := c.String("output-dir")
	bundlePath := c.String("bundle-path")

	logger.Infof("Exporting worker artifacts to: %s", outputDir)
	logger.Infof("Referencing k0rdent bundle: %s", bundlePath)

	// Create exporter with bundle path
	exporter := airgap.NewExporter(bundlePath)

	// Create directory structure
	k0sBinaryDir := fmt.Sprintf("%s/k0s-binary", outputDir)
	imagesDir := fmt.Sprintf("%s/images", outputDir)
	scriptsDir := fmt.Sprintf("%s/scripts", outputDir)

	// Extract k0s binary
	if err := exporter.ExtractK0sBinary(k0sBinaryDir); err != nil {
		return fmt.Errorf("failed to extract k0s binary: %w", err)
	}

	// Create bundle reference
	if err := exporter.ExtractImageBundles(imagesDir); err != nil {
		return fmt.Errorf("failed to create bundle reference: %w", err)
	}

	// Generate helper scripts
	if err := exporter.GenerateScripts(scriptsDir); err != nil {
		return fmt.Errorf("failed to generate scripts: %w", err)
	}

	// Generate README
	if err := exporter.GenerateReadme(outputDir); err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}

	logger.Info("Worker artifacts exported successfully")
	logger.Infof("Artifacts location: %s", outputDir)
	logger.Info("Remember to also copy the k0rdent bundle to worker nodes!")
	logger.Infof("Bundle location: %s", bundlePath)

	return nil
}

var ShowFlavorCommand = &cli.Command{
	Name:      "show-flavor",
	Usage:     "Show the build flavor (online or airgap)",
	UsageText: "k0rdentd show-flavor",
	Action:    showFlavorAction,
}

func showFlavorAction(c *cli.Context) error {
	metadata := airgap.GetBuildMetadata()

	fmt.Printf("Build Flavor: %s\n", metadata.Flavor)
	fmt.Printf("Version: %s\n", metadata.Version)
	if metadata.Flavor == "airgap" {
		fmt.Printf("K0s Version: %s\n", metadata.K0sVersion)
		fmt.Printf("K0rdent Version: %s\n", metadata.K0rdentVersion)
	}
	fmt.Printf("Build Time: %s\n", metadata.BuildTime)

	return nil
}
