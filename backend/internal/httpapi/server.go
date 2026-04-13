package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/harshpn/taskflow/internal/auth"
	"github.com/harshpn/taskflow/internal/store"
)

type Dependencies struct {
	Logger       *slog.Logger
	Store        *store.Store
	TokenManager auth.TokenManager
	BcryptCost   int
}

type Server struct {
	logger       *slog.Logger
	store        *store.Store
	tokenManager auth.TokenManager
	bcryptCost   int
}

func NewServer(deps Dependencies) *Server {
	return &Server{
		logger:       deps.Logger,
		store:        deps.Store,
		tokenManager: deps.TokenManager,
		bcryptCost:   deps.BcryptCost,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /auth/register", s.handleRegister)
	mux.HandleFunc("POST /auth/login", s.handleLogin)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /projects", s.handleListProjects)
	protected.HandleFunc("POST /projects", s.handleCreateProject)
	protected.HandleFunc("GET /projects/{id}", s.handleGetProject)
	protected.HandleFunc("PATCH /projects/{id}", s.handleUpdateProject)
	protected.HandleFunc("DELETE /projects/{id}", s.handleDeleteProject)
	protected.HandleFunc("GET /projects/{id}/tasks", s.handleListProjectTasks)
	protected.HandleFunc("POST /projects/{id}/tasks", s.handleCreateTask)
	protected.HandleFunc("GET /users", s.handleListUsers)
	protected.HandleFunc("PATCH /tasks/{id}", s.handleUpdateTask)
	protected.HandleFunc("DELETE /tasks/{id}", s.handleDeleteTask)

	mux.Handle("/projects", auth.Middleware(s.tokenManager, s.writeUnauthorized)(protected))
	mux.Handle("/projects/", auth.Middleware(s.tokenManager, s.writeUnauthorized)(protected))
	mux.Handle("/users", auth.Middleware(s.tokenManager, s.writeUnauthorized)(protected))
	mux.Handle("/users/", auth.Middleware(s.tokenManager, s.writeUnauthorized)(protected))
	mux.Handle("/tasks/", auth.Middleware(s.tokenManager, s.writeUnauthorized)(protected))

	return s.withLogging(s.enforceJSON(mux))
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type authRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	fields := map[string]string{}
	if strings.TrimSpace(req.Name) == "" {
		fields["name"] = "is required"
	}
	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "is required"
	}
	if len(strings.TrimSpace(req.Password)) < 8 {
		fields["password"] = "must be at least 8 characters"
	}
	if len(fields) > 0 {
		writeValidationError(w, fields)
		return
	}

	hash, err := auth.HashPassword(req.Password, s.bcryptCost)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	user, err := s.store.CreateUser(r.Context(), store.CreateUserInput{
		Name:         strings.TrimSpace(req.Name),
		Email:        strings.TrimSpace(req.Email),
		PasswordHash: hash,
	})
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeValidationError(w, map[string]string{"email": "is already registered"})
			return
		}
		s.writeInternalError(w, err)
		return
	}

	token, err := s.tokenManager.IssueToken(user.ID, user.Email)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"user":  user,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	fields := map[string]string{}
	if strings.TrimSpace(req.Email) == "" {
		fields["email"] = "is required"
	}
	if strings.TrimSpace(req.Password) == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		writeValidationError(w, fields)
		return
	}

	user, err := s.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			s.writeUnauthorized(w, "invalid credentials")
			return
		}
		s.writeInternalError(w, err)
		return
	}

	if err := auth.CheckPassword(user.Password, req.Password); err != nil {
		s.writeUnauthorized(w, "invalid credentials")
		return
	}

	token, err := s.tokenManager.IssueToken(user.ID, user.Email)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, "missing authenticated user")
		return
	}

	projects, err := s.store.ListAccessibleProjects(r.Context(), user.UserID)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserFromContext(r.Context()); !ok {
		s.writeUnauthorized(w, "missing authenticated user")
		return
	}

	users, err := s.store.ListUsers(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

type projectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, "missing authenticated user")
		return
	}

	var req projectRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeValidationError(w, map[string]string{"name": "is required"})
		return
	}

	project, err := s.store.CreateProject(r.Context(), user.UserID, strings.TrimSpace(req.Name), strings.TrimSpace(req.Description))
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"project": project})
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, "missing authenticated user")
		return
	}

	projectID := r.PathValue("id")
	projectWithTasks, err := s.store.GetProjectWithTasks(r.Context(), projectID)
	if err != nil {
		s.handleStoreError(w, err)
		return
	}

	if !s.canAccessProject(r.Context(), user.UserID, projectWithTasks.Project, projectWithTasks.Tasks) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	writeJSON(w, http.StatusOK, projectWithTasks)
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, "missing authenticated user")
		return
	}

	var req projectRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	project, err := s.store.UpdateProject(r.Context(), r.PathValue("id"), user.UserID, strings.TrimSpace(req.Name), strings.TrimSpace(req.Description))
	if err != nil {
		s.handleStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"project": project})
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, "missing authenticated user")
		return
	}

	err := s.store.DeleteProject(r.Context(), r.PathValue("id"), user.UserID)
	if err != nil {
		s.handleStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleListProjectTasks(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, "missing authenticated user")
		return
	}

	projectID := r.PathValue("id")
	project, err := s.store.GetProject(r.Context(), projectID)
	if err != nil {
		s.handleStoreError(w, err)
		return
	}

	if !s.canAccessProject(r.Context(), user.UserID, project, nil) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	filters := store.TaskFilters{
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		AssigneeID: strings.TrimSpace(r.URL.Query().Get("assignee")),
	}
	tasks, err := s.store.ListProjectTasks(r.Context(), projectID, filters)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

type taskRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, "missing authenticated user")
		return
	}

	projectID := r.PathValue("id")
	project, err := s.store.GetProject(r.Context(), projectID)
	if err != nil {
		s.handleStoreError(w, err)
		return
	}
	if !s.canAccessProject(r.Context(), user.UserID, project, nil) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req taskRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	fields, parsedDueDate := validateTaskRequest(req, true)
	assigneeID, assigneeFieldError, err := s.validateAssigneeID(r.Context(), req.AssigneeID)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}
	if assigneeFieldError != "" {
		fields["assignee_id"] = assigneeFieldError
	}
	if len(fields) > 0 {
		writeValidationError(w, fields)
		return
	}

	task, err := s.store.CreateTask(r.Context(), store.CreateTaskInput{
		ProjectID:   projectID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Status:      defaultString(req.Status, "todo"),
		Priority:    defaultString(req.Priority, "medium"),
		AssigneeID:  assigneeID,
		DueDate:     parsedDueDate,
		CreatorID:   user.UserID,
	})
	if err != nil {
		s.writeInternalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"task": task})
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var req taskRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	fields, parsedDueDate := validateTaskRequest(req, false)
	assigneeID, assigneeFieldError, err := s.validateAssigneeID(r.Context(), req.AssigneeID)
	if err != nil {
		s.writeInternalError(w, err)
		return
	}
	if assigneeFieldError != "" {
		fields["assignee_id"] = assigneeFieldError
	}
	if len(fields) > 0 {
		writeValidationError(w, fields)
		return
	}

	input := store.UpdateTaskInput{
		ID:          r.PathValue("id"),
		Title:       trimStringPointer(stringPointerOrNil(req.Title)),
		Description: trimStringPointer(stringPointerOrNil(req.Description)),
		Status:      trimStringPointer(stringPointerOrNil(req.Status)),
		Priority:    trimStringPointer(stringPointerOrNil(req.Priority)),
		AssigneeID:  assigneeID,
		DueDate:     parsedDueDate,
	}

	task, err := s.store.UpdateTask(r.Context(), input)
	if err != nil {
		s.handleStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"task": task})
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		s.writeUnauthorized(w, "missing authenticated user")
		return
	}

	err := s.store.DeleteTask(r.Context(), r.PathValue("id"), user.UserID)
	if err != nil {
		s.handleStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) canAccessProject(ctx context.Context, userID string, project store.Project, tasks []store.Task) bool {
	if project.OwnerID == userID {
		return true
	}
	if tasks == nil {
		projectTasks, err := s.store.ListProjectTasks(ctx, project.ID, store.TaskFilters{})
		if err != nil {
			return false
		}
		tasks = projectTasks
	}
	for _, task := range tasks {
		if task.AssigneeID != nil && *task.AssigneeID == userID {
			return true
		}
	}
	return false
}

func (s *Server) validateAssigneeID(ctx context.Context, assigneeID *string) (*string, string, error) {
	trimmed := trimStringPointer(assigneeID)
	if trimmed == nil {
		return nil, "", nil
	}

	_, err := s.store.GetUserByID(ctx, *trimmed)
	switch {
	case err == nil:
		return trimmed, "", nil
	case errors.Is(err, store.ErrBadRequest):
		return nil, "must be a valid user id", nil
	case errors.Is(err, store.ErrNotFound):
		return nil, "must reference an existing user", nil
	default:
		return nil, "", err
	}
}

func (s *Server) enforceJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodDelete && r.Method != http.MethodHead {
			if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
				writeError(w, http.StatusBadRequest, "content type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		s.logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func (s *Server) writeUnauthorized(w http.ResponseWriter, _ string) {
	writeError(w, http.StatusUnauthorized, "unauthorized")
}

func (s *Server) writeInternalError(w http.ResponseWriter, err error) {
	s.logger.Error("request failed", "error", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func (s *Server) handleStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, store.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	default:
		s.writeInternalError(w, err)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dest any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeValidationError(w http.ResponseWriter, fields map[string]string) {
	writeJSON(w, http.StatusBadRequest, map[string]any{
		"error":  "validation failed",
		"fields": fields,
	})
}

func validateTaskRequest(req taskRequest, requireTitle bool) (map[string]string, *time.Time) {
	fields := map[string]string{}
	if requireTitle && strings.TrimSpace(req.Title) == "" {
		fields["title"] = "is required"
	}

	status := strings.TrimSpace(req.Status)
	if status != "" && !contains([]string{"todo", "in_progress", "done"}, status) {
		fields["status"] = "must be one of todo, in_progress, done"
	}

	priority := strings.TrimSpace(req.Priority)
	if priority != "" && !contains([]string{"low", "medium", "high"}, priority) {
		fields["priority"] = "must be one of low, medium, high"
	}

	var parsedDueDate *time.Time
	if req.DueDate != nil && strings.TrimSpace(*req.DueDate) != "" {
		value, err := time.Parse("2006-01-02", strings.TrimSpace(*req.DueDate))
		if err != nil {
			fields["due_date"] = "must be in YYYY-MM-DD format"
		} else {
			parsedDueDate = &value
		}
	}

	return fields, parsedDueDate
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func defaultString(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func trimStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func stringPointerOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
