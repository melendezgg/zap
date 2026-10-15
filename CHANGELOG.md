# Changelog

Notable changes to Zap are documented in this file. The current `Unreleased` section contains only changes verified in Git after `v0.1.4`.

## Unreleased

### Added

- Preserve React component state with embedded React Fast Refresh.
- Keep the last working UI visible behind a compilation error overlay.
- Print network URLs and a terminal QR code for mobile preview.
- Include inline source maps and structured development diagnostics.
- Shut down the development server gracefully on termination signals.

### Changed

- Prefer native filesystem notifications with a short debounce and retain polling as a fallback.
- Reuse incremental esbuild contexts and release contexts for deleted routes.
- Serve embedded React assets locally and keep development responses uncached.
- Allow intentional public runtime configuration files such as `public/config.json` while continuing to block common secret and private-key formats.

### Fixed

- Exclude underscore-prefixed directories, as well as files, from public routing while keeping them importable.
- Prevent public symlinks and traversal attempts from escaping the public directory.
- Redact absolute filesystem paths from browser-facing compilation diagnostics.
- Serialize concurrent cache misses for the same route.

### Documentation

- Add the MIT license.
- Define project goals, non-goals, routing conventions, initialization behavior, and compatibility expectations.
