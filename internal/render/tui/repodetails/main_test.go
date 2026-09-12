package repodetails

import (
	"testing"

	"charm.land/lipgloss/v2"
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

func TestNewWithNilRepoStateDoesNotPanic(t *testing.T) {
	model := New(nil, theme.Theme{})

	assert.Equal(t, 0, model.selectedTabIndex)
	assert.Empty(t, model.tabs)
	assert.Nil(t, model.repoState)
}

func TestNewWithNilRepoStateReturnsEmptyView(t *testing.T) {
	model := New(nil, theme.Theme{})

	output := model.View()
	assert.Equal(t, "", output)
}

func TestUpdateDataWithNilRepoStateDoesNotPanic(t *testing.T) {
	repoState := createRepoState([]string{"f1"}, []string{"s1"})
	model := createModel(0, repoState)

	assert.Equal(t, 2, len(model.tabs))

	model.UpdateData(nil)

	assert.Nil(t, model.repoState)
	assert.Empty(t, model.tabs)
}

func TestUpdateDataWithNilRepoStateReturnsEmptyView(t *testing.T) {
	repoState := createRepoState([]string{"f1"}, []string{"s1"})
	model := createModel(0, repoState)

	model.UpdateData(nil)

	output := model.View()
	assert.Equal(t, "", output)
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

func TestInitReturnsNoCommand(t *testing.T) {
	model := New(nil, theme.Theme{})

	assert.Nil(t, model.Init())
}

func TestUpdateThemeChangesModelTheme(t *testing.T) {
	model := New(nil, theme.Theme{})
	newTheme := theme.Theme{
		Colors: theme.ColorScheme{
			Name:   "test theme",
			Accent: lipgloss.Color("#123456"),
		},
	}

	model.UpdateTheme(newTheme)

	assert.Equal(t, newTheme, model.theme)
}
