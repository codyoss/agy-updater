package installer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersionFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://storage.googleapis.com/antigravity-hub/1.2.3/linux-x64/Antigravity.tar.gz", "1.2.3"},
		{"https://storage.googleapis.com/stable/4.5.6-beta/linux-arm/Antigravity.tar.gz", "4.5.6"},
		{"https://storage.googleapis.com/foo/7.8.9/linux-x64/Antigravity%20IDE.tar.gz", "7.8.9"},
		{"https://storage.googleapis.com/no-version/linux-x64/Antigravity.tar.gz", "unknown"},
	}

	for _, tt := range tests {
		got := versionFromURL(tt.url)
		if got != tt.expected {
			t.Errorf("versionFromURL(%q) = %q; expected %q", tt.url, got, tt.expected)
		}
	}
}

func TestResolveDownload(t *testing.T) {
	dummyJS := `
		// Some header comments
		id:"antigravity-2",url:"https://storage.googleapis.com/stable/2.0.4/linux-x64/Antigravity.tar.gz"
		id:"antigravity-cli",url:"https://storage.googleapis.com/stable/2.0.4/linux-x64/Antigravity-CLI.tar.gz"
		id:"antigravity-ide",url:"https://storage.googleapis.com/stable/1.89.2/linux-x64/Antigravity%20IDE.tar.gz"
		id:"antigravity-sdk",url:"https://storage.googleapis.com/stable/1.89.2/linux-x64/Antigravity-SDK.tar.gz"
	`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(dummyJS))
	}))
	defer ts.Close()

	// 1. Test desktop resolution
	ver, urlStr, err := ResolveDownload(ts.URL, "desktop", "linux-x64")
	if err != nil {
		t.Fatalf("unexpected error resolving desktop: %v", err)
	}
	if ver != "2.0.4" {
		t.Errorf("expected version 2.0.4, got %q", ver)
	}
	if urlStr != "https://storage.googleapis.com/stable/2.0.4/linux-x64/Antigravity.tar.gz" {
		t.Errorf("expected URL, got %q", urlStr)
	}

	// 2. Test IDE resolution
	ver, urlStr, err = ResolveDownload(ts.URL, "ide", "linux-x64")
	if err != nil {
		t.Fatalf("unexpected error resolving ide: %v", err)
	}
	if ver != "1.89.2" {
		t.Errorf("expected version 1.89.2, got %q", ver)
	}
	if urlStr != "https://storage.googleapis.com/stable/1.89.2/linux-x64/Antigravity%20IDE.tar.gz" {
		t.Errorf("expected URL, got %q", urlStr)
	}
}

func TestLiveDownloadPage(t *testing.T) {
	jsURL, err := ResolveMainBundle(DownloadPage)
	if err != nil {
		t.Fatalf("ResolveMainBundle failed: %v", err)
	}
	t.Logf("Resolved download bundle URL: %s", jsURL)

	verDesktop, urlDesktop, errDesktop := ResolveDownload(jsURL, "desktop", "linux-x64")
	if errDesktop != nil {
		t.Fatalf("ResolveDownload desktop failed: %v", errDesktop)
	}
	t.Logf("Desktop linux-x64: version=%s, url=%s", verDesktop, urlDesktop)

	verIDE, urlIDE, errIDE := ResolveDownload(jsURL, "ide", "linux-x64")
	if errIDE != nil {
		t.Fatalf("ResolveDownload ide failed: %v", errIDE)
	}
	t.Logf("IDE linux-x64: version=%s, url=%s", verIDE, urlIDE)
}


