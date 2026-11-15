package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestValidationResultIsValid_ReportsWhetherErrorsExist(t *testing.T) {
	tests := []struct {
		name   string
		result ValidationResult
		want   bool
	}{
		{name: "has errors", result: ValidationResult{Errors: []Issue{{Field: "root"}}}, want: true},
		{name: "has no errors", result: ValidationResult{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsValid(); got != tt.want {
				t.Fatalf("expected IsValid() = %v, got %v", tt.want, got)
			}
		})
	}
}

func TestValidate_WhenConfigIsInvalid_ReturnsErrorsAndWarnings(t *testing.T) {
	rootDir := t.TempDir()
	// NUL forces the filesystem error branches consistently across platforms.
	invalidPath := "\x00"

	cfg := Defaults()
	cfg.Roots = []string{rootDir, invalidPath}
	cfg.Only = OnlyFilter("bad")
	cfg.Output = Output{
		Type:            OutputFormat("yaml"),
		JSONPath:        invalidPath,
		ColorSchemeName: "definitely-not-a-scheme",
	}

	result := Validate(cfg)

	wantErrors := []string{"root", "Only", "Output"}
	if got := issueFields(result.Errors); !reflect.DeepEqual(got, wantErrors) {
		t.Fatalf("expected errors %v, got %v", wantErrors, got)
	}
	wantWarnings := []string{"jsonOutputPath", "output.colorscheme"}
	if got := issueFields(result.Warnings); !reflect.DeepEqual(got, wantWarnings) {
		t.Fatalf("expected warnings %v, got %v", wantWarnings, got)
	}
}

func TestValidate_WhenConfigIsValid_ReturnsNoIssues(t *testing.T) {
	rootDir := t.TempDir()
	jsonPathDir := filepath.Join(rootDir, "reports")
	if err := os.MkdirAll(jsonPathDir, 0o755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}

	cfg := Defaults()
	cfg.Roots = []string{rootDir}
	cfg.Output.JSONPath = jsonPathDir
	cfg.Output.ColorSchemeName = ""

	result := Validate(cfg)

	if len(result.Errors) != 0 || len(result.Warnings) != 0 {
		t.Fatalf("expected no issues, got %+v", result)
	}
}

func TestValidate_ExpandsTildeInRootsAndJSONPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // os.UserHomeDir on Windows
	for _, dir := range []string{"code", "reports"} {
		if err := os.MkdirAll(filepath.Join(home, dir), 0o755); err != nil {
			t.Fatalf("create %s: %v", dir, err)
		}
	}

	t.Run("existing dirs are valid", func(t *testing.T) {
		cfg := Defaults()
		cfg.Roots = []string{"~/code"}
		cfg.Output.JSONPath = "~/reports"

		result := Validate(cfg)
		if len(result.Errors) != 0 || len(result.Warnings) != 0 {
			t.Fatalf("expected no issues, got %+v", result)
		}
	})

	t.Run("missing dirs report the expanded path", func(t *testing.T) {
		cfg := Defaults()
		cfg.Roots = []string{"~/missing"}
		cfg.Output.JSONPath = "~/missing-reports"

		result := Validate(cfg)
		want := filepath.Join(home, "missing")
		if len(result.Errors) != 1 || !strings.Contains(result.Errors[0].Message, want) {
			t.Fatalf("expected one root error naming %q, got %+v", want, result.Errors)
		}
		if len(result.Warnings) != 1 || !strings.Contains(result.Warnings[0].Message, want+"-reports") {
			t.Fatalf("expected one jsonOutputPath warning naming %q, got %+v", want+"-reports", result.Warnings)
		}
	})
}

func TestValidationResultLog_DoesNotPanic(t *testing.T) {
	result := ValidationResult{
		Warnings: []Issue{{Field: "output", Message: "warn"}},
		Errors:   []Issue{{Field: "root", Message: "err"}},
	}

	result.Log()
}

func TestValidationResultPrint_WritesIssues(t *testing.T) {
	result := ValidationResult{
		Warnings: []Issue{{Field: "output", Message: "warn"}},
		Errors:   []Issue{{Field: "root", Message: "err"}},
	}

	output := captureStdout(t, func() {
		result.Print()
	})

	for _, want := range []string{"Warning:", "Error:", "field=output", "field=root"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got %q", want, output)
		}
	}
}

func issueFields(issues []Issue) []string {
	fields := make([]string, len(issues))
	for i, issue := range issues {
		fields[i] = issue.Field
	}
	return fields
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = original
		reader.Close()
	})

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	var output bytes.Buffer
	if _, err := io.Copy(&output, reader); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return output.String()
}
