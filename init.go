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
	if err := createStarterProject(); err != nil {
		fmt.Printf("error inicializando proyecto: %v\n", err)
		return false
	}

	if _, err := os.Stat(routesDir); err == nil {
		fmt.Println("Listo! Servidor iniciando...")
		fmt.Println()
		return true
	}
	return false
}

func createStarterProject() error {
	dirs := []string{routesDir}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("creando %s: %w", d, err)
		}
	}

	files := map[string]string{
		"routes/index.tsx": `import { useState } from "react";

export const title = "Inicio - Zap App";

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
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("escribiendo %s: %w", path, err)
		}
		fmt.Printf("  %s\n", path)
	}

	return nil
}
