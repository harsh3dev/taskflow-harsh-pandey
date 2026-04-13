package service

import (
	"context"

	"github.com/harshpn/taskflow/internal/store"
)

// userStore covers user-record persistence operations used by AuthService.
type userStore interface {
	CreateUser(ctx context.Context, input store.CreateUserInput) (store.User, error)
	GetUserByEmail(ctx context.Context, email string) (store.User, error)
	GetUserByID(ctx context.Context, userID string) (store.User, error)
}

// sessionStore covers refresh-session persistence operations used by AuthService.
type sessionStore interface {
	CreateRefreshSession(ctx context.Context, input store.CreateRefreshSessionInput) (store.RefreshSession, error)
	RotateRefreshSession(ctx context.Context, input store.RotateRefreshSessionInput) (store.RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, tokenHash, reason string) error
}

// projectStore covers project and task-list persistence operations used by ProjectService.
type projectStore interface {
	ListAccessibleProjects(ctx context.Context, userID string, limit, offset int) ([]store.Project, int, error)
	CreateProject(ctx context.Context, ownerID, name, description string) (store.Project, error)
	GetProject(ctx context.Context, projectID string) (store.Project, error)
	UpdateProject(ctx context.Context, projectID, ownerID, name, description string) (store.Project, error)
	DeleteProject(ctx context.Context, projectID, ownerID string) error
	CanAccessProject(ctx context.Context, projectID, userID string) (bool, error)
	ListProjectTasks(ctx context.Context, projectID string, filters store.TaskFilters, limit, offset int) ([]store.Task, int, error)
	GetProjectStats(ctx context.Context, projectID, userID string) (store.ProjectStats, error)
}

// taskStore covers task persistence operations used by TaskService.
// It also includes GetUserByID for assignee existence validation.
type taskStore interface {
	CreateTask(ctx context.Context, input store.CreateTaskInput) (store.Task, error)
	UpdateTask(ctx context.Context, input store.UpdateTaskInput) (store.Task, error)
	DeleteTask(ctx context.Context, taskID, actorID string) error
	GetTaskAccess(ctx context.Context, taskID, actorID string) (store.TaskAccess, error)
	GetUserByID(ctx context.Context, userID string) (store.User, error)
}

// userListStore covers the single user-listing operation needed by UserService.
type userListStore interface {
	ListUsers(ctx context.Context, search string) ([]store.User, error)
}
