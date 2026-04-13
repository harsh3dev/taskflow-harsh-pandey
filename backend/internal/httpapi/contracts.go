package httpapi

import (
	"context"

	"github.com/harshpn/taskflow/internal/service"
)

type authService interface {
	Register(ctx context.Context, input service.RegisterInput, meta service.SessionMetadata) (service.AuthSession, map[string]string, error)
	Login(ctx context.Context, input service.LoginInput, meta service.SessionMetadata) (service.AuthSession, map[string]string, error)
	Refresh(ctx context.Context, input service.RefreshInput) (service.AuthSession, map[string]string, error)
	Logout(ctx context.Context, refreshToken string) error
}

type projectService interface {
	ListAccessibleProjects(ctx context.Context, userID string) ([]service.Project, error)
	CreateProject(ctx context.Context, userID, name, description string) (service.Project, map[string]string, error)
	GetProject(ctx context.Context, projectID, userID string) (service.ProjectWithTasks, error)
	UpdateProject(ctx context.Context, projectID, userID, name, description string) (service.Project, error)
	DeleteProject(ctx context.Context, projectID, userID string) error
	ListProjectTasks(ctx context.Context, projectID, userID string, filters service.TaskFilters) ([]service.Task, error)
}

type taskService interface {
	CreateTask(ctx context.Context, projectID, actorID string, input service.TaskCreateInput) (service.Task, map[string]string, error)
	UpdateTask(ctx context.Context, taskID, actorID string, input service.TaskUpdateInput) (service.Task, map[string]string, error)
	DeleteTask(ctx context.Context, taskID, actorID string) error
}

type userService interface {
	ListUsers(ctx context.Context, search string) ([]service.User, error)
}
