package config

import "testing"

func TestPathsGetConfigFullPath_WhenUserHomeDirFails_ReturnsError(t *testing.T) {
	setHomeEnvironment(t, "")

	if _, err := DefaultPaths().GetConfigFullPath(); err == nil {
		t.Fatal("expected error when home lookup fails")
	}
}

func setHomeEnvironment(t *testing.T, value string) {
	t.Helper()
	t.Setenv("HOME", value)
	t.Setenv("USERPROFILE", value)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
}
