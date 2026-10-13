package main

import (
	"strings"
	"testing"
)

func resetConfig() {
	config.Host = "localhost"
	config.Port = "8080"
}

func TestParseArgsDefaultsToServe(t *testing.T) {
	resetConfig()

	action, err := parseArgs(nil)
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}

	if action != cliActionServe {
		t.Fatalf("expected serve action, got %v", action)
	}
	if config.Host != "localhost" || config.Port != "8080" {
		t.Fatalf("unexpected config: host=%q port=%q", config.Host, config.Port)
	}
}

func TestParseArgsAcceptsHostAndPort(t *testing.T) {
	resetConfig()

	action, err := parseArgs([]string{"--host", "0.0.0.0", "--port", "3000"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}

	if action != cliActionServe {
		t.Fatalf("expected serve action, got %v", action)
	}
	if config.Host != "0.0.0.0" || config.Port != "3000" {
		t.Fatalf("unexpected config: host=%q port=%q", config.Host, config.Port)
	}
}

func TestParseArgsRejectsInvalidPort(t *testing.T) {
	resetConfig()

	_, err := parseArgs([]string{"--port", ":3000"})
	if err == nil {
		t.Fatal("expected invalid port error")
	}
	if !strings.Contains(err.Error(), "use only the number") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseArgsReturnsHelpAndVersionActions(t *testing.T) {
	resetConfig()

	action, err := parseArgs([]string{"--help"})
	if err != nil {
		t.Fatalf("parseArgs help returned error: %v", err)
	}
	if action != cliActionHelp {
		t.Fatalf("expected help action, got %v", action)
	}

	action, err = parseArgs([]string{"-v"})
	if err != nil {
		t.Fatalf("parseArgs version returned error: %v", err)
	}
	if action != cliActionVersion {
		t.Fatalf("expected version action, got %v", action)
	}
}

func TestIsLocalOnlyHost(t *testing.T) {
	for _, host := range []string{"localhost", "LOCALHOST", "127.0.0.1", "::1"} {
		if !isLocalOnlyHost(host) {
			t.Fatalf("expected %q to be local-only", host)
		}
	}
	for _, host := range []string{"0.0.0.0", "::", "192.168.1.20", "devbox.local"} {
		if isLocalOnlyHost(host) {
			t.Fatalf("expected %q to be network-accessible", host)
		}
	}
}
