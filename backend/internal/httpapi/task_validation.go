package httpapi

import (
	"bytes"
	"encoding/json"

	"github.com/harshpn/taskflow/internal/service"
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

func taskCreateInputFromRequest(req taskRequest) service.TaskCreateInput {
	return service.TaskCreateInput{
		Title:       optionalStringToService(req.Title),
		Description: optionalStringToService(req.Description),
		Status:      optionalStringToService(req.Status),
		Priority:    optionalStringToService(req.Priority),
		AssigneeID:  optionalStringToService(req.AssigneeID),
		DueDate:     optionalStringToService(req.DueDate),
	}
}

func taskUpdateInputFromRequest(req taskRequest) service.TaskUpdateInput {
	return service.TaskUpdateInput(taskCreateInputFromRequest(req))
}

func optionalStringToService(value optionalString) service.OptionalString {
	return service.OptionalString{
		Set:   value.Set,
		Null:  value.Null,
		Value: value.Value,
	}
}
