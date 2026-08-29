package utils

import "testing"

func TestHash(t *testing.T) {
	// FNV-1a 64-bit known vectors.
	cases := []struct {
		in   string
		want string
	}{
		{in: "", want: "cbf29ce484222325"},
		{in: "hello", want: "a430d84680aabd0b"},
	}

	for _, tc := range cases {
		if got := Hash(tc.in); got != tc.want {
			t.Fatalf("Hash(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestHashDeterministic(t *testing.T) {
	s := "/home/user/projects/foo"
	if Hash(s) != Hash(s) {
		t.Fatal("expected identical input to produce identical hash")
	}
}

func TestHashDistinctInputs(t *testing.T) {
	if Hash("/a/b") == Hash("/a/c") {
		t.Fatal("expected distinct inputs to produce distinct hashes")
	}
}
