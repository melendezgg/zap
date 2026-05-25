package main

import (
	"embed"
	"net/http"
	"strings"
)

const reactVersion = "19.2.6"

//go:embed internal/assets/react/*.mjs
var vendorAssets embed.FS

func handleVendorAsset(w http.ResponseWriter, r *http.Request) (bool, int) {
	if !strings.HasPrefix(r.URL.Path, "/__zap/assets/") {
		return false, 0
	}

	if !isReadMethod(r.Method) {
		methodNotAllowed(w, "GET, HEAD")
		return true, http.StatusMethodNotAllowed
	}

	name := strings.TrimPrefix(r.URL.Path, "/__zap/assets/")
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
	return `import React from "/__zap/assets/react/react.development.mjs";
import * as ReactDOM from "/__zap/assets/react/react-dom.development.mjs";
import { createRoot, hydrateRoot } from "/__zap/assets/react/react-dom-client.development.mjs";

window.React = React;
window.ReactDOM = { ...ReactDOM, createRoot, hydrateRoot };
const { useState, useEffect, useContext, createContext, useRef, useMemo, useCallback } = React;`
}
