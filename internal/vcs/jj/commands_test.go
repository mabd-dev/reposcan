package jj

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

// TestCommandError verifies that command failures retain their invocation,
// stderr, and underlying error for diagnostics and errors.Is checks.
func TestCommandError(t *testing.T) {
	underlying := errors.New("exit status 1")
	err := commandError{
		Binary:   "jj",
		RepoPath: "/tmp/repo",
		Args:     []string{"git", "fetch"},
		Stderr:   "network unavailable",
		Err:      underlying,
	}

	want := `command="jj -R /tmp/repo git fetch" failed: exit status 1: network unavailable`
	if got := err.Error(); got != want {
		t.Fatalf("commandError.Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, underlying) {
		t.Fatalf("commandError should unwrap to %v", underlying)
	}
}

// TestPublicCommandWrappers verifies the package's current exported command
// surface. These entry points hardcode "jj", so the fake must be on PATH.
func TestPublicCommandWrappers(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "checkout")
	responses := map[string]fakeJJResponse{
		fakeJJCommandKey("git", "remote", "list"): {Stdout: "origin https://example.com/acme/widgets.git\n"},
		branchCommandKey():                        {Stdout: "main*?|abc123\n"},
		fakeJJCommandKey("diff", "--summary"):     {Stdout: "M README.md\n"},
		trackedBookmarksCommandKey():              {},
		fakeJJCommandKey("version"):               {Stdout: "jj test version\n"},
	}
	installFakeJJ(t, responses)

	repoName, err := GetRepoName(repoPath)
	if err != nil || repoName != "widgets" {
		t.Fatalf("GetRepoName() = %q, %v, want widgets, nil", repoName, err)
	}

	remoteURL, err := GetFirstRemoteURL(repoPath)
	if err != nil || remoteURL != "https://example.com/acme/widgets.git" {
		t.Fatalf("GetFirstRemoteURL() = %q, %v", remoteURL, err)
	}

	branch, err := GetBranchDisplay(repoPath)
	if err != nil || branch != "main" {
		t.Fatalf("GetBranchDisplay() = %q, %v, want main, nil", branch, err)
	}

	files, err := GetUncommittedFiles(repoPath)
	if err != nil || !reflect.DeepEqual(files, []string{"M README.md"}) {
		t.Fatalf("GetUncommittedFiles() = %v, %v", files, err)
	}

	outgoing, err := GetOutgoingCommits(repoPath)
	if err != nil || len(outgoing) != 0 {
		t.Fatalf("GetOutgoingCommits() = %v, %v, want empty, nil", outgoing, err)
	}

	incoming, err := GetIncomingCommits(repoPath)
	if err != nil || len(incoming) != 0 {
		t.Fatalf("GetIncomingCommits() = %v, %v, want empty, nil", incoming, err)
	}

	output, err := RunJJCommand(repoPath, "version")
	if err != nil || output != "jj test version\n" {
		t.Fatalf("RunJJCommand() = %q, %v", output, err)
	}
}

// TestGetRepoName covers command failures, local-directory fallback, supported
// remote formats, and fallback when a remote cannot be parsed.
func TestGetRepoName(t *testing.T) {
	tests := []struct {
		name       string
		remoteList fakeJJResponse
		want       string
		wantErr    bool
	}{
		{
			name:       "remote command fails",
			remoteList: fakeJJResponse{Stderr: "not a repository", ExitCode: 1},
			wantErr:    true,
		},
		{
			name: "falls back to directory without remote",
			want: "checkout",
		},
		{
			name:       "uses parsed remote name",
			remoteList: fakeJJResponse{Stdout: "origin git@github.com:acme/widgets.git\n"},
			want:       "widgets",
		},
		{
			name:       "falls back when remote cannot be parsed",
			remoteList: fakeJJResponse{Stdout: "origin https://\n"},
			want:       "checkout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := filepath.Join(t.TempDir(), "checkout")
			binary := useFakeJJ(t, map[string]fakeJJResponse{
				fakeJJCommandKey("git", "remote", "list"): tt.remoteList,
			})

			got, err := getRepoName(binary, repoPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getRepoName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("getRepoName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestGetFirstRemoteURL verifies that informational, blank, and malformed lines
// are ignored before returning the first valid remote URL.
func TestGetFirstRemoteURL(t *testing.T) {
	binary := useFakeJJ(t, map[string]fakeJJResponse{
		fakeJJCommandKey("git", "remote", "list"): {
			Stdout: "\nDone importing changes from the underlying Git repo.\ninvalid\norigin ssh://git@example.com/acme/widgets.git extra\n",
		},
	})

	got, err := getFirstRemoteURL(binary, t.TempDir())
	if err != nil {
		t.Fatalf("getFirstRemoteURL() error = %v", err)
	}
	if want := "ssh://git@example.com/acme/widgets.git"; got != want {
		t.Fatalf("getFirstRemoteURL() = %q, want %q", got, want)
	}
}

// TestGetBranchDisplay covers command failures and every display fallback,
// including local bookmark cleanup and change-ID selection.
func TestGetBranchDisplay(t *testing.T) {
	tests := []struct {
		name     string
		response fakeJJResponse
		want     string
		wantErr  bool
	}{
		{name: "command fails", response: fakeJJResponse{Stderr: "bad revset", ExitCode: 1}, wantErr: true},
		{name: "empty output", want: "-"},
		{name: "change id fallback", response: fakeJJResponse{Stdout: "|abc123\n"}, want: "abc123"},
		{name: "malformed line fallback", response: fakeJJResponse{Stdout: "fallback\n\nsecond\n"}, want: "fallback"},
		{
			name:     "local bookmarks",
			response: fakeJJResponse{Stdout: "origin@origin*?, , main*?|abc123\n"},
			want:     "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binary := useFakeJJ(t, map[string]fakeJJResponse{branchCommandKey(): tt.response})
			got, err := getBranchDisplay(binary, t.TempDir())
			if (err != nil) != tt.wantErr {
				t.Fatalf("getBranchDisplay() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("getBranchDisplay() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestGetUncommittedFiles verifies diff failures are propagated and successful
// summaries are trimmed without retaining blank lines.
func TestGetUncommittedFiles(t *testing.T) {
	t.Run("command fails", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			fakeJJCommandKey("diff", "--summary"): {Stderr: "diff failed", ExitCode: 1},
		})
		if _, err := getUncommittedFiles(binary, t.TempDir()); err == nil {
			t.Fatal("getUncommittedFiles() error = nil, want error")
		}
	})

	t.Run("trims and skips blank lines", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			fakeJJCommandKey("diff", "--summary"): {Stdout: " M README.md \n\nA docs.md\n"},
		})
		got, err := getUncommittedFiles(binary, t.TempDir())
		if err != nil {
			t.Fatalf("getUncommittedFiles() error = %v", err)
		}
		want := []string{"M README.md", "A docs.md"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("getUncommittedFiles() = %v, want %v", got, want)
		}
	})
}

// TestRunJJCommandError verifies that command stderr is trimmed and preserved
// in the structured error returned to callers.
func TestRunJJCommandError(t *testing.T) {
	binary := useFakeJJ(t, map[string]fakeJJResponse{
		fakeJJCommandKey("git", "fetch"): {Stderr: "authentication failed\n", ExitCode: 1},
	})
	_, err := runJJCommand(binary, "/tmp/repo", "git", "fetch")
	if err == nil {
		t.Fatal("runJJCommand() error = nil, want error")
	}
	var commandErr commandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("runJJCommand() error = %T, want commandError", err)
	}
	if commandErr.Stderr != "authentication failed" {
		t.Fatalf("runJJCommand() stderr = %q, want trimmed stderr", commandErr.Stderr)
	}
}

// TestParseRepoName verifies repository names from SCP-style remotes, URLs, and
// filesystem paths, plus rejection of an unparseable remote.
func TestParseRepoName(t *testing.T) {
	tests := []struct {
		name   string
		remote string
		want   string
		ok     bool
	}{
		{name: "scp syntax", remote: "git@github.com:acme/widgets.git", want: "widgets", ok: true},
		{name: "URL", remote: "https://github.com/acme/widgets.git", want: "widgets", ok: true},
		{name: "filesystem fallback", remote: `C:\checkouts\widgets.git`, want: "widgets", ok: true},
		{name: "unparseable", remote: "https://", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseRepoName(tt.remote)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("parseRepoName(%q) = %q, %v, want %q, %v", tt.remote, got, ok, tt.want, tt.ok)
			}
		})
	}
}

// branchCommandKey records the exact jj log invocation used to select the
// current bookmark display.
func branchCommandKey() string {
	return fakeJJCommandKey(
		"log",
		"-r",
		"latest(ancestors(@) & bookmarks()) | @",
		"--no-graph",
		"-T",
		`bookmarks.join(",") ++ "|" ++ change_id.short() ++ "\n"`,
	)
}

// trackedBookmarksCommandKey records the exact template used to list tracked
// bookmark and remote pairs.
func trackedBookmarksCommandKey() string {
	return fakeJJCommandKey(
		"bookmark",
		"list",
		"--tracked",
		"-T",
		`name ++ "|" ++ remote ++ "\n"`,
	)
}

// untrackedRemotesCommandKey records the bookmark-specific invocation used to
// discover remote bookmarks that are not tracked locally.
func untrackedRemotesCommandKey(bookmarkName string) string {
	return fakeJJCommandKey(
		"bookmark",
		"list",
		"--all",
		bookmarkName,
		"-T",
		`name ++ "|" ++ remote ++ "\n"`,
	)
}

// commitLogCommandKey records the exact jj log invocation used to collect
// commit summaries for a generated revset.
func commitLogCommandKey(revset string) string {
	return fakeJJCommandKey(
		"log",
		"-r",
		revset,
		"--no-graph",
		"-T",
		`commit_id.short() ++ "|" ++ description.first_line() ++ "\n"`,
	)
}
