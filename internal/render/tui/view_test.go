package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/render/tui/repodetails"
	"github.com/mabd-dev/reposcan/internal/render/tui/repostable"
	rth "github.com/mabd-dev/reposcan/internal/render/tui/repostableheader"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func headerTestModel() Model {
	colors := theme.ColorScheme{Foreground: lipgloss.Color("#ffffff"), Muted: lipgloss.Color("#888888")}
	theme := theme.Theme{Colors: colors, Styles: theme.CreateStyles(colors)}
	report := report.ScanReport{RepoStates: []report.RepoState{{Repo: "alpha", Branch: "main", ID: "alpha", Path: "/workspace/" + strings.Repeat("long-path/", 12)}}}
	header := rth.Header{Theme: theme}
	header.SetReport(report, false)

	m := Model{
		width:       80,
		height:      24,
		theme:       theme,
		rtHeader:    header,
		reposTable:  repostable.New(theme, report, 80, 12, repostable.Options{}),
		repoDetails: repodetails.New(&report.RepoStates[0], theme),
		focusStack:  []FocusState{FocusReposTable},
	}

	return m
}

func TestViewIncludesRepositoryHeader(t *testing.T) {
	m := headerTestModel()
	view := m.View().Content
	for _, want := range []string{"reposcan", "1 repos"} {
		if !strings.Contains(view, want) {
			t.Fatalf("production view missing %q:\n%s", want, view)
		}
	}
}

func TestViewHeaderUpdatesAfterReportRefresh(t *testing.T) {
	m := headerTestModel()
	updated, _ := m.Update(tea.KeyPressMsg{Text: "r"})
	refreshed := report.ScanReport{RepoStates: []report.RepoState{
		{Repo: "alpha", Branch: "main", ID: "alpha"},
		{Repo: "beta", Branch: "main", ID: "beta"},
	}}
	updated, _ = updated.Update(generateReportResponse{report: refreshed})
	view := updated.View().Content
	if !strings.Contains(view, "2 repos") {
		t.Fatalf("refreshed production view has stale repository count:\n%s", view)
	}
}

func TestViewHeaderUsesCurrentTheme(t *testing.T) {
	m := headerTestModel()
	m.theme.Styles.Base = lipgloss.NewStyle().Transform(func(s string) string {
		return strings.ToUpper(s)
	})
	if view := m.View().Content; !strings.Contains(view, "REPOSCAN") {
		t.Fatalf("header did not use current theme:\n%s", view)
	}
}

func TestViewFitsTerminalAfterWrappingFooter(t *testing.T) {
	m := headerTestModel()
	view := m.View().Content
	if got := lipgloss.Height(view); got != m.height {
		t.Fatalf("view height = %d, want terminal height %d", got, m.height)
	}
}
