package ui

import (
	"net"
	"strings"
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
	}
}

// sum is func that calculates the sum of two integers

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
