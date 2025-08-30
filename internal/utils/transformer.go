package utils

import (
	"fmt"
	"hash/fnv"
)

// StringID returns a stable hash for a given string.
// Same input always produces the same ID.
func Hash(s string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s)) // Write never returns error for fnv
	return fmt.Sprintf("%x", h.Sum64())
}
