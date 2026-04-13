package httpapi

import (
	"time"

	"github.com/harshpn/taskflow/internal/service"
)

type authResponse struct {
	Token            string       `json:"token"`
	AccessToken      string       `json:"access_token"`
	RefreshToken     string       `json:"refresh_token"`
	TokenType        string       `json:"token_type"`
	ExpiresInSeconds int          `json:"expires_in_seconds"`
	User             userResponse `json:"user"`
}

type userResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type projectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type taskResponse struct {
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

type projectDetailResponse struct {
	Project projectResponse `json:"project"`
	Tasks   []taskResponse  `json:"tasks"`
}

func newAuthResponse(session service.AuthSession) authResponse {
	return authResponse{
		Token:            session.Tokens.AccessToken,
		AccessToken:      session.Tokens.AccessToken,
		RefreshToken:     session.Tokens.RefreshToken,
		TokenType:        session.Tokens.TokenType,
		ExpiresInSeconds: session.Tokens.ExpiresInSeconds,
		User:             newUserResponse(session.User),
	}
}

func newUserResponse(user service.User) userResponse {
	return userResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func newUsersResponse(users []service.User) []userResponse {
	result := make([]userResponse, 0, len(users))
	for _, user := range users {
		result = append(result, newUserResponse(user))
	}
	return result
}

func newProjectResponse(project service.Project) projectResponse {
	return projectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt,
	}
}

func newProjectsResponse(projects []service.Project) []projectResponse {
	result := make([]projectResponse, 0, len(projects))
	for _, project := range projects {
		result = append(result, newProjectResponse(project))
	}
	return result
}

func newTaskResponse(task service.Task) taskResponse {
	return taskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		ProjectID:   task.ProjectID,
		AssigneeID:  task.AssigneeID,
		CreatorID:   task.CreatorID,
		DueDate:     task.DueDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func newTasksResponse(tasks []service.Task) []taskResponse {
	result := make([]taskResponse, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, newTaskResponse(task))
	}
	return result
}

func newProjectDetailResponse(project service.ProjectWithTasks) projectDetailResponse {
	return projectDetailResponse{
		Project: newProjectResponse(project.Project),
		Tasks:   newTasksResponse(project.Tasks),
	}
}
