package installer

import (
	"os"
	"os/exec"
)

// DesktopEntryTemplate returns the .desktop file content for Antigravity 2.0.
const DesktopEntryTemplate = `[Desktop Entry]
Name=Antigravity
Comment=Google Antigravity 2.0 agent platform
Exec=/usr/local/bin/antigravity %U
Icon=antigravity
Terminal=false
Type=Application
Categories=Development;IDE;
StartupNotify=true
StartupWMClass=Antigravity
`

// IDEDesktopEntryTemplate returns the .desktop file content for Antigravity IDE.
const IDEDesktopEntryTemplate = `[Desktop Entry]
Name=Antigravity IDE
Comment=Google Antigravity IDE
Exec=/usr/local/bin/antigravity-ide %F
Icon=antigravity-ide
Terminal=false
Type=Application
Categories=Development;IDE;
MimeType=inode/directory;text/plain;application/x-code-workspace;application/x-antigravity-workspace;x-scheme-handler/antigravity-ide;
StartupNotify=true
StartupWMClass=antigravity-ide
`

// FixChromeSandbox updates ownership to root:root and permissions to 4755 for the given sandbox binary path.
func FixChromeSandbox(sandboxPath string) error {
	info, err := os.Stat(sandboxPath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}

	// Change ownership to root:root (uid 0, gid 0)
	if err := os.Chown(sandboxPath, 0, 0); err != nil {
		return err
	}

	// Set suid bit and read/execute permissions (4755 octal is 04755, which is -rwsr-xr-x)
	if err := os.Chmod(sandboxPath, 04755|os.ModeSetuid); err != nil {
		return err
	}

	return nil
}

// RefreshDesktopCaches runs update-desktop-database and gtk-update-icon-cache if they are present on the system.
func RefreshDesktopCaches() {
	if path, err := exec.LookPath("update-desktop-database"); err == nil {
		_ = exec.Command(path, "/usr/share/applications").Run()
	}
	if path, err := exec.LookPath("gtk-update-icon-cache"); err == nil {
		_ = exec.Command(path, "-q", "/usr/share/icons/hicolor").Run()
	}
}
