package repodetails

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/theme"
)

func generateFakeLipGlossScheme() theme.LipglossScheme {
	return theme.LipglossScheme{
		Background: lipgloss.Color("#000001"),
		Foreground: lipgloss.Color("#000002"),
		Accent:     lipgloss.Color("#000003"),

		Muted:   lipgloss.Color("#000004"),
		Error:   lipgloss.Color("#000005"),
		Warning: lipgloss.Color("#000006"),
		Success: lipgloss.Color("#000007"),
		Info:    lipgloss.Color("#000008"),

		Border:       lipgloss.Color("#000009"),
		BorderActive: lipgloss.Color("#000010"),

		TableHeader: lipgloss.Color("#000011"),
		TableRow:    lipgloss.Color("#000012"),
		TableAltRow: lipgloss.Color("#000013"),

		PopupBackground: lipgloss.Color("#000014"),
		PopupBorder:     lipgloss.Color("#000015"),
		PopupTitle:      lipgloss.Color("#000010"),
	}
}

func TestGetFileStatusColor(t *testing.T) {
	colors := generateFakeLipGlossScheme()

	tests := []struct {
		symbol        string
		expectedColor lipgloss.Color
	}{
		{
			symbol:        "??",
			expectedColor: colors.Muted,
		},
		{
			symbol:        "A ",
			expectedColor: colors.Success,
		},
		{
			symbol:        "D ",
			expectedColor: colors.Error,
		},
		{
			symbol:        " D",
			expectedColor: colors.Error,
		},
		{
			symbol:        "R ",
			expectedColor: colors.Accent,
		},
		{
			symbol:        "U ",
			expectedColor: colors.Warning,
		},
		{
			symbol:        " U",
			expectedColor: colors.Warning,
		},
		{
			symbol:        "M ",
			expectedColor: colors.PopupTitle,
		},
		{
			symbol:        " M",
			expectedColor: colors.PopupTitle,
		},
		{
			symbol:        "asdf",
			expectedColor: colors.Foreground,
		},
		{
			symbol:        "",
			expectedColor: colors.Foreground,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Test Color %v", test.symbol), func(t *testing.T) {
			if color := getFileStatusColor(test.symbol, colors); color != test.expectedColor {
				t.Fatalf("expected %v found %v (symbol=%v)", test.expectedColor, color, test.symbol)
			}
		})
	}

}
