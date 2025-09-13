package scan

import "testing"

func TestIgnoreMatcher_BasicPatterns(t *testing.T) {
	roots := []string{"/proj"}
	m := NewIgnoreMatcher(roots, []string{"**/node_modules/**", "/vendor/**", "**/cache/", "config"})

	// anywhere match
	if !m.ShouldIgnore("/proj/a/node_modules/pkg") {
		t.Fatalf("expected to ignore node_modules")
	}
	// anchored to root
	if !m.ShouldIgnore("/proj/vendor/lib") {
		t.Fatalf("expected to ignore /vendor under root")
	}
	if m.ShouldIgnore("/other/vendor/lib") {
		t.Fatalf("should not ignore /vendor outside roots")
	}
	// token expanded to **/cache/**
	if !m.ShouldIgnore("/proj/tmp/cache/data") {
		t.Fatalf("expected to ignore token cache/")
	}

	if !m.ShouldIgnore("/proj/tmp/config/data") {
		t.Fatalf("expected to ignore token config/")
	}
}
