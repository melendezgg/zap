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
	setDevResponseHeaders(w)

	if path == "/__zap/events" {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, "GET")
			logRequest(r.Method, path, http.StatusMethodNotAllowed, start)
			return
		}
		handleDevEvents(w, r)
		return
	}

	if !isReadMethod(r.Method) {
		methodNotAllowed(w, "GET, HEAD")
		logRequest(r.Method, path, http.StatusMethodNotAllowed, start)
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
		logRequest(r.Method, path, http.StatusNoContent, start)
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
		writeResponse(w, r, http.StatusOK, []byte(injectDevReloadScript(string(content))))
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
		writeResponse(w, r, http.StatusOK, content)
		logRequest(r.Method, path, http.StatusOK, start)
		return
	}

	// JSX/TSX
	if info.Type == "page" {
		bundledCode, err := bundleJSX(info.File, info.IsTS)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			writeResponse(w, r, http.StatusInternalServerError, []byte(renderDevErrorHTML(path, err)))
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
		writeResponse(w, r, http.StatusOK, []byte(html))
		logRequest(r.Method, path, http.StatusOK, start)
		return
	}

	http.Error(w, "404 - Ruta no encontrada", http.StatusNotFound)
	logRequest(r.Method, path, http.StatusNotFound, start)
}

func isReadMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	http.Error(w, "405 - Metodo no permitido", http.StatusMethodNotAllowed)
}

func writeResponse(w http.ResponseWriter, r *http.Request, status int, content []byte) {
	w.WriteHeader(status)
	if r.Method == http.MethodHead {
		return
	}
	w.Write(content)
}

func setDevResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}
