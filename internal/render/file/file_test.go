package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

// TestWriteScanReport_WriteErrorCoverage exercises the utils.WriteToFile
// error path by pointing dirPath at a location that cannot be created
// (a regular file used as a parent directory yields ENOTDIR).
func TestWriteScanReport_WriteErrorCoverage(t *testing.T) {
	root := t.TempDir()
	blocker := filepath.Join(root, "blocker.txt")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}

	r := report.ScanReport{GeneratedAt: time.Now()}
	// dirPath has a regular-file component, so creating the report dir fails.
	err := WriteScanReport(r, filepath.Join(blocker, "child"))
	if err == nil {
		t.Fatal("expected error when target directory cannot be created")
	}
}

// TestWriteScanReport_MarshalError verifies the JSON marshal error path: a
// time.Time whose year exceeds 9999 cannot be JSON-encoded, so WriteScanReport
// must return that error instead of reporting success with a zero-byte file.
func TestWriteScanReport_MarshalError(t *testing.T) {
	dir := t.TempDir()
	r := report.ScanReport{
		GeneratedAt: time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	err := WriteScanReport(r, dir)
	if err == nil {
		t.Fatal("expected error when GeneratedAt year is out of JSON range")
	}
	if !strings.Contains(err.Error(), "Error convert report to json") {
		t.Fatalf("expected json conversion error, got %q", err)
	}

	// No file must have been written.
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("readdir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no file written on marshal error, got %v", entries)
	}
}

// TestWriteScanReport_ExpandsDirPath checks the report lands in the folder the
// user meant. dirPath is expanded exactly once, before the filename is joined:
// filepath.Join would clean "~/.." or "$VAR/.." away, and a second expansion
// would mangle an env value that contains a literal "$".
func TestWriteScanReport_ExpandsDirPath(t *testing.T) {
	base := t.TempDir()
	home := filepath.Join(base, "home")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // os.UserHomeDir on Windows
	t.Setenv("REPOSCAN_TEST_REPORTS", filepath.Join(base, "reports"))
	t.Setenv("REPOSCAN_TEST_INNER", "inner")
	t.Setenv("REPOSCAN_TEST_LITERAL", filepath.Join(base, "dev$REPOSCAN_TEST_INNER"))

	tests := []struct {
		dirPath string
		want    string // relative to base
	}{
		{dirPath: "$REPOSCAN_TEST_REPORTS", want: "reports"},
		{dirPath: "$REPOSCAN_TEST_REPORTS/../out", want: "out"},
		{dirPath: "~/reports", want: "home/reports"},
		{dirPath: "~/../out", want: "out"},
		{dirPath: "$REPOSCAN_TEST_LITERAL", want: "dev$REPOSCAN_TEST_INNER"},
	}

	for _, tt := range tests {
		t.Run(tt.dirPath, func(t *testing.T) {
			// A path that loses its prefix becomes relative to the working
			// dir, so run in an empty one and check nothing lands there.
			cwd := t.TempDir()
			t.Chdir(cwd)
			want := filepath.Join(base, filepath.FromSlash(tt.want))
			t.Cleanup(func() { os.RemoveAll(want) })

			r := report.ScanReport{Version: 1, GeneratedAt: time.Date(2025, 8, 31, 22, 4, 45, 0, time.UTC)}
			if err := WriteScanReport(r, tt.dirPath); err != nil {
				t.Fatalf("WriteScanReport(%q) error: %v", tt.dirPath, err)
			}

			if entries, err := os.ReadDir(want); err != nil || len(entries) != 1 {
				t.Fatalf("expected 1 report in %s, got %d entries (err=%v)", want, len(entries), err)
			}
			if entries, _ := os.ReadDir(cwd); len(entries) != 0 {
				t.Fatalf("expected nothing written to the working dir, found %v", entries)
			}
		})
	}
}

func TestWriteScanReport_ErrorWhenHomeCannotBeDetermined(t *testing.T) {
	for _, key := range []string{"HOME", "USERPROFILE", "HOMEDRIVE", "HOMEPATH"} {
		t.Setenv(key, "")
	}
	t.Chdir(t.TempDir()) // nothing may be written, but keep the package dir clean

	r := report.ScanReport{Version: 1, GeneratedAt: time.Now()}
	if err := WriteScanReport(r, "~/reports"); err == nil {
		t.Fatal("expected an error when the home directory cannot be determined")
	}
}
