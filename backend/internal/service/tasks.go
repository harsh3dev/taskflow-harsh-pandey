package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/harshpn/taskflow/internal/store"
)

type TaskService struct {
	repo taskStore
}

type OptionalString struct {
	Set   bool
	Null  bool
	Value string
}

type TaskCreateInput struct {
	Title       OptionalString
	Description OptionalString
	Status      OptionalString
	Priority    OptionalString
	AssigneeID  OptionalString
	DueDate     OptionalString
}

type TaskUpdateInput = TaskCreateInput

type TaskFilters struct {
	Status     string
	AssigneeID string
}

type ValidatedTaskInput struct {
	CreateInput store.CreateTaskInput
	UpdateInput store.UpdateTaskInput
}

func NewTaskService(repo taskStore) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(ctx context.Context, projectID, actorID string, input TaskCreateInput) (Task, map[string]string, error) {
	fields, validated, err := s.validateTaskInput(ctx, input, true)
	if err != nil {
		return Task{}, nil, err
	}
	if len(fields) > 0 {
		return Task{}, fields, nil
	}

	createInput := validated.CreateInput
	createInput.ProjectID = projectID
	createInput.CreatorID = actorID

	task, err := s.repo.CreateTask(ctx, createInput)
	if err != nil {
		return Task{}, nil, err
	}
	return taskFromStore(task), nil, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID, actorID string, input TaskUpdateInput) (Task, map[string]string, error) {
	fields, validated, err := s.validateTaskInput(ctx, input, false)
	if err != nil {
		return Task{}, nil, err
	}
	if len(fields) > 0 {
		return Task{}, fields, nil
	}

	access, err := s.repo.GetTaskAccess(ctx, taskID, actorID)
	if err != nil {
		return Task{}, nil, err
	}
	if !access.IsOwner && !access.IsCreator && !access.IsAssignee {
		return Task{}, nil, store.ErrForbidden
	}
	if access.IsAssignee && !access.IsOwner && !access.IsCreator {
		if !isStatusOnlyTaskUpdate(validated.UpdateInput, input.AssigneeID.Set) {
			return Task{}, map[string]string{"status": "assignees may only update task status"}, nil
		}
	}

	updateInput := validated.UpdateInput
	updateInput.ID = taskID
	updateInput.ActorID = actorID

	task, err := s.repo.UpdateTask(ctx, updateInput)
	if err != nil {
		return Task{}, nil, err
	}
	return taskFromStore(task), nil, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, taskID, actorID string) error {
	return s.repo.DeleteTask(ctx, taskID, actorID)
}

func (s *TaskService) validateTaskInput(ctx context.Context, req TaskCreateInput, requireTitle bool) (map[string]string, ValidatedTaskInput, error) {
	fields := map[string]string{}

	title := stringPatchFromOptional(req.Title)
	description := nullableStringPatchFromOptional(req.Description)
	status := stringPatchFromOptional(req.Status)
	priority := stringPatchFromOptional(req.Priority)
	dueDate := nullableDatePatchFromOptional(req.DueDate, fields)

	if requireTitle {
		if !req.Title.Set || req.Title.Null || strings.TrimSpace(req.Title.Value) == "" {
			fields["title"] = "is required"
		}
	}
	if title.Set && strings.TrimSpace(title.Value) == "" {
		fields["title"] = "must not be empty"
	}
	if req.Status.Set && (req.Status.Null || status.Value == "") {
		fields["status"] = "must not be empty"
	}
	if status.Set && !contains([]string{"todo", "in_progress", "done"}, status.Value) {
		fields["status"] = "must be one of todo, in_progress, done"
	}
	if req.Priority.Set && (req.Priority.Null || priority.Value == "") {
		fields["priority"] = "must not be empty"
	}
	if priority.Set && !contains([]string{"low", "medium", "high"}, priority.Value) {
		fields["priority"] = "must be one of low, medium, high"
	}

	assigneeID, assigneeFieldError, err := s.validateAssigneeID(ctx, req.AssigneeID)
	if err != nil {
		return nil, ValidatedTaskInput{}, err
	}
	if assigneeFieldError != "" {
		fields["assignee_id"] = assigneeFieldError
	}

	return fields, ValidatedTaskInput{
		CreateInput: store.CreateTaskInput{
			Title:       title.Value,
			Description: derefOrEmpty(description.Value),
			Status:      defaultTaskValue(status, "todo"),
			Priority:    defaultTaskValue(priority, "medium"),
			AssigneeID:  assigneeID,
			DueDate:     dueDate.Value,
		},
		UpdateInput: store.UpdateTaskInput{
			Title:       title,
			Description: description,
			Status:      status,
			Priority:    priority,
			AssigneeID:  store.NullableStringPatch{Set: req.AssigneeID.Set, Value: assigneeID},
			DueDate:     dueDate,
		},
	}, nil
}

func (s *TaskService) validateAssigneeID(ctx context.Context, assigneeID OptionalString) (*string, string, error) {
	if !assigneeID.Set || assigneeID.Null {
		return nil, "", nil
	}
	trimmedValue := strings.TrimSpace(assigneeID.Value)
	if trimmedValue == "" {
		return nil, "", nil
	}
	trimmed := &trimmedValue

	_, err := s.repo.GetUserByID(ctx, *trimmed)
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

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func defaultTaskValue(value store.StringPatch, fallback string) string {
	if !value.Set || value.Value == "" {
		return fallback
	}
	return value.Value
}

func stringPatchFromOptional(value OptionalString) store.StringPatch {
	if !value.Set || value.Null {
		return store.StringPatch{}
	}
	return store.StringPatch{
		Set:   true,
		Value: strings.TrimSpace(value.Value),
	}
}

func nullableStringPatchFromOptional(value OptionalString) store.NullableStringPatch {
	if !value.Set {
		return store.NullableStringPatch{}
	}
	if value.Null {
		return store.NullableStringPatch{Set: true, Value: nil}
	}
	trimmed := strings.TrimSpace(value.Value)
	if trimmed == "" {
		return store.NullableStringPatch{Set: true, Value: nil}
	}
	return store.NullableStringPatch{Set: true, Value: &trimmed}
}

func nullableDatePatchFromOptional(value OptionalString, fields map[string]string) store.NullableDatePatch {
	if !value.Set {
		return store.NullableDatePatch{}
	}
	if value.Null || strings.TrimSpace(value.Value) == "" {
		return store.NullableDatePatch{Set: true, Value: nil}
	}
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value.Value))
	if err != nil {
		fields["due_date"] = "must be in YYYY-MM-DD format"
		return store.NullableDatePatch{}
	}
	return store.NullableDatePatch{Set: true, Value: &parsed}
}

func derefOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func isStatusOnlyTaskUpdate(input store.UpdateTaskInput, assigneeSet bool) bool {
	return input.Status.Set &&
		!input.Title.Set &&
		!input.Description.Set &&
		!input.Priority.Set &&
		!input.DueDate.Set &&
		!assigneeSet
}
