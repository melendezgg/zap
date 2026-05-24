package main

import (
	"fmt"
	"net/http"
	"os"
)

var version = "dev"

// === MAIN ===
func main() {
	parseArgs()
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "--help", "-h":
			printUsage()
			return
		case "--version", "-v":
			printVersion()
			return
		}
	}

	magicInit()

	fmt.Println("Escaneando...")
	routeStore.SetAll(scanAllRoutes())

	fmt.Printf("\nZAP %s (DESARROLLO) en http://localhost%s\n", version, config.Port)
	fmt.Printf("Rutas: %d\n\n", routeStore.Len())
	fmt.Println("Hot-reload activo")
	fmt.Println("React 18 (CDN)")
	fmt.Println()
	fmt.Println("Servidor listo. Ctrl+C para detener.")
	fmt.Println()

	go watchChanges()
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(config.Port, nil); err != nil {
		fmt.Printf("error iniciando servidor en %s: %v\n", config.Port, err)
		os.Exit(1)
	}
}
