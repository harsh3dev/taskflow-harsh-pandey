package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/harshpn/taskflow/internal/auth"
	"github.com/harshpn/taskflow/internal/store"
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

	fields, validated := validateTaskRequest(req, true)
	assigneeID, assigneeFieldError, err := s.validateAssigneeID(r.Context(), req.AssigneeID)
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}
	if assigneeFieldError != "" {
		fields["assignee_id"] = assigneeFieldError
	}
	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}

	task, err := s.store.CreateTask(r.Context(), store.CreateTaskInput{
		ProjectID:   projectID,
		Title:       validated.Title.Value,
		Description: derefOrEmpty(validated.Description.Value),
		Status:      defaultTaskValue(validated.Status, "todo"),
		Priority:    defaultTaskValue(validated.Priority, "medium"),
		AssigneeID:  assigneeID,
		DueDate:     validated.DueDate.Value,
		CreatorID:   user.UserID,
	})
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"task": task})
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

	fields, validated := validateTaskRequest(req, false)
	assigneeID, assigneeFieldError, err := s.validateAssigneeID(r.Context(), req.AssigneeID)
	if err != nil {
		s.writeInternalError(w, r, err)
		return
	}
	if assigneeFieldError != "" {
		fields["assignee_id"] = assigneeFieldError
	}
	if len(fields) > 0 {
		writeValidationError(w, r, fields)
		return
	}

	access, err := s.store.GetTaskAccess(r.Context(), r.PathValue("id"), user.UserID)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}
	if !access.IsOwner && !access.IsCreator && !access.IsAssignee {
		writeError(w, r, http.StatusForbidden, "forbidden", "forbidden")
		return
	}
	if access.IsAssignee && !access.IsOwner && !access.IsCreator {
		if !isStatusOnlyTaskUpdate(validated, req.AssigneeID.Set) {
			writeValidationError(w, r, map[string]string{
				"status": "assignees may only update task status",
			})
			return
		}
	}

	input := store.UpdateTaskInput{
		ID:          r.PathValue("id"),
		ActorID:     user.UserID,
		Title:       validated.Title,
		Description: validated.Description,
		Status:      validated.Status,
		Priority:    validated.Priority,
		AssigneeID:  store.NullableStringPatch{Set: req.AssigneeID.Set, Value: assigneeID},
		DueDate:     validated.DueDate,
	}

	task, err := s.store.UpdateTask(r.Context(), input)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"task": task})
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, r, "missing authenticated user")
		return
	}

	err := s.store.DeleteTask(r.Context(), r.PathValue("id"), user.UserID)
	if err != nil {
		s.handleStoreError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) validateAssigneeID(ctx context.Context, assigneeID optionalString) (*string, string, error) {
	if !assigneeID.Set || assigneeID.Null {
		return nil, "", nil
	}
	trimmedValue := strings.TrimSpace(assigneeID.Value)
	if trimmedValue == "" {
		return nil, "", nil
	}
	trimmed := &trimmedValue

	_, err := s.store.GetUserByID(ctx, *trimmed)
	switch {
	case err == nil:
		return trimmed, "", nil
	case errors.Is(err, store.ErrBadRequest):
		return nil, "must be a valid user id", nil
	case errors.Is(err, store.ErrNotFound):
		return nil, "must reference an existing user", nil
	default:
		return nil, "", err
	}
}
