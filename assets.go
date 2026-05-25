package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var sensitivePublicBaseNames = map[string]bool{
	".env":            true,
	".env.local":      true,
	".env.test":       true,
	".env.production": true,
	"id_rsa":          true,
	"id_ed25519":      true,
}

var sensitivePublicExtensions = map[string]bool{
	".key":    true,
	".pem":    true,
	".crt":    true,
	".p12":    true,
	".pfx":    true,
	".sqlite": true,
	".db":     true,
}

var sensitivePublicExactFiles = map[string]bool{
	"config.json":  true,
	"config.yaml":  true,
	"config.yml":   true,
	"secrets.json": true,
	"secrets.yaml": true,
	"secrets.yml":  true,
}

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

func isSensitivePublicFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if sensitivePublicBaseNames[base] || sensitivePublicExactFiles[base] {
		return true
	}
	if strings.HasPrefix(base, ".env.") {
		return true
	}
	return sensitivePublicExtensions[strings.ToLower(filepath.Ext(base))]
}

func warnSensitivePublicFiles() {
	matches := findSensitivePublicFiles()
	for _, path := range matches {
		fmt.Printf("warning: %s looks sensitive and will not be served\n", path)
	}
}

func findSensitivePublicFiles() []string {
	var matches []string

	if _, err := os.Stat(publicDir); err != nil {
		return matches
	}

	filepath.WalkDir(publicDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry == nil || entry.IsDir() {
			return nil
		}
		if shouldIgnore(path) {
			return nil
		}
		if isSensitivePublicFile(path) {
			matches = append(matches, filepath.ToSlash(path))
		}
		return nil
	})

	return matches
}
