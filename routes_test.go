package main

import (
	"os"
	"path/filepath"
	"strings"
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

func TestScanAllRoutesExcludesPrivateDirectories(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/index.tsx":                     `export default function App() { return <main>Home</main>; }`,
		"routes/_components/Card.tsx":          `export default function Card() { return <p>Private</p>; }`,
		"routes/_utils/helper.tsx":             `export default function Helper() { return null; }`,
		"routes/blog/index.tsx":                `export default function Blog() { return <main>Blog</main>; }`,
		"routes/blog/post.tsx":                 `export default function Post() { return <main>Post</main>; }`,
		"routes/blog/_components/PostCard.tsx": `export default function PostCard() { return null; }`,
	})

	routes := scanAllRoutes()
	for _, route := range []string{"/", "/blog", "/blog/post"} {
		if _, ok := routes[route]; !ok {
			t.Fatalf("expected public route %s in %#v", route, routes)
		}
	}
	for _, route := range []string{"/_components/Card", "/_utils/helper", "/blog/_components/PostCard"} {
		if _, ok := routes[route]; ok {
			t.Fatalf("did not expect private route %s in %#v", route, routes)
		}
	}
}

func TestBundleJSXImportsFromPrivateRouteDirectory(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/index.tsx": `
import Card from "./_components/Card";
export default function App() { return <main><Card /></main>; }
`,
		"routes/_components/Card.tsx": `export default function Card() { return <p>Private Card</p>; }`,
	})
	bundleCache.Clear()
	bundleBuilds.dispose()
	t.Cleanup(bundleBuilds.dispose)

	bundle, err := bundleJSX(filepath.Join(routesDir, "index.tsx"), true)
	if err != nil {
		t.Fatalf("bundle public route with private import: %v", err)
	}
	if !strings.Contains(bundle, "Private Card") {
		t.Fatalf("expected private component in bundle")
	}
}

func TestReloadProjectKeepsNewPrivateDirectoriesOutOfRouteStore(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/index.tsx": `export default function App() { return <main>Home</main>; }`,
	})
	routeStore.SetAll(scanAllRoutes())
	writeTestFile(t, filepath.Join(routesDir, "_components", "Card.tsx"), `export default function Card() { return null; }`)

	reloadProject()

	if _, ok := routeStore.Get("/_components/Card"); ok {
		t.Fatal("did not expect a private route after project reload")
	}
}

func TestRoutePathForFileDocumentsSupportedConventions(t *testing.T) {
	tests := []struct {
		file string
		want string
	}{
		{"index.tsx", "/"},
		{"about.tsx", "/about"},
		{"about.jsx", "/about"},
		{"about.html", "/about"},
		{filepath.Join("blog", "index.tsx"), "/blog"},
		{filepath.Join("blog", "post.tsx"), "/blog/post"},
		{filepath.Join("scripts", "app.js"), "/scripts/app.js"},
		{filepath.Join("scripts", "index.js"), "/scripts/index.js"},
	}

	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
			route, ok := routePathForFile(routesDir, filepath.Join(routesDir, test.file))
			if !ok {
				t.Fatal("expected route path")
			}
			if route != test.want {
				t.Fatalf("expected %s, got %s", test.want, route)
			}
		})
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
