package overlay

import "testing"

func TestPlaceOverlayWithPositionAndPaddingHonorsBottomPadding(t *testing.T) {
	got := PlaceOverlayWithPositionAndPadding(
		OverlayPositionBottomRight,
		5,
		3,
		1,
		1,
		"X",
		".....\n.....\n.....",
		false,
	)
	want := ".....\n...X.\n....."

	if got != want {
		t.Fatalf("PlaceOverlayWithPositionAndPadding() = %q, want %q", got, want)
	}
}

func TestCutLeftDoesNotSplitWideRune(t *testing.T) {
	if got, want := cutLeft("界ab", 1), " ab"; got != want {
		t.Fatalf("cutLeft() = %q, want %q", got, want)
	}
}

func TestCutLeftPreservesANSISequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		cutWidth int
		want     string
	}{
		{
			name:     "sequence starts at cut",
			input:    "ab\x1b[31mcd\x1b[0m",
			cutWidth: 2,
			want:     "\x1b[31mcd\x1b[0m",
		},
		{
			name:     "cut splits wide rune before sequence",
			input:    "界\x1b[31mab\x1b[0m",
			cutWidth: 1,
			want:     " \x1b[31mab\x1b[0m",
		},
		{
			name:     "style active at cut",
			input:    "\x1b[31mabcdef\x1b[0m",
			cutWidth: 2,
			want:     "\x1b[31mcdef\x1b[0m",
		},
		{
			name:     "reset before cut",
			input:    "\x1b[31mab\x1b[0mcd",
			cutWidth: 3,
			want:     "d",
		},
		{
			name:     "zero cut",
			input:    "\x1b[31mab\x1b[0m",
			cutWidth: 0,
			want:     "\x1b[31mab\x1b[0m",
		},
		{
			name:     "cut beyond end",
			input:    "\x1b[31mab\x1b[0m",
			cutWidth: 3,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cutLeft(tt.input, tt.cutWidth); got != tt.want {
				t.Fatalf("cutLeft(%q, %d) = %q, want %q", tt.input, tt.cutWidth, got, tt.want)
			}
		})
	}
}

func TestWhitespaceRenderHandlesRuneWidths(t *testing.T) {
	tests := []struct {
		name  string
		chars string
		width int
		want  string
	}{
		{name: "combining pattern width one", chars: "e\u0301", width: 1, want: "e\u0301"},
		{name: "combining pattern width two", chars: "e\u0301", width: 2, want: "e\u0301e\u0301"},
		{name: "combining mark inside pattern", chars: "a\u0301b", width: 4, want: "a\u0301ba\u0301b"},
		{name: "combining mark only", chars: "\u0301", width: 2, want: "  "},
		{name: "zero width space only", chars: "\u200b", width: 2, want: "  "},
		{name: "zero width", chars: "e\u0301", width: 0, want: ""},
		{name: "negative width", chars: "e\u0301", width: -1, want: ""},
		{name: "wide pattern narrower target", chars: "界", width: 1, want: " "},
		{name: "wide pattern with remainder", chars: "界", width: 3, want: "界 "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (whitespace{chars: tt.chars}).render(tt.width); got != tt.want {
				t.Fatalf("render(%d) = %q, want %q", tt.width, got, tt.want)
			}
		})
	}
}

func TestPlaceOverlayUsesWhitespacePatternCellWidths(t *testing.T) {
	tests := []struct {
		name  string
		chars string
		want  string
	}{
		{name: "combining pattern", chars: "e\u0301", want: "ae\u0301X\n...."},
		{name: "wide pattern", chars: "界", want: "a X\n...."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlaceOverlay(2, 0, "X", "a\n....", false, WithWhitespaceChars(tt.chars))
			if got != tt.want {
				t.Fatalf("PlaceOverlay() = %q, want %q", got, tt.want)
			}
		})
	}
}
