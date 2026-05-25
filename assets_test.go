package main

import (
	"path/filepath"
	"testing"
)

func TestIsSensitivePublicFile(t *testing.T) {
	sensitive := []string{
		"public/.env",
		"public/.env.local",
		"public/.env.preview",
		"public/config.json",
		"public/secrets.yaml",
		"public/private.pem",
		"public/database.sqlite",
		"public/id_rsa",
	}

	for _, path := range sensitive {
		if !isSensitivePublicFile(path) {
			t.Fatalf("expected %s to be sensitive", path)
		}
	}
}

func TestIsSensitivePublicFileAllowsNormalAssets(t *testing.T) {
	normal := []string{
		"public/app.js",
		"public/configurator.js",
		"public/settings.css",
		"public/app.json",
		"public/images/logo.svg",
	}

	for _, path := range normal {
		if isSensitivePublicFile(path) {
			t.Fatalf("expected %s to be allowed", path)
		}
	}
}

func TestFindSensitivePublicFiles(t *testing.T) {
	withTempProject(t, map[string]string{
		"public/.env":              "TOKEN=secret",
		"public/config.json":       `{"token":"secret"}`,
		"public/assets/app.js":     "console.log('ok');",
		"public/assets/id_ed25519": "secret",
	})

	matches := findSensitivePublicFiles()
	expected := map[string]bool{
		filepath.ToSlash(filepath.Join(publicDir, ".env")):              true,
		filepath.ToSlash(filepath.Join(publicDir, "config.json")):       true,
		filepath.ToSlash(filepath.Join(publicDir, "assets/id_ed25519")): true,
	}

	if len(matches) != len(expected) {
		t.Fatalf("expected %d matches, got %#v", len(expected), matches)
	}
	for _, match := range matches {
		if !expected[match] {
			t.Fatalf("unexpected sensitive match %s in %#v", match, matches)
		}
	}
}
