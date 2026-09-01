package analytics

import (
	"bytes"
	"encoding/json"
	"errors"
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

// errWriter always fails writes so we can exercise the writer-error branch of
// Send.
type errWriter struct{}

func (errWriter) Write(p []byte) (int, error) {
	return 0, errors.New("boom")
}

func TestStdoutAnalytics_Send_WriterError(t *testing.T) {
	s := StdoutAnalytics{Writer: errWriter{}}

	err := s.Send("usage", map[string]any{"os": "linux"})
	if err == nil {
		t.Fatal("expected error when underlying writer fails")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected wrapped writer error, got %v", err)
	}
}

func TestStdoutAnalytics_Send_MarshalError(t *testing.T) {
	var buf bytes.Buffer
	s := StdoutAnalytics{Writer: &buf}

	// Values that cannot be represented in JSON (e.g. a channel) make
	// json.Marshal fail.
	err := s.Send("usage", map[string]any{"bad": make(chan int)})
	if err == nil {
		t.Fatal("expected error when properties cannot be marshaled")
	}
	if !strings.Contains(err.Error(), "marshal properties") {
		t.Fatalf("expected marshal error to be wrapped, got %v", err)
	}
	// Nothing should have been written on the marshal path.
	if buf.Len() != 0 {
		t.Fatalf("expected no output on marshal error, got %q", buf.String())
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
