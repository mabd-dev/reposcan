package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/repodetails"
	"github.com/mabd-dev/reposcan/internal/render/tui/repostable"
	rth "github.com/mabd-dev/reposcan/internal/render/tui/repostableheader"
	"github.com/mabd-dev/reposcan/internal/theme"
)

type reposFilter struct {
	textInput textinput.Model
	show      bool
}

func (rf reposFilter) IsVisible() bool {
	return rf.show && rf.textInput.Focused()
}

type Model struct {
	reposTable        repostable.Model
	repoDetails       repodetails.Model
	rtHeader          rth.Header
	alerts            alerts.AlertModel
	isPushing         bool
	width             int
	height            int
	reposBeingUpdated []string
	warnings          []string
	showHelp          bool
	reposFilter       reposFilter
	theme             theme.Theme
}
