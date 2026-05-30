package repostable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

const (
	RepoW        = 32
	BranchW      = 20
	VCSW         = 6
	StashW       = 6
	RemoteStateW = 36
)

type columnDef struct {
	title        string
	widthPercent int
	show         func(Options) bool
	cell         func(report.RepoState, theme.Theme) string
	expand       bool
}

func createColumns(maxWidth int, options Options) []table.Column {
	defs := activeColumnDefs(options)
	columns := make([]table.Column, 0, len(defs))
	for _, def := range defs {
		columns = append(columns, table.Column{
			Title: def.title,
			Width: maxWidth * def.widthPercent / 100,
		})
	}
	return columns
}

func createRows(repoStates []report.RepoState, theme theme.Theme, options Options) []table.Row {
	defs := activeColumnDefs(options)
	rows := make([]table.Row, 0, len(repoStates))
	for _, rs := range repoStates {
		row := make(table.Row, 0, len(defs))
		for _, def := range defs {
			row = append(row, def.cell(rs, theme))
		}
		rows = append(rows, row)
	}
	return rows
}

func activeColumnDefs(options Options) []columnDef {
	defs := []columnDef{
		{
			title:        "Repo",
			widthPercent: RepoW,
			cell: func(rs report.RepoState, _ theme.Theme) string {
				return rs.Repo
			},
		},
		{
			title:        "Branch",
			widthPercent: BranchW,
			cell: func(rs report.RepoState, _ theme.Theme) string {
				return rs.Branch
			},
		},
		{
			title:        "VCS",
			widthPercent: VCSW,
			show: func(options Options) bool {
				return options.ShowVCS
			},
			cell: func(rs report.RepoState, _ theme.Theme) string {
				return rs.VCSType
			},
		},
		{
			title:        "Stash",
			widthPercent: StashW,
			cell: func(rs report.RepoState, _ theme.Theme) string {
				return stashColumnStr(rs)
			},
		},
		{
			title:        "State",
			widthPercent: RemoteStateW,
			expand:       true,
			cell: func(rs report.RepoState, theme theme.Theme) string {
				return getStateColumnStr(rs, theme)
			},
		},
	}

	hiddenWidth := 0
	active := make([]columnDef, 0, len(defs))
	for _, def := range defs {
		if def.show != nil && !def.show(options) {
			hiddenWidth += def.widthPercent
			continue
		}
		active = append(active, def)
	}
	for i := range active {
		if active[i].expand {
			active[i].widthPercent += hiddenWidth
		}
	}

	return active
}

func stashColumnStr(rs report.RepoState) string {
	n := rs.StashCount()
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d", n)
}

func getStateColumnStr(rs report.RepoState, theme theme.Theme) string {
	parts := []string{}

	uc := len(rs.UncommitedFiles)
	ucStr := fmt.Sprintf("⏳%-d ", uc)

	for _, remoteStatus := range rs.RemoteStatus {
		var statusParts []string

		if remoteStatus.Ahead > 0 {
			statusParts = append(statusParts, fmt.Sprintf("↑%-d", remoteStatus.Ahead))
		} else if remoteStatus.Ahead < 0 {
			statusParts = append(statusParts, "x")
		} else {
			statusParts = append(statusParts, fmt.Sprintf("↑%-d", 0))
		}

		if remoteStatus.Behind > 0 {
			statusParts = append(statusParts, fmt.Sprintf("↓%-d", remoteStatus.Behind))
		} else if remoteStatus.Behind < 0 {
			statusParts = append(statusParts, "x")
		} else {
			statusParts = append(statusParts, fmt.Sprintf("↓%-d", 0))
		}

		if remoteStatus.Remote != "" && !(len(rs.RemoteStatus) == 1 && remoteStatus.Remote == "origin") {
			remoteName := theme.Styles.Base.Render(fmt.Sprintf("(%s)", remoteStatus.Remote))
			statusParts = append(statusParts, remoteName)
		}

		parts = append(parts, strings.Join(statusParts, " "))
	}

	// Combine uncommitted count with all remote statuses, separated by " | "
	s := ucStr
	s += strings.Join(parts, " | ")

	return s
}

func setKeymaps(km table.KeyMap) {
	km.LineUp.SetKeys("up", "k")
	km.LineDown.SetKeys("down", "j")
	km.PageUp.SetKeys("pgup", tea.KeyCtrlU.String())
	km.PageDown.SetKeys("pgdn", tea.KeyCtrlD.String())
	km.GotoTop.SetKeys("home", "g")
	km.GotoBottom.SetKeys("end", "G")
}

func getRepoIndex(repos []report.RepoState, id string) int {
	for i, s := range repos {
		if s.ID == id {
			return i
		}
	}
	return -1
}
