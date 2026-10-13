package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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

func TestBundleJSXSupportsReact19NamedImports(t *testing.T) {
	bundleCache.Clear()
	dir := t.TempDir()
	entry := filepath.Join(dir, "page.tsx")
	content := `import { useActionState, useOptimistic } from "react";
import { preconnect, useFormStatus } from "react-dom";

export default function App() {
  return <main>{typeof useActionState}{typeof useOptimistic}{typeof preconnect}{typeof useFormStatus}</main>;
}
`

	if err := os.WriteFile(entry, []byte(content), 0644); err != nil {
		t.Fatalf("write entry: %v", err)
	}

	if _, err := bundleJSX(entry, true); err != nil {
		t.Fatalf("expected React 19 named imports to bundle: %v", err)
	}
}

func TestBundleJSXEscapesEntryPath(t *testing.T) {
	bundleCache.Clear()
	dir := filepath.Join(t.TempDir(), "developer's app")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir entry directory: %v", err)
	}
	entry := filepath.Join(dir, "page.tsx")
	content := `export default function App() { return <main>OK</main>; }`

	if err := os.WriteFile(entry, []byte(content), 0644); err != nil {
		t.Fatalf("write entry: %v", err)
	}

	if _, err := bundleJSX(entry, true); err != nil {
		t.Fatalf("expected quoted entry path to bundle: %v", err)
	}
}

func TestBundleJSXReusesIncrementalBuildContext(t *testing.T) {
	bundleCache.Clear()
	dir := t.TempDir()
	entry := filepath.Join(dir, "page.tsx")
	if err := os.WriteFile(entry, []byte(`export default function App() { return <main>First</main>; }`), 0644); err != nil {
		t.Fatalf("write first entry: %v", err)
	}

	if _, err := bundleJSX(entry, true); err != nil {
		t.Fatalf("first build: %v", err)
	}
	absPath, err := filepath.Abs(entry)
	if err != nil {
		t.Fatalf("absolute entry path: %v", err)
	}
	key := fmt.Sprintf("%s|%t", absPath, true)
	firstContext := bundleBuilds.get(key).context
	if firstContext == nil {
		t.Fatal("expected build context after first build")
	}

	if err := os.WriteFile(entry, []byte(`export default function App() { return <main>Second</main>; }`), 0644); err != nil {
		t.Fatalf("write second entry: %v", err)
	}
	bundleCache.Clear()
	secondBundle, err := bundleJSX(entry, true)
	if err != nil {
		t.Fatalf("incremental rebuild: %v", err)
	}
	if !strings.Contains(secondBundle, "Second") {
		t.Fatalf("expected rebuilt output, got %s", secondBundle)
	}
	if secondContext := bundleBuilds.get(key).context; secondContext != firstContext {
		t.Fatal("expected esbuild context to be reused")
	}
}

func TestBundleJSXIncludesInlineSourceMap(t *testing.T) {
	bundleCache.Clear()
	dir := t.TempDir()
	entry := filepath.Join(dir, "page.tsx")
	content := `export default function App() { return <main>Mapped</main>; }`
	if err := os.WriteFile(entry, []byte(content), 0644); err != nil {
		t.Fatalf("write entry: %v", err)
	}

	bundle, err := bundleJSX(entry, true)
	if err != nil {
		t.Fatalf("bundle entry: %v", err)
	}
	const marker = "sourceMappingURL=data:application/json;base64,"
	index := strings.LastIndex(bundle, marker)
	if index == -1 {
		t.Fatal("expected inline source map")
	}
	encoded := strings.TrimSpace(bundle[index+len(marker):])
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode source map: %v", err)
	}
	var sourceMap struct {
		Sources        []string `json:"sources"`
		SourcesContent []string `json:"sourcesContent"`
	}
	if err := json.Unmarshal(decoded, &sourceMap); err != nil {
		t.Fatalf("parse source map: %v", err)
	}
	if len(sourceMap.Sources) == 0 || len(sourceMap.SourcesContent) == 0 {
		t.Fatalf("expected source paths and content in source map: %s", decoded)
	}
	if !strings.Contains(strings.Join(sourceMap.SourcesContent, "\n"), "Mapped") {
		t.Fatal("expected original TSX in source map")
	}
}

func TestBundleJSXSourceMapUsesProjectRelativePaths(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/index.tsx": `export default function App() { return <main>Mapped</main>; }`,
	})
	bundleCache.Clear()

	bundle, err := bundleJSX(filepath.Join("routes", "index.tsx"), true)
	if err != nil {
		t.Fatalf("bundle entry: %v", err)
	}
	const marker = "sourceMappingURL=data:application/json;base64,"
	index := strings.LastIndex(bundle, marker)
	if index == -1 {
		t.Fatal("expected inline source map")
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(bundle[index+len(marker):]))
	if err != nil {
		t.Fatalf("decode source map: %v", err)
	}
	var sourceMap struct {
		Sources []string `json:"sources"`
	}
	if err := json.Unmarshal(decoded, &sourceMap); err != nil {
		t.Fatalf("parse source map: %v", err)
	}
	for _, source := range sourceMap.Sources {
		if filepath.IsAbs(source) || pathEscapesRoot(filepath.FromSlash(source)) {
			t.Fatalf("expected project-relative source path, got %q", source)
		}
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
