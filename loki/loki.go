package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client represents a Loki client for sending logs
type Client struct {
	URL          string
	BatchSize    int
	BatchWait    time.Duration
	Labels       map[string]string
	HTTPClient   *http.Client
	entriesQueue chan entry
	done         chan struct{}
}

// Config holds configuration for Loki client
type Config struct {
	URL        string            // Loki push API endpoint (e.g., "http://loki:3100/loki/api/v1/push")
	BatchSize  int               // Number of entries to batch before sending
	BatchWait  time.Duration     // Maximum time to wait before sending batch
	Labels     map[string]string // Default labels to add to all log entries
	HTTPClient *http.Client      // Custom HTTP client (optional)
}

// entry represents a log entry to be sent to Loki
type entry struct {
	Timestamp time.Time
	Message   string
	Level     string
}

// stream represents a stream of log entries with the same labels
type stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// pushRequest represents the request body for Loki's push API
type pushRequest struct {
	Streams []stream `json:"streams"`
}

// NewClient creates a new Loki client
func NewClient(config Config) *Client {
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	if config.BatchWait <= 0 {
		config.BatchWait = 1 * time.Second
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	client := &Client{
		URL:          config.URL,
		BatchSize:    config.BatchSize,
		BatchWait:    config.BatchWait,
		Labels:       config.Labels,
		HTTPClient:   config.HTTPClient,
		entriesQueue: make(chan entry, config.BatchSize*2),
		done:         make(chan struct{}),
	}

	go client.processQueue()
	return client
}

// Stop gracefully shuts down the client
func (c *Client) Stop() {
	close(c.done)
}

// Log sends a log entry to Loki
func (c *Client) Log(timestamp time.Time, level, message string) {
	select {
	case c.entriesQueue <- entry{
		Timestamp: timestamp,
		Level:     level,
		Message:   message,
	}:
	default:
		// Queue is full, log to stderr
		fmt.Printf("Loki client queue full, dropping log entry: %s\n", message)
	}
}

// processQueue batches and sends log entries to Loki
func (c *Client) processQueue() {
	ticker := time.NewTicker(c.BatchWait)
	defer ticker.Stop()

	batch := make([]entry, 0, c.BatchSize)
	for {
		select {
		case <-c.done:
			if len(batch) > 0 {
				c.sendBatch(batch)
			}
			return
		case e := <-c.entriesQueue:
			batch = append(batch, e)
			if len(batch) >= c.BatchSize {
				c.sendBatch(batch)
				batch = make([]entry, 0, c.BatchSize)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				c.sendBatch(batch)
				batch = make([]entry, 0, c.BatchSize)
			}
		}
	}
}

// sendBatch sends a batch of log entries to Loki
func (c *Client) sendBatch(entries []entry) {
	// Group entries by their level
	entriesByLevel := make(map[string][]entry)
	for _, e := range entries {
		entriesByLevel[e.Level] = append(entriesByLevel[e.Level], e)
	}

	// Create a stream for each level
	streams := make([]stream, 0, len(entriesByLevel))
	for level, levelEntries := range entriesByLevel {
		// Create labels for this stream
		labels := make(map[string]string)
		for k, v := range c.Labels {
			labels[k] = v
		}
		labels["level"] = level

		// Create values for this stream
		values := make([][]string, 0, len(levelEntries))
		for _, e := range levelEntries {
			// Convert timestamp to nanosecond precision string
			ts := fmt.Sprintf("%d", e.Timestamp.UnixNano())
			values = append(values, []string{ts, e.Message})
		}

		streams = append(streams, stream{
			Stream: labels,
			Values: values,
		})
	}

	// Create push request
	reqBody := pushRequest{Streams: streams}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error marshalling Loki push request: %v\n", err)
		return
	}

	// Send request to Loki
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Printf("Error creating Loki push request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending logs to Loki: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("Error response from Loki: %s\n", resp.Status)
	}
}
