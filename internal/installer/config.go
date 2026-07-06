package installer

// Constants representing project defaults and official URLs.
const (
	ProjectName       = "agy-updater"
	DownloadPage      = "https://antigravity.google/download"
	CLIInstallerURL   = "https://antigravity.google/cli/install.sh"
	DesktopOptRoot    = "/opt/antigravity"
	IDEOptRoot        = "/opt/antigravity-ide"
	DesktopBinLink    = "/usr/local/bin/antigravity"
	IDEBinLink        = "/usr/local/bin/antigravity-ide"
	ManagerBinLink    = "/usr/local/bin/agy-updater"
	UpdateDesktopLink = "/usr/local/bin/update-antigravity"
	UpdateIDELink     = "/usr/local/bin/update-antigravity-ide"
)

// Config holds the parsed CLI arguments/flags.
type Config struct {
	InstallDesktop   bool
	InstallIDE       bool
	InstallCLI       bool
	InstallNautilus  bool
	InstallDeps      bool
	Force            bool
	Yes              bool
	DoStatus         bool
	DoPrintDownloads bool
	DoUninstall      bool
}

// NewDefaultConfig returns a Config with default options.
func NewDefaultConfig() *Config {
	return &Config{
		InstallDesktop:  true,
		InstallIDE:      false,
		InstallCLI:      false,
		InstallNautilus: true,
		InstallDeps:     false,
		Force:           false,
		Yes:             false,
	}
}
