package reposcan

import (
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal"
	"github.com/spf13/cobra"
)

func TestVersionCmd_PrintsVersion(t *testing.T) {
	out := captureStdout(t, func() { versionCmd.Run(versionCmd, nil) })

	if want := "reposcan " + internal.VERSION + "\n"; out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestAnalyticsCmd_DebugPrintsEvent(t *testing.T) {
	cmd := &cobra.Command{Use: "analytics-test"}
	cmd.Flags().Bool("debug-analytics", false, "")
	parseFlags(t, cmd, "--debug-analytics")

	var err error
	out := captureStdout(t, func() { err = analyticsCmd.RunE(cmd, nil) })
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	if !strings.Contains(out, `event="analytics_test"`) {
		t.Fatalf("expected the test event on stdout, got %q", out)
	}
}

func TestAnalyticsCmd_MissingFlagErrors(t *testing.T) {
	if err := analyticsCmd.RunE(&cobra.Command{Use: "analytics-test"}, nil); err == nil {
		t.Fatal("expected an error when debug-analytics is not defined")
	}
}
