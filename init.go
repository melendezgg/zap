package main

import (
	"fmt"
	"os"
)

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

	fmt.Println("Listo! Servidor iniciando...")
	fmt.Println()
	return true
}
