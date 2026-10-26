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

// newRootFlagsCmd returns a throwaway command with the same flags RootCmd
// registers, minus the flag named in skip. RootCmd's own flag defaults come
// from the developer's real config file, so tests must not rely on them.
func newRootFlagsCmd(skip string) *cobra.Command {
	cmd := &cobra.Command{Use: "reposcan"}
	flags := map[string]func(){
		"root":             func() { cmd.Flags().StringArrayP("root", "r", nil, "") },
		"dirIgnore":        func() { cmd.Flags().StringArrayP("dirIgnore", "d", nil, "") },
		"output":           func() { cmd.Flags().StringP("output", "o", "json", "") },
		"filter":           func() { cmd.Flags().StringP("filter", "f", "all", "") },
		"stale-days":       func() { cmd.Flags().Int("stale-days", 0, "") },
		"json-output-path": func() { cmd.Flags().String("json-output-path", "", "") },
		"max-workers":      func() { cmd.Flags().IntP("max-workers", "w", 1, "") },
		"debug":            func() { cmd.Flags().Bool("debug", false, "") },
		"no-telemetry":     func() { cmd.Flags().Bool("no-telemetry", true, "") },
	}
	for name, register := range flags {
		if name != skip {
			register()
		}
	}
	return cmd
}

func parseFlags(t *testing.T, cmd *cobra.Command, args ...string) {
	t.Helper()
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("parse flags %v: %v", args, err)
	}
}
