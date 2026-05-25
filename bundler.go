package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

const reactGlobalsNamespace = "zap-react-globals"

// === BUNDLE JSX/TSX ===
func bundleJSX(entryFile string, isTS bool) (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cwd error: %v", err)
	}
	absPath, err := filepath.Abs(entryFile)
	if err != nil {
		return "", fmt.Errorf("path error: %v", err)
	}

	cacheKey := fmt.Sprintf("%s|%t", absPath, isTS)
	if cachedBundle, ok := bundleCache.Get(cacheKey); ok {
		return cachedBundle, nil
	}

	virtualEntry := fmt.Sprintf("import App from '%s'; window.App = App;", absPath)

	result := api.Build(api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents:   virtualEntry,
			ResolveDir: workingDir,
			Sourcefile: "entry.js",
		},
		Write:             false,
		Format:            api.FormatIIFE,
		Target:            api.ES2020,
		Bundle:            true,
		MinifyWhitespace:  false,
		MinifyIdentifiers: false,
		MinifySyntax:      false,
		Sourcemap:         api.SourceMapNone,
		Platform:          api.PlatformBrowser,
		LegalComments:     api.LegalCommentsNone,
		TreeShaking:       api.TreeShakingTrue,
		AbsWorkingDir:     workingDir,
		Plugins:           []api.Plugin{reactGlobalsPlugin()},
	})

	if len(result.Errors) > 0 {
		msg := ""
		for _, e := range result.Errors {
			msg += e.Text + "\n"
		}
		return "", fmt.Errorf("esbuild: %s", msg)
	}

	if len(result.OutputFiles) > 0 {
		output := string(result.OutputFiles[0].Contents)
		bundleCache.Set(cacheKey, output)
		return output, nil
	}

	return "", fmt.Errorf("no output")
}

func reactGlobalsPlugin() api.Plugin {
	return api.Plugin{
		Name: "zap-react-globals",
		Setup: func(build api.PluginBuild) {
			build.OnResolve(api.OnResolveOptions{Filter: `^react$|^react-dom$|^react-dom/client$`},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					return api.OnResolveResult{
						Path:      args.Path,
						Namespace: reactGlobalsNamespace,
					}, nil
				})

			build.OnLoad(api.OnLoadOptions{Filter: `.*`, Namespace: reactGlobalsNamespace},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					contents, err := reactGlobalModule(args.Path)
					if err != nil {
						return api.OnLoadResult{}, err
					}

					return api.OnLoadResult{
						Contents: &contents,
						Loader:   api.LoaderJS,
					}, nil
				})
		},
	}
}

func reactGlobalModule(path string) (string, error) {
	switch path {
	case "react":
		return `const React = window.React;
export default React;
export const Children = React.Children;
export const Component = React.Component;
export const Fragment = React.Fragment;
export const Profiler = React.Profiler;
export const PureComponent = React.PureComponent;
export const StrictMode = React.StrictMode;
export const Suspense = React.Suspense;
export const cloneElement = React.cloneElement;
export const createContext = React.createContext;
export const createElement = React.createElement;
export const createFactory = React.createFactory;
export const createRef = React.createRef;
export const forwardRef = React.forwardRef;
export const isValidElement = React.isValidElement;
export const lazy = React.lazy;
export const memo = React.memo;
export const startTransition = React.startTransition;
export const useCallback = React.useCallback;
export const useContext = React.useContext;
export const useDebugValue = React.useDebugValue;
export const useDeferredValue = React.useDeferredValue;
export const useEffect = React.useEffect;
export const useId = React.useId;
export const useImperativeHandle = React.useImperativeHandle;
export const useInsertionEffect = React.useInsertionEffect;
export const useLayoutEffect = React.useLayoutEffect;
export const useMemo = React.useMemo;
export const useReducer = React.useReducer;
export const useRef = React.useRef;
export const useState = React.useState;
export const useSyncExternalStore = React.useSyncExternalStore;
export const useTransition = React.useTransition;
export const version = React.version;`, nil
	case "react-dom":
		return reactDOMGlobalModule(), nil
	case "react-dom/client":
		return `const ReactDOM = window.ReactDOM;
export const createRoot = ReactDOM.createRoot;
export const hydrateRoot = ReactDOM.hydrateRoot;`, nil
	default:
		return "", fmt.Errorf("unsupported React import %q", path)
	}
}

func reactDOMGlobalModule() string {
	exports := []string{
		"createPortal",
		"createRoot",
		"findDOMNode",
		"flushSync",
		"hydrate",
		"hydrateRoot",
		"render",
		"unmountComponentAtNode",
		"unstable_batchedUpdates",
		"version",
	}

	var builder strings.Builder
	builder.WriteString("const ReactDOM = window.ReactDOM;\nexport default ReactDOM;\n")
	for _, name := range exports {
		builder.WriteString("export const ")
		builder.WriteString(name)
		builder.WriteString(" = ReactDOM.")
		builder.WriteString(name)
		builder.WriteString(";\n")
	}
	return builder.String()
}
