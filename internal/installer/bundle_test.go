package installer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResolveMainBundleAndDownload(t *testing.T) {
	dummyHTML := `<!DOCTYPE html><html><head><script src="/static/js/main-123.js"></script></head></html>`
	dummyJS := `
		id:"antigravity-2",url:"https://storage.googleapis.com/stable/2.0.4/linux-x64/Antigravity.tar.gz"
		id:"antigravity-ide",url:"https://storage.googleapis.com/stable/1.89.2/linux-x64/Antigravity%20IDE.tar.gz"
	`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(dummyHTML))
			return
		}
		if r.URL.Path == "/static/js/main-123.js" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(dummyJS))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	bundleURL, err := ResolveMainBundle(ts.URL + "/download")
	if err != nil {
		t.Fatalf("ResolveMainBundle failed: %v", err)
	}

	if !strings.Contains(bundleURL, "/static/js/main-123.js") {
		t.Fatalf("expected bundle URL to contain /static/js/main-123.js, got %q", bundleURL)
	}

	ver, urlStr, err := ResolveDownload(bundleURL, "desktop", "linux-x64")
	if err != nil {
		t.Fatalf("ResolveDownload failed: %v", err)
	}
	if ver != "2.0.4" {
		t.Errorf("expected version 2.0.4, got %q", ver)
	}
	if urlStr != "https://storage.googleapis.com/stable/2.0.4/linux-x64/Antigravity.tar.gz" {
		t.Errorf("unexpected URL: %s", urlStr)
	}
}
