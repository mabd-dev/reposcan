//go:build !cgo && !windows

package tui

import "testing"

func TestCopyTextToClipboardIsUnavailableWithoutCGO(t *testing.T) {
	if copyTextToClipboard("path") {
		t.Fatal("expected clipboard to be unavailable without CGO")
	}
}
