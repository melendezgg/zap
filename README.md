# Zap

**Frontend development runtime without the setup.**

A small, portable development runtime for modern frontend work.

Download it, run it, edit files, and start building immediately. No Node.js, no npm, and no bundler setup. Zap serves HTML, JavaScript, JSX, and TSX from a single executable, so a new developer can try frontend development almost the same way they would with a simple static server, but with React-style components built in.

It is designed for frontend prototypes, demos, internal tools, small apps, and learning. If you want to sketch an interface without first learning a heavy toolchain, Zap is the shortest path from an empty folder to a running app.

## Status

Zap is currently a pre-`1.0` development runtime. Its core workflow is usable, but its public conventions may still evolve while the project finds the smallest stable interface worth supporting long term.

- Development-only runtime
- Frontend-only runtime; no backend JavaScript execution
- Static file-based routing
- Global styles via `public/styles/global.css`
- Files starting with `_` are private modules, not public routes

## Goals

- Ship as a self-contained executable
- Support frontend development with React, JSX, and TypeScript
- Start from an empty directory without required configuration
- Provide automatic reloads and fast feedback during development
- Keep the workflow simple, transparent, and predictable

## Non-goals

- Replace Node.js as a general-purpose runtime
- Become a full-stack framework
- Manage deployments, hosting, or infrastructure
- Provide an application backend, database layer, or authentication system
- Compete with complete production build and distribution toolchains
- Create a package ecosystem that replaces npm

New features should reduce frontend development friction without making the basic `zap` workflow harder to understand. They should remain optional, fit inside the executable, and produce projects whose behavior can be understood by reading their files.

## Compatibility

- During `0.x`, documented behavior may still evolve as Zap establishes its smallest stable interface.
- A future `1.0` will make the documented CLI, routing conventions, starter structure, and supported React imports a stable compatibility contract.
- After `1.0`, incompatible changes must be identified explicitly and explained in the release notes.
- Security fixes and corrections to defective behavior may change incorrect behavior, but user-visible impact must be documented.
- Internal development endpoints and embedded asset URLs under `/__zap/` are implementation details, not public compatibility guarantees.

Verified user-facing changes are recorded in [CHANGELOG.md](CHANGELOG.md).

## Quick Start

```bash
# Create a new project
mkdir my-app && cd my-app

# Run Zap (auto-initializes project)
zap

# Open http://localhost:8080
```

That is enough to get a site running. If `routes/` does not exist, Zap creates the missing starter files in `routes/` plus `public/styles/global.css`, without overwriting files that already exist. If `routes/` already exists, automatic initialization is skipped. The `public/` directory is optional when no static assets or global stylesheet are needed.

## Installation

Download the archive for your platform from [Releases](https://github.com/melendezgg/zap/releases) and extract the `zap` executable (`zap.exe` on Windows), or build from source:

```bash
go install github.com/melendezgg/zap@latest
```

## Features

- **Zero configuration** - Single binary, no Node.js toolchain
- **React 19.2** - Embedded in the binary for offline development
- **React imports** - Supports `"react"`, `"react-dom"`, and `"react-dom/client"`
- **Fast Refresh** - Updates React routes without losing component state
- **Error overlay** - Keeps the last working app visible while compile errors are fixed
- **Mobile preview** - `--host 0.0.0.0` prints network URLs and a scannable QR code
- **TypeScript/JSX** - Native support via esbuild
- **Source maps** - Inline development source maps for browser stack traces
- **Multi-format routes** - `.tsx`, `.jsx`, `.html`, `.js`
- **Dynamic titles** - `export const title = "Page Title"`
- **Private route modules** - Files and directories prefixed with `_` are not routed
- **Global CSS** - `public/styles/global.css` is loaded automatically
- **Bundle cache** - JSX/TSX bundles are cached in memory during development

## Project Structure

```text
my-app/
├── routes/
│   ├── index.tsx          -> Home page (/)
│   ├── about.tsx          -> About page (/about)
│   ├── contact.html       -> Static HTML (/contact)
│   ├── _Card.tsx          -> Private reusable module
│   └── _components/       -> Private reusable modules
│       └── Button.tsx
└── public/
    └── styles/
        └── global.css     -> Global styles
```

## Usage

```bash
zap                        # Start the dev server
zap --port 3000            # Custom port
zap -P 3000                # Short port form
zap --host 0.0.0.0         # Preview on another device with the printed QR code
zap --help                 # Show help
zap -h                     # Short help form
zap --version              # Show version
zap -v                     # Short version form
```

## Example: `routes/index.tsx`

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

## Frontend Scope

Zap is intentionally a frontend runtime. It does not execute backend JavaScript, install npm packages, provide API routes, or connect to databases. If your frontend needs data, run a separate API/backend server and call it from Zap with `fetch`.

Zap only handles a small set of controlled package imports today: `"react"`, `"react-dom"`, and `"react-dom/client"`. Those imports are mapped to the embedded React runtime that Zap serves locally, so code can follow normal React patterns without requiring `node_modules` or a network connection.

## Private Modules

Any file or directory inside `routes/` whose name starts with `_` is excluded from public routing. Its modules can still be imported normally from public pages.

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

## Routing Conventions

| File | URL |
| --- | --- |
| `routes/index.tsx` | `/` |
| `routes/about.tsx` | `/about` |
| `routes/about.jsx` | `/about` |
| `routes/about.html` | `/about` |
| `routes/blog/index.tsx` | `/blog` |
| `routes/blog/post.tsx` | `/blog/post` |
| `routes/scripts/app.js` | `/scripts/app.js` |
| `routes/scripts/index.js` | `/scripts/index.js` |

TSX, JSX, and HTML files are page routes: their extension is removed, and `index` maps to its containing directory. JavaScript files are served as JavaScript resources and keep `.js` in the URL; `index.js` is therefore not a directory index.

### Route Conflicts

If multiple page files map to the same URL, Zap keeps the route deterministic and prints a warning. The page priority is:

```text
.tsx > .jsx > .html
```

For example, if both `routes/about.tsx` and `routes/about.html` exist, `/about` uses `routes/about.tsx` and Zap reports that `routes/about.html` was ignored.

## React Imports

Zap serves an embedded React 19.2 runtime and supports normal imports from `"react"`, `"react-dom"`, and `"react-dom/client"`.

You can write components with standard React imports:

```tsx
import { useState } from "react";

export default function App() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(count + 1)}>{count}</button>;
}
```

For small examples, React hooks are also available as globals:

```tsx
export default function App() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(count + 1)}>{count}</button>;
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

Zap uses [esbuild](https://esbuild.github.io/) to bundle JSX/TSX for development. Routes are discovered from `routes/`, private files prefixed with `_` are excluded from the public router but can still be reused locally, React imports are mapped to embedded runtime globals, and `public/styles/global.css` is injected automatically when present.

Zap uses native filesystem notifications with a short debounce (excluding `.git` and `node_modules`), clears its in-memory output cache on changes, and rebuilds the route map. JSX/TSX routes keep an incremental esbuild context so subsequent rebuilds can reuse the parsed dependency graph. React Fast Refresh then applies successful updates in place so local component state survives. A polling watcher is retained as a fallback when native notifications are unavailable.

Compilation failures are logged with structured context in the terminal. During hot updates, a browser overlay reports compiler diagnostics with project-relative paths and redacted embedded filesystem paths while the last successful app remains interactive underneath. Fixing the source dismisses the overlay automatically. Unexpected non-compiler errors stay in the terminal for local debugging.

For mobile preview, start Zap with `--host 0.0.0.0`. Zap prints reachable IPv4 network URLs and a QR code for the first one. Devices must be on the same network, and the development server should only be exposed on networks you trust.

## Examples

See the `examples/` folder for small sample apps built around the current dev workflow:

- `hello-world/` - Minimal app
- `counter/` - `useState` example
- `multi-page/` - Multiple static routes
- `html-only/` - Static HTML/JS

## License

MIT
