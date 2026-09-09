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

func TestIgnoreMatcher_SkipsEmptyAndCommentPatterns(t *testing.T) {
	// Blank, whitespace-only, and '#'-prefixed patterns are dropped.
	m := NewIgnoreMatcher([]string{"/proj"}, []string{"", "   ", "# comment", "node_modules"})

	if m.ShouldIgnore("/proj/something") {
		t.Fatalf("non-matching path must not be ignored")
	}
	if !m.ShouldIgnore("/proj/x/node_modules/y") {
		t.Fatalf("expected remaining valid pattern to still match")
	}
}

func TestIgnoreMatcher_NoPatternsIgnoresNothing(t *testing.T) {
	m := NewIgnoreMatcher([]string{"/proj"}, nil)
	if m.ShouldIgnore("/proj/vendor/lib") || m.ShouldIgnore("/proj") {
		t.Fatalf("empty pattern set must ignore nothing")
	}
}

// TestIgnoreMatcher_ShouldIgnoreDocumentedBranches locks in the reachable
// ShouldIgnore behavior. Note: the filepath.IsAbs(pat) branch (a Windows-only
// path like "C:/...") is not reachable on POSIX, where filepath.IsAbs and the
// anchored "/"-prefix branch are equivalent; that branch is covered by the
// Windows CI runner.
func TestIgnoreMatcher_ShouldIgnoreDocumentedBranches(t *testing.T) {
	m := NewIgnoreMatcher([]string{"/proj"}, []string{"**/build/**"})

	if !m.ShouldIgnore("/proj/cache/build/x") {
		t.Fatalf("expected build anywhere to be ignored")
	}
}
