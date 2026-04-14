# Backend Production Checklist

Track backend hardening work here and mark items complete as they land.

## Security And Authorization

- [x] Enforce authorization for `PATCH /tasks/{id}` so arbitrary authenticated users cannot update tasks they do not control.
- [x] Centralize project/task access rules in backend logic instead of scattering authorization checks across handlers.
- [x] Align task write permissions across create, update, and delete flows.

## API Semantics And Validation

- [x] Redesign task PATCH semantics so omitted fields, explicit clears, and updates are distinct operations.
- [x] Allow clearing nullable task fields such as `assignee_id`, `due_date`, and optional text fields in a controlled way.
- [x] Tighten request decoding and validation behavior for production-grade API usage.
- [x] Standardize API error responses with stable machine-readable error codes.

## Runtime Hardening

- [x] Add production-grade HTTP server timeouts beyond `ReadHeaderTimeout`.
- [x] Add request body size limits for JSON endpoints.
- [x] Add panic recovery middleware.
- [x] Add request IDs to logs and responses for traceability.
- [x] Make DB pool sizing and connection lifetime configurable.

## Configuration

- [x] Fail fast on invalid numeric configuration instead of silently falling back.
- [x] Validate security-sensitive configuration such as minimum JWT secret strength.
- [x] Separate production-safe config validation from local convenience defaults.

## Persistence And Architecture

- [x] Reduce handler-side authorization/data-loading logic by moving access-aware operations into store/service methods.
- [x] Split oversized backend files or otherwise improve code organization for maintainability.
- [x] Improve store error mapping for malformed identifiers and forbidden/not-found distinctions.

## Database

- [x] Review and add indexes needed for common authorization and filtering paths.
- [x] Add stronger schema-level constraints where they improve invariants.

## Testing And Quality Gates

- [x] Add unit tests for config validation and auth helpers.
- [x] Add handler tests for auth, authorization, and validation paths.
- [x] Add store tests for critical mutation and access behaviors.
- [x] Ensure backend tests can run with local writable caches in the current dev environment.

## Production Gaps From Review

### 1. Authorization At The Write Boundary

- [x] Move permission-sensitive writes such as task creation, task updates, and project mutations into store or service methods that enforce authorization in the same transaction/query as the write.
- [x] Remove time-of-check/time-of-use gaps created by handler-side `GetProject` or `CanAccessProject` checks before mutations.
- [x] Add permission-focused integration tests that prove unauthorized users cannot mutate data even under concurrent access patterns.

### 2. Authentication And Session Hardening

- [x] Introduce refresh token logic with rotation on every refresh and server-side revocation support.
- [x] Add a persisted session or token table for refresh tokens, token family tracking, logout, and compromise handling.
- [x] Add `jti`, `iss`, `aud`, and stronger claim validation to issued access tokens.
- [x] Plan for signing key rotation instead of relying on one long-lived shared HS256 secret.
- [x] Define access token TTL and refresh token TTL separately and document the session lifecycle.

### 3. Health Checks And Operational Readiness

- [ ] Split health endpoints into liveness and readiness instead of returning a static healthy response for all cases.
- [ ] Make readiness verify critical dependencies such as database connectivity.
- [ ] Add container healthchecks, restart policies, and runtime wiring for the backend service in deployment manifests.
- [ ] Add migration startup strategy or deployment sequencing so schema drift does not break boot or requests.

### 4. Layering, DI, And Decoupling

- [x] Introduce a service or use-case layer between HTTP handlers and the store for orchestration, policy enforcement, and transactions.
- [x] Keep transport DTOs, domain models, and persistence models separate instead of sharing store structs directly with JSON tags.
- [x] Depend on interfaces at boundaries for token issuing, user/project/task services, clock, and ID generation where testing or replacement benefits are real.
- [x] Move validation, authorization policy, and cross-entity coordination out of handlers so handlers stay transport-focused.
- [x] Refactor `httpapi.Dependencies` and `Server` construction so the HTTP layer depends on service interfaces rather than the concrete `*store.Store`.
- [x] Review dependency injection wiring and remove direct concrete coupling where interfaces would reduce package-level entanglement.

### 5. Observability

- [ ] Expand request logs to include route template, authenticated user id, remote IP, user agent, and response size.
- [ ] Capture panic stacks and structured error context instead of only logging the recovered value.
- [ ] Add metrics for request rates, latencies, status codes, auth failures, and database pool health.
- [ ] Add tracing hooks and request context propagation for cross-layer diagnostics.

### 6. Database Invariants And Data Integrity

- [x] Enforce case-insensitive email uniqueness with `CITEXT` or a unique index on `lower(email)`.
- [x] Move `updated_at` maintenance to database triggers or another centralized persistence mechanism.
- [x] Review remaining schema invariants that currently live only in application code and push the critical ones into constraints.
- [x] Add migration tests or verification steps that validate constraints and indexes against a real Postgres instance.

### 7. Production-Grade Testing

- [ ] Add integration tests against a real Postgres database instead of relying only on SQL mocks.
- [ ] Add end-to-end tests for register, login, refresh, logout, and authenticated CRUD flows.
- [ ] Add authorization matrix tests across owner, creator, assignee, and unrelated-user roles.
- [ ] Add migration smoke tests in CI so schema changes are exercised before deployment.

## 8. Dependency Injection Assessment

- [x] Dependency injection is partially implemented, not fully production-grade.
- [x] `cmd/api/main.go` does composition cleanly, which is good, but the HTTP layer still depends on concrete types like `*store.Store` and concrete auth/token implementations.
- [x] Current wiring exists, but the boundaries are still tightly coupled.
- [x] Replace direct `*store.Store` usage in handlers with narrower service interfaces so the HTTP layer is not coupled to the persistence implementation.
- [x] Inject cross-cutting dependencies such as clock, request ID generator, and logger abstractions where deterministic testing is valuable.
- [x] Keep composition rooted in `cmd/api/main.go` and avoid hidden service locators or package-level singletons.
- [x] Audit constructor dependencies so each layer only receives the collaborators it actually needs.

## 9. Further Decoupling Opportunities

- [x] Add a service or use-case layer between handlers and store.
- [x] Make handlers depend on service interfaces, not `*store.Store`.
- [x] Separate HTTP DTOs from DB models.
- [x] Move authz policy into dedicated policy or service code.
- [x] Isolate JWT and session logic behind an auth service.
- [x] Inject clock and ID generator dependencies where deterministic tests help.
- [x] Extract authorization policy logic into dedicated policy or service types instead of spreading role checks across handlers and SQL queries.
- [x] Separate query-oriented read models from command-oriented write flows where response assembly is becoming handler-heavy.
- [x] Introduce repository-style interfaces only where they improve testability or multiple implementations are plausible; avoid interface proliferation without a boundary need.
- [x] Isolate token/session management behind an auth service rather than coupling handlers directly to JWT mechanics.

## Single Points Of Failure To Address

- [ ] Remove the authentication single point of failure created by one long-lived signing secret with no rotation path.
- [ ] Remove the session-management single point of failure by adding revocable refresh-token persistence instead of stateless access tokens only.
- [ ] Reduce infrastructure single points of failure by planning for managed Postgres backups, failover, and monitored connection exhaustion handling.
- [ ] Avoid a single application instance assumption in deployment design; ensure horizontal scaling works without in-memory auth or request state.
- [ ] Ensure background responsibilities such as migrations, cleanup, or token revocation jobs are not tied to one fragile process with no operational fallback.
