package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

var lokiDynamicLabelKeys = []string{"component", "log_kind", "level", "operation_type"}

type lokiWriter struct {
	url          string
	tenantID     string
	username     string
	password     string
	client       *http.Client
	staticLabels map[string]string
	batchWait    time.Duration
	batchSize    int

	entries chan lokiEntry
	wg      sync.WaitGroup
	once    sync.Once
}

type lokiEntry struct {
	timestamp string
	line      string
	labels    map[string]string
}

type lokiPushRequest struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"`
}

func newLokiWriter(cfg Config) (*lokiWriter, error) {
	if strings.TrimSpace(cfg.LokiURL) == "" {
		return nil, fmt.Errorf("LOKI_URL is required when LOG_SINK is loki or both")
	}
	batchWait := cfg.LokiBatchWait
	if batchWait <= 0 {
		batchWait = time.Second
	}
	batchSize := cfg.LokiBatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	timeout := cfg.LokiTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	w := &lokiWriter{
		url:          cfg.LokiURL,
		tenantID:     cfg.LokiTenantID,
		username:     cfg.LokiUsername,
		password:     cfg.LokiPassword,
		client:       &http.Client{Timeout: timeout},
		staticLabels: cloneLabels(cfg.LokiLabels),
		batchWait:    batchWait,
		batchSize:    batchSize,
		entries:      make(chan lokiEntry, batchSize*4),
	}
	if len(w.staticLabels) == 0 {
		w.staticLabels = map[string]string{"service": "frontol-etl"}
	}
	w.wg.Add(1)
	go w.run()
	return w, nil
}

func (w *lokiWriter) Write(p []byte) (int, error) {
	entry, ok := w.buildEntry(p)
	if !ok {
		return len(p), nil
	}
	select {
	case w.entries <- entry:
	default:
		// Drop on backpressure to avoid blocking the application.
	}
	return len(p), nil
}

func (w *lokiWriter) Close() error {
	w.once.Do(func() {
		close(w.entries)
		w.wg.Wait()
	})
	return nil
}

func (w *lokiWriter) run() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.batchWait)
	defer ticker.Stop()
	batch := make([]lokiEntry, 0, w.batchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		_, _ = w.flush(batch)
		batch = batch[:0]
	}
	for {
		select {
		case entry, ok := <-w.entries:
			if !ok {
				flush()
				return
			}
			batch = append(batch, entry)
			if len(batch) >= w.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (w *lokiWriter) flush(batch []lokiEntry) (int, error) {
	streamsByKey := make(map[string]*lokiStream)
	order := make([]string, 0)
	for _, entry := range batch {
		key := labelsKey(entry.labels)
		stream, ok := streamsByKey[key]
		if !ok {
			stream = &lokiStream{Stream: entry.labels, Values: make([][2]string, 0, 1)}
			streamsByKey[key] = stream
			order = append(order, key)
		}
		stream.Values = append(stream.Values, [2]string{entry.timestamp, entry.line})
	}
	request := lokiPushRequest{Streams: make([]lokiStream, 0, len(order))}
	for _, key := range order {
		request.Streams = append(request.Streams, *streamsByKey[key])
	}
	body, err := json.Marshal(request)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest(http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if w.tenantID != "" {
		req.Header.Set("X-Scope-OrgID", w.tenantID)
	}
	if w.username != "" || w.password != "" {
		req.SetBasicAuth(w.username, w.password)
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return 0, fmt.Errorf("loki push returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return len(batch), nil
}

func (w *lokiWriter) buildEntry(p []byte) (lokiEntry, bool) {
	line := strings.TrimSpace(string(p))
	if line == "" {
		return lokiEntry{}, false
	}
	payload := make(map[string]any)
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		return lokiEntry{timestamp: fmt.Sprintf("%d", time.Now().UnixNano()), line: line, labels: cloneLabels(w.staticLabels)}, true
	}
	labels := cloneLabels(w.staticLabels)
	for _, key := range lokiDynamicLabelKeys {
		if value, ok := payload[key].(string); ok && value != "" {
			labels[key] = sanitizeLabelValue(value)
		}
	}
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	if rawTime, ok := payload["time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339Nano, rawTime); err == nil {
			timestamp = fmt.Sprintf("%d", parsed.UnixNano())
		}
	}
	return lokiEntry{timestamp: timestamp, line: line, labels: labels}, true
}

func labelsKey(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+labels[key])
	}
	return strings.Join(parts, ",")
}

func cloneLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(labels))
	for key, value := range labels {
		cloned[key] = value
	}
	return cloned
}

func sanitizeLabelValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer(" ", "_", "/", "_", ".", "_", "-", "_", ":", "_")
	return replacer.Replace(value)
}
