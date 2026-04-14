package events

import "time"

// ChangeKind identifies which task field changed.
type ChangeKind string

const (
	ChangeKindStatus   ChangeKind = "status_changed"
	ChangeKindPriority ChangeKind = "priority_changed"
	ChangeKindAssignee ChangeKind = "assignee_changed"
	ChangeKindDueDate  ChangeKind = "due_date_changed"
)

// TaskChangedEvent is published to the Redis Stream whenever a watched task
// field changes. One event is emitted per changed field so each email has a
// single, clear subject line.
type TaskChangedEvent struct {
	EventID    string     `json:"event_id"`
	TaskID     string     `json:"task_id"`
	ProjectID  string     `json:"project_id"`
	TaskTitle  string     `json:"task_title"`
	AssigneeID string     `json:"assignee_id"` // user to notify
	ChangeKind ChangeKind `json:"change_kind"`
	OldValue   string     `json:"old_value"`
	NewValue   string     `json:"new_value"`
	ChangedAt  time.Time  `json:"changed_at"`
}
