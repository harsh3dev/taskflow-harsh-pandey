package httpapi

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/harshpn/taskflow/internal/auth"
)

func newTestServer(t *testing.T, maxBodyBytes int64) *Server {
	t.Helper()

	return NewServer(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		TokenManager: auth.NewTokenManager(auth.TokenManagerConfig{
			ActiveKeyID:    "default",
			SigningKeys:    map[string]string{"default": "12345678901234567890123456789012"},
			AccessTokenTTL: time.Hour,
			Issuer:         "taskflow",
			Audience:       "taskflow-api",
		}),
		RefreshTokenTTL:     24 * time.Hour,
		BcryptCost:          12,
		MaxRequestBodyBytes: maxBodyBytes,
	})
}

func TestRoutesRejectProtectedRequestWithoutBearerToken(t *testing.T) {
	server := newTestServer(t, 1024)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rec := httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
	if got := rec.Header().Get("X-Request-Id"); got == "" {
		t.Fatal("expected X-Request-Id header to be set")
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"unauthorized"`)) {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestRoutesRejectInvalidContentType(t *testing.T) {
	server := newTestServer(t, 1024)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid_content_type"`)) {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestRoutesReturnValidationErrorsForBadRegisterPayload(t *testing.T) {
	server := newTestServer(t, 1024)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"email":"","password":"short"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"validation_failed"`)) {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"name":"is required"`)) {
		t.Fatalf("expected field validation error, got: %s", rec.Body.String())
	}
}

func TestRoutesRejectOversizedJSONBodies(t *testing.T) {
	server := newTestServer(t, 16)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"name":"avery","email":"test@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"request_too_large"`)) {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}
