package db_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func TestMigrationsEnforceCriticalIntegrityRules(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("sql.Open returned error: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	schemaName := fmt.Sprintf("taskflow_migration_%d", time.Now().UnixNano())
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA `+pq.QuoteIdentifier(schemaName)); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	defer func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA `+pq.QuoteIdentifier(schemaName)+` CASCADE`)
	}()

	if _, err := db.ExecContext(ctx, `SET search_path TO `+pq.QuoteIdentifier(schemaName)+`,public`); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	applyUpMigrations(t, ctx, db)

	var userID string
	if err := db.QueryRowContext(ctx, `
		INSERT INTO users (name, email, password)
		VALUES ('Alice', 'alice@example.com', 'password-hash')
		RETURNING id
	`).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (name, email, password)
		VALUES ('Alice 2', 'ALICE@example.com', 'password-hash')
	`)
	if err == nil {
		t.Fatal("expected case-insensitive email uniqueness to reject duplicate email")
	}

	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
		t.Fatalf("expected unique violation for lower(email), got %v", err)
	}

	var projectID string
	if err := db.QueryRowContext(ctx, `
		INSERT INTO projects (name, owner_id)
		VALUES ('Project', $1)
		RETURNING id
	`, userID).Scan(&projectID); err != nil {
		t.Fatalf("insert project: %v", err)
	}

	var taskID string
	var createdUpdatedAt time.Time
	if err := db.QueryRowContext(ctx, `
		INSERT INTO tasks (title, project_id, creator_id)
		VALUES ('Task', $1, $2)
		RETURNING id, updated_at
	`, projectID, userID).Scan(&taskID, &createdUpdatedAt); err != nil {
		t.Fatalf("insert task: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	var updatedAt time.Time
	if err := db.QueryRowContext(ctx, `
		UPDATE tasks
		SET title = 'Task updated'
		WHERE id = $1
		RETURNING updated_at
	`, taskID).Scan(&updatedAt); err != nil {
		t.Fatalf("update task: %v", err)
	}

	if !updatedAt.After(createdUpdatedAt) {
		t.Fatalf("expected updated_at trigger to advance timestamp, before=%v after=%v", createdUpdatedAt, updatedAt)
	}
}

func applyUpMigrations(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()

	files, err := filepath.Glob(filepath.Join("migrations", "*.up.sql"))
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	sort.Strings(files)

	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if _, err := db.ExecContext(ctx, string(contents)); err != nil {
			t.Fatalf("apply %s: %v", filepath.Base(file), err)
		}
	}
}
