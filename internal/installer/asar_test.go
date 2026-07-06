package installer

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractIconFromAsar(t *testing.T) {
	// Let's create a synthetic ASAR file
	// JSON header:
	// {"files":{"icon.png":{"size":5,"offset":"0"}}}
	jsonStr := `{"files":{"icon.png":{"size":5,"offset":"0"}}}`
	jsonBytes := []byte(jsonStr)

	jsonSize := uint32(len(jsonBytes))
	headerSize := uint32(8 + jsonSize)

	buf := new(bytes.Buffer)
	// Offset 0-4: 4 bytes ignored
	_ = binary.Write(buf, binary.LittleEndian, uint32(4))
	// Offset 4-8: headerSize
	_ = binary.Write(buf, binary.LittleEndian, headerSize)
	// Offset 8-12: 4 bytes ignored
	_ = binary.Write(buf, binary.LittleEndian, uint32(4))
	// Offset 12-16: jsonSize
	_ = binary.Write(buf, binary.LittleEndian, jsonSize)
	// Offset 16+: json data
	buf.Write(jsonBytes)
	// Binary payload at 8 + headerSize: "hello"
	payload := []byte("hello")
	buf.Write(payload)

	tmpDir, err := os.MkdirTemp("", "asar-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	asarPath := filepath.Join(tmpDir, "app.asar")
	if err := os.WriteFile(asarPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed to write test ASAR: %v", err)
	}

	outPath := filepath.Join(tmpDir, "extracted-icon.png")
	if err := ExtractIconFromAsar(asarPath, outPath); err != nil {
		t.Fatalf("failed to extract icon: %v", err)
	}

	gotBytes, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read extracted icon: %v", err)
	}

	if !bytes.Equal(gotBytes, payload) {
		t.Errorf("extracted icon content = %q; expected %q", gotBytes, payload)
	}
}
