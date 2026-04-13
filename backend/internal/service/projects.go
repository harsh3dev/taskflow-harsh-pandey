package service

import (
	"context"
	"strings"

	"github.com/harshpn/taskflow/internal/store"
)

type ProjectService struct {
	store *store.Store
}

func NewProjectService(store *store.Store) *ProjectService {
	return &ProjectService{store: store}
}

func (s *ProjectService) ListAccessibleProjects(ctx context.Context, userID string) ([]Project, error) {
	projects, err := s.store.ListAccessibleProjects(ctx, userID)
	if err != nil {
		return nil, err
	}
	return projectsFromStore(projects), nil
}

func (s *ProjectService) CreateProject(ctx context.Context, userID, name, description string) (Project, map[string]string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return Project{}, map[string]string{"name": "is required"}, nil
	}

	project, err := s.store.CreateProject(ctx, userID, trimmedName, strings.TrimSpace(description))
	if err != nil {
		return Project{}, nil, err
	}
	return projectFromStore(project), nil, nil
}

func (s *ProjectService) GetProject(ctx context.Context, projectID, userID string) (ProjectWithTasks, error) {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return ProjectWithTasks{}, err
	}

	allowed, err := s.store.CanAccessProject(ctx, projectID, userID)
	if err != nil {
		return ProjectWithTasks{}, err
	}
	if !allowed {
		return ProjectWithTasks{}, store.ErrForbidden
	}

	tasks, err := s.store.ListProjectTasks(ctx, project.ID, store.TaskFilters{})
	if err != nil {
		return ProjectWithTasks{}, err
	}

	return ProjectWithTasks{
		Project: projectFromStore(project),
		Tasks:   tasksFromStore(tasks),
	}, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, projectID, userID, name, description string) (Project, error) {
	project, err := s.store.UpdateProject(ctx, projectID, userID, strings.TrimSpace(name), strings.TrimSpace(description))
	if err != nil {
		return Project{}, err
	}
	return projectFromStore(project), nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectID, userID string) error {
	return s.store.DeleteProject(ctx, projectID, userID)
}

func (s *ProjectService) ListProjectTasks(ctx context.Context, projectID, userID string, filters TaskFilters) ([]Task, error) {
	if _, err := s.store.GetProject(ctx, projectID); err != nil {
		return nil, err
	}

	allowed, err := s.store.CanAccessProject(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, store.ErrForbidden
	}

	tasks, err := s.store.ListProjectTasks(ctx, projectID, store.TaskFilters{
		Status:     strings.TrimSpace(filters.Status),
		AssigneeID: strings.TrimSpace(filters.AssigneeID),
	})
	if err != nil {
		return nil, err
	}
	return tasksFromStore(tasks), nil
}
