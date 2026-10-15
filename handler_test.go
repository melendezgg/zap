package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlerRejectsUnsupportedMethods(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/index.html": `<h1>Home</h1>`,
	})
	routeStore.SetAll(scanAllRoutes())

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("expected Allow header, got %q", rec.Header().Get("Allow"))
	}
}

func TestHandlerSupportsHeadWithoutBody(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/index.html": `<h1>Home</h1>`,
	})
	routeStore.SetAll(scanAllRoutes())

	req := httptest.NewRequest(http.MethodHead, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty HEAD body, got %q", rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected html content type, got %q", contentType)
	}
}

func TestHandlerServesPublicFiles(t *testing.T) {
	withTempProject(t, map[string]string{
		"public/app.js": `console.log("zap");`,
	})
	routeStore.SetAll(map[string]*RouteInfo{})

	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, `console.log("zap")`) {
		t.Fatalf("unexpected body: %q", body)
	}
	if rec.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("expected no-cache header, got %q", rec.Header().Get("Cache-Control"))
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("expected nosniff header, got %q", rec.Header().Get("X-Content-Type-Options"))
	}
}

func TestHandlerBlocksSensitivePublicFiles(t *testing.T) {
	withTempProject(t, map[string]string{
		"public/.env":         "TOKEN=secret",
		"public/secrets.json": `{"token":"secret"}`,
		"public/private.key":  "secret",
	})
	routeStore.SetAll(map[string]*RouteInfo{})

	for _, path := range []string{"/.env", "/secrets.json", "/private.key"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for %s, got %d", path, rec.Code)
		}
	}
}

func TestHandlerAllowsPublicRuntimeConfig(t *testing.T) {
	withTempProject(t, map[string]string{
		"public/config.json": `{"apiURL":"https://example.test"}`,
	})
	routeStore.SetAll(map[string]*RouteInfo{})

	req := httptest.NewRequest(http.MethodGet, "/config.json", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected public runtime config to be served, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "apiURL") {
		t.Fatalf("unexpected config response: %s", rec.Body.String())
	}
}

func TestHandlerDoesNotServePublicDirectories(t *testing.T) {
	withTempProject(t, map[string]string{
		"public/assets/logo.txt": "logo",
	})
	routeStore.SetAll(map[string]*RouteInfo{})

	req := httptest.NewRequest(http.MethodGet, "/assets", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestResolvePublicPathRejectsTraversal(t *testing.T) {
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
		t.Fatalf("chdir: %v", err)
	}
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		t.Fatalf("mkdir public: %v", err)
	}

	path, ok := resolvePublicPath("/../secret.txt")
	if !ok {
		t.Fatal("expected cleaned path inside public")
	}

	publicRoot, err := filepath.Abs(publicDir)
	if err != nil {
		t.Fatalf("public abs: %v", err)
	}
	if !strings.HasPrefix(path, publicRoot) {
		t.Fatalf("expected path inside public, got %s", path)
	}
}

func TestResolvePublicPathRejectsSymlinkOutsidePublic(t *testing.T) {
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
		t.Fatalf("chdir: %v", err)
	}
	if err := os.MkdirAll(publicDir, 0755); err != nil {
		t.Fatalf("mkdir public: %v", err)
	}
	secretPath := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(secretPath, []byte("secret"), 0644); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	if err := os.Symlink(secretPath, filepath.Join(publicDir, "leak.txt")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if path, ok := resolvePublicPath("/leak.txt"); ok {
		t.Fatalf("expected external symlink to be rejected, resolved to %s", path)
	}

	routeStore.SetAll(map[string]*RouteInfo{})
	req := httptest.NewRequest(http.MethodGet, "/leak.txt", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected external symlink to return 404, got %d", rec.Code)
	}
}

func TestResolvePublicPathAllowsDotDotPrefixedFileName(t *testing.T) {
	withTempProject(t, map[string]string{
		"public/..safe.txt": "safe",
	})

	path, ok := resolvePublicPath("/..safe.txt")
	if !ok || filepath.Base(path) != "..safe.txt" {
		t.Fatalf("expected safe dot-prefixed file, got path=%q ok=%v", path, ok)
	}
}

func TestDevEventsRejectsHead(t *testing.T) {
	req := httptest.NewRequest(http.MethodHead, "/__zap/events", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != "GET" {
		t.Fatalf("expected Allow GET, got %q", rec.Header().Get("Allow"))
	}
}

func TestDevUpdateReturnsReactBundleAndCSS(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/index.tsx":         `export default function App() { return <main>Updated</main>; }`,
		"public/styles/global.css": `body { color: rebeccapurple; }`,
	})
	bundleCache.Clear()
	routeStore.SetAll(scanAllRoutes())

	req := httptest.NewRequest(http.MethodGet, "/__zap/update?route=/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":`) || !strings.Contains(rec.Body.String(), "Updated") {
		t.Fatalf("expected updated bundle: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "rebeccapurple") {
		t.Fatalf("expected updated CSS: %s", rec.Body.String())
	}
}

func TestDevUpdateReturnsSafeCompileError(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/index.tsx": `export default function App() { return <main>; }`,
	})
	bundleCache.Clear()
	routeStore.SetAll(scanAllRoutes())

	req := httptest.NewRequest(http.MethodGet, "/__zap/update?route=/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"error":`) || !strings.Contains(rec.Body.String(), "index.tsx") {
		t.Fatalf("expected safe compilation diagnostic: %s", rec.Body.String())
	}
}
