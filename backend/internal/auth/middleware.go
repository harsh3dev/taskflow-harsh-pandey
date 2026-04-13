package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userContextKey contextKey = "authenticated_user"

type AuthenticatedUser struct {
	UserID string
	Email  string
}

func UserFromContext(ctx context.Context) (AuthenticatedUser, bool) {
	user, ok := ctx.Value(userContextKey).(AuthenticatedUser)
	return user, ok
}

func Middleware(tokenManager TokenManager, unauthorized func(http.ResponseWriter, *http.Request, string)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if header == "" {
				unauthorized(w, r, "missing bearer token")
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				unauthorized(w, r, "invalid authorization header")
				return
			}

			claims, err := tokenManager.ParseToken(strings.TrimSpace(strings.TrimPrefix(header, prefix)))
			if err != nil {
				unauthorized(w, r, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, AuthenticatedUser{
				UserID: claims.UserID,
				Email:  claims.Email,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
