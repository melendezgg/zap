package main

import (
	"os"
	"path/filepath"
	"strings"
)

// === CARGAR CSS ===
func loadGlobalCSS() string {
	globalCSSPath := filepath.Join(publicDir, "styles", "global.css")
	content, err := os.ReadFile(globalCSSPath)
	if err != nil {
		return ""
	}
	return string(content)
}

func loadCSS(f string) string {
	if f == "" {
		return ""
	}
	content, err := os.ReadFile(f)
	if err != nil {
		return ""
	}
	return string(content)
}

func resolvePublicPath(requestPath string) (string, bool) {
	cleanPath := filepath.Clean("/" + requestPath)
	trimmedPath := strings.TrimPrefix(cleanPath, "/")

	publicRoot, err := filepath.Abs(publicDir)
	if err != nil {
		return "", false
	}

	resolvedPath, err := filepath.Abs(filepath.Join(publicRoot, trimmedPath))
	if err != nil {
		return "", false
	}

	relToPublic, err := filepath.Rel(publicRoot, resolvedPath)
	if err != nil {
		return "", false
	}

	if relToPublic == "." {
		return resolvedPath, true
	}

	if strings.HasPrefix(relToPublic, "..") || filepath.IsAbs(relToPublic) {
		return "", false
	}

	return resolvedPath, true
}
