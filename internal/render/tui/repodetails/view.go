package repodetails

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/theme"
)

func (m *Model) View() string {

	if m.repoState == nil {
		return ""
	}

	style := m.theme.Styles.Base.Foreground(m.theme.Colors.Info).Bold(true)
	pathStyle := m.theme.Styles.Base.Foreground(m.theme.Colors.Foreground)

	lines := []string{
		fmt.Sprintf("%s %s", style.Render("Path:"), pathStyle.Render(m.repoState.Path)),
		style.Render("File Changes:"),
	}

	lines = append(lines, m.buildUncommittedFiles()...)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *Model) buildUncommittedFiles() []string {

	files := m.repoState.UncommitedFiles
	if len(files) == 0 {
		return []string{
			m.theme.Styles.Muted.Render("    no changes"),
		}
	}

	lines := []string{}

	fileStyle := m.theme.Styles.Base.Foreground(m.theme.Colors.Foreground)

	maxUncommitedFilesToShow := m.height - len(files) - 1
	trimUncommitedFiles := len(files) > maxUncommitedFilesToShow

	if trimUncommitedFiles {
		files = files[:maxUncommitedFilesToShow]
	}

	for _, f := range files {
		changeSymbol := f[:2]
		color := getFileStatusColor(changeSymbol, m.theme.Colors)
		lines = append(lines, "  "+fileStyle.Foreground(color).Render(f))
	}

	if trimUncommitedFiles {
		more := len(m.repoState.UncommitedFiles) - maxUncommitedFilesToShow
		lines = append(lines, fileStyle.Render("  ... (+"+strconv.Itoa(more)+" more)"))
	}

	return lines
}

func getFileStatusColor(symbol string, colors theme.LipglossScheme) lipgloss.Color {
	if len(symbol) != 2 {
		return colors.Foreground
	}

	staged := string(symbol[0])
	unstaged := string(symbol[1])

	if symbol == "??" {
		return colors.Muted
	}

	if staged == "A" {
		return colors.Success
	}

	if staged == "D" || unstaged == "D" {
		return colors.Error
	}

	if staged == "R" {
		return colors.Accent
	}

	if staged == "U" || unstaged == "U" {
		return colors.Warning
	}

	if staged == "M" || unstaged == "M" {
		return colors.PopupTitle
	}

	return colors.Foreground
}
