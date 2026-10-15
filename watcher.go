package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	watchDebounceInterval = 150 * time.Millisecond
	watchFallbackInterval = 2 * time.Second
)

type watchedFile struct {
	modTime time.Time
	size    int64
}

func collectWatchedFiles() map[string]watchedFile {
	files, err := collectWatchedFilesFrom([]string{"."})
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
			if entry == nil {
				return nil
			}
			if entry.IsDir() {
				if path != root && shouldIgnoreWatchDir(entry.Name()) {
					return filepath.SkipDir
				}
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
		return files, fmt.Errorf("%d file(s) could not be checked; first: %w", len(walkErrors), walkErrors[0])
	}

	return files, nil
}

func shouldIgnoreWatchDir(name string) bool {
	return name == ".git" || name == "node_modules"
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

func newProjectWatcher(root string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := addWatchTree(watcher, root); err != nil {
		watcher.Close()
		return nil, err
	}
	return watcher, nil
}

func addWatchTree(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry == nil || !entry.IsDir() {
			return nil
		}
		if path != root && shouldIgnoreWatchDir(entry.Name()) {
			return filepath.SkipDir
		}
		return watcher.Add(path)
	})
}

func runProjectWatcher(watcher *fsnotify.Watcher, stop <-chan struct{}, onChange func()) error {
	var debounceTimer *time.Timer
	var debounce <-chan time.Time
	defer func() {
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
	}()

	scheduleReload := func() {
		if debounceTimer == nil {
			debounceTimer = time.NewTimer(watchDebounceInterval)
			debounce = debounceTimer.C
			return
		}
		if !debounceTimer.Stop() {
			select {
			case <-debounceTimer.C:
			default:
			}
		}
		debounceTimer.Reset(watchDebounceInterval)
		debounce = debounceTimer.C
	}

	for {
		select {
		case <-stop:
			return nil
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("filesystem watcher stopped")
			}
			fmt.Printf("warning: watcher: %v\n", err)
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("filesystem watcher stopped")
			}
			if shouldIgnore(event.Name) {
				continue
			}
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() &&
					!shouldIgnoreWatchDir(filepath.Base(event.Name)) {
					if err := addWatchTree(watcher, event.Name); err != nil {
						fmt.Printf("warning: watcher: adding %s: %v\n", event.Name, err)
					}
				}
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				scheduleReload()
			}
		case <-debounce:
			debounce = nil
			onChange()
		}
	}
}

func reloadProject() {
	fmt.Println("Changes detected!")
	bundleCache.Clear()
	routes := scanAllRoutes()
	routeStore.SetAll(routes)
	bundleBuilds.prune(activeBundleBuildKeys(routes))
	warnSensitivePublicFiles()
	devReload.broadcast()
}

// watchChanges prefers native filesystem events and falls back to polling if
// the platform cannot initialize recursive watches for the project.
func watchChanges() {
	watcher, err := newProjectWatcher(".")
	if err == nil {
		runErr := runProjectWatcher(watcher, devReload.done, reloadProject)
		watcher.Close()
		if runErr == nil {
			return
		}
		fmt.Printf("warning: watcher: %v; falling back to polling\n", runErr)
	} else {
		fmt.Printf("warning: watcher: %v; falling back to polling\n", err)
	}

	watchChangesByPolling(devReload.done)
}

func watchChangesByPolling(stop <-chan struct{}) {
	lastSnapshot := collectWatchedFiles()
	ticker := time.NewTicker(watchFallbackInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			currentSnapshot := collectWatchedFiles()
			if watchedFilesChanged(lastSnapshot, currentSnapshot) {
				lastSnapshot = currentSnapshot
				reloadProject()
			}
		}
	}
}
