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

	fmt.Println("Escaneando...")
	routeStore.SetAll(scanAllRoutes())

	fmt.Printf("\nZAP %s (DESARROLLO) en %s\n", version, displayURL())
	fmt.Printf("Rutas: %d\n\n", routeStore.Len())
	fmt.Println("Hot-reload activo")
	fmt.Printf("React %s (embebido)\n", reactVersion)
	fmt.Println()
	fmt.Println("Servidor listo. Ctrl+C para detener.")
	fmt.Println()

	go watchChanges()
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(listenAddress(), nil); err != nil {
		fmt.Printf("error iniciando servidor en %s: %v\n", listenAddress(), err)
		os.Exit(1)
	}
}
