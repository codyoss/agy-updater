package installer

import (
	"fmt"
	"os"
)

// UninstallAll removes all files and directories installed by the installer.
func UninstallAll(cfg *Config) error {
	if err := RequireRoot(cfg); err != nil {
		return err
	}

	fmt.Println("Uninstalling helper-managed Antigravity installations...")

	dirs := []string{
		"/opt/antigravity",
		"/opt/antigravity.new",
		"/opt/antigravity.previous",
		"/opt/antigravity-ide",
		"/opt/antigravity-ide.new",
		"/opt/antigravity-ide.previous",
	}

	files := []string{
		DesktopBinLink,
		IDEBinLink,
		UpdateDesktopLink,
		UpdateIDELink,
		ManagerBinLink,
		"/usr/share/applications/antigravity.desktop",
		"/usr/share/applications/antigravity-ide.desktop",
		"/usr/share/icons/hicolor/512x512/apps/antigravity.png",
		"/usr/share/icons/hicolor/512x512/apps/antigravity-ide.png",
		"/usr/share/nautilus-python/extensions/open-in-antigravity-ide.py",
	}

	for _, d := range dirs {
		if err := os.RemoveAll(d); err != nil {
			fmt.Printf("Warning: failed to remove directory %s: %v\n", d, err)
		}
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove file %s: %v\n", f, err)
		}
	}

	RefreshDesktopCaches()

	fmt.Println("Removed helper-managed Antigravity files. User settings under home directories were left untouched.")
	return nil
}
