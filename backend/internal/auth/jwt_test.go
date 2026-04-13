package auth

import (
	"testing"
	"time"
)

func TestTokenManagerIssuesAndParsesAccessTokenWithRequiredClaims(t *testing.T) {
	now := time.Now().UTC()
	manager := NewTokenManager(TokenManagerConfig{
		ActiveKeyID:    "current",
		SigningKeys:    map[string]string{"current": "12345678901234567890123456789012"},
		AccessTokenTTL: 15 * time.Minute,
		Issuer:         "taskflow",
		Audience:       "taskflow-api",
		Now:            func() time.Time { return now },
	})

	token, err := manager.IssueAccessToken("user-1", "test@example.com")
	if err != nil {
		t.Fatalf("IssueAccessToken returned error: %v", err)
	}

	claims, err := manager.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}

	if claims.UserID != "user-1" || claims.Email != "test@example.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.Issuer != "taskflow" {
		t.Fatalf("unexpected issuer: %s", claims.Issuer)
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != "taskflow-api" {
		t.Fatalf("unexpected audience: %v", claims.Audience)
	}
	if claims.ID == "" {
		t.Fatal("expected jti to be set")
	}
}

func TestTokenManagerSupportsKeyRotationForValidation(t *testing.T) {
	issuer := "taskflow"
	audience := "taskflow-api"
	now := time.Now().UTC()

	oldManager := NewTokenManager(TokenManagerConfig{
		ActiveKeyID:    "old",
		SigningKeys:    map[string]string{"old": "12345678901234567890123456789012"},
		AccessTokenTTL: time.Hour,
		Issuer:         issuer,
		Audience:       audience,
		Now:            func() time.Time { return now },
	})
	token, err := oldManager.IssueAccessToken("user-1", "test@example.com")
	if err != nil {
		t.Fatalf("IssueAccessToken returned error: %v", err)
	}

	rotatedManager := NewTokenManager(TokenManagerConfig{
		ActiveKeyID: "new",
		SigningKeys: map[string]string{
			"old": "12345678901234567890123456789012",
			"new": "abcdefghijklmnopqrstuvwxyz123456",
		},
		AccessTokenTTL: time.Hour,
		Issuer:         issuer,
		Audience:       audience,
		Now:            func() time.Time { return now },
	})

	claims, err := rotatedManager.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken returned error: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Fatalf("unexpected user id: %s", claims.UserID)
	}
}
