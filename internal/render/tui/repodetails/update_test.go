package repodetails

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

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
