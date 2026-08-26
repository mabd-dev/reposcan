package repodetails

import (
	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if len(m.tabs) == 0 {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		case "right", "l":
			m.selectedTabIndex = (m.selectedTabIndex + 1) % len(m.tabs)
			return m, nil
		case "left", "h":
			m.selectedTabIndex = (len(m.tabs) + m.selectedTabIndex - 1) % len(m.tabs)
			return m, nil
		}
	}
	return m, nil
}
