package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateHTMLUsesEmbeddedReactRuntime(t *testing.T) {
	html := generateHTML("console.log('app');", "", "Test")

	if strings.Contains(html, "unpkg.com") {
		t.Fatalf("expected no CDN references: %s", html)
	}
	if !strings.Contains(html, "/__zap/vendor/react.development.mjs") {
		t.Fatalf("expected embedded React import: %s", html)
	}
	if !strings.Contains(html, "/__zap/vendor/react-dom-client.development.mjs") {
		t.Fatalf("expected embedded ReactDOM import: %s", html)
	}
	if !strings.Contains(html, "/__zap/vendor/react-dom.development.mjs") {
		t.Fatalf("expected embedded ReactDOM import: %s", html)
	}
	if !strings.Contains(html, `type="module"`) {
		t.Fatalf("expected module runtime script: %s", html)
	}
	if strings.Contains(html, "ReactDOM.createRoot") {
		t.Fatalf("expected render to use imported createRoot binding: %s", html)
	}
	if !strings.Contains(html, "const root = createRoot(") {
		t.Fatalf("expected render to use createRoot directly: %s", html)
	}
}

func TestHandlerServesVendorAssets(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/__zap/vendor/react.development.mjs", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "application/javascript") {
		t.Fatalf("expected JavaScript content type, got %q", contentType)
	}
	if !strings.Contains(rec.Body.String(), "react@19.2.6") {
		t.Fatalf("expected vendored React 19.2.6 asset")
	}
}

func TestHandlerServesReactDOMVendorAsset(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/__zap/vendor/react-dom.development.mjs", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "react-dom@19.2.6") {
		t.Fatalf("expected vendored ReactDOM 19.2.6 asset")
	}
	if strings.Contains(rec.Body.String(), `"/react@19.2.6/`) {
		t.Fatalf("expected ReactDOM asset to use local imports")
	}
}

func TestHandlerRejectsUnsupportedVendorMethods(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/__zap/vendor/react.development.mjs", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("expected Allow header, got %q", rec.Header().Get("Allow"))
	}
}
