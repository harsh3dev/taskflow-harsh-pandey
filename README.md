# TaskFlow

## Overview

TaskFlow is a full-stack take-home project for managing projects and tasks with authentication, relational data, and a responsive React UI.

Current repository status:
- Phase 0 foundation is in place
- Phase 1 database migrations and seed data are in place
- Backend and frontend feature implementation are still pending

## Planned Tech Stack

- Backend: Go
- HTTP routing: standard `net/http` with modular handlers
- Database: PostgreSQL
- Migrations: `golang-migrate`-compatible SQL-first migration files
- Logging: Go `slog`
- Frontend: React + TypeScript + Vite
- Routing: React Router
- State: Zustand for auth, React Query for server state
- UI: Tailwind CSS with shadcn/ui
- Infra: Docker Compose

## Architecture Direction

- Monorepo structure with `/backend` and `/frontend`
- SQL-first schema management instead of ORM auto-migration
- Clear separation between transport, domain, and persistence concerns in the backend
- Component-driven frontend with route-level data fetching and shared auth state

## Repository Layout

```text
.
├── backend
│   ├── cmd/api
│   ├── db
│   │   ├── migrations
│   │   └── seeds
│   └── internal
├── frontend
│   ├── public
│   └── src
├── docker-compose.yml
├── .env.example
├── implementation-plan.md
└── requirement.md
```

## Local Setup Foundation

```bash
cp .env.example .env
docker compose up postgres
```

At this stage, `docker-compose.yml` provisions PostgreSQL only. The repository also includes minimal backend and frontend developer scaffolds so later phases can build on a stable structure. Full-stack runtime wiring will be completed in the infrastructure phase.

## Environment Variables

All expected local environment variables are listed in `.env.example`.

Backend auth/session highlights:
- Access tokens and refresh tokens now have separate lifetimes. The backend expects `ACCESS_TOKEN_TTL` and `REFRESH_TOKEN_TTL`.
- Access tokens are JWTs with `jti`, `iss`, and `aud` claims. `JWT_ISSUER` and `JWT_AUDIENCE` control validation.
- Signing keys support rotation through `JWT_SIGNING_KEYS` plus `JWT_ACTIVE_KEY_ID`. `JWT_SECRET` remains as a single-key local fallback.
- Login and register issue both an access token and a refresh token. Refresh tokens are persisted server-side, rotated on every refresh, revocable on logout, and family-wide revocation is triggered if an already-rotated refresh token is reused.

## Next Planned Milestones

- Phase 2: backend API and auth
- Phase 3: frontend app flows
- Phase 4: full Docker runtime
- Phase 5: final README and submission polish
