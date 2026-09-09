//go:build !windows

// os.Symlink to a dangling target is used to simulate a .git entry that is a
// non-directory yet unreadable, which forces isGitRepo's ReadFile error path.
// Symlink creation is treated as a filesystem-level capability, so this file
// is restricted to POSIX platforms.
package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGitRepo_UnreadableGitFileReturnsFalse(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// A .git symlink pointing at a nonexistent target passes os.Lstat as a
	// non-directory entry but os.ReadFile fails on it.
	if err := os.Symlink("/nonexistent/.git-target", filepath.Join(repo, ".git")); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	if isGitRepo(repo) {
		t.Fatalf("isGitRepo should be false when .git exists but cannot be read")
	}
}
