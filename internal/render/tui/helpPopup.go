package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func generateHelpPopup(m Model) string {
	helpBox := PopupStyle.
		Render(`
Keybindings:
  ↑/↓    - Navigate up and down (or using j/k)
  Enter  - Open git repository report details
  ?      - Keybindings
  p	  - Pull changes
  P	  - Push changes 
  f	  - Fetch remote
  q	  - Quit
        `)

	popup := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpBox)
	return popup
}
