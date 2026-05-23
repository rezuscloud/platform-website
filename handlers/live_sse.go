package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/rezuscloud/platform-website/obs"
)

// LiveSSE streams live infrastructure data via Server-Sent Events.
func LiveSSE(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	clientIP := c.IP()
	log.Printf("SSE client connected: %s", clientIP)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		log.Printf("SSE stream started for %s", clientIP)
		defer log.Printf("SSE stream ended for %s", clientIP)

		snapshotCount := 0
		errCount := 0

		if sendSnapshot(w) {
			snapshotCount++
		} else {
			errCount++
		}

		for range ticker.C {
			if sendSnapshot(w) {
				snapshotCount++
				if errCount > 0 {
					log.Printf("SSE recovered for %s after %d errors (%d snapshots sent)", clientIP, errCount, snapshotCount)
					errCount = 0
				}
			} else {
				errCount++
				if errCount == 1 {
					log.Printf("SSE first error for %s, continuing to retry", clientIP)
				}
			}
		}
	})

	return nil
}

func sendSnapshot(w *bufio.Writer) bool {
	data, err := liveClient.Fetch(context.Background())
	if err != nil {
		log.Printf("SSE fetch error: %v", err)
		fmt.Fprint(w, ": keepalive\n\n")
		w.Flush()
		return false
	}

	if len(data.Hosts) == 0 && len(data.Services) == 0 {
		data = obs.DefaultMockData()
	}

	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("SSE marshal error: %v", err)
		fmt.Fprint(w, ": keepalive\n\n")
		w.Flush()
		return false
	}

	fmt.Fprintf(w, "event: update\ndata: %s\n\n", payload)
	w.Flush()
	return true
}
