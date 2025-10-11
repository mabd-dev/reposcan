package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/logger"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

var (
	schemesDir        string = "internal/theme/schemes/base24/"
	defaultSchemeName string = "catppuccin-mocha"
)

func LoadBase24(path string) (ColorScheme, error) {
	if !strings.HasSuffix(path, ".yaml") {
		path += ".yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ColorScheme{}, err
	}

	var b Base24ColorSchema
	if err := yaml.Unmarshal(data, &b); err != nil {
		return ColorScheme{}, err
	}

	c := ColorScheme{
		Background:      b.Palette.Base00,
		Foreground:      b.Palette.Base05,
		Accent:          b.Palette.Base0D,
		Muted:           b.Palette.Base03,
		Error:           b.Palette.Base08,
		Warning:         b.Palette.Base09,
		Success:         b.Palette.Base0B,
		Info:            b.Palette.Base0C,
		Border:          b.Palette.Base02,
		BorderActive:    b.Palette.Base0D,
		TableHeader:     b.Palette.Base04,
		TableRow:        b.Palette.Base05,
		TableAltRow:     b.Palette.Base01,
		PopupBackground: b.Palette.Base00,
		PopupBorder:     b.Palette.Base0D,
		PopupTitle:      b.Palette.Base0A,
	}

	return c, nil
}

func CreateColors(colorSchemeName string) (LipglossScheme, error) {
	colorScheme, err := LoadBase24(schemesDir + colorSchemeName)
	if err != nil {
		logger.Debug("using default color scheme", logger.StringAttr("name", defaultSchemeName))
		colorScheme, err := LoadBase24(schemesDir + defaultSchemeName)
		if err != nil {
			return LipglossScheme{}, err
		}
		return toLipglossTheme(colorScheme), nil
	}

	logger.Debug("used color scheme", logger.StringAttr("name", colorSchemeName))
	return toLipglossTheme(colorScheme), nil
}

func toLipglossTheme(cs ColorScheme) LipglossScheme {
	return LipglossScheme{
		Background:      lipgloss.Color(cs.Background),
		Foreground:      lipgloss.Color(cs.Foreground),
		Accent:          lipgloss.Color(cs.Accent),
		Muted:           lipgloss.Color(cs.Muted),
		Error:           lipgloss.Color(cs.Error),
		Warning:         lipgloss.Color(cs.Warning),
		Success:         lipgloss.Color(cs.Success),
		Info:            lipgloss.Color(cs.Info),
		Border:          lipgloss.Color(cs.Border),
		BorderActive:    lipgloss.Color(cs.BorderActive),
		TableHeader:     lipgloss.Color(cs.TableHeader),
		TableRow:        lipgloss.Color(cs.TableRow),
		TableAltRow:     lipgloss.Color(cs.TableAltRow),
		PopupBackground: lipgloss.Color(cs.PopupBackground),
		PopupBorder:     lipgloss.Color(cs.PopupBorder),
		PopupTitle:      lipgloss.Color(cs.PopupTitle),
	}
}

func CreateStyles(colors LipglossScheme) Styles {
	return Styles{
		Base:  lipgloss.NewStyle(),
		Muted: lipgloss.NewStyle().Foreground(colors.Muted), //.Faint(true) // TODO: do i need Faint as well?

		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.BorderActive),
		// Background(colors.Background)
		BoxMuted: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.Border),
		// Background(colors.Background),
		TableHeader: lipgloss.NewStyle().
			Foreground(colors.TableHeader).
			// Background(colors.Background).
			Bold(true),
		TableSelectedRow: lipgloss.NewStyle().
			Background(colors.TableAltRow).
			Foreground(colors.Accent).
			Bold(true),
		TableSelectedRowMuted: lipgloss.NewStyle().
			Background(colors.TableAltRow).
			Foreground(colors.Muted),
		TableRow: lipgloss.NewStyle(),

		Popup: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colors.BorderActive).
			Padding(1, 2).
			Align(lipgloss.Center),
		PopupHeader: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 2, 0, 2).
			Italic(true).
			MarginBottom(1),
		PopupText: lipgloss.NewStyle().Foreground(colors.Foreground),
	}
}
