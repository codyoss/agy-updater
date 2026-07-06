# Antigravity Linux Updater & Manager

A lightweight, minimal CLI tool written in pure Go to download, install, update, and manage the Google **Antigravity 2.0 Desktop Application** and **Antigravity IDE** on Debian-based Linux distributions (e.g., Debian, Ubuntu, Pop!_OS).

> **Disclaimer:** This is a community‑maintained project and is not affiliated with or endorsed by Google.
>
> **System Requirements:** This tool is designed specifically for Debian-based distributions as it utilizes `apt-get` for automated system dependencies installation.

---

## Key Features

- 🔋 **Zero External Dependencies**: Implemented purely using the Go standard library.
- ⚡ **Official Release Resolution**: Fetches and parses Google's official download pages dynamically to extract the latest release tarball URLs and version details.
- 📦 **Pure Go ASAR Parser**: Includes a custom binary header parser for Electron `.asar` files. It extracts high-resolution icons (`icon.png`) for desktop menus directly from target archives, completely eliminating Node.js or Python runtime requirements.
- 🔒 **Atomic Deployments**: Implements staging directories and atomic directory swaps via `os.Rename` to guarantee that your `/opt` installation folders never end up in a corrupted or half-extracted state.
- 🛠️ **Seamless Desktop Integration**: Generates desktop shortcuts (`.desktop` files), fixes Chrome sandbox setuid permissions (`chrome-sandbox`), and refreshes gtk-icon and desktop databases automatically.
- 📂 **GNOME Files / Nautilus Integration**: Opt-in integration to open folders in the Antigravity IDE directly from the Nautilus context menu.
- 🔑 **Automated Root Elevation**: Automatically detects if root permissions are required and securely re-executes itself under `sudo` (forwarding options) if needed.

---

## Installation & Building

To build the tool from source, ensure you have Go 1.25 or higher installed:

```bash
go install github.com/codyoss/agy-updater/cmd/agy-updater
```

## Usage Guide

```text
Usage:
  agy-updater <subcommand> [options]

Subcommands:
  install, update       Install or update Antigravity 2.0 desktop app and Antigravity IDE
  status                Show installed helper-managed apps and versions
  uninstall             Remove helper-managed Antigravity desktop/IDE files
```

### Install / Update Options

| Flag | Description |
|---|---|
| `--desktop` | Install/update Antigravity 2.0 desktop app only |
| `--ide` | Install/update Antigravity IDE only |
| `--cli` | Also execute Google's official Antigravity CLI installer |
| `--nautilus-support` | Enable GNOME Files/Nautilus context-menu helper |
| `--apt` | Install apt dependencies automatically (e.g. `ca-certificates`, `desktop-file-utils`) |
| `--force` | Reinstall even when local recorded versions are up-to-date |
| `-y, --yes` | Non-interactive mode; assume yes to all prompts |
| `-v, --verbose` | Print the resolved official Google tarball URLs and exit |
| `-h, --help` | Show subcommand help |

### Examples

**1. Install both Desktop and IDE (Default flow):**
```bash
agy-updater install
```

**2. Update only the Antigravity IDE:**
```bash
agy-updater update --ide
```

**3. Install with GNOME Files / Nautilus context menus and install required system dependencies automatically:**
```bash
agy-updater install --nautilus-support --apt -y
```

**4. Check current status & installed versions:**
```bash
agy-updater status
```

**5. Uninstall all components and helpers:**
```bash
agy-updater uninstall
```

---

## Credits

This project is a Go port of the original Bash-based installer/updater, [opensnap/antigravity](https://github.com/opensnap/antigravity).

