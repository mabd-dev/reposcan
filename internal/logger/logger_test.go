package logger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

// isolateLoggerState points the home directory at a fresh temp dir and
// restores the package globals when the test ends, closing any log file the
// test left open so the temp dir can be removed.
func isolateLoggerState(t *testing.T) string {
	t.Helper()
	previousEnabled := enabled
	previousLogger := logger
	previousLogFile := logFile

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	// Registered after TempDir so it runs before the dir is removed.
	t.Cleanup(func() {
		if logFile != nil && logFile != previousLogFile {
			if err := logFile.Close(); err != nil {
				t.Errorf("Close() error = %v", err)
			}
		}
		enabled = previousEnabled
		logger = previousLogger
		logFile = previousLogFile
	})
	return homeDir
}

// readLogDir returns the contents of every log file in dir, in name order.
// Reading all files keeps tests stable if a run crosses midnight and the
// dated file name changes between calls.
func readLogDir(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%q) error = %v", dir, err)
	}
	var contents strings.Builder
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", entry.Name(), err)
		}
		contents.Write(data)
	}
	return contents.String()
}

func capturePanic(f func()) (recovered any) {
	defer func() { recovered = recover() }()
	f()
	return nil
}

func TestCreateLogFile(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	logDir := fmt.Sprintf("logs-%d", time.Now().UnixNano())

	file, err := createLogFile(logDir)
	if err != nil {
		t.Fatalf("createLogFile() error = %v", err)
	}
	defer file.Close()

	if got, want := filepath.Dir(file.Name()), filepath.Join(homeDir, logDir); got != want {
		t.Errorf("log directory = %q, want %q", got, want)
	}
	if got, want := filepath.Ext(file.Name()), ".log"; got != want {
		t.Errorf("log extension = %q, want %q", got, want)
	}
}

func TestInitEnabledWritesInitializationLog(t *testing.T) {
	previousEnabled := enabled
	previousLogger := logger
	previousLogFile := logFile

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	t.Cleanup(func() {
		if logFile != nil && logFile != previousLogFile {
			if err := logFile.Close(); err != nil {
				t.Errorf("Close() error = %v", err)
			}
		}
		enabled = previousEnabled
		logger = previousLogger
		logFile = previousLogFile
	})
	Init(true, "logs")

	if !enabled {
		t.Fatal("Init(true) did not enable logging")
	}
	if logger == nil {
		t.Fatal("Init(true) did not initialize a logger")
	}

	logDir := filepath.Join(homeDir, "logs")
	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatalf("ReadDir(%q) error = %v", logDir, err)
	}
	if len(entries) != 1 {
		t.Fatalf("log files = %d, want 1", len(entries))
	}
	logFile := filepath.Join(logDir, entries[0].Name())
	contents, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", logFile, err)
	}
	if !strings.Contains(string(contents), "Logger initialized") {
		t.Errorf("log file = %q, want initialization message", contents)
	}
}

func TestLogCallsRespectEnabledState(t *testing.T) {
	previousEnabled := enabled
	previousLogger := logger
	t.Cleanup(func() {
		enabled = previousEnabled
		logger = previousLogger
	})

	enabled = false
	logger = nil
	Debug("debug")
	Info("info")
	Warn("warning")
	Error("error")

	enabled = true
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	Debug("debug")
	Info("info")
	Warn("warning")
	Error("error")
}

func TestInitDisabled(t *testing.T) {
	previousEnabled := enabled
	previousLogger := logger
	t.Cleanup(func() {
		enabled = previousEnabled
		logger = previousLogger
	})

	logger = nil
	Init(false, "logs")
	if enabled {
		t.Fatal("Init(false) enabled logging")
	}
	if logger != nil {
		t.Fatal("Init(false) initialized a logger")
	}
}

func TestAttributeHelpers(t *testing.T) {
	tests := []struct {
		name string
		got  slog.Attr
		want slog.Attr
	}{
		{name: "bool", got: BoolAttr("enabled", true), want: slog.Bool("enabled", true)},
		{name: "string", got: StringAttr("branch", "main"), want: slog.String("branch", "main")},
		{name: "int", got: IntAttr("count", 3), want: slog.Int("count", 3)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Errorf("attribute = %#v, want %#v", tt.got, tt.want)
			}
		})
	}
}

func TestInitClosesPreviousLogFile(t *testing.T) {
	tests := []struct {
		name          string
		secondEnabled bool
		wantInitLogs  int
	}{
		{name: "re-enable appends to the dated log file", secondEnabled: true, wantInitLogs: 2},
		{name: "disable stops writing", secondEnabled: false, wantInitLogs: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			homeDir := isolateLoggerState(t)
			Init(true, "logs")
			previous := logFile
			if previous == nil {
				t.Fatal("Init(true) did not open a log file")
			}

			Init(tt.secondEnabled, "logs")
			Info("after-second-init")

			if err := previous.Close(); !errors.Is(err, os.ErrClosed) {
				t.Errorf("previous log file Close() error = %v, want %v", err, os.ErrClosed)
			}
			if enabled != tt.secondEnabled {
				t.Errorf("enabled = %v, want %v", enabled, tt.secondEnabled)
			}
			if tt.secondEnabled && (logFile == nil || logFile == previous) {
				t.Errorf("Init(true) did not open a new log file")
			}
			if !tt.secondEnabled && logFile != nil {
				t.Errorf("Init(false) left log file %q open", logFile.Name())
			}

			contents := readLogDir(t, filepath.Join(homeDir, "logs"))
			if got := strings.Count(contents, "Logger initialized"); got != tt.wantInitLogs {
				t.Errorf("initialization logs = %d, want %d in %q", got, tt.wantInitLogs, contents)
			}
			if got := strings.Contains(contents, "after-second-init"); got != tt.secondEnabled {
				t.Errorf("log contains message after second Init = %v, want %v", got, tt.secondEnabled)
			}
		})
	}
}

func TestLogHelpersWriteLevelMessageAndAttributes(t *testing.T) {
	tests := []struct {
		name  string
		log   func(string, ...any)
		level string
	}{
		{name: "debug", log: Debug, level: "DEBUG"},
		{name: "info", log: Info, level: "INFO"},
		{name: "warn", log: Warn, level: "WARN"},
		{name: "error", log: Error, level: "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			homeDir := isolateLoggerState(t)
			Init(true, "logs")

			tt.log(tt.name+"-message",
				BoolAttr("dirty", true),
				StringAttr("branch", "main"),
				IntAttr("ahead", 2),
			)

			contents := readLogDir(t, filepath.Join(homeDir, "logs"))
			lines := strings.Split(strings.TrimSpace(contents), "\n")
			last := lines[len(lines)-1]
			for _, want := range []string{
				"level=" + tt.level,
				"msg=" + tt.name + "-message",
				"dirty=true",
				"branch=main",
				"ahead=2",
			} {
				if !strings.Contains(last, want) {
					t.Errorf("log line = %q, want it to contain %q", last, want)
				}
			}
		})
	}
}

func TestInitPanicsWhenLogFileCannotBeCreated(t *testing.T) {
	tests := []struct {
		name string
		// setup prepares homeDir and returns the paths the failing
		// *os.PathError may name, or nil when no PathError is expected.
		setup      func(t *testing.T, homeDir string) []string
		wantPrefix string
		wantOp     string
	}{
		{
			name: "home dir unset",
			setup: func(t *testing.T, homeDir string) []string {
				if runtime.GOOS == "android" || runtime.GOOS == "ios" {
					t.Skipf("os.UserHomeDir does not fail on %s", runtime.GOOS)
				}
				t.Setenv("HOME", "")
				t.Setenv("USERPROFILE", "")
				return nil
			},
			wantPrefix: "failed to get user home dir: ",
		},
		{
			name: "log dir path is a file",
			setup: func(t *testing.T, homeDir string) []string {
				logDir := filepath.Join(homeDir, "logs")
				if err := os.WriteFile(logDir, nil, 0o644); err != nil {
					t.Fatalf("WriteFile(%q) error = %v", logDir, err)
				}
				return []string{logDir}
			},
			wantPrefix: "failed to create log dir: ",
			wantOp:     "mkdir",
		},
		{
			name: "dated log file path is a directory",
			setup: func(t *testing.T, homeDir string) []string {
				// Block tomorrow's file too in case the run crosses midnight.
				now := time.Now()
				var paths []string
				for _, day := range []time.Time{now, now.AddDate(0, 0, 1)} {
					path := filepath.Join(homeDir, "logs", day.Format("2006-01-02")+".log")
					if err := os.MkdirAll(path, 0o755); err != nil {
						t.Fatalf("MkdirAll(%q) error = %v", path, err)
					}
					paths = append(paths, path)
				}
				return paths
			},
			wantPrefix: "failed to open log file: ",
			wantOp:     "open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			homeDir := isolateLoggerState(t)
			wantPaths := tt.setup(t, homeDir)

			recovered := capturePanic(func() { Init(true, "logs") })

			err, ok := recovered.(error)
			if !ok {
				t.Fatalf("Init(true) panic = %#v, want an error", recovered)
			}
			if !strings.HasPrefix(err.Error(), tt.wantPrefix) {
				t.Errorf("panic error = %q, want prefix %q", err, tt.wantPrefix)
			}
			if wantPaths != nil {
				var pathErr *os.PathError
				if !errors.As(err, &pathErr) {
					t.Fatalf("panic error = %#v, want it to wrap *os.PathError", err)
				}
				if pathErr.Op != tt.wantOp {
					t.Errorf("PathError.Op = %q, want %q", pathErr.Op, tt.wantOp)
				}
				if !slices.Contains(wantPaths, pathErr.Path) {
					t.Errorf("PathError.Path = %q, want one of %q", pathErr.Path, wantPaths)
				}
			}
			if logFile != nil {
				t.Errorf("failed Init left log file %q open", logFile.Name())
			}
		})
	}
}
