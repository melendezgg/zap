package main

import (
	"fmt"
	"os"
)

// === PARSE ARGUMENTOS ===
func parseArgs() {
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port", "-P":
			if i+1 < len(args) {
				config.Port = ":" + args[i+1]
				i++
			}
		}
	}
}

// === MOSTRAR AYUDA ===
func printUsage() {
	fmt.Printf(`
ZAP - Runtime de desarrollo React/TypeScript
Versión: %s

USO:
  zap                        Iniciar servidor de desarrollo
  zap --port 3000            Puerto personalizado
  zap --help                 Esta ayuda
  zap --version              Mostrar versión

NOTA:
  0.1 está enfocado en desarrollo.
  No hay modo de producción ni comando de build todavía.

INICIO RÁPIDO:
  mkdir mi-app && cd mi-app && zap
  -> Proyecto creado automáticamente

CARACTERÍSTICAS:
  - Un solo binario
  - React 18 vía CDN
  - Hot-reload automático
  - TypeScript/JSX nativo
  - HTML/JS estático soportado
  - Título por página (export const title)
  - CSS global vía /public/styles/global.css
  - Archivos privados con prefijo _

ESTRUCTURA:
  /routes/*.tsx,jsx,html,js  -> Rutas públicas
  /routes/_*                 -> Módulos privados, no rutas públicas
  /public/styles/global.css  -> Estilos globales automáticos
  /public/                   -> Assets estáticos (opcional)
`, version)
}

func printVersion() {
	fmt.Println(version)
}
