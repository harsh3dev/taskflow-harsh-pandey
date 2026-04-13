package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/harshpn/taskflow/internal/auth"
)

type Dependencies struct {
	Logger              *slog.Logger
	TokenParser         auth.TokenParser
	AuthService         authService
	ProjectService      projectService
	TaskService         taskService
	UserService         userService
	MaxRequestBodyBytes int64
}

type Server struct {
	logger       *slog.Logger
	tokenParser  auth.TokenParser
	authService  authService
	projects     projectService
	tasks        taskService
	users        userService
	maxBodyBytes int64
}

func NewServer(deps Dependencies) *Server {
	return &Server{
		logger:       deps.Logger,
		tokenParser:  deps.TokenParser,
		authService:  deps.AuthService,
		projects:     deps.ProjectService,
		tasks:        deps.TaskService,
		users:        deps.UserService,
		maxBodyBytes: deps.MaxRequestBodyBytes,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("POST /auth/register", s.handleRegister)
	mux.HandleFunc("POST /auth/login", s.handleLogin)
	mux.HandleFunc("POST /auth/refresh", s.handleRefresh)
	mux.HandleFunc("POST /auth/logout", s.handleLogout)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /projects", s.handleListProjects)
	protected.HandleFunc("POST /projects", s.handleCreateProject)
	protected.HandleFunc("GET /projects/{id}", s.handleGetProject)
	protected.HandleFunc("PATCH /projects/{id}", s.handleUpdateProject)
	protected.HandleFunc("DELETE /projects/{id}", s.handleDeleteProject)
	protected.HandleFunc("GET /projects/{id}/tasks", s.handleListProjectTasks)
	protected.HandleFunc("POST /projects/{id}/tasks", s.handleCreateTask)
	protected.HandleFunc("GET /projects/{id}/stats", s.handleGetProjectStats)
	protected.HandleFunc("GET /users", s.handleListUsers)
	protected.HandleFunc("PATCH /tasks/{id}", s.handleUpdateTask)
	protected.HandleFunc("DELETE /tasks/{id}", s.handleDeleteTask)

	mux.Handle("/projects", auth.Middleware(s.tokenParser, s.writeUnauthorized)(protected))
	mux.Handle("/projects/", auth.Middleware(s.tokenParser, s.writeUnauthorized)(protected))
	mux.Handle("/users", auth.Middleware(s.tokenParser, s.writeUnauthorized)(protected))
	mux.Handle("/users/", auth.Middleware(s.tokenParser, s.writeUnauthorized)(protected))
	mux.Handle("/tasks/", auth.Middleware(s.tokenParser, s.writeUnauthorized)(protected))

	return s.withRequestID(s.withLogging(s.withRecovery(s.enforceJSON(mux))))
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
