package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var routeExtensionPriority = map[string]int{
	".tsx":  0,
	".jsx":  1,
	".html": 2,
	".js":   3,
}

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

	files := collectRouteFiles(routesDir)
	for _, path := range files {
		route, ok := routePathForFile(routesDir, path)
		if !ok {
			continue
		}

		next := routeInfoForFile(path)
		if current, exists := routes[route]; exists {
			winner, loser := chooseRouteFile(current.File, next.File)
			if winner == current.File {
				fmt.Printf("  advertencia: conflicto de ruta %s, usando %s e ignorando %s\n", route, current.File, loser)
				continue
			}

			fmt.Printf("  advertencia: conflicto de ruta %s, usando %s e ignorando %s\n", route, next.File, loser)
		}

		fmt.Printf("  %s -> %s\n", route, path)
		routes[route] = next
	}

	return routes
}

func collectRouteFiles(root string) []string {
	var files []string

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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

		files = append(files, path)
		return nil
	})

	sort.Strings(files)
	return files
}

func routePathForFile(root string, path string) (string, bool) {
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		fmt.Printf("error resolviendo ruta %s: %v\n", path, err)
		return "", false
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

	return route, true
}

func routeInfoForFile(path string) *RouteInfo {
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

	return &RouteInfo{
		File:    path,
		Type:    fileType,
		IsTS:    strings.HasSuffix(path, ".tsx"),
		CSSFile: "",
		Title:   title,
	}
}

func chooseRouteFile(a string, b string) (winner string, loser string) {
	aPriority := routeFilePriority(a)
	bPriority := routeFilePriority(b)

	if aPriority < bPriority {
		return a, b
	}
	if bPriority < aPriority {
		return b, a
	}
	if a <= b {
		return a, b
	}
	return b, a
}

func routeFilePriority(path string) int {
	priority, ok := routeExtensionPriority[filepath.Ext(path)]
	if !ok {
		return len(routeExtensionPriority)
	}
	return priority
}
