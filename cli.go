package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type cliAction int

const (
	cliActionServe cliAction = iota
	cliActionHelp
	cliActionVersion
)

func parseArgs(args []string) (cliAction, error) {
	flags := flag.NewFlagSet("zap", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	port := flags.String("port", config.Port, "server port")
	portShort := flags.String("P", config.Port, "server port")
	host := flags.String("host", config.Host, "server host")
	help := flags.Bool("help", false, "show help")
	helpShort := flags.Bool("h", false, "show help")
	versionFlag := flags.Bool("version", false, "show version")
	versionShort := flags.Bool("v", false, "show version")

	if err := flags.Parse(args); err != nil {
		return cliActionServe, fmt.Errorf("%w\nuse `zap --help` to see available options", err)
	}

	if flags.NArg() > 0 {
		return cliActionServe, fmt.Errorf("unknown argument: %s\nuse `zap --help` to see available options", flags.Arg(0))
	}

	if *help || *helpShort {
		return cliActionHelp, nil
	}
	if *versionFlag || *versionShort {
		return cliActionVersion, nil
	}

	selectedPort := *port
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "P" {
			selectedPort = *portShort
		}
	})

	selectedHost := strings.TrimSpace(*host)
	if err := validateHost(selectedHost); err != nil {
		return cliActionServe, err
	}
	if err := validatePort(selectedPort); err != nil {
		return cliActionServe, err
	}

	config.Host = selectedHost
	config.Port = selectedPort
	return cliActionServe, nil
}

func validateHost(host string) error {
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("invalid host: cannot be empty")
	}
	return nil
}

func validatePort(port string) error {
	if strings.Contains(port, ":") {
		return fmt.Errorf("invalid port %q: use only the number, for example `--port 3000`", port)
	}
	value, err := strconv.Atoi(port)
	if err != nil || value < 1 || value > 65535 {
		return fmt.Errorf("invalid port %q: must be between 1 and 65535", port)
	}
	return nil
}

func listenAddress() string {
	return net.JoinHostPort(config.Host, config.Port)
}

func displayURL() string {
	return "http://" + net.JoinHostPort(config.Host, config.Port)
}

func isLocalOnlyHost(host string) bool {
	if strings.EqualFold(strings.TrimSpace(host), "localhost") {
		return true
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	return ip != nil && ip.IsLoopback()
}

// === SHOW HELP ===
func printUsage() {
	fmt.Printf(`
ZAP - React/TypeScript development runtime
Version: %s

USAGE:
  zap                        Start the development server
  zap --port 3000            Use a custom port
  zap --host 0.0.0.0         Use a custom host
  zap --help                 Show this help
  zap --version              Show version

NOTE:
  0.1 is focused on development.
  There is no production mode or build command yet.

QUICK START:
  mkdir my-app && cd my-app && zap
  -> Project created automatically

FEATURES:
  - Single binary
  - React 19.2 embedded
  - React Fast Refresh with state preservation
  - Compile error overlay with last-good UI
  - Network URLs and QR code for mobile preview
  - Native TypeScript/JSX
  - Static HTML/JS support
  - Per-page title (export const title)
  - Global CSS via /public/styles/global.css
  - Private files with _ prefix

STRUCTURE:
  /routes/*.tsx,jsx,html,js  -> Public routes
  /routes/_*                 -> Private modules, not public routes
  /public/styles/global.css  -> Automatic global styles
  /public/                   -> Static assets (optional)
`, version)
}

func printVersion() {
	fmt.Println(version)
}
