package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/evanw/esbuild/pkg/api"
)

const reactGlobalsNamespace = "zap-react-globals"

type bundleBuild struct {
	mu      sync.Mutex
	context api.BuildContext
}

type bundleBuildStore struct {
	mu      sync.Mutex
	entries map[string]*bundleBuild
}

func newBundleBuildStore() *bundleBuildStore {
	return &bundleBuildStore{entries: make(map[string]*bundleBuild)}
}

func (s *bundleBuildStore) get(key string) *bundleBuild {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := s.entries[key]
	if entry == nil {
		entry = &bundleBuild{}
		s.entries[key] = entry
	}
	return entry
}

func (s *bundleBuildStore) acquire(key string) *bundleBuild {
	s.mu.Lock()
	entry := s.entries[key]
	if entry == nil {
		entry = &bundleBuild{}
		s.entries[key] = entry
	}
	entry.mu.Lock()
	s.mu.Unlock()
	return entry
}

func (s *bundleBuildStore) prune(activeKeys map[string]struct{}) {
	s.mu.Lock()
	var stale []*bundleBuild
	for key, entry := range s.entries {
		if _, active := activeKeys[key]; active {
			continue
		}
		entry.mu.Lock()
		delete(s.entries, key)
		stale = append(stale, entry)
	}
	s.mu.Unlock()

	for _, entry := range stale {
		if entry.context != nil {
			entry.context.Dispose()
			entry.context = nil
		}
		entry.mu.Unlock()
	}
}

func (s *bundleBuildStore) dispose() {
	s.prune(map[string]struct{}{})
}

var bundleBuilds = newBundleBuildStore()

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

	cacheKey := bundleBuildKey(absPath, isTS)
	if cachedBundle, ok := bundleCache.Get(cacheKey); ok {
		return cachedBundle, nil
	}

	// Serialize cache misses for the same route. Without this, simultaneous
	// first requests can start duplicate esbuild jobs.
	build := bundleBuilds.acquire(cacheKey)
	defer build.mu.Unlock()

	if cachedBundle, ok := bundleCache.Get(cacheKey); ok {
		return cachedBundle, nil
	}

	// Use a JavaScript string literal instead of interpolating the path between
	// quotes. This keeps projects with quotes in their names working and avoids
	// treating Windows path separators as escape sequences.
	entryPath := filepath.ToSlash(absPath)
	if relativePath, relErr := filepath.Rel(workingDir, absPath); relErr == nil && !pathEscapesRoot(relativePath) {
		entryPath = filepath.ToSlash(relativePath)
		if !strings.HasPrefix(entryPath, "./") {
			entryPath = "./" + entryPath
		}
	}
	virtualEntry := fmt.Sprintf(
		"import App from %s; window.App = App; window.__zapRefreshID = %s;",
		strconv.Quote(entryPath),
		strconv.Quote("zap:"+filepath.ToSlash(absPath)),
	)

	options := api.BuildOptions{
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
		Sourcemap:         api.SourceMapInline,
		SourcesContent:    api.SourcesContentInclude,
		Platform:          api.PlatformBrowser,
		LegalComments:     api.LegalCommentsNone,
		TreeShaking:       api.TreeShakingTrue,
		AbsWorkingDir:     workingDir,
		Plugins:           []api.Plugin{reactGlobalsPlugin()},
		Banner: map[string]string{
			"js": reactRuntimeImportScript(),
		},
		Footer: map[string]string{
			"js": mountAppScript,
		},
	}

	if build.context == nil {
		buildContext, contextErr := api.Context(options)
		if contextErr != nil {
			return "", newCompileError("context", contextErr.Errors)
		}
		build.context = buildContext
	}

	// Rebuild reuses esbuild's parsed dependency graph from the prior build.
	result := build.context.Rebuild()

	if len(result.Errors) > 0 {
		return "", newCompileError("build", result.Errors)
	}

	if len(result.OutputFiles) > 0 {
		output := string(result.OutputFiles[0].Contents)
		bundleCache.Set(cacheKey, output)
		return output, nil
	}

	return "", fmt.Errorf("no output")
}

func bundleBuildKey(absPath string, isTS bool) string {
	return fmt.Sprintf("%s|%t", absPath, isTS)
}

func activeBundleBuildKeys(routes map[string]*RouteInfo) map[string]struct{} {
	keys := make(map[string]struct{})
	for _, info := range routes {
		if info.Type != "page" {
			continue
		}
		absPath, err := filepath.Abs(info.File)
		if err != nil {
			continue
		}
		keys[bundleBuildKey(absPath, info.IsTS)] = struct{}{}
	}
	return keys
}

type compileError struct {
	Phase    string
	Messages []api.Message
}

func newCompileError(phase string, messages []api.Message) *compileError {
	return &compileError{Phase: phase, Messages: append([]api.Message(nil), messages...)}
}

func (e *compileError) Error() string {
	return fmt.Sprintf("esbuild %s:\n%s", e.Phase, formatBuildMessages(e.Messages))
}

func (e *compileError) browserMessage() string {
	messages := make([]api.Message, len(e.Messages))
	copy(messages, e.Messages)
	for i := range messages {
		messages[i].Text = sanitizeDiagnosticText(messages[i].Text)
		if messages[i].Location == nil {
			continue
		}
		location := *messages[i].Location
		location.File = safeDisplayPath(location.File)
		messages[i].Location = &location
	}
	return fmt.Sprintf("esbuild %s:\n%s", e.Phase, formatBuildMessages(messages))
}

func safeDisplayPath(path string) string {
	if path == "" {
		return ""
	}
	if !filepath.IsAbs(path) {
		if pathEscapesRoot(filepath.FromSlash(path)) {
			return filepath.Base(filepath.FromSlash(path))
		}
		return filepath.ToSlash(path)
	}

	workingDir, err := os.Getwd()
	if err == nil {
		if relative, relErr := filepath.Rel(workingDir, path); relErr == nil && !pathEscapesRoot(relative) {
			return filepath.ToSlash(relative)
		}
	}
	return filepath.Base(path)
}

var quotedAbsolutePathPattern = regexp.MustCompile("([\"'])(/[^\\r\\n\"']+|[A-Za-z]:[\\\\/][^\\r\\n\"']+)([\"'])")
var unquotedAbsolutePathPattern = regexp.MustCompile("(^|[\\s(])(/[^\\s\"'<>:,;)}\\]]+|[A-Za-z]:[\\\\/][^\\s\"'<>:,;)}\\]]+)")

func sanitizeDiagnosticText(text string) string {
	text = quotedAbsolutePathPattern.ReplaceAllStringFunc(text, func(match string) string {
		quote := match[:1]
		return quote + diagnosticPathBase(match[1:len(match)-1]) + quote
	})
	return unquotedAbsolutePathPattern.ReplaceAllStringFunc(text, func(match string) string {
		prefix := ""
		candidate := match
		if strings.ContainsRune(" \t\n\r(", rune(match[0])) {
			prefix = match[:1]
			candidate = match[1:]
		}
		return prefix + diagnosticPathBase(candidate)
	})
}

func diagnosticPathBase(value string) string {
	normalized := strings.ReplaceAll(value, "\\", "/")
	return path.Base(normalized)
}

func formatBuildMessages(messages []api.Message) string {
	var builder strings.Builder
	for i, message := range messages {
		if i > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(formatBuildMessage(message))
	}
	return builder.String()
}

func formatBuildMessage(message api.Message) string {
	var builder strings.Builder
	if message.Location != nil {
		location := message.Location
		if location.File != "" {
			builder.WriteString(location.File)
			if location.Line > 0 {
				builder.WriteString(fmt.Sprintf(":%d", location.Line))
				if location.Column > 0 {
					builder.WriteString(fmt.Sprintf(":%d", location.Column+1))
				}
			}
			builder.WriteString(": ")
		}
	}

	builder.WriteString(message.Text)

	if message.Location != nil && message.Location.LineText != "" {
		location := message.Location
		builder.WriteString("\n\n")
		builder.WriteString(location.LineText)
		if location.Column >= 0 {
			builder.WriteString("\n")
			builder.WriteString(strings.Repeat(" ", location.Column))
			length := location.Length
			if length < 1 {
				length = 1
			}
			builder.WriteString(strings.Repeat("^", length))
		}
	}

	return builder.String()
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
export const Activity = React.Activity;
export const Children = React.Children;
export const Component = React.Component;
export const Fragment = React.Fragment;
export const Profiler = React.Profiler;
export const PureComponent = React.PureComponent;
export const StrictMode = React.StrictMode;
export const Suspense = React.Suspense;
export const act = React.act;
export const cache = React.cache;
export const cacheSignal = React.cacheSignal;
export const captureOwnerStack = React.captureOwnerStack;
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
export const use = React.use;
export const useActionState = React.useActionState;
export const useCallback = React.useCallback;
export const useContext = React.useContext;
export const useDebugValue = React.useDebugValue;
export const useDeferredValue = React.useDeferredValue;
export const useEffect = React.useEffect;
export const useEffectEvent = React.useEffectEvent;
export const useId = React.useId;
export const useImperativeHandle = React.useImperativeHandle;
export const useInsertionEffect = React.useInsertionEffect;
export const useLayoutEffect = React.useLayoutEffect;
export const useMemo = React.useMemo;
export const useOptimistic = React.useOptimistic;
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
		"preconnect",
		"prefetchDNS",
		"preinit",
		"preinitModule",
		"preload",
		"preloadModule",
		"render",
		"requestFormReset",
		"unmountComponentAtNode",
		"unstable_batchedUpdates",
		"useFormState",
		"useFormStatus",
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
