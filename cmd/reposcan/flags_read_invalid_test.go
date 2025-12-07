package reposcan

import (
	"testing"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/spf13/cobra"
)

func TestReadFlags_InvalidOutput_Errors(t *testing.T) {
	cfg := config.Defaults()
	cmd := &cobra.Command{Use: "reposcan"}
	cmd.Flags().String("output", "table", "")
	cmd.SetArgs([]string{"--output", "yaml"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if err := readFlags(cmd, &cfg); err == nil {
		t.Fatalf("expected error for invalid output format")
	}
}

func TestReadFlags_InvalidFilter_Errors(t *testing.T) {
	cfg := config.Defaults()
	cmd := &cobra.Command{Use: "reposcan"}
	cmd.Flags().String("filter", "dirty", "")
	cmd.SetArgs([]string{"--filter", "staged"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}
	if err := readFlags(cmd, &cfg); err == nil {
		t.Fatalf("expected error for invalid filter value")
	}
}
