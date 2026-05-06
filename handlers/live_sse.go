package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LiveSSE streams live infrastructure data via Server-Sent Events.
func LiveSSE(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// Send initial snapshot immediately
		if !sendSnapshot(w) {
			return
		}

		for range ticker.C {
			if !sendSnapshot(w) {
				return
			}
		}
	})

	return nil
}

func sendSnapshot(w *bufio.Writer) bool {
	data, err := liveClient.Fetch(context.Background())
	if err != nil {
		log.Printf("SSE fetch error: %v", err)
		return false
	}

	log.Printf("SSE snapshot: %d categories, hasMetrics=%v", len(data.Categories), data.HasMetrics)

	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("SSE marshal error: %v", err)
		return false
	}

	fmt.Fprintf(w, "event: update\ndata: %s\n\n", payload)
	w.Flush()
	return true
}
