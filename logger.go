package main

import (
	"fmt"
	"time"
)

// === LOG DE REQUESTS ===
func logRequest(method, path string, status int, start time.Time) {
	duration := time.Since(start)
	fmt.Printf("[%s] %s - %d (%v)\n", method, path, status, duration)
}
