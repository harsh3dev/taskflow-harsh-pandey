package httpapi

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/harshpn/taskflow/internal/store"
)

type optionalString struct {
	Set   bool
	Null  bool
	Value string
}

func (o *optionalString) UnmarshalJSON(data []byte) error {
	o.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		o.Null = true
		o.Value = ""
		return nil
	}
	return json.Unmarshal(data, &o.Value)
}

type validatedTaskInput struct {
	Title       store.StringPatch
	Description store.NullableStringPatch
	Status      store.StringPatch
	Priority    store.StringPatch
	DueDate     store.NullableDatePatch
}

func validateTaskRequest(req taskRequest, requireTitle bool) (map[string]string, validatedTaskInput) {
	fields := map[string]string{}
	input := validatedTaskInput{
		Title:       stringPatchFromOptional(req.Title),
		Description: nullableStringPatchFromOptional(req.Description),
		Status:      stringPatchFromOptional(req.Status),
		Priority:    stringPatchFromOptional(req.Priority),
		DueDate:     nullableDatePatchFromOptional(req.DueDate, fields),
	}

	if requireTitle {
		if !req.Title.Set || req.Title.Null || strings.TrimSpace(req.Title.Value) == "" {
			fields["title"] = "is required"
		}
	}
	if input.Title.Set && strings.TrimSpace(input.Title.Value) == "" {
		fields["title"] = "must not be empty"
	}
	if req.Status.Set && (req.Status.Null || input.Status.Value == "") {
		fields["status"] = "must not be empty"
	}
	if input.Status.Set && !contains([]string{"todo", "in_progress", "done"}, input.Status.Value) {
		fields["status"] = "must be one of todo, in_progress, done"
	}
	if req.Priority.Set && (req.Priority.Null || input.Priority.Value == "") {
		fields["priority"] = "must not be empty"
	}
	if input.Priority.Set && !contains([]string{"low", "medium", "high"}, input.Priority.Value) {
		fields["priority"] = "must be one of low, medium, high"
	}

	return fields, input
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

func stringPatchFromOptional(value optionalString) store.StringPatch {
	if !value.Set || value.Null {
		return store.StringPatch{}
	}
	return store.StringPatch{
		Set:   true,
		Value: strings.TrimSpace(value.Value),
	}
}

func nullableStringPatchFromOptional(value optionalString) store.NullableStringPatch {
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

func nullableDatePatchFromOptional(value optionalString, fields map[string]string) store.NullableDatePatch {
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

func isStatusOnlyTaskUpdate(input validatedTaskInput, assigneeSet bool) bool {
	return input.Status.Set &&
		!input.Title.Set &&
		!input.Description.Set &&
		!input.Priority.Set &&
		!input.DueDate.Set &&
		!assigneeSet
}
