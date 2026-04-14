package main

import (
	"net/http/httptest"
	"testing"
)

func TestRequestClientIPPrefersForwardedHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/load", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.1")
	req.Header.Set("X-Real-IP", "198.51.100.10")
	req.RemoteAddr = "127.0.0.1:1234"

	if got := requestClientIP(req); got != "203.0.113.10" {
		t.Fatalf("requestClientIP() = %q, want %q", got, "203.0.113.10")
	}
}

func TestRequestClientIPFallsBackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/load", nil)
	req.RemoteAddr = "127.0.0.1:1234"

	if got := requestClientIP(req); got != "127.0.0.1" {
		t.Fatalf("requestClientIP() = %q, want %q", got, "127.0.0.1")
	}
}
