package store

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestRotateRefreshSessionRotatesWithinSingleTransaction(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, family_id, parent_session_id, replaced_by_session_id, expires_at, created_at, rotated_at, revoked_at
		FROM auth_sessions
		WHERE token_hash = $1
		FOR UPDATE
	`)).
		WithArgs("old-hash").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "family_id", "parent_session_id", "replaced_by_session_id", "expires_at", "created_at", "rotated_at", "revoked_at",
		}).AddRow("session-1", "user-1", "family-1", nil, nil, expiresAt, now, nil, nil))

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO auth_sessions (user_id, token_hash, family_id, parent_session_id, expires_at, user_agent, ip_address)
		VALUES ($1, $2, COALESCE($3::uuid, gen_random_uuid()), $4::uuid, $5, NULLIF($6, ''), NULLIF($7, '')::inet)
		RETURNING id, user_id, family_id, parent_session_id, replaced_by_session_id, expires_at, created_at, rotated_at, revoked_at
	`)).
		WithArgs("user-1", "new-hash", "family-1", "session-1", expiresAt, "test-agent", "127.0.0.1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "family_id", "parent_session_id", "replaced_by_session_id", "expires_at", "created_at", "rotated_at", "revoked_at",
		}).AddRow("session-2", "user-1", "family-1", "session-1", nil, expiresAt, now, nil, nil))

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE auth_sessions
		SET rotated_at = $2, last_used_at = $2, replaced_by_session_id = $3
		WHERE id = $1
	`)).
		WithArgs("session-1", sqlmock.AnyArg(), "session-2").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	session, err := store.RotateRefreshSession(context.Background(), RotateRefreshSessionInput{
		TokenHash:    "old-hash",
		NewTokenHash: "new-hash",
		ExpiresAt:    expiresAt,
		UserAgent:    "test-agent",
		IPAddress:    "127.0.0.1:8080",
	})
	if err != nil {
		t.Fatalf("RotateRefreshSession returned error: %v", err)
	}
	if session.ID != "session-2" {
		t.Fatalf("unexpected rotated session id: %s", session.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRotateRefreshSessionRevokesFamilyOnReuse(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now().UTC()
	rotatedAt := now.Add(-time.Minute)
	expiresAt := now.Add(24 * time.Hour)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, family_id, parent_session_id, replaced_by_session_id, expires_at, created_at, rotated_at, revoked_at
		FROM auth_sessions
		WHERE token_hash = $1
		FOR UPDATE
	`)).
		WithArgs("reused-hash").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "family_id", "parent_session_id", "replaced_by_session_id", "expires_at", "created_at", "rotated_at", "revoked_at",
		}).AddRow("session-1", "user-1", "family-1", nil, "session-2", expiresAt, now, rotatedAt, nil))

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, $2), revocation_reason = COALESCE(revocation_reason, NULLIF($3, ''))
		WHERE family_id = $1
	`)).
		WithArgs("family-1", sqlmock.AnyArg(), "refresh_token_reuse_detected").
		WillReturnResult(sqlmock.NewResult(0, 2))

	mock.ExpectCommit()

	_, err := store.RotateRefreshSession(context.Background(), RotateRefreshSessionInput{
		TokenHash:    "reused-hash",
		NewTokenHash: "ignored-hash",
		ExpiresAt:    expiresAt,
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRevokeRefreshSessionReturnsUnauthorizedForUnknownToken(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE auth_sessions
		SET revoked_at = COALESCE(revoked_at, NOW()), revocation_reason = COALESCE(revocation_reason, NULLIF($2, ''))
		WHERE token_hash = $1
	`)).
		WithArgs("missing-hash", "user_logout").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := store.RevokeRefreshSession(context.Background(), "missing-hash", "user_logout")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRotateRefreshSessionReturnsUnauthorizedWhenTokenMissing(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, family_id, parent_session_id, replaced_by_session_id, expires_at, created_at, rotated_at, revoked_at
		FROM auth_sessions
		WHERE token_hash = $1
		FOR UPDATE
	`)).
		WithArgs("missing-hash").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := store.RotateRefreshSession(context.Background(), RotateRefreshSessionInput{
		TokenHash:    "missing-hash",
		NewTokenHash: "new-hash",
		ExpiresAt:    time.Now().UTC().Add(time.Hour),
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
