package reposcan

import (
	"runtime"

	"github.com/mabd-dev/reposcan/internal"
	"github.com/mabd-dev/reposcan/internal/analytics"
	"github.com/spf13/cobra"
)

// mixpanelToken is the Mixpanel project token. It is intentionally declared
// as a package-level var (not const) so it can be injected at build time via
// ldflags:
//
//	go build -ldflags="-X github.com/mabd-dev/reposcan/cmd/reposcan.mixpanelToken=REAL_TOKEN"
//
// When empty — the default for local dev builds and the binaries distributed
// through `go install` — the analytics factory returns StdoutAnalytics and
// never touches the network.
var mixpanelToken string

// newAnalytics returns the Analytics implementation this CLI should use.
// Callers pass the same debug flag that already gates logger verbosity; when
// debug is on, events go to stdout regardless of whether a token is wired in.
func newAnalytics(debug bool) analytics.Analytics {
	return analytics.New(mixpanelToken, debug)
}

// analyticsCmd is a hidden maintainer-facing command that emits a single
// test event through the analytics layer. It exists to satisfy the #24
// acceptance criteria ("Send a dummy event in debug mode to verify the
// integration end-to-end") without wiring any real events into the normal
// scan flow — that is the telemetry issue's job (#25).
var analyticsCmd = &cobra.Command{
	Use:    "analytics-test",
	Short:  "Emit a single test event through the analytics layer",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		debug, err := cmd.Flags().GetBool("debug-analytics")
		if err != nil {
			return err
		}
		a := newAnalytics(debug)
		return a.Send("analytics_test", map[string]any{
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
			"version": internal.VERSION,
		})
	},
}

func init() {
	analyticsCmd.Flags().Bool(
		"debug-analytics",
		false,
		"Force the stdout analytics backend and print events locally instead of sending them",
	)
	RootCmd.AddCommand(analyticsCmd)
}
