package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getVersion reads the version string from versionFile, returning "unknown" on error/not found.
// It falls back to checking .antigravity-linux-version for legacy installations.
func getVersion(versionFile string) string {
	b, err := os.ReadFile(versionFile)
	if err == nil {
		return strings.TrimSpace(string(b))
	}
	dir := filepath.Dir(versionFile)
	legacyFile := filepath.Join(dir, ".antigravity-linux-version")
	if b2, err2 := os.ReadFile(legacyFile); err2 == nil {
		return strings.TrimSpace(string(b2))
	}
	return "unknown"
}

// isExecutable checks if a file exists and is executable.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	// Check if executable by owner, group, or other
	return info.Mode().IsRegular() && (info.Mode()&0111 != 0)
}

// PrintStatus displays the current installation status of all components.
func PrintStatus() {
	fmt.Println("Antigravity Linux status")

	if isExecutable(DesktopBinLink) {
		ver := getVersion("/opt/antigravity/.agy-updater-version")
		fmt.Printf("- Antigravity 2.0: installed (%s)\n", ver)
		fmt.Printf("  Command: %s\n", DesktopBinLink)
	} else {
		fmt.Println("- Antigravity 2.0: not installed by this helper")
	}

	if isExecutable(IDEBinLink) {
		ver := getVersion("/opt/antigravity-ide/.agy-updater-version")
		fmt.Printf("- Antigravity IDE: installed (%s)\n", ver)
		fmt.Printf("  Command: %s\n", IDEBinLink)
	} else {
		fmt.Println("- Antigravity IDE: not installed by this helper")
	}

	if isExecutable(ManagerBinLink) {
		fmt.Println("- Update helper: installed")
	} else {
		fmt.Println("- Update helper: not installed")
	}
}

// PrintDownloads resolves and prints the official download URLs.
func PrintDownloads(cfg *Config) error {
	platInfo, err := GetPlatformInfo()
	if err != nil {
		return err
	}

	fmt.Println("Resolving download URLs from Google...")
	jsURL, err := ResolveMainBundle(DownloadPage)
	if err != nil {
		return err
	}

	if cfg.InstallDesktop {
		ver, urlStr, err := ResolveDownload(jsURL, "desktop", platInfo.Platform)
		if err != nil {
			return fmt.Errorf("failed to resolve Desktop download: %w", err)
		}
		fmt.Printf("Antigravity 2.0 %s: %s\n", ver, urlStr)
	}

	if cfg.InstallIDE {
		ver, urlStr, err := ResolveDownload(jsURL, "ide", platInfo.Platform)
		if err != nil {
			return fmt.Errorf("failed to resolve IDE download: %w", err)
		}
		fmt.Printf("Antigravity IDE %s: %s\n", ver, urlStr)
	}

	return nil
}
