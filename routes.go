package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// === EXTRAER TÍTULO DE TSX/JSX ===
func extractTitle(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Zap App"
	}

	regex := regexp.MustCompile(`export\s+const\s+title\s*=\s*["']([^"']+)["']`)
	matches := regex.FindSubmatch(content)

	if len(matches) > 1 {
		return string(matches[1])
	}

	return "Zap App"
}

// === EXTRAER TÍTULO DE HTML ===
func extractTitleFromHTML(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Zap App"
	}

	regex := regexp.MustCompile(`(?i)<title>([^<]+)</title>`)
	matches := regex.FindSubmatch(content)

	if len(matches) > 1 {
		return strings.TrimSpace(string(matches[1]))
	}

	return "Zap App"
}

// === ESCANEAR RUTAS ===
func scanAllRoutes() map[string]*RouteInfo {
	routes := make(map[string]*RouteInfo)

	if err := os.MkdirAll(routesDir, 0755); err != nil {
		fmt.Printf("error creando %s: %v\n", routesDir, err)
		return routes
	}

	filepath.Walk(routesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if shouldIgnore(path) {
			return nil
		}

		if !strings.HasSuffix(path, ".tsx") && !strings.HasSuffix(path, ".jsx") &&
			!strings.HasSuffix(path, ".html") && !strings.HasSuffix(path, ".js") {
			return nil
		}

		baseName := filepath.Base(path)
		if strings.HasPrefix(baseName, "_") {
			return nil
		}

		relPath, err := filepath.Rel(routesDir, path)
		if err != nil {
			fmt.Printf("error resolviendo ruta %s: %v\n", path, err)
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		route := relPath
		route = strings.TrimSuffix(route, ".tsx")
		route = strings.TrimSuffix(route, ".jsx")
		route = strings.TrimSuffix(route, ".html")
		if !strings.HasSuffix(path, ".js") {
			route = strings.TrimSuffix(route, "/index")
		}

		if route == "" || route == "index" {
			route = "/"
		} else if !strings.HasPrefix(route, "/") {
			route = "/" + route
		}

		fileType := "page"
		if strings.HasSuffix(path, ".html") {
			fileType = "html"
		} else if strings.HasSuffix(path, ".js") {
			fileType = "js"
		}

		title := "Zap App"
		if fileType == "page" {
			title = extractTitle(path)
		} else if fileType == "html" {
			title = extractTitleFromHTML(path)
		}

		fmt.Printf("  %s -> %s\n", route, path)

		routes[route] = &RouteInfo{
			File:    path,
			Type:    fileType,
			IsTS:    strings.HasSuffix(path, ".tsx"),
			CSSFile: "",
			Title:   title,
		}

		return nil
	})

	return routes
}
