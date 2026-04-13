package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/harshpn/taskflow/internal/store"
)

func (s *Server) enforceJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodDelete && r.Method != http.MethodHead {
			r.Body = http.MaxBytesReader(w, r.Body, s.maxBodyBytes)
			if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
				writeError(w, r, http.StatusBadRequest, "invalid_content_type", "content type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set("X-Request-Id", requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		s.logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", requestIDFromContext(r.Context()),
		)
	})
}

func (s *Server) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				s.logger.Error("request panic",
					"panic", recovered,
					"path", r.URL.Path,
					"request_id", requestIDFromContext(r.Context()),
				)
				writeError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) writeUnauthorized(w http.ResponseWriter, r *http.Request, _ string) {
	writeError(w, r, http.StatusUnauthorized, "unauthorized", "unauthorized")
}

func (s *Server) writeInternalError(w http.ResponseWriter, r *http.Request, err error) {
	s.logger.Error("request failed", "error", err, "request_id", requestIDFromContext(r.Context()))
	writeError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
}

func (s *Server) handleStoreError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, r, http.StatusNotFound, "not_found", "not found")
	case errors.Is(err, store.ErrForbidden):
		writeError(w, r, http.StatusForbidden, "forbidden", "forbidden")
	case errors.Is(err, store.ErrBadRequest):
		writeError(w, r, http.StatusBadRequest, "bad_request", "bad request")
	default:
		s.writeInternalError(w, r, err)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type contextKey string

const requestIDContextKey contextKey = "request_id"

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey).(string)
	return requestID
}

func newRequestID() string {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(raw[:])
}
