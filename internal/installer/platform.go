package installer

import (
	"fmt"
	"runtime"
)

// PlatformInfo contains architecture and top-level directory name mappings.
type PlatformInfo struct {
	Platform   string // e.g. "linux-x64" or "linux-arm"
	DesktopTop string // e.g. "Antigravity-x64" or "Antigravity-arm64"
}

// GetPlatformInfo detects the current OS and CPU architecture.
// Returns an error if the system is not supported.
func GetPlatformInfo() (*PlatformInfo, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("this installer is for Linux only")
	}

	switch runtime.GOARCH {
	case "amd64":
		return &PlatformInfo{
			Platform:   "linux-x64",
			DesktopTop: "Antigravity-x64",
		}, nil
	case "arm64":
		return &PlatformInfo{
			Platform:   "linux-arm",
			DesktopTop: "Antigravity-arm64",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported CPU architecture: %s. Google currently provides x64 and ARM64 Linux builds", runtime.GOARCH)
	}
}
