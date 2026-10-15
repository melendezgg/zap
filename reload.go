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
	done    chan struct{}
	once    sync.Once
}

func newReloadHub() *reloadHub {
	return &reloadHub{
		clients: make(map[chan struct{}]struct{}),
		done:    make(chan struct{}),
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

func (h *reloadHub) close() {
	h.once.Do(func() {
		close(h.done)
	})
}

var devReload = newReloadHub()

func handleDevEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is not supported", http.StatusInternalServerError)
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
		case <-devReload.done:
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

  const showError = (message) => {
    let host = document.getElementById("__zap-error-overlay");
    if (!host) {
      host = document.createElement("div");
      host.id = "__zap-error-overlay";
      host.style.cssText = "position:fixed;inset:0;z-index:2147483647;background:rgba(9,12,20,.94);color:#f8fafc;padding:32px;overflow:auto;font:14px/1.55 ui-monospace,SFMono-Regular,Menlo,monospace";
      const title = document.createElement("h1");
      title.textContent = "Zap compilation error";
      title.style.cssText = "margin:0 0 18px;font:700 22px/1.2 system-ui,sans-serif;color:#fb7185";
      const detail = document.createElement("pre");
      detail.style.cssText = "white-space:pre-wrap;word-break:break-word;margin:0;max-width:1100px";
      host.append(title, detail);
      document.body.appendChild(host);
    }
    host.querySelector("pre").textContent = message;
  };

  const hideError = () => document.getElementById("__zap-error-overlay")?.remove();

  const applyUpdate = async () => {
    const response = await fetch("/__zap/update?route=" + encodeURIComponent(window.location.pathname), {
      cache: "no-store",
    });
    const update = await response.json();
    if (!response.ok) throw new Error(update.error || "Update failed");

    let styles = document.getElementById("__zap-styles");
    if (update.css && !styles) {
      styles = document.createElement("style");
      styles.id = "__zap-styles";
      document.head.appendChild(styles);
    }
    if (styles) styles.textContent = update.css || "";

    const executableCode = update.code.replaceAll(
      "\"/__zap/assets/react/",
      "\"" + window.location.origin + "/__zap/assets/react/",
    );
    const url = URL.createObjectURL(new Blob([executableCode], { type: "text/javascript" }));
    try {
      await import(url);
      hideError();
    } finally {
      URL.revokeObjectURL(url);
    }
  };

  const source = new EventSource("/__zap/events");
  source.addEventListener("reload", async () => {
    if (!window.__zapHotUpdate) {
      window.location.reload();
      return;
    }
    try {
      await applyUpdate();
    } catch (error) {
      showError(error instanceof Error ? error.message : String(error));
    }
  });
})();
</script>`

func injectDevReloadScript(content string) string {
	index := strings.LastIndex(strings.ToLower(content), "</body>")
	if index == -1 {
		return content + "\n" + devReloadScript
	}
	return content[:index] + devReloadScript + "\n" + content[index:]
}
