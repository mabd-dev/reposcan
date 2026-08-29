package utils

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func isWindows() bool {
	return runtime.GOOS == "windows"
}

// setHome points the OS home-directory lookup at dir. os.UserHomeDir()
// reads different env vars per platform: HOME on unix, USERPROFILE on
// Windows. Setting both keeps the tilde tests reproducible on every runner.
func setHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
}

// unsetHome removes the env vars os.UserHomeDir() relies on so the lookup
// fails on every platform.
func unsetHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
}

func TestFileExists(t *testing.T) {
	t.Run("existing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "f.txt")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		ok, err := FileExists(path)
		if err != nil || !ok {
			t.Fatalf("FileExists(%q) = (%v, %v), want (true, nil)", path, ok, err)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nope.txt")

		ok, err := FileExists(path)
		if err != nil || ok {
			t.Fatalf("FileExists(%q) = (%v, %v), want (false, nil)", path, ok, err)
		}
	})

	t.Run("stat error other than not-exist", func(t *testing.T) {
		if isWindows() {
			// On Windows, stat of "<file>\child" returns ERROR_PATH_NOT_FOUND,
			// which os.Stat surfaces as os.ErrNotExist rather than ENOTDIR.
			// The not-exist branch returns (false, nil), so the unix-only
			// premise of this test does not hold.
			t.Skip("ENOTDIR semantics are not exercised on Windows")
		}

		root := t.TempDir()
		file := filepath.Join(root, "f.txt")
		if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		// Using a regular file as a path component yields ENOTDIR.
		ok, err := FileExists(filepath.Join(file, "child"))
		if err == nil || ok {
			t.Fatalf("expected non-NotExist error, got (%v, %v)", ok, err)
		}
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected error other than ErrNotExist, got %v", err)
		}
	})
}

func TestDirExists(t *testing.T) {
	t.Run("existing dir", func(t *testing.T) {
		dir := t.TempDir()

		ok, err := DirExists(dir)
		if err != nil || !ok {
			t.Fatalf("DirExists(%q) = (%v, %v), want (true, nil)", dir, ok, err)
		}
	})

	t.Run("path is a file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "f.txt")
		if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		ok, err := DirExists(file)
		if err != nil || ok {
			t.Fatalf("DirExists(%q) = (%v, %v), want (false, nil)", file, ok, err)
		}
	})

	t.Run("missing dir", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "nope")

		ok, err := DirExists(dir)
		if err != nil || ok {
			t.Fatalf("DirExists(%q) = (%v, %v), want (false, nil)", dir, ok, err)
		}
	})

	t.Run("stat error other than not-exist", func(t *testing.T) {
		if isWindows() {
			t.Skip("ENOTDIR semantics are not exercised on Windows")
		}

		root := t.TempDir()
		file := filepath.Join(root, "f.txt")
		if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		ok, err := DirExists(filepath.Join(file, "child"))
		if err == nil || ok {
			t.Fatalf("expected non-NotExist error, got (%v, %v)", ok, err)
		}
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected error other than ErrNotExist, got %v", err)
		}
	})
}

func TestWriteToFile(t *testing.T) {
	t.Run("creates parent dirs and writes", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "a", "b", "out.txt")
		data := []byte("hello")

		if err := WriteToFile(data, path); err != nil {
			t.Fatalf("WriteToFile: %v", err)
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read written file: %v", err)
		}
		if string(got) != string(data) {
			t.Fatalf("expected content %q, got %q", data, got)
		}
	})

	t.Run("expands leading tilde using HOME", func(t *testing.T) {
		root := t.TempDir()
		setHome(t, root)

		relRel := filepath.Join("proj", "out.txt")
		if err := WriteToFile([]byte("x"), filepath.Join("~", relRel)); err != nil {
			t.Fatalf("WriteToFile: %v", err)
		}

		want := filepath.Join(root, relRel)
		got, err := os.ReadFile(want)
		if err != nil {
			t.Fatalf("expected file at %q: %v", want, err)
		}
		if string(got) != "x" {
			t.Fatalf("expected content %q, got %q", "x", got)
		}
	})

	t.Run("error when home cannot be determined", func(t *testing.T) {
		unsetHome(t)

		if err := WriteToFile([]byte("x"), "~/out.txt"); err == nil {
			t.Fatal("expected error when $HOME is not defined")
		}
	})

	t.Run("error when a parent component is a file", func(t *testing.T) {
		root := t.TempDir()
		blocker := filepath.Join(root, "blocker.txt")
		if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
			t.Fatalf("write blocker: %v", err)
		}

		// MkdirAll fails because "blocker.txt" exists as a regular file.
		if err := WriteToFile([]byte("x"), filepath.Join(blocker, "child", "out.txt")); err == nil {
			t.Fatal("expected error when a parent component is a regular file")
		}
	})
}
