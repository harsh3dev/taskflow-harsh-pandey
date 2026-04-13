import { cn } from "../../lib/utils";
import { useSidebar } from "../../app/sidebar";

type WorkspaceSidebarProps = {
  projectCount: number;
  ownedCount: number;
};

export function WorkspaceSidebar({ projectCount, ownedCount }: WorkspaceSidebarProps) {
  const { isOpen } = useSidebar();
  const contributedCount = projectCount - ownedCount;

  return (
    <aside
      className={cn(
        // Shared styles
        "flex flex-col gap-6 overflow-y-auto border-r border-border bg-card p-5",
        // Mobile: fixed slide-out drawer
        "fixed inset-y-0 left-0 z-50 w-72 shadow-xl transition-transform duration-200 md:shadow-none",
        isOpen ? "translate-x-0" : "-translate-x-full",
        // Desktop: sticky sidebar (always visible)
        "md:static md:sticky md:top-[60px] md:h-[calc(100vh-60px)] md:w-56 md:translate-x-0 md:transition-none"
      )}
    >
      <div className="flex flex-col gap-2">
        <p className="text-[0.65rem] font-semibold uppercase tracking-[0.2em] text-primary">
          Workspace
        </p>
        <h2 className="text-lg font-semibold leading-tight tracking-tight">
          Projects that matter today.
        </h2>
        <p className="text-xs text-muted-foreground">
          Your owned projects and work where you are assigned.
        </p>
      </div>

      <div className="flex flex-col gap-1">
        <p className="mb-2 text-[0.65rem] font-semibold uppercase tracking-[0.2em] text-muted-foreground">
          Overview
        </p>
        <StatRow label="Total" value={projectCount} />
        <StatRow label="Owned" value={ownedCount} color="primary" />
        <StatRow label="Contributed" value={contributedCount} />
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
  value: number;
  color?: "primary";
}) {
  return (
    <div className="flex items-center justify-between rounded-md px-2 py-1.5 text-sm hover:bg-muted/50">
      <span className="text-muted-foreground">{label}</span>
      <span
        className={cn(
          "font-semibold tabular-nums",
          color === "primary" && "text-primary"
        )}
      >
        {value}
      </span>
    </div>
  );
}
