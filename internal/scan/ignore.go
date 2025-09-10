package scan

import (
	"path/filepath"
	"strings"

	doublestar "github.com/bmatcuk/doublestar/v4"
)

// IgnoreMatcher holds compiled patterns and the normalized scan roots.
type IgnoreMatcher struct {
	patterns []string // normalized (forward slashes)
	roots    []string // normalized (forward slashes), expanded from user roots
}

// NewIgnoreMatcher normalizes and post-processes patterns for friendlier DX.
func NewIgnoreMatcher(roots []string, raw []string) *IgnoreMatcher {
	normRoots := make([]string, 0, len(roots))
	for _, r := range roots {
		normRoots = append(normRoots, filepath.ToSlash(r))
	}

	var pats []string
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		p = filepath.ToSlash(p)

		// If no slash and no wildcard, treat as directory token anywhere.
		if !strings.Contains(p, "/") && !strings.ContainsAny(p, "*?[") {
			p = "**/" + p + "/**"
		}

		// If it looks like a plain directory name with a trailing slash, ensure recursive.
		if strings.HasSuffix(p, "/") && !strings.HasSuffix(p, "/**") {
			p = p + "**"
		}

		// Special case: patterns like "cache/" should also match the directory itself.
		// doublestar PathMatch("**/cache/**", "/proj/tmp/cache") returns false, so we
		// add a sibling pattern without the trailing "/**" to match the folder path.
		if strings.HasSuffix(p, "/**") {
			base := strings.TrimSuffix(p, "/**")
			pats = append(pats, base)
		}

		pats = append(pats, p)
	}

	return &IgnoreMatcher{
		patterns: pats,
		roots:    normRoots,
	}
}

// ShouldIgnore reports whether absPath should be ignored.
// absPath must be absolute or at least rooted; it is normalized to forward slashes.
func (m *IgnoreMatcher) ShouldIgnore(absPath string) bool {
	if len(m.patterns) == 0 {
		return false
	}
	path := filepath.ToSlash(absPath)

	for _, pat := range m.patterns {
		if strings.HasPrefix(pat, "/") {
			// Anchor to each root: join root + pat
			anch := pat // e.g., "/vendor/**"
			for _, root := range m.roots {
				// Build a pattern like "<root>/vendor/**" (avoid double slashes)
				rootPat := root + anch
				if ok, _ := doublestar.PathMatch(rootPat, path); ok {
					return true
				}
			}
		} else if filepath.IsAbs(pat) {
			// Absolute OS path pattern (already normalized to /)
			if ok, _ := doublestar.PathMatch(pat, path); ok {
				return true
			}
		} else {
			// Anywhere pattern, e.g., "**/node_modules/**"
			if ok, _ := doublestar.PathMatch(pat, path); ok {
				return true
			}
		}
	}
	return false
}
