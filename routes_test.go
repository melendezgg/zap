package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanAllRoutesDiscoversSupportedFiles(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/index.tsx":        `export const title = "Home"; export default function App() { return <h1>Home</h1>; }`,
		"routes/about.html":       `<title>About</title><h1>About</h1>`,
		"routes/scripts/tool.js":  `console.log("tool");`,
		"routes/_Private.tsx":     `export default function Private() { return null; }`,
		"routes/notes/readme.txt": `ignore me`,
	})

	routes := scanAllRoutes()

	if _, ok := routes["/"]; !ok {
		t.Fatal("expected / route")
	}
	if route := routes["/about"]; route == nil || route.Type != "html" || route.Title != "About" {
		t.Fatalf("unexpected /about route: %#v", route)
	}
	if route := routes["/scripts/tool.js"]; route == nil || route.Type != "js" {
		t.Fatalf("unexpected /scripts/tool.js route: %#v", route)
	}
	if _, ok := routes["/_Private"]; ok {
		t.Fatal("did not expect private route")
	}
}

func TestScanAllRoutesUsesExplicitConflictPriority(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/about.html": `<title>HTML About</title>`,
		"routes/about.jsx":  `export const title = "JSX About"; export default function App() { return <h1>About</h1>; }`,
		"routes/about.tsx":  `export const title = "TSX About"; export default function App() { return <h1>About</h1>; }`,
	})

	routes := scanAllRoutes()

	route := routes["/about"]
	if route == nil {
		t.Fatal("expected /about route")
	}
	if filepath.Base(route.File) != "about.tsx" {
		t.Fatalf("expected about.tsx to win, got %s", route.File)
	}
	if route.Title != "TSX About" {
		t.Fatalf("expected title from winning route, got %q", route.Title)
	}
}

func TestRoutePathForFileKeepsJSFileExtension(t *testing.T) {
	route, ok := routePathForFile("routes", filepath.Join("routes", "script.js"))
	if !ok {
		t.Fatal("expected route path")
	}
	if route != "/script.js" {
		t.Fatalf("expected /script.js, got %s", route)
	}
}

func TestChooseRouteFilePriority(t *testing.T) {
	winner, loser := chooseRouteFile("routes/page.html", "routes/page.jsx")
	if winner != "routes/page.jsx" || loser != "routes/page.html" {
		t.Fatalf("unexpected winner=%s loser=%s", winner, loser)
	}
}

func withTempProject(t *testing.T, files map[string]string) {
	t.Helper()

	dir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir temp project: %v", err)
	}

	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
}
