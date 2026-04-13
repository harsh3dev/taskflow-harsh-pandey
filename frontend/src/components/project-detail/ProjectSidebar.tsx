import { Link } from "react-router-dom";
import { cn } from "../../lib/utils";
import { useSidebar } from "../../app/sidebar";
import { Project, ProjectStats } from "../../lib/types";

type ProjectSidebarProps = {
  project: Project;
  stats: ProjectStats | null;
  roleLabel: string;
  allTaskCount: number;
  visibleTaskCount: number;
};

export function ProjectSidebar({
  project,
  stats,
  roleLabel,
  allTaskCount,
  visibleTaskCount
}: ProjectSidebarProps) {
  const { isOpen, close } = useSidebar();

  return (
    <aside
      className={cn(
        // Shared styles
        "flex flex-col gap-6 overflow-y-auto border-r border-border bg-card p-5",
        // Mobile: fixed slide-out drawer
        "fixed inset-y-0 left-0 z-50 w-72 shadow-xl transition-transform duration-200 md:shadow-none",
        isOpen ? "translate-x-0" : "-translate-x-full",
        // Desktop: sticky sidebar (always visible)
        "md:static md:sticky md:top-[60px] md:h-[calc(100vh-60px)] md:w-60 md:translate-x-0 md:transition-none"
      )}
    >
      <div className="flex flex-col gap-3">
        <Link
          className="text-sm font-medium text-primary hover:underline"
          to="/"
          onClick={close}
        >
          ← Back to projects
        </Link>
        <p className="text-[0.65rem] font-semibold uppercase tracking-[0.2em] text-primary">
          Project detail
        </p>
        <h2 className="text-xl font-semibold leading-tight tracking-tight">{project.name}</h2>
        {project.description ? (
          <p className="line-clamp-3 text-sm text-muted-foreground">{project.description}</p>
        ) : (
          <p className="text-sm italic text-muted-foreground/60">No description yet.</p>
        )}
      </div>

      <div className="flex flex-col gap-1">
        <p className="mb-2 text-[0.65rem] font-semibold uppercase tracking-[0.2em] text-muted-foreground">
          Stats
        </p>
        <StatRow label="Visible" value={visibleTaskCount} />
        <StatRow label="Total" value={allTaskCount} />
        {stats ? (
          <>
            <StatRow label="Done" value={stats.status_counts.done ?? 0} color="success" />
            <StatRow label="In progress" value={stats.status_counts.in_progress ?? 0} color="warning" />
            <StatRow label="To do" value={stats.status_counts.todo ?? 0} />
          </>
        ) : null}
        <StatRow label="Role" value={roleLabel} />
      </div>
    </aside>
  );
}

function StatRow({
  label,
  value,
  color
}: {
  label: string;
  value: number | string;
  color?: "success" | "warning";
}) {
  return (
    <div className="flex items-center justify-between rounded-md px-2 py-1.5 text-sm hover:bg-muted/50">
      <span className="text-muted-foreground">{label}</span>
      <span
        className={cn(
          "font-semibold tabular-nums",
          color === "success" && "text-success",
          color === "warning" && "text-warning"
        )}
      >
        {value}
      </span>
    </div>
  );
}
