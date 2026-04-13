package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/harshpn/taskflow/internal/auth"
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

	if err := s.writeSessionAuthResponse(w, r, user, http.StatusCreated); err != nil {
		s.writeInternalError(w, r, err)
		return
	}
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

	if err := s.writeSessionAuthResponse(w, r, user, http.StatusOK); err != nil {
		s.writeInternalError(w, r, err)
		return
	}
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

	nextRefreshToken, nextRefreshTokenHash, err := auth.NewRefreshToken()
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	session, err := s.store.RotateRefreshSession(r.Context(), store.RotateRefreshSessionInput{
		TokenHash:    auth.HashRefreshToken(refreshToken),
		NewTokenHash: nextRefreshTokenHash,
		ExpiresAt:    time.Now().UTC().Add(s.refreshTokenTTL),
		UserAgent:    r.UserAgent(),
		IPAddress:    r.RemoteAddr,
	})
	if err != nil {
		if errors.Is(err, store.ErrUnauthorized) {
			s.writeUnauthorized(w, r, "invalid refresh token")
			return
		}
		s.handleStoreError(w, r, err)
		return
	}

	user, err := s.store.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	if err := s.writeAuthResponse(w, http.StatusOK, user, nextRefreshToken); err != nil {
		s.writeInternalError(w, r, err)
		return
	}
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

	err := s.store.RevokeRefreshSession(r.Context(), auth.HashRefreshToken(refreshToken), "user_logout")
	if err != nil && !errors.Is(err, store.ErrUnauthorized) {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (s *Server) writeSessionAuthResponse(w http.ResponseWriter, r *http.Request, user store.User, status int) error {
	refreshToken, refreshTokenHash, err := auth.NewRefreshToken()
	if err != nil {
		return err
	}

	if _, err := s.store.CreateRefreshSession(r.Context(), store.CreateRefreshSessionInput{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().UTC().Add(s.refreshTokenTTL),
		UserAgent: r.UserAgent(),
		IPAddress: r.RemoteAddr,
	}); err != nil {
		return err
	}

	return s.writeAuthResponse(w, status, user, refreshToken)
}

func (s *Server) writeAuthResponse(w http.ResponseWriter, status int, user store.User, refreshToken string) error {
	accessToken, err := s.tokenManager.IssueAccessToken(user.ID, user.Email)
	if err != nil {
		return err
	}

	writeJSON(w, status, map[string]any{
		"token":              accessToken,
		"access_token":       accessToken,
		"refresh_token":      refreshToken,
		"token_type":         "Bearer",
		"expires_in_seconds": int(s.tokenManager.AccessTokenTTL().Seconds()),
		"user":               user,
	})
	return nil
}
