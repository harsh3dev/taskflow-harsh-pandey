package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/harshpn/taskflow/internal/auth"
	"github.com/harshpn/taskflow/internal/store"
)

type Dependencies struct {
	Logger              *slog.Logger
	Store               *store.Store
	TokenManager        auth.TokenManager
	BcryptCost          int
	MaxRequestBodyBytes int64
}

type Server struct {
	logger       *slog.Logger
	store        *store.Store
	tokenManager auth.TokenManager
	bcryptCost   int
	maxBodyBytes int64
}

func NewServer(deps Dependencies) *Server {
	return &Server{
		logger:       deps.Logger,
		store:        deps.Store,
		tokenManager: deps.TokenManager,
		bcryptCost:   deps.BcryptCost,
		maxBodyBytes: deps.MaxRequestBodyBytes,
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

	return s.withRequestID(s.withLogging(s.withRecovery(s.enforceJSON(mux))))
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
