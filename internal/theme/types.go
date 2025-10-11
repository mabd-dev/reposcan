// Package theme define app theme and handles reading colorschemes.
// ColorSchemes relies on https://github.com/tinted-theming/schemes
package theme

import (
	"github.com/charmbracelet/lipgloss"
)

type Base24ColorSchema struct {
	Scheme  string        `yaml:"scheme"`
	Slug    string        `yaml:"slug"`
	Author  string        `yaml:"author"`
	Palette Base24Palette `yaml:"palette"`
}

type Base24Palette struct {
	Base00 string `yaml:"base00"`
	Base01 string `yaml:"base01"`
	Base02 string `yaml:"base02"`
	Base03 string `yaml:"base03"`
	Base04 string `yaml:"base04"`
	Base05 string `yaml:"base05"`
	Base06 string `yaml:"base06"`
	Base07 string `yaml:"base07"`
	Base08 string `yaml:"base08"`
	Base09 string `yaml:"base09"`
	Base0A string `yaml:"base0A"`
	Base0B string `yaml:"base0B"`
	Base0C string `yaml:"base0C"`
	Base0D string `yaml:"base0D"`
	Base0E string `yaml:"base0E"`
	Base0F string `yaml:"base0F"`
	Base10 string `yaml:"base10"`
	Base11 string `yaml:"base11"`
	Base12 string `yaml:"base12"`
	Base13 string `yaml:"base13"`
	Base14 string `yaml:"base14"`
	Base15 string `yaml:"base15"`
	Base16 string `yaml:"base16"`
	Base17 string `yaml:"base17"`
}

type ColorScheme struct {
	Background string
	Foreground string
	Accent     string
	Muted      string
	Error      string
	Warning    string
	Success    string
	Info       string

	Border       string
	BorderActive string

	TableHeader string
	TableRow    string
	TableAltRow string

	PopupBackground string
	PopupBorder     string
	PopupTitle      string
}

type LipglossScheme struct {
	Background lipgloss.Color
	Foreground lipgloss.Color
	Accent     lipgloss.Color
	Muted      lipgloss.Color
	Error      lipgloss.Color
	Warning    lipgloss.Color
	Success    lipgloss.Color
	Info       lipgloss.Color

	Border       lipgloss.Color
	BorderActive lipgloss.Color

	TableHeader lipgloss.Color
	TableRow    lipgloss.Color
	TableAltRow lipgloss.Color

	PopupBackground lipgloss.Color
	PopupBorder     lipgloss.Color
	PopupTitle      lipgloss.Color
}

type Styles struct {
	Base  lipgloss.Style
	Muted lipgloss.Style

	Box                   lipgloss.Style
	BoxMuted              lipgloss.Style
	TableHeader           lipgloss.Style
	TableSelectedRow      lipgloss.Style
	TableSelectedRowMuted lipgloss.Style
	TableRow              lipgloss.Style

	Popup       lipgloss.Style
	PopupHeader lipgloss.Style
	PopupText   lipgloss.Style

	Notification     lipgloss.Style
	NotificationText lipgloss.Style
}

// BoxFor return [Box] or [BoxMuted] based on isActive
func (s Styles) BoxFor(isActive bool) lipgloss.Style {
	if isActive {
		return s.Box
	}
	return s.BoxMuted
}

type Theme struct {
	Colors LipglossScheme
	Styles Styles
}
