package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
