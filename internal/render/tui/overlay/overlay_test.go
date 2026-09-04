package overlay

import (
	"io"
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

func TestPlaceOverlayWithPosition(t *testing.T) {
	tests := []struct {
		name     string
		position OverlayPosition
		want     string
	}{
		{name: "center", position: OverlayPositionCenter, want: ".....\n..X..\n....."},
		{name: "top right", position: OverlayPositionTopRight, want: "....X\n.....\n....."},
		{name: "top left", position: OverlayPositionTopLeft, want: "X....\n.....\n....."},
		{name: "bottom right", position: OverlayPositionBottomRight, want: ".....\n.....\n....X"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlaceOverlayWithPosition(tt.position, 5, 3, "X", ".....\n.....\n.....", false)
			if got != tt.want {
				t.Fatalf("PlaceOverlayWithPosition() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlaceOverlayWithPositionAndPadding(t *testing.T) {
	tests := []struct {
		name     string
		position OverlayPosition
		hPadding int
		want     string
	}{
		{name: "top right", position: OverlayPositionTopRight, hPadding: 1, want: "...X.\n.....\n....."},
		{name: "top left", position: OverlayPositionTopLeft, hPadding: 1, want: ".X...\n.....\n....."},
		{name: "bottom right", position: OverlayPositionBottomRight, hPadding: 1, want: ".....\n.....\n...X."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlaceOverlayWithPositionAndPadding(
				tt.position, 5, 3, tt.hPadding, 0, "X", ".....\n.....\n.....", false,
			)
			if got != tt.want {
				t.Fatalf("PlaceOverlayWithPositionAndPadding() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlaceOverlayWithPositionPanicsForUnknownPosition(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("PlaceOverlayWithPosition() did not panic")
		}
	}()

	PlaceOverlayWithPosition(OverlayPosition("unknown"), 1, 1, "X", ".", false)
}

func TestPlaceOverlay(t *testing.T) {
	tests := []struct {
		name string
		x, y int
		fg   string
		bg   string
		opts []WhitespaceOption
		want string
	}{
		{
			name: "places multiline foreground",
			x:    2,
			y:    1,
			fg:   "AB\nCD",
			bg:   "......\n......\n......\n......",
			want: "......\n..AB..\n..CD..\n......",
		},
		{
			name: "clamps negative coordinates",
			x:    -3,
			y:    -2,
			fg:   "X",
			bg:   "...\n...",
			want: "X..\n...",
		},
		{
			name: "clamps coordinates beyond background",
			x:    99,
			y:    99,
			fg:   "X",
			bg:   "...\n...",
			want: "...\n..X",
		},
		{
			name: "fills a short background line",
			x:    2,
			fg:   "X",
			bg:   "a\n....",
			opts: []WhitespaceOption{WithWhitespaceChars("-")},
			want: "a-X\n....",
		},
		{
			name: "foreground matching background size is returned",
			fg:   "ABC\nDEF",
			bg:   "...\n...",
			want: "ABC\nDEF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlaceOverlay(tt.x, tt.y, tt.fg, tt.bg, false, tt.opts...)
			if got != tt.want {
				t.Fatalf("PlaceOverlay() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlaceOverlayWithShadow(t *testing.T) {
	got := PlaceOverlay(0, 0, "X", "....\n....\n....", true)
	lines := strings.Split(got, "\n")

	if len(lines) != 3 {
		t.Fatalf("PlaceOverlay() returned %d lines, want 3: %q", len(lines), got)
	}
	if lines[0] != "X .." {
		t.Errorf("first line = %q, want %q", lines[0], "X ..")
	}
	if !strings.Contains(lines[1], "░") {
		t.Errorf("second line does not contain shadow character: %q", lines[1])
	}
}

func TestGetLines(t *testing.T) {
	lines, widest := GetLines("a\n\x1b[31m界界\x1b[0m\n")

	if len(lines) != 3 {
		t.Fatalf("GetLines() returned %d lines, want 3", len(lines))
	}
	if widest != 4 {
		t.Fatalf("GetLines() widest = %d, want 4", widest)
	}
}

func TestCutLeft(t *testing.T) {
	tests := []struct {
		name string
		s    string
		cut  int
		want string
	}{
		{name: "ascii", s: "abcdef", cut: 2, want: "cdef"},
		{name: "preserves active ANSI style", s: "\x1b[31mabcdef\x1b[0m", cut: 2, want: "\x1b[31mcdef\x1b[0m"},
		{name: "does not preserve reset style", s: "\x1b[31mab\x1b[0mcd", cut: 3, want: "d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cutLeft(tt.s, tt.cut); got != tt.want {
				t.Fatalf("cutLeft(%q, %d) = %q, want %q", tt.s, tt.cut, got, tt.want)
			}
		})
	}
}

func TestWhitespaceRender(t *testing.T) {
	ansiStyle := termenv.NewOutput(io.Discard, termenv.WithProfile(termenv.ANSI)).String().
		Foreground(termenv.ANSIRed)
	tests := []struct {
		name  string
		ws    whitespace
		width int
		want  string
	}{
		{name: "defaults to spaces", ws: whitespace{}, width: 3, want: "   "},
		{name: "cycles characters", ws: whitespace{chars: "ab"}, width: 3, want: "aba"},
		{name: "fills remainder after mixed width characters", ws: whitespace{chars: "a界"}, width: 2, want: "a "},
		{name: "applies style", ws: whitespace{chars: "x", style: ansiStyle}, width: 3, want: "\x1b[31mxxx\x1b[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ws.render(tt.width); got != tt.want {
				t.Fatalf("render(%d) = %q, want %q", tt.width, got, tt.want)
			}
		})
	}
}
