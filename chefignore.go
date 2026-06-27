package cinc

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Chefignore holds the glob patterns parsed from a cookbook's `chefignore`
// file and reports whether a cookbook-relative path should be excluded from
// upload or packaging. It is the single, canonical implementation of knife's
// chefignore handling: cookbook uploads, cookbook archives, and Policyfile
// content-identifier computation all match files through one Chefignore so
// they agree on exactly which files belong to a cookbook.
type Chefignore struct {
	patterns []string
}

// LoadChefignore reads the `chefignore` file at the root of cookbookDir and
// returns the parsed patterns. A missing file is not an error — the returned
// Chefignore ignores nothing.
func LoadChefignore(cookbookDir string) (*Chefignore, error) {
	data, err := os.ReadFile(filepath.Join(cookbookDir, "chefignore"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Chefignore{}, nil
		}
		return nil, err
	}
	return &Chefignore{patterns: parseChefignore(data)}, nil
}

// parseChefignore extracts the glob patterns from a chefignore file's bytes,
// dropping blank lines and comments (lines whose first non-space character is
// "#"). Surrounding whitespace is trimmed from each pattern.
func parseChefignore(data []byte) []string {
	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

// Patterns returns the chefignore globs, in file order. The result is a copy;
// mutating it does not affect the Chefignore.
func (c *Chefignore) Patterns() []string {
	if c == nil || len(c.patterns) == 0 {
		return nil
	}
	return append([]string(nil), c.patterns...)
}

// Ignores reports whether relPath — a forward-slash, cookbook-relative path —
// is excluded by any chefignore pattern. It matches knife: each pattern is
// tested (with filepath glob semantics) against the full path, the basename,
// and every ancestor directory along with each ancestor's basename. That makes
// a pattern like `spec/*`, or a bare directory name like `.kitchen`, exclude
// everything nested beneath it.
//
// A nil receiver, no patterns, or an empty/"." path all ignore nothing.
func (c *Chefignore) Ignores(relPath string) bool {
	if c == nil || len(c.patterns) == 0 || relPath == "" || relPath == "." {
		return false
	}
	candidates := chefignoreCandidates(relPath)
	for _, pattern := range c.patterns {
		for _, cand := range candidates {
			if ok, _ := path.Match(pattern, cand); ok {
				return true
			}
		}
	}
	return false
}

// chefignoreCandidates returns the strings a pattern is tested against for one
// path: the full path, its basename, and every ancestor directory plus that
// ancestor's basename.
func chefignoreCandidates(relPath string) []string {
	candidates := []string{relPath, path.Base(relPath)}
	dir := path.Dir(relPath)
	for dir != "." && dir != "/" && dir != "" {
		candidates = append(candidates, dir, path.Base(dir))
		dir = path.Dir(dir)
	}
	return candidates
}
