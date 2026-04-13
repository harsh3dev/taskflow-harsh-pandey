package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrForbidden  = errors.New("forbidden")
	ErrBadRequest = errors.New("bad request")
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	ProjectID   string     `json:"project_id"`
	AssigneeID  *string    `json:"assignee_id"`
	CreatorID   string     `json:"creator_id"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ProjectWithTasks struct {
	Project Project `json:"project"`
	Tasks   []Task  `json:"tasks"`
}

type CreateUserInput struct {
	Name         string
	Email        string
	PasswordHash string
}

func (s *Store) ListUsers(ctx context.Context, search string) ([]User, error) {
	base := `
		SELECT id, name, email, '' AS password, created_at
		FROM users
	`

	args := []any{}
	search = strings.TrimSpace(search)
	if search != "" {
		base += `
			WHERE name ILIKE $1 OR email ILIKE $1
		`
		args = append(args, "%"+search+"%")
	}

	base += `
		ORDER BY name ASC, email ASC
		LIMIT 50
	`

	rows, err := s.db.QueryContext(ctx, base, args...)
	if err != nil {
		if isInvalidTextRepresentation(err) {
			return nil, ErrBadRequest
		}
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (s *Store) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	const query = `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password, created_at
	`
	var user User
	err := s.db.QueryRowContext(ctx, query, input.Name, strings.ToLower(input.Email), input.PasswordHash).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrConflict
		}
		return User{}, err
	}
	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		SELECT id, name, email, password, created_at
		FROM users
		WHERE email = $1
	`
	var user User
	err := s.db.QueryRowContext(ctx, query, strings.ToLower(email)).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, userID string) (User, error) {
	const query = `
		SELECT id, name, email, password, created_at
		FROM users
		WHERE id = $1
	`
	var user User
	err := s.db.QueryRowContext(ctx, query, strings.TrimSpace(userID)).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		if isInvalidTextRepresentation(err) {
			return User{}, ErrBadRequest
		}
		return User{}, err
	}
	return user, nil
}

func (s *Store) ListAccessibleProjects(ctx context.Context, userID string) ([]Project, error) {
	const query = `
		SELECT DISTINCT p.id, p.name, COALESCE(p.description, ''), p.owner_id, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1
		ORDER BY p.created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

func (s *Store) CreateProject(ctx context.Context, ownerID, name, description string) (Project, error) {
	const query = `
		INSERT INTO projects (name, description, owner_id)
		VALUES ($1, NULLIF($2, ''), $3)
		RETURNING id, name, COALESCE(description, ''), owner_id, created_at
	`
	var project Project
	err := s.db.QueryRowContext(ctx, query, name, description, ownerID).
		Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt)
	return project, err
}

func (s *Store) GetProject(ctx context.Context, projectID string) (Project, error) {
	const query = `
		SELECT id, name, COALESCE(description, ''), owner_id, created_at
		FROM projects
		WHERE id = $1
	`
	var project Project
	err := s.db.QueryRowContext(ctx, query, projectID).
		Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		if isInvalidTextRepresentation(err) {
			return Project{}, ErrBadRequest
		}
		return Project{}, err
	}
	return project, nil
}

func (s *Store) UpdateProject(ctx context.Context, projectID, ownerID, name, description string) (Project, error) {
	const query = `
		UPDATE projects
		SET
			name = COALESCE(NULLIF($3, ''), name),
			description = COALESCE($4, description)
		WHERE id = $1 AND owner_id = $2
		RETURNING id, name, COALESCE(description, ''), owner_id, created_at
	`
	var project Project
	err := s.db.QueryRowContext(ctx, query, projectID, ownerID, name, nullableString(description)).
		Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if _, projectErr := s.GetProject(ctx, projectID); projectErr != nil {
				return Project{}, projectErr
			}
			return Project{}, ErrForbidden
		}
		if isInvalidTextRepresentation(err) {
			return Project{}, ErrBadRequest
		}
		return Project{}, err
	}
	return project, nil
}

func (s *Store) DeleteProject(ctx context.Context, projectID, ownerID string) error {
	const query = `DELETE FROM projects WHERE id = $1 AND owner_id = $2`
	result, err := s.db.ExecContext(ctx, query, projectID, ownerID)
	if err != nil {
		if isInvalidTextRepresentation(err) {
			return ErrBadRequest
		}
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		if _, projectErr := s.GetProject(ctx, projectID); projectErr != nil {
			return projectErr
		}
		return ErrForbidden
	}
	return nil
}

func (s *Store) ListProjectTasks(ctx context.Context, projectID string, filters TaskFilters) ([]Task, error) {
	base := `
		SELECT id, title, COALESCE(description, ''), status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		FROM tasks
		WHERE project_id = $1
	`
	args := []any{projectID}
	index := 2
	if filters.Status != "" {
		base += fmt.Sprintf(" AND status = $%d", index)
		args = append(args, filters.Status)
		index++
	}
	if filters.AssigneeID != "" {
		base += fmt.Sprintf(" AND assignee_id = $%d", index)
		args = append(args, filters.AssigneeID)
		index++
	}
	base += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		task, scanErr := scanTask(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (s *Store) GetProjectWithTasks(ctx context.Context, projectID string) (ProjectWithTasks, error) {
	project, err := s.GetProject(ctx, projectID)
	if err != nil {
		return ProjectWithTasks{}, err
	}
	tasks, err := s.ListProjectTasks(ctx, projectID, TaskFilters{})
	if err != nil {
		return ProjectWithTasks{}, err
	}
	return ProjectWithTasks{Project: project, Tasks: tasks}, nil
}

type TaskFilters struct {
	Status     string
	AssigneeID string
}

type TaskAccess struct {
	TaskID     string
	ProjectID  string
	IsOwner    bool
	IsCreator  bool
	IsAssignee bool
}

func (s *Store) CanAccessProject(ctx context.Context, projectID, userID string) (bool, error) {
	const query = `
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
	`
	var allowed bool
	if err := s.db.QueryRowContext(ctx, query, projectID, userID).Scan(&allowed); err != nil {
		if isInvalidTextRepresentation(err) {
			return false, ErrBadRequest
		}
		return false, err
	}
	return allowed, nil
}

func (s *Store) GetTaskAccess(ctx context.Context, taskID, actorID string) (TaskAccess, error) {
	const query = `
		SELECT
			t.id,
			t.project_id,
			(p.owner_id = $2) AS is_owner,
			(t.creator_id = $2) AS is_creator,
			(t.assignee_id = $2) AS is_assignee
		FROM tasks t
		INNER JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1
	`

	var access TaskAccess
	err := s.db.QueryRowContext(ctx, query, taskID, actorID).Scan(
		&access.TaskID,
		&access.ProjectID,
		&access.IsOwner,
		&access.IsCreator,
		&access.IsAssignee,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TaskAccess{}, ErrNotFound
		}
		if isInvalidTextRepresentation(err) {
			return TaskAccess{}, ErrBadRequest
		}
		return TaskAccess{}, err
	}

	return access, nil
}

type CreateTaskInput struct {
	ProjectID   string
	Title       string
	Description string
	Status      string
	Priority    string
	AssigneeID  *string
	DueDate     *time.Time
	CreatorID   string
}

func (s *Store) CreateTask(ctx context.Context, input CreateTaskInput) (Task, error) {
	const query = `
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
	`
	var dueDate any
	if input.DueDate != nil {
		dueDate = input.DueDate.Format("2006-01-02")
	}
	row := s.db.QueryRowContext(ctx, query,
		input.Title,
		input.Description,
		input.Status,
		input.Priority,
		input.ProjectID,
		input.AssigneeID,
		dueDate,
		input.CreatorID,
	)

	task, err := scanTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if _, projectErr := s.GetProject(ctx, input.ProjectID); projectErr != nil {
				return Task{}, projectErr
			}
			return Task{}, ErrForbidden
		}
		if isInvalidTextRepresentation(err) {
			return Task{}, ErrBadRequest
		}
		return Task{}, err
	}
	return task, nil
}

type UpdateTaskInput struct {
	ID          string
	ActorID     string
	Title       StringPatch
	Description NullableStringPatch
	Status      StringPatch
	Priority    StringPatch
	AssigneeID  NullableStringPatch
	DueDate     NullableDatePatch
}

type StringPatch struct {
	Set   bool
	Value string
}

type NullableStringPatch struct {
	Set   bool
	Value *string
}

type NullableDatePatch struct {
	Set   bool
	Value *time.Time
}

func (s *Store) UpdateTask(ctx context.Context, input UpdateTaskInput) (Task, error) {
	const query = `
		UPDATE tasks t
		SET
			title = CASE WHEN $2 THEN $3 ELSE t.title END,
			description = CASE WHEN $4 THEN NULLIF($5, '') ELSE t.description END,
			status = CASE WHEN $6 THEN $7::task_status ELSE t.status END,
			priority = CASE WHEN $8 THEN $9::task_priority ELSE t.priority END,
			assignee_id = CASE WHEN $10 THEN $11::uuid ELSE t.assignee_id END,
			due_date = CASE WHEN $12 THEN $13::date ELSE t.due_date END
		FROM projects p
		WHERE t.project_id = p.id
		  AND t.id = $1
		  AND (
			p.owner_id = $14
			OR t.creator_id = $14
			OR t.assignee_id = $14
		  )
		RETURNING t.id, t.title, COALESCE(t.description, ''), t.status, t.priority, t.project_id, t.assignee_id, t.creator_id, t.due_date, t.created_at, t.updated_at
	`
	var dueDate any
	if input.DueDate.Value != nil {
		dueDate = input.DueDate.Value.Format("2006-01-02")
	}
	row := s.db.QueryRowContext(
		ctx,
		query,
		input.ID,
		input.Title.Set,
		nullableString(input.Title.Value),
		input.Description.Set,
		derefString(input.Description.Value),
		input.Status.Set,
		nullableString(input.Status.Value),
		input.Priority.Set,
		nullableString(input.Priority.Value),
		input.AssigneeID.Set,
		input.AssigneeID.Value,
		input.DueDate.Set,
		dueDate,
		input.ActorID,
	)
	task, err := scanTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if _, taskErr := s.GetTask(ctx, input.ID); taskErr != nil {
				return Task{}, taskErr
			}
			return Task{}, ErrForbidden
		}
		if isInvalidTextRepresentation(err) {
			return Task{}, ErrBadRequest
		}
		return Task{}, err
	}
	return task, nil
}

func (s *Store) GetTask(ctx context.Context, taskID string) (Task, error) {
	const query = `
		SELECT id, title, COALESCE(description, ''), status, priority, project_id, assignee_id, creator_id, due_date, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	task, err := scanTask(s.db.QueryRowContext(ctx, query, taskID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		if isInvalidTextRepresentation(err) {
			return Task{}, ErrBadRequest
		}
		return Task{}, err
	}
	return task, nil
}

func (s *Store) DeleteTask(ctx context.Context, taskID, actorID string) error {
	const query = `
		DELETE FROM tasks t
		USING projects p
		WHERE t.project_id = p.id
		  AND t.id = $1
		  AND (p.owner_id = $2 OR t.creator_id = $2)
	`
	result, err := s.db.ExecContext(ctx, query, taskID, actorID)
	if err != nil {
		if isInvalidTextRepresentation(err) {
			return ErrBadRequest
		}
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		if _, taskErr := s.GetTask(ctx, taskID); taskErr != nil {
			return taskErr
		}
		return ErrForbidden
	}
	return nil
}

func scanTask(scanner interface {
	Scan(dest ...any) error
}) (Task, error) {
	var task Task
	var assigneeID sql.NullString
	var dueDate sql.NullTime
	err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.ProjectID,
		&assigneeID,
		&task.CreatorID,
		&dueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return Task{}, err
	}
	if assigneeID.Valid {
		task.AssigneeID = &assigneeID.String
	}
	if dueDate.Valid {
		dateOnly := dueDate.Time.UTC()
		task.DueDate = &dateOnly
	}
	return task, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func isInvalidTextRepresentation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "22P02"
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
