package repodetails

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {

	if m.repoState == nil {
		return ""
	}

	style := m.theme.Styles.Base.Foreground(m.theme.Colors.Info).Bold(true)

	lines := []string{
		fmt.Sprintf("%s %s", style.Render("Path:"), m.repoState.Path),
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
		lines = append(lines, "  "+fileStyle.Render(f))
	}

	if trimUncommitedFiles {
		more := len(m.repoState.UncommitedFiles) - maxUncommitedFilesToShow
		lines = append(lines, fileStyle.Render("  ... (+"+strconv.Itoa(more)+" more)"))
	}

	return lines
}
