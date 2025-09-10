package stdout

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mabd-dev/reposcan/pkg/report"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func sampleReport() report.ScanReport {
	return report.ScanReport{
		Version:     1,
		GeneratedAt: time.Date(2025, 8, 31, 22, 0, 0, 0, time.UTC),
		RepoStates: []report.RepoState{
			{Repo: "clean", Branch: "main", Path: "/tmp/clean"},
			{Repo: "dirty", Branch: "dev", Path: "/tmp/dirty", UncommitedFiles: []string{"a.txt"}},
		},
		Warnings: []string{"test warning"},
	}
}

func TestRenderScanReportAsJson_OutputsValidJSON(t *testing.T) {
	out := captureStdout(t, func() {
		if err := RenderScanReportAsJson(sampleReport()); err != nil {
			t.Fatalf("json render error: %v", err)
		}
	})
	// Verify it is JSON
	var v map[string]any
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if v["version"].(float64) != 1 {
		t.Fatalf("unexpected version: %v", v["Version"])
	}
}

func TestRenderScanReportAsTable_PrintsHeaderAndDetails(t *testing.T) {
	out := captureStdout(t, func() { RenderScanReportAsTable(sampleReport()) })
	if !strings.Contains(out, "Repo Scan Report") {
		t.Fatalf("missing header: %s", out)
	}
	if !strings.Contains(out, "Details:") {
		t.Fatalf("missing details section for dirty repos: %s", out)
	}
	if !strings.Contains(out, "dirty") || !strings.Contains(out, "/tmp/dirty") {
		t.Fatalf("missing repo details: %s", out)
	}
	if !strings.Contains(out, "test warning") {
		t.Fatalf("missing warnings: %s", out)
	}
}
