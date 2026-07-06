package installer

import (
	"fmt"
	"os"
	"os/exec"
)

// InstallDepsDebian installs dependencies on Debian/Ubuntu systems using apt-get.
// On non-Debian systems, it checks if necessary commands are available.
func InstallDepsDebian(cfg *Config) error {
	if !cfg.InstallDeps {
		return nil
	}

	aptPath, err := exec.LookPath("apt-get")
	if err == nil {
		// apt-get is available
		cmdUpdate := exec.Command(aptPath, "update")
		cmdUpdate.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
		cmdUpdate.Stdout = os.Stdout
		cmdUpdate.Stderr = os.Stderr
		if err := cmdUpdate.Run(); err != nil {
			// Non-fatal, just warn
			fmt.Printf("Warning: failed to run apt-get update: %v\n", err)
		}

		packages := []string{"ca-certificates", "curl", "tar", "desktop-file-utils", "xdg-utils"}
		if cfg.InstallNautilus && cfg.InstallIDE {
			packages = append(packages, "python3", "python3-nautilus")
		}

		args := append([]string{"install", "-y"}, packages...)
		cmdInstall := exec.Command(aptPath, args...)
		cmdInstall.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
		cmdInstall.Stdout = os.Stdout
		cmdInstall.Stderr = os.Stderr
		if err := cmdInstall.Run(); err != nil {
			return fmt.Errorf("failed to install packages: %w", err)
		}
	} else {
		// Check that required tools are in path
		requiredCmds := []string{"curl", "tar"}
		if cfg.InstallNautilus && cfg.InstallIDE {
			requiredCmds = append(requiredCmds, "python3")
		}
		for _, cmd := range requiredCmds {
			if _, err := exec.LookPath(cmd); err != nil {
				return fmt.Errorf("required command not found: %s", cmd)
			}
		}
	}
	return nil
}
