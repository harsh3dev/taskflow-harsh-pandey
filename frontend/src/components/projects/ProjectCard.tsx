import { Link } from "react-router-dom";
import { Badge } from "../ui/badge";
import { Card, CardContent } from "../ui/card";
import { Project } from "../../lib/types";
import { formatDateTime } from "../../lib/utils";

type ProjectCardProps = {
  project: Project;
  currentUserId?: string;
};

export function ProjectCard({ project, currentUserId }: ProjectCardProps) {
  const isOwner = project.owner_id === currentUserId;

  return (
    <Card className="group transition-shadow hover:shadow-md">
      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <Badge variant="secondary">{isOwner ? "Owner" : "Contributor"}</Badge>
          <Badge variant="outline">{formatDateTime(project.created_at)}</Badge>
        </div>
        <div className="flex flex-col gap-2">
          <h2 className="m-0 text-xl font-semibold tracking-tight">{project.name}</h2>
          <p className="line-clamp-2 text-sm text-muted-foreground">
            {project.description || "No description yet. Add one when the project scope firms up."}
          </p>
        </div>
        <Link
          className="inline-flex items-center gap-2 self-start rounded-full bg-primary px-[1.15rem] py-[0.85rem] font-bold text-primary-foreground transition-transform duration-200 hover:-translate-y-px"
          to={`/projects/${project.id}`}
        >
          Open project
          <span className="inline-block transition-transform duration-200 group-hover:translate-x-1">→</span>
        </Link>
      </CardContent>
    </Card>
  );
}
