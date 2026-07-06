package installer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ExtractIconFromAsar extracts icon.png from an Electron .asar file and writes it to outPath.
func ExtractIconFromAsar(asarPath, outPath string) error {
	f, err := os.Open(asarPath)
	if err != nil {
		return fmt.Errorf("failed to open ASAR file: %w", err)
	}
	defer f.Close()

	// Read format header fields
	// Offset 0: 4 bytes ignored
	var dummy uint32
	if err := binary.Read(f, binary.LittleEndian, &dummy); err != nil {
		return fmt.Errorf("failed to read ASAR header offset 0: %w", err)
	}

	var headerSize uint32
	if err := binary.Read(f, binary.LittleEndian, &headerSize); err != nil {
		return fmt.Errorf("failed to read ASAR header size: %w", err)
	}

	if err := binary.Read(f, binary.LittleEndian, &dummy); err != nil {
		return fmt.Errorf("failed to read ASAR header offset 8: %w", err)
	}

	var jsonSize uint32
	if err := binary.Read(f, binary.LittleEndian, &jsonSize); err != nil {
		return fmt.Errorf("failed to read ASAR JSON size: %w", err)
	}

	jsonBytes := make([]byte, jsonSize)
	if _, err := io.ReadFull(f, jsonBytes); err != nil {
		return fmt.Errorf("failed to read ASAR JSON header: %w", err)
	}

	type AsarFileInfo struct {
		Size   int64       `json:"size"`
		Offset interface{} `json:"offset"` // Can be string or number
	}

	type AsarHeader struct {
		Files map[string]AsarFileInfo `json:"files"`
	}

	var header AsarHeader
	if err := json.Unmarshal(jsonBytes, &header); err != nil {
		return fmt.Errorf("failed to unmarshal ASAR JSON header: %w", err)
	}

	icon, ok := header.Files["icon.png"]
	if !ok {
		return fmt.Errorf("icon.png not found in ASAR header")
	}

	var offset int64
	switch val := icon.Offset.(type) {
	case string:
		if _, err := fmt.Sscan(val, &offset); err != nil {
			return fmt.Errorf("failed to parse offset string %q: %w", val, err)
		}
	case float64:
		offset = int64(val)
	case int64:
		offset = val
	case int:
		offset = int64(val)
	default:
		return fmt.Errorf("unexpected offset type: %T", val)
	}

	// Seek to file start: 8 + headerSize + offset
	startPos := 8 + int64(headerSize) + offset
	if _, err := f.Seek(startPos, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to icon data offset: %w", err)
	}

	outF, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create output icon file: %w", err)
	}
	defer outF.Close()

	if _, err := io.CopyN(outF, f, icon.Size); err != nil {
		return fmt.Errorf("failed to copy icon data: %w", err)
	}

	return nil
}
