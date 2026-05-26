package main

import "testing"

func TestShouldIgnoreSystemAndTemporaryFiles(t *testing.T) {
	ignored := []string{
		"routes/.DS_Store",
		"routes/Thumbs.db",
		"routes/.gitignore",
		"routes/index.tsx.swp",
		"routes/index.tsx.swo",
		"routes/index.tsx.swn",
		"routes/index.tsx.bak",
		"routes/index.tsx~",
	}

	for _, path := range ignored {
		if !shouldIgnore(path) {
			t.Fatalf("expected %s to be ignored", path)
		}
	}
}

func TestShouldIgnoreKeepsRegularFiles(t *testing.T) {
	kept := []string{
		"routes/index.tsx",
		"routes/.well-known.tsx",
		"public/images/logo.svg",
	}

	for _, path := range kept {
		if shouldIgnore(path) {
			t.Fatalf("expected %s to be kept", path)
		}
	}
}
