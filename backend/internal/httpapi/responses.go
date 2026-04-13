package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func decodeJSON(w http.ResponseWriter, r *http.Request, dest any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		var maxBytesErr *http.MaxBytesError
		switch {
		case errors.As(err, &maxBytesErr):
			writeError(w, r, http.StatusRequestEntityTooLarge, "request_too_large", "request body is too large")
		default:
			writeError(w, r, http.StatusBadRequest, "invalid_json", "invalid JSON body")
		}
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != nil {
		if !errors.Is(err, io.EOF) {
			writeError(w, r, http.StatusBadRequest, "invalid_json", "invalid JSON body")
			return false
		}
	}
	if decoder.More() {
		writeError(w, r, http.StatusBadRequest, "invalid_json", "invalid JSON body")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"code":       code,
		"error":      message,
		"request_id": requestIDFromContext(r.Context()),
	})
}

func writeValidationError(w http.ResponseWriter, r *http.Request, fields map[string]string) {
	writeJSON(w, http.StatusBadRequest, map[string]any{
		"code":       "validation_failed",
		"error":      "validation failed",
		"fields":     fields,
		"request_id": requestIDFromContext(r.Context()),
	})
}
