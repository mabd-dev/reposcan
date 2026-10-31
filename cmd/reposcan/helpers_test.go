package reposcan

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// isolateHome points every home/config lookup at a temp dir so tests never
// read or write the real ~/.config/reposcan. USERPROFILE and AppData cover
// os.UserHomeDir and os.UserConfigDir on Windows.
func isolateHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("AppData", filepath.Join(home, "AppData"))
	return home
}

// captureStdout returns everything fn writes to os.Stdout. The deferred
// closes also release the pipe and drain goroutine if fn exits via t.Fatal.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()

	// Drain concurrently so a large write cannot block on a full pipe buffer.
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	_ = w.Close()
	return <-done
}

// newRootFlagsCmd returns a throwaway command with the flags RootCmd
// registers, so tests can parse flags without touching RootCmd's state.
func newRootFlagsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "reposcan"}
	addRootFlags(cmd)
	return cmd
}

// writeConfig writes contents to the config.toml under an isolated home.
func writeConfig(t *testing.T, home, contents string) {
	t.Helper()
	path := filepath.Join(home, ".config", "reposcan", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func parseFlags(t *testing.T, cmd *cobra.Command, args ...string) {
	t.Helper()
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("parse flags %v: %v", args, err)
	}
}
