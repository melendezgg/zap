package main

import "testing"

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
