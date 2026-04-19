package analytics

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestStdoutAnalytics_Send_WritesSingleLine(t *testing.T) {
	var buf bytes.Buffer
	s := StdoutAnalytics{Writer: &buf}

	err := s.Send("usage", map[string]any{
		"os":      "linux",
		"arch":    "arm64",
		"version": "1.4.0",
	})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, `[analytics] event="usage" properties=`) {
		t.Fatalf("unexpected prefix: %q", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("line should end with newline, got %q", out)
	}
	if strings.Count(out, "\n") != 1 {
		t.Fatalf("expected exactly one line, got %d", strings.Count(out, "\n"))
	}
}

func TestStdoutAnalytics_Send_PropertiesAreValidJSON(t *testing.T) {
	var buf bytes.Buffer
	s := StdoutAnalytics{Writer: &buf}

	props := map[string]any{"os": "darwin", "cpu_count": 10}
	if err := s.Send("usage", props); err != nil {
		t.Fatalf("Send: %v", err)
	}

	// Extract the properties= payload and confirm it round-trips through JSON.
	_, jsonPart, ok := strings.Cut(buf.String(), "properties=")
	if !ok {
		t.Fatalf("no properties= segment in output: %q", buf.String())
	}
	jsonPart = strings.TrimSuffix(jsonPart, "\n")

	var got map[string]any
	if err := json.Unmarshal([]byte(jsonPart), &got); err != nil {
		t.Fatalf("properties not valid JSON: %v (%q)", err, jsonPart)
	}
	if got["os"] != "darwin" {
		t.Fatalf("expected os=darwin, got %v", got["os"])
	}
}

func TestStdoutAnalytics_Send_NilWriterUsesStdoutImplicitly(t *testing.T) {
	// The zero value StdoutAnalytics{} should not panic on Send — it falls
	// back to os.Stdout. We can't easily capture real stdout from a test, so
	// we just assert the call returns nil.
	if err := (StdoutAnalytics{}).Send("usage", nil); err != nil {
		t.Fatalf("Send with nil writer and nil props should not error, got %v", err)
	}
}

func TestNew_EmptyTokenReturnsStdoutAnalytics(t *testing.T) {
	a := New("", false)
	if _, ok := a.(StdoutAnalytics); !ok {
		t.Fatalf("expected StdoutAnalytics, got %T", a)
	}
}

func TestNew_DebugReturnsStdoutEvenWithToken(t *testing.T) {
	a := New("real-token", true)
	if _, ok := a.(StdoutAnalytics); !ok {
		t.Fatalf("expected StdoutAnalytics in debug mode, got %T", a)
	}
}

func TestNew_TokenAndNoDebugReturnsMixpanel(t *testing.T) {
	a := New("real-token", false)
	if _, ok := a.(*MixpanelAnalytics); !ok {
		t.Fatalf("expected *MixpanelAnalytics, got %T", a)
	}
}

func TestAnalytics_InterfaceIsSatisfied(t *testing.T) {
	// Compile-time: confirm both implementations satisfy the interface.
	// The var declarations at the bottom of analytics.go already assert
	// this, so this test is a belt-and-suspenders check that the tests
	// import the package correctly.
	var a Analytics
	a = StdoutAnalytics{}
	a = NewMixpanelAnalytics("x")
	_ = a
}
