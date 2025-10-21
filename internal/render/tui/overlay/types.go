package overlay

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
