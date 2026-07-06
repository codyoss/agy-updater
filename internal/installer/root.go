package installer

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// RequireRoot Checks if running as root. If not, attempts to re-execute the program
// using sudo, forwarding all command line arguments and environment variables.
func RequireRoot(cfg *Config) error {
	if os.Getuid() == 0 {
		return nil
	}

	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		return fmt.Errorf("system-wide install needs root. Please run with sudo")
	}

	selfExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine self executable: %w", err)
	}

	// Prepare sudo arguments
	var args []string
	args = append(args, selfExe)
	args = append(args, os.Args[1:]...)

	cmd := exec.Command(sudoPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
			os.Exit(1)
		}
		return fmt.Errorf("failed to re-execute with sudo: %w", err)
	}

	os.Exit(0)
	return nil
}
