package repodetails

import (
	"testing"

	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/stretchr/testify/assert"
)

func TestModelCreatedWithCorrectDefaults(t *testing.T) {
	repoState := createRepoState([]string{}, []string{})
	theme := theme.Theme{}
	model := New(&repoState, theme)

	assert.Equal(t, 0, model.selectedTabIndex)
	assert.Equal(t, 2, len(model.tabs))
}

func TestUpdatingRepoStateChangesTabsContent(t *testing.T) {
	repoState1 := createRepoState([]string{"f1", "f2"}, []string{})
	repoState2 := createRepoState([]string{"f1"}, []string{"s1"})

	model := createModel(0, repoState1)

	assert := assert.New(t)

	assert.Equal("(2)", model.tabs[0].highlightedText)
	assert.Equal("(0)", model.tabs[1].highlightedText)

	model.UpdateData(&repoState2)

	assert.Equal("(1)", model.tabs[0].highlightedText)
	assert.Equal("(1)", model.tabs[1].highlightedText)

}
