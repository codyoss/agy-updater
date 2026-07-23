package installer

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// copyFile copies a file from src to dest, setting permissions to mode.
func copyFile(src, dest string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	_ = os.Remove(dest)
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

// extractTarGz extracts a gzipped tarball from r to destDir.
func extractTarGz(r io.Reader, destDir string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip reader error: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar reader error: %w", err)
		}

		target := filepath.Join(destDir, filepath.Clean(hdr.Name))
		// Avoid path traversal attacks
		if filepath.Clean(target) != destDir && !filepath.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(filepath.Separator)) {
			return fmt.Errorf("tar traversal detected: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("failed to create dir: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent dir: %w", err)
			}
			outF, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(outF, tr); err != nil {
				outF.Close()
				return fmt.Errorf("failed to write file content: %w", err)
			}
			outF.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent dir for symlink: %w", err)
			}
			_ = os.Remove(target)
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return fmt.Errorf("failed to create symlink: %w", err)
			}
		}
	}
	return nil
}

// safeReplaceDir atomically moves newdir to target.
func safeReplaceDir(newdir, target string) error {
	prev := target + ".previous"
	_ = os.RemoveAll(prev)
	if _, err := os.Stat(target); err == nil {
		if err := os.Rename(target, prev); err != nil {
			return fmt.Errorf("failed to rename existing target to previous: %w", err)
		}
	}
	if err := os.Rename(newdir, target); err != nil {
		return fmt.Errorf("failed to rename new dir to target: %w", err)
	}
	return nil
}

// downloadAndExtract downloads a tarball from urlStr and extracts it to destDir.
func downloadAndExtract(urlStr, destDir string) error {
	resp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http GET failed with status %d", resp.StatusCode)
	}

	return extractTarGz(resp.Body, destDir)
}

// InstallDesktopApp downloads and installs the Antigravity 2.0 desktop application.
func InstallDesktopApp(cfg *Config, platInfo *PlatformInfo, jsBundleURL string) error {
	ver, urlStr, err := ResolveDownload(jsBundleURL, "desktop", platInfo.Platform)
	if err != nil {
		return err
	}

	root := DesktopOptRoot
	targetBin := filepath.Join(root, platInfo.DesktopTop, "antigravity")
	versionFile := filepath.Join(root, ".agy-updater-version")

	if !cfg.Force && isExecutable(targetBin) && getVersion(versionFile) == ver {
		fmt.Printf("Antigravity 2.0 %s is already installed.\n", ver)
		return nil
	}

	fmt.Printf("Downloading Antigravity 2.0 %s for %s from Google...\n", ver, platInfo.Platform)
	if err := os.MkdirAll("/opt", 0755); err != nil {
		return fmt.Errorf("failed to create /opt: %w", err)
	}
	tmpdir, err := os.MkdirTemp("/opt", "antigravity-desktop-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpdir)

	if err := downloadAndExtract(urlStr, tmpdir); err != nil {
		return fmt.Errorf("failed to download or extract: %w", err)
	}

	extractedBin := filepath.Join(tmpdir, platInfo.DesktopTop, "antigravity")
	if !isExecutable(extractedBin) {
		return fmt.Errorf("Antigravity launcher not found inside tarball")
	}

	// Extract icon
	iconStaged := filepath.Join(tmpdir, "antigravity.png")
	asarPath := filepath.Join(tmpdir, platInfo.DesktopTop, "resources", "app.asar")
	if err := ExtractIconFromAsar(asarPath, iconStaged); err != nil {
		fmt.Printf("Warning: could not extract desktop icon: %v\n", err)
	}

	// Stage atomically
	rootNew := root + ".new"
	_ = os.RemoveAll(rootNew)
	if err := os.MkdirAll(rootNew, 0755); err != nil {
		return err
	}

	if err := os.Rename(filepath.Join(tmpdir, platInfo.DesktopTop), filepath.Join(rootNew, platInfo.DesktopTop)); err != nil {
		return fmt.Errorf("failed to move files to staging directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(rootNew, ".agy-updater-version"), []byte(ver), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(rootNew, ".agy-updater-source-url"), []byte(urlStr), 0644); err != nil {
		return err
	}

	_ = FixChromeSandbox(filepath.Join(rootNew, platInfo.DesktopTop, "chrome-sandbox"))

	if err := safeReplaceDir(rootNew, root); err != nil {
		return err
	}

	// Symlink
	_ = os.Remove(DesktopBinLink)
	if err := os.Symlink(filepath.Join(root, platInfo.DesktopTop, "antigravity"), DesktopBinLink); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Icon
	if _, err := os.Stat(iconStaged); err == nil {
		_ = os.MkdirAll("/usr/share/icons/hicolor/512x512/apps", 0755)
		_ = copyFile(iconStaged, "/usr/share/icons/hicolor/512x512/apps/antigravity.png", 0644)
	}

	// Desktop entry
	_ = os.MkdirAll("/usr/share/applications", 0755)
	if err := os.WriteFile("/usr/share/applications/antigravity.desktop", []byte(DesktopEntryTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write desktop entry: %w", err)
	}

	CleanUserDesktopOverrides()
	RefreshDesktopCaches()
	fmt.Printf("Installed Antigravity 2.0 %s at %s/%s\n", ver, root, platInfo.DesktopTop)
	return nil
}

// InstallIDEApp downloads and installs the Antigravity IDE application.
func InstallIDEApp(cfg *Config, platInfo *PlatformInfo, jsBundleURL string) error {
	ver, urlStr, err := ResolveDownload(jsBundleURL, "ide", platInfo.Platform)
	if err != nil {
		return err
	}

	root := IDEOptRoot
	installDir := "Antigravity-IDE"
	targetBin := filepath.Join(root, installDir, "antigravity-ide")
	versionFile := filepath.Join(root, ".agy-updater-version")

	if !cfg.Force && isExecutable(targetBin) && getVersion(versionFile) == ver {
		fmt.Printf("Antigravity IDE %s is already installed.\n", ver)
		return nil
	}

	fmt.Printf("Downloading Antigravity IDE %s for %s from Google...\n", ver, platInfo.Platform)
	if err := os.MkdirAll("/opt", 0755); err != nil {
		return fmt.Errorf("failed to create /opt: %w", err)
	}
	tmpdir, err := os.MkdirTemp("/opt", "antigravity-ide-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpdir)

	if err := downloadAndExtract(urlStr, tmpdir); err != nil {
		return fmt.Errorf("failed to download or extract: %w", err)
	}

	extractedBin := filepath.Join(tmpdir, "Antigravity IDE", "antigravity-ide")
	if !isExecutable(extractedBin) {
		return fmt.Errorf("Antigravity IDE launcher not found inside tarball")
	}

	// Stage atomically
	rootNew := root + ".new"
	_ = os.RemoveAll(rootNew)
	if err := os.MkdirAll(rootNew, 0755); err != nil {
		return err
	}

	if err := os.Rename(filepath.Join(tmpdir, "Antigravity IDE"), filepath.Join(rootNew, installDir)); err != nil {
		return fmt.Errorf("failed to move files to staging directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(rootNew, ".agy-updater-version"), []byte(ver), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(rootNew, ".agy-updater-source-url"), []byte(urlStr), 0644); err != nil {
		return err
	}

	_ = FixChromeSandbox(filepath.Join(rootNew, installDir, "chrome-sandbox"))

	if err := safeReplaceDir(rootNew, root); err != nil {
		return err
	}

	// Symlink
	_ = os.Remove(IDEBinLink)
	if err := os.Symlink(filepath.Join(root, installDir, "antigravity-ide"), IDEBinLink); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Icon
	iconSource := filepath.Join(root, installDir, "resources", "app", "resources", "linux", "code.png")
	if _, err := os.Stat(iconSource); err == nil {
		_ = os.MkdirAll("/usr/share/icons/hicolor/512x512/apps", 0755)
		_ = copyFile(iconSource, "/usr/share/icons/hicolor/512x512/apps/antigravity-ide.png", 0644)
	}

	// Desktop entry
	_ = os.MkdirAll("/usr/share/applications", 0755)
	if err := os.WriteFile("/usr/share/applications/antigravity-ide.desktop", []byte(IDEDesktopEntryTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write desktop entry: %w", err)
	}

	CleanUserDesktopOverrides()
	RefreshDesktopCaches()
	fmt.Printf("Installed Antigravity IDE %s at %s/%s\n", ver, root, installDir)
	return nil
}

// InstallNautilusExtension installs the GNOME Files/Nautilus context menu extension.
func InstallNautilusExtension(cfg *Config) error {
	if !cfg.InstallNautilus || !cfg.InstallIDE {
		return nil
	}

	if _, err := exec.LookPath("nautilus"); err != nil {
		return nil // Nautilus not installed
	}

	// Verify Python Nautilus bindings are available
	pythonCheck := "import gi; gi.require_version('Nautilus', '4.0')"
	if err := exec.Command("python3", "-c", pythonCheck).Run(); err != nil {
		fmt.Println("Warning: Skipping Nautilus extension because Python Nautilus bindings are unavailable.")
		return nil
	}

	nautilusExtDir := "/usr/share/nautilus-python/extensions"
	if err := os.MkdirAll(nautilusExtDir, 0755); err != nil {
		return fmt.Errorf("failed to create nautilus extensions dir: %w", err)
	}

	extContent := `import subprocess
from urllib.parse import unquote, urlparse
from gi.repository import Nautilus, GObject

class OpenInAntigravityIDE(GObject.GObject, Nautilus.MenuProvider):
    def _path(self, file_info):
        uri = file_info.get_uri()
        parsed = urlparse(uri)
        if parsed.scheme != 'file':
            return None
        return unquote(parsed.path)

    def get_file_items(self, files):
        if not files or len(files) != 1:
            return []
        path = self._path(files[0])
        if not path:
            return []
        item = Nautilus.MenuItem(
            name='OpenInAntigravityIDE::open',
            label='Open in Antigravity IDE',
            tip='Open this folder or file in Antigravity IDE'
        )
        item.connect('activate', lambda _item: subprocess.Popen(['antigravity-ide', path]))
        return [item]

    def get_background_items(self, folder):
        path = self._path(folder)
        if not path:
            return []
        item = Nautilus.MenuItem(
            name='OpenInAntigravityIDE::open_background',
            label='Open Folder in Antigravity IDE',
            tip='Open the current folder in Antigravity IDE'
        )
        item.connect('activate', lambda _item: subprocess.Popen(['antigravity-ide', path]))
        return [item]
`

	extPath := filepath.Join(nautilusExtDir, "open-in-antigravity-ide.py")
	if err := os.WriteFile(extPath, []byte(extContent), 0644); err != nil {
		return fmt.Errorf("failed to write nautilus extension: %w", err)
	}

	fmt.Println("Installed Nautilus context-menu helper. Restart Files/Nautilus to see it.")
	return nil
}

// InstallCLI installs Google's official Antigravity CLI.
func InstallCLI(cfg *Config) error {
	if !cfg.InstallCLI {
		return nil
	}

	fmt.Println("Running Google's official Antigravity CLI installer for the non-root user...")
	sudoUser := os.Getenv("SUDO_USER")

	var cmd *exec.Cmd
	if sudoUser != "" && sudoUser != "root" {
		cmd = exec.Command("sudo", "-u", sudoUser, "-H", "bash", "-lc", "curl -fsSL "+CLIInstallerURL+" | bash")
	} else {
		cmd = exec.Command("bash", "-c", "curl -fsSL "+CLIInstallerURL+" | bash")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("official CLI installer failed: %w", err)
	}

	return nil
}

// PrintSuccessSummary displays the final status and execution details to the user.
func PrintSuccessSummary(cfg *Config) {
	selfCmd := os.Args[0]
	fmt.Println()
	fmt.Println("Antigravity Linux install complete.")
	fmt.Println()
	fmt.Println("Installed:")
	if cfg.InstallDesktop {
		fmt.Printf("- Antigravity 2.0: %s\n", DesktopBinLink)
	}
	if cfg.InstallIDE {
		fmt.Printf("- Antigravity IDE: %s\n", IDEBinLink)
	}
	fmt.Println()
	fmt.Println("Manage:")
	fmt.Printf("- Status:    %s status\n", selfCmd)
	fmt.Printf("- Update:    %s install\n", selfCmd)
	fmt.Printf("- Uninstall: %s uninstall\n", selfCmd)

	if cfg.InstallIDE {
		fmt.Println()
		fmt.Println("Folder open integration: use your file manager's Open With menu, or Nautilus context menu after restarting Files.")
	}
}

// NeedsUpdate returns true if an update/install is required.
func NeedsUpdate(cfg *Config, platInfo *PlatformInfo, jsBundleURL string) (bool, error) {
	if cfg.Force {
		return true, nil
	}

	if cfg.InstallDesktop {
		ver, _, err := ResolveDownload(jsBundleURL, "desktop", platInfo.Platform)
		if err != nil {
			return false, err
		}
		root := DesktopOptRoot
		targetBin := filepath.Join(root, platInfo.DesktopTop, "antigravity")
		versionFile := filepath.Join(root, ".agy-updater-version")
		if !isExecutable(targetBin) || getVersion(versionFile) != ver {
			return true, nil
		}
	}

	if cfg.InstallIDE {
		ver, _, err := ResolveDownload(jsBundleURL, "ide", platInfo.Platform)
		if err != nil {
			return false, err
		}
		root := IDEOptRoot
		installDir := "Antigravity-IDE"
		targetBin := filepath.Join(root, installDir, "antigravity-ide")
		versionFile := filepath.Join(root, ".agy-updater-version")
		if !isExecutable(targetBin) || getVersion(versionFile) != ver {
			return true, nil
		}
	}

	return false, nil
}

