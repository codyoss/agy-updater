package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codyoss/agy-updater/internal/installer"
)

// Options for 'install' / 'update':
//   --desktop              Install/update Antigravity 2.0 desktop app only
//   --ide                  Install/update Antigravity IDE only
//   --cli                  Also run Google's official Antigravity CLI installer
//   --nautilus-support     Enable GNOME Files/Nautilus context-menu helper
//   --apt                  Install apt dependencies automatically
//   --force                Reinstall even when the recorded version matches
//   -y, --yes              Non-interactive; assume yes where possible
//   -v, --verbose          Print the resolved official Google tarball URLs and exit
//   -h, --help             Show this help
func usage() {
	fmt.Print(`Antigravity Linux Installer & Manager

Usage:
  agy-updater <subcommand> [options]

Subcommands:
  install, update       Install or update Antigravity 2.0 desktop app and Antigravity IDE
  status                Show installed helper-managed apps and versions
  uninstall             Remove helper-managed Antigravity desktop/IDE files

Options for 'install' / 'update':
  --desktop              Install/update Antigravity 2.0 desktop app only
  --ide                  Install/update Antigravity IDE only
  --cli                  Also run Google's official Antigravity CLI installer
  --nautilus-support     Enable GNOME Files/Nautilus context-menu helper
  --apt                  Install apt dependencies automatically
  --force                Reinstall even when the recorded version matches
  -y, --yes              Non-interactive; assume yes where possible
  -v, --verbose          Print the resolved official Google tarball URLs and exit
  -h, --help             Show this help
`)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		usage()
		return fmt.Errorf("missing subcommand")
	}

	subcommand := os.Args[1]
	if subcommand == "help" || subcommand == "-h" || subcommand == "--help" {
		usage()
		return nil
	}

	if subcommand != "install" && subcommand != "update" && subcommand != "status" && subcommand != "uninstall" {
		usage()
		return fmt.Errorf("unknown subcommand %q", subcommand)
	}

	cfg := &installer.Config{}


	if subcommand == "status" {
		cfg.DoStatus = true
		installer.PrintStatus()
		return nil
	}

	if subcommand == "uninstall" {
		cfg.DoUninstall = true
		if err := installer.UninstallAll(cfg); err != nil {
			return err
		}
		return nil
	}

	// subcommand is "install" or "update"
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.Usage = usage

	flagDesktop := fs.Bool("desktop", false, "")
	flagIDE := fs.Bool("ide", false, "")
	flagCLI := fs.Bool("cli", false, "")
	flagNautilusSupport := fs.Bool("nautilus-support", false, "")
	flagApt := fs.Bool("apt", false, "")
	flagForce := fs.Bool("force", false, "")
	flagV := fs.Bool("v", false, "")
	flagVerbose := fs.Bool("verbose", false, "")
	flagY := fs.Bool("y", false, "")
	flagYes := fs.Bool("yes", false, "")

	if err := fs.Parse(os.Args[2:]); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return fmt.Errorf("parsing flags: %w", err)
	}

	cfg.InstallDesktop = true
	cfg.InstallIDE = true
	cfg.InstallCLI = *flagCLI
	cfg.InstallNautilus = *flagNautilusSupport
	cfg.InstallDeps = *flagApt
	cfg.Force = *flagForce
	cfg.Yes = *flagY || *flagYes
	cfg.DoPrintDownloads = *flagV || *flagVerbose

	// Resolve desktop/ide selection flags
	if *flagIDE && *flagDesktop {
		cfg.InstallDesktop = true
		cfg.InstallIDE = true
	} else if *flagIDE {
		cfg.InstallDesktop = false
		cfg.InstallIDE = true
	} else if *flagDesktop {
		cfg.InstallDesktop = true
		cfg.InstallIDE = false
	} else {
		// Default
		cfg.InstallDesktop = true
		cfg.InstallIDE = true
	}

	if cfg.DoPrintDownloads {
		if err := installer.PrintDownloads(cfg); err != nil {
			return err
		}
		return nil
	}

	// Default flow: Install or Update
	platInfo, err := installer.GetPlatformInfo()
	if err != nil {
		return err
	}

	fmt.Println("Resolving download URLs from Google...")
	jsBundleURL, err := installer.ResolveMainBundle(installer.DownloadPage)
	if err != nil {
		return err
	}

	needsUpdate, err := installer.NeedsUpdate(cfg, platInfo, jsBundleURL)
	if err != nil {
		return err
	}

	if !needsUpdate {
		if cfg.InstallDesktop {
			ver, _, _ := installer.ResolveDownload(jsBundleURL, "desktop", platInfo.Platform)
			fmt.Printf("Antigravity 2.0 %s is already installed.\n", ver)
		}
		if cfg.InstallIDE {
			ver, _, _ := installer.ResolveDownload(jsBundleURL, "ide", platInfo.Platform)
			fmt.Printf("Antigravity IDE %s is already installed.\n", ver)
		}
		installer.PrintSuccessSummary(cfg)
		return nil
	}

	if err := installer.RequireRoot(cfg); err != nil {
		return err
	}

	if err := installer.InstallDepsDebian(cfg); err != nil {
		return err
	}

	if cfg.InstallDesktop {
		if err := installer.InstallDesktopApp(cfg, platInfo, jsBundleURL); err != nil {
			return err
		}
	}

	if cfg.InstallIDE {
		if err := installer.InstallIDEApp(cfg, platInfo, jsBundleURL); err != nil {
			return err
		}
	}

	if err := installer.InstallNautilusExtension(cfg); err != nil {
		return err
	}

	if err := installer.InstallCLI(cfg); err != nil {
		return err
	}

	installer.PrintSuccessSummary(cfg)
	return nil
}

