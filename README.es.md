# Zap

**Runtime de desarrollo frontend sin configuración.**

Un runtime de desarrollo pequeño y portable para frontend moderno.

Descárgalo, ejecútalo, edita archivos y empieza a construir de inmediato. Sin Node.js, sin npm y sin configurar bundlers. Zap sirve HTML, JavaScript, JSX y TSX desde un solo ejecutable, así que un desarrollador nuevo puede probar desarrollo frontend casi igual que con un servidor estático simple, pero con componentes tipo React integrados.

Está pensado para prototipos frontend, demos, herramientas internas, apps pequeñas y aprendizaje. Si quieres bocetar una interfaz sin tener que aprender antes un toolchain pesado, Zap es el camino más corto entre una carpeta vacía y una app corriendo.

## Estado

Zap es actualmente un runtime de desarrollo pre-`1.0`. Su flujo principal ya es utilizable, pero sus convenciones públicas todavía pueden evolucionar mientras el proyecto encuentra la interfaz estable más pequeña que valga la pena mantener a largo plazo.

- Runtime exclusivo para desarrollo
- Runtime centrado en frontend; no ejecuta backend JavaScript
- Routing estático basado en archivos
- Estilos globales vía `public/styles/global.css`
- Archivos que empiezan con `_` son módulos privados, no rutas públicas

## Objetivos

- Distribuirse como un ejecutable autónomo
- Permitir desarrollo frontend con React, JSX y TypeScript
- Comenzar desde una carpeta vacía sin configuración obligatoria
- Ofrecer recarga automática y retroalimentación rápida durante el desarrollo
- Mantener una experiencia simple, transparente y predecible

## No objetivos

- Sustituir Node.js como runtime de propósito general
- Convertirse en un framework full-stack
- Gestionar despliegues, hosting o infraestructura
- Incorporar un backend de aplicaciones, base de datos o sistema de autenticación
- Competir con toolchains completos de compilación y distribución para producción
- Crear un ecosistema de paquetes que sustituya npm

Las nuevas funcionalidades deben reducir la fricción del desarrollo frontend sin volver más difícil de entender el flujo básico de `zap`. Deben ser opcionales, caber dentro del ejecutable y producir proyectos cuyo comportamiento se pueda comprender leyendo sus archivos.

## Inicio Rápido

```bash
# Crear nuevo proyecto
mkdir mi-app && cd mi-app

# Ejecutar Zap (auto-inicializa el proyecto)
zap

# Abrir http://localhost:8080
```

Con eso ya tienes un sitio inicial funcionando. Si `routes/` no existe, Zap crea los archivos iniciales faltantes en `routes/` junto con `public/styles/global.css`, sin sobrescribir archivos existentes. Si `routes/` ya existe, omite la inicialización automática. El directorio `public/` es opcional cuando no se necesitan recursos estáticos ni una hoja de estilos global.

## Instalación

Descarga el binario para tu plataforma desde [Releases](https://github.com/melendezgg/zap/releases) o compila desde el código:

```bash
go install github.com/melendezgg/zap@latest
```

## Características

- **Cero configuración** - Un solo binario, sin toolchain de Node.js
- **React 19.2** - Embebido en el binario para desarrollo offline
- **Imports de React** - Soporta `"react"`, `"react-dom"` y `"react-dom/client"`
- **Fast Refresh** - Actualiza rutas React sin perder el estado de los componentes
- **Overlay de errores** - Mantiene visible la última app funcional mientras corriges errores
- **Vista móvil** - `--host 0.0.0.0` muestra URLs de red y un código QR escaneable
- **TypeScript/JSX** - Soporte nativo vía esbuild
- **Source maps** - Source maps inline para stacks del navegador durante desarrollo
- **Rutas multi-formato** - `.tsx`, `.jsx`, `.html`, `.js`
- **Títulos dinámicos** - `export const title = "Título"`
- **Módulos privados** - Los archivos y directorios con prefijo `_` no son rutas
- **CSS global** - `public/styles/global.css` se carga automáticamente
- **Cache de bundles** - Las páginas JSX/TSX se cachean en memoria durante desarrollo

## Estructura del Proyecto

```text
mi-app/
├── routes/
│   ├── index.tsx          -> Home page (/)
│   ├── about.tsx          -> About page (/about)
│   ├── contact.html       -> HTML estático (/contact)
│   ├── _Card.tsx          -> Módulo privado reutilizable
│   └── _components/       -> Módulos privados reutilizables
│       └── Button.tsx
└── public/
    └── styles/
        └── global.css     -> Estilos globales
```

## Uso

```bash
zap                        # Iniciar servidor de desarrollo
zap --port 3000            # Puerto personalizado
zap -P 3000                # Forma corta del puerto
zap --host 0.0.0.0         # Probar en otro dispositivo con el QR mostrado
zap --help                 # Mostrar ayuda
zap -h                     # Forma corta de ayuda
zap --version              # Mostrar versión
zap -v                     # Forma corta de versión
```

## Ejemplo: `routes/index.tsx`

```tsx
import { useState } from "react";

export const title = "Home - My App";

export default function App() {
  const [count, setCount] = useState(0);

  return (
    <div>
      <h1>{title}</h1>
      <p>Count: {count}</p>
      <button onClick={() => setCount(count + 1)}>
        Increment
      </button>
    </div>
  );
}
```

## Alcance Frontend

Zap es intencionalmente un runtime de frontend. No ejecuta backend JavaScript, no instala paquetes npm, no provee rutas API y no se conecta a bases de datos. Si tu frontend necesita datos, ejecuta un servidor API/backend aparte y consúmelo desde Zap con `fetch`.

Zap solo maneja un conjunto pequeño de imports de paquetes controlados por ahora: `"react"`, `"react-dom"` y `"react-dom/client"`. Esos imports se mapean al runtime embebido de React que Zap sirve localmente, así que el código puede seguir patrones normales de React sin requerir `node_modules` ni conexión de red.

## Módulos Privados

Cualquier archivo o directorio dentro de `routes/` cuyo nombre empiece con `_` queda fuera del routing público. Sus módulos todavía se pueden importar normalmente desde páginas públicas.

```tsx
import Button from "./_Button";
import Card from "./_components/Card";

export default function App() {
  return (
    <div>
      <Button />
      <Card />
    </div>
  );
}
```

## Convenciones de Enrutamiento

| Archivo | URL |
| --- | --- |
| `routes/index.tsx` | `/` |
| `routes/about.tsx` | `/about` |
| `routes/about.jsx` | `/about` |
| `routes/about.html` | `/about` |
| `routes/blog/index.tsx` | `/blog` |
| `routes/blog/post.tsx` | `/blog/post` |
| `routes/scripts/app.js` | `/scripts/app.js` |
| `routes/scripts/index.js` | `/scripts/index.js` |

Los archivos TSX, JSX y HTML son rutas de página: se elimina su extensión y `index` representa al directorio que lo contiene. Los archivos JavaScript se sirven como recursos JavaScript y conservan `.js` en la URL; por eso `index.js` no funciona como índice de directorio.

### Conflictos de Rutas

Si varios archivos de página producen la misma URL, Zap mantiene el resultado determinístico y muestra una advertencia. La prioridad de páginas es:

```text
.tsx > .jsx > .html
```

Por ejemplo, si existen `routes/about.tsx` y `routes/about.html`, `/about` usa `routes/about.tsx` y Zap reporta que `routes/about.html` fue ignorado.

## Imports de React

Zap sirve un runtime embebido de React 19.2 y soporta imports normales desde `"react"`, `"react-dom"` y `"react-dom/client"`.

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

Zap usa [esbuild](https://esbuild.github.io/) para bundlear JSX/TSX en desarrollo. Las rutas se descubren desde `routes/`, los archivos privados con prefijo `_` quedan fuera del router público pero se pueden reutilizar localmente, los imports de React se mapean a globals del runtime embebido, y `public/styles/global.css` se inyecta automáticamente cuando existe.

Zap usa notificaciones nativas del sistema de archivos con un debounce corto (excepto `.git` y `node_modules`), limpia su cache de salida en memoria y vuelve a escanear las rutas. Las rutas JSX/TSX mantienen un contexto incremental de esbuild para reutilizar el grafo de dependencias en compilaciones posteriores. React Fast Refresh aplica las compilaciones exitosas en el sitio para conservar el estado local de los componentes. El polling se conserva como fallback cuando las notificaciones nativas no están disponibles.

Los fallos de compilación se registran con contexto estructurado en la terminal. Durante una actualización, un overlay muestra diagnósticos del compilador con rutas relativas al proyecto y rutas internas incrustadas redactadas mientras la última app exitosa sigue visible debajo. Al corregir el código, el overlay desaparece automáticamente. Los errores inesperados ajenos al compilador permanecen solo en la terminal.

Para probar desde un móvil, inicia Zap con `--host 0.0.0.0`. Zap muestra las URLs IPv4 alcanzables y un código QR para la primera. Los dispositivos deben estar en la misma red y conviene exponer el servidor de desarrollo únicamente en redes de confianza.

## Ejemplos

Ver la carpeta `examples/` para apps pequeñas de ejemplo alineadas con el flujo actual de desarrollo:

- `hello-world/` - App mínima
- `counter/` - Ejemplo con `useState`
- `multi-page/` - Múltiples rutas estáticas
- `html-only/` - Solo HTML/JS estático

## Licencia

MIT
