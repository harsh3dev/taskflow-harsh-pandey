import {
  Component,
  ErrorInfo,
  FormEvent,
  ReactNode,
  createContext,
  useContext,
  useEffect,
  useState
} from "react";
import {
  BrowserRouter,
  Link,
  Navigate,
  Outlet,
  Route,
  Routes,
  useLocation,
  useNavigate,
  useParams
} from "react-router-dom";

type User = {
  id: string;
  name: string;
  email: string;
  created_at: string;
};

type Project = {
  id: string;
  name: string;
  description: string;
  owner_id: string;
  created_at: string;
};

type Task = {
  id: string;
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  project_id: string;
  assignee_id: string | null;
  creator_id: string;
  due_date: string | null;
  created_at: string;
  updated_at: string;
};

type TaskStatus = "todo" | "in_progress" | "done";
type TaskPriority = "low" | "medium" | "high";

type AuthResponse = {
  token: string;
  user: User;
};

type ApiErrorShape = {
  error?: string;
  fields?: Record<string, string>;
};

class ApiError extends Error {
  status: number;
  fields?: Record<string, string>;

  constructor(status: number, message: string, fields?: Record<string, string>) {
    super(message);
    this.status = status;
    this.fields = fields;
  }
}

type AuthContextValue = {
  token: string | null;
  user: User | null;
  login: (payload: { email: string; password: string }) => Promise<void>;
  register: (payload: {
    name: string;
    email: string;
    password: string;
  }) => Promise<void>;
  logout: () => void;
};

const AUTH_STORAGE_KEY = "taskflow.auth";
const API_BASE_URL = (
  import.meta.env.VITE_API_BASE_URL?.trim() || "/api"
).replace(/\/$/, "");

const statusOptions: Array<{ value: TaskStatus; label: string }> = [
  { value: "todo", label: "To do" },
  { value: "in_progress", label: "In progress" },
  { value: "done", label: "Done" }
];

const priorityOptions: Array<{ value: TaskPriority; label: string }> = [
  { value: "low", label: "Low" },
  { value: "medium", label: "Medium" },
  { value: "high", label: "High" }
];

const AuthContext = createContext<AuthContextValue | null>(null);

class AppErrorBoundary extends Component<
  { children: ReactNode },
  { hasError: boolean }
> {
  constructor(props: { children: ReactNode }) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  componentDidCatch(_error: Error, _errorInfo: ErrorInfo) {}

  render() {
    if (this.state.hasError) {
      return (
        <main className="page">
          <section className="panel empty-state app-fallback">
            <p className="eyebrow">Application error</p>
            <h1>TaskFlow hit an unexpected problem.</h1>
            <p>Refresh the page and try again. Your saved session will be kept in local storage.</p>
            <button
              className="button button-primary"
              onClick={() => window.location.reload()}
              type="button"
            >
              Reload app
            </button>
          </section>
        </main>
      );
    }

    return this.props.children;
  }
}

function readStoredAuth() {
  const fallback = { token: null as string | null, user: null as User | null };
  const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
  if (!raw) {
    return fallback;
  }

  try {
    const parsed = JSON.parse(raw) as { token?: string; user?: User };
    return {
      token: parsed.token ?? null,
      user: parsed.user ?? null
    };
  } catch {
    window.localStorage.removeItem(AUTH_STORAGE_KEY);
    return fallback;
  }
}

async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
  token?: string | null
): Promise<T> {
  const headers = new Headers(options.headers ?? {});
  if (options.body !== undefined && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  headers.set("Accept", "application/json");
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers
  });

  const text = await response.text();
  const trimmed = text.trim();
  const payload = trimmed ? (JSON.parse(trimmed) as ApiErrorShape & T) : ({} as T);

  if (!response.ok) {
    const apiError = payload as ApiErrorShape;
    throw new ApiError(
      response.status,
      apiError.error || "Request failed",
      apiError.fields
    );
  }

  return payload as T;
}

function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => readStoredAuth().token);
  const [user, setUser] = useState<User | null>(() => readStoredAuth().user);

  useEffect(() => {
    if (token && user) {
      window.localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify({ token, user }));
      return;
    }
    window.localStorage.removeItem(AUTH_STORAGE_KEY);
  }, [token, user]);

  async function handleAuth(path: "/auth/login" | "/auth/register", body: Record<string, string>) {
    const response = await apiRequest<AuthResponse>(path, {
      method: "POST",
      body: JSON.stringify(body)
    });
    setToken(response.token);
    setUser(response.user);
  }

  const value: AuthContextValue = {
    token,
    user,
    login: (payload) => handleAuth("/auth/login", payload),
    register: (payload) => handleAuth("/auth/register", payload),
    logout: () => {
      setToken(null);
      setUser(null);
    }
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("Auth context is unavailable");
  }
  return context;
}

function useApi() {
  const { token, logout } = useAuth();

  return async function request<T>(path: string, options: RequestInit = {}) {
    try {
      return await apiRequest<T>(path, options, token);
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        logout();
      }
      throw error;
    }
  };
}

function App() {
  return (
    <AppErrorBoundary>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<AuthPage mode="login" />} />
            <Route path="/register" element={<AuthPage mode="register" />} />
            <Route element={<ProtectedLayout />}>
              <Route index element={<ProjectsPage />} />
              <Route path="/projects/:projectId" element={<ProjectDetailPage />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </AppErrorBoundary>
  );
}

function ProtectedLayout() {
  const { token, user, logout } = useAuth();
  const location = useLocation();

  if (!token || !user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return (
    <div className="app-shell">
      <nav className="shell-nav">
        <div className="brand">
          <div className="brand-mark">TF</div>
          <div className="brand-copy">
            <strong>TaskFlow</strong>
            <p>Projects, tasks, and ownership in one clean workspace.</p>
          </div>
        </div>
        <div className="shell-actions">
          <div className="nav-user">{user.name}</div>
          <button className="button button-ghost" onClick={logout} type="button">
            Logout
          </button>
        </div>
      </nav>
      <Outlet />
    </div>
  );
}

function AuthPage({ mode }: { mode: "login" | "register" }) {
  const auth = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const nextPath = (location.state as { from?: string } | null)?.from || "/";
  const [form, setForm] = useState({
    name: "",
    email: "",
    password: ""
  });
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [errorMessage, setErrorMessage] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (auth.token && auth.user) {
      navigate(nextPath, { replace: true });
    }
  }, [auth.token, auth.user, navigate, nextPath]);

  function validate() {
    const nextErrors: Record<string, string> = {};
    if (mode === "register" && form.name.trim().length < 2) {
      nextErrors.name = "Enter at least 2 characters";
    }
    if (!form.email.trim()) {
      nextErrors.email = "Email is required";
    } else if (!/\S+@\S+\.\S+/.test(form.email.trim())) {
      nextErrors.email = "Enter a valid email";
    }
    if (form.password.trim().length < 8) {
      nextErrors.password = "Use at least 8 characters";
    }
    return nextErrors;
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const nextErrors = validate();
    setFieldErrors(nextErrors);
    setErrorMessage("");

    if (Object.keys(nextErrors).length > 0) {
      return;
    }

    setSubmitting(true);
    try {
      if (mode === "login") {
        await auth.login({
          email: form.email.trim(),
          password: form.password
        });
      } else {
        await auth.register({
          name: form.name.trim(),
          email: form.email.trim(),
          password: form.password
        });
      }
      navigate(nextPath, { replace: true });
    } catch (error) {
      if (error instanceof ApiError) {
        setFieldErrors(error.fields ?? {});
        setErrorMessage(error.message);
      } else {
        setErrorMessage("Unable to reach the API.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  const title = mode === "login" ? "Welcome back." : "Build your workspace.";
  const subtitle =
    mode === "login"
      ? "Sign in to manage projects, track tasks, and pick up where you left off."
      : "Create an account, save your session, and move straight into project planning.";

  return (
    <main className="auth-page">
      <section className="auth-hero">
        <div className="auth-hero-panel stack">
          <p className="eyebrow">Phase 3 Frontend</p>
          <h1>{title}</h1>
          <p>{subtitle}</p>
          <div className="stats-grid">
            <div className="stat-card">
              <span>Protected routes</span>
              <strong>JWT</strong>
            </div>
            <div className="stat-card">
              <span>Project views</span>
              <strong>Kanban</strong>
            </div>
            <div className="stat-card">
              <span>Task updates</span>
              <strong>Optimistic</strong>
            </div>
            <div className="stat-card">
              <span>Session</span>
              <strong>Persistent</strong>
            </div>
          </div>
        </div>
      </section>

      <section className="auth-card">
        <form className="stack" onSubmit={handleSubmit}>
          <div>
            <p className="eyebrow">{mode === "login" ? "Login" : "Register"}</p>
            <h2>{mode === "login" ? "Sign in to TaskFlow" : "Create your account"}</h2>
            <p className="helper-text">
              {mode === "login"
                ? "Use the seeded credentials or your own account."
                : "Your account will be logged in immediately after registration."}
            </p>
          </div>

          {errorMessage ? <div className="alert alert-error">{errorMessage}</div> : null}

          {mode === "register" ? (
            <div className="field">
              <label htmlFor="name">Name</label>
              <input
                id="name"
                name="name"
                value={form.name}
                onChange={(event) =>
                  setForm((current) => ({ ...current, name: event.target.value }))
                }
                placeholder="Avery Chen"
              />
              {fieldErrors.name ? <div className="field-error">{fieldErrors.name}</div> : null}
            </div>
          ) : null}

          <div className="field">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              name="email"
              type="email"
              value={form.email}
              onChange={(event) =>
                setForm((current) => ({ ...current, email: event.target.value }))
              }
              placeholder="test@example.com"
              autoComplete="email"
            />
            {fieldErrors.email ? <div className="field-error">{fieldErrors.email}</div> : null}
          </div>

          <div className="field">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              name="password"
              type="password"
              value={form.password}
              onChange={(event) =>
                setForm((current) => ({ ...current, password: event.target.value }))
              }
              placeholder="password123"
              autoComplete={mode === "login" ? "current-password" : "new-password"}
            />
            {fieldErrors.password ? (
              <div className="field-error">{fieldErrors.password}</div>
            ) : (
              <div className="field-note">Minimum length: 8 characters.</div>
            )}
          </div>

          <button className="button button-primary" disabled={submitting} type="submit">
            {submitting
              ? mode === "login"
                ? "Signing in..."
                : "Creating account..."
              : mode === "login"
                ? "Sign in"
                : "Create account"}
          </button>
        </form>

        <p className="helper-text">
          {mode === "login" ? "Need an account? " : "Already registered? "}
          <Link className="link-text" to={mode === "login" ? "/register" : "/login"}>
            {mode === "login" ? "Create one" : "Sign in"}
          </Link>
        </p>
      </section>
    </main>
  );
}

function ProjectsPage() {
  const api = useApi();
  const { user } = useAuth();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState("");
  const [createState, setCreateState] = useState({
    name: "",
    description: ""
  });
  const [createErrors, setCreateErrors] = useState<Record<string, string>>({});
  const [creating, setCreating] = useState(false);

  async function loadProjects() {
    setLoading(true);
    setErrorMessage("");
    try {
      const response = await api<{ projects: Project[] }>("/projects");
      setProjects(response.projects);
    } catch (error) {
      setErrorMessage(getErrorMessage(error, "Unable to load projects."));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadProjects();
  }, []);

  async function handleCreateProject(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const errors: Record<string, string> = {};
    if (!createState.name.trim()) {
      errors.name = "Project name is required";
    }
    setCreateErrors(errors);
    if (Object.keys(errors).length > 0) {
      return;
    }

    setCreating(true);
    try {
      const response = await api<{ project: Project }>("/projects", {
        method: "POST",
        body: JSON.stringify({
          name: createState.name.trim(),
          description: createState.description.trim()
        })
      });

      setProjects((current) => [response.project, ...current]);
      setCreateState({ name: "", description: "" });
      setCreateErrors({});
    } catch (error) {
      if (error instanceof ApiError && error.fields) {
        setCreateErrors(error.fields);
      }
      setErrorMessage(getErrorMessage(error, "Unable to create project."));
    } finally {
      setCreating(false);
    }
  }

  const ownedCount = projects.filter((project) => project.owner_id === user?.id).length;

  return (
    <main className="page stack">
      <section className="page-header">
        <p className="eyebrow">Workspace</p>
        <div className="split">
          <div className="stack">
            <h1>Projects that matter today.</h1>
            <p>
              Create a project, keep a clean overview of your accessible workspaces,
              and jump straight into grouped task management.
            </p>
          </div>
          <div className="stats-grid">
            <div className="stat-card">
              <span>Total projects</span>
              <strong>{projects.length}</strong>
            </div>
            <div className="stat-card">
              <span>Owned by you</span>
              <strong>{ownedCount}</strong>
            </div>
          </div>
        </div>
      </section>

      <section className="panel">
        <div className="split">
          <div>
            <h2>Create a project</h2>
            <p className="helper-text">Keep the scope tight and the description useful.</p>
          </div>
        </div>

        <form className="stack" onSubmit={handleCreateProject}>
          <div className="field">
            <label htmlFor="project-name">Project name</label>
            <input
              id="project-name"
              value={createState.name}
              onChange={(event) =>
                setCreateState((current) => ({ ...current, name: event.target.value }))
              }
              placeholder="Q2 Launch Readiness"
            />
            {createErrors.name ? <div className="field-error">{createErrors.name}</div> : null}
          </div>

          <div className="field">
            <label htmlFor="project-description">Description</label>
            <textarea
              id="project-description"
              value={createState.description}
              onChange={(event) =>
                setCreateState((current) => ({
                  ...current,
                  description: event.target.value
                }))
              }
              placeholder="Goals, handoff notes, and delivery boundaries."
            />
          </div>

          <div className="row">
            <button className="button button-primary" disabled={creating} type="submit">
              {creating ? "Creating..." : "Create project"}
            </button>
            <span className="helper-text">Projects are visible if you own them or have tasks in them.</span>
          </div>
        </form>
      </section>

      {errorMessage ? <div className="alert alert-error">{errorMessage}</div> : null}

      <section className="stack">
        <div className="split">
          <div>
            <h2>Accessible projects</h2>
            <p className="helper-text">Your owned projects and work where you are assigned.</p>
          </div>
          <button className="button button-subtle" onClick={() => void loadProjects()} type="button">
            Refresh
          </button>
        </div>

        {loading ? <div className="loading">Loading projects...</div> : null}

        {!loading && projects.length === 0 ? (
          <div className="panel empty-state">
            <h3>No projects yet</h3>
            <p>Create your first project to start organizing tasks.</p>
          </div>
        ) : null}

        {!loading && projects.length > 0 ? (
          <div className="projects-grid">
            {projects.map((project) => {
              const isOwner = project.owner_id === user?.id;
              return (
                <article className="project-card" key={project.id}>
                  <div className="stack">
                    <div className="split">
                      <span className="role-chip">{isOwner ? "Owner" : "Contributor"}</span>
                      <span className="meta-pill">{formatDateTime(project.created_at)}</span>
                    </div>
                    <div className="stack">
                      <h2>{project.name}</h2>
                      <p className="project-summary">
                        {project.description || "No description yet. Add one when the project scope firms up."}
                      </p>
                    </div>
                  </div>
                  <Link className="button button-secondary" to={`/projects/${project.id}`}>
                    Open project
                  </Link>
                </article>
              );
            })}
          </div>
        ) : null}
      </section>
    </main>
  );
}

function ProjectDetailPage() {
  const { projectId = "" } = useParams();
  const api = useApi();
  const { user } = useAuth();
  const navigate = useNavigate();
  const [users, setUsers] = useState<User[]>([]);
  const [project, setProject] = useState<Project | null>(null);
  const [allTasks, setAllTasks] = useState<Task[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loadingProject, setLoadingProject] = useState(true);
  const [loadingTasks, setLoadingTasks] = useState(false);
  const [projectError, setProjectError] = useState("");
  const [taskError, setTaskError] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("");
  const [assigneeFilter, setAssigneeFilter] = useState<string>("");
  const [modalState, setModalState] = useState<
    { mode: "create"; task: null } | { mode: "edit"; task: Task }
  >();
  const [deletingTaskId, setDeletingTaskId] = useState<string | null>(null);
  const [statusSavingId, setStatusSavingId] = useState<string | null>(null);

  async function loadProjectShell() {
    setLoadingProject(true);
    setProjectError("");
    try {
      const [response, usersResponse] = await Promise.all([
        api<{ project: Project; tasks: Task[] }>(`/projects/${projectId}`),
        api<{ users: User[] }>("/users")
      ]);
      setProject(response.project);
      setAllTasks(response.tasks);
      setTasks(response.tasks);
      setUsers(usersResponse.users);
    } catch (error) {
      setProjectError(getErrorMessage(error, "Unable to load this project."));
    } finally {
      setLoadingProject(false);
    }
  }

  async function loadTasks(filters?: { status?: string; assignee?: string }) {
    setLoadingTasks(true);
    setTaskError("");
    const params = new URLSearchParams();
    if (filters?.status) {
      params.set("status", filters.status);
    }
    if (filters?.assignee) {
      params.set("assignee", filters.assignee);
    }
    const query = params.toString();

    try {
      const response = await api<{ tasks: Task[] }>(
        `/projects/${projectId}/tasks${query ? `?${query}` : ""}`
      );
      setTasks(response.tasks);
    } catch (error) {
      setTaskError(getErrorMessage(error, "Unable to refresh tasks."));
    } finally {
      setLoadingTasks(false);
    }
  }

  useEffect(() => {
    void loadProjectShell();
  }, [projectId]);

  useEffect(() => {
    if (!project) {
      return;
    }
    void loadTasks({
      status: statusFilter,
      assignee: assigneeFilter
    });
  }, [project, statusFilter, assigneeFilter]);

  async function refreshProjectAndTasks() {
    const [response, usersResponse] = await Promise.all([
      api<{ project: Project; tasks: Task[] }>(`/projects/${projectId}`),
      api<{ users: User[] }>("/users")
    ]);
    setProject(response.project);
    setAllTasks(response.tasks);
    setUsers(usersResponse.users);
    if (statusFilter || assigneeFilter) {
      await loadTasks({ status: statusFilter, assignee: assigneeFilter });
      return;
    }
    setTasks(response.tasks);
  }

  async function handleDeleteTask(taskId: string) {
    setDeletingTaskId(taskId);
    setTaskError("");
    try {
      await api(`/tasks/${taskId}`, { method: "DELETE" });
      setTasks((current) => current.filter((task) => task.id !== taskId));
      setAllTasks((current) => current.filter((task) => task.id !== taskId));
    } catch (error) {
      setTaskError(getErrorMessage(error, "Unable to delete the task."));
    } finally {
      setDeletingTaskId(null);
    }
  }

  async function handleStatusChange(task: Task, status: TaskStatus) {
    const previousTasks = tasks;
    const previousAllTasks = allTasks;
    const applyStatus = (collection: Task[]) =>
      collection.map((item) => (item.id === task.id ? { ...item, status } : item));

    setStatusSavingId(task.id);
    setTaskError("");
    setTasks((current) => applyStatus(current));
    setAllTasks((current) => applyStatus(current));

    try {
      const response = await api<{ task: Task }>(`/tasks/${task.id}`, {
        method: "PATCH",
        body: JSON.stringify({ status })
      });

      setTasks((current) =>
        current.map((item) => (item.id === task.id ? response.task : item))
      );
      setAllTasks((current) =>
        current.map((item) => (item.id === task.id ? response.task : item))
      );
    } catch (error) {
      setTasks(previousTasks);
      setAllTasks(previousAllTasks);
      setTaskError(getErrorMessage(error, "Unable to update task status."));
    } finally {
      setStatusSavingId(null);
    }
  }

  async function handleTaskSaved() {
    setModalState(undefined);
    try {
      await refreshProjectAndTasks();
    } catch (error) {
      setTaskError(getErrorMessage(error, "Task saved, but the refreshed view failed."));
    }
  }

  const visibleColumns = statusOptions.map((column) => ({
    ...column,
    tasks: tasks.filter((task) => task.status === column.value)
  }));

  const assigneeOptions = Array.from(
    new Set(
      allTasks
        .map((task) => task.assignee_id)
        .filter((value): value is string => Boolean(value))
    )
  );
  const userMap = new Map(users.map((entry) => [entry.id, entry]));

  if (loadingProject) {
    return <main className="page loading">Loading project...</main>;
  }

  if (projectError || !project) {
    return (
      <main className="page stack">
        <div className="alert alert-error">{projectError || "Project not found."}</div>
        <button className="button button-secondary" onClick={() => navigate("/")} type="button">
          Back to projects
        </button>
      </main>
    );
  }

  const canEditProject = project.owner_id === user?.id;

  return (
    <main className="page stack">
      <section className="page-header">
        <div className="split">
          <div className="stack">
            <Link className="link-text" to="/">
              ← Back to projects
            </Link>
            <p className="eyebrow">Project detail</p>
            <h1>{project.name}</h1>
            <p>{project.description || "No project description yet."}</p>
          </div>
          <div className="stats-grid">
            <div className="stat-card">
              <span>Visible tasks</span>
              <strong>{tasks.length}</strong>
            </div>
            <div className="stat-card">
              <span>All tasks</span>
              <strong>{allTasks.length}</strong>
            </div>
            <div className="stat-card">
              <span>Role</span>
              <strong>{canEditProject ? "Owner" : "Member"}</strong>
            </div>
          </div>
        </div>
      </section>

      <section className="panel stack">
        <div className="split">
          <div>
            <h2>Filters and actions</h2>
            <p className="helper-text">
              Filter by task status or assignee. Status changes apply immediately and roll back on API failure.
            </p>
          </div>
          <div className="inline-actions">
            <button
              className="button button-primary"
              onClick={() => setModalState({ mode: "create", task: null })}
              type="button"
            >
              New task
            </button>
            <button
              className="button button-subtle"
              onClick={() => void refreshProjectAndTasks()}
              type="button"
            >
              Refresh
            </button>
          </div>
        </div>

        <div className="filter-bar">
          <div className="field">
            <label htmlFor="status-filter">Status</label>
            <select
              id="status-filter"
              value={statusFilter}
              onChange={(event) => setStatusFilter(event.target.value)}
            >
              <option value="">All statuses</option>
              {statusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>

          <div className="field">
            <label htmlFor="assignee-filter">Assignee</label>
            <select
              id="assignee-filter"
              value={assigneeFilter}
              onChange={(event) => setAssigneeFilter(event.target.value)}
            >
              <option value="">All assignees</option>
              {assigneeOptions.map((assigneeId) => (
                <option key={assigneeId} value={assigneeId}>
                  {userMap.get(assigneeId)?.name || (assigneeId === user?.id ? "You" : abbreviateId(assigneeId))}
                </option>
              ))}
            </select>
          </div>
        </div>

        {taskError ? <div className="alert alert-error">{taskError}</div> : null}
      </section>

      {loadingTasks ? <div className="loading">Refreshing tasks...</div> : null}

      {!loadingTasks && tasks.length === 0 ? (
        <section className="panel empty-state">
          <h3>No matching tasks</h3>
          <p>Adjust the filters or add a new task to this project.</p>
        </section>
      ) : null}

      {!loadingTasks && tasks.length > 0 ? (
        <section className="task-board">
          {visibleColumns.map((column) => (
            <div className="task-column" key={column.value}>
              <div className="split">
                <h3>{column.label}</h3>
                <span className="meta-pill">{column.tasks.length}</span>
              </div>

              {column.tasks.length === 0 ? (
                <div className="task-card">
                  <p className="helper-text">No tasks in this column.</p>
                </div>
              ) : null}

              {column.tasks.map((task) => (
                <article className="task-card" key={task.id}>
                  <div className="stack">
                    <div className="split">
                      <span className="status-chip" data-status={task.status}>
                        {labelForStatus(task.status)}
                      </span>
                      <span className="priority-chip" data-priority={task.priority}>
                        {labelForPriority(task.priority)}
                      </span>
                    </div>
                    <div className="stack">
                      <h4>{task.title}</h4>
                      <p className="project-summary">
                        {task.description || "No task description yet."}
                      </p>
                    </div>
                    <div className="task-meta">
                      <span>
                        Assignee: {task.assignee_id
                          ? userMap.get(task.assignee_id)?.name || (task.assignee_id === user?.id ? "You" : abbreviateId(task.assignee_id))
                          : "Unassigned"}
                      </span>
                      <span>Due: {task.due_date ? formatDate(task.due_date) : "No date"}</span>
                    </div>
                  </div>

                  <div className="task-actions">
                    <div className="field">
                      <label htmlFor={`status-${task.id}`}>Move task</label>
                      <select
                        id={`status-${task.id}`}
                        disabled={statusSavingId === task.id}
                        value={task.status}
                        onChange={(event) =>
                          void handleStatusChange(task, event.target.value as TaskStatus)
                        }
                      >
                        {statusOptions.map((option) => (
                          <option key={option.value} value={option.value}>
                            {option.label}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>

                  <div className="row">
                    <button
                      className="button button-subtle"
                      onClick={() => setModalState({ mode: "edit", task })}
                      type="button"
                    >
                      Edit
                    </button>
                    <button
                      className="button button-danger"
                      disabled={deletingTaskId === task.id}
                      onClick={() => void handleDeleteTask(task.id)}
                      type="button"
                    >
                      {deletingTaskId === task.id ? "Deleting..." : "Delete"}
                    </button>
                  </div>
                </article>
              ))}
            </div>
          ))}
        </section>
      ) : null}

      {modalState ? (
        <TaskModal
          key={modalState.mode === "edit" ? modalState.task.id : "new-task"}
          mode={modalState.mode}
          onClose={() => setModalState(undefined)}
          onSaved={() => void handleTaskSaved()}
          projectId={projectId}
          task={modalState.task}
          users={users}
        />
      ) : null}
    </main>
  );
}

function TaskModal({
  mode,
  projectId,
  task,
  users,
  onClose,
  onSaved
}: {
  mode: "create" | "edit";
  projectId: string;
  task: Task | null;
  users: User[];
  onClose: () => void;
  onSaved: () => void;
}) {
  const api = useApi();
  const { user } = useAuth();
  const [form, setForm] = useState({
    title: task?.title ?? "",
    description: task?.description ?? "",
    status: task?.status ?? "todo",
    priority: task?.priority ?? "medium",
    assignee_id: task?.assignee_id ?? "",
    due_date: toDateInputValue(task?.due_date ?? "")
  });
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [errorMessage, setErrorMessage] = useState("");
  const [submitting, setSubmitting] = useState(false);

  function validate() {
    const nextErrors: Record<string, string> = {};
    if (!form.title.trim()) {
      nextErrors.title = "Title is required";
    }
    if (form.due_date && !/^\d{4}-\d{2}-\d{2}$/.test(form.due_date)) {
      nextErrors.due_date = "Use YYYY-MM-DD";
    }
    return nextErrors;
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const errors = validate();
    setFieldErrors(errors);
    setErrorMessage("");
    if (Object.keys(errors).length > 0) {
      return;
    }

    setSubmitting(true);
    const payload = {
      title: form.title.trim(),
      description: form.description.trim(),
      status: form.status,
      priority: form.priority,
      assignee_id: form.assignee_id.trim() || null,
      due_date: form.due_date || null
    };

    try {
      if (mode === "create") {
        await api(`/projects/${projectId}/tasks`, {
          method: "POST",
          body: JSON.stringify(payload)
        });
      } else if (task) {
        await api(`/tasks/${task.id}`, {
          method: "PATCH",
          body: JSON.stringify(payload)
        });
      }
      onSaved();
    } catch (error) {
      if (error instanceof ApiError && error.fields) {
        setFieldErrors(error.fields);
      }
      setErrorMessage(getErrorMessage(error, "Unable to save the task."));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="modal-backdrop" role="presentation">
      <div className="modal-card">
        <div className="split">
          <div>
            <p className="eyebrow">{mode === "create" ? "Create task" : "Edit task"}</p>
            <h2>{mode === "create" ? "Add a new task" : "Update task details"}</h2>
            <p className="helper-text">
              Pick an assignee from the backend user directory, or leave the task unassigned.
            </p>
          </div>
          <button className="button button-subtle" onClick={onClose} type="button">
            Close
          </button>
        </div>

        <form className="stack" onSubmit={handleSubmit}>
          {errorMessage ? <div className="alert alert-error">{errorMessage}</div> : null}

          <div className="field">
            <label htmlFor="task-title">Title</label>
            <input
              id="task-title"
              value={form.title}
              onChange={(event) =>
                setForm((current) => ({ ...current, title: event.target.value }))
              }
              placeholder="Draft API release notes"
            />
            {fieldErrors.title ? <div className="field-error">{fieldErrors.title}</div> : null}
          </div>

          <div className="field">
            <label htmlFor="task-description">Description</label>
            <textarea
              id="task-description"
              value={form.description}
              onChange={(event) =>
                setForm((current) => ({ ...current, description: event.target.value }))
              }
              placeholder="Context, expected output, and review notes."
            />
          </div>

          <div className="row">
            <div className="field" style={{ flex: "1 1 180px" }}>
              <label htmlFor="task-status">Status</label>
              <select
                id="task-status"
                value={form.status}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    status: event.target.value as TaskStatus
                  }))
                }
              >
                {statusOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="field" style={{ flex: "1 1 180px" }}>
              <label htmlFor="task-priority">Priority</label>
              <select
                id="task-priority"
                value={form.priority}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    priority: event.target.value as TaskPriority
                  }))
                }
              >
                {priorityOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="field">
            <label htmlFor="task-assignee">Assignee</label>
            <select
              id="task-assignee"
              value={form.assignee_id}
              onChange={(event) =>
                setForm((current) => ({ ...current, assignee_id: event.target.value }))
              }
            >
              <option value="">Unassigned</option>
              {users.map((candidate) => (
                <option key={candidate.id} value={candidate.id}>
                  {candidate.name}
                  {candidate.id === user?.id ? " (You)" : ""}
                  {" · "}
                  {candidate.email}
                </option>
              ))}
            </select>
            <div className="row">
              <button
                className="button button-subtle"
                onClick={() =>
                  setForm((current) => ({ ...current, assignee_id: user?.id ?? "" }))
                }
                type="button"
              >
                Assign to me
              </button>
              <button
                className="button button-subtle"
                onClick={() => setForm((current) => ({ ...current, assignee_id: "" }))}
                type="button"
              >
                Clear assignee
              </button>
            </div>
            {fieldErrors.assignee_id ? (
              <div className="field-error">{fieldErrors.assignee_id}</div>
            ) : (
              <div className="field-note">Choose from users returned by the authenticated backend directory.</div>
            )}
          </div>

          <div className="field">
            <label htmlFor="task-due-date">Due date</label>
            <input
              id="task-due-date"
              type="date"
              value={form.due_date}
              onChange={(event) =>
                setForm((current) => ({ ...current, due_date: event.target.value }))
              }
            />
            {fieldErrors.due_date ? (
              <div className="field-error">{fieldErrors.due_date}</div>
            ) : (
              <div className="field-note">Optional. Stored as a date-only value.</div>
            )}
          </div>

          <div className="row">
            <button className="button button-primary" disabled={submitting} type="submit">
              {submitting ? "Saving..." : mode === "create" ? "Create task" : "Save changes"}
            </button>
            <button className="button button-subtle" onClick={onClose} type="button">
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function getErrorMessage(error: unknown, fallback: string) {
  if (error instanceof ApiError) {
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}

function labelForStatus(status: TaskStatus) {
  return statusOptions.find((option) => option.value === status)?.label ?? status;
}

function labelForPriority(priority: TaskPriority) {
  return priorityOptions.find((option) => option.value === priority)?.label ?? priority;
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric"
  });
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric"
  });
}

function abbreviateId(value: string) {
  if (value.length <= 10) {
    return value;
  }
  return `${value.slice(0, 8)}…${value.slice(-4)}`;
}

function toDateInputValue(value: string) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return date.toISOString().slice(0, 10);
}

export default App;
