package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type reloadHub struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

func newReloadHub() *reloadHub {
	return &reloadHub{
		clients: make(map[chan struct{}]struct{}),
	}
}

func (h *reloadHub) subscribe() (chan struct{}, func()) {
	ch := make(chan struct{}, 1)

	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		delete(h.clients, ch)
		close(ch)
		h.mu.Unlock()
	}
}

func (h *reloadHub) broadcast() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for ch := range h.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

var devReload = newReloadHub()

func handleDevEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming no soportado", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	events, unsubscribe := devReload.subscribe()
	defer unsubscribe()

	fmt.Fprint(w, "event: ready\ndata: connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-events:
			fmt.Fprint(w, "event: reload\ndata: changed\n\n")
			flusher.Flush()
		}
	}
}

const devReloadScript = `<script>
(() => {
  if (!("EventSource" in window)) return;
  const source = new EventSource("/__zap/events");
  source.addEventListener("reload", () => window.location.reload());
})();
</script>`

func injectDevReloadScript(content string) string {
	index := strings.LastIndex(strings.ToLower(content), "</body>")
	if index == -1 {
		return content + "\n" + devReloadScript
	}
	return content[:index] + devReloadScript + "\n" + content[index:]
}
