package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mixpanel "github.com/mixpanel/mixpanel-go"
)

// newTestMixpanel returns a MixpanelAnalytics wired to a self-hosted HTTP
// handler via the SDK's ProxyApiLocation + HttpClient options, so Send runs
// against localhost instead of the real Mixpanel API.
func newTestMixpanel(t *testing.T, handler http.HandlerFunc) *MixpanelAnalytics {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return NewMixpanelAnalyticsWithClient(
		"test-token",
		mixpanel.HttpClient(server.Client()),
		mixpanel.ProxyApiLocation(server.URL),
	)
}

// TestMixpanelAnalytics_Send_ServerRoundTrip runs Send against a real local
// httptest server so we assert both the request payload and that a
// non-erroring response is accepted.
func TestMixpanelAnalytics_Send_ServerRoundTrip(t *testing.T) {
	var requests [][]map[string]any
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/track" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		var events []map[string]any
		if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		requests = append(requests, events)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":1,"error":""}`))
	}

	m := newTestMixpanel(t, handler)

	if err := m.Send("usage", map[string]any{"os": "darwin", "cpu": 10}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}
	events := requests[0]
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ev := events[0]
	if ev["event"] != "usage" {
		t.Fatalf("expected event name %q, got %v", "usage", ev["event"])
	}
	props, ok := ev["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties object, got %#v", ev["properties"])
	}
	if props["os"] != "darwin" {
		t.Fatalf("expected os=darwin, got %v", props["os"])
	}
	if props["distinct_id"] != "anonymous" {
		t.Fatalf("expected distinct_id=anonymous, got %v", props["distinct_id"])
	}
	if props["token"] != "test-token" {
		t.Fatalf("expected token=test-token, got %v", props["token"])
	}
}

func TestMixpanelAnalytics_Send_BackendError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":0,"error":"invalid token"}`))
	}

	m := newTestMixpanel(t, handler)

	err := m.Send("usage", nil)
	if err == nil {
		t.Fatal("expected error from backend rejection")
	}
	if !strings.Contains(err.Error(), "invalid token") {
		t.Fatalf("expected error to surface backend message, got %v", err)
	}
}

func TestMixpanelAnalytics_Send_ServerError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}

	m := newTestMixpanel(t, handler)

	if err := m.Send("usage", nil); err == nil {
		t.Fatal("expected error on HTTP 500")
	}
}
