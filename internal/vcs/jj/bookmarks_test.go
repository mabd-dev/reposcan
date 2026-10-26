package jj

import (
	"reflect"
	"testing"
)

// TestGetBookmarkRemoteStatuses verifies bookmark cleanup, deduplication,
// per-remote lookup regardless of tracking, remotes without the bookmark, and
// failure paths.
func TestGetBookmarkRemoteStatuses(t *testing.T) {
	repoPath := t.TempDir()
	remotes := []string{"origin", "upstream"}
	bookmark := trackedBookmark{Name: "main", Remote: "origin"}
	outgoingRevset := buildTrackedOutgoingRevset([]trackedBookmark{bookmark})
	incomingRevset := buildTrackedIncomingRevset([]trackedBookmark{bookmark})
	upstreamBookmark := trackedBookmark{Name: "main", Remote: "upstream"}
	upstreamOutgoingRevset := buildTrackedOutgoingRevset([]trackedBookmark{upstreamBookmark})
	upstreamIncomingRevset := buildTrackedIncomingRevset([]trackedBookmark{upstreamBookmark})

	t.Run("no bookmark names marks every remote without a lookup", func(t *testing.T) {
		got, err := getBookmarkRemoteStatuses("unused", repoPath, nil, remotes)
		if err != nil {
			t.Fatalf("getBookmarkRemoteStatuses() error = %v", err)
		}
		want := []bookmarkRemoteStatus{
			{Remote: "origin", NoRemoteBookmark: true},
			{Remote: "upstream", NoRemoteBookmark: true},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("getBookmarkRemoteStatuses() = %#v, want %#v", got, want)
		}
	})

	t.Run("no remotes", func(t *testing.T) {
		got, err := getBookmarkRemoteStatuses("unused", repoPath, []string{"main"}, nil)
		if err != nil || len(got) != 0 {
			t.Fatalf("getBookmarkRemoteStatuses() = %v, %v, want empty, nil", got, err)
		}
	})

	t.Run("remote bookmark command fails", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			remoteBookmarksCommandKey(): {Stderr: "list failed", ExitCode: 1},
		})
		if _, err := getBookmarkRemoteStatuses(binary, repoPath, []string{"main"}, remotes); err == nil {
			t.Fatal("getBookmarkRemoteStatuses() error = nil, want error")
		}
	})

	t.Run("normalizes names and returns every remote with the bookmark", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			remoteBookmarksCommandKey():                 {Stdout: "dev|origin\nmain|origin\nmain|upstream\n"},
			commitLogCommandKey(outgoingRevset):         {Stdout: "abc123|outgoing\n"},
			commitLogCommandKey(incomingRevset):         {Stdout: "def456|incoming\n"},
			commitLogCommandKey(upstreamOutgoingRevset): {Stdout: "ghi789|upstream outgoing\n"},
			commitLogCommandKey(upstreamIncomingRevset): {Stdout: "jkl012|upstream incoming\n"},
		})
		got, err := getBookmarkRemoteStatuses(binary, repoPath, []string{" ", "main*?", "main"}, remotes)
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

	t.Run("marks a remote without the bookmark in remote order", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			remoteBookmarksCommandKey():                 {Stdout: "main|upstream\n"},
			commitLogCommandKey(upstreamOutgoingRevset): {},
			commitLogCommandKey(upstreamIncomingRevset): {},
		})
		got, err := getBookmarkRemoteStatuses(binary, repoPath, []string{"main"}, remotes)
		if err != nil {
			t.Fatalf("getBookmarkRemoteStatuses() error = %v", err)
		}
		want := []bookmarkRemoteStatus{{
			Remote:           "origin",
			NoRemoteBookmark: true,
		}, {
			Remote:          "upstream",
			OutgoingCommits: []string{},
			IncomingCommits: []string{},
		}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("getBookmarkRemoteStatuses() = %#v, want %#v", got, want)
		}
	})

	t.Run("returns statuses for several bookmarks on different remotes", func(t *testing.T) {
		devBookmark := trackedBookmark{Name: "dev", Remote: "origin"}
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			remoteBookmarksCommandKey():                                                     {Stdout: "dev|origin\nmain|upstream\n"},
			commitLogCommandKey(upstreamOutgoingRevset):                                     {Stdout: "ghi789|main outgoing\n"},
			commitLogCommandKey(upstreamIncomingRevset):                                     {},
			commitLogCommandKey(buildTrackedOutgoingRevset([]trackedBookmark{devBookmark})): {Stdout: "abc123|dev outgoing\n"},
			commitLogCommandKey(buildTrackedIncomingRevset([]trackedBookmark{devBookmark})): {},
		})
		got, err := getBookmarkRemoteStatuses(binary, repoPath, []string{"main", "dev"}, remotes)
		if err != nil {
			t.Fatalf("getBookmarkRemoteStatuses() error = %v", err)
		}
		want := []bookmarkRemoteStatus{{
			Remote:          "origin",
			OutgoingCommits: []string{"abc123 dev outgoing"},
			IncomingCommits: []string{},
		}, {
			Remote:          "upstream",
			OutgoingCommits: []string{"ghi789 main outgoing"},
			IncomingCommits: []string{},
		}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("getBookmarkRemoteStatuses() = %#v, want %#v", got, want)
		}
	})

	t.Run("outgoing command fails", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			remoteBookmarksCommandKey():         {Stdout: "main|origin\n"},
			commitLogCommandKey(outgoingRevset): {Stderr: "log failed", ExitCode: 1},
		})
		if _, err := getBookmarkRemoteStatuses(binary, repoPath, []string{"main"}, remotes); err == nil {
			t.Fatal("getBookmarkRemoteStatuses() error = nil, want error")
		}
	})

	t.Run("incoming command fails", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			remoteBookmarksCommandKey():         {Stdout: "main|origin\n"},
			commitLogCommandKey(outgoingRevset): {},
			commitLogCommandKey(incomingRevset): {Stderr: "log failed", ExitCode: 1},
		})
		if _, err := getBookmarkRemoteStatuses(binary, repoPath, []string{"main"}, remotes); err == nil {
			t.Fatal("getBookmarkRemoteStatuses() error = nil, want error")
		}
	})
}

// TestGetRemoteBookmarks covers command failures and filtering of malformed,
// empty, and synthetic git remote entries.
func TestGetRemoteBookmarks(t *testing.T) {
	t.Run("command fails", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			remoteBookmarksCommandKey(): {Stderr: "list failed", ExitCode: 1},
		})
		if _, err := getRemoteBookmarks(binary, t.TempDir()); err == nil {
			t.Fatal("getRemoteBookmarks() error = nil, want error")
		}
	})

	t.Run("filters malformed and ineligible entries", func(t *testing.T) {
		binary := useFakeJJ(t, map[string]fakeJJResponse{
			remoteBookmarksCommandKey(): {
				Stdout: "\ninvalid\ndev|origin\nmain|\nmain|git\n|origin\nmain|origin\nmain|upstream\n",
			},
		})
		got, err := getRemoteBookmarks(binary, t.TempDir())
		if err != nil {
			t.Fatalf("getRemoteBookmarks() error = %v", err)
		}
		want := []trackedBookmark{
			{Name: "dev", Remote: "origin"},
			{Name: "main", Remote: "origin"},
			{Name: "main", Remote: "upstream"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("getRemoteBookmarks() = %v, want %v", got, want)
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

	wantOutgoing := `(remote_bookmarks(exact:"ma\"in", remote=exact:"up\\stream")..@-) | ` +
		`(remote_bookmarks(exact:"dev", remote=exact:"origin")..@-)`
	if got := buildTrackedOutgoingRevset(bookmarks); got != wantOutgoing {
		t.Fatalf("buildTrackedOutgoingRevset() = %q, want %q", got, wantOutgoing)
	}

	wantIncoming := `(@..remote_bookmarks(exact:"ma\"in", remote=exact:"up\\stream")) | ` +
		`(@..remote_bookmarks(exact:"dev", remote=exact:"origin"))`
	if got := buildTrackedIncomingRevset(bookmarks); got != wantIncoming {
		t.Fatalf("buildTrackedIncomingRevset() = %q, want %q", got, wantIncoming)
	}
}
