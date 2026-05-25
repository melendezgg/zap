package main

import (
	"errors"
	"strings"
	"testing"
)

func TestInjectDevReloadScriptBeforeBody(t *testing.T) {
	html := "<html><body><h1>Zap</h1></body></html>"

	result := injectDevReloadScript(html)

	if !strings.Contains(result, `new EventSource("/__zap/events")`) {
		t.Fatalf("expected reload script in result: %s", result)
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

func TestRenderDevErrorHTMLIncludesReloadClient(t *testing.T) {
	result := renderDevErrorHTML("/labs", errors.New("compile failed"))

	if !strings.Contains(result, `new EventSource("/__zap/events")`) {
		t.Fatalf("expected reload client in dev error HTML: %s", result)
	}
}
