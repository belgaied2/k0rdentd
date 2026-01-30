package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/belgaied2/k0rdentd/pkg/utils"
)

// CloudProvider represents type of cloud provider
type CloudProvider string

const (
	CloudProviderAWS   CloudProvider = "aws"
	CloudProviderGCP   CloudProvider = "gcp"
	CloudProviderAzure CloudProvider = "azure"
	CloudProviderNone  CloudProvider = "none"
)

// awsMetadataResponse represents response from AWS metadata service
type awsMetadataResponse struct {
	PublicIPv4 string `json:"public-ipv4"`
}

// gcpMetadataResponse represents response from GCP metadata service
type gcpMetadataResponse struct {
	NetworkInterfaces []struct {
		AccessConfigs []struct {
			ExternalIP string `json:"externalIp"`
		} `json:"accessConfigs"`
	} `json:"networkInterfaces"`
}

// azureMetadataResponse represents response from Azure metadata service
type azureMetadataResponse struct {
	Network struct {
		PublicIPAddress string `json:"publicIpAddress"`
	} `json:"network"`
}

// DetectCloudProvider attempts to detect cloud provider using dmidecode
func DetectCloudProvider() CloudProvider {
	cmd := exec.Command("dmidecode", "-s", "bios-vendor")
	output, err := cmd.Output()
	if err != nil {
		utils.GetLogger().Debug("dmidecode not available, cannot detect cloud provider via BIOS vendor")
		return CloudProviderNone
	}

	vendor := strings.ToLower(strings.TrimSpace(string(output)))

	switch {
	case strings.Contains(vendor, "amazon"):
		return CloudProviderAWS
	case strings.Contains(vendor, "google"):
		return CloudProviderGCP
	case strings.Contains(vendor, "microsoft"):
		return CloudProviderAzure
	default:
		return CloudProviderNone
	}
}

// getAWSMetadata retrieves public IP from AWS metadata service
func getAWSMetadata() (string, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("http://checkip.amazonaws.com")
	if err != nil {
		return "", fmt.Errorf("failed to fetch AWS metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AWS metadata service returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read AWS metadata response: %w", err)
	}

	// AWS metadata service returns plain text, not JSON
	return strings.TrimSpace(string(body)), nil
}

// getGCPMetadata retrieves public IP from GCP metadata service
func getGCPMetadata() (string, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("http://metadata.google.internal/computeMetadata/v1/")
	if err != nil {
		return "", fmt.Errorf("failed to fetch GCP metadata endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GCP metadata service returned status: %d", resp.StatusCode)
	}

	_, _ = io.ReadAll(resp.Body)

	// GCP returns directory listing, we need to fetch network interface data
	resp2, err := client.Get("http://metadata.google.internal/computeMetadata/v1/network-interfaces/")
	if err != nil {
		return "", fmt.Errorf("failed to fetch GCP network metadata: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GCP network metadata service returned status: %d", resp2.StatusCode)
	}

	interfacesBody, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read GCP network metadata response: %w", err)
	}

	interfacesList := strings.Split(strings.TrimSpace(string(interfacesBody)), "\n")
	if len(interfacesList) == 0 {
		return "", fmt.Errorf("no network interfaces found in GCP metadata")
	}

	// Get metadata for first interface
	resp3, err := client.Get(fmt.Sprintf("http://metadata.google.internal/computeMetadata/v1/network-interfaces/%s/", interfacesList[0]))
	if err != nil {
		return "", fmt.Errorf("failed to fetch GCP interface metadata: %w", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GCP interface metadata service returned status: %d", resp3.StatusCode)
	}

	interfaceBody, err := io.ReadAll(resp3.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read GCP interface metadata response: %w", err)
	}

	var gcpResp gcpMetadataResponse
	if err := json.Unmarshal(interfaceBody, &gcpResp); err != nil {
		return "", fmt.Errorf("failed to parse GCP metadata response: %w", err)
	}

	if len(gcpResp.NetworkInterfaces) > 0 && len(gcpResp.NetworkInterfaces[0].AccessConfigs) > 0 {
		return gcpResp.NetworkInterfaces[0].AccessConfigs[0].ExternalIP, nil
	}

	return "", nil
}

// getAzureMetadata retrieves public IP from Azure metadata service
func getAzureMetadata() (string, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/instance?api-version=2021-02-01", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create Azure metadata request: %w", err)
	}
	req.Header.Set("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Azure metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Azure metadata service returned status: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Azure metadata response: %w", err)
	}

	var azureResp azureMetadataResponse
	if err := json.Unmarshal(respBody, &azureResp); err != nil {
		return "", fmt.Errorf("failed to parse Azure metadata response: %w", err)
	}

	return azureResp.Network.PublicIPAddress, nil
}

// GetExternalIP attempts to get external IP from cloud metadata service
// If no cloud is detected or metadata is unavailable, returns empty string
func GetExternalIP() string {
	provider := DetectCloudProvider()

	var ip string
	var err error

	switch provider {
	case CloudProviderAWS:
		ip, err = getAWSMetadata()
	case CloudProviderGCP:
		ip, err = getGCPMetadata()
	case CloudProviderAzure:
		ip, err = getAzureMetadata()
	default:
		utils.GetLogger().Debug("No known cloud provider detected")
		return ""
	}

	if err != nil {
		utils.GetLogger().Debugf("Failed to get external IP from %s cloud: %v", provider, err)
		return ""
	}

	utils.GetLogger().Debugf("Detected external IP %s from %s cloud", ip, provider)
	return ip
}
