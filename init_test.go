package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateStarterProjectCreatesStarterRoutes(t *testing.T) {
	withTempProject(t, map[string]string{})

	if err := createStarterProject(); err != nil {
		t.Fatalf("createStarterProject returned error: %v", err)
	}

	indexPath := filepath.Join(routesDir, "index.tsx")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read %s: %v", indexPath, err)
	}
	if !strings.Contains(string(content), `import { useState } from "react";`) {
		t.Fatalf("expected React import in starter page: %s", content)
	}
	if !strings.Contains(string(content), `className="container"`) {
		t.Fatalf("expected starter page layout classes: %s", content)
	}
	if !strings.Contains(string(content), `Count: {count}`) {
		t.Fatalf("expected English starter content: %s", content)
	}

	if _, err := os.Stat(filepath.Join(routesDir, "about.tsx")); err != nil {
		t.Fatalf("expected about route: %v", err)
	}
	if _, err := os.Stat(filepath.Join(publicDir, "styles", "global.css")); err != nil {
		t.Fatalf("expected global css: %v", err)
	}
}

func TestCreateStarterProjectDoesNotOverwriteExistingFiles(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/index.tsx":         "custom route",
		"public/styles/global.css": "custom styles",
	})

	if err := createStarterProject(); err != nil {
		t.Fatalf("expected existing file to be skipped without error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(routesDir, "index.tsx"))
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if string(content) != "custom route" {
		t.Fatalf("expected existing route to remain unchanged, got %q", content)
	}

	styles, err := os.ReadFile(filepath.Join(publicDir, "styles", "global.css"))
	if err != nil {
		t.Fatalf("read global styles: %v", err)
	}
	if string(styles) != "custom styles" {
		t.Fatalf("expected existing global styles to remain unchanged, got %q", styles)
	}
}

func TestMagicInitSkipsProjectsWithRoutesDirectory(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/custom.tsx": `export default function Custom() { return null; }`,
	})

	if initialized := magicInit(); initialized {
		t.Fatal("expected automatic initialization to be skipped")
	}
	if _, err := os.Stat(filepath.Join(routesDir, "about.tsx")); !os.IsNotExist(err) {
		t.Fatalf("expected starter routes not to be created, got err=%v", err)
	}
	if _, err := os.Stat(publicDir); !os.IsNotExist(err) {
		t.Fatalf("expected optional public directory not to be created, got err=%v", err)
	}
}
