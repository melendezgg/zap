# Zap

**React/TypeScript development runtime without the setup.**

A small, portable development runtime for modern web development.

Download it, run it, edit files, and start building immediately. No Node.js, no npm, and no bundler setup. Zap serves HTML, JSX, and TSX from a single executable, so a new developer can try web development almost the same way they would with a simple static server, but with React-style components built in.

It is designed for prototypes, demos, internal tools, small apps, and learning. If you want to try modern frontend development without first learning a heavy toolchain, Zap is the shortest path from an empty folder to a running app.

## Status

Zap is currently focused on the `0.1` development experience.

- Dev-only runtime for now
- Static file-based routing
- Global styles via `public/styles/global.css`
- Files starting with `_` are private modules, not public routes

## Quick Start

```bash
# Create a new project
mkdir my-app && cd my-app

# Run Zap (auto-initializes project)
zap

# Open http://localhost:8080
```

That is enough to get a site running.

## Installation

Download the binary for your platform from [Releases](https://github.com/melendezgg/zap/releases) or build from source:

```bash
go install github.com/melendezgg/zap@latest
```

## Features

- **Zero configuration** - Single binary, no Node.js toolchain
- **React 18** - Loaded via CDN
- **Hot reload** - Detects file changes automatically
- **TypeScript/JSX** - Native support via esbuild
- **Multi-format routes** - `.tsx`, `.jsx`, `.html`, `.js`
- **Dynamic titles** - `export const title = "Page Title"`
- **Private route files** - `_Component.tsx`, `_utils.jsx`, `_helpers.js` are not routed
- **Global CSS** - `public/styles/global.css` is loaded automatically
- **Bundle cache** - JSX/TSX bundles are cached in memory during development

## Project Structure

```text
my-app/
├── routes/
│   ├── index.tsx          -> Home page (/)
│   ├── about.tsx          -> About page (/about)
│   ├── contact.html       -> Static HTML (/contact)
│   └── _Card.tsx          -> Private module (not a route)
├── components/
│   └── Button.jsx         -> Reusable component
└── public/
    └── styles/
        └── global.css     -> Global styles
```

## Usage

```bash
zap                        # Start the dev server
zap --port 3000            # Custom port
zap --help                 # Show help
```

## Example: `routes/index.tsx`

```tsx
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

## Private Files

Files inside `routes/` that start with `_` are excluded from public routing, but can still be imported normally.

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

## Global Styles

Create `public/styles/global.css` and Zap will inject it into every React page automatically.

```css
body {
  font-family: sans-serif;
  margin: 0;
}
```

## Use Cases

- **Teaching web development** - Students start coding immediately
- **Prototyping** - Quick experiments without scaffolding
- **Learning React** - Focus on concepts instead of tooling
- **Internal tools** - Small dashboards and utilities with minimal setup
- **Small projects** - Apps that do not need a heavy toolchain

## How It Works

Zap uses [esbuild](https://esbuild.github.io/) to bundle JSX/TSX for development. Routes are discovered from `routes/`, private files prefixed with `_` are excluded from the public router, and `public/styles/global.css` is injected automatically when present.

Zap watches the project every 2 seconds, clears its in-memory bundle cache on changes, and rebuilds the route map.

## Examples

See the `examples/` folder for small sample apps built around the current dev workflow:

- `hello-world/` - Minimal app
- `counter/` - `useState` example
- `multi-page/` - Multiple static routes
- `html-only/` - Static HTML/JS

## License

MIT
