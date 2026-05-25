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
		"routes/index.tsx": "custom",
	})

	if err := createStarterProject(); err != nil {
		t.Fatalf("expected existing file to be skipped without error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(routesDir, "index.tsx"))
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if string(content) != "custom" {
		t.Fatalf("expected existing route to remain unchanged, got %q", content)
	}
}
