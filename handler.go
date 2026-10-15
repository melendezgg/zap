package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	path := r.URL.Path
	setDevResponseHeaders(w)

	if handled, status := handleVendorAsset(w, r); handled {
		logRequest(r.Method, path, status, start)
		return
	}

	if path == "/__zap/events" {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, "GET")
			logRequest(r.Method, path, http.StatusMethodNotAllowed, start)
			return
		}
		handleDevEvents(w, r)
		return
	}

	if path == "/__zap/update" {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, "GET")
			logRequest(r.Method, path, http.StatusMethodNotAllowed, start)
			return
		}
		handleDevUpdate(w, r)
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
			http.Error(w, "404 - Route not found", http.StatusNotFound)
			logRequest(r.Method, path, http.StatusNotFound, start)
			return
		}
		if isSensitivePublicFile(faviconPath) {
			blockSensitivePublicFile(w, r, path, start)
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

	info, exists := findRoute(path)

	if !exists {
		staticPath, ok := resolvePublicPath(path)
		if ok {
			if fileInfo, err := os.Stat(staticPath); err == nil && !fileInfo.IsDir() {
				if isSensitivePublicFile(staticPath) {
					blockSensitivePublicFile(w, r, path, start)
					return
				}
				http.ServeFile(w, r, staticPath)
				logRequest(r.Method, path, http.StatusOK, start)
				return
			}
		}
		http.Error(w, "404 - Route not found", http.StatusNotFound)
		logRequest(r.Method, path, http.StatusNotFound, start)
		return
	}

	// Static HTML
	if info.Type == "html" {
		content, err := os.ReadFile(info.File)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			logRequest(r.Method, path, http.StatusInternalServerError, start)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		writeResponse(w, r, http.StatusOK, []byte(injectDevReloadScript(string(content))))
		logRequest(r.Method, path, http.StatusOK, start)
		return
	}

	// Static JS
	if info.Type == "js" {
		content, err := os.ReadFile(info.File)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
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
			logCompileError(path, err)
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

	http.Error(w, "404 - Route not found", http.StatusNotFound)
	logRequest(r.Method, path, http.StatusNotFound, start)
}

type devUpdateResponse struct {
	Code  string `json:"code,omitempty"`
	CSS   string `json:"css,omitempty"`
	Error string `json:"error,omitempty"`
}

func findRoute(path string) (*RouteInfo, bool) {
	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	info, exists := routeStore.Get(path)
	if !exists {
		info, exists = routeStore.Get(path + "/index")
	}
	return info, exists
}

func handleDevUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	info, exists := findRoute(r.URL.Query().Get("route"))
	if !exists || info.Type != "page" {
		writeJSONResponse(w, r, http.StatusNotFound, devUpdateResponse{Error: "React route not found"})
		return
	}

	bundledCode, err := bundleJSX(info.File, info.IsTS)
	if err != nil {
		logCompileError(r.URL.Query().Get("route"), err)
		writeJSONResponse(w, r, http.StatusInternalServerError, devUpdateResponse{Error: browserErrorMessage(err)})
		return
	}

	css := loadGlobalCSS()
	if info.CSSFile != "" {
		if scopedCSS := loadCSS(info.CSSFile); scopedCSS != "" {
			css += "\n" + scopedCSS
		}
	}
	writeJSONResponse(w, r, http.StatusOK, devUpdateResponse{Code: bundledCode, CSS: css})
}

func writeJSONResponse(w http.ResponseWriter, r *http.Request, status int, value any) {
	content, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
	writeResponse(w, r, status, content)
}

func isReadMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

func methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	http.Error(w, "405 - Method not allowed", http.StatusMethodNotAllowed)
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

func blockSensitivePublicFile(w http.ResponseWriter, r *http.Request, path string, start time.Time) {
	http.Error(w, "403 - Public file blocked for safety", http.StatusForbidden)
	logRequest(r.Method, path, http.StatusForbidden, start)
}
