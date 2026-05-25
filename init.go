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
	dirs := []string{routesDir, "public/styles"}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("creando %s: %w", d, err)
		}
	}

	files := map[string]string{
		"routes/index.tsx": `import { useState } from "react";

export const title = "Home - Zap App";

export default function App() {
  const [count, setCount] = useState(0);

  return (
    <main className="container">
      <nav>
        <a href="/">Home</a>
        <a href="/about">About</a>
      </nav>

      <h1>Zap App</h1>
      <p>
        Edit files in <code>routes/</code> and the browser will update
        automatically.
      </p>

      <p>Count: {count}</p>
      <button onClick={() => setCount(count + 1)}>Increment</button>
    </main>
  );
}
`,
		"routes/about.tsx": `export const title = "About - Zap App";

export default function About() {
  return (
    <main className="container">
      <nav>
        <a href="/">Home</a>
        <a href="/about">About</a>
      </nav>

      <h1>About</h1>
      <p>
        Zap serves HTML, JavaScript, JSX, and TSX from a single executable so
        you can sketch frontends quickly.
      </p>
    </main>
  );
}
`,
		"public/styles/global.css": `:root {
  color: #222;
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  line-height: 1.5;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
}

.container {
  margin: 0 auto;
  max-width: 720px;
  padding: 48px 24px;
}

nav {
  display: flex;
  gap: 16px;
  margin-bottom: 32px;
}

a {
  color: #0b57d0;
}

h1 {
  font-size: 40px;
  line-height: 1.1;
  margin: 0 0 16px;
}

p {
  margin: 0 0 16px;
}

button {
  font: inherit;
  padding: 8px 12px;
}

code {
  background: #f2f2f2;
  padding: 2px 4px;
}

@media (max-width: 640px) {
  .container {
    padding: 32px 18px;
  }

  h1 {
    font-size: 32px;
  }
}
`,
	}

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
