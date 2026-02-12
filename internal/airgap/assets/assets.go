//go:build airgap
// +build airgap

package assets

import "embed"

// K0sBinary embeds the k0s binary for the target platform
// The binary is placed in internal/airgap/assets/k0s/ during build
//
//go:embed k0s/k0s-*-amd64
var K0sBinary embed.FS

// SkopeoBinary embeds the skopeo binary for the target platform
// The binary is placed in internal/airgap/assets/skopeo/ during build
//
//go:embed skopeo/skopeo-*-amd64
var SkopeoBinary embed.FS

// BuildMetadataJSON embeds the build metadata
// Generated during the build process
//
//go:embed metadata.json
var BuildMetadataJSON []byte

// Note: k0rdent helm charts and image bundles are NOT embedded.
// They are loaded from external bundle path configured by the user.
// This avoids redistributing k0rdent enterprise content we don't own.
