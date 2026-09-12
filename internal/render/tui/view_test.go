package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/render/tui/repodetails"
	"github.com/mabd-dev/reposcan/internal/render/tui/repostable"
	rth "github.com/mabd-dev/reposcan/internal/render/tui/repostableheader"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func TestViewIncludesRepositoryHeader(t *testing.T) {
	colors := theme.ColorScheme{Foreground: lipgloss.Color("#ffffff"), Muted: lipgloss.Color("#888888")}
	theme := theme.Theme{Colors: colors, Styles: theme.Styles{Base: lipgloss.NewStyle(), Muted: lipgloss.NewStyle()}}
	report := report.ScanReport{RepoStates: []report.RepoState{{Repo: "alpha", Branch: "main", ID: "alpha"}}}
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

	view := m.View().Content
	for _, want := range []string{"reposcan", "1 repos"} {
		if !strings.Contains(view, want) {
			t.Fatalf("production view missing %q:\n%s", want, view)
		}
	}
}
