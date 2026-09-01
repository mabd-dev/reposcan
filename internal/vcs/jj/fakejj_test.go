package jj

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const fakeJJResponsesEnv = "REPOSCAN_TEST_JJ_RESPONSES"

type fakeJJResponse struct {
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exitCode,omitempty"`
}

// TestMain runs the package tests normally unless the test binary was re-executed
// as a fake jj command. Fake invocations skip "-R <repo>" and match the remaining
// arguments against responses configured by useFakeJJ.
func TestMain(m *testing.M) {
	responsesJSON := os.Getenv(fakeJJResponsesEnv)
	if responsesJSON == "" {
		os.Exit(m.Run())
	}

	var responses map[string]fakeJJResponse
	if err := json.Unmarshal([]byte(responsesJSON), &responses); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "decode fake jj responses: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 3 || os.Args[1] != "-R" {
		_, _ = fmt.Fprintf(os.Stderr, "unexpected fake jj arguments: %q\n", os.Args[1:])
		os.Exit(1)
	}

	key := fakeJJCommandKey(os.Args[3:]...)
	response, ok := responses[key]
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "no fake jj response for arguments: %q\n", os.Args[3:])
		os.Exit(1)
	}

	_, _ = os.Stdout.WriteString(response.Stdout)
	_, _ = os.Stderr.WriteString(response.Stderr)
	os.Exit(response.ExitCode)
}

// fakeJJCommandKey preserves argument boundaries and normalizes empty lists.
func fakeJJCommandKey(args ...string) string {
	if args == nil {
		args = []string{}
	}

	key, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}

	return string(key)
}

// useFakeJJ configures the current test binary to respond as jj when a unit
// under test executes it directly.
func useFakeJJ(t *testing.T, responses map[string]fakeJJResponse) string {
	t.Helper()

	responsesJSON, err := json.Marshal(responses)
	if err != nil {
		t.Fatalf("encode fake jj responses: %v", err)
	}
	t.Setenv(fakeJJResponsesEnv, string(responsesJSON))

	// Race-enabled child processes wait one second before exiting by default.
	// Remove that wait so every fake jj command does not add a second to the test.
	if gorace := os.Getenv("GORACE"); !strings.Contains(gorace, "atexit_sleep_ms=") {
		t.Setenv("GORACE", strings.TrimSpace(gorace+" atexit_sleep_ms=0"))
	}

	return os.Args[0]
}

// installFakeJJ places a platform-specific copy of the test binary on PATH for
// exported wrappers and Provider.Fetch, which currently hardcode "jj".
func installFakeJJ(t *testing.T, responses map[string]fakeJJResponse) {
	t.Helper()

	useFakeJJ(t, responses)

	binDir := t.TempDir()
	testBinary, err := filepath.Abs(os.Args[0])
	if err != nil {
		t.Fatalf("resolve test binary path: %v", err)
	}

	fakeBinaryName := "jj"
	if runtime.GOOS == "windows" {
		fakeBinaryName += ".exe"
	}
	fakeBinaryPath := filepath.Join(binDir, fakeBinaryName)

	source, err := os.Open(testBinary)
	if err != nil {
		t.Fatalf("open test binary: %v", err)
	}
	defer source.Close()

	destination, err := os.OpenFile(fakeBinaryPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		t.Fatalf("create fake jj: %v", err)
	}
	if _, err := io.Copy(destination, source); err != nil {
		_ = destination.Close()
		t.Fatalf("copy fake jj: %v", err)
	}
	if err := destination.Close(); err != nil {
		t.Fatalf("install fake jj: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}
