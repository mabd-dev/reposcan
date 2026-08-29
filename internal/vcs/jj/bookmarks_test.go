package jj

import (
	"reflect"
	"testing"
)

// TestGetTrackedBookmarkCommits verifies that both outgoing and incoming
// helpers handle bookmark lookup, empty results, log failures, and formatting.
func TestGetTrackedBookmarkCommits(t *testing.T) {
	bookmark := trackedBookmark{Name: "main", Remote: "origin"}
	directions := []struct {
		name   string
		revset string
		get    func(string, string) ([]string, error)
		commit string
		want   string
	}{
		{
			name:   "outgoing",
			revset: buildTrackedOutgoingRevset([]trackedBookmark{bookmark}),
			get:    getOutgoingCommits,
			commit: "abc123|first change\n",
			want:   "abc123 first change",
		},
		{
			name:   "incoming",
			revset: buildTrackedIncomingRevset([]trackedBookmark{bookmark}),
			get:    getIncomingCommits,
			commit: "def456|remote change\n",
			want:   "def456 remote change",
		},
	}

	for _, direction := range directions {
		t.Run(direction.name, func(t *testing.T) {
			tests := []struct {
				name      string
				responses map[string]fakeJJResponse
				want      []string
				wantErr   bool
			}{
				{
					name: "tracked bookmark command fails",
					responses: map[string]fakeJJResponse{
						trackedBookmarksCommandKey(): {Stderr: "bookmark failed", ExitCode: 1},
					},
					wantErr: true,
				},
				{
					name: "no tracked bookmarks",
					responses: map[string]fakeJJResponse{
						trackedBookmarksCommandKey(): {},
					},
					want: []string{},
				},
				{
					name: "log command fails",
					responses: map[string]fakeJJResponse{
						trackedBookmarksCommandKey():          {Stdout: "main|origin\n"},
						commitLogCommandKey(direction.revset): {Stderr: "log failed", ExitCode: 1},
					},
					wantErr: true,
				},
				{
					name: "returns commit summaries",
					responses: map[string]fakeJJResponse{
						trackedBookmarksCommandKey():          {Stdout: "main|origin\n"},
						commitLogCommandKey(direction.revset): {Stdout: direction.commit},
					},
					want: []string{direction.want},
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					binary := useFakeJJ(t, tt.responses)
					got, err := direction.get(binary, t.TempDir())
					if (err != nil) != tt.wantErr {
						t.Fatalf("commit lookup error = %v, wantErr %v", err, tt.wantErr)
					}
					if !reflect.DeepEqual(got, tt.want) {
						t.Fatalf("commit lookup = %v, want %v", got, tt.want)
					}
				})
			}
		})
	}
}

// TestGetBookmarkRemoteStatuses verifies bookmark cleanup, deduplication,
// tracked and untracked remote discovery, and outgoing/incoming failure paths.
func TestGetBookmarkRemoteStatuses(t *testing.T) {
	repoPath := t.TempDir()
	bookmark := trackedBookmark{Name: "main", Remote: "origin"}
	outgoingRevset := buildTrackedOutgoingRevset([]trackedBookmark{bookmark})
	incomingRevset := buildTrackedIncomingRevset([]trackedBookmark{bookmark})
	upstreamBookmark := trackedBookmark{Name: "main", Remote: "upstream"}
	upstreamOutgoingRevset := buildTrackedOutgoingRevset([]trackedBookmark{upstreamBookmark})
	upstreamIncomingRevset := buildTrackedIncomingRevset([]trackedBookmark{upstreamBookmark})

	t.Run("no bookmark names", func(t *testing.T) {
		got, err := getBookmarkRemoteStatuses("unused", repoPath, nil)
		if err != nil || len(got) != 0 {
			t.Fatalf("getBookmarkRemoteStatuses() = %v, %v, want empty, nil", got, err)
		}
	})

	t.Run("tracked bookmark command fails", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			trackedBookmarksCommandKey(): {Stderr: "bookmark failed", ExitCode: 1},
		})
		if _, err := getBookmarkRemoteStatuses(binary, repoPath, []string{"main"}); err == nil {
			t.Fatal("getBookmarkRemoteStatuses() error = nil, want error")
		}
	})

	t.Run("normalizes names and returns every tracked remote", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			trackedBookmarksCommandKey():                {Stdout: "dev|origin\nmain|origin\nmain|upstream\n"},
			commitLogCommandKey(outgoingRevset):         {Stdout: "abc123|outgoing\n"},
			commitLogCommandKey(incomingRevset):         {Stdout: "def456|incoming\n"},
			commitLogCommandKey(upstreamOutgoingRevset): {Stdout: "ghi789|upstream outgoing\n"},
			commitLogCommandKey(upstreamIncomingRevset): {Stdout: "jkl012|upstream incoming\n"},
		})
		got, err := getBookmarkRemoteStatuses(binary, repoPath, []string{" ", "main*?", "main"})
		if err != nil {
			t.Fatalf("getBookmarkRemoteStatuses() error = %v", err)
		}
		want := []bookmarkRemoteStatus{{
			Remote:          "origin",
			OutgoingCommits: []string{"abc123 outgoing"},
			IncomingCommits: []string{"def456 incoming"},
		}, {
			Remote:          "upstream",
			OutgoingCommits: []string{"ghi789 upstream outgoing"},
			IncomingCommits: []string{"jkl012 upstream incoming"},
		}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("getBookmarkRemoteStatuses() = %#v, want %#v", got, want)
		}
	})

	t.Run("returns status for an untracked remote", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			trackedBookmarksCommandKey():                {},
			untrackedRemotesCommandKey("main"):          {Stdout: "main|upstream\n"},
			commitLogCommandKey(upstreamOutgoingRevset): {Stdout: "ghi789|outgoing\n"},
			commitLogCommandKey(upstreamIncomingRevset): {Stdout: "jkl012|incoming\n"},
		})
		got, err := getBookmarkRemoteStatuses(binary, repoPath, []string{"main"})
		if err != nil {
			t.Fatalf("getBookmarkRemoteStatuses() error = %v", err)
		}
		want := []bookmarkRemoteStatus{{
			Remote:          "upstream",
			OutgoingCommits: []string{"ghi789 outgoing"},
			IncomingCommits: []string{"jkl012 incoming"},
		}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("getBookmarkRemoteStatuses() = %#v, want %#v", got, want)
		}
	})

	t.Run("untracked remote command fails", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			trackedBookmarksCommandKey():       {},
			untrackedRemotesCommandKey("main"): {Stderr: "list failed", ExitCode: 1},
		})
		if _, err := getBookmarkRemoteStatuses(binary, repoPath, []string{"main"}); err == nil {
			t.Fatal("getBookmarkRemoteStatuses() error = nil, want error")
		}
	})

	t.Run("bookmark has no remotes", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			trackedBookmarksCommandKey():       {},
			untrackedRemotesCommandKey("main"): {},
		})
		got, err := getBookmarkRemoteStatuses(binary, repoPath, []string{"main"})
		if err != nil || len(got) != 0 {
			t.Fatalf("getBookmarkRemoteStatuses() = %v, %v, want empty, nil", got, err)
		}
	})

	t.Run("outgoing command fails", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			trackedBookmarksCommandKey():        {Stdout: "main|origin\n"},
			commitLogCommandKey(outgoingRevset): {Stderr: "log failed", ExitCode: 1},
		})
		if _, err := getBookmarkRemoteStatuses(binary, repoPath, []string{"main"}); err == nil {
			t.Fatal("getBookmarkRemoteStatuses() error = nil, want error")
		}
	})

	t.Run("incoming command fails", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			trackedBookmarksCommandKey():        {Stdout: "main|origin\n"},
			commitLogCommandKey(outgoingRevset): {},
			commitLogCommandKey(incomingRevset): {Stderr: "log failed", ExitCode: 1},
		})
		if _, err := getBookmarkRemoteStatuses(binary, repoPath, []string{"main"}); err == nil {
			t.Fatal("getBookmarkRemoteStatuses() error = nil, want error")
		}
	})
}

// TestGetUntrackedRemotesForBookmark covers command failures and filtering of
// malformed, mismatched, empty, and synthetic git remote entries.
func TestGetUntrackedRemotesForBookmark(t *testing.T) {
	t.Run("command fails", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			untrackedRemotesCommandKey("main"): {Stderr: "list failed", ExitCode: 1},
		})
		if _, err := getUntrackedRemotesForBookmark(binary, t.TempDir(), "main"); err == nil {
			t.Fatal("getUntrackedRemotesForBookmark() error = nil, want error")
		}
	})

	t.Run("filters malformed and ineligible entries", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			untrackedRemotesCommandKey("main"): {
				Stdout: "\ninvalid\ndev|origin\nmain|\nmain|git\nmain|origin\nmain|upstream\n",
			},
		})
		got, err := getUntrackedRemotesForBookmark(binary, t.TempDir(), "main")
		if err != nil {
			t.Fatalf("getUntrackedRemotesForBookmark() error = %v", err)
		}
		want := []string{"origin", "upstream"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("getUntrackedRemotesForBookmark() = %v, want %v", got, want)
		}
	})
}

// TestGetCommitsForRevset verifies the empty-revset shortcut and parsing of
// commit summaries returned from a non-empty revset.
func TestGetCommitsForRevset(t *testing.T) {
	if got, err := getCommitsForRevset("unused", t.TempDir(), "  "); err != nil || len(got) != 0 {
		t.Fatalf("getCommitsForRevset(empty) = %v, %v, want empty, nil", got, err)
	}

	binary := useFakeJJ(t, map[string]fakeJJResponse{
		commitLogCommandKey("bookmarks(main)"): {Stdout: "abc123|change\n"},
	})
	got, err := getCommitsForRevset(binary, t.TempDir(), "bookmarks(main)")
	if err != nil || !reflect.DeepEqual(got, []string{"abc123 change"}) {
		t.Fatalf("getCommitsForRevset() = %v, %v", got, err)
	}
}

// TestGetTrackedBookmarks covers command failures and parsing, validation, and
// deduplication of tracked bookmark and remote pairs.
func TestGetTrackedBookmarks(t *testing.T) {
	t.Run("command fails", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			trackedBookmarksCommandKey(): {Stderr: "list failed", ExitCode: 1},
		})
		if _, err := getTrackedBookmarks(binary, t.TempDir()); err == nil {
			t.Fatal("getTrackedBookmarks() error = nil, want error")
		}
	})

	t.Run("parses unique complete bookmarks", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			trackedBookmarksCommandKey(): {
				Stdout: "\ninvalid\n|origin\nmain|\nmain|origin\nmain|origin\ndev|upstream\n",
			},
		})
		got, err := getTrackedBookmarks(binary, t.TempDir())
		if err != nil {
			t.Fatalf("getTrackedBookmarks() error = %v", err)
		}
		want := []trackedBookmark{
			{Name: "main", Remote: "origin"},
			{Name: "dev", Remote: "upstream"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("getTrackedBookmarks() = %#v, want %#v", got, want)
		}
	})
}

// TestParseCommitSummaries verifies blank and invalid lines are ignored,
// duplicate IDs are removed, and missing descriptions receive a stable label.
func TestParseCommitSummaries(t *testing.T) {
	input := "\n|missing id\nabc123|first change\nabc123|duplicate\ndef456|\nghi789\n"
	want := []string{
		"abc123 first change",
		"def456 (no description set)",
		"ghi789 (no description set)",
	}
	if got := parseCommitSummaries(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("parseCommitSummaries() = %v, want %v", got, want)
	}
}

// TestBuildTrackedRevsets verifies outgoing and incoming expressions combine
// multiple bookmarks while escaping quotes and backslashes safely.
func TestBuildTrackedRevsets(t *testing.T) {
	bookmarks := []trackedBookmark{
		{Name: `ma"in`, Remote: `up\stream`},
		{Name: "dev", Remote: "origin"},
	}

	wantOutgoing := `(remote_bookmarks("ma\"in", remote="up\\stream")..bookmarks("ma\"in")) | ` +
		`(remote_bookmarks("dev", remote="origin")..bookmarks("dev"))`
	if got := buildTrackedOutgoingRevset(bookmarks); got != wantOutgoing {
		t.Fatalf("buildTrackedOutgoingRevset() = %q, want %q", got, wantOutgoing)
	}

	wantIncoming := `(remote_bookmarks("ma\"in", remote="git")..remote_bookmarks("ma\"in", remote="up\\stream")) | ` +
		`(remote_bookmarks("dev", remote="git")..remote_bookmarks("dev", remote="origin"))`
	if got := buildTrackedIncomingRevset(bookmarks); got != wantIncoming {
		t.Fatalf("buildTrackedIncomingRevset() = %q, want %q", got, wantIncoming)
	}
}
