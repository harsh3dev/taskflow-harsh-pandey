package events

import "context"

// EventPublisher is the interface TaskService uses to emit domain events.
// Concrete implementations (Redis Streams, in-memory) are injected at startup.
type EventPublisher interface {
	Publish(ctx context.Context, event TaskChangedEvent) error
}

// NoopPublisher discards all events. Used in tests and when notifications are
// disabled.
type NoopPublisher struct{}

func (NoopPublisher) Publish(_ context.Context, _ TaskChangedEvent) error { return nil }
