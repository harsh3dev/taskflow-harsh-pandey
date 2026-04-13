package store

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strings"
	"time"
)

var ErrUnauthorized = errors.New("unauthorized")

type RefreshSession struct {
	ID                  string
	UserID              string
	FamilyID            string
	ParentSessionID     *string
	ReplacedBySessionID *string
	ExpiresAt           time.Time
	CreatedAt           time.Time
	RotatedAt           *time.Time
	RevokedAt           *time.Time
}

type CreateRefreshSessionInput struct {
	UserID          string
	TokenHash       string
	FamilyID        *string
	ParentSessionID *string
	ExpiresAt       time.Time
	UserAgent       string
	IPAddress       string
}

type RotateRefreshSessionInput struct {
	TokenHash     string
	NewTokenHash  string
	ExpiresAt     time.Time
	UserAgent     string
	IPAddress     string
	ReuseDetected string
}

func (s *Store) CreateRefreshSession(ctx context.Context, input CreateRefreshSessionInput) (RefreshSession, error) {
	const query = `
		INSERT INTO auth_sessions (user_id, token_hash, family_id, parent_session_id, expires_at, user_agent, ip_address)
		VALUES ($1, $2, COALESCE($3::uuid, gen_random_uuid()), $4::uuid, $5, NULLIF($6, ''), NULLIF($7, '')::inet)
		RETURNING id, user_id, family_id, parent_session_id, replaced_by_session_id, expires_at, created_at, rotated_at, revoked_at
	`

	row := s.db.QueryRowContext(
		ctx,
		query,
		input.UserID,
		input.TokenHash,
		input.FamilyID,
		input.ParentSessionID,
		input.ExpiresAt,
		input.UserAgent,
		normalizeIPAddress(input.IPAddress),
	)

	session, err := scanRefreshSession(row)
	if err != nil {
		if isInvalidTextRepresentation(err) {
			return RefreshSession{}, ErrBadRequest
		}
		return RefreshSession{}, err
	}
	return session, nil
}

func (s *Store) RotateRefreshSession(ctx context.Context, input RotateRefreshSessionInput) (RefreshSession, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RefreshSession{}, err
	}
	defer func() { _ = tx.Rollback() }()

	current, err := getRefreshSessionByTokenHashForUpdate(ctx, tx, input.TokenHash)
	if err != nil {
		return RefreshSession{}, err
	}

	now := time.Now().UTC()
	if current.RevokedAt != nil {
		if revokeErr := revokeRefreshSessionFamily(ctx, tx, current.FamilyID, defaultString(input.ReuseDetected, "refresh_token_reuse_detected"), now); revokeErr != nil {
			return RefreshSession{}, revokeErr
		}
		if err := tx.Commit(); err != nil {
			return RefreshSession{}, err
		}
		return RefreshSession{}, ErrUnauthorized
	}
	if current.RotatedAt != nil || current.ReplacedBySessionID != nil {
		if revokeErr := revokeRefreshSessionFamily(ctx, tx, current.FamilyID, defaultString(input.ReuseDetected, "refresh_token_reuse_detected"), now); revokeErr != nil {
			return RefreshSession{}, revokeErr
		}
		if err := tx.Commit(); err != nil {
			return RefreshSession{}, err
		}
		return RefreshSession{}, ErrUnauthorized
	}
	if !current.ExpiresAt.After(now) {
		if _, execErr := tx.ExecContext(ctx, `
			UPDATE auth_sessions
			SET revoked_at = COALESCE(revoked_at, $2), revocation_reason = COALESCE(revocation_reason, 'expired')
			WHERE id = $1
		`, current.ID, now); execErr != nil {
			return RefreshSession{}, execErr
		}
		if err := tx.Commit(); err != nil {
			return RefreshSession{}, err
		}
		return RefreshSession{}, ErrUnauthorized
	}

	next, err := createRefreshSessionTx(ctx, tx, CreateRefreshSessionInput{
		UserID:          current.UserID,
		TokenHash:       input.NewTokenHash,
		FamilyID:        &current.FamilyID,
		ParentSessionID: &current.ID,
		ExpiresAt:       input.ExpiresAt,
		UserAgent:       input.UserAgent,
		IPAddress:       input.IPAddress,
	})
	if err != nil {
		return RefreshSession{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE auth_sessions
		SET rotated_at = $2, last_used_at = $2, replaced_by_session_id = $3
		WHERE id = $1
	`, current.ID, now, next.ID); err != nil {
		return RefreshSession{}, err
	}

	if err := tx.Commit(); err != nil {
		return RefreshSession{}, err
	}

	return next, nil
}

func (s *Store) RevokeRefreshSession(ctx context.Context, tokenHash, reason string) error {
	const query = `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, NOW()), revocation_reason = COALESCE(revocation_reason, NULLIF($2, ''))
		WHERE token_hash = $1
	`

	result, err := s.db.ExecContext(ctx, query, tokenHash, reason)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUnauthorized
	}
	return nil
}

func getRefreshSessionByTokenHashForUpdate(ctx context.Context, tx *sql.Tx, tokenHash string) (RefreshSession, error) {
	const query = `
		SELECT id, user_id, family_id, parent_session_id, replaced_by_session_id, expires_at, created_at, rotated_at, revoked_at
		FROM auth_sessions
		WHERE token_hash = $1
		FOR UPDATE
	`

	session, err := scanRefreshSession(tx.QueryRowContext(ctx, query, tokenHash))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RefreshSession{}, ErrUnauthorized
		}
		return RefreshSession{}, err
	}
	return session, nil
}

func createRefreshSessionTx(ctx context.Context, tx *sql.Tx, input CreateRefreshSessionInput) (RefreshSession, error) {
	const query = `
		INSERT INTO auth_sessions (user_id, token_hash, family_id, parent_session_id, expires_at, user_agent, ip_address)
		VALUES ($1, $2, COALESCE($3::uuid, gen_random_uuid()), $4::uuid, $5, NULLIF($6, ''), NULLIF($7, '')::inet)
		RETURNING id, user_id, family_id, parent_session_id, replaced_by_session_id, expires_at, created_at, rotated_at, revoked_at
	`

	session, err := scanRefreshSession(tx.QueryRowContext(
		ctx,
		query,
		input.UserID,
		input.TokenHash,
		input.FamilyID,
		input.ParentSessionID,
		input.ExpiresAt,
		input.UserAgent,
		normalizeIPAddress(input.IPAddress),
	))
	if err != nil {
		if isInvalidTextRepresentation(err) {
			return RefreshSession{}, ErrBadRequest
		}
		return RefreshSession{}, err
	}
	return session, nil
}

func revokeRefreshSessionFamily(ctx context.Context, tx *sql.Tx, familyID, reason string, now time.Time) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, $2), revocation_reason = COALESCE(revocation_reason, NULLIF($3, ''))
		WHERE family_id = $1
	`, familyID, now, reason)
	return err
}

func scanRefreshSession(scanner interface {
	Scan(dest ...any) error
}) (RefreshSession, error) {
	var session RefreshSession
	var parentSessionID sql.NullString
	var replacedBySessionID sql.NullString
	var rotatedAt sql.NullTime
	var revokedAt sql.NullTime

	err := scanner.Scan(
		&session.ID,
		&session.UserID,
		&session.FamilyID,
		&parentSessionID,
		&replacedBySessionID,
		&session.ExpiresAt,
		&session.CreatedAt,
		&rotatedAt,
		&revokedAt,
	)
	if err != nil {
		return RefreshSession{}, err
	}

	if parentSessionID.Valid {
		session.ParentSessionID = &parentSessionID.String
	}
	if replacedBySessionID.Valid {
		session.ReplacedBySessionID = &replacedBySessionID.String
	}
	if rotatedAt.Valid {
		timestamp := rotatedAt.Time.UTC()
		session.RotatedAt = &timestamp
	}
	if revokedAt.Valid {
		timestamp := revokedAt.Time.UTC()
		session.RevokedAt = &timestamp
	}

	return session, nil
}

func normalizeIPAddress(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	host, _, err := net.SplitHostPort(trimmed)
	if err == nil {
		trimmed = host
	}

	if ip := net.ParseIP(trimmed); ip != nil {
		return ip.String()
	}

	return ""
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
