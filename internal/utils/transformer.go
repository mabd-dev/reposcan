package utils

import (
	"fmt"
	"hash/fnv"
)

// Hash returns a stable hexadecimal FNV-1a 64-bit hash for s.
// The same input always produces the same identifier string.
func Hash(s string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s)) // Write never returns error for fnv
	return fmt.Sprintf("%x", h.Sum64())
}
