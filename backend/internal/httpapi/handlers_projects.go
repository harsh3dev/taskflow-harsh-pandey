package httpapi

import (
	"net/http"
	"strings"

	"github.com/harshpn/taskflow/internal/auth"
	"github.com/harshpn/taskflow/internal/service"
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

	projects, err := s.projects.ListAccessibleProjects(r.Context(), user.UserID)
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"projects": newProjectsResponse(projects)})
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserFromContext(r.Context()); !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	users, err := s.users.ListUsers(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"users": newUsersResponse(users)})
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

	project, fields, err := s.projects.CreateProject(r.Context(), user.UserID, req.Name, req.Description)
	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"project": newProjectResponse(project)})
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	project, err := s.projects.GetProject(r.Context(), r.PathValue("id"), user.UserID)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, newProjectDetailResponse(project))
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

	project, err := s.projects.UpdateProject(r.Context(), r.PathValue("id"), user.UserID, req.Name, req.Description)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"project": newProjectResponse(project)})
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	err := s.projects.DeleteProject(r.Context(), r.PathValue("id"), user.UserID)
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

	tasks, err := s.projects.ListProjectTasks(r.Context(), r.PathValue("id"), user.UserID, service.TaskFilters{
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		AssigneeID: strings.TrimSpace(r.URL.Query().Get("assignee")),
	})
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"tasks": newTasksResponse(tasks)})
}
