# Zap

**Runtime de desarrollo React/TypeScript sin configuración.**

Un runtime de desarrollo pequeño y portable para desarrollo web moderno.

Descárgalo, ejecútalo, edita archivos y empieza a construir de inmediato. Sin Node.js, sin npm y sin configurar bundlers. Zap sirve HTML, JSX y TSX desde un solo ejecutable, así que un desarrollador nuevo puede probar desarrollo web casi igual que con un servidor estático simple, pero con componentes tipo React integrados.

Está pensado para prototipos, demos, herramientas internas, apps pequeñas y aprendizaje. Si quieres probar frontend moderno sin tener que aprender antes un toolchain pesado, Zap es el camino más corto entre una carpeta vacía y una app corriendo.

## Estado

Zap está enfocado actualmente en la experiencia de desarrollo `0.1`.

- Runtime centrado en desarrollo
- Routing estático basado en archivos
- Estilos globales vía `public/styles/global.css`
- Archivos que empiezan con `_` son módulos privados, no rutas públicas

## Inicio Rápido

```bash
# Crear nuevo proyecto
mkdir mi-app && cd mi-app

# Ejecutar Zap (auto-inicializa el proyecto)
zap

# Abrir http://localhost:8080
```

Con eso ya tienes un sitio inicial funcionando.

## Instalación

Descarga el binario para tu plataforma desde [Releases](https://github.com/melendezgg/zap/releases) o compila desde el código:

```bash
go install github.com/melendezgg/zap@latest
```

## Características

- **Cero configuración** - Un solo binario, sin toolchain de Node.js
- **React 18** - Cargado vía CDN
- **Hot reload** - Detecta cambios automáticamente
- **TypeScript/JSX** - Soporte nativo vía esbuild
- **Rutas multi-formato** - `.tsx`, `.jsx`, `.html`, `.js`
- **Títulos dinámicos** - `export const title = "Título"`
- **Archivos privados** - `_Componente.tsx`, `_utils.jsx`, `_helpers.js` no son rutas
- **CSS global** - `public/styles/global.css` se carga automáticamente
- **Cache de bundles** - Las páginas JSX/TSX se cachean en memoria durante desarrollo

## Estructura del Proyecto

```text
mi-app/
├── routes/
│   ├── index.tsx          -> Página principal (/)
│   ├── about.tsx          -> Acerca de (/about)
│   ├── contact.html       -> HTML estático (/contact)
│   └── _Card.tsx          -> Módulo privado (no es ruta)
├── components/
│   └── Button.jsx         -> Componente reutilizable
└── public/
    └── styles/
        └── global.css     -> Estilos globales
```

## Uso

```bash
zap                        # Iniciar servidor de desarrollo
zap --port 3000            # Puerto personalizado
zap --help                 # Mostrar ayuda
```

## Ejemplo: `routes/index.tsx`

```tsx
export const title = "Inicio - Mi App";

export default function App() {
  const [count, setCount] = useState(0);

  return (
    <div>
      <h1>{title}</h1>
      <p>Contador: {count}</p>
      <button onClick={() => setCount(count + 1)}>
        Incrementar
      </button>
    </div>
  );
}
```

## Archivos Privados

Los archivos dentro de `routes/` que empiezan con `_` quedan fuera del routing público, pero se pueden importar normalmente.

```tsx
import Button from "../components/Button";
import Card from "./_Card";

export default function App() {
  return (
    <div>
      <Button />
      <Card />
    </div>
  );
}
```

## Imports de React

Zap `0.1` sirve React desde CDN y espera que APIs de React como `useState` estén disponibles como globales en runtime.

Eso significa:

- todavía no importes desde `"react"` ni `"react-dom"`
- usa `useState`, `useEffect` y APIs similares directamente
- los componentes locales importados deben seguir la misma regla

Esto funciona:

```tsx
export default function App() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(count + 1)}>{count}</button>;
}
```

Esto todavía no funciona:

```tsx
import { useState } from "react";
```

## Estilos Globales

Crea `public/styles/global.css` y Zap lo inyectará automáticamente en cada página React.

```css
body {
  font-family: sans-serif;
  margin: 0;
}
```

## Casos de Uso

- **Enseñanza de desarrollo web** - Estudiantes empiezan a codificar de inmediato
- **Prototipado** - Experimentos rápidos sin scaffolding
- **Aprender React** - Enfocarse en conceptos, no en herramientas
- **Herramientas internas** - Dashboards y utilidades pequeñas con setup mínimo
- **Proyectos pequeños** - Apps que no necesitan un toolchain pesado

## Cómo Funciona

Zap usa [esbuild](https://esbuild.github.io/) para bundlear JSX/TSX en desarrollo. Las rutas se descubren desde `routes/`, los archivos privados con prefijo `_` quedan fuera del router público y `public/styles/global.css` se inyecta automáticamente cuando existe.

Zap observa cambios cada 2 segundos, limpia su cache de bundles en memoria y vuelve a escanear las rutas.

## Ejemplos

Ver la carpeta `examples/` para apps pequeñas de ejemplo alineadas con el flujo actual de desarrollo:

- `hello-world/` - App mínima
- `counter/` - Ejemplo con `useState`
- `multi-page/` - Múltiples rutas estáticas
- `html-only/` - Solo HTML/JS estático

## Licencia

MIT
