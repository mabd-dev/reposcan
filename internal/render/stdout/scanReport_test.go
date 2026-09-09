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
			{Repo: "clean", VCSType: "git", Branch: "main", Path: "/tmp/clean"},
			{Repo: "dirty", VCSType: "git", Branch: "dev", Path: "/tmp/dirty", UncommitedFiles: []string{"a.txt"}},
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

func TestRenderScanReportAsJson_MarshalError(t *testing.T) {
	// A GeneratedAt year outside [0,9999] makes json.MarshalIndent fail.
	r := report.ScanReport{GeneratedAt: time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)}

	err := RenderScanReportAsJson(r)
	if err == nil {
		t.Fatal("expected error when report cannot be marshaled")
	}
	if !strings.Contains(err.Error(), "Error convert report to json") {
		t.Fatalf("expected conversion error, got %v", err)
	}
}

func TestRenderReportHeader_WithDirtyRepos(t *testing.T) {
	out := captureStdout(t, func() {
		renderReportHeader(sampleReport(), 5, 2)
	})

	if !strings.Contains(out, "Repo Scan Report") {
		t.Fatalf("expected title, got %q", out)
	}
	if !strings.Contains(out, "Generated at:") {
		t.Fatalf("expected generated-at line, got %q", out)
	}
	if !strings.Contains(out, "Total repositories:") {
		t.Fatalf("expected total-repos line, got %q", out)
	}
}

func TestRenderReportHeader_NoDirtyRepos(t *testing.T) {
	out := captureStdout(t, func() {
		renderReportHeader(sampleReport(), 3, 0)
	})

	if !strings.Contains(out, "Total repositories:") {
		t.Fatalf("expected total-repos line, got %q", out)
	}
}

func TestTruncateRunes(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want string
	}{
		{in: "hello", n: 0, want: ""},
		{in: "hello", n: -1, want: ""},
		{in: "hey", n: 3, want: "hey"},
		{in: "hello world", n: 5, want: "he..."},
		{in: "abcdef", n: 3, want: "abc"},
		{in: "ééééé", n: 4, want: "é..."},
	}

	for _, tc := range cases {
		if got := truncateRunes(tc.in, tc.n); got != tc.want {
			t.Fatalf("truncateRunes(%q, %d) = %q, want %q", tc.in, tc.n, got, tc.want)
		}
	}
}

func TestWarnings(t *testing.T) {
	out := captureStdout(t, func() {
		Warnings([]string{"w1", "w2"})
	})

	if !strings.Contains(out, "Warning:") || !strings.Contains(out, "w1") || !strings.Contains(out, "w2") {
		t.Fatalf("expected warning lines, got %q", out)
	}
}

func TestWarning(t *testing.T) {
	out := captureStdout(t, func() {
		Warning("single warning")
	})

	if !strings.Contains(out, "Warning:") || !strings.Contains(out, "single warning") {
		t.Fatalf("expected single warning line, got %q", out)
	}
}

func TestError(t *testing.T) {
	out := captureStdout(t, func() {
		Error("something failed")
	})

	if !strings.Contains(out, "Error:") || !strings.Contains(out, "something failed") {
		t.Fatalf("expected error line, got %q", out)
	}
}
