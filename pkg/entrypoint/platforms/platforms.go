package platforms

import (
	"errors"
	"fmt"
	"log/slog"
	"path"
	"runtime"
	"sync"
)

var (
	errNotImplemented = errors.New("not implemented")
)

// Platform describes the platform which the image in the manifest runs on.
type Platform struct {
	// Architecture field specifies the CPU architecture, for example
	// `amd64` or `ppc64le`.
	Architecture string `json:"architecture"`

	// OS specifies the operating system, for example `linux` or `windows`.
	OS string `json:"os"`

	// OSVersion is an optional field specifying the operating system
	// version, for example on Windows `10.0.14393.1066`.
	OSVersion string `json:"os.version,omitempty"`

	// OSFeatures is an optional field specifying an array of strings,
	// each listing a required OS feature (for example on Windows `win32k`).
	OSFeatures []string `json:"os.features,omitempty"`

	// Variant is an optional field specifying a variant of the CPU, for
	// example `v7` to specify ARMv7 when architecture is `arm`.
	Variant string `json:"variant,omitempty"`
}

func NewPlatform() *Platform {
	p := &Platform{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Variant:      cpuVariant(),
	}
	return p
}

func (p *Platform) Format() string {
	if p.OS == "" {
		return "unknown"
	}

	return path.Join(p.OS, p.Architecture, p.Variant)
}

// Present the ARM instruction set architecture, eg: v7, v8
// Don't use this value directly; call cpuVariant() instead.
var cpuVariantValue string

var cpuVariantOnce sync.Once

func cpuVariant() string {
	cpuVariantOnce.Do(func() {
		if isArmArch(runtime.GOARCH) {
			var err error
			cpuVariantValue, err = getCPUVariant()
			if err != nil {
				slog.Error("failed to get CPU variant", "os", runtime.GOOS, "error", err)
			}
		}
	})
	return cpuVariantValue
}

// isArmArch returns true if the architecture is ARM.
//
// The arch value should be normalized before being passed to this function.
func isArmArch(arch string) bool {
	switch arch {
	case "arm", "arm64":
		return true
	}
	return false
}

func getCPUVariant() (string, error) {

	var variant string

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		// Windows/Darwin only supports v7 for ARM32 and v8 for ARM64 and so we can use
		// runtime.GOARCH to determine the variants
		switch runtime.GOARCH {
		case "arm64":
			variant = "v8"
		case "arm":
			variant = "v7"
		default:
			variant = "unknown"
		}
	} else if runtime.GOOS == "freebsd" {
		// FreeBSD supports ARMv6 and ARMv7 as well as ARMv4 and ARMv5 (though deprecated)
		// detecting those variants is currently unimplemented
		switch runtime.GOARCH {
		case "arm64":
			variant = "v8"
		default:
			variant = "unknown"
		}
	} else {
		return "", fmt.Errorf("getCPUVariant for OS %s: %v", runtime.GOOS, errNotImplemented)
	}

	return variant, nil
}
