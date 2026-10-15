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

func TestParseArgsAcceptsShortPort(t *testing.T) {
	resetConfig()

	action, err := parseArgs([]string{"-P", "3000"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if action != cliActionServe || config.Port != "3000" {
		t.Fatalf("unexpected result: action=%v port=%q", action, config.Port)
	}
}

func TestParseArgsAcceptsLoopbackHost(t *testing.T) {
	resetConfig()

	action, err := parseArgs([]string{"--host", "127.0.0.1"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if action != cliActionServe || config.Host != "127.0.0.1" {
		t.Fatalf("unexpected result: action=%v host=%q", action, config.Host)
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
	tests := []struct {
		flag string
		want cliAction
	}{
		{"--help", cliActionHelp},
		{"-h", cliActionHelp},
		{"--version", cliActionVersion},
		{"-v", cliActionVersion},
	}

	for _, test := range tests {
		t.Run(test.flag, func(t *testing.T) {
			resetConfig()
			action, err := parseArgs([]string{test.flag})
			if err != nil {
				t.Fatalf("parseArgs returned error: %v", err)
			}
			if action != test.want {
				t.Fatalf("expected action %v, got %v", test.want, action)
			}
		})
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
