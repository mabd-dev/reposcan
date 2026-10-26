package reposcan

import (
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal/config"
)

func TestReadFlags_InvalidValues_Error(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "output", args: []string{"--output", "yaml"}, wantErr: "yaml"},
		{name: "filter", args: []string{"--filter", "staged"}, wantErr: "staged"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Defaults()
			cmd := newRootFlagsCmd("")
			parseFlags(t, cmd, tt.args...)

			err := readFlags(cmd, &cfg)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error mentioning %q, got %v", tt.wantErr, err)
			}
		})
	}
}

// Each flag lookup returns an error when the command does not define the flag.
func TestReadFlags_UndefinedFlag_Errors(t *testing.T) {
	flags := []string{
		"root", "dirIgnore", "output", "filter", "stale-days",
		"json-output-path", "max-workers", "debug", "no-telemetry",
	}

	for _, flag := range flags {
		t.Run(flag, func(t *testing.T) {
			cfg := config.Defaults()
			cmd := newRootFlagsCmd(flag)

			err := readFlags(cmd, &cfg)
			if err == nil || !strings.Contains(err.Error(), flag) {
				t.Fatalf("expected error naming %q, got %v", flag, err)
			}
		})
	}
}
