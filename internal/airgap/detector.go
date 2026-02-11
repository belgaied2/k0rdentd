package airgap

// Flavor is set at build time via -ldflags
// Possible values: "online" (default) or "airgap"
var Flavor = "online"

// Version is set at build time via -ldflags
var Version = "dev"

// K0sVersion is the version of embedded k0s binary (airgap builds only)
var K0sVersion = ""

// K0rdentVersion is the version of embedded k0rdent helm chart (airgap builds only)
var K0rdentVersion = ""

// BuildTime is the build timestamp
var BuildTime = ""

// IsAirGap returns true if this binary was built as an airgap flavor
func IsAirGap() bool {
	return Flavor == "airgap"
}

// BuildMetadata contains build-time information
type BuildMetadata struct {
	Flavor         string `json:"flavor"`
	Version        string `json:"version"`
	K0sVersion     string `json:"k0sVersion,omitempty"`
	K0rdentVersion string `json:"k0rdentVersion,omitempty"`
	BuildTime      string `json:"buildTime"`
}

// GetBuildMetadata returns the build metadata
func GetBuildMetadata() BuildMetadata {
	return BuildMetadata{
		Flavor:         Flavor,
		Version:        Version,
		K0sVersion:     K0sVersion,
		K0rdentVersion: K0rdentVersion,
		BuildTime:      BuildTime,
	}
}
