package tui

import (
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
)

type FocusState int

const (
	FocusReposTable FocusState = iota
	FocusReposFilter
	FocusKeybindingPopup
)

func (m Model) currentFocus() FocusState {
	if len(m.focusStack) == 0 {
		return FocusReposTable
	}
	return m.focusStack[len(m.focusStack)-1]
}

func (m *Model) pushFocus(state FocusState) {
	m.blurCurrentModel()
	m.focusStack = append(m.focusStack, state)
	m.focusCurrentModel()
}

func (m *Model) popFocus(reset bool) Model {
	m.blurCurrentModel()
	if reset {
		m.resetCurrentModel()
	}

	if len(m.focusStack) > 1 {
		m.focusStack = m.focusStack[:len(m.focusStack)-1]
	}

	m.focusCurrentModel()
	if reset {
		m.resetCurrentModel()
	}

	return *m
}

func (m *Model) focusCurrentModel() {
	switch m.currentFocus() {
	case FocusReposTable:
		m.reposTable.Focus()
	case FocusReposFilter:
		m.reposFilter.Focus()
	case FocusKeybindingPopup:
		break
	}
}

func (m *Model) blurCurrentModel() {
	switch m.currentFocus() {
	case FocusReposTable:
		m.reposTable.Blur()
	case FocusReposFilter:
		m.reposFilter.Blur()
	case FocusKeybindingPopup:
		break
	}
}

func (m *Model) resetCurrentModel() {
	switch m.currentFocus() {
	case FocusReposTable:
		m.reposTable.Filter("")
	case FocusReposFilter:
		m.reposFilter.SetValue("")
	case FocusKeybindingPopup:
		break
	}
}

func (m *Model) keybindings() []common.Keybinding {
	switch m.currentFocus() {
	case FocusReposTable:
		return reposTableKeybindings
	case FocusReposFilter:
		return reposTableFilterKeybindings
	case FocusKeybindingPopup:
		return helpPopupKeybindings
	}
	return []common.Keybinding{}
}
