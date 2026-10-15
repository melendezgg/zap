package main

import (
	"embed"
	"net/http"
	"strings"
)

const reactVersion = "19.2.6"

const reactRefreshVersion = "0.17.0"

const mountAppScript = `if (typeof window.App !== 'undefined') {
    if (!window.__zapRoot) {
        window.__zapRoot = createRoot(document.getElementById('root'));
        ZapRefreshRuntime.register(window.App, window.__zapRefreshID);
        window.__zapRoot.render(React.createElement(window.App));
    } else {
        ZapRefreshRuntime.register(window.App, window.__zapRefreshID);
        ZapRefreshRuntime.performReactRefresh();
    }
    window.__zapHotUpdate = true;
}`

//go:embed internal/assets/react/*.mjs
var vendorAssets embed.FS

func handleVendorAsset(w http.ResponseWriter, r *http.Request) (bool, int) {
	if !strings.HasPrefix(r.URL.Path, "/__zap/assets/react/") {
		return false, 0
	}

	if !isReadMethod(r.Method) {
		methodNotAllowed(w, "GET, HEAD")
		return true, http.StatusMethodNotAllowed
	}

	name := strings.TrimPrefix(r.URL.Path, "/__zap/assets/react/")
	if name == "" || strings.Contains(name, "/") || !strings.HasSuffix(name, ".mjs") {
		http.NotFound(w, r)
		return true, http.StatusNotFound
	}

	content, err := vendorAssets.ReadFile("internal/assets/react/" + name)
	if err != nil {
		http.NotFound(w, r)
		return true, http.StatusNotFound
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	writeResponse(w, r, http.StatusOK, content)
	return true, http.StatusOK
}

func reactRuntimeImportScript() string {
	return `import ZapRefreshRuntime from "/__zap/assets/react/react-refresh-runtime.development.mjs";

if (!window.__zapRefreshReady) {
  ZapRefreshRuntime.injectIntoGlobalHook(window);
  window.__zapRefreshReady = true;
}

const { default: React } = await import("/__zap/assets/react/react.development.mjs");
const ReactDOM = await import("/__zap/assets/react/react-dom.development.mjs");
const { createRoot, hydrateRoot } = await import("/__zap/assets/react/react-dom-client.development.mjs");

window.React = React;
window.ReactDOM = { ...ReactDOM, createRoot, hydrateRoot };
const { useState, useEffect, useContext, createContext, useRef, useMemo, useCallback } = React;`
}
