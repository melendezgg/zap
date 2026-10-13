package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

func TestShouldSkipRequestLogSkipsInternalSuccessfulRequests(t *testing.T) {
	if !shouldSkipRequestLog("/__zap/assets/react/react.development.mjs", 200) {
		t.Fatal("expected successful vendor request to be skipped")
	}
	if !shouldSkipRequestLog("/__zap/events", 200) {
		t.Fatal("expected successful event stream request to be skipped")
	}
}

func TestShouldSkipRequestLogKeepsUserAndErrorRequests(t *testing.T) {
	if shouldSkipRequestLog("/about", 200) {
		t.Fatal("expected user route to be logged")
	}
	if shouldSkipRequestLog("/__zap/assets/react/missing.mjs", 404) {
		t.Fatal("expected internal errors to be logged")
	}
}

func TestLogCompileErrorIncludesStructuredContext(t *testing.T) {
	var output bytes.Buffer
	previousLogger := devLogger
	devLogger = slog.New(slog.NewTextHandler(&output, nil))
	t.Cleanup(func() { devLogger = previousLogger })

	err := newCompileError("build", []api.Message{{
		Text: "Expected identifier",
		Location: &api.Location{
			File:   "routes/broken.tsx",
			Line:   5,
			Column: 2,
		},
	}})
	logCompileError("/broken", err)

	message := output.String()
	for _, field := range []string{
		"msg=\"compilation failed\"",
		"route=/broken",
		"phase=build",
		"diagnostics=1",
		"file=routes/broken.tsx",
		"line=5",
		"column=3",
	} {
		if !strings.Contains(message, field) {
			t.Fatalf("expected structured field %q in %s", field, message)
		}
	}
}
