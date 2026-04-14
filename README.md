# TaskFlow

A full-stack task management system with authentication, projects, tasks, and team assignment — built as a take-home engineering project.

---

## 1. Overview

TaskFlow lets users register, log in, create projects, add tasks, and assign them to team members. It ships as a single `docker compose up` command that stands up PostgreSQL, runs migrations, seeds demo data, and starts both the API and React frontend.

**Key Features:**
- **Secure Authentication**: JWT-based authentication with refresh token rotation and replay detection.
- **Projects & Tasks**: Full CRUD for projects and associated tasks. Focus on access control boundaries.
- **Interactive Board**: Drag-and-drop task ordering and status updates using `@dnd-kit/core`.
- **Optimistic Updates**: Immediate UI updates that revert automatically on API failures.
- **Background Processes**: Asynchronous worker thread handling SMTP email deliveries out-of-band.
- **Data Integrity**: Enforced via strict PostgreSQL schemas, parameterization, and indices rather than ORMs.

**Tech stack**

| Layer | Choice |
|---|---|
| Backend | Go 1.22 · `net/http` (no framework) |
| Database | PostgreSQL 16 · SQL-first migrations |
| Auth | JWT access tokens + refresh token rotation |
| Frontend | React 18 · TypeScript · Vite · Tailwind CSS v4 |
| State | Zustand (auth) · controller hooks (server data) |
| UI | shadcn/ui component primitives |
| DnD | @dnd-kit/core |
| Infra | Docker Compose · golang-migrate |

---

## 2. Architecture decisions

### Backend

![alt text](image.png)

**No framework.** Go's `net/http` with 1.22 pattern matching handles routing without adding a dependency. The route surface is small enough that a framework would be overhead, not help.

**Layered architecture.** Transport (`httpapi`) → Service (business logic, validation, auth policy) → Store (SQL). Each layer depends only on interfaces defined by the layer above it — `httpapi` consumes `projectService`/`taskService`/`authService` interfaces; services consume narrow repository interfaces. This lets each layer be tested in isolation.

**SQL-first, no ORM.** Every query is explicit SQL. Schema lives in numbered migration files compatible with `golang-migrate`. This keeps the data model honest and avoids hidden N+1 behaviour.

**Refresh token rotation.** On every `/auth/refresh`, the old token is invalidated and a new one is issued. The server persists token hashes in `auth_sessions`. Replay detection is implemented: if an already-rotated token is presented, the entire token family is revoked.

**Authorization at the write boundary.** PATCH/DELETE SQL queries include the `owner_id`/`creator_id` check directly in the `WHERE` clause — no separate pre-check that could be raced. The service layer enforces read-side access rules before returning data.

**Intentional omissions:** rate limiting, WebSocket real-time updates, and Redis caching were scoped out. The architecture makes each straightforward to add.

### Email / notification system

![alt text](image-1.png)

I wanted task assignment notifications to feel reliable, not just "fire and forget and hope."

**The naive approach** (`go sendEmail(...)`) silently drops notifications if the process crashes after the DB write but before the goroutine finishes. That felt wrong to me for something users would actually notice missing.

**Why Redis Streams?** Events get written to a durable log (`taskflow:task:events`) and stay there until a consumer explicitly ACKs them. That means notifications survive restarts and deploys — the worker picks up where it left off.

**How the flow works:**
- `TaskService` is the only producer. On task create (with an assignee) or update, it diffs old vs. new state and publishes a `TaskChangedEvent` per changed field (`status`, `priority`, `assignee_id`, `due_date`). Publishing happens in a detached goroutine with a 5-second timeout so a slow Redis never blocks the HTTP response.
- A `NotificationWorker` goroutine reads from the stream via `XREADGROUP` (consumer group `notif-workers`, batch of 10, 5-second block). It looks up the assignee, sends the email via stdlib `net/smtp`, then ACKs. If the send fails, it skips the ACK — the message stays pending and gets retried automatically.
- A 60-second ticker runs `XAUTOCLAIM` to reclaim messages that have been sitting unacknowledged for more than 5 minutes. That covers the case where a worker crashes mid-send.

**A few other details:**
- `NOTIFICATIONS_ENABLED=false` (the default) wires in a `NoopPublisher` — no Redis client or worker goroutine is started at all.
- The SMTP sender skips `PlainAuth` when no credentials are configured, so it works out of the box against MailHog in dev.
- The stream is soft-capped at ~10,000 entries via approximate `MAXLEN` trimming on every `XADD`.

### Frontend

![alt text](image-2.png)

**Zustand for auth, hook controllers for server state.** Auth state is tiny and global. Server state (projects, tasks) is fetched and cached by page-scoped controller hooks that call the API directly. I didn't reach for React Query or SWR — at this scale, the extra dependency isn't worth it.

**Optimistic task updates.** Dragging a card or changing a status dropdown applies the change immediately and reverts on API error. The board feels instant rather than sluggish.

**Component hierarchy.** Page → ViewModel hook → UI components. The ViewModel derives `visibleColumns`, `userMap`, and `assigneeOptions` from raw store state. Components stay presentation-only — they receive data and fire callbacks, nothing else.

**Honest tradeoffs:** no end-to-end tests, and the frontend always fetches up to 1,000 tasks per project (the API returns pagination metadata; the UI just doesn't use it yet).

---

## 3. Running locally

> Requires Docker Desktop. Nothing else.

```bash
git clone https://github.com/harshpn/taskflow-harsh-pandey
cd taskflow-harsh-pandey

cp .env.example .env

# First run — builds images, applies migrations, seeds data, starts everything
docker compose up --build

# Every run after that
docker compose up
```

**App:** http://localhost:3000  
**API:** http://localhost:8080

Hot reload works in dev mode:
- **Backend** — `air` recompiles on every `.go` save (~1s restart)
- **Frontend** — Vite HMR pushes changes without a full reload

### Production build

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build
```

---

## 4. Migrations

Migrations run automatically on `docker compose up` via the `migrate` service. The backend service waits until migrations and seed complete before starting.

To run them manually against a running database:

```bash
# Up
docker compose run --rm migrate

# Down (rolls back everything)
docker compose run --rm migrate \
  -path=/migrations \
  -database "postgres://taskflow:taskflow@postgres:5432/taskflow?sslmode=disable" \
  down -all
```

---

## 5. Test credentials

The seed script creates a ready-to-use account:

```
Email:    test@example.com
Password: password123
```

---

## 6. API reference

Base URL: `http://localhost:8080`  
All endpoints accept and return `application/json`.  
Protected endpoints require `Authorization: Bearer <access_token>`.

### Auth

| Method | Path | Description |
|---|---|---|
| POST | `/auth/register` | Register — returns access + refresh tokens |
| POST | `/auth/login` | Login — returns access + refresh tokens |
| POST | `/auth/refresh` | Rotate refresh token — returns new pair |
| POST | `/auth/logout` | Revoke refresh token |

**Register / Login request:**
```json
{ "name": "Jane Doe", "email": "jane@example.com", "password": "secret123" }
```

**Auth response:**
```json
{
  "access_token": "<jwt>",
  "refresh_token": "<opaque>",
  "token_type": "Bearer",
  "expires_in_seconds": 900,
  "user": { "id": "uuid", "name": "Jane Doe", "email": "jane@example.com", "created_at": "..." }
}
```

### Projects

| Method | Path | Description |
|---|---|---|
| GET | `/projects?page=1&limit=50` | List accessible projects (owned or assigned) |
| POST | `/projects` | Create project |
| GET | `/projects/:id` | Get project + tasks |
| PATCH | `/projects/:id` | Update name/description (owner only) |
| DELETE | `/projects/:id` | Delete project + all tasks (owner only) |
| GET | `/projects/:id/stats` | Task counts by status and assignee |

**Pagination response:**
```json
{
  "projects": [...],
  "pagination": { "page": 1, "limit": 50, "total": 12 }
}
```

**Stats response:**
```json
{
  "status_counts": { "todo": 2, "in_progress": 1, "done": 3 },
  "assignee_counts": [{ "user_id": "uuid", "name": "Jane", "count": 3 }]
}
```

### Tasks

| Method | Path | Description |
|---|---|---|
| GET | `/projects/:id/tasks?status=todo&assignee=uuid&page=1&limit=50` | List tasks with filters + pagination |
| POST | `/projects/:id/tasks` | Create task |
| PATCH | `/tasks/:id` | Update task fields |
| DELETE | `/tasks/:id` | Delete task (owner or creator only) |

**Task PATCH body** — all fields optional; send only what changed:
```json
{
  "title": "Updated title",
  "status": "done",
  "priority": "high",
  "assignee_id": "uuid",
  "due_date": "2026-05-01",
  "description": "New description"
}
```

To clear a nullable field, send `null`:
```json
{ "assignee_id": null, "due_date": null }
```

### Users

| Method | Path | Description |
|---|---|---|
| GET | `/users?q=search` | List users (for the assignee picker) |

### Error responses

```json
// 400 validation
{ "code": "validation_failed", "error": "validation failed", "fields": { "email": "is required" } }

// 401
{ "code": "unauthorized", "error": "unauthorized" }

// 403
{ "code": "forbidden", "error": "forbidden" }

// 404
{ "code": "not_found", "error": "not found" }
```

---

## 7. What I'd do with more time

**Testing**
- Integration tests against a real Postgres instance for the full CRUD lifecycle — the kind that would have caught the auth edge cases I had to think through manually
- An authorization matrix test: owner vs. creator vs. assignee vs. unrelated user for every mutation
- Frontend component tests with Testing Library

**Production hardening**
- Split `/health` into liveness and readiness — readiness should actually ping the DB before reporting healthy
- A key rotation path for JWT signing keys (the infrastructure is there; the tooling and UI aren't)
- Rate limiting on auth endpoints

**Features I wanted to add**
- Real-time task updates via SSE — the structured logging and request-ID plumbing are already in place; it's mostly wiring
- Pagination controls in the frontend (the API already returns the metadata; the UI just ignores it)
- A project stats chart — stacked bar, done/in_progress/todo over time

**Developer experience**
- `make` targets for the common operations (build, test, migrate, seed)
- A pre-commit hook to run `go vet` and `tsc --noEmit` before every push
