package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type watchedFile struct {
	modTime time.Time
	size    int64
}

func collectWatchedFiles() map[string]watchedFile {
	files, err := collectWatchedFilesFrom([]string{routesDir, publicDir})
	if err != nil {
		fmt.Printf("warning: watcher: %v\n", err)
	}
	return files
}

func collectWatchedFilesFrom(roots []string) (map[string]watchedFile, error) {
	files := make(map[string]watchedFile)
	var walkErrors []error

	for _, root := range roots {
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			walkErrors = append(walkErrors, err)
			continue
		}

		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				walkErrors = append(walkErrors, err)
				return nil
			}
			if entry == nil || entry.IsDir() {
				return nil
			}
			if shouldIgnore(path) {
				return nil
			}

			info, err := entry.Info()
			if err != nil {
				walkErrors = append(walkErrors, err)
				return nil
			}

			files[filepath.ToSlash(path)] = watchedFile{
				modTime: info.ModTime(),
				size:    info.Size(),
			}
			return nil
		})
		if err != nil {
			walkErrors = append(walkErrors, err)
		}
	}

	if len(walkErrors) > 0 {
		return files, fmt.Errorf("%d archivo(s) no se pudieron revisar; primero: %w", len(walkErrors), walkErrors[0])
	}

	return files, nil
}

func watchedFilesChanged(previous, current map[string]watchedFile) bool {
	if len(previous) != len(current) {
		return true
	}

	for path, prev := range previous {
		currentFile, ok := current[path]
		if !ok {
			return true
		}
		if !currentFile.modTime.Equal(prev.modTime) || currentFile.size != prev.size {
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
			fmt.Println("Changes detected!")
			bundleCache.Clear()
			routeStore.SetAll(scanAllRoutes())
			warnSensitivePublicFiles()
			lastSnapshot = currentSnapshot
			devReload.broadcast()
		}
	}
}
