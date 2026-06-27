package cinc

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestParseChefignoreSkipsCommentsAndBlanks(t *testing.T) {
	patterns := parseChefignore([]byte("# a comment\n\n*.bak\n  spec/* \n\n# another\nBerksfile.lock\n"))
	want := []string{"*.bak", "spec/*", "Berksfile.lock"}
	if !slices.Equal(patterns, want) {
		t.Fatalf("patterns = %v, want %v", patterns, want)
	}
}

func TestLoadChefignoreMissingFileIgnoresNothing(t *testing.T) {
	ci, err := LoadChefignore(t.TempDir())
	if err != nil {
		t.Fatalf("LoadChefignore: %v", err)
	}
	if ci == nil {
		t.Fatal("LoadChefignore returned nil for missing file, want empty Chefignore")
	}
	if ci.Patterns() != nil {
		t.Fatalf("Patterns() = %v, want nil for missing file", ci.Patterns())
	}
	if ci.Ignores("anything.rb") {
		t.Fatal("a missing chefignore should ignore nothing")
	}
}

func TestLoadChefignoreReadsFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "chefignore"), []byte("*.bak\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ci, err := LoadChefignore(dir)
	if err != nil {
		t.Fatalf("LoadChefignore: %v", err)
	}
	if !slices.Equal(ci.Patterns(), []string{"*.bak"}) {
		t.Fatalf("Patterns() = %v, want [*.bak]", ci.Patterns())
	}
}

func TestChefignoreIgnoresByBasename(t *testing.T) {
	ci := &Chefignore{patterns: []string{"*.bak"}}
	cases := map[string]bool{
		"recipes/default.rb":  false,
		"recipes/default.bak": true,
		"deep/nested/foo.bak": true,
		"foo.bak":             true,
		"foobak":              false,
	}
	for relPath, want := range cases {
		if got := ci.Ignores(relPath); got != want {
			t.Errorf("%q: got %v, want %v", relPath, got, want)
		}
	}
}

func TestChefignoreIgnoresByDirectorySegment(t *testing.T) {
	// A pattern matching a directory (a glob like `spec/*` or a bare directory
	// name like `.kitchen`) excludes everything nested beneath it. This is the
	// behavior cookbook packaging relies on and is preserved from cinc-cli's
	// cookbook chefignore copy.
	ci := &Chefignore{patterns: []string{"spec/*", ".kitchen"}}
	cases := map[string]bool{
		"spec/foo_spec.rb":        true,
		"spec/fixtures/sample.rb": true,
		".kitchen/state.yml":      true,
		".kitchen":                true,
		"recipes/default.rb":      false,
		"libraries/helper.rb":     false,
	}
	for relPath, want := range cases {
		if got := ci.Ignores(relPath); got != want {
			t.Errorf("%q: got %v, want %v", relPath, got, want)
		}
	}
}

func TestChefignoreIgnoresFullPath(t *testing.T) {
	ci := &Chefignore{patterns: []string{"recipes/default.bak"}}
	if !ci.Ignores("recipes/default.bak") {
		t.Fatal("expected full-path match")
	}
	if ci.Ignores("recipes/other.bak") {
		t.Fatal("did not expect match for a different file")
	}
}

func TestChefignoreEmptyIgnoresNothing(t *testing.T) {
	if (&Chefignore{}).Ignores("anything.rb") {
		t.Fatal("an empty Chefignore should not match")
	}
	var nilCI *Chefignore
	if nilCI.Ignores("anything.rb") {
		t.Fatal("a nil Chefignore should not match")
	}
	ci := &Chefignore{patterns: []string{"*.bak"}}
	if ci.Ignores("") || ci.Ignores(".") {
		t.Fatal(`empty path and "." should never be ignored`)
	}
}
