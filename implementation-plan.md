**DO NOT COMMIT THIS FILE**
# TaskFlow Implementation Checklist

This checklist converts the assignment requirements into a practical delivery plan. The goal is to finish the core submission first, then add bonuses only if time remains.

## Phase 0 — Project Setup & Foundations

### Repository and structure
- [x] Create a monorepo structure with `/backend` and `/frontend`
- [x] Add root `docker-compose.yml`
- [x] Add root `.env.example` with all required variables
- [x] Ensure no secrets are committed

### Technical decisions
- [x] Finalize backend stack and note it for the README
- [x] Finalize frontend stack and note it for the README
- [x] Choose migration tool
- [x] Choose logging library
- [x] Choose UI library or custom component approach

### Definition of done for setup
- [x] Project boots locally in developer mode
- [x] Environment variables are documented
- [x] Folder structure is stable enough to build on

---

## Phase 1 — Database Schema, Migrations, and Seed Data

### Required schema
- [x] Create `users` table
- [x] Create `projects` table
- [x] Create `tasks` table

### Required fields
- [x] `users`: `id`, `name`, `email`, `password`, `created_at`
- [x] `projects`: `id`, `name`, `description`, `owner_id`, `created_at`
- [x] `tasks`: `id`, `title`, `description`, `status`, `priority`, `project_id`, `assignee_id`, `due_date`, `created_at`, `updated_at`

### Constraints and relationships
- [x] Use UUID primary keys
- [x] Enforce unique email on users
- [x] Add foreign key: `projects.owner_id -> users.id`
- [x] Add foreign key: `tasks.project_id -> projects.id`
- [x] Add foreign key: `tasks.assignee_id -> users.id`
- [x] Define task status enum: `todo | in_progress | done`
- [x] Define task priority enum: `low | medium | high`

### Performance-minded indexes
- [x] Add index on `tasks.project_id`
- [x] Add index on `tasks.assignee_id`
- [x] Add composite index on `tasks(project_id, status)`

### Migrations and seed
- [x] Create explicit up migration files
- [x] Create explicit down migration files
- [x] Avoid ORM auto-migrate behavior
- [x] Add seed script or SQL file
- [x] Seed 1 test user with known credentials
- [x] Seed 1 project
- [x] Seed 3 tasks with different statuses

### Definition of done for data layer
- [x] Fresh database can be created from migrations only
- [x] Database can be rolled back with down migrations
- [x] Seed data is available for reviewers immediately

---

## Phase 2 — Backend Core API

### Server foundation
- [x] Set up HTTP server with route structure
- [x] Add JSON-only response handling
- [x] Add graceful shutdown on `SIGTERM`
- [x] Add structured request/error logging

### Authentication
- [x] Implement `POST /auth/register`
- [x] Implement `POST /auth/login`
- [x] Hash passwords with bcrypt using cost `>= 12`
- [x] Issue JWT with 24-hour expiry
- [x] Include `user_id` and `email` in JWT claims
- [x] Add auth middleware for protected routes

### Auth and error handling rules
- [x] Return `400` for validation errors with structured field errors
- [x] Return `401` for unauthenticated requests
- [x] Return `403` for authenticated but unauthorized actions
- [x] Return `404` for missing resources
- [x] Keep error responses consistent across endpoints
- [x] Validate `assignee_id` before task writes and return structured field errors for bad or unknown user IDs

### Projects API
- [x] Implement `GET /projects`
- [x] Ensure it returns projects owned by the user or linked via assigned tasks
- [x] Implement `POST /projects`
- [x] Implement `GET /projects/:id`
- [x] Include project details and its tasks
- [x] Implement `PATCH /projects/:id`
- [x] Restrict project updates to the owner
- [x] Implement `DELETE /projects/:id`
- [x] Restrict project deletion to the owner
- [x] Delete project tasks when a project is removed

### Tasks API
- [x] Implement `GET /projects/:id/tasks`
- [x] Support `status` filter
- [x] Support `assignee` filter
- [x] Implement `POST /projects/:id/tasks`
- [x] Implement `PATCH /tasks/:id`
- [x] Allow updates to title, description, status, priority, assignee, and due date
- [x] Implement `DELETE /tasks/:id`
- [x] Restrict delete to project owner or task creator only
- [x] Implement protected `GET /users`
- [x] Support optional user directory search via `?q=`

### Definition of done for backend core
- [x] Core auth, project, and task flows work end to end
- [x] All protected routes require bearer token auth
- [x] Permission boundaries match the assignment rules

---

## Phase 3 — Frontend Core Experience

### Application foundation
- [x] Set up React app with TypeScript
- [x] Configure React Router
- [x] Configure API client
- [x] Set up global auth state
- [x] Persist auth across refreshes with localStorage or equivalent

### Authentication UX
- [x] Build register page
- [x] Build login page
- [x] Add client-side validation
- [x] Handle API errors visibly
- [x] Store JWT after login/register
- [x] Add logout flow
- [x] Add protected routes with redirect to `/login`

### Shared layout
- [x] Build navbar
- [x] Show logged-in user's name
- [x] Show logout action in navbar

### Projects experience
- [x] Build projects list page
- [x] Fetch and display accessible projects
- [x] Add create project flow
- [x] Show name, description, and useful summary information
- [x] Add sensible empty state when no projects exist

### Project detail and tasks
- [x] Build project detail page
- [x] Show project tasks clearly
- [x] Support grouped view by status or equivalent task organization
- [x] Add create task UI using modal or side panel
- [x] Add edit task UI
- [x] Add delete task action
- [x] Include fields: title, description, status, priority, assignee, due date
- [x] Populate assignee controls from backend user data instead of manual UUID entry

### Filtering and task interactions
- [x] Add task filter by status
- [x] Add task filter by assignee
- [x] Implement optimistic UI for task status changes
- [x] Revert optimistic update cleanly on API failure

### UX quality requirements
- [x] Visible loading states on async screens and actions
- [x] Visible error states with clear feedback
- [x] Sensible empty states for no tasks and no projects
- [x] No blank screens on failed requests
- [x] No console errors in production build

### Responsive UI
- [x] Validate layout at `375px`
- [x] Validate layout at `1280px`
- [x] Ensure no broken layouts between those sizes

### Definition of done for frontend core
- [x] User can register/login, create a project, create tasks, and manage them
- [x] Auth persists across refreshes
- [x] UX handles loading, empty, and error states properly

---

## Phase 4 — Docker, Runtime, and Delivery Readiness

### Containers
- [ ] Add backend Dockerfile
- [ ] Use multi-stage build for backend Dockerfile
- [ ] Add frontend Dockerfile
- [ ] Add PostgreSQL service to `docker-compose.yml`
- [ ] Add backend service to `docker-compose.yml`
- [ ] Add frontend service to `docker-compose.yml`

### Runtime configuration
- [ ] Read PostgreSQL config from environment variables
- [ ] Wire app services together through Docker networking
- [ ] Ensure migration execution is automatic on container start
- [ ] If not automatic, document exact migration commands in README

### Single-command startup
- [ ] Verify `docker compose up` works from repo root
- [ ] Verify the full stack starts with zero manual steps
- [ ] Verify seed data is accessible after startup

### Definition of done for infra
- [ ] Reviewer can clone, copy `.env.example`, run `docker compose up`, and use the app

---

## Phase 5 — README and Submission Requirements

### README content
- [ ] Add project overview
- [ ] Add tech stack summary
- [ ] Add architecture decisions
- [ ] Explain tradeoffs and intentional omissions
- [ ] Add exact local setup steps from clone to running app
- [ ] Add migration instructions
- [ ] Add seed test credentials
- [ ] Add API reference or link to Postman/Bruno collection
- [ ] Add "What I'd do with more time"

### Submission hygiene
- [ ] Confirm `.env` is ignored
- [ ] Confirm `.env.example` is complete
- [ ] Confirm no hardcoded JWT secret in source
- [ ] Confirm README is present and accurate
- [ ] Confirm repository structure matches assignment expectations

### Final review against automatic disqualifiers
- [ ] App runs with `docker compose up`
- [ ] Migrations are present and usable
- [ ] Passwords are hashed, not plaintext
- [ ] JWT secret comes from environment variables
- [ ] README exists

---

## Phase 6 — Optional Bonus Work

Only start this phase after all required phases are complete and verified.

### Backend bonuses
- [ ] Add pagination to list endpoints with `page` and `limit`
- [ ] Add `GET /projects/:id/stats`
- [ ] Return task counts by status
- [ ] Return task counts by assignee
- [ ] Add at least 3 integration tests

### Frontend bonuses
- [ ] Add drag-and-drop for task movement across status columns
- [ ] Persist task status update after drop
- [ ] Add stats dashboard visualization
- [ ] Add dark mode toggle with persistence
- [ ] Add real-time updates via WebSocket or SSE if backend supports it

### Infra / performance extras
- [ ] Add Redis caching where it genuinely improves repeated reads
- [ ] Add rate limiting middleware

---

## Suggested Execution Order

- [ ] Finish Phase 0 before writing feature code
- [ ] Finish Phase 1 before building API handlers
- [ ] Finish Phase 2 before full frontend integration
- [ ] Finish Phase 3 before Docker polish
- [ ] Finish Phase 4 before README finalization
- [ ] Finish Phase 5 before any bonus work

---

## Reviewer Demo Checklist

- [ ] Register or log in with seed credentials
- [ ] View accessible projects
- [ ] Create a project
- [ ] Open a project detail page
- [ ] Create a task
- [ ] Update task status and other fields
- [ ] Filter tasks by status and assignee
- [ ] Delete a task with correct permissions
- [ ] Restart the app and confirm auth persistence
- [ ] Confirm the app works on mobile and desktop widths
