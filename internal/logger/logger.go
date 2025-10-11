// Package logger is a logger package :)
package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var (
	enabled bool = false
	logger  *slog.Logger
)

func Init(
	isEnabled bool,
	logFilePath string,
) {
	enabled = isEnabled
	if !enabled {
		return
	}

	file, err := createLogFile(logFilePath)
	if err != nil {
		panic(err)
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger = slog.New(slog.NewTextHandler(file, opts))
	logger.Info("Logger initialized")
}

func createLogFile(logFilePath string) (*os.File, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("failed to get user home dir: %w", err))
	}
	logDir := filepath.Join(homeDir, logFilePath)

	// Ensure directory exists
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		panic(fmt.Errorf("failed to create log dir: %w", err))
	}

	// Create log file with today's date
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, today+".log")

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return file, nil
}

func Debug(msg string, args ...any) {
	if enabled {
		logger.Debug(msg, args...)
	}
}

func Info(msg string, args ...any) {
	if enabled {
		logger.Info(msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if enabled {
		logger.Warn(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if enabled {
		logger.Error(msg, args...)
	}
}

func BoolAttr(key string, v bool) slog.Attr {
	return slog.Bool(key, v)
}

func StringAttr(key string, v string) slog.Attr {
	return slog.String(key, v)
}

func IntAttr(key string, v int) slog.Attr {
	return slog.Int(key, v)
}
