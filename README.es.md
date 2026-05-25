# Zap

**Runtime de desarrollo frontend sin configuración.**

Un runtime de desarrollo pequeño y portable para frontend moderno.

Descárgalo, ejecútalo, edita archivos y empieza a construir de inmediato. Sin Node.js, sin npm y sin configurar bundlers. Zap sirve HTML, JavaScript, JSX y TSX desde un solo ejecutable, así que un desarrollador nuevo puede probar desarrollo frontend casi igual que con un servidor estático simple, pero con componentes tipo React integrados.

Está pensado para prototipos frontend, demos, herramientas internas, apps pequeñas y aprendizaje. Si quieres bocetar una interfaz sin tener que aprender antes un toolchain pesado, Zap es el camino más corto entre una carpeta vacía y una app corriendo.

## Estado

Zap está enfocado actualmente en la experiencia de desarrollo `0.1`.

- Runtime centrado en desarrollo
- Runtime centrado en frontend; no ejecuta backend JavaScript
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

Con eso ya tienes un sitio inicial funcionando. Si `routes/` todavía no existe, Zap crea las rutas iniciales e inicia el servidor de desarrollo de inmediato.

## Instalación

Descarga el binario para tu plataforma desde [Releases](https://github.com/melendezgg/zap/releases) o compila desde el código:

```bash
go install github.com/melendezgg/zap@latest
```

## Características

- **Cero configuración** - Un solo binario, sin toolchain de Node.js
- **React 18** - Cargado vía CDN
- **Imports de React** - Soporta `"react"`, `"react-dom"` y `"react-dom/client"`
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
│   └── _Card.tsx          -> Módulo privado reutilizable
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
import { useState } from "react";

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

## Alcance Frontend

Zap es intencionalmente un runtime de frontend. No ejecuta backend JavaScript, no instala paquetes npm, no provee rutas API y no se conecta a bases de datos. Si tu frontend necesita datos, ejecuta un servidor API/backend aparte y consúmelo desde Zap con `fetch`.

Zap solo maneja un conjunto pequeño de imports de paquetes controlados por ahora: `"react"`, `"react-dom"` y `"react-dom/client"`. Esos imports se mapean a los scripts CDN de React que Zap inyecta en runtime, así que el código puede seguir patrones normales de React sin requerir `node_modules`.

## Archivos Privados

Los archivos dentro de `routes/` que empiezan con `_` quedan fuera del routing público, pero se pueden importar normalmente.

```tsx
import Button from "./_Button";
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

## Conflictos de Rutas

Si varios archivos producen la misma ruta, Zap mantiene el resultado determinístico y muestra una advertencia. La prioridad actual es:

```text
.tsx > .jsx > .html > .js
```

Por ejemplo, si existen `routes/about.tsx` y `routes/about.html`, `/about` usa `routes/about.tsx` y Zap reporta que `routes/about.html` fue ignorado.

## Imports de React

Zap `0.1` sirve React desde CDN y soporta imports normales desde `"react"`, `"react-dom"` y `"react-dom/client"`.

Puedes escribir componentes con imports estándar de React:

```tsx
import { useState } from "react";

export default function App() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(count + 1)}>{count}</button>;
}
```

Para ejemplos pequeños, los hooks de React también están disponibles como globales:

```tsx
export default function App() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(count + 1)}>{count}</button>;
}
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

Zap usa [esbuild](https://esbuild.github.io/) para bundlear JSX/TSX en desarrollo. Las rutas se descubren desde `routes/`, los archivos privados con prefijo `_` quedan fuera del router público pero se pueden reutilizar localmente, los imports de React se mapean a globals del CDN, y `public/styles/global.css` se inyecta automáticamente cuando existe.

Zap observa cambios cada 2 segundos, limpia su cache de bundles en memoria y vuelve a escanear las rutas.

## Ejemplos

Ver la carpeta `examples/` para apps pequeñas de ejemplo alineadas con el flujo actual de desarrollo:

- `hello-world/` - App mínima
- `counter/` - Ejemplo con `useState`
- `multi-page/` - Múltiples rutas estáticas
- `html-only/` - Solo HTML/JS estático

## Licencia

MIT
