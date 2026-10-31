package repodetails

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"unicode"

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

	switch m.tabs[m.selectedTabIndex].key {
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
	box := lipgloss.NewStyle().Border(lipgloss.Border{Bottom: "━"})

	renderedTabs := []string{}

	for _, tab := range m.tabs {
		line := ""
		styledName := ""
		styledHighlightedText := ""

		if tab.key == m.tabs[m.selectedTabIndex].key {
			styledHighlightedText = styles.Base.Foreground(colors.Accent).Render(tab.highlightedText)
			styledName = styles.Base.Foreground(colors.Foreground).Render(tab.name)
			line = box.BorderForeground(colors.BorderActive).Render(fmt.Sprintf("%v %v", styledName, styledHighlightedText))
		} else {
			styledHighlightedText = styles.Base.Foreground(colors.Muted).Render(tab.highlightedText)
			styledName = styles.Base.Foreground(colors.Muted).Render(tab.name)
			line = box.BorderForeground(colors.Border).Render(fmt.Sprintf("%v %v", styledName, styledHighlightedText))
		}
		renderedTabs = append(renderedTabs, line)
	}

	return lipgloss.JoinHorizontal(lipgloss.Bottom, renderedTabs...)
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

	trimUncommitedFiles := len(files) > maxUncommitedFilesToShow
	if trimUncommitedFiles {
		files = files[:maxUncommitedFilesToShow]
	}

	fileStyle := m.theme.Styles.Base.Foreground(m.theme.Colors.Foreground)

	for _, f := range files {
		changeSymbol := f[:2]
		color := getFileStatusColor(changeSymbol, m.theme.Colors)
		lines = append(lines, fileStyle.Foreground(color).Render(escapeControlChars(f)))
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

// escapeControlChars shows control characters escaped, e.g. a newline as \n.
// git -z and jj print file names raw, so a name could otherwise span several
// rows or send terminal escape sequences.
func escapeControlChars(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) {
			quoted := strconv.QuoteRune(r)
			b.WriteString(quoted[1 : len(quoted)-1])
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
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
