package tui

import (
	"testing"

	"golang.design/x/clipboard"
)

// Runs wherever the clipboard cannot be initialised: CGO disabled, or a Linux
// host without an X display. Elsewhere copying would overwrite the user's real
// clipboard, so the test skips.
func TestCopyTextToClipboardReportsUnavailableClipboard(t *testing.T) {
	if err := clipboard.Init(); err == nil {
		t.Skip("system clipboard is available; copying would overwrite the user's clipboard")
	}

	if copyTextToClipboard("path") {
		t.Fatal("expected copyTextToClipboard to report false when the clipboard is unavailable")
	}
}
