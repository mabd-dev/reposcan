package config

import "testing"

func TestOutputFormatIsValid(t *testing.T) {
	tests := []struct {
		name   string
		format OutputFormat
		want   bool
	}{
		{name: "json", format: OutputJson, want: true},
		{name: "table", format: OutputTable, want: true},
		{name: "interactive", format: OutputInteractive, want: true},
		{name: "none", format: OutputNone, want: true},
		{name: "unknown", format: OutputFormat("yaml"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.IsValid(); got != tt.want {
				t.Fatalf("expected IsValid() = %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCreateOutputFormat(t *testing.T) {
	tests := []struct {
		input string
		want  OutputFormat
	}{
		// "table" remains a compatibility alias for interactive output.
		{input: " table ", want: OutputInteractive},
		{input: "interactive", want: OutputInteractive},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := CreateOutputFormat(tt.input)
			if err != nil {
				t.Fatalf("create output format: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
