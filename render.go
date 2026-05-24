package main

import (
	"fmt"
	"html"
)

// / CDNs - React
const reactCDN = `https://unpkg.com/react@18/umd/react.development.js`
const reactDOMCDN = `https://unpkg.com/react-dom@18/umd/react-dom.development.js`

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
