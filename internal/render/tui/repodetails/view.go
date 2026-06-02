package repodetails

import (
	"fmt"
	"image/color"
	"strconv"

	"charm.land/lipgloss/v2"
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
		m.buildTabs(),
	}

	switch m.selectedTab().key {
	case tabChanges:
		lines = append(lines, m.buildUncommittedFiles()...)
	case tabStashes:
		lines = append(lines, m.buildStashes()...)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *Model) buildTabs() string {
	styles := m.theme.Styles
	colors := m.theme.Colors

	lines := []string{}

	for _, tab := range m.tabs {
		line := ""
		styledName := ""
		styledHighlightedText := ""
		if tab.key == m.selectedTab().key {
			styledHighlightedText = styles.Base.Foreground(colors.Accent).Render(tab.highlightedText)
			styledName = styles.Base.Foreground(colors.Foreground).Render(tab.name)
			line = styles.Box.Render(fmt.Sprintf("%v %v", styledName, styledHighlightedText))
		} else {
			styledHighlightedText = styles.Base.Foreground(colors.Muted).Render(tab.highlightedText)
			styledName = styles.Base.Foreground(colors.Muted).Render(tab.name)
			line = styles.BoxMuted.Render(fmt.Sprintf("%v %v", styledName, styledHighlightedText))
		}
		lines = append(lines, line)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, lines...)
}

func (m *Model) buildUncommittedFiles() []string {
	files := m.repoState.UncommitedFiles
	if len(files) == 0 {
		return []string{
			m.theme.Styles.Muted.Render("no changes"),
		}
	}

	lines := []string{}

	maxUncommitedFilesToShow := m.height - 3
	if maxUncommitedFilesToShow <= 0 {
		return lines
	}

	fileStyle := m.theme.Styles.Base.Foreground(m.theme.Colors.Foreground)

	trimUncommitedFiles := len(files) > m.height-3
	if trimUncommitedFiles {
		files = files[:maxUncommitedFilesToShow]
	}

	for _, f := range files {
		changeSymbol := f[:2]
		color := getFileStatusColor(changeSymbol, m.theme.Colors)
		lines = append(lines, fileStyle.Foreground(color).Render(f))
	}

	if trimUncommitedFiles {
		more := len(m.repoState.UncommitedFiles) - maxUncommitedFilesToShow
		lines = append(lines, fileStyle.Render("  ... (+"+strconv.Itoa(more)+" more)"))
	}

	return lines
}

func (m *Model) buildStashes() []string {
	if len(m.repoState.Stashes) == 0 {
		return []string{
			m.theme.Styles.Muted.Render("no stashes"),
		}
	}

	lines := []string{}
	for _, line := range m.repoState.Stashes {
		lines = append(lines, m.theme.Styles.Base.Foreground(m.theme.Colors.Foreground).Render(line))
	}
	return lines
}

func getFileStatusColor(symbol string, colors theme.ColorScheme) color.Color {
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
