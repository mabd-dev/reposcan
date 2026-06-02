package repodetails

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/stretchr/testify/assert"
)

func generateFakeLipGlossScheme() theme.ColorScheme {
	return theme.ColorScheme{
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
	assert := assert.New(t)
	colors := generateFakeLipGlossScheme()

	tests := []struct {
		symbol        string
		expectedColor color.Color
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
			color := getFileStatusColor(test.symbol, colors)
			assert.Equal(test.expectedColor, color)
		})
	}

}

func TestUncommitedFilesRenderWhenChangesIsSelected(t *testing.T) {
	repoState := createRepoState([]string{}, []string{})
	model := createModel(0, repoState)

	output := model.View()
	lines := strings.Split(output, "\n")

	assert.Equal(t, 5, len(lines))
	assert.Equal(t, "no changes", strings.TrimSpace(lines[4]))
}

func TestStashesRenderWhenChangesIsSelected(t *testing.T) {
	repoState := createRepoState([]string{}, []string{})
	model := createModel(1, repoState)

	output := model.View()
	lines := strings.Split(output, "\n")

	assert.Equal(t, 5, len(lines))
	assert.Equal(t, "no stashes", strings.TrimSpace(lines[4]))
}

func TestTooLowTabIndexReverBackToFirstTab(t *testing.T) {
	repoState := createRepoState([]string{}, []string{})
	model := createModel(-1, repoState)

	output := model.View()
	lines := strings.Split(output, "\n")

	assert.Equal(t, 5, len(lines))
	assert.Equal(t, "no changes", strings.TrimSpace(lines[4]))
}

func TestTooHighTabIndexReverBackToLastTab(t *testing.T) {
	repoState := createRepoState([]string{}, []string{})
	model := createModel(2, repoState)

	output := model.View()
	lines := strings.Split(output, "\n")

	assert.Equal(t, 5, len(lines))
	assert.Equal(t, "no stashes", strings.TrimSpace(lines[4]))
}

func TestRenderingUncommitedFiles(t *testing.T) {
	assert := assert.New(t)
	uncommitedFiles := []string{
		"file1", "file2", "file3",
	}
	repoState := createRepoState(uncommitedFiles, []string{})
	model := createModel(0, repoState)

	output := model.View()
	lines := strings.Split(output, "\n")

	assert.Equal(7, len(lines)) // lines: [path, space tab, tab line, content]

	for i, line := range uncommitedFiles {
		assert.Equal(line, strings.TrimSpace(lines[4+i]))
	}
}

func TestRenderingStashedFiles(t *testing.T) {
	assert := assert.New(t)
	stashedLines := []string{
		"file1", "file2", "file3",
	}
	repoState := createRepoState([]string{}, stashedLines)
	model := createModel(1, repoState)

	output := model.View()
	lines := strings.Split(output, "\n")

	assert.Equal(7, len(lines)) // lines: [path, space tab, tab line, content]

	for i, line := range stashedLines {
		assert.Equal(line, strings.TrimSpace(lines[4+i]))
	}
}
