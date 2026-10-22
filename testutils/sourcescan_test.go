package testutils

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestFindSourceMatches_FindsAMatchInANonTestGoFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "offender.go", "package x\nfunc f() { bad() }\n")
	writeFile(t, dir, "clean.go", "package x\nfunc g() { ok() }\n")

	matches, err := FindSourceMatches(dir, regexp.MustCompile(`bad\(\)`), nil)
	if err != nil {
		t.Fatalf("FindSourceMatches: %v", err)
	}
	if len(matches) != 1 || filepath.Base(matches[0]) != "offender.go" {
		t.Fatalf("matches = %v, want exactly [offender.go]", matches)
	}
}

// The red-proof for any guard built on this: a match inside a _test.go file must never count,
// or a guard's own red-proof test (which deliberately plants the forbidden shape) would fail
// itself.
func TestFindSourceMatches_IgnoresTestFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "offender_test.go", "package x\nfunc f() { bad() }\n")

	matches, err := FindSourceMatches(dir, regexp.MustCompile(`bad\(\)`), nil)
	if err != nil {
		t.Fatalf("FindSourceMatches: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("matches = %v, want none — _test.go files must be skipped", matches)
	}
}

func TestFindSourceMatches_SkipsDefaultAndExtraDirs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("vendor", "offender.go"), "package x\nfunc f() { bad() }\n")
	writeFile(t, dir, filepath.Join("mycache", "offender.go"), "package x\nfunc f() { bad() }\n")
	writeFile(t, dir, "real_offender.go", "package x\nfunc f() { bad() }\n")

	matches, err := FindSourceMatches(dir, regexp.MustCompile(`bad\(\)`), map[string]bool{"mycache": true})
	if err != nil {
		t.Fatalf("FindSourceMatches: %v", err)
	}
	if len(matches) != 1 || filepath.Base(matches[0]) != "real_offender.go" {
		t.Fatalf("matches = %v, want exactly [real_offender.go] (vendor/ and mycache/ must be skipped)", matches)
	}
}
