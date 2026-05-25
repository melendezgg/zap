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

	port := flags.String("port", config.Port, "puerto del servidor")
	portShort := flags.String("P", config.Port, "puerto del servidor")
	host := flags.String("host", config.Host, "host del servidor")
	help := flags.Bool("help", false, "mostrar ayuda")
	helpShort := flags.Bool("h", false, "mostrar ayuda")
	versionFlag := flags.Bool("version", false, "mostrar version")
	versionShort := flags.Bool("v", false, "mostrar version")

	if err := flags.Parse(args); err != nil {
		return cliActionServe, fmt.Errorf("%w\nusa `zap --help` para ver las opciones disponibles", err)
	}

	if flags.NArg() > 0 {
		return cliActionServe, fmt.Errorf("argumento desconocido: %s\nusa `zap --help` para ver las opciones disponibles", flags.Arg(0))
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
		return fmt.Errorf("host invalido: no puede estar vacio")
	}
	return nil
}

func validatePort(port string) error {
	if strings.Contains(port, ":") {
		return fmt.Errorf("puerto invalido %q: usa solo el numero, por ejemplo `--port 3000`", port)
	}
	value, err := strconv.Atoi(port)
	if err != nil || value < 1 || value > 65535 {
		return fmt.Errorf("puerto invalido %q: debe estar entre 1 y 65535", port)
	}
	return nil
}

func listenAddress() string {
	return net.JoinHostPort(config.Host, config.Port)
}

func displayURL() string {
	return "http://" + net.JoinHostPort(config.Host, config.Port)
}

// === MOSTRAR AYUDA ===
func printUsage() {
	fmt.Printf(`
ZAP - Runtime de desarrollo React/TypeScript
Versión: %s

USO:
  zap                        Iniciar servidor de desarrollo
  zap --port 3000            Puerto personalizado
  zap --host 0.0.0.0         Host personalizado
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
  - React 19.2 embebido
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
