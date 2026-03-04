package k0s

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestValidateVersion(t *testing.T) {
	g := gomega.NewWithT(t)

	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "valid version v1.32.4+k0s.0",
			version: "v1.32.4+k0s.0",
			wantErr: false,
		},
		{
			name:    "valid version v1.31.0+k0s.0",
			version: "v1.31.0+k0s.0",
			wantErr: false,
		},
		{
			name:    "valid version v1.30.0+k0s.0",
			version: "v1.30.0+k0s.0",
			wantErr: false,
		},
		{
			name:    "valid version v1.28.5+k0s.3",
			version: "v1.28.5+k0s.3",
			wantErr: false,
		},
		{
			name:    "empty version",
			version: "",
			wantErr: true,
		},
		{
			name:    "missing v prefix",
			version: "1.32.4+k0s.0",
			wantErr: true,
		},
		{
			name:    "missing k0s suffix",
			version: "v1.32.4",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong k0s suffix",
			version: "v1.32.4+k0s",
			wantErr: true,
		},
		{
			name:    "invalid format - extra characters",
			version: "v1.32.4+k0s.0-extra",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.version)
			if tt.wantErr {
				g.Expect(err).ToNot(gomega.BeNil())
			} else {
				g.Expect(err).To(gomega.BeNil())
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	g := gomega.NewWithT(t)

	tests := []struct {
		name         string
		version      string
		wantErr      bool
		wantMajor    int
		wantMinor    int
		wantPatch    int
		wantK0sPatch int
	}{
		{
			name:         "v1.32.4+k0s.0",
			version:      "v1.32.4+k0s.0",
			wantErr:      false,
			wantMajor:    1,
			wantMinor:    32,
			wantPatch:    4,
			wantK0sPatch: 0,
		},
		{
			name:         "v1.31.0+k0s.5",
			version:      "v1.31.0+k0s.5",
			wantErr:      false,
			wantMajor:    1,
			wantMinor:    31,
			wantPatch:    0,
			wantK0sPatch: 5,
		},
		{
			name:    "invalid version",
			version: "invalid",
			wantErr: true,
		},
		{
			name:    "empty version",
			version: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseVersion(tt.version)
			if tt.wantErr {
				g.Expect(err).ToNot(gomega.BeNil())
			} else {
				g.Expect(err).To(gomega.BeNil())
				g.Expect(result.Major).To(gomega.Equal(tt.wantMajor))
				g.Expect(result.Minor).To(gomega.Equal(tt.wantMinor))
				g.Expect(result.Patch).To(gomega.Equal(tt.wantPatch))
				g.Expect(result.K0sPatch).To(gomega.Equal(tt.wantK0sPatch))
				g.Expect(result.Original).To(gomega.Equal(tt.version))
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	g := gomega.NewWithT(t)

	tests := []struct {
		name      string
		v1        string
		v2        string
		want      int
		wantErr   bool
	}{
		{
			name: "equal versions",
			v1:   "v1.32.4+k0s.0",
			v2:   "v1.32.4+k0s.0",
			want: 0,
		},
		{
			name: "v1 < v2 - major version",
			v1:   "v1.31.0+k0s.0",
			v2:   "v1.32.0+k0s.0",
			want: -1,
		},
		{
			name: "v1 > v2 - major version",
			v1:   "v1.32.0+k0s.0",
			v2:   "v1.31.0+k0s.0",
			want: 1,
		},
		{
			name: "v1 < v2 - minor version",
			v1:   "v1.31.0+k0s.0",
			v2:   "v1.31.5+k0s.0",
			want: -1,
		},
		{
			name: "v1 > v2 - minor version",
			v1:   "v1.31.5+k0s.0",
			v2:   "v1.31.0+k0s.0",
			want: 1,
		},
		{
			name: "v1 < v2 - patch version",
			v1:   "v1.31.0+k0s.0",
			v2:   "v1.31.0+k0s.1",
			want: -1,
		},
		{
			name: "v1 > v2 - patch version",
			v1:   "v1.31.0+k0s.2",
			v2:   "v1.31.0+k0s.1",
			want: 1,
		},
		{
			name: "empty v1",
			v1:   "",
			v2:   "v1.32.0+k0s.0",
			want: -1,
		},
		{
			name: "empty v2",
			v1:   "v1.32.0+k0s.0",
			v2:   "",
			want: 1,
		},
		{
			name: "both empty",
			v1:   "",
			v2:   "",
			want: 0,
		},
		{
			name:    "invalid v1",
			v1:     "invalid",
			v2:     "v1.32.0+k0s.0",
			wantErr: true,
		},
		{
			name:    "invalid v2",
			v1:     "v1.32.0+k0s.0",
			v2:     "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CompareVersions(tt.v1, tt.v2)
			if tt.wantErr {
				g.Expect(err).ToNot(gomega.BeNil())
			} else {
				g.Expect(err).To(gomega.BeNil())
				g.Expect(result).To(gomega.Equal(tt.want))
			}
		})
	}
}

func TestVersionsEqual(t *testing.T) {
	g := gomega.NewWithT(t)

	tests := []struct {
		name    string
		v1      string
		v2      string
		want    bool
		wantErr bool
	}{
		{
			name: "equal versions",
			v1:   "v1.32.4+k0s.0",
			v2:   "v1.32.4+k0s.0",
			want: true,
		},
		{
			name: "different versions",
			v1:   "v1.32.4+k0s.0",
			v2:   "v1.31.0+k0s.0",
			want: false,
		},
		{
			name:    "invalid v1",
			v1:     "invalid",
			v2:     "v1.32.0+k0s.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VersionsEqual(tt.v1, tt.v2)
			if tt.wantErr {
				g.Expect(err).ToNot(gomega.BeNil())
			} else {
				g.Expect(err).To(gomega.BeNil())
				g.Expect(result).To(gomega.Equal(tt.want))
			}
		})
	}
}

func TestVersionConflict(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("NeedsResolution - no conflict when versions equal", func(t *testing.T) {
		conflict := &VersionConflict{
			InstalledVersion: "v1.32.4+k0s.0",
			ConfigVersion:    "v1.32.4+k0s.0",
			IsRunning:        false,
		}
		g.Expect(conflict.NeedsResolution()).To(gomega.BeFalse())
	})

	t.Run("NeedsResolution - conflict when versions differ", func(t *testing.T) {
		conflict := &VersionConflict{
			InstalledVersion: "v1.31.0+k0s.0",
			ConfigVersion:    "v1.32.4+k0s.0",
			IsRunning:        false,
		}
		g.Expect(conflict.NeedsResolution()).To(gomega.BeTrue())
	})

	t.Run("NeedsResolution - no conflict when installed is empty", func(t *testing.T) {
		conflict := &VersionConflict{
			InstalledVersion: "",
			ConfigVersion:    "v1.32.4+k0s.0",
			IsRunning:        false,
		}
		g.Expect(conflict.NeedsResolution()).To(gomega.BeFalse())
	})

	t.Run("RequiresManualIntervention - true when running and conflict", func(t *testing.T) {
		conflict := &VersionConflict{
			InstalledVersion: "v1.31.0+k0s.0",
			ConfigVersion:    "v1.32.4+k0s.0",
			IsRunning:        true,
		}
		g.Expect(conflict.RequiresManualIntervention()).To(gomega.BeTrue())
	})

	t.Run("RequiresManualIntervention - false when not running", func(t *testing.T) {
		conflict := &VersionConflict{
			InstalledVersion: "v1.31.0+k0s.0",
			ConfigVersion:    "v1.32.4+k0s.0",
			IsRunning:        false,
		}
		g.Expect(conflict.RequiresManualIntervention()).To(gomega.BeFalse())
	})

	t.Run("CanAutoReplace - true when not running and conflict", func(t *testing.T) {
		conflict := &VersionConflict{
			InstalledVersion: "v1.31.0+k0s.0",
			ConfigVersion:    "v1.32.4+k0s.0",
			IsRunning:        false,
		}
		g.Expect(conflict.CanAutoReplace()).To(gomega.BeTrue())
	})

	t.Run("CanAutoReplace - false when running", func(t *testing.T) {
		conflict := &VersionConflict{
			InstalledVersion: "v1.31.0+k0s.0",
			ConfigVersion:    "v1.32.4+k0s.0",
			IsRunning:        true,
		}
		g.Expect(conflict.CanAutoReplace()).To(gomega.BeFalse())
	})
}

func TestFormatWarningMessage(t *testing.T) {
	g := gomega.NewWithT(t)

	result := FormatWarningMessage("v1.31.0+k0s.0", "v1.32.4+k0s.0")
	g.Expect(result).To(gomega.ContainSubstring("v1.31.0+k0s.0"))
	g.Expect(result).To(gomega.ContainSubstring("v1.32.4+k0s.0"))
	g.Expect(result).To(gomega.ContainSubstring("Using bundled version"))
}

func TestFormatConflictMessage(t *testing.T) {
	g := gomega.NewWithT(t)

	result := FormatConflictMessage("v1.30.0+k0s.0", "v1.32.4+k0s.0")
	g.Expect(result).To(gomega.ContainSubstring("v1.30.0+k0s.0"))
	g.Expect(result).To(gomega.ContainSubstring("v1.32.4+k0s.0"))
	g.Expect(result).To(gomega.ContainSubstring("conflict"))
}

func TestFormatRunningConflictError(t *testing.T) {
	g := gomega.NewWithT(t)

	result := FormatRunningConflictError("v1.30.0+k0s.0", "v1.32.4+k0s.0")
	g.Expect(result).To(gomega.ContainSubstring("v1.30.0+k0s.0"))
	g.Expect(result).To(gomega.ContainSubstring("v1.32.4+k0s.0"))
	g.Expect(result).To(gomega.ContainSubstring("running"))
	g.Expect(result).To(gomega.ContainSubstring("sudo k0s stop"))
}
