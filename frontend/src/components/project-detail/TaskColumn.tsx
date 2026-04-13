import { useDroppable } from "@dnd-kit/core";
import { Task, TaskStatus, User } from "../../lib/types";
import { TaskCard } from "./TaskCard";
import { cn } from "../../lib/utils";

type StatusDot = { color: string };
const statusDot: Record<string, StatusDot> = {
  todo: { color: "bg-muted-foreground/40" },
  in_progress: { color: "bg-warning" },
  done: { color: "bg-success" }
};

type TaskColumnProps = {
  column: { value: string; label: string; tasks: Task[] };
  currentUserId?: string;
  userMap: Map<string, User>;
  deletingTaskId: string | null;
  statusSavingId: string | null;
  onCardClick: (task: Task) => void;
  onStatusChange: (task: Task, status: TaskStatus) => void;
  onDeleteTask: (taskId: string) => void;
};

export function TaskColumn({
  column,
  currentUserId,
  userMap,
  deletingTaskId,
  statusSavingId,
  onCardClick,
  onStatusChange,
  onDeleteTask
}: TaskColumnProps) {
  const { setNodeRef, isOver } = useDroppable({ id: column.value });
  const dot = statusDot[column.value] ?? { color: "bg-muted-foreground/40" };

  return (
    <div className="flex flex-col">
      {/* JIRA-style column header */}
      <div className="flex items-center gap-2 border-b border-border px-4 py-2.5">
        <span className={cn("size-2 rounded-full shrink-0", dot.color)} />
        <span className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          {column.label}
        </span>
        <span className="ml-auto rounded-full bg-muted px-2 py-0.5 text-xs font-medium tabular-nums">
          {column.tasks.length}
        </span>
      </div>

      {/* Drop zone */}
      <div
        ref={setNodeRef}
        className={cn(
          "flex flex-1 flex-col gap-2 p-3 transition-colors min-h-[280px] sm:min-h-[400px]",
          isOver ? "bg-primary/5" : "bg-muted/20"
        )}
      >
        {column.tasks.length === 0 ? (
          <div className="m-2 rounded-lg border-2 border-dashed border-border/50 p-6 text-center text-sm text-muted-foreground">
            {isOver ? "Drop here" : "No tasks"}
          </div>
        ) : null}

        {column.tasks.map((task) => (
          <TaskCard
            currentUserId={currentUserId}
            deletingTaskId={deletingTaskId}
            key={task.id}
            onCardClick={onCardClick}
            onDeleteTask={onDeleteTask}
            onStatusChange={onStatusChange}
            statusSavingId={statusSavingId}
            task={task}
            userMap={userMap}
          />
        ))}
      </div>
    </div>
  );
}
