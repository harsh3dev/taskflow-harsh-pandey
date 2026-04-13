package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type TokenManagerConfig struct {
	ActiveKeyID    string
	SigningKeys    map[string]string
	AccessTokenTTL time.Duration
	Issuer         string
	Audience       string
	Now            func() time.Time
}

type TokenManager struct {
	activeKeyID    string
	signingKeys    map[string][]byte
	accessTokenTTL time.Duration
	issuer         string
	audience       string
	now            func() time.Time
}

func NewTokenManager(cfg TokenManagerConfig) TokenManager {
	keys := make(map[string][]byte, len(cfg.SigningKeys))
	for keyID, secret := range cfg.SigningKeys {
		keys[keyID] = []byte(secret)
	}

	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = func() time.Time { return time.Now().UTC() }
	}

	return TokenManager{
		activeKeyID:    cfg.ActiveKeyID,
		signingKeys:    keys,
		accessTokenTTL: cfg.AccessTokenTTL,
		issuer:         cfg.Issuer,
		audience:       cfg.Audience,
		now:            nowFn,
	}
}

func (tm TokenManager) IssueAccessToken(userID, email string) (string, error) {
	now := tm.now()
	tokenID, err := randomHex(16)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Issuer:    tm.issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{tm.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.accessTokenTTL)),
		},
	})
	token.Header["kid"] = tm.activeKeyID

	signingKey, ok := tm.signingKeys[tm.activeKeyID]
	if !ok {
		return "", fmt.Errorf("missing active signing key")
	}

	return token.SignedString(signingKey)
}

func (tm TokenManager) ParseToken(raw string) (Claims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tm.issuer),
		jwt.WithAudience(tm.audience),
	)

	token, err := parser.ParseWithClaims(raw, &Claims{}, func(token *jwt.Token) (any, error) {
		keyID, _ := token.Header["kid"].(string)
		if strings.TrimSpace(keyID) == "" {
			return nil, fmt.Errorf("missing key id")
		}

		signingKey, ok := tm.signingKeys[keyID]
		if !ok {
			return nil, fmt.Errorf("unknown key id")
		}
		return signingKey, nil
	})
	if err != nil {
		return Claims{}, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}
	if strings.TrimSpace(claims.ID) == "" || strings.TrimSpace(claims.Subject) == "" {
		return Claims{}, fmt.Errorf("invalid token claims")
	}

	return *claims, nil
}

func (tm TokenManager) AccessTokenTTL() time.Duration {
	return tm.accessTokenTTL
}

func NewRefreshToken() (string, string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", "", err
	}

	token := base64.RawURLEncoding.EncodeToString(raw[:])
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomHex(bytesLen int) (string, error) {
	raw := make([]byte, bytesLen)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
