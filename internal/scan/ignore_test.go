package scan

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

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
// anchored "/"-prefix branch are equivalent; that branch is covered by
// TestIgnoreMatcher_WindowsDrivePatternIsNotRootAnchored on the Windows CI
// runner.
func TestIgnoreMatcher_ShouldIgnoreDocumentedBranches(t *testing.T) {
	m := NewIgnoreMatcher([]string{"/proj"}, []string{"**/build/**"})

	if !m.ShouldIgnore("/proj/cache/build/x") {
		t.Fatalf("expected build anywhere to be ignored")
	}
}

// "~/" patterns are real paths under the home dir, not anchored to each root
// the way "/..." patterns are.
func TestIgnoreMatcher_TildePatternsMatchUnderHome(t *testing.T) {
	home := t.TempDir()
	setHome(t, home)

	m := NewIgnoreMatcher([]string{filepath.Join(home, "code")}, []string{"~/Library/**", "~/stuff/"})

	tests := []struct {
		path string
		want bool
	}{
		{path: filepath.Join(home, "Library"), want: true},
		{path: filepath.Join(home, "Library", "Caches", "x"), want: true},
		{path: filepath.Join(home, "stuff"), want: true},
		{path: filepath.Join(home, "stuff", "repo"), want: true},
		{path: filepath.Join(home, "code", "Library", "x"), want: false},
		{path: filepath.Join(home, "code", "app"), want: false},
	}
	for _, tt := range tests {
		if got := m.ShouldIgnore(tt.path); got != tt.want {
			t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// Without a home dir a "~/" pattern names no real path, so it is dropped: it
// is neither kept unexpanded nor expanded against an empty home. The other
// patterns still apply.
func TestIgnoreMatcher_TildePatternsDroppedWithoutHome(t *testing.T) {
	setHome(t, "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	if home, err := os.UserHomeDir(); err == nil {
		t.Skipf("home dir still resolves to %q with HOME unset", home)
	}

	m := NewIgnoreMatcher([]string{"/proj"}, []string{"~/Library/**", "~/stuff/", "node_modules"})

	tests := []struct {
		path string
		want bool
	}{
		// A "~/" pattern kept unexpanded would match these literal paths.
		{path: "~/Library", want: false},
		{path: "~/Library/Caches", want: false},
		{path: "~/stuff/repo", want: false},
		// An empty home would turn "~/Library/**" into "/Library/**".
		{path: "/Library/Caches", want: false},
		{path: "/stuff/repo", want: false},
		{path: "/proj/Library/Caches", want: false},
		// The remaining pattern is unaffected.
		{path: "/proj/app/node_modules", want: true},
		{path: "/proj/app/node_modules/pkg", want: true},
	}
	for _, tt := range tests {
		if got := m.ShouldIgnore(tt.path); got != tt.want {
			t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// A Windows drive-letter pattern is absolute but does not start with "/", so
// it is matched as-is rather than anchored to each root. On POSIX every
// absolute path starts with "/" and is root-anchored instead.
func TestIgnoreMatcher_WindowsDrivePatternIsNotRootAnchored(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("drive-letter paths are only absolute on Windows")
	}

	m := NewIgnoreMatcher([]string{`C:\proj`}, []string{`C:\proj\skip\**`})

	tests := []struct {
		path string
		want bool
	}{
		{path: `C:\proj\skip`, want: true},
		{path: `C:\proj\skip\repo`, want: true},
		{path: `C:\proj\skip\a\b`, want: true},
		{path: `C:\proj\keep`, want: false},
		{path: `C:\proj\skipped`, want: false},
		{path: `D:\proj\skip\repo`, want: false},
	}
	for _, tt := range tests {
		if got := m.ShouldIgnore(tt.path); got != tt.want {
			t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
