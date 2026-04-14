package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddlewarePropagatesHeaderToRequest(t *testing.T) {
	var seen string
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("X-Request-ID")
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if seen == "" {
		t.Fatal("expected request header X-Request-ID to be propagated")
	}
	if got := rec.Header().Get("X-Request-ID"); got == "" || got != seen {
		t.Fatalf("response X-Request-ID = %q, want propagated value %q", got, seen)
	}
}

func TestRequestIDMiddlewarePreservesExistingHeader(t *testing.T) {
	var seen string
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("X-Request-ID")
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	req.Header.Set("X-Request-ID", "req_existing")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if seen != "req_existing" {
		t.Fatalf("request X-Request-ID = %q, want req_existing", seen)
	}
	if got := rec.Header().Get("X-Request-ID"); got != "req_existing" {
		t.Fatalf("response X-Request-ID = %q, want req_existing", got)
	}
}
