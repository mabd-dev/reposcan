package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mabd-dev/reposcan/pkg/report"
)

func TestWriteScanReport_WritesFileWithTimestampedName(t *testing.T) {
	dir := t.TempDir()
	r := report.ScanReport{
		Version:     1,
		GeneratedAt: time.Date(2025, 8, 31, 22, 4, 45, 0, time.UTC),
		RepoStates:  nil,
		Warnings:    nil,
	}

	if err := WriteScanReport(r, dir); err != nil {
		t.Fatalf("WriteScanReport error: %v", err)
	}

	// Expect exactly one file named like "ScanReport 2025-08-31 22:04:45.json"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}
	name := entries[0].Name()
	if filepath.Ext(name) != ".json" || name[:10] != "ScanReport" {
		t.Fatalf("unexpected filename: %s", name)
	}

	// Validate JSON content
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("readfile: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, string(b))
	}
	if obj["version"].(float64) != 1 {
		t.Fatalf("unexpected version in JSON: %v", obj["Version"])
	}
}
