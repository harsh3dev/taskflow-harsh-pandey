import { useEffect, useState } from "react";
import { useApi, useAuth } from "../app/auth";
import { ApiError } from "../lib/api";
import { Alert, AlertDescription } from "../components/ui/alert";
import { Button } from "../components/ui/button";
import { Card } from "../components/ui/card";
import { WorkspaceSidebar } from "../components/projects/WorkspaceSidebar";
import { CreateProjectModal } from "../components/projects/CreateProjectModal";
import { ProjectListItem } from "../components/projects/ProjectListItem";
import { ProjectsEmptyState } from "../components/projects/ProjectsEmptyState";
import { createProject, listProjects } from "../lib/services/projects";
import { getErrorMessage } from "../lib/utils";
import { Project } from "../lib/types";

export function ProjectsPage() {
  const api = useApi();
  const { user } = useAuth();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState("");
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createErrors, setCreateErrors] = useState<Record<string, string>>({});
  const [creating, setCreating] = useState(false);

  async function loadProjects() {
    setLoading(true);
    setErrorMessage("");
    try {
      const response = await listProjects(api);
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

  async function handleCreateProject(name: string, description: string) {
    const errors: Record<string, string> = {};
    if (!name.trim()) {
      errors.name = "Project name is required";
    }
    setCreateErrors(errors);
    if (Object.keys(errors).length > 0) return;

    setCreating(true);
    try {
      const response = await createProject(api, {
        name: name.trim(),
        description: description.trim()
      });
      setProjects((current) => [response.project, ...current]);
      setShowCreateModal(false);
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

  const ownedCount = projects.filter((p) => p.owner_id === user?.id).length;

  return (
    <div className="flex min-h-[calc(100vh-60px)]">
      {/* Sidebar */}
      <WorkspaceSidebar projectCount={projects.length} ownedCount={ownedCount} />

      {/* Main content */}
      <div className="flex min-w-0 flex-1 flex-col">
        {/* Filter / action bar */}
        <div className="sticky top-[60px] z-4 flex items-center gap-3 border-b border-border bg-background px-4 py-2">
          <span className="text-sm font-medium text-muted-foreground">
            Projects
          </span>
          <span className="rounded-full bg-muted px-2 py-0.5 text-xs font-medium tabular-nums">
            {projects.length}
          </span>
          <div className="ml-auto flex items-center gap-2">
            <Button
              size="sm"
              onClick={() => setShowCreateModal(true)}
              type="button"
            >
              + New project
            </Button>
            <Button
              size="icon-sm"
              variant="ghost"
              onClick={() => void loadProjects()}
              type="button"
              title="Refresh"
              aria-label="Refresh projects"
            >
              ↻
            </Button>
          </div>
        </div>

        {/* Error */}
        {errorMessage ? (
          <div className="px-4 pt-4">
            <Alert variant="destructive">
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
          </div>
        ) : null}

        {/* Loading */}
        {loading ? (
          <div className="flex flex-1 items-center justify-center py-16 text-sm text-muted-foreground">
            Loading projects…
          </div>
        ) : null}

        {/* Empty state */}
        {!loading && projects.length === 0 ? (
          <div className="flex flex-1 items-center justify-center px-4 py-16">
            <ProjectsEmptyState />
          </div>
        ) : null}

        {/* Project list */}
        {!loading && projects.length > 0 ? (
          <Card className="m-4 rounded-xl p-0 overflow-hidden">
            {projects.map((project) => (
              <ProjectListItem
                currentUserId={user?.id}
                key={project.id}
                project={project}
              />
            ))}
          </Card>
        ) : null}
      </div>

      {/* Create project modal */}
      {showCreateModal ? (
        <CreateProjectModal
          creating={creating}
          errors={createErrors}
          onClose={() => {
            setShowCreateModal(false);
            setCreateErrors({});
          }}
          onSubmit={handleCreateProject}
        />
      ) : null}
    </div>
  );
}
