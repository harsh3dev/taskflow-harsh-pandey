import { Button } from "../ui/button";
import { Project } from "../../lib/types";
import { ProjectCard } from "./ProjectCard";
import { ProjectsEmptyState } from "./ProjectsEmptyState";

type ProjectsListSectionProps = {
  loading: boolean;
  projects: Project[];
  currentUserId?: string;
  onRefresh: () => void;
};

export function ProjectsListSection({
  loading,
  projects,
  currentUserId,
  onRefresh
}: ProjectsListSectionProps) {
  return (
    <section className="flex flex-col gap-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Accessible projects</h2>
          <p className="text-sm text-muted-foreground">
            Your owned projects and work where you are assigned.
          </p>
        </div>
        <Button variant="outline" onClick={onRefresh} type="button">
          Refresh
        </Button>
      </div>

      {loading ? <div className="py-8 text-muted-foreground">Loading projects...</div> : null}

      {!loading && projects.length === 0 ? <ProjectsEmptyState /> : null}

      {!loading && projects.length > 0 ? (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {projects.map((project) => (
            <ProjectCard currentUserId={currentUserId} key={project.id} project={project} />
          ))}
        </div>
      ) : null}
    </section>
  );
}
