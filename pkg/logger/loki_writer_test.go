package logger

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogger_LokiWriterPushesLogs(t *testing.T) {
	requests := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests <- payload
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	buf := &bytes.Buffer{}
	log := New(Config{
		Output:        buf,
		Format:        "json",
		Sink:          "both",
		LokiURL:       server.URL,
		LokiBatchWait: 10 * time.Millisecond,
		LokiBatchSize: 1,
		LokiTimeout:   time.Second,
		LokiLabels:    map[string]string{"service": "etl"},
	})
	t.Cleanup(func() { _ = log.Close() })

	log.WithComponent("webhook-server").Info("hello", "log_kind", "loki_operational")

	select {
	case payload := <-requests:
		streams, ok := payload["streams"].([]any)
		if !ok || len(streams) != 1 {
			t.Fatalf("streams = %#v, want 1 stream", payload["streams"])
		}
		stream := streams[0].(map[string]any)
		labels := stream["stream"].(map[string]any)
		if labels["service"] != "etl" {
			t.Fatalf("service label = %v, want etl", labels["service"])
		}
		if labels["component"] != "webhook_server" {
			t.Fatalf("component label = %v, want webhook_server", labels["component"])
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Loki push")
	}

	if buf.Len() == 0 {
		t.Fatal("expected stdout copy in both mode")
	}
}

func TestLogger_LokiWriterUsesBasicAuth(t *testing.T) {
	authHeaders := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeaders <- r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	log := New(Config{
		Output:        &bytes.Buffer{},
		Format:        "json",
		Sink:          "loki",
		LokiURL:       server.URL,
		LokiUsername:  "kadmin",
		LokiPassword:  "Qasdasd",
		LokiBatchWait: 10 * time.Millisecond,
		LokiBatchSize: 1,
		LokiTimeout:   time.Second,
	})
	t.Cleanup(func() { _ = log.Close() })

	log.Info("hello")

	select {
	case authHeader := <-authHeaders:
		want := "Basic " + base64.StdEncoding.EncodeToString([]byte("kadmin:Qasdasd"))
		if authHeader != want {
			t.Fatalf("Authorization = %q, want %q", authHeader, want)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Loki push")
	}
}
