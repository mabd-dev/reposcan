package repostable

import tea "github.com/charmbracelet/bubbletea"

func (rt Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	rt.tbl, cmd = rt.tbl.Update(msg)

	return rt, cmd
}
