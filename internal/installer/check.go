package installer

import (
	"fmt"
	"io"
	"path/filepath"
)

// formatStatusMessage returns a human-readable status string for an application check.
func formatStatusMessage(appName string, installed bool, installedVer, newestVer string) string {
	if installed {
		if installedVer != newestVer {
			return fmt.Sprintf("- %s: installed (%s), newest version %s is available for download", appName, installedVer, newestVer)
		}
		return fmt.Sprintf("- %s: installed (%s) is up to date", appName, installedVer)
	}
	return fmt.Sprintf("- %s: not installed (newest version available: %s)", appName, newestVer)
}

// CheckVersions resolves the latest available versions and compares them against installed versions.
func CheckVersions(out io.Writer, cfg *Config) error {
	platInfo, err := GetPlatformInfo()
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "Resolving download URLs from Google...")
	jsURL, err := ResolveMainBundle(DownloadPage)
	if err != nil {
		return err
	}

	return CheckVersionsWithURL(out, cfg, platInfo, jsURL)
}

// CheckVersionsWithURL compares installed app versions with resolved versions from jsBundleURL.
func CheckVersionsWithURL(out io.Writer, cfg *Config, platInfo *PlatformInfo, jsBundleURL string) error {
	fmt.Fprintln(out, "Checking for Antigravity updates...")

	if cfg.InstallDesktop {
		ver, _, err := ResolveDownload(jsBundleURL, "desktop", platInfo.Platform)
		if err != nil {
			fmt.Fprintf(out, "- Antigravity 2.0: failed to check version: %v\n", err)
		} else {
			targetBin := filepath.Join(DesktopOptRoot, platInfo.DesktopTop, "antigravity")
			versionFile := filepath.Join(DesktopOptRoot, ".agy-updater-version")
			installed := isExecutable(targetBin)
			installedVer := getVersion(versionFile)
			fmt.Fprintln(out, formatStatusMessage("Antigravity 2.0", installed, installedVer, ver))
		}
	}

	if cfg.InstallIDE {
		ver, _, err := ResolveDownload(jsBundleURL, "ide", platInfo.Platform)
		if err != nil {
			fmt.Fprintf(out, "- Antigravity IDE: failed to check version: %v\n", err)
		} else {
			installDir := "Antigravity-IDE"
			targetBin := filepath.Join(IDEOptRoot, installDir, "antigravity-ide")
			versionFile := filepath.Join(IDEOptRoot, ".agy-updater-version")
			installed := isExecutable(targetBin)
			installedVer := getVersion(versionFile)
			fmt.Fprintln(out, formatStatusMessage("Antigravity IDE", installed, installedVer, ver))
		}
	}

	return nil
}
