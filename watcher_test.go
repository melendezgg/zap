package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchedFilesChangedDetectsSizeChanges(t *testing.T) {
	now := time.Now()
	previous := map[string]watchedFile{
		"routes/index.tsx": {modTime: now, size: 10},
	}
	current := map[string]watchedFile{
		"routes/index.tsx": {modTime: now, size: 11},
	}

	if !watchedFilesChanged(previous, current) {
		t.Fatal("expected size change to be detected")
	}
}

func TestWatchedFilesChangedDetectsDeletedFiles(t *testing.T) {
	previous := map[string]watchedFile{
		"routes/index.tsx": {modTime: time.Now(), size: 10},
	}
	current := map[string]watchedFile{}

	if !watchedFilesChanged(previous, current) {
		t.Fatal("expected deleted file to be detected")
	}
}

func TestCollectWatchedFilesFromIgnoresPrivateSystemFiles(t *testing.T) {
	dir := t.TempDir()
	routes := filepath.Join(dir, "routes")
	if err := os.MkdirAll(routes, 0755); err != nil {
		t.Fatalf("mkdir routes: %v", err)
	}

	writeTestFile(t, filepath.Join(routes, "index.tsx"), "export default function App() { return null; }")
	writeTestFile(t, filepath.Join(routes, ".DS_Store"), "ignore")
	writeTestFile(t, filepath.Join(routes, "draft.tsx~"), "ignore")

	files, err := collectWatchedFilesFrom([]string{routes})
	if err != nil {
		t.Fatalf("collect watched files: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected one watched file, got %#v", files)
	}
	if _, ok := files[filepath.ToSlash(filepath.Join(routes, "index.tsx"))]; !ok {
		t.Fatalf("expected index.tsx in watched files: %#v", files)
	}
}

func TestCollectWatchedFilesFromAllowsMissingOptionalRoot(t *testing.T) {
	files, err := collectWatchedFilesFrom([]string{filepath.Join(t.TempDir(), "missing")})
	if err != nil {
		t.Fatalf("expected missing root to be ignored, got %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected no files, got %#v", files)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
