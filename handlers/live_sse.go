package handlers

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rezuscloud/platform-website/obs"
)

// LiveSSE streams live infrastructure data via Server-Sent Events.
func LiveSSE(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		data, err := liveClient.Fetch(c.Context())
		if err != nil {
			sseError(w, err)
			return
		}
		sseNodes(w, data)
		sseMetrics(w, data)
		sseHeartbeat(w)
		w.Flush()

		for {
			select {
			case <-ticker.C:
				data, err := liveClient.Fetch(c.Context())
				if err != nil {
					sseError(w, err)
					return
				}
				sseNodes(w, data)
				sseMetrics(w, data)
				sseHeartbeat(w)
				w.Flush()
			}
		}
	})

	return nil
}

func sseError(w *bufio.Writer, err error) {
	fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
	w.Flush()
}

func sseNodes(w *bufio.Writer, data obs.LiveData) {
	fmt.Fprintf(w, "event: services\ndata: %d\n\n", len(data.Services))
}

func sseMetrics(w *bufio.Writer, data obs.LiveData) {
	var count int
	for _, s := range data.Services {
		count += len(s.Metrics)
	}
	fmt.Fprintf(w, "event: metrics\ndata: %d\n\n", count)
}

func sseHeartbeat(w *bufio.Writer) {
	fmt.Fprintf(w, "event: heartbeat\ndata: %d\n\n", time.Now().Unix())
}
