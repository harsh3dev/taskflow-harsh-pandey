package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/harshpn/taskflow/internal/auth"
	"github.com/harshpn/taskflow/internal/store"
)

type authRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	fields := map[string]string{}
	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "is required"
	}
	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "is required"
	}
	if len(strings.TrimSpace(req.Password)) < 8 {
		fields["password"] = "must be at least 8 characters"
	}
	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}

	hash, err := auth.HashPassword(req.Password, s.bcryptCost)
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	user, err := s.store.CreateUser(r.Context(), store.CreateUserInput{
		Name:         strings.TrimSpace(req.Name),
		Email:        strings.TrimSpace(req.Email),
		PasswordHash: hash,
	})
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeValidationError(w, r, map[string]string{"email": "is already registered"})
			return
		}
		s.writeInternalError(w, r, err)
		return
	}

	token, err := s.tokenManager.IssueToken(user.ID, user.Email)
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user":  user,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	fields := map[string]string{}
	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "is required"
	}
	if strings.TrimSpace(req.Password) == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}

	user, err := s.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			s.writeUnauthorized(w, r, "invalid credentials")
			return
		}
		s.writeInternalError(w, r, err)
		return
	}

	if err := auth.CheckPassword(user.Password, req.Password); err != nil {
		s.writeUnauthorized(w, r, "invalid credentials")
		return
	}

	token, err := s.tokenManager.IssueToken(user.ID, user.Email)
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}
