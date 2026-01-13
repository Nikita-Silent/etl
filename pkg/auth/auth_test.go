package auth

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBearerAuthMiddleware_ValidToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	token := "test-token-123"
	middleware := BearerAuthMiddleware(logger, token)

	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestBearerAuthMiddleware_InvalidToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	token := "test-token-123"
	middleware := BearerAuthMiddleware(logger, token)

	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestBearerAuthMiddleware_MissingHeader(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	token := "test-token-123"
	middleware := BearerAuthMiddleware(logger, token)

	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestBearerAuthMiddleware_InvalidFormat(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	token := "test-token-123"
	middleware := BearerAuthMiddleware(logger, token)

	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat test-token-123")
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestBearerAuthMiddleware_EmptyToken_AllowsAccess(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	token := "" // Пустой токен - авторизация отключена
	middleware := BearerAuthMiddleware(logger, token)

	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// Заголовок Authorization не установлен
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 when token is empty (auth disabled), got %d", rr.Code)
	}
}
