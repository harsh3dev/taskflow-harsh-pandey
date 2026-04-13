package service

import "time"

type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
}

type Project struct {
	ID          string
	Name        string
	Description string
	OwnerID     string
	CreatedAt   time.Time
}

type Task struct {
	ID          string
	Title       string
	Description string
	Status      string
	Priority    string
	ProjectID   string
	AssigneeID  *string
	CreatorID   string
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectWithTasks struct {
	Project Project
	Tasks   []Task
}

type AuthTokens struct {
	AccessToken      string
	RefreshToken     string
	TokenType        string
	ExpiresInSeconds int
}

type AuthSession struct {
	User   User
	Tokens AuthTokens
}

// Pagination is returned alongside paginated list results.
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type AssigneeCount struct {
	UserID string
	Name   string
	Count  int
}

type ProjectStats struct {
	StatusCounts   map[string]int
	AssigneeCounts []AssigneeCount
}
