package jj

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal/vcs"
)

// TestNewProvider verifies that the constructor selects the jj executable and
// advertises the provider type used by the VCS registry.
func TestNewProvider(t *testing.T) {
	provider := New()
	if provider.binary != "jj" {
		t.Fatalf("New().binary = %q, want jj", provider.binary)
	}
	if got := provider.Type(); got != vcs.TypeJJ {
		t.Fatalf("Provider.Type() = %q, want %q", got, vcs.TypeJJ)
	}
}

// TestProviderCheckRepoStateWithFakeJJ verifies that successful command results
// are assembled into a complete RepoState with per-remote ahead/behind counts.
func TestProviderCheckRepoStateWithFakeJJ(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "checkout")
	bookmark := trackedBookmark{Name: "main", Remote: "origin"}
	outgoingRevset := buildTrackedOutgoingRevset([]trackedBookmark{bookmark})
	incomingRevset := buildTrackedIncomingRevset([]trackedBookmark{bookmark})
	responses := map[string]fakeJJResponse{
		fakeJJCommandKey("git", "remote", "list"): {Stdout: "origin https://example.com/acme/widgets.git\n"},
		branchCommandKey():                        {Stdout: "main*?|abc123\n"},
		fakeJJCommandKey("diff", "--summary"):     {Stdout: "M README.md\n"},
		trackedBookmarksCommandKey():              {Stdout: "main|origin\n"},
		commitLogCommandKey(outgoingRevset):       {Stdout: "abc123|local change\n"},
		commitLogCommandKey(incomingRevset):       {Stdout: "def456|remote change\n"},
	}
	binary := useFakeJJ(t, responses)

	state, warnings := (&Provider{binary: binary}).CheckRepoState(repoPath)
	if len(warnings) != 0 {
		t.Fatalf("CheckRepoState() warnings = %v, want none", warnings)
	}
	if state.ID == "" {
		t.Fatal("CheckRepoState() ID is empty")
	}
	if state.Path != repoPath {
		t.Fatalf("CheckRepoState() path = %q, want %q", state.Path, repoPath)
	}
	if state.Repo != "widgets" {
		t.Fatalf("CheckRepoState() repo = %q, want widgets", state.Repo)
	}
	if state.Branch != "main" {
		t.Fatalf("CheckRepoState() branch = %q, want main", state.Branch)
	}
	if state.VCSType != string(vcs.TypeJJ) {
		t.Fatalf("CheckRepoState() VCS type = %q, want %q", state.VCSType, vcs.TypeJJ)
	}
	if !reflect.DeepEqual(state.UncommitedFiles, []string{"M README.md"}) {
		t.Fatalf("CheckRepoState() uncommitted files = %v", state.UncommitedFiles)
	}
	if len(state.RemoteStatus) != 1 {
		t.Fatalf("CheckRepoState() remote statuses = %v, want one", state.RemoteStatus)
	}
	remoteStatus := state.RemoteStatus[0]
	if remoteStatus.Remote != "origin" || remoteStatus.Ahead != 1 || remoteStatus.Behind != 1 {
		t.Fatalf("CheckRepoState() remote status = %#v", remoteStatus)
	}
	if !reflect.DeepEqual(remoteStatus.OutgoingCommits, []string{"abc123 local change"}) {
		t.Fatalf("CheckRepoState() outgoing commits = %v", remoteStatus.OutgoingCommits)
	}
}

// TestProviderActions verifies fetch output and the explicit unsupported-action
// contract for push and pull.
func TestProviderActions(t *testing.T) {
	t.Run("fetch", func(t *testing.T) {
		installFakeJJ(t, map[string]fakeJJResponse{
			fakeJJCommandKey("git", "fetch"): {Stdout: "fetch complete\n"},
		})
		output, err := New().Fetch(t.TempDir())
		if err != nil || output != "fetch complete\n" {
			t.Fatalf("Provider.Fetch() = %q, %v", output, err)
		}
	})

	t.Run("push", func(t *testing.T) {
		output, err := New().Push(t.TempDir())
		if output != "" {
			t.Fatalf("Provider.Push() output = %q, want empty", output)
		}
		if !errors.Is(err, ErrJJActionNotImplemented) {
			t.Fatalf("Provider.Push() error = %v, want ErrJJActionNotImplemented", err)
		}
		if !strings.Contains(err.Error(), "push bookmark behavior") {
			t.Fatalf("Provider.Push() error = %q", err)
		}
	})

	t.Run("pull", func(t *testing.T) {
		output, err := New().Pull(t.TempDir())
		if output != "" {
			t.Fatalf("Provider.Pull() output = %q, want empty", output)
		}
		if !errors.Is(err, ErrJJActionNotImplemented) {
			t.Fatalf("Provider.Pull() error = %v, want ErrJJActionNotImplemented", err)
		}
		if !strings.Contains(err.Error(), "no direct Git-equivalent") {
			t.Fatalf("Provider.Pull() error = %q", err)
		}
	})
}
