package httpapi

import (
	"net/http"
	"strings"

	"github.com/harshpn/taskflow/internal/auth"
	"github.com/harshpn/taskflow/internal/store"
)

type projectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	projects, err := s.store.ListAccessibleProjects(r.Context(), user.UserID)
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserFromContext(r.Context()); !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	users, err := s.store.ListUsers(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	var req projectRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeValidationError(w, r, map[string]string{"name": "is required"})
		return
	}

	project, err := s.store.CreateProject(r.Context(), user.UserID, strings.TrimSpace(req.Name), strings.TrimSpace(req.Description))
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"project": project})
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	projectID := r.PathValue("id")
	project, err := s.store.GetProject(r.Context(), projectID)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	allowed, err := s.store.CanAccessProject(r.Context(), projectID, user.UserID)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	if !allowed {
		writeError(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}

	tasks, err := s.store.ListProjectTasks(r.Context(), project.ID, store.TaskFilters{})
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, store.ProjectWithTasks{Project: project, Tasks: tasks})
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	var req projectRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	project, err := s.store.UpdateProject(r.Context(), r.PathValue("id"), user.UserID, strings.TrimSpace(req.Name), strings.TrimSpace(req.Description))
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"project": project})
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	err := s.store.DeleteProject(r.Context(), r.PathValue("id"), user.UserID)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleListProjectTasks(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	projectID := r.PathValue("id")
	if _, err := s.store.GetProject(r.Context(), projectID); err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	allowed, err := s.store.CanAccessProject(r.Context(), projectID, user.UserID)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	if !allowed {
		writeError(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}

	filters := store.TaskFilters{
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		AssigneeID: strings.TrimSpace(r.URL.Query().Get("assignee")),
	}
	tasks, err := s.store.ListProjectTasks(r.Context(), projectID, filters)
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}
