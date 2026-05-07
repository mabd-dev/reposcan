package vcs

import (
	"testing"

	"github.com/mabd-dev/reposcan/pkg/report"
)

type stubActionProvider struct {
	repoType Type
}

func (p stubActionProvider) Type() Type {
	return p.repoType
}

func (p stubActionProvider) CheckRepoState(path string) (report.RepoState, []string) {
	return report.RepoState{Path: path, VCSType: string(p.repoType)}, nil
}

func (p stubActionProvider) Fetch(path string) (string, error) {
	return path, nil
}

func (p stubActionProvider) Push(path string) (string, error) {
	return path, nil
}

func (p stubActionProvider) Pull(path string) (string, error) {
	return path, nil
}

func TestRegistryGetActionProvider(t *testing.T) {
	registry := NewRegistry(
		stubProvider{repoType: TypeGit},
		stubActionProvider{repoType: TypeJJ},
	)

	if _, ok := registry.GetActionProvider(TypeGit); ok {
		t.Fatal("expected git stub to not implement ActionProvider")
	}

	actionProvider, ok := registry.GetActionProvider(TypeJJ)
	if !ok {
		t.Fatal("expected jj stub to implement ActionProvider")
	}

	output, err := actionProvider.Fetch("/tmp/repo")
	if err != nil {
		t.Fatalf("expected fetch to succeed: %v", err)
	}

	if output != "/tmp/repo" {
		t.Fatalf("expected fetch output to be repo path, got %q", output)
	}
}
