package repodetails

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/stretchr/testify/assert"
)

func TestKeyPressesWithNilRepoStateDoNotPanic(t *testing.T) {
	model := New(nil, theme.Theme{})

	m, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	assert.Equal(t, 0, m.selectedTabIndex)

	m, _ = m.Update(tea.KeyPressMsg{Text: "l"})
	assert.Equal(t, 0, m.selectedTabIndex)

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	assert.Equal(t, 0, m.selectedTabIndex)

	m, _ = m.Update(tea.KeyPressMsg{Text: "h"})
	assert.Equal(t, 0, m.selectedTabIndex)
}

func TestKeyPressesMoveToCorrectTab(t *testing.T) {
	assert := assert.New(t)

	repoState := createRepoState([]string{}, []string{})
	m := createModel(0, repoState)

	// starts at 0
	assert.Equal(0, m.selectedTabIndex)

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	assert.Equal(1, m.selectedTabIndex)

	m, _ = m.Update(tea.KeyPressMsg{Text: "l"})
	assert.Equal(0, m.selectedTabIndex)

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	assert.Equal(1, m.selectedTabIndex)

	m, _ = m.Update(tea.KeyPressMsg{Text: "h"})
	assert.Equal(0, m.selectedTabIndex)
}

func TestUnhandledMessageDoesNotChangeSelectedTab(t *testing.T) {
	repoState := createRepoState([]string{}, []string{})
	model := createModel(1, repoState)

	updatedModel, cmd := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	assert.Equal(t, 1, updatedModel.selectedTabIndex)
	assert.Nil(t, cmd)
}
