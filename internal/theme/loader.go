package theme

import (
	"embed"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/logger"
	"gopkg.in/yaml.v3"
)

var (
	SchemesDir        string = "base24-schemas/"
	defaultSchemeName string = "catppuccin-mocha"
)

//go:embed base24-schemas/*.yaml
var schemesFS embed.FS

func LoadBase24Schema(path string) (Base24Scheme, error) {
	if !strings.HasSuffix(path, ".yaml") {
		path += ".yaml"
	}

	data, err := schemesFS.ReadFile(path)
	if err != nil {
		return Base24Scheme{}, err
	}

	var b Base24Scheme
	if err := yaml.Unmarshal(data, &b); err != nil {
		return Base24Scheme{}, err
	}
	return b, nil
}

func LoadBase24(path string) (ColorScheme, error) {
	b, err := LoadBase24Schema(path)
	if err != nil {
		return ColorScheme{}, err
	}

	c := ColorScheme{
		Name:            b.Name,
		Background:      lipgloss.Color(b.Palette.Base00),
		Foreground:      lipgloss.Color(b.Palette.Base05),
		Accent:          lipgloss.Color(b.Palette.Base0D),
		Muted:           lipgloss.Color(b.Palette.Base03),
		Error:           lipgloss.Color(b.Palette.Base08),
		Warning:         lipgloss.Color(b.Palette.Base09),
		Success:         lipgloss.Color(b.Palette.Base0B),
		Info:            lipgloss.Color(b.Palette.Base0C),
		Border:          lipgloss.Color(b.Palette.Base02),
		BorderActive:    lipgloss.Color(b.Palette.Base0D),
		TableHeader:     lipgloss.Color(b.Palette.Base04),
		TableRow:        lipgloss.Color(b.Palette.Base05),
		TableAltRow:     lipgloss.Color(b.Palette.Base01),
		PopupBackground: lipgloss.Color(b.Palette.Base00),
		PopupBorder:     lipgloss.Color(b.Palette.Base0D),
		PopupTitle:      lipgloss.Color(b.Palette.Base0A),
	}

	return c, nil
}

func CreateColors(colorSchemeName string) (ColorScheme, error) {
	colorScheme, err := LoadBase24(SchemesDir + colorSchemeName)
	if err != nil {
		logger.Debug("using default color scheme", logger.StringAttr("name", defaultSchemeName))
		colorScheme, err := LoadBase24(SchemesDir + defaultSchemeName)
		if err != nil {
			return ColorScheme{}, err
		}
		return colorScheme, nil
	}

	logger.Debug("used color scheme", logger.StringAttr("name", colorSchemeName))
	return colorScheme, nil
}

func CreateStyles(colors ColorScheme) Styles {
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
