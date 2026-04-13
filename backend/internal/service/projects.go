package service

import (
	"context"
	"strings"

	"github.com/harshpn/taskflow/internal/store"
)

type ProjectService struct {
	repo projectStore
}

func NewProjectService(repo projectStore) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) ListAccessibleProjects(ctx context.Context, userID string, page, limit int) ([]Project, Pagination, error) {
	offset := (page - 1) * limit
	projects, total, err := s.repo.ListAccessibleProjects(ctx, userID, limit, offset)
	if err != nil {
		return nil, Pagination{}, err
	}
	return projectsFromStore(projects), Pagination{Page: page, Limit: limit, Total: total}, nil
}

func (s *ProjectService) CreateProject(ctx context.Context, userID, name, description string) (Project, map[string]string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return Project{}, map[string]string{"name": "is required"}, nil
	}

	project, err := s.repo.CreateProject(ctx, userID, trimmedName, strings.TrimSpace(description))
	if err != nil {
		return Project{}, nil, err
	}
	return projectFromStore(project), nil, nil
}

func (s *ProjectService) GetProject(ctx context.Context, projectID, userID string) (ProjectWithTasks, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return ProjectWithTasks{}, err
	}

	allowed, err := s.repo.CanAccessProject(ctx, projectID, userID)
	if err != nil {
		return ProjectWithTasks{}, err
	}
	if !allowed {
		return ProjectWithTasks{}, store.ErrForbidden
	}

	// No pagination for project detail — return all tasks.
	tasks, _, err := s.repo.ListProjectTasks(ctx, project.ID, store.TaskFilters{}, 1000, 0)
	if err != nil {
		return ProjectWithTasks{}, err
	}

	return ProjectWithTasks{
		Project: projectFromStore(project),
		Tasks:   tasksFromStore(tasks),
	}, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, projectID, userID, name, description string) (Project, error) {
	project, err := s.repo.UpdateProject(ctx, projectID, userID, strings.TrimSpace(name), strings.TrimSpace(description))
	if err != nil {
		return Project{}, err
	}
	return projectFromStore(project), nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectID, userID string) error {
	return s.repo.DeleteProject(ctx, projectID, userID)
}

func (s *ProjectService) ListProjectTasks(ctx context.Context, projectID, userID string, filters TaskFilters, page, limit int) ([]Task, Pagination, error) {
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, Pagination{}, err
	}

	allowed, err := s.repo.CanAccessProject(ctx, projectID, userID)
	if err != nil {
		return nil, Pagination{}, err
	}
	if !allowed {
		return nil, Pagination{}, store.ErrForbidden
	}

	offset := (page - 1) * limit
	tasks, total, err := s.repo.ListProjectTasks(ctx, projectID, store.TaskFilters{
		Status:     strings.TrimSpace(filters.Status),
		AssigneeID: strings.TrimSpace(filters.AssigneeID),
	}, limit, offset)
	if err != nil {
		return nil, Pagination{}, err
	}
	return tasksFromStore(tasks), Pagination{Page: page, Limit: limit, Total: total}, nil
}

func (s *ProjectService) GetProjectStats(ctx context.Context, projectID, userID string) (ProjectStats, error) {
	stats, err := s.repo.GetProjectStats(ctx, projectID, userID)
	if err != nil {
		return ProjectStats{}, err
	}
	return projectStatsFromStore(stats), nil
}
