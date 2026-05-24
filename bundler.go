package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
)

// === BUNDLE JSX/TSX ===
func bundleJSX(entryFile string, isTS bool) (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cwd error: %v", err)
	}
	absPath, err := filepath.Abs(entryFile)
	if err != nil {
		return "", fmt.Errorf("path error: %v", err)
	}

	cacheKey := fmt.Sprintf("%s|%t", absPath, isTS)
	if cachedBundle, ok := bundleCache.Get(cacheKey); ok {
		return cachedBundle, nil
	}

	virtualEntry := fmt.Sprintf("import App from '%s'; window.App = App;", absPath)

	result := api.Build(api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents:   virtualEntry,
			ResolveDir: workingDir,
			Sourcefile: "entry.js",
		},
		Write:             false,
		Format:            api.FormatIIFE,
		Target:            api.ES2020,
		Bundle:            true,
		MinifyWhitespace:  false,
		MinifyIdentifiers: false,
		MinifySyntax:      false,
		Sourcemap:         api.SourceMapNone,
		Platform:          api.PlatformBrowser,
		LegalComments:     api.LegalCommentsNone,
		TreeShaking:       api.TreeShakingTrue,
		AbsWorkingDir:     workingDir,
		// ✅ Sin configuración JSX especial (usa React por defecto)
	})

	if len(result.Errors) > 0 {
		msg := ""
		for _, e := range result.Errors {
			msg += e.Text + "\n"
		}
		return "", fmt.Errorf("esbuild: %s", msg)
	}

	if len(result.OutputFiles) > 0 {
		output := string(result.OutputFiles[0].Contents)
		bundleCache.Set(cacheKey, output)
		return output, nil
	}

	return "", fmt.Errorf("no output")
}
