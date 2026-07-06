# Architecture: Antigravity Linux Updater Go CLI

This document describes the structure and implementation details of the Go CLI port of the Antigravity Linux Installer.

## Per-File Granularity

### [go.mod](../go.mod)
Defines the Go module path `github.com/codyoss/agy-updater` and sets the Go version to `1.20`. It does not import any external dependencies, sticking strictly to the Go standard library.

### [README.md](../README.md)
Provides a high-level overview of the tool, its features, building and testing instructions, detailed subcommand usage/examples, and developer guidelines.

### [cmd/agy-updater/main.go](../cmd/agy-updater/main.go)
The entry point of the CLI application. Its `main` function delegates execution to a `run` wrapper function that returns errors up to `main` for cleaner error handling and fewer `os.Exit` calls.
- Dispatches execution flow based on the requested subcommand (`install`/`update`, `status`, or `uninstall`).
- Parses subcommand-specific arguments. For the `install` subcommand, flags default to installing both desktop and IDE apps unless restricted, and offers `--nautilus-support` and `--apt` opt-in flags.
- Resolves the download URLs from Google and checks if installation or updates are actually needed before attempting to elevate privileges.
- Builds the `installer.Config` configuration struct and executes the appropriate flow (atomic installation/update, offline status checks, or full component uninstallation).

### [internal/installer/config.go](../internal/installer/config.go)
Declares project constants (URLs, directory paths, and symlink targets) and the configuration structures for the app.
- Defines default target paths in `/opt` and `/usr/local/bin`.
- Holds the `Config` struct which maps to CLI flag choices, defaulting dependencies installation (`InstallDeps`) to `false`.

### [internal/installer/platform.go](../internal/installer/platform.go)
Detects the host OS and architecture.
- Gated to run on `linux` only using Go's `runtime` library.
- Maps `amd64` to `linux-x64` platform name and `Antigravity-x64` archive root directory.
- Maps `arm64` to `linux-arm` platform name and `Antigravity-arm64` archive root directory.

### [internal/installer/resolve.go](../internal/installer/resolve.go)
Handles network requests and text processing to find, fetch, and parse Google's official downloads.
- Implements `ResolveMainBundle` which fetches `https://antigravity.google/download` and scans it using regular expressions to locate the main JavaScript bundle containing download metadata.
- Implements `ResolveDownload` which extracts the tarball URLs and versions for either the `desktop` or `ide` package.
- Implements `versionFromURL` to extract version strings from official Google URL patterns.

### [internal/installer/asar.go](../internal/installer/asar.go)
Implements a pure Go parser for Electron `.asar` archives.
- Parses the binary headers of Electron `app.asar` files.
- Extracts `icon.png` from the file layout by traversing the JSON header index and seeking directly to the byte offset in the archive.
- Eliminates the need for a python runtime for icon extraction.

### [internal/installer/desktop.go](../internal/installer/desktop.go)
Manages desktop integrations, shortcuts, and caching.
- Contains the `.desktop` file specifications/templates.
- Implements `RefreshDesktopCaches` which updates the local desktop database and icon caches by executing `update-desktop-database` and `gtk-update-icon-cache` if present.
- Implements `FixChromeSandbox` to set the owner to `root` and correct permissions (setuid/executable) on the Electron Chrome sandbox executable.

### [internal/installer/deps.go](../internal/installer/deps.go)
Manages installation of system dependencies.
- Triggers non-interactive `apt-get` updates and package installs for needed dependencies (e.g. `ca-certificates`, `curl`, `tar`, `desktop-file-utils`, `xdg-utils`, and Python Nautilus bindings if applicable).
- Validates binary existence on path for non-apt systems.

### [internal/installer/root.go](../internal/installer/root.go)
Verifies current permissions.
- Ensures the installer is running with root permissions.
- If run as an unprivileged user, it attempts to re-execute itself under `sudo`, forwarding command line arguments without using the `-E` flag.

### [internal/installer/install.go](../internal/installer/install.go)
Coordinates the installation phase of each tool.
- Implements `NeedsUpdate` to determine whether the target application versions are already installed and up-to-date.
- Handles atomic folder deployment: downloads the package, extracts the `.tar.gz` to a staging folder, performs checks, sets up desktop configurations, updates permissions, and moves directories atomically via `os.Rename`.
- Manages custom symlinks for desktop and IDE binaries.
- Installs the Nautilus integration file.
- Executes Google's official CLI installer.

### [internal/installer/status.go](../internal/installer/status.go)
Reads installed component details to print status info offline (with fallback support for legacy `.antigravity-linux-version` files). Also contains support for printing raw resolved downloads.

### [internal/installer/uninstall.go](../internal/installer/uninstall.go)
Performs cleanup by removing `/opt` directories, symlinks, desktop shortcuts, icons, and shell helper scripts.
