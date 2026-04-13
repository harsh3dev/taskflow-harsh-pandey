# TaskFlow

A full-stack task management system with authentication, projects, tasks, and team assignment — built as a take-home engineering project.

---

## 1. Overview

TaskFlow lets users register, log in, create projects, add tasks, and assign them to team members. It ships as a single `docker compose up` command that stands up PostgreSQL, runs migrations, seeds demo data, and starts both the API and React frontend.

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

## 2. Architecture Decisions

### Backend

**No framework.** Go's `net/http` with 1.22 pattern matching handles routing without adding a dependency. The route surface is small enough that a framework would be overhead, not help.

**Layered architecture.** Transport (`httpapi`) → Service (business logic, validation, auth policy) → Store (SQL). Each layer depends only on interfaces defined by the layer above it — `httpapi` consumes `projectService`/`taskService`/`authService` interfaces; services consume narrow repository interfaces. This lets each layer be tested in isolation.

**SQL-first, no ORM.** Every query is explicit SQL. Schema lives in numbered migration files compatible with `golang-migrate`. This keeps the data model honest and avoids hidden N+1 behaviour.

**Refresh token rotation.** On every `/auth/refresh`, the old token is invalidated and a new one is issued. The server persists token hashes in `auth_sessions`. Replay detection is implemented: if an already-rotated token is presented, the entire token family is revoked.

**Authorization at the write boundary.** PATCH/DELETE SQL queries include the `owner_id`/`creator_id` check directly in the `WHERE` clause — no separate pre-check that could be raced. The service layer enforces read-side access rules before returning data.

**Intentional omissions:** rate limiting, WebSocket real-time updates, and Redis caching were scoped out. The architecture makes each straightforward to add.

### Frontend

**Zustand for auth, hook controllers for server state.** Auth state is tiny and global (token + user). Server state (projects, tasks) is fetched and stored by page-scoped controllers that call the API directly — no React Query or SWR dependency needed at this scale.

**Optimistic UI for task status.** Dragging a task card or selecting a new status from the dropdown applies the change immediately and reverts on API error — so the board feels instant.

**Component hierarchy.** Page → ViewModel hook (derives display data) → UI components. The ViewModel produces `visibleColumns`, `userMap`, and `assigneeOptions` from raw store state. Components stay presentation-only.

**Tradeoffs:** no end-to-end tests; no pagination UI (the API supports it, the frontend always fetches up to 1 000 tasks per project).

---

## 3. Running Locally

> Requires Docker Desktop. No other tools needed.

```bash
git clone https://github.com/harshpn/taskflow-harsh-pandey
cd taskflow-harsh-pandey

cp .env.example .env

# First run — build images, apply migrations, seed, start all services
docker compose up --build

# Every subsequent run
docker compose up
```

**App:** http://localhost:3000  
**API:** http://localhost:8080

Hot reload is active in this mode:
- **Backend** — `air` recompiles on every `.go` save (~1 s restart)
- **Frontend** — Vite HMR pushes changes to the browser without a full reload

### Production build

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build
```

---

## 4. Running Migrations

Migrations run **automatically** on `docker compose up` via the `migrate` service (uses the official `golang-migrate/migrate` Docker image). The `backend` service only starts after migrations and seed complete.

To run migrations manually against a running database:

```bash
# Up
docker compose run --rm migrate

# Down (rolls back all)
docker compose run --rm migrate \
  -path=/migrations \
  -database "postgres://taskflow:taskflow@postgres:5432/taskflow?sslmode=disable" \
  down -all
```

---

## 5. Test Credentials

The seed script inserts a ready-to-use account:

```
Email:    test@example.com
Password: password123
```

---

## 6. API Reference

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

**Pagination response shape:**
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

**Task PATCH body** — all fields optional; send only what changes:
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
| GET | `/users?q=search` | List users (for assignee picker) |

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

## 7. What I'd Do With More Time

**Testing**
- Integration tests against a real Postgres instance for the full CRUD lifecycle
- Authorization matrix tests: owner vs. creator vs. assignee vs. unrelated user for every mutation
- Frontend component tests with Testing Library

**Production hardening**
- Liveness/readiness split on `/health` — readiness should ping the DB
- Key rotation path for JWT signing keys (the infrastructure is present; UI and tooling are not)
- Rate limiting on auth endpoints

**Features**
- Real-time task updates via SSE — the backend request-ID plumbing and structured logging make it straightforward to wire up
- Pagination controls in the frontend (the API already returns `pagination` metadata)
- Project stats chart — a stacked bar showing done/in_progress/todo trends over time

**DevEx**
- `make` targets for common tasks (build, test, migrate, seed)
- Pre-commit hook to run `go vet` and `tsc --noEmit` before every commit
