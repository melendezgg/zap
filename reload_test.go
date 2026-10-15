package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

func TestInjectDevReloadScriptBeforeBody(t *testing.T) {
	html := "<html><body><h1>Zap</h1></body></html>"

	result := injectDevReloadScript(html)

	if !strings.Contains(result, `new EventSource("/__zap/events")`) {
		t.Fatalf("expected reload script in result: %s", result)
	}
	if !strings.Contains(result, `fetch("/__zap/update?route="`) {
		t.Fatalf("expected hot update client in result: %s", result)
	}
	if !strings.Contains(result, `__zap-error-overlay`) {
		t.Fatalf("expected compilation error overlay in result: %s", result)
	}
	if !strings.Contains(result, `window.location.origin + "/__zap/assets/react/"`) {
		t.Fatalf("expected blob updates to resolve embedded React assets from the page origin: %s", result)
	}
	if strings.Index(result, devReloadScript) > strings.Index(result, "</body>") {
		t.Fatalf("expected reload script before closing body: %s", result)
	}
}

func TestInjectDevReloadScriptAppendsWhenBodyIsMissing(t *testing.T) {
	html := "<h1>Zap</h1>"

	result := injectDevReloadScript(html)

	if !strings.HasPrefix(result, html) {
		t.Fatalf("expected original content prefix: %s", result)
	}
	if !strings.Contains(result, devReloadScript) {
		t.Fatalf("expected reload script in result: %s", result)
	}
}

func TestReloadHubBroadcastsToSubscribers(t *testing.T) {
	hub := newReloadHub()
	events, unsubscribe := hub.subscribe()
	defer unsubscribe()

	hub.broadcast()

	select {
	case <-events:
	default:
		t.Fatal("expected reload event")
	}
}

func TestReloadHubCloseSignalsSubscribers(t *testing.T) {
	hub := newReloadHub()
	hub.close()

	select {
	case <-hub.done:
	default:
		t.Fatal("expected reload hub shutdown signal")
	}
}

func TestRenderDevErrorHTMLIncludesReloadClient(t *testing.T) {
	result := renderDevErrorHTML("/labs", errors.New("compile failed"))

	if !strings.Contains(result, `new EventSource("/__zap/events")`) {
		t.Fatalf("expected reload client in dev error HTML: %s", result)
	}
}

func TestRenderDevErrorHTMLUsesProjectRelativePaths(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	absPath := filepath.Join(workingDir, "routes", "broken.tsx")
	buildErr := newCompileError("build", []api.Message{{
		Text: "Expected identifier",
		Location: &api.Location{
			File:     absPath,
			Line:     7,
			Column:   3,
			LineText: "bad code",
		},
	}})

	result := renderDevErrorHTML("/broken", buildErr)
	if strings.Contains(result, workingDir) {
		t.Fatalf("did not expect absolute project path in browser error: %s", result)
	}
	if !strings.Contains(result, "routes/broken.tsx:7:4") {
		t.Fatalf("expected relative diagnostic path: %s", result)
	}
}

func TestRenderDevErrorHTMLHidesUnknownInternalErrors(t *testing.T) {
	result := renderDevErrorHTML("/broken", errors.New("open /Users/private/secret: permission denied"))
	if strings.Contains(result, "/Users/private/secret") {
		t.Fatalf("did not expect internal path in browser error: %s", result)
	}
	if !strings.Contains(result, "See the Zap server log") {
		t.Fatalf("expected safe fallback message: %s", result)
	}
}
