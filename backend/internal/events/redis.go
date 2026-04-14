package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const StreamName = "taskflow:task:events"

// RedisStreamPublisher publishes TaskChangedEvent messages to a Redis Stream
// using XADD. The stream has a soft cap (MAXLEN ~) to prevent unbounded growth.
type RedisStreamPublisher struct {
	client *redis.Client
}

func NewRedisStreamPublisher(client *redis.Client) *RedisStreamPublisher {
	return &RedisStreamPublisher{client: client}
}

func (p *RedisStreamPublisher) Publish(ctx context.Context, event TaskChangedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamName,
		MaxLen: 10_000, // soft cap; old entries trimmed automatically
		Approx: true,
		Values: map[string]any{"payload": string(payload)},
	}).Err()
}
