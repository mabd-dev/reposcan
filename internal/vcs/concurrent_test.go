package vcs

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/pkg/report"
)

type stubProvider struct {
	repoType Type
	states   map[string]report.RepoState
	warnings map[string][]string
}

func (p stubProvider) Type() Type {
	return p.repoType
}

func (p stubProvider) CheckRepoState(path string) (report.RepoState, []string) {
	state := p.states[path]
	state.Path = path
	state.VCSType = string(p.repoType)

	return state, p.warnings[path]
}

func TestGetRepoStatesConcurrent_SortsResultsAndPreservesWarnings(t *testing.T) {
	repos := []RepoPath{
		{Path: "/tmp/z-jj", Type: TypeJJ},
		{Path: "/tmp/a-git", Type: TypeGit},
		{Path: "/tmp/m-jj", Type: TypeJJ},
	}

	registry := NewRegistry(
		stubProvider{
			repoType: TypeGit,
			states: map[string]report.RepoState{
				"/tmp/a-git": {Repo: "a"},
			},
			warnings: map[string][]string{
				"/tmp/a-git": {"git warning"},
			},
		},
		stubProvider{
			repoType: TypeJJ,
			states: map[string]report.RepoState{
				"/tmp/z-jj": {Repo: "z"},
				"/tmp/m-jj": {Repo: "m"},
			},
			warnings: map[string][]string{
				"/tmp/z-jj": {"jj warning"},
			},
		},
	)

	states, warnings := GetRepoStatesConcurrent(repos, registry, 2)

	gotPaths := []string{}
	for _, state := range states {
		gotPaths = append(gotPaths, state.Path)
	}

	wantPaths := []string{"/tmp/a-git", "/tmp/m-jj", "/tmp/z-jj"}
	if !reflect.DeepEqual(gotPaths, wantPaths) {
		t.Fatalf("expected sorted paths %v, got %v", wantPaths, gotPaths)
	}

	slices.Sort(warnings)
	wantWarnings := []string{"git warning", "jj warning"}
	slices.Sort(wantWarnings)

	if !reflect.DeepEqual(warnings, wantWarnings) {
		t.Fatalf("expected warnings %v, got %v", wantWarnings, warnings)
	}
}

func TestGetRepoStatesConcurrent_WarnsWhenProviderIsMissing(t *testing.T) {
	repos := []RepoPath{{Path: "/tmp/repo", Type: TypeJJ}}

	states, warnings := GetRepoStatesConcurrent(repos, NewRegistry(), 1)

	if len(states) != 0 {
		t.Fatalf("expected no states, got %v", states)
	}

	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}

	if !strings.Contains(warnings[0], `No provider registered for vcs type="jj"`) {
		t.Fatalf("unexpected warning: %v", warnings)
	}
}
