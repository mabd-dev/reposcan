package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

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
