package installer

import (
	"runtime"
	"testing"
)

func TestPlatformDetection(t *testing.T) {
	platInfo, err := GetPlatformInfo()
	if runtime.GOOS != "linux" {
		if err == nil {
			t.Errorf("expected error on non-linux OS %s", runtime.GOOS)
		}
		return
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runtime.GOARCH == "amd64" {
		if platInfo.Platform != "linux-x64" {
			t.Errorf("expected platform linux-x64, got %q", platInfo.Platform)
		}
		if platInfo.DesktopTop != "Antigravity-x64" {
			t.Errorf("expected desktop top Antigravity-x64, got %q", platInfo.DesktopTop)
		}
	} else if runtime.GOARCH == "arm64" {
		if platInfo.Platform != "linux-arm" {
			t.Errorf("expected platform linux-arm, got %q", platInfo.Platform)
		}
		if platInfo.DesktopTop != "Antigravity-arm64" {
			t.Errorf("expected desktop top Antigravity-arm64, got %q", platInfo.DesktopTop)
		}
	}
}
