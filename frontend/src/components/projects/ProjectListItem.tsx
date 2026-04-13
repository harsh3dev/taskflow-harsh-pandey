import { Link } from "react-router-dom";
import { Badge } from "../ui/badge";
import { Project } from "../../lib/types";
import { formatDateTime } from "../../lib/utils";

type ProjectListItemProps = {
  project: Project;
  currentUserId?: string;
};

export function ProjectListItem({ project, currentUserId }: ProjectListItemProps) {
  const isOwner = project.owner_id === currentUserId;

  return (
    <div className="group flex flex-col gap-2 border-b border-border px-4 py-3 transition-colors last:border-0 hover:bg-muted/40 sm:flex-row sm:items-center sm:gap-4">
      {/* Badge */}
      <Badge variant="secondary" className="w-fit shrink-0 text-[0.7rem]">
        {isOwner ? "Owner" : "Contributor"}
      </Badge>

      {/* Name + description */}
      <div className="min-w-0 flex-1">
        <p className="truncate font-medium">{project.name}</p>
        <p className="truncate text-xs text-muted-foreground">
          {project.description || "No description yet."}
        </p>
      </div>

      {/* Date — hidden on mobile */}
      <span className="hidden shrink-0 text-xs text-muted-foreground sm:block">
        {formatDateTime(project.created_at)}
      </span>

      {/* Open link */}
      <Link
        className="w-fit shrink-0 text-sm font-semibold text-primary transition-transform duration-150 group-hover:translate-x-0.5"
        to={`/projects/${project.id}`}
      >
        Open →
      </Link>
    </div>
  );
}
