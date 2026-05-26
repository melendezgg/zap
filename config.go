package main

import (
	"path/filepath"
	"strings"
)

// Configuration
var config = &Config{
	Host: "localhost",
	Port: "8080",
}

type Config struct {
	Host string
	Port string
}

// Directories
const (
	routesDir = "routes"
	publicDir = "public"
)

func shouldIgnore(path string) bool {
	base := filepath.Base(path)
	return base == ".DS_Store" || base == "Thumbs.db" || base == ".gitignore" ||
		strings.HasSuffix(base, ".swp") || strings.HasSuffix(base, ".swo") ||
		strings.HasSuffix(base, ".swn") || strings.HasSuffix(base, ".bak") ||
		strings.HasSuffix(base, "~")
}
