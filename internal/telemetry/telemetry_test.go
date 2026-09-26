package telemetry

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal/analytics"
	"github.com/mabd-dev/reposcan/internal/config"
)

// mockAnalytics records what was passed to Send.
type mockAnalytics struct {
	event      string
	properties map[string]any
	err        error
}

func (m *mockAnalytics) Send(event string, properties map[string]any) error {
	m.event = event
	m.properties = properties
	return m.err
}

// setConfigDir overrides userConfigDir for the duration of the test.
func setConfigDir(t *testing.T, dir string) {
	t.Helper()
	orig := userConfigDir
	userConfigDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { userConfigDir = orig })
}

// setConfigDirError overrides userConfigDir to return an error.
func setConfigDirError(t *testing.T) {
	t.Helper()
	orig := userConfigDir
	userConfigDir = func() (string, error) { return "", errors.New("no config dir") }
	t.Cleanup(func() { userConfigDir = orig })
}

// clearHomeEnv unsets every environment variable os.UserHomeDir consults so it
// reports an error. HOME is used on Unix; Windows uses USERPROFILE and falls
// back to HOMEDRIVE+HOMEPATH (both set on CI runners).
func clearHomeEnv(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
}

// setAnalytics overrides newAnalyticsService for the duration of the test.
func setAnalytics(t *testing.T, mock *mockAnalytics) {
	t.Helper()
	orig := newAnalyticsService
	newAnalyticsService = func(token string, debug bool) analytics.Analytics { return mock }
	t.Cleanup(func() { newAnalyticsService = orig })
}

// captureStdout redirects the package-level stdout to a buffer.
func captureStdout(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	orig := stdout
	stdout = &buf
	t.Cleanup(func() { stdout = orig })
	return &buf
}

// writeTelemetryFile writes a Telemetry struct as JSON at the given path.
func writeTelemetryFile(t *testing.T, path string, tel Telemetry) {
	t.Helper()
	data, err := json.MarshalIndent(tel, "", "    ")
	if err != nil {
		t.Fatalf("marshal telemetry: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write telemetry file: %v", err)
	}
}

// ---- expandPath ----

func TestExpandPath_TildeExpandsToHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}
	got, err := expandPath("~/foo/bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, "foo/bar")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExpandPath_TildeOnly_ExpandsToHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home dir: %v", err)
	}
	got, err := expandPath("~")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != home {
		t.Errorf("got %q, want %q", got, home)
	}
}

func TestExpandPath_AbsolutePath_Unchanged(t *testing.T) {
	path := "/some/absolute/path"
	got, err := expandPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != path {
		t.Errorf("got %q, want %q", got, path)
	}
}

func TestExpandPath_RelativePath_Unchanged(t *testing.T) {
	path := "relative/path/file.txt"
	got, err := expandPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != path {
		t.Errorf("got %q, want %q", got, path)
	}
}

// ---- fileExists ----

func TestFileExists_ExistingFile_ReturnsTrue(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "telemetry-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	exists, err := fileExists(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected true for existing file, got false")
	}
}

func TestFileExists_NonExistentFile_ReturnsFalse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-such-file.json")
	exists, err := fileExists(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false for non-existent file, got true")
	}
}

func TestFileExists_Directory_ReturnsTrue(t *testing.T) {
	dir := t.TempDir()
	exists, err := fileExists(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected true for directory, got false")
	}
}

// ---- readTelemetry ----

func TestReadTelemetry_ValidFile_ReturnsTelemetry(t *testing.T) {
	want := Telemetry{UUID: "test-uuid-1234", Warned: true}
	path := filepath.Join(t.TempDir(), "telemetry.json")
	writeTelemetryFile(t, path, want)

	got, err := readTelemetry(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestReadTelemetry_NonExistentFile_ReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-such.json")
	_, err := readTelemetry(path)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestReadTelemetry_InvalidJSON_ReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(path, []byte("not-valid-json{{{"), 0o644)

	_, err := readTelemetry(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// ---- writeTelemetry ----

func TestWriteTelemetry_CreatesFileWithCorrectContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "telemetry.json")
	want := Telemetry{UUID: "abc-123", Warned: true}

	if err := writeTelemetry(path, want); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := readTelemetry(path)
	if err != nil {
		t.Fatalf("readTelemetry after write: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestWriteTelemetry_NonExistentDirectory_ReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "subdir", "telemetry.json")
	err := writeTelemetry(path, Telemetry{UUID: "x"})
	if err == nil {
		t.Fatal("expected error for non-existent parent directory, got nil")
	}
}

// ---- getOrCreateTelemetry ----

func TestGetOrCreateTelemetry_ExistingFile_ReturnsTelemetry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "telemetry.json")
	want := Telemetry{UUID: "existing-uuid", Warned: true}
	writeTelemetryFile(t, path, want)

	got, err := getOrCreateTelemetry(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGetOrCreateTelemetry_FileNotExists_CreatesNewWithUUID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "telemetry.json")

	got, err := getOrCreateTelemetry(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UUID == "" {
		t.Error("expected non-empty UUID for new telemetry")
	}
	if got.Warned {
		t.Error("new telemetry should have Warned=false")
	}
	if exists, _ := fileExists(path); !exists {
		t.Error("telemetry file should have been written to disk")
	}
}

func TestGetOrCreateTelemetry_TwoCalls_ProduceSameUUID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "telemetry.json")

	first, _ := getOrCreateTelemetry(path)
	second, _ := getOrCreateTelemetry(path)
	if first.UUID != second.UUID {
		t.Errorf("UUID mismatch: %q vs %q", first.UUID, second.UUID)
	}
}

func TestGetOrCreateTelemetry_InvalidJSON_ReturnsEmptyTelemetry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "telemetry.json")
	os.WriteFile(path, []byte("bad json{{"), 0o644)

	got, err := getOrCreateTelemetry(path)
	if err == nil {
		t.Fatal("Error expected, but found nil")
	}
	if got != (Telemetry{}) {
		t.Errorf("expected empty Telemetry on bad JSON, got %+v", got)
	}
}

// ---- getTelemetryFilePath ----

func TestGetTelemetryFilePath_Success_ReturnsPathUnderToolSubdir(t *testing.T) {
	tmpDir := t.TempDir()
	setConfigDir(t, tmpDir)

	path, err := getTelemetryFilePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedDir := filepath.Join(tmpDir, toolName)
	if !strings.HasPrefix(path, expectedDir) {
		t.Errorf("path %q should be under %q", path, expectedDir)
	}
	if !strings.HasSuffix(path, telemtryFileName) {
		t.Errorf("path %q should end with %q", path, telemtryFileName)
	}
}

func TestGetTelemetryFilePath_DirAlreadyExists_Succeeds(t *testing.T) {
	tmpDir := t.TempDir()
	// pre-create the subdir to verify MkdirAll doesn't error on existing dirs
	os.Mkdir(filepath.Join(tmpDir, toolName), 0755)
	setConfigDir(t, tmpDir)

	_, err := getTelemetryFilePath()
	if err != nil {
		t.Fatalf("expected success when dir already exists, got: %v", err)
	}
}

func TestGetTelemetryFilePath_ConfigDirError_ReturnsError(t *testing.T) {
	setConfigDirError(t)

	_, err := getTelemetryFilePath()
	if err == nil {
		t.Fatal("expected error when config dir unavailable, got nil")
	}
}

// ---- Send ----

func TestSend_CIEnvironment_SendsAnalytics(t *testing.T) {
	t.Setenv("CI", "true")
	mock := &mockAnalytics{}
	setAnalytics(t, mock)

	Send("token", false, config.OnlyAll, config.OutputTable, 1)

	if mock.event == "" {
		t.Error("analytics.Send should be called in CI")
	}
}

func TestSend_FilePathError_SkipsAnalytics(t *testing.T) {
	t.Setenv("CI", "")
	setConfigDirError(t)
	mock := &mockAnalytics{}
	setAnalytics(t, mock)

	Send("token", false, config.OnlyAll, config.OutputTable, 1)

	if mock.event != "" {
		t.Error("analytics.Send should not be called when file path resolution fails")
	}
}

func TestSend_FirstRun_PrintsWarningAndPersistsWarnedTrue(t *testing.T) {
	t.Setenv("CI", "")
	tmpDir := t.TempDir()
	setConfigDir(t, tmpDir)
	mock := &mockAnalytics{}
	setAnalytics(t, mock)
	buf := captureStdout(t)

	Send("token", false, config.OnlyAll, config.OutputTable, 2)

	if !strings.Contains(buf.String(), "anonymous usage telemetry") {
		t.Errorf("expected telemetry warning in output, got: %q", buf.String())
	}
	filePath := filepath.Join(tmpDir, toolName, telemtryFileName)
	tel, err := readTelemetry(filePath)
	if err != nil {
		t.Fatalf("readTelemetry: %v", err)
	}
	if !tel.Warned {
		t.Error("Warned should be true after first Send")
	}
}

func TestSend_AlreadyWarned_NoPrintNoFileUpdate(t *testing.T) {
	t.Setenv("CI", "")
	tmpDir := t.TempDir()
	setConfigDir(t, tmpDir)

	// pre-create the dir + file with Warned=true
	dir := filepath.Join(tmpDir, toolName)
	os.MkdirAll(dir, 0755)
	writeTelemetryFile(t, filepath.Join(dir, telemtryFileName), Telemetry{UUID: "pre", Warned: true})

	mock := &mockAnalytics{}
	setAnalytics(t, mock)
	buf := captureStdout(t)

	Send("token", false, config.OnlyAll, config.OutputTable, 1)

	if strings.Contains(buf.String(), "anonymous usage telemetry") {
		t.Error("warning should not be printed when already warned")
	}
}

func TestSend_AnalyticsCalledWithUsageEvent(t *testing.T) {
	t.Setenv("CI", "")
	tmpDir := t.TempDir()
	setConfigDir(t, tmpDir)
	mock := &mockAnalytics{}
	setAnalytics(t, mock)
	captureStdout(t)

	Send("token", false, config.OnlyAll, config.OutputTable, 3)

	if mock.event != "usage" {
		t.Errorf("expected event 'usage', got %q", mock.event)
	}
	if mock.properties["repo_count"] != 3 {
		t.Errorf("expected repo_count=3, got %v", mock.properties["repo_count"])
	}
	if mock.properties["project"] != toolName {
		t.Errorf("expected project=%q, got %v", toolName, mock.properties["project"])
	}
}

func TestSend_AnalyticsError_DoesNotPanic(t *testing.T) {
	t.Setenv("CI", "")
	tmpDir := t.TempDir()
	setConfigDir(t, tmpDir)
	mock := &mockAnalytics{err: errors.New("network error")}
	setAnalytics(t, mock)
	captureStdout(t)

	// should complete without panic
	Send("token", false, config.OnlyAll, config.OutputTable, 1)
}

// makeDanglingSymlink creates path as a symlink pointing at a target whose
// parent directory does not exist, so writes through it fail with ENOENT.
func makeDanglingSymlink(t *testing.T, path string) {
	t.Helper()
	target := filepath.Join(filepath.Dir(path), "missing-dir", "target.json")
	if err := os.Symlink(target, path); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
}

func TestSend_GetOrCreateTelemetryError_ReturnsEarly(t *testing.T) {
	t.Setenv("CI", "")
	tmpDir := t.TempDir()
	setConfigDir(t, tmpDir)
	mock := &mockAnalytics{}
	setAnalytics(t, mock)

	// telemetry.json is a directory, so reading it fails.
	dir := filepath.Join(tmpDir, toolName)
	if err := os.MkdirAll(filepath.Join(dir, telemtryFileName), 0755); err != nil {
		t.Fatal(err)
	}

	Send("token", false, config.OnlyAll, config.OutputTable, 1)

	if mock.event != "" {
		t.Error("analytics.Send should not be called when telemetry read fails")
	}
}

func TestSend_WarnedStatePersistFails(t *testing.T) {
	t.Setenv("CI", "")
	tmpDir := t.TempDir()
	setConfigDir(t, tmpDir)
	mock := &mockAnalytics{}
	setAnalytics(t, mock)
	buf := captureStdout(t)

	// Pre-create the tool dir with a dangling telemetry symlink so the new-file
	// write and the warned-state write both fail (ENOENT) without erroring Send.
	dir := filepath.Join(tmpDir, toolName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	makeDanglingSymlink(t, filepath.Join(dir, telemtryFileName))

	Send("token", false, config.OnlyAll, config.OutputTable, 1)

	if !strings.Contains(buf.String(), "anonymous usage telemetry") {
		t.Errorf("expected warning to still be printed, got %q", buf.String())
	}
	if mock.event != "usage" {
		t.Errorf("analytics.Send should still be called, got event %q", mock.event)
	}
}

func TestGetOrCreateTelemetry_FileExistsError(t *testing.T) {
	// A path with an embedded NUL byte is rejected by the OS stat call with a
	// non-ErrNotExist error on both Unix (EINVAL) and Windows (EINVAL).
	_, err := getOrCreateTelemetry("invalid\x00path")
	if err == nil {
		t.Fatal("expected error when stat fails, got nil")
	}
}

func TestGetOrCreateTelemetry_WriteNewFileFails_ReturnsTelemetry(t *testing.T) {
	link := filepath.Join(t.TempDir(), "telemetry.json")
	makeDanglingSymlink(t, link)

	// ensure it is treated as non-existent, then written (which fails).
	if exists, _ := fileExists(link); exists {
		t.Fatal("dangling symlink should not count as existing")
	}

	got, err := getOrCreateTelemetry(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UUID == "" {
		t.Error("expected a UUID even when the write fails")
	}
}

func TestGetTelemetryFilePath_MkdirAllError(t *testing.T) {
	// config dir is a regular file, so MkdirAll(configDir/toolName) fails.
	file := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	setConfigDir(t, file)

	if _, err := getTelemetryFilePath(); err == nil {
		t.Fatal("expected error when MkdirAll fails, got nil")
	}
}

func TestFileExists_StatError(t *testing.T) {
	exists, err := fileExists("invalid\x00path")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
	if exists {
		t.Error("expected exists=false on error")
	}
}

func TestExpandPath_HomeDirError(t *testing.T) {
	clearHomeEnv(t)
	if _, err := expandPath("~/file"); err == nil {
		t.Fatal("expected error when home dir is unavailable, got nil")
	}
}

func TestWriteTelemetry_ExpandPathError(t *testing.T) {
	clearHomeEnv(t)
	if err := writeTelemetry("~/telemetry.json", Telemetry{UUID: "x"}); err == nil {
		t.Fatal("expected error when home dir is unavailable, got nil")
	}
}
