package tui

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/termenv"
)

// Most of this code is borrowed from
// https://github.com/charmbracelet/lipgloss/pull/102
// as well as the lipgloss library, with some modification for what I needed.

// PlaceOverlayWithPosition places fg on top of bg with overlay position
func PlaceOverlayWithPosition(
	position OverlayPosition,
	fullWidth, fullHeight int,
	fg, bg string,
	shadow bool, opts ...WhitespaceOption,
) string {
	fgLines, fgWidth := GetLines(fg)
	fgHeight := len(fgLines)

	var x, y int

	switch position {
	case OverlayPositionCenter:
		x = (fullWidth - fgWidth) / 2
		y = (fullHeight - fgHeight) / 2
	case OverlayPositionTopRight:
		x = fullWidth - fgWidth
		y = 0
	case OverlayPositionTopLeft:
		x = 0
		y = 0
	case OverlayPositionBottomRight:
		x = fullWidth - fgWidth
		y = (fullHeight)
	default:
		panic("Unknown overlay position")
	}

	return PlaceOverlay(x, y, fg, bg, shadow, opts...)
}

// PlaceOverlay places fg on top of bg.
func PlaceOverlay(
	x, y int,
	fg, bg string,
	shadow bool, opts ...WhitespaceOption,
) string {
	fgLines, fgWidth := GetLines(fg)
	bgLines, bgWidth := GetLines(bg)
	bgHeight := len(bgLines)
	fgHeight := len(fgLines)

	if shadow {
		var shadowbg string = ""
		shadowchar := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333")).
			Render("░")
		for i := 0; i <= fgHeight; i++ {
			if i == 0 {
				shadowbg += " " + strings.Repeat(" ", fgWidth) + "\n"
			} else {
				shadowbg += " " + strings.Repeat(shadowchar, fgWidth) + "\n"
			}
		}

		fg = PlaceOverlay(0, 0, fg, shadowbg, false, opts...)
		fgLines, fgWidth = GetLines(fg)
		fgHeight = len(fgLines)
	}

	if fgWidth >= bgWidth && fgHeight >= bgHeight {
		// FIXME: return fg or bg?
		return fg
	}
	// TODO: allow placement outside of the bg box?
	x = clamp(x, 0, bgWidth-fgWidth)
	y = clamp(y, 0, bgHeight-fgHeight)

	ws := &whitespace{}
	for _, opt := range opts {
		opt(ws)
	}

	var b strings.Builder
	for i, bgLine := range bgLines {
		if i > 0 {
			b.WriteByte('\n')
		}
		if i < y || i >= y+fgHeight {
			b.WriteString(bgLine)
			continue
		}

		pos := 0
		if x > 0 {
			left := truncate.String(bgLine, uint(x))
			pos = ansi.PrintableRuneWidth(left)
			b.WriteString(left)
			if pos < x {
				b.WriteString(ws.render(x - pos))
				pos = x
			}
		}

		fgLine := fgLines[i-y]
		b.WriteString(fgLine)
		pos += ansi.PrintableRuneWidth(fgLine)

		right := cutLeft(bgLine, pos)
		bgWidth := ansi.PrintableRuneWidth(bgLine)
		rightWidth := ansi.PrintableRuneWidth(right)
		if rightWidth <= bgWidth-pos {
			b.WriteString(ws.render(bgWidth - rightWidth - pos))
		}

		b.WriteString(right)
	}

	return b.String()
}

// GetLines Split a string into lines, additionally returning the size of the widest
// line.
func GetLines(s string) (lines []string, widest int) {
	lines = strings.Split(s, "\n")

	for _, l := range lines {
		w := ansi.PrintableRuneWidth(l)
		if widest < w {
			widest = w
		}
	}

	return lines, widest
}

// cutLeft cuts printable characters from the left.
// This function is heavily based on muesli's ansi and truncate packages.
func cutLeft(s string, cutWidth int) string {
	var (
		pos    int
		isAnsi bool
		ab     bytes.Buffer
		b      bytes.Buffer
	)
	for _, c := range s {
		var w int
		if c == ansi.Marker || isAnsi {
			isAnsi = true
			ab.WriteRune(c)
			if ansi.IsTerminator(c) {
				isAnsi = false
				if bytes.HasSuffix(ab.Bytes(), []byte("[0m")) {
					ab.Reset()
				}
			}
		} else {
			w = runewidth.RuneWidth(c)
		}

		if pos >= cutWidth {
			if b.Len() == 0 {
				if ab.Len() > 0 {
					b.Write(ab.Bytes())
				}
				if pos-cutWidth > 1 {
					b.WriteByte(' ')
					continue
				}
			}
			b.WriteRune(c)
		}
		pos += w
	}
	return b.String()
}

func clamp(v, lower, upper int) int {
	return min(max(v, lower), upper)
}

type whitespace struct {
	style termenv.Style
	chars string
}

// Render whitespaces.
func (w whitespace) render(width int) string {
	if w.chars == "" {
		w.chars = " "
	}

	r := []rune(w.chars)
	j := 0
	b := strings.Builder{}

	// Cycle through runes and print them into the whitespace.
	for i := 0; i < width; {
		b.WriteRune(r[j])
		j++
		if j >= len(r) {
			j = 0
		}
		i += ansi.PrintableRuneWidth(string(r[j]))
	}

	// Fill any extra gaps white spaces. This might be necessary if any runes
	// are more than one cell wide, which could leave a one-rune gap.
	short := width - ansi.PrintableRuneWidth(b.String())
	if short > 0 {
		b.WriteString(strings.Repeat(" ", short))
	}

	return w.style.Styled(b.String())
}

// WhitespaceOption sets a styling rule for rendering whitespace.
type WhitespaceOption func(*whitespace)

// WithWhitespaceChars sets the characters to be rendered in the whitespace.
func WithWhitespaceChars(s string) WhitespaceOption {
	return func(w *whitespace) {
		w.chars = s
	}
}

type OverlayPosition string

const (
	OverlayPositionCenter      OverlayPosition = "center"
	OverlayPositionTopRight    OverlayPosition = "topRight"
	OverlayPositionTopLeft     OverlayPosition = "topLeft"
	OverlayPositionBottomRight OverlayPosition = "bottomright"
)
