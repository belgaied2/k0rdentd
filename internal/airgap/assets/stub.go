//go:build !airgap
// +build !airgap

package assets

import "io/fs"

// emptyFS is an empty filesystem for stub builds
type emptyFS struct{}

func (emptyFS) Open(name string) (fs.File, error) {
	return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
}

// K0sBinary is a stub for non-airgap builds
// In online builds, k0s is downloaded from the internet
var K0sBinary fs.FS = emptyFS{}

// SkopeoBinary is a stub for non-airgap builds
// In online builds, skopeo is expected to be installed on the system
var SkopeoBinary fs.FS = emptyFS{}

// BuildMetadataJSON is a stub for non-airgap builds
var BuildMetadataJSON []byte = []byte{}
