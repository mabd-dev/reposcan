package repostable

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
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
	// show controls whether a column is included for the current options.
	// A nil show function means the column is always visible.
	show func(Options) bool
	// cell renders the table cell for this column. It must not be nil because
	// createRows calls it for every active column.
	cell   func(report.RepoState, theme.Theme) string
	expand bool
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
		rows = append(rows, createRow(rs, theme, defs))
	}
	return rows
}

func createRow(rs report.RepoState, theme theme.Theme, defs []columnDef) table.Row {
	row := make(table.Row, 0, len(defs))
	for _, def := range defs {
		row = append(row, def.cell(rs, theme))
	}
	return row
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
	redistributeHiddenWidth(active, hiddenWidth)

	return active
}

func redistributeHiddenWidth(active []columnDef, hiddenWidth int) {
	if hiddenWidth == 0 || len(active) == 0 {
		return
	}
	for i := range active {
		if active[i].expand {
			active[i].widthPercent += hiddenWidth
			return
		}
	}
	active[len(active)-1].widthPercent += hiddenWidth
}

func stashColumnStr(rs report.RepoState) string {
	n := rs.StashCount()
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}

func getStateColumnStr(rs report.RepoState, theme theme.Theme) string {
	parts := make([]string, 0, len(rs.RemoteStatus))

	uc := len(rs.UncommitedFiles)
	ucStr := "⏳" + strconv.Itoa(uc) + " "

	for _, remoteStatus := range rs.RemoteStatus {
		statusParts := make([]string, 0, 3)

		if remoteStatus.Ahead > 0 {
			statusParts = append(statusParts, "↑"+strconv.Itoa(remoteStatus.Ahead))
		} else if remoteStatus.Ahead < 0 {
			statusParts = append(statusParts, "x")
		} else {
			statusParts = append(statusParts, "↑0")
		}

		if remoteStatus.Behind > 0 {
			statusParts = append(statusParts, "↓"+strconv.Itoa(remoteStatus.Behind))
		} else if remoteStatus.Behind < 0 {
			statusParts = append(statusParts, "x")
		} else {
			statusParts = append(statusParts, "↓0")
		}

		if remoteStatus.Remote != "" && !(len(rs.RemoteStatus) == 1 && remoteStatus.Remote == "origin") {
			remoteName := theme.Styles.Base.Render("(" + remoteStatus.Remote + ")")
			statusParts = append(statusParts, remoteName)
		}

		parts = append(parts, strings.Join(statusParts, " "))
	}

	return ucStr + strings.Join(parts, " | ")
}

func getRepoIndex(repos []report.RepoState, id string) int {
	for i, s := range repos {
		if s.ID == id {
			return i
		}
	}
	return -1
}
