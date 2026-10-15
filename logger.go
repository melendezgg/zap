package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

var devLogger = slog.New(slog.NewTextHandler(os.Stderr, nil))

// === REQUEST LOGGING ===
func logRequest(method, path string, status int, start time.Time) {
	if shouldSkipRequestLog(path, status) {
		return
	}

	duration := time.Since(start)
	fmt.Printf("[%s] %s - %d (%v)\n", method, path, status, duration)
}

func logCompileError(route string, err error) {
	attributes := []any{"route", route, "error", err.Error()}
	var buildErr *compileError
	if errors.As(err, &buildErr) {
		attributes = append(attributes, "phase", buildErr.Phase, "diagnostics", len(buildErr.Messages))
		if len(buildErr.Messages) > 0 && buildErr.Messages[0].Location != nil {
			location := buildErr.Messages[0].Location
			attributes = append(attributes,
				"file", safeDisplayPath(location.File),
				"line", location.Line,
				"column", location.Column+1,
			)
		}
	}
	devLogger.Error("compilation failed", attributes...)
}

func shouldSkipRequestLog(path string, status int) bool {
	if status >= 400 {
		return false
	}
	return strings.HasPrefix(path, "/__zap/assets/react/") || path == "/__zap/events"
}
