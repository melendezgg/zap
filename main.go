package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	if !isLocalOnlyHost(config.Host) {
		fmt.Println("Warning: network clients can access development bundles, source maps, and diagnostics.")
	}
	printNetworkAccess()
	fmt.Printf("Routes: %d\n\n", routeStore.Len())
	fmt.Println("Hot reload active")
	fmt.Printf("React %s + Fast Refresh %s (embedded)\n", reactVersion, reactRefreshVersion)
	fmt.Println()
	fmt.Println("Server ready. Press Ctrl+C to stop.")
	fmt.Println()

	go watchChanges()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	server := &http.Server{
		Addr:    listenAddress(),
		Handler: mux,
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	stopSignals := make(chan os.Signal, 1)
	signal.Notify(stopSignals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stopSignals)

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		fmt.Printf("error starting server at %s: %v\n", listenAddress(), err)
		os.Exit(1)
	case <-stopSignals:
		fmt.Println("\nShutting down server...")
	}

	devReload.close()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("error shutting down server: %v\n", err)
		os.Exit(1)
	}
	bundleBuilds.dispose()

	fmt.Println("Server stopped.")
}
