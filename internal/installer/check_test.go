package installer

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFormatStatusMessage(t *testing.T) {
	tests := []struct {
		name         string
		appName      string
		installed    bool
		installedVer string
		newestVer    string
		expected     string
	}{
		{
			name:         "installed and up to date",
			appName:      "Antigravity 2.0",
			installed:    true,
			installedVer: "2.0.4",
			newestVer:    "2.0.4",
			expected:     "- Antigravity 2.0: installed (2.0.4) is up to date",
		},
		{
			name:         "installed and update available",
			appName:      "Antigravity 2.0",
			installed:    true,
			installedVer: "2.0.0",
			newestVer:    "2.0.4",
			expected:     "- Antigravity 2.0: installed (2.0.0), newest version 2.0.4 is available for download",
		},
		{
			name:         "not installed",
			appName:      "Antigravity IDE",
			installed:    false,
			installedVer: "unknown",
			newestVer:    "1.89.2",
			expected:     "- Antigravity IDE: not installed (newest version available: 1.89.2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatStatusMessage(tt.appName, tt.installed, tt.installedVer, tt.newestVer)
			if got != tt.expected {
				t.Errorf("formatStatusMessage() = %q; want %q", got, tt.expected)
			}
		})
	}
}

func TestCheckVersionsWithURL(t *testing.T) {
	dummyJS := `
		id:"antigravity-2",url:"https://storage.googleapis.com/stable/2.0.4/linux-x64/Antigravity.tar.gz"
		id:"antigravity-ide",url:"https://storage.googleapis.com/stable/1.89.2/linux-x64/Antigravity%20IDE.tar.gz"
	`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(dummyJS))
	}))
	defer ts.Close()

	platInfo := &PlatformInfo{
		Platform:   "linux-x64",
		DesktopTop: "Antigravity-x64",
	}

	tests := []struct {
		name           string
		cfg            *Config
		expectedSubstr []string
	}{
		{
			name: "check both components",
			cfg: &Config{
				InstallDesktop: true,
				InstallIDE:     true,
			},
			expectedSubstr: []string{
				"Checking for Antigravity updates...",
				"Antigravity 2.0",
				"2.0.4",
				"Antigravity IDE",
				"1.89.2",
			},
		},
		{
			name: "check desktop only",
			cfg: &Config{
				InstallDesktop: true,
				InstallIDE:     false,
			},
			expectedSubstr: []string{
				"Antigravity 2.0",
				"2.0.4",
			},
		},
		{
			name: "check ide only",
			cfg: &Config{
				InstallDesktop: false,
				InstallIDE:     true,
			},
			expectedSubstr: []string{
				"Antigravity IDE",
				"1.89.2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := CheckVersionsWithURL(&buf, tt.cfg, platInfo, ts.URL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			out := buf.String()
			for _, sub := range tt.expectedSubstr {
				if !strings.Contains(out, sub) {
					t.Errorf("expected output to contain %q, got output:\n%s", sub, out)
				}
			}
			if !tt.cfg.InstallDesktop && strings.Contains(out, "Antigravity 2.0") {
				t.Errorf("expected output NOT to contain Antigravity 2.0 when InstallDesktop is false")
			}
			if !tt.cfg.InstallIDE && strings.Contains(out, "Antigravity IDE") {
				t.Errorf("expected output NOT to contain Antigravity IDE when InstallIDE is false")
			}
		})
	}
}
