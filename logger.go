package main

import (
	"fmt"
	"strings"
	"time"
)

// === LOG DE REQUESTS ===
func logRequest(method, path string, status int, start time.Time) {
	if shouldSkipRequestLog(path, status) {
		return
	}

	duration := time.Since(start)
	fmt.Printf("[%s] %s - %d (%v)\n", method, path, status, duration)
}

func shouldSkipRequestLog(path string, status int) bool {
	if status >= 400 {
		return false
	}
	return strings.HasPrefix(path, "/__zap/assets/react/") || path == "/__zap/events"
}
