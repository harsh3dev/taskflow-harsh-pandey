import { FormEvent, useEffect, useState } from "react";
import { useApi, useAuth } from "../app/auth";
import { ApiError } from "../lib/api";
import { Alert, AlertDescription } from "../components/ui/alert";
import { CreateProjectCard } from "../components/projects/CreateProjectCard";
import { ProjectsListSection } from "../components/projects/ProjectsListSection";
import { ProjectsPageHeader } from "../components/projects/ProjectsPageHeader";
import { createProject, listProjects } from "../lib/services/projects";
import { getErrorMessage } from "../lib/utils";
import { Project } from "../lib/types";

export function ProjectsPage() {
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
      const response = await createProject(api, {
        name: createState.name.trim(),
        description: createState.description.trim()
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
    <main className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-6 sm:px-5">
      <ProjectsPageHeader ownedCount={ownedCount} projectCount={projects.length} />

      <CreateProjectCard
        creating={creating}
        description={createState.description}
        errors={createErrors}
        name={createState.name}
        onChange={(field, value) =>
          setCreateState((current) => ({
            ...current,
            [field]: value
          }))
        }
        onSubmit={handleCreateProject}
      />

      {errorMessage ? (
        <Alert variant="destructive">
          <AlertDescription>{errorMessage}</AlertDescription>
        </Alert>
      ) : null}

      <ProjectsListSection
        currentUserId={user?.id}
        loading={loading}
        projects={projects}
        onRefresh={() => void loadProjects()}
      />
    </main>
  );
}
