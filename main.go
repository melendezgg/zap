package main

import (
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/evanw/esbuild/pkg/api"
)

var version = "dev"

// Configuración
var config = &Config{
	Port: ":8080",
}

type Config struct {
	Port string
}

// Directorios
const (
	routesDir     = "routes"
	apiDir        = "api"
	componentsDir = "components"
	publicDir     = "public"
)

// Archivos a ignorar
var ignoreFiles = map[string]bool{
	".DS_Store":  true,
	"Thumbs.db":  true,
	".gitignore": true,
	".swp":       true,
	".swo":       true,
	".swn":       true,
	".bak":       true,
	"~":          true,
}

func shouldIgnore(path string) bool {
	base := filepath.Base(path)
	for ignore := range ignoreFiles {
		if strings.HasPrefix(base, ".") && ignore == base {
			return true
		}
		if strings.HasSuffix(base, ignore) {
			return true
		}
		if strings.HasSuffix(base, "~") {
			return true
		}
	}
	return false
}

// / CDNs - React
const reactCDN = `https://unpkg.com/react@18/umd/react.development.js`
const reactDOMCDN = `https://unpkg.com/react-dom@18/umd/react-dom.development.js`

// Tipos
type RouteInfo struct {
	File    string
	Type    string
	IsTS    bool
	CSSFile string
	Title   string
}

type RouteStore struct {
	mu     sync.RWMutex
	routes map[string]*RouteInfo
}

func NewRouteStore() *RouteStore {
	return &RouteStore{
		routes: make(map[string]*RouteInfo),
	}
}

func (s *RouteStore) Get(path string) (*RouteInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, ok := s.routes[path]
	return info, ok
}

func (s *RouteStore) SetAll(routes map[string]*RouteInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.routes = routes
}

func (s *RouteStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.routes)
}

var routeStore = NewRouteStore()

type BundleCache struct {
	mu      sync.RWMutex
	bundles map[string]string
}

func NewBundleCache() *BundleCache {
	return &BundleCache{
		bundles: make(map[string]string),
	}
}

func (c *BundleCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bundle, ok := c.bundles[key]
	return bundle, ok
}

func (c *BundleCache) Set(key, bundle string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bundles[key] = bundle
}

func (c *BundleCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.bundles = make(map[string]string)
}

var bundleCache = NewBundleCache()

// === BUNDLE JSX/TSX ===
func bundleJSX(entryFile string, isTS bool) (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cwd error: %v", err)
	}
	absPath, err := filepath.Abs(entryFile)
	if err != nil {
		return "", fmt.Errorf("path error: %v", err)
	}

	cacheKey := fmt.Sprintf("%s|%t", absPath, isTS)
	if cachedBundle, ok := bundleCache.Get(cacheKey); ok {
		return cachedBundle, nil
	}

	virtualEntry := fmt.Sprintf("import App from '%s'; window.App = App;", absPath)

	result := api.Build(api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents:   virtualEntry,
			ResolveDir: workingDir,
			Sourcefile: "entry.js",
		},
		Write:             false,
		Format:            api.FormatIIFE,
		Target:            api.ES2020,
		Bundle:            true,
		MinifyWhitespace:  false,
		MinifyIdentifiers: false,
		MinifySyntax:      false,
		Sourcemap:         api.SourceMapNone,
		Platform:          api.PlatformBrowser,
		LegalComments:     api.LegalCommentsNone,
		TreeShaking:       api.TreeShakingTrue,
		AbsWorkingDir:     workingDir,
		// ✅ Sin configuración JSX especial (usa React por defecto)
	})

	if len(result.Errors) > 0 {
		msg := ""
		for _, e := range result.Errors {
			msg += e.Text + "\n"
		}
		return "", fmt.Errorf("esbuild: %s", msg)
	}

	if len(result.OutputFiles) > 0 {
		output := string(result.OutputFiles[0].Contents)
		bundleCache.Set(cacheKey, output)
		return output, nil
	}

	return "", fmt.Errorf("no output")
}

// === EXTRAER TÍTULO DE TSX/JSX ===
func extractTitle(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Zap App"
	}

	regex := regexp.MustCompile(`export\s+const\s+title\s*=\s*["']([^"']+)["']`)
	matches := regex.FindSubmatch(content)

	if len(matches) > 1 {
		return string(matches[1])
	}

	return "Zap App"
}

// === EXTRAER TÍTULO DE HTML ===
func extractTitleFromHTML(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Zap App"
	}

	regex := regexp.MustCompile(`(?i)<title>([^<]+)</title>`)
	matches := regex.FindSubmatch(content)

	if len(matches) > 1 {
		return strings.TrimSpace(string(matches[1]))
	}

	return "Zap App"
}

// === ESCANEAR RUTAS ===
func scanAllRoutes() map[string]*RouteInfo {
	routes := make(map[string]*RouteInfo)

	if err := os.MkdirAll(routesDir, 0755); err != nil {
		fmt.Printf("error creando %s: %v\n", routesDir, err)
		return routes
	}

	filepath.Walk(routesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if shouldIgnore(path) {
			return nil
		}

		if !strings.HasSuffix(path, ".tsx") && !strings.HasSuffix(path, ".jsx") &&
			!strings.HasSuffix(path, ".html") && !strings.HasSuffix(path, ".js") {
			return nil
		}

		baseName := filepath.Base(path)
		if strings.HasPrefix(baseName, "_") {
			return nil
		}

		relPath, err := filepath.Rel(routesDir, path)
		if err != nil {
			fmt.Printf("error resolviendo ruta %s: %v\n", path, err)
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		route := relPath
		route = strings.TrimSuffix(route, ".tsx")
		route = strings.TrimSuffix(route, ".jsx")
		route = strings.TrimSuffix(route, ".html")
		if !strings.HasSuffix(path, ".js") {
			route = strings.TrimSuffix(route, "/index")
		}

		if route == "" || route == "index" {
			route = "/"
		} else if !strings.HasPrefix(route, "/") {
			route = "/" + route
		}

		fileType := "page"
		if strings.HasSuffix(path, ".html") {
			fileType = "html"
		} else if strings.HasSuffix(path, ".js") {
			fileType = "js"
		}

		// Extraer título
		title := "Zap App"
		if fileType == "page" {
			title = extractTitle(path)
		} else if fileType == "html" {
			title = extractTitleFromHTML(path)
		}

		fmt.Printf("  %s -> %s\n", route, path)

		routes[route] = &RouteInfo{
			File:    path,
			Type:    fileType,
			IsTS:    strings.HasSuffix(path, ".tsx"),
			CSSFile: "",
			Title:   title,
		}

		return nil
	})

	return routes
}

// === GENERAR HTML ===
func generateHTML(jsCode string, css string, title string) string {
	styleTag := ""
	if css != "" {
		styleTag = fmt.Sprintf("<style>%s</style>", css)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    %s
    <script crossorigin src="%s"></script>
    <script crossorigin src="%s"></script>
</head>
<body>
    <div id="root"></div>
    <script>
        const { useState, useEffect, useContext, createContext, useRef, useMemo, useCallback } = React;
        
        %s
        
        if (typeof window.App !== 'undefined') {
            const root = ReactDOM.createRoot(document.getElementById('root'));
            root.render(React.createElement(window.App));
        }
    </script>
</body>
</html>`, html.EscapeString(title), styleTag, reactCDN, reactDOMCDN, jsCode)
}

func renderDevErrorHTML(routePath string, err error) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Error - Zap</title>
    <style>
        body {
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
            background: #111827;
            color: #f9fafb;
            margin: 0;
            padding: 32px;
        }
        .panel {
            max-width: 960px;
            margin: 0 auto;
            background: #1f2937;
            border: 1px solid #374151;
            border-radius: 12px;
            padding: 24px;
        }
        h1 {
            margin-top: 0;
            font-size: 24px;
        }
        p {
            color: #d1d5db;
        }
        pre {
            white-space: pre-wrap;
            word-break: break-word;
            background: #0b1220;
            border-radius: 8px;
            padding: 16px;
            overflow: auto;
        }
    </style>
</head>
<body>
    <div class="panel">
        <h1>Error de compilacion</h1>
        <p>La ruta <strong>%s</strong> no pudo compilarse.</p>
        <pre>%s</pre>
    </div>
</body>
</html>`, html.EscapeString(routePath), html.EscapeString(err.Error()))
}

// === CARGAR CSS ===
func loadGlobalCSS() string {
	globalCSSPath := filepath.Join(publicDir, "styles", "global.css")
	content, err := os.ReadFile(globalCSSPath)
	if err != nil {
		return ""
	}
	return string(content)
}

func loadCSS(f string) string {
	if f == "" {
		return ""
	}
	content, err := os.ReadFile(f)
	if err != nil {
		return ""
	}
	return string(content)
}

func resolvePublicPath(requestPath string) (string, bool) {
	cleanPath := filepath.Clean("/" + requestPath)
	trimmedPath := strings.TrimPrefix(cleanPath, "/")

	publicRoot, err := filepath.Abs(publicDir)
	if err != nil {
		return "", false
	}

	resolvedPath, err := filepath.Abs(filepath.Join(publicRoot, trimmedPath))
	if err != nil {
		return "", false
	}

	relToPublic, err := filepath.Rel(publicRoot, resolvedPath)
	if err != nil {
		return "", false
	}

	if relToPublic == "." {
		return resolvedPath, true
	}

	if strings.HasPrefix(relToPublic, "..") || filepath.IsAbs(relToPublic) {
		return "", false
	}

	return resolvedPath, true
}

// === HANDLER PRINCIPAL ===
func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	path := r.URL.Path

	// Favicon
	if path == "/favicon.ico" {
		faviconPath, ok := resolvePublicPath(path)
		if !ok {
			http.Error(w, "404 - Ruta no encontrada", http.StatusNotFound)
			logRequest(r.Method, path, http.StatusNotFound, start)
			return
		}
		if _, err := os.Stat(faviconPath); err == nil {
			http.ServeFile(w, r, faviconPath)
			logRequest(r.Method, path, http.StatusOK, start)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if path != "/" && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	info, exists := routeStore.Get(path)
	if !exists {
		info, exists = routeStore.Get(path + "/index")
	}

	if !exists {
		staticPath, ok := resolvePublicPath(path)
		if ok {
			if fileInfo, err := os.Stat(staticPath); err == nil && !fileInfo.IsDir() {
				http.ServeFile(w, r, staticPath)
				logRequest(r.Method, path, http.StatusOK, start)
				return
			}
		}
		http.Error(w, "404 - Ruta no encontrada", http.StatusNotFound)
		logRequest(r.Method, path, http.StatusNotFound, start)
		return
	}

	// HTML estático
	if info.Type == "html" {
		content, err := os.ReadFile(info.File)
		if err != nil {
			http.Error(w, "Error leyendo archivo", http.StatusInternalServerError)
			logRequest(r.Method, path, http.StatusInternalServerError, start)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
		logRequest(r.Method, path, http.StatusOK, start)
		return
	}

	// JS estático
	if info.Type == "js" {
		content, err := os.ReadFile(info.File)
		if err != nil {
			http.Error(w, "Error leyendo archivo", http.StatusInternalServerError)
			logRequest(r.Method, path, http.StatusInternalServerError, start)
			return
		}
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Write(content)
		logRequest(r.Method, path, http.StatusOK, start)
		return
	}

	// JSX/TSX
	if info.Type == "page" {
		bundledCode, err := bundleJSX(info.File, info.IsTS)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(renderDevErrorHTML(path, err)))
			logRequest(r.Method, path, http.StatusInternalServerError, start)
			return
		}

		css := loadGlobalCSS()
		if info.CSSFile != "" {
			if scopedCSS := loadCSS(info.CSSFile); scopedCSS != "" {
				css += "\n" + scopedCSS
			}
		}

		html := generateHTML(bundledCode, css, info.Title)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
		logRequest(r.Method, path, http.StatusOK, start)
		return
	}

	http.Error(w, "404 - Ruta no encontrada", http.StatusNotFound)
	logRequest(r.Method, path, http.StatusNotFound, start)
}

// === LOG DE REQUESTS ===
func logRequest(method, path string, status int, start time.Time) {
	duration := time.Since(start)
	fmt.Printf("[%s] %s - %d (%v)\n", method, path, status, duration)
}

func collectWatchedFiles() map[string]time.Time {
	files := make(map[string]time.Time)

	for _, d := range []string{routesDir, apiDir, componentsDir, publicDir} {
		filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if shouldIgnore(path) {
				return nil
			}

			files[path] = info.ModTime()
			return nil
		})
	}

	return files
}

func watchedFilesChanged(previous, current map[string]time.Time) bool {
	if len(previous) != len(current) {
		return true
	}

	for path, prevModTime := range previous {
		currentModTime, ok := current[path]
		if !ok {
			return true
		}
		if !currentModTime.Equal(prevModTime) {
			return true
		}
	}

	return false
}

// === HOT-RELOAD ===
func watchChanges() {
	lastSnapshot := collectWatchedFiles()
	for {
		time.Sleep(2 * time.Second)
		currentSnapshot := collectWatchedFiles()
		if watchedFilesChanged(lastSnapshot, currentSnapshot) {
			fmt.Println("Cambios detectados!")
			bundleCache.Clear()
			routeStore.SetAll(scanAllRoutes())
			lastSnapshot = currentSnapshot
		}
	}
}

// === MAGIC INIT ===
func magicInit() bool {
	if _, err := os.Stat(routesDir); err == nil {
		return false
	}

	fmt.Println("Carpeta vacia detectada. Inicializando proyecto Zap...")

	dirs := []string{routesDir}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			fmt.Printf("error creando %s: %v\n", d, err)
			return false
		}
	}

	files := map[string]string{
		"routes/index.tsx": `export const title = "Inicio - Zap App";

export default function App() {
  const [count, setCount] = useState(0);

  return (
    <div style={{ padding: "50px" }}>
      <h1>Zap!</h1>
      <p>Tu app esta viva.</p>
      <p>Contador: {count}</p>
      <button onClick={() => setCount(count + 1)}>Incrementar</button>
    </div>
  );
}
`,
		"routes/about.tsx": `export const title = "Acerca de - Zap App";

export default function About() {
  return (
    <div style={{ padding: "50px" }}>
      <h1>Acerca de</h1>
      <a href="/">Volver al inicio</a>
    </div>
  );
}
`}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Printf("error escribiendo %s: %v\n", path, err)
			return false
		}
		fmt.Printf("  %s\n", path)
	}

	fmt.Println("Listo! Servidor iniciando...\n")
	return true
}

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
  /components/               -> Componentes UI (opcional)
  /public/styles/global.css  -> Estilos globales automáticos
  /public/                   -> Assets estáticos (opcional)
`, version)
}

func printVersion() {
	fmt.Println(version)
}

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
	fmt.Println("\nServidor listo. Ctrl+C para detener.\n")

	go watchChanges()
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(config.Port, nil); err != nil {
		fmt.Printf("error iniciando servidor en %s: %v\n", config.Port, err)
		os.Exit(1)
	}
}
