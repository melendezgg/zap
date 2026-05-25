package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

func TestBundleJSXSupportsReactImports(t *testing.T) {
	bundleCache.Clear()
	dir := t.TempDir()
	entry := filepath.Join(dir, "page.tsx")
	content := `import React, { useState } from "react";

export default function App() {
  const [count] = useState(1);
  return <main>{React.createElement("span", null, count)}</main>;
}
`

	if err := os.WriteFile(entry, []byte(content), 0644); err != nil {
		t.Fatalf("write entry: %v", err)
	}

	bundle, err := bundleJSX(entry, true)
	if err != nil {
		t.Fatalf("bundleJSX returned error: %v", err)
	}

	if !strings.Contains(bundle, "window.React") {
		t.Fatalf("expected bundle to reference window.React: %s", bundle)
	}
}

func TestBundleJSXSupportsReactDOMClientImports(t *testing.T) {
	bundleCache.Clear()
	dir := t.TempDir()
	entry := filepath.Join(dir, "page.jsx")
	content := `import { createRoot } from "react-dom/client";

export default function App() {
  return <main>{typeof createRoot}</main>;
}
`

	if err := os.WriteFile(entry, []byte(content), 0644); err != nil {
		t.Fatalf("write entry: %v", err)
	}

	bundle, err := bundleJSX(entry, false)
	if err != nil {
		t.Fatalf("bundleJSX returned error: %v", err)
	}

	if !strings.Contains(bundle, "window.ReactDOM") {
		t.Fatalf("expected bundle to reference window.ReactDOM: %s", bundle)
	}
}

func TestBundleJSXReturnsUsefulCompileErrors(t *testing.T) {
	bundleCache.Clear()
	dir := t.TempDir()
	entry := filepath.Join(dir, "broken.tsx")
	content := `export default function App() {
  return <main>
}
`

	if err := os.WriteFile(entry, []byte(content), 0644); err != nil {
		t.Fatalf("write entry: %v", err)
	}

	_, err := bundleJSX(entry, true)
	if err == nil {
		t.Fatal("expected compile error")
	}

	message := err.Error()
	if !strings.Contains(message, "broken.tsx") {
		t.Fatalf("expected file name in error: %s", message)
	}
	if !strings.Contains(message, "^") {
		t.Fatalf("expected source marker in error: %s", message)
	}
}

func TestFormatBuildMessageIncludesLocationAndSourceLine(t *testing.T) {
	message := formatBuildMessage(api.Message{
		Text: "Expected identifier but found \"}\"",
		Location: &api.Location{
			File:     "routes/index.tsx",
			Line:     3,
			Column:   9,
			Length:   1,
			LineText: "  return }",
		},
	})

	expected := "routes/index.tsx:3:10: Expected identifier but found \"}\""
	if !strings.Contains(message, expected) {
		t.Fatalf("expected %q in %q", expected, message)
	}
	if !strings.Contains(message, "  return }\n         ^") {
		t.Fatalf("expected source marker in %q", message)
	}
}
