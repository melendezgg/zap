package main

import (
	"net/http"
	"os"
	"strings"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	path := r.URL.Path

	if path == "/__zap/events" {
		handleDevEvents(w, r)
		return
	}

	// Favicon
	if path == "/favicon.ico" {
		faviconPath, ok := resolvePublicPath(path)
		if !ok {
			http.Error(w, "404 - Ruta no encontrada", http.StatusNotFound)
			logRequest(r.Method, path, http.StatusNotFound, start)
			return
		}
		if _, err := os.Stat(faviconPath); err == nil {
			http.ServeFile(w, r, faviconPath)
			logRequest(r.Method, path, http.StatusOK, start)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	info, exists := routeStore.Get(path)
	if !exists {
		info, exists = routeStore.Get(path + "/index")
	}

	if !exists {
		staticPath, ok := resolvePublicPath(path)
		if ok {
			if fileInfo, err := os.Stat(staticPath); err == nil && !fileInfo.IsDir() {
				http.ServeFile(w, r, staticPath)
				logRequest(r.Method, path, http.StatusOK, start)
				return
			}
		}
		http.Error(w, "404 - Ruta no encontrada", http.StatusNotFound)
		logRequest(r.Method, path, http.StatusNotFound, start)
		return
	}

	// HTML estático
	if info.Type == "html" {
		content, err := os.ReadFile(info.File)
		if err != nil {
			http.Error(w, "Error leyendo archivo", http.StatusInternalServerError)
			logRequest(r.Method, path, http.StatusInternalServerError, start)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(injectDevReloadScript(string(content))))
		logRequest(r.Method, path, http.StatusOK, start)
		return
	}

	// JS estático
	if info.Type == "js" {
		content, err := os.ReadFile(info.File)
		if err != nil {
			http.Error(w, "Error leyendo archivo", http.StatusInternalServerError)
			logRequest(r.Method, path, http.StatusInternalServerError, start)
			return
		}
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Write(content)
		logRequest(r.Method, path, http.StatusOK, start)
		return
	}

	// JSX/TSX
	if info.Type == "page" {
		bundledCode, err := bundleJSX(info.File, info.IsTS)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(renderDevErrorHTML(path, err)))
			logRequest(r.Method, path, http.StatusInternalServerError, start)
			return
		}

		css := loadGlobalCSS()
		if info.CSSFile != "" {
			if scopedCSS := loadCSS(info.CSSFile); scopedCSS != "" {
				css += "\n" + scopedCSS
			}
		}

		html := generateHTML(bundledCode, css, info.Title)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
		logRequest(r.Method, path, http.StatusOK, start)
		return
	}

	http.Error(w, "404 - Ruta no encontrada", http.StatusNotFound)
	logRequest(r.Method, path, http.StatusNotFound, start)
}
