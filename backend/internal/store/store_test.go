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

func newMockStore(t *testing.T) (*Store, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}

	return New(db), mock, func() { _ = db.Close() }
}

func TestCanAccessProjectReturnsAllowed(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT EXISTS (
			SELECT 1
			FROM projects p
			WHERE p.id = $1
			  AND (
				p.owner_id = $2
				OR EXISTS (
					SELECT 1
					FROM tasks t
					WHERE t.project_id = p.id AND t.assignee_id = $2
				)
			  )
		)
	`)).
		WithArgs("project-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	allowed, err := store.CanAccessProject(context.Background(), "project-1", "user-1")
	if err != nil {
		t.Fatalf("CanAccessProject returned error: %v", err)
	}
	if !allowed {
		t.Fatal("expected project access to be allowed")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestGetTaskAccessReturnsRoles(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			t.id,
			t.project_id,
			(p.owner_id = $2) AS is_owner,
			(t.creator_id = $2) AS is_creator,
			(t.assignee_id = $2) AS is_assignee
		FROM tasks t
		INNER JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1
	`)).
		WithArgs("task-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "is_owner", "is_creator", "is_assignee"}).
			AddRow("task-1", "project-1", true, false, false))

	access, err := store.GetTaskAccess(context.Background(), "task-1", "user-1")
	if err != nil {
		t.Fatalf("GetTaskAccess returned error: %v", err)
	}
	if !access.IsOwner || access.IsCreator || access.IsAssignee {
		t.Fatalf("unexpected task access: %+v", access)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCreateTaskEnforcesProjectAccessInWriteQuery(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, due_date, creator_id)
		SELECT $1, NULLIF($2, ''), $3, $4, p.id, $6, $7, $8
		FROM projects p
		WHERE p.id = $5
		  AND (
			p.owner_id = $8
			OR EXISTS (
				SELECT 1
				FROM tasks existing
				WHERE existing.project_id = p.id
				  AND existing.assignee_id = $8
			)
		  )
		RETURNING id, title, COALESCE(description, ''), status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
	`)).
		WithArgs("Task", "Desc", "todo", "medium", "project-1", nil, nil, "user-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "priority", "project_id", "assignee_id", "creator_id", "due_date", "created_at", "updated_at",
		}).AddRow("task-1", "Task", "Desc", "todo", "medium", "project-1", nil, "user-1", nil, now, now))

	task, err := store.CreateTask(context.Background(), CreateTaskInput{
		ProjectID:   "project-1",
		Title:       "Task",
		Description: "Desc",
		Status:      "todo",
		Priority:    "medium",
		CreatorID:   "user-1",
	})
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}
	if task.ID != "task-1" {
		t.Fatalf("unexpected task id: %s", task.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCreateTaskReturnsForbiddenWhenActorCannotWriteProject(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO tasks (title, description, status, priority, project_id, assignee_id, due_date, creator_id)
		SELECT $1, NULLIF($2, ''), $3, $4, p.id, $6, $7, $8
		FROM projects p
		WHERE p.id = $5
		  AND (
			p.owner_id = $8
			OR EXISTS (
				SELECT 1
				FROM tasks existing
				WHERE existing.project_id = p.id
				  AND existing.assignee_id = $8
			)
		  )
		RETURNING id, title, COALESCE(description, ''), status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
	`)).
		WithArgs("Task", "", "todo", "medium", "project-1", nil, nil, "user-2").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, COALESCE(description, ''), owner_id, created_at
		FROM projects
		WHERE id = $1
	`)).
		WithArgs("project-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "owner_id", "created_at"}).
			AddRow("project-1", "Project", "", "owner-1", time.Now().UTC()))

	_, err := store.CreateTask(context.Background(), CreateTaskInput{
		ProjectID: "project-1",
		Title:     "Task",
		Status:    "todo",
		Priority:  "medium",
		CreatorID: "user-2",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUpdateTaskClearsNullableFieldsAndChecksActor(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE tasks t
		SET
			title = CASE WHEN $2 THEN $3 ELSE t.title END,
			description = CASE WHEN $4 THEN NULLIF($5, '') ELSE t.description END,
			status = CASE WHEN $6 THEN $7::task_status ELSE t.status END,
			priority = CASE WHEN $8 THEN $9::task_priority ELSE t.priority END,
			assignee_id = CASE WHEN $10 THEN $11::uuid ELSE t.assignee_id END,
			due_date = CASE WHEN $12 THEN $13::date ELSE t.due_date END,
			updated_at = NOW()
		FROM projects p
		WHERE t.project_id = p.id
		  AND t.id = $1
		  AND (
			p.owner_id = $14
			OR t.creator_id = $14
			OR t.assignee_id = $14
		  )
		RETURNING t.id, t.title, COALESCE(t.description, ''), t.status, t.priority, t.project_id, t.assignee_id, t.creator_id, t.due_date, t.created_at, t.updated_at
	`)).
		WithArgs(
			"task-1",
			false,
			nil,
			true,
			"",
			true,
			"done",
			false,
			nil,
			true,
			nil,
			true,
			nil,
			"user-1",
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "priority", "project_id", "assignee_id", "creator_id", "due_date", "created_at", "updated_at",
		}).AddRow("task-1", "Task", "", "done", "medium", "project-1", nil, "user-1", nil, now, now))

	task, err := store.UpdateTask(context.Background(), UpdateTaskInput{
		ID:      "task-1",
		ActorID: "user-1",
		Description: NullableStringPatch{
			Set:   true,
			Value: nil,
		},
		Status: StringPatch{
			Set:   true,
			Value: "done",
		},
		AssigneeID: NullableStringPatch{
			Set:   true,
			Value: nil,
		},
		DueDate: NullableDatePatch{
			Set:   true,
			Value: nil,
		},
	})
	if err != nil {
		t.Fatalf("UpdateTask returned error: %v", err)
	}
	if task.Status != "done" {
		t.Fatalf("unexpected task status: %s", task.Status)
	}
	if task.AssigneeID != nil || task.DueDate != nil {
		t.Fatalf("expected nullable fields to be cleared: %+v", task)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestUpdateTaskReturnsForbiddenWhenActorCannotMutate(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`
		UPDATE tasks t
		SET
			title = CASE WHEN $2 THEN $3 ELSE t.title END,
			description = CASE WHEN $4 THEN NULLIF($5, '') ELSE t.description END,
			status = CASE WHEN $6 THEN $7::task_status ELSE t.status END,
			priority = CASE WHEN $8 THEN $9::task_priority ELSE t.priority END,
			assignee_id = CASE WHEN $10 THEN $11::uuid ELSE t.assignee_id END,
			due_date = CASE WHEN $12 THEN $13::date ELSE t.due_date END,
			updated_at = NOW()
		FROM projects p
		WHERE t.project_id = p.id
		  AND t.id = $1
		  AND (
			p.owner_id = $14
			OR t.creator_id = $14
			OR t.assignee_id = $14
		  )
		RETURNING t.id, t.title, COALESCE(t.description, ''), t.status, t.priority, t.project_id, t.assignee_id, t.creator_id, t.due_date, t.created_at, t.updated_at
	`)).
		WithArgs(
			"task-1",
			false,
			nil,
			false,
			"",
			true,
			"done",
			false,
			nil,
			false,
			nil,
			false,
			nil,
			"user-2",
		).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, title, COALESCE(description, ''), status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`)).
		WithArgs("task-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "status", "priority", "project_id", "assignee_id", "creator_id", "due_date", "created_at", "updated_at",
		}).AddRow("task-1", "Task", "", "todo", "medium", "project-1", nil, "user-1", nil, time.Now().UTC(), time.Now().UTC()))

	_, err := store.UpdateTask(context.Background(), UpdateTaskInput{
		ID:      "task-1",
		ActorID: "user-2",
		Status: StringPatch{
			Set:   true,
			Value: "done",
		},
	})
	if err == nil {
		t.Fatal("expected forbidden error")
	}
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
