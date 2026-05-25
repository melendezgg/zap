package main

import (
	"path/filepath"
	"strings"
)

// Configuración
var config = &Config{
	Host: "localhost",
	Port: "8080",
}

type Config struct {
	Host string
	Port string
}

// Directorios
const (
	routesDir = "routes"
	publicDir = "public"
)

// Archivos a ignorar
var ignoreFiles = map[string]bool{
	".DS_Store":  true,
	"Thumbs.db":  true,
	".gitignore": true,
	".swp":       true,
	".swo":       true,
	".swn":       true,
	".bak":       true,
	"~":          true,
}

func shouldIgnore(path string) bool {
	base := filepath.Base(path)
	for ignore := range ignoreFiles {
		if strings.HasPrefix(base, ".") && ignore == base {
			return true
		}
		if strings.HasSuffix(base, ignore) {
			return true
		}
		if strings.HasSuffix(base, "~") {
			return true
		}
	}
	return false
}
