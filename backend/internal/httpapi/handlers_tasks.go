package httpapi

import (
	"net/http"

	"github.com/harshpn/taskflow/internal/auth"
)

type taskRequest struct {
	Title       optionalString `json:"title"`
	Description optionalString `json:"description"`
	Status      optionalString `json:"status"`
	Priority    optionalString `json:"priority"`
	AssigneeID  optionalString `json:"assignee_id"`
	DueDate     optionalString `json:"due_date"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	projectID := r.PathValue("id")

	var req taskRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	task, fields, err := s.tasks.CreateTask(r.Context(), projectID, user.UserID, taskCreateInputFromRequest(req))
	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"task": newTaskResponse(task)})
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	var req taskRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	task, fields, err := s.tasks.UpdateTask(r.Context(), r.PathValue("id"), user.UserID, taskUpdateInputFromRequest(req))
	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"task": newTaskResponse(task)})
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	err := s.tasks.DeleteTask(r.Context(), r.PathValue("id"), user.UserID)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
