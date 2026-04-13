package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/harshpn/taskflow/internal/service"
	"github.com/harshpn/taskflow/internal/store"
)

type authRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	session, fields, err := s.authService.Register(r.Context(), service.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}, service.SessionMetadata{
		UserAgent: r.UserAgent(),
		IPAddress: r.RemoteAddr,
	})
	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}
	if err != nil {
		if errors.Is(err, store.ErrUnauthorized) {
			s.writeUnauthorized(w, r, "invalid credentials")
			return
		}
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, newAuthResponse(session))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	session, fields, err := s.authService.Login(r.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}, service.SessionMetadata{
		UserAgent: r.UserAgent(),
		IPAddress: r.RemoteAddr,
	})
	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}
	if err != nil {
		if errors.Is(err, store.ErrUnauthorized) {
			s.writeUnauthorized(w, r, "invalid credentials")
			return
		}
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, newAuthResponse(session))
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshTokenRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		writeValidationError(w, r, map[string]string{"refresh_token": "is required"})
		return
	}

	session, fields, err := s.authService.Refresh(r.Context(), service.RefreshInput{
		RefreshToken: refreshToken,
		UserAgent:    r.UserAgent(),
		IPAddress:    r.RemoteAddr,
	})
	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}
	if err != nil {
		if errors.Is(err, store.ErrUnauthorized) {
			s.writeUnauthorized(w, r, "invalid refresh token")
			return
		}
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, newAuthResponse(session))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req refreshTokenRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		writeValidationError(w, r, map[string]string{"refresh_token": "is required"})
		return
	}

	err := s.authService.Logout(r.Context(), refreshToken)
	if err != nil && !errors.Is(err, store.ErrUnauthorized) {
		if errors.Is(err, store.ErrBadRequest) {
			writeValidationError(w, r, map[string]string{"refresh_token": "is required"})
			return
		}
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}
