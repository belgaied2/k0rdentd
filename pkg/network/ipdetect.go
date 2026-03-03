package network

import (
	"fmt"
	"net"
	"strings"

	"github.com/belgaied2/k0rdentd/pkg/utils"
)

// interfaceFilter is a function that returns true if an interface should be ignored
type interfaceFilter func(string) bool

// commonIgnoredInterfaces returns filters for interfaces that should be ignored
func commonIgnoredInterfaces() []interfaceFilter {
	return []interfaceFilter{
		func(name string) bool {
			return strings.HasPrefix(name, "cali")
		},
		func(name string) bool {
			return strings.HasPrefix(name, "vxlan.calico")
		},
		func(name string) bool {
			return strings.HasPrefix(name, "kube-bridge")
		},
		func(name string) bool {
			return strings.HasPrefix(name, "tunl") // tailscale
		},
		func(name string) bool {
			return strings.HasPrefix(name, "docker") // docker
		},
		func(name string) bool {
			return strings.HasPrefix(name, "wg") // wireguard
		},
		func(name string) bool {
			return strings.HasPrefix(name, "veth") // virtual ethernet
		},
		func(name string) bool {
			return strings.HasPrefix(name, "br-") // docker bridges
		},
		func(name string) bool {
			return strings.HasPrefix(name, "cni") // CNI interfaces
		},
		func(name string) bool {
			return strings.HasPrefix(name, "flannel") // flannel
		},
	}
}

// shouldIgnoreInterface checks if an interface should be ignored based on common patterns
func shouldIgnoreInterface(name string) bool {
	for _, filter := range commonIgnoredInterfaces() {
		if filter(name) {
			return true
		}
	}
	return false
}

// GetLocalIPs returns all non-local, non-ignored IP addresses from system
func GetLocalIPs() ([]net.IP, error) {
	var ips []net.IP

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		// Skip interfaces that should be ignored
		if shouldIgnoreInterface(iface.Name) {
			continue
		}

		// Skip interfaces that are down
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Skip loopback interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip IPv6 addresses for now
			if ip == nil || ip.To4() == nil {
				continue
			}

			ips = append(ips, ip)
		}
	}

	return ips, nil
}

// GetInternalIP returns the first internal IP address (for join config)
// This is the IP address that other nodes can use to connect to this node
func GetInternalIP() (string, error) {
	ips, err := GetLocalIPs()
	if err != nil {
		return "", fmt.Errorf("failed to get local IPs: %w", err)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no internal IP addresses found")
	}

	// Return the first IP (usually the primary interface)
	return ips[0].String(), nil
}

// GetInternalIPWithOverride returns the override IP if provided, otherwise auto-detects
func GetInternalIPWithOverride(override string) (string, error) {
	if override != "" {
		// Validate the override IP
		ip := net.ParseIP(override)
		if ip == nil {
			return "", fmt.Errorf("invalid IP address: %s", override)
		}
		return override, nil
	}

	return GetInternalIP()
}

// GetControllerIP is an alias for GetInternalIPWithOverride for clarity
func GetControllerIP(override string) (string, error) {
	return GetInternalIPWithOverride(override)
}

// LogDetectedIPs logs all detected IP addresses for debugging
func LogDetectedIPs() {
	logger := utils.GetLogger()
	ips, err := GetLocalIPs()
	if err != nil {
		logger.Warnf("Failed to get local IPs: %v", err)
		return
	}

	logger.Debugf("Detected %d IP addresses:", len(ips))
	for i, ip := range ips {
		logger.Debugf("  [%d] %s", i+1, ip.String())
	}
}
