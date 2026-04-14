package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/harshpn/taskflow/internal/email"
	"github.com/harshpn/taskflow/internal/events"
	"github.com/redis/go-redis/v9"
)

const (
	streamName     = events.StreamName
	groupName      = "notif-workers"
	claimMinIdle   = 5 * time.Minute // reclaim messages idle this long
	claimInterval  = 60 * time.Second
	blockTimeout   = 5 * time.Second
	batchSize      = 10
)

// UserLookup fetches a user's name and email by ID.
type UserLookup interface {
	GetUserByID(ctx context.Context, userID string) (userName, userEmail string, err error)
}

// NotificationWorker reads TaskChangedEvents from a Redis Stream consumer
// group and sends email notifications. It provides at-least-once delivery:
// messages are only ACKed after successful email delivery. Unacknowledged
// messages are reclaimed via XAUTOCLAIM after claimMinIdle.
type NotificationWorker struct {
	redis      *redis.Client
	emailer    *email.SMTPSender
	userLookup UserLookup
	logger     *slog.Logger
	workerID   string
}

func NewNotificationWorker(
	rdb *redis.Client,
	emailer *email.SMTPSender,
	userLookup UserLookup,
	logger *slog.Logger,
) *NotificationWorker {
	hostname, _ := os.Hostname()
	return &NotificationWorker{
		redis:      rdb,
		emailer:    emailer,
		userLookup: userLookup,
		logger:     logger,
		workerID:   fmt.Sprintf("worker-%s-%d", hostname, os.Getpid()),
	}
}

// Run starts the consumer loop. It blocks until ctx is cancelled.
func (w *NotificationWorker) Run(ctx context.Context) {
	if err := w.ensureConsumerGroup(ctx); err != nil {
		w.logger.Error("notification worker: create consumer group", "error", err)
		return
	}

	w.logger.Info("notification worker started", "worker_id", w.workerID)

	claimTicker := time.NewTicker(claimInterval)
	defer claimTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("notification worker stopping")
			return
		case <-claimTicker.C:
			w.reclaimStale(ctx)
		default:
		}

		messages, err := w.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: w.workerID,
			Streams:  []string{streamName, ">"},
			Count:    batchSize,
			Block:    blockTimeout,
		}).Result()

		if err != nil {
			if isContextDone(ctx) {
				return
			}
			// NOGROUP means the group was deleted; recreate and retry.
			if isNOGROUP(err) {
				_ = w.ensureConsumerGroup(ctx)
				continue
			}
			// Blocking read timeout returns redis.Nil — not a real error.
			if err != redis.Nil {
				w.logger.Error("notification worker: xreadgroup", "error", err)
				time.Sleep(time.Second)
			}
			continue
		}

		for _, stream := range messages {
			for _, msg := range stream.Messages {
				w.handleMessage(ctx, msg)
			}
		}
	}
}

func (w *NotificationWorker) handleMessage(ctx context.Context, msg redis.XMessage) {
	payload, ok := msg.Values["payload"].(string)
	if !ok {
		w.logger.Error("notification worker: missing payload field", "msg_id", msg.ID)
		w.ack(ctx, msg.ID) // malformed; ack to discard
		return
	}

	var event events.TaskChangedEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		w.logger.Error("notification worker: unmarshal event", "error", err, "msg_id", msg.ID)
		w.ack(ctx, msg.ID) // unparseable; discard
		return
	}

	if event.AssigneeID == "" {
		w.ack(ctx, msg.ID)
		return
	}

	userName, userEmail, err := w.userLookup.GetUserByID(ctx, event.AssigneeID)
	if err != nil {
		w.logger.Warn("notification worker: assignee not found", "assignee_id", event.AssigneeID, "msg_id", msg.ID)
		w.ack(ctx, msg.ID) // user deleted; discard
		return
	}

	if err := w.emailer.Send(userName, userEmail, event); err != nil {
		// Do NOT ack — message stays in PEL for retry.
		w.logger.Error("notification worker: send email",
			"error", err,
			"assignee_email", userEmail,
			"task_id", event.TaskID,
			"msg_id", msg.ID,
		)
		return
	}

	w.logger.Info("notification worker: email sent",
		"assignee_email", userEmail,
		"task_id", event.TaskID,
		"change_kind", event.ChangeKind,
	)
	w.ack(ctx, msg.ID)
}

func (w *NotificationWorker) ack(ctx context.Context, msgID string) {
	if err := w.redis.XAck(ctx, streamName, groupName, msgID).Err(); err != nil {
		w.logger.Error("notification worker: xack failed", "msg_id", msgID, "error", err)
	}
}

// reclaimStale uses XAUTOCLAIM to pull any PEL messages that have been idle
// longer than claimMinIdle back to this worker so they are retried.
func (w *NotificationWorker) reclaimStale(ctx context.Context) {
	msgs, _, err := w.redis.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   streamName,
		Group:    groupName,
		Consumer: w.workerID,
		MinIdle:  claimMinIdle,
		Start:    "0-0",
		Count:    batchSize,
	}).Result()
	if err != nil {
		if !isContextDone(ctx) {
			w.logger.Error("notification worker: xautoclaim", "error", err)
		}
		return
	}
	for _, msg := range msgs {
		w.handleMessage(ctx, msg)
	}
}

func (w *NotificationWorker) ensureConsumerGroup(ctx context.Context) error {
	err := w.redis.XGroupCreateMkStream(ctx, streamName, groupName, "$").Err()
	if err != nil && !isGroupExistsErr(err) {
		return err
	}
	return nil
}

func isGroupExistsErr(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "BUSYGROUP Consumer Group name already exists"
}

func isNOGROUP(err error) bool {
	if err == nil {
		return false
	}
	return len(err.Error()) >= 7 && err.Error()[:7] == "NOGROUP"
}

func isContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
