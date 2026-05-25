package main

import (
	"fmt"
	"net/http"
	"os"
)

var version = "dev"

// === MAIN ===
func main() {
	action, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	switch action {
	case cliActionHelp:
		printUsage()
		return
	case cliActionVersion:
		printVersion()
		return
	}

	magicInit()

	fmt.Println("Scanning...")
	routeStore.SetAll(scanAllRoutes())
	warnSensitivePublicFiles()

	fmt.Printf("\nZAP %s (DEVELOPMENT) at %s\n", version, displayURL())
	fmt.Printf("Routes: %d\n\n", routeStore.Len())
	fmt.Println("Hot reload active")
	fmt.Printf("React %s (embedded)\n", reactVersion)
	fmt.Println()
	fmt.Println("Server ready. Press Ctrl+C to stop.")
	fmt.Println()

	go watchChanges()
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(listenAddress(), nil); err != nil {
		fmt.Printf("error starting server at %s: %v\n", listenAddress(), err)
		os.Exit(1)
	}
}
