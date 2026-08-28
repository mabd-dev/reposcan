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

func TestNewRegistry_SkipsNilProviders(t *testing.T) {
	registry := NewRegistry(nil, stubProvider{repoType: TypeGit})

	if _, ok := registry.Get(TypeGit); !ok {
		t.Fatal("expected git provider to be registered")
	}

	if len(registry.providers) != 1 {
		t.Fatalf("expected 1 registered provider, got %d", len(registry.providers))
	}
}

func TestRegister_InitializesNilProviderMap(t *testing.T) {
	var registry Registry

	registry.Register(stubProvider{repoType: TypeGit})

	if _, ok := registry.Get(TypeGit); !ok {
		t.Fatal("expected provider to be registered on zero-value Registry")
	}
}

func TestGet_NilReceiverReturnsNotFound(t *testing.T) {
	var registry *Registry

	provider, ok := registry.Get(TypeGit)
	if ok || provider != nil {
		t.Fatalf("expected no provider for nil receiver, got (%v, %v)", provider, ok)
	}
}

func TestGetActionProvider_UnregisteredTypeReturnsNotFound(t *testing.T) {
	registry := NewRegistry(stubActionProvider{repoType: TypeGit})

	actionProvider, ok := registry.GetActionProvider(TypeJJ)
	if ok || actionProvider != nil {
		t.Fatalf("expected no action provider for unregistered type, got (%v, %v)", actionProvider, ok)
	}
}
