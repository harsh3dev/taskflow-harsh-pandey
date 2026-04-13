package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/harshpn/taskflow/internal/auth"
	"github.com/harshpn/taskflow/internal/store"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

type tokenIssuer interface {
	IssueAccessToken(userID, email string) (string, error)
	AccessTokenTTL() time.Duration
}

type AuthService struct {
	store           *store.Store
	tokenIssuer     tokenIssuer
	refreshTokenTTL time.Duration
	bcryptCost      int
	clock           Clock
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type RefreshInput struct {
	RefreshToken string
	UserAgent    string
	IPAddress    string
}

type SessionMetadata struct {
	UserAgent string
	IPAddress string
}

func NewAuthService(store *store.Store, tokenIssuer tokenIssuer, refreshTokenTTL time.Duration, bcryptCost int, clock Clock) *AuthService {
	if clock == nil {
		clock = realClock{}
	}
	return &AuthService{
		store:           store,
		tokenIssuer:     tokenIssuer,
		refreshTokenTTL: refreshTokenTTL,
		bcryptCost:      bcryptCost,
		clock:           clock,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput, meta SessionMetadata) (AuthSession, map[string]string, error) {
	fields := map[string]string{}
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	password := strings.TrimSpace(input.Password)

	if name == "" {
		fields["name"] = "is required"
	}
	if email == "" {
		fields["email"] = "is required"
	}
	if len(password) < 8 {
		fields["password"] = "must be at least 8 characters"
	}
	if len(fields) > 0 {
		return AuthSession{}, fields, nil
	}

	hash, err := auth.HashPassword(input.Password, s.bcryptCost)
	if err != nil {
		return AuthSession{}, nil, err
	}

	user, err := s.store.CreateUser(ctx, store.CreateUserInput{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
	})
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			return AuthSession{}, map[string]string{"email": "is already registered"}, nil
		}
		return AuthSession{}, nil, err
	}

	session, err := s.newAuthSession(ctx, user, meta)
	return session, nil, err
}

func (s *AuthService) Login(ctx context.Context, input LoginInput, meta SessionMetadata) (AuthSession, map[string]string, error) {
	fields := map[string]string{}
	email := strings.TrimSpace(input.Email)
	password := strings.TrimSpace(input.Password)

	if email == "" {
		fields["email"] = "is required"
	}
	if password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		return AuthSession{}, fields, nil
	}

	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return AuthSession{}, nil, store.ErrUnauthorized
		}
		return AuthSession{}, nil, err
	}
	if err := auth.CheckPassword(user.Password, input.Password); err != nil {
		return AuthSession{}, nil, store.ErrUnauthorized
	}

	session, err := s.newAuthSession(ctx, user, meta)
	return session, nil, err
}

func (s *AuthService) Refresh(ctx context.Context, input RefreshInput) (AuthSession, map[string]string, error) {
	refreshToken := strings.TrimSpace(input.RefreshToken)
	if refreshToken == "" {
		return AuthSession{}, map[string]string{"refresh_token": "is required"}, nil
	}

	nextRefreshToken, nextRefreshTokenHash, err := auth.NewRefreshToken()
	if err != nil {
		return AuthSession{}, nil, err
	}

	session, err := s.store.RotateRefreshSession(ctx, store.RotateRefreshSessionInput{
		TokenHash:    auth.HashRefreshToken(refreshToken),
		NewTokenHash: nextRefreshTokenHash,
		ExpiresAt:    s.clock.Now().Add(s.refreshTokenTTL),
		UserAgent:    input.UserAgent,
		IPAddress:    input.IPAddress,
	})
	if err != nil {
		return AuthSession{}, nil, err
	}

	user, err := s.store.GetUserByID(ctx, session.UserID)
	if err != nil {
		return AuthSession{}, nil, err
	}

	tokens, err := s.issueTokens(ctx, user, nextRefreshToken)
	if err != nil {
		return AuthSession{}, nil, err
	}

	return AuthSession{
		User:   userFromStore(user),
		Tokens: tokens,
	}, nil, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return store.ErrBadRequest
	}

	err := s.store.RevokeRefreshSession(ctx, auth.HashRefreshToken(refreshToken), "user_logout")
	if err != nil && errors.Is(err, store.ErrUnauthorized) {
		return nil
	}
	return err
}

func (s *AuthService) newAuthSession(ctx context.Context, user store.User, meta SessionMetadata) (AuthSession, error) {
	refreshToken, refreshTokenHash, err := auth.NewRefreshToken()
	if err != nil {
		return AuthSession{}, err
	}

	if _, err := s.store.CreateRefreshSession(ctx, store.CreateRefreshSessionInput{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: s.clock.Now().Add(s.refreshTokenTTL),
		UserAgent: meta.UserAgent,
		IPAddress: meta.IPAddress,
	}); err != nil {
		return AuthSession{}, err
	}

	tokens, err := s.issueTokens(ctx, user, refreshToken)
	if err != nil {
		return AuthSession{}, err
	}

	return AuthSession{
		User:   userFromStore(user),
		Tokens: tokens,
	}, nil
}

func (s *AuthService) issueTokens(_ context.Context, user store.User, refreshToken string) (AuthTokens, error) {
	accessToken, err := s.tokenIssuer.IssueAccessToken(user.ID, user.Email)
	if err != nil {
		return AuthTokens{}, err
	}

	return AuthTokens{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresInSeconds: int(s.tokenIssuer.AccessTokenTTL().Seconds()),
	}, nil
}
