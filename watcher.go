package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func collectWatchedFiles() map[string]time.Time {
	files := make(map[string]time.Time)

	for _, d := range []string{routesDir, publicDir} {
		filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if shouldIgnore(path) {
				return nil
			}

			files[path] = info.ModTime()
			return nil
		})
	}

	return files
}

func watchedFilesChanged(previous, current map[string]time.Time) bool {
	if len(previous) != len(current) {
		return true
	}

	for path, prevModTime := range previous {
		currentModTime, ok := current[path]
		if !ok {
			return true
		}
		if !currentModTime.Equal(prevModTime) {
			return true
		}
	}

	return false
}

// === HOT-RELOAD ===
func watchChanges() {
	lastSnapshot := collectWatchedFiles()
	for {
		time.Sleep(2 * time.Second)
		currentSnapshot := collectWatchedFiles()
		if watchedFilesChanged(lastSnapshot, currentSnapshot) {
			fmt.Println("Cambios detectados!")
			bundleCache.Clear()
			routeStore.SetAll(scanAllRoutes())
			lastSnapshot = currentSnapshot
		}
	}
}
