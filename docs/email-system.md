# Plan: Scalable Task Change Email Notifications via Redis Streams

## Context

When a task's status, priority, assignee, or due date changes, the assigned user should receive an email. The notification path must be async, durable, and horizontally scalable — the HTTP handler must not block on email delivery, and events must survive server restarts.

Choices confirmed by user:
- **Message broker**: Redis Streams (already in `.env.example`, just unused)
- **Email provider**: SMTP (generic)
- **Trigger events**: status change, priority change, task assigned, due date change

---

## Architecture Overview

```
HTTP Handler
    │
    ▼
TaskService.UpdateTask()
    │   (captures old task state before update, diffs fields)
    ▼
EventPublisher.Publish(TaskChangedEvent) ──► Redis Stream: taskflow:task:events
                                                         │
                                            ┌────────────▼───────────────┐
                                            │   NotificationWorker       │
                                            │   (XREADGROUP loop)        │
                                            │   • fetch assignee email   │
                                            │   • render email template  │
                                            │   • send via SMTP          │
                                            │   • XACK on success        │
                                            │   • retry up to N times    │
                                            │   • XCLAIM stale messages  │
                                            └────────────────────────────┘
```

**Why Redis Streams over Pub/Sub**: Streams are persistent — events survive worker restarts. Consumer groups enable at-least-once delivery with per-message acknowledgement. The Pending Entry List (PEL) lets a restarted worker re-claim unacknowledged messages.

---

## New Files to Create

| File | Purpose |
|------|---------|
| `backend/internal/events/events.go` | `TaskChangedEvent` struct + field change constants |
| `backend/internal/events/publisher.go` | `EventPublisher` interface + `NoopPublisher` (used in tests) |
| `backend/internal/events/redis.go` | `RedisStreamPublisher` — XADD to Redis Stream |
| `backend/internal/email/smtp.go` | `SMTPSender` — dial, auth, send via `net/smtp` (stdlib) |
| `backend/internal/worker/notification.go` | Consumer loop: XREADGROUP → email → XACK |

No new DB migration needed — all state lives in Redis Streams.

---

## Files to Modify

### 1. `backend/internal/store/store.go`
Add `GetTask(ctx, taskID string) (Task, error)` method to fetch the current task state before an update. Uses a simple `SELECT … WHERE id = $1`.

### 2. `backend/internal/service/repositories.go`
Add `GetTask(ctx context.Context, taskID string) (store.Task, error)` to the `taskStore` interface.

### 3. `backend/internal/service/tasks.go`
- Add `EventPublisher` field to `TaskService` struct and constructor
- In `UpdateTask`: fetch old task with `s.repo.GetTask()` before calling `s.repo.UpdateTask()`
- After successful update, diff old vs new (status, priority, assigneeID, dueDate)
- Publish a `TaskChangedEvent` if any watched field changed (async: fire in goroutine, log error)
- In `CreateTask`: publish `TaskAssignedEvent` if assignee is set

### 4. `backend/internal/config/config.go`
Add fields and loading for:
```go
RedisURL        string  // REDIS_URL, default "redis://localhost:6379"
SMTPHost        string  // SMTP_HOST (required in production)
SMTPPort        int     // SMTP_PORT, default 587
SMTPUsername    string  // SMTP_USERNAME
SMTPPassword    string  // SMTP_PASSWORD
SMTPFromAddress string  // SMTP_FROM, default "noreply@taskflow.app"
NotificationsEnabled bool  // NOTIFICATIONS_ENABLED, default false in dev
```

### 5. `backend/cmd/api/main.go`
Wire up:
1. Create Redis client from `cfg.RedisURL`
2. Create `email.SMTPSender` from SMTP config
3. Create `events.RedisStreamPublisher`
4. Inject publisher into `service.NewTaskService(st, publisher)`
5. Create `worker.NotificationWorker` and start it in a goroutine
6. On shutdown signal, call `worker.Stop()` before `httpServer.Shutdown()`

### 6. `docker-compose.yml`
Add Redis 7 Alpine service:
```yaml
redis:
  image: redis:7-alpine
  container_name: taskflow-redis
  ports:
    - "${REDIS_PORT:-6379}:6379"
  volumes:
    - redis_data:/data
  command: redis-server --appendonly yes   # enable AOF persistence
  restart: unless-stopped
```
Add `redis_data` to volumes section. Make backend service depend on redis being healthy.

### 7. `.env.example`
Uncomment/add:
```
REDIS_URL=redis://localhost:6379
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=noreply@taskflow.app
NOTIFICATIONS_ENABLED=false
```

### 8. `backend/go.mod`
Add `github.com/redis/go-redis/v9` (run `go get github.com/redis/go-redis/v9`).

---

## Event Struct

```go
// backend/internal/events/events.go
type ChangeKind string

const (
    ChangeKindStatus   ChangeKind = "status_changed"
    ChangeKindPriority ChangeKind = "priority_changed"
    ChangeKindAssignee ChangeKind = "assignee_changed"
    ChangeKindDueDate  ChangeKind = "due_date_changed"
)

type TaskChangedEvent struct {
    EventID    string     // uuid
    TaskID     string
    ProjectID  string
    TaskTitle  string
    AssigneeID string     // who to notify
    ChangeKind ChangeKind
    OldValue   string
    NewValue   string
    ChangedAt  time.Time
}
```

Events are serialized as JSON and stored in the Redis Stream field `payload`.

---

## Worker Logic (notification.go)

```
On start:
  XGROUP CREATE taskflow:task:events notif-workers $ MKSTREAM (idempotent)

Poll loop (blocking XREADGROUP, 5s timeout):
  1. XREADGROUP GROUP notif-workers worker-{hostname} COUNT 10 BLOCK 5000ms STREAMS taskflow:task:events >
  2. For each message:
     a. Decode TaskChangedEvent from JSON
     b. Fetch assignee email: store.GetUserByID(assigneeID)
     c. Compose email (subject + body via text/template)
     d. Send via SMTPSender
     e. On success: XACK taskflow:task:events notif-workers <msgID>
     f. On error: log; do NOT ack (message stays in PEL for retry)
  3. Periodically (every 60s): XAUTOCLAIM stale PEL entries older than 5min back to self

Shutdown: cancel context → blocking read returns → goroutine exits cleanly
```

---

## Scalability Properties

| Property | How Achieved |
|---|---|
| Non-blocking HTTP | Events published in goroutine after response is already written |
| At-least-once delivery | Redis Streams PEL + XACK only on success |
| Horizontal scaling | Consumer groups — run multiple workers, each gets distinct messages |
| Restart safety | Redis AOF persistence + unacknowledged messages stay in PEL |
| Backpressure | XLEN check before publish; drop + log if stream too large |
| No new DB tables | Event state lives entirely in Redis |

---

## Verification Steps

1. **Unit**: `TaskService.UpdateTask` emits event when status/priority/assignee/due_date changes; `NoopPublisher` used in existing tests (no test changes needed for passing tests).
2. **Integration (docker compose up)**:
   - Set `NOTIFICATIONS_ENABLED=true`, configure a local SMTP server (e.g. `mailhog`)
   - Update a task's status via `PATCH /tasks/{id}` → check mailhog inbox for email
   - Kill the worker mid-delivery → restart → verify email eventually delivered (PEL re-claim)
3. **Redis inspection**: `XLEN taskflow:task:events`, `XPENDING taskflow:task:events notif-workers - + 10`
