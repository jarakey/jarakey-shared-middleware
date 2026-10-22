// Package testutils holds small, reusable pieces for the "source-text guard test" pattern used
// across Jarakey services (e.g. ENG-250's banned-call-shape checks, ENG-337's migration
// transaction-control guard): walk the repo's own .go source, skip the files that can't matter,
// and fail when a forbidden pattern appears outside its one sanctioned call site. Before this
// package existed, each service re-implemented the walk-and-skip boilerplate from scratch —
// consolidated here so a future service adopting this pattern has one place to import it from.
package testutils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DefaultSkipDirs are non-Go asset directories no source-scan guard has a reason to descend
// into. Callers can pass additional entries to FindSourceMatches on top of these.
var DefaultSkipDirs = map[string]bool{
	"vendor": true, ".git": true, "helm": true, "infrastructure": true,
	"docs": true, "doc": true, "bin": true, "node_modules": true,
}

// FindSourceMatches walks root for non-test .go files whose contents match pattern, skipping
// DefaultSkipDirs plus any directory named in extraSkipDirs. Returns the matching file paths
// (empty/nil when none), or an error if the walk itself fails (e.g. root does not exist).
func FindSourceMatches(root string, pattern *regexp.Regexp, extraSkipDirs map[string]bool) ([]string, error) {
	var matches []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if DefaultSkipDirs[info.Name()] || extraSkipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if pattern.Match(src) {
			matches = append(matches, path)
		}
		return nil
	})
	return matches, err
}
