// Package analytics provides a thin abstraction over anonymous usage events.
//
// Callers depend only on the Analytics interface, so the rest of the codebase
// never imports a concrete SDK. Two implementations are provided:
//
//   - MixpanelAnalytics  sends events via the official Mixpanel Go SDK.
//   - StdoutAnalytics    prints events to stdout as JSON for local / CI use.
//
// The factory New() picks between them based on whether a build-time token is
// present and whether the caller asked for debug mode.
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	mixpanel "github.com/mixpanel/mixpanel-go"
)

// Analytics is the single entry point for emitting a usage event. A nil
// implementation is never returned by New; callers can rely on Send never
// panicking.
type Analytics interface {
	Send(event string, properties map[string]any) error
}

// StdoutAnalytics prints each event to an io.Writer (stdout by default) as a
// single human-readable line. It is the zero-configuration implementation
// used for local development, CI, and when no Mixpanel token is wired in.
type StdoutAnalytics struct {
	// Writer is where events are printed. Nil means os.Stdout.
	Writer io.Writer
}

// Send formats the event and writes it as a single line. Errors from the
// underlying writer are returned unchanged.
func (s StdoutAnalytics) Send(event string, properties map[string]any) error {
	w := s.Writer
	if w == nil {
		w = os.Stdout
	}
	payload, err := json.Marshal(properties)
	if err != nil {
		return fmt.Errorf("analytics: marshal properties: %w", err)
	}
	if _, err := fmt.Fprintf(w, "[analytics] event=%q properties=%s\n", event, payload); err != nil {
		return fmt.Errorf("analytics: write event: %w", err)
	}
	return nil
}

// MixpanelAnalytics sends events to the Mixpanel HTTP API via the official
// mixpanel/mixpanel-go SDK. It uses a distinctID of "anonymous" because
// reposcan does not have a persistent user identity (that is handled in the
// telemetry layer built on top of this).
type MixpanelAnalytics struct {
	client     *mixpanel.ApiClient
	distinctID string
}

// NewMixpanelAnalytics constructs a live client against the Mixpanel API.
// The token is required; passing an empty token is a caller bug.
func NewMixpanelAnalytics(token string) *MixpanelAnalytics {
	return &MixpanelAnalytics{
		client:     mixpanel.NewApiClient(token),
		distinctID: "anonymous",
	}
}

// Send forwards the event to Mixpanel. Failures are returned to the caller
// verbatim - it is the caller's responsibility to decide whether telemetry
// failures should be logged or swallowed.
func (m *MixpanelAnalytics) Send(event string, properties map[string]any) error {
	ev := m.client.NewEvent(event, m.distinctID, properties)
	if err := m.client.Track(context.Background(), []*mixpanel.Event{ev}); err != nil {
		return fmt.Errorf("analytics: mixpanel track: %w", err)
	}
	return nil
}

// compile-time interface checks
var (
	_ Analytics = StdoutAnalytics{}
	_ Analytics = (*MixpanelAnalytics)(nil)
)

// New picks the implementation based on build- and run-time inputs.
//
//   - If debug is true, events go to stdout - even when a real token is wired
//     in. This keeps local development loops noisy and predictable and mirrors
//     the behavior expected by CI harnesses.
//   - If token is empty (a local dev build without the -ldflags injection),
//     fall back silently to stdout. This means no-one has to set up a Mixpanel
//     project just to run the CLI.
//   - Otherwise, return a live MixpanelAnalytics.
func New(token string, debug bool) Analytics {
	if debug || token == "" {
		return StdoutAnalytics{}
	}
	return NewMixpanelAnalytics(token)
}
