package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHotUpdateFlowFromFileChangeToRebuiltBundle(t *testing.T) {
	withTempProject(t, map[string]string{
		"routes/index.tsx": `
import Card from "./_components/Card";
export default function App() { return <main><Card /></main>; }
`,
		"routes/_components/Card.tsx": `export default function Card() { return <p>Before</p>; }`,
	})
	bundleCache.Clear()
	bundleBuilds.dispose()
	t.Cleanup(bundleBuilds.dispose)
	routeStore.SetAll(scanAllRoutes())

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	eventsRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/__zap/events", nil)
	if err != nil {
		t.Fatalf("create events request: %v", err)
	}
	eventsResponse, err := server.Client().Do(eventsRequest)
	if err != nil {
		t.Fatalf("connect to events: %v", err)
	}
	defer eventsResponse.Body.Close()
	if eventsResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected SSE connection, got %d", eventsResponse.StatusCode)
	}
	if status := requestStatus(t, server.Client(), server.URL+"/_components/Card"); status != http.StatusNotFound {
		t.Fatalf("expected private component route to return 404, got %d", status)
	}

	eventReader := bufio.NewReader(eventsResponse.Body)
	if event, err := readSSEEvent(eventReader); err != nil || event != "ready" {
		t.Fatalf("expected ready event, got event=%q err=%v", event, err)
	}

	watcher, err := newProjectWatcher(".")
	if err != nil {
		t.Fatalf("create project watcher: %v", err)
	}
	stop := make(chan struct{})
	watcherDone := make(chan error, 1)
	go func() {
		watcherDone <- runProjectWatcher(watcher, stop, reloadProject)
	}()
	watcherStopped := false
	defer func() {
		if !watcherStopped {
			close(stop)
		}
		watcher.Close()
	}()

	updatedSource := `export default function Card() { return <p>After</p>; }`
	if err := os.WriteFile(filepath.Join("routes", "_components", "Card.tsx"), []byte(updatedSource), 0644); err != nil {
		t.Fatalf("update private component: %v", err)
	}

	if event, err := readSSEEvent(eventReader); err != nil || event != "reload" {
		t.Fatalf("expected reload event, got event=%q err=%v", event, err)
	}

	updateResponse, err := server.Client().Get(server.URL + "/__zap/update?route=/")
	if err != nil {
		t.Fatalf("request rebuilt bundle: %v", err)
	}
	defer updateResponse.Body.Close()
	if updateResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected rebuilt bundle, got %d", updateResponse.StatusCode)
	}
	var update devUpdateResponse
	if err := json.NewDecoder(updateResponse.Body).Decode(&update); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if !strings.Contains(update.Code, "After") {
		t.Fatalf("expected updated private component in bundle")
	}

	writeTestFile(t, filepath.Join(routesDir, "about.tsx"), `export default function About() { return <main>About</main>; }`)
	if event, err := readSSEEvent(eventReader); err != nil || event != "reload" {
		t.Fatalf("expected reload event after adding route, got event=%q err=%v", event, err)
	}
	if status := requestStatus(t, server.Client(), server.URL+"/about"); status != http.StatusOK {
		t.Fatalf("expected added route to return 200, got %d", status)
	}

	if err := os.Remove(filepath.Join(routesDir, "about.tsx")); err != nil {
		t.Fatalf("remove route: %v", err)
	}
	if event, err := readSSEEvent(eventReader); err != nil || event != "reload" {
		t.Fatalf("expected reload event after removing route, got event=%q err=%v", event, err)
	}
	if status := requestStatus(t, server.Client(), server.URL+"/about"); status != http.StatusNotFound {
		t.Fatalf("expected removed route to return 404, got %d", status)
	}

	close(stop)
	watcherStopped = true
	if err := <-watcherDone; err != nil {
		t.Fatalf("stop project watcher: %v", err)
	}
}

func requestStatus(t *testing.T, client *http.Client, url string) int {
	t.Helper()
	response, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer response.Body.Close()
	return response.StatusCode
}

func readSSEEvent(reader *bufio.Reader) (string, error) {
	event := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			if event != "" {
				return event, nil
			}
			continue
		}
		if strings.HasPrefix(line, "event:") {
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.Contains(line, ":") {
			return "", fmt.Errorf("malformed SSE line %q", line)
		}
	}
}
