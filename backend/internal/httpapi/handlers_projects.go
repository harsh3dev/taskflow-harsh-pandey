package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/harshpn/taskflow/internal/auth"
	"github.com/harshpn/taskflow/internal/service"
)

type projectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// parsePagination reads ?page= and ?limit= from the request, applying sane defaults and caps.
func parsePagination(r *http.Request, defaultLimit int) (page, limit int) {
	page, limit = 1, defaultLimit
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	return
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	page, limit := parsePagination(r, 50)
	projects, pagination, err := s.projects.ListAccessibleProjects(r.Context(), user.UserID, page, limit)
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"projects":   newProjectsResponse(projects),
		"pagination": newPaginationResponse(pagination),
	})
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

	page, limit := parsePagination(r, 50)
	tasks, pagination, err := s.projects.ListProjectTasks(r.Context(), r.PathValue("id"), user.UserID, service.TaskFilters{
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		AssigneeID: strings.TrimSpace(r.URL.Query().Get("assignee")),
	}, page, limit)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tasks":      newTasksResponse(tasks),
		"pagination": newPaginationResponse(pagination),
	})
}

func (s *Server) handleGetProjectStats(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	stats, err := s.projects.GetProjectStats(r.Context(), r.PathValue("id"), user.UserID)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, newProjectStatsResponse(stats))
}
