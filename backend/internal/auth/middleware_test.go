package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareRejectsMissingBearerToken(t *testing.T) {
	tokenManager := NewTokenManager("12345678901234567890123456789012", time.Hour)
	called := false

	handler := Middleware(tokenManager, func(w http.ResponseWriter, _ *http.Request, reason string) {
		called = true
		if reason == "" {
			t.Fatal("expected unauthorized reason")
		}
		w.WriteHeader(http.StatusUnauthorized)
	})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected unauthorized callback to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
}

func TestMiddlewareInjectsAuthenticatedUser(t *testing.T) {
	tokenManager := NewTokenManager("12345678901234567890123456789012", time.Hour)
	token, err := tokenManager.IssueToken("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("IssueToken returned error: %v", err)
	}

	handler := Middleware(tokenManager, func(w http.ResponseWriter, _ *http.Request, _ string) {
		w.WriteHeader(http.StatusUnauthorized)
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			t.Fatal("expected authenticated user in request context")
		}
		if user.UserID != "user-123" || user.Email != "test@example.com" {
			t.Fatalf("unexpected authenticated user: %+v", user)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
}
