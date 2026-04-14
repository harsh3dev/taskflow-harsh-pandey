import { useDraggable } from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import { Card, CardContent } from "../ui/card";
import { cn } from "../../lib/utils";
import { formatDate, labelForPriority } from "../../lib/utils";
import { Task, TaskStatus, User } from "../../lib/types";

type TaskCardProps = {
  task: Task;
  currentUserId?: string;
  userMap: Map<string, User>;
  deletingTaskId: string | null;
  statusSavingId: string | null;
  onStatusChange: (task: Task, status: TaskStatus) => void;
  onCardClick: (task: Task) => void;
  onDeleteTask: (taskId: string) => void;
};

const priorityClasses: Record<string, string> = {
  high: "border-destructive/30 bg-destructive/10 text-destructive",
  medium: "border-border bg-muted/60 text-muted-foreground",
  low: "border-success/30 bg-success/10 text-success"
};

function initials(name: string) {
  return name
    .split(" ")
    .map((w) => w[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

export function TaskCard({
  task,
  currentUserId,
  userMap,
  deletingTaskId,
  onCardClick
}: TaskCardProps) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({
    id: task.id,
    data: { task }
  });

  const style = transform
    ? { transform: CSS.Translate.toString(transform), opacity: isDragging ? 0.45 : 1 }
    : undefined;

  const assignee = task.assignee_id ? userMap.get(task.assignee_id) : null;
  const assigneeName = assignee?.name ?? (task.assignee_id === currentUserId ? "You" : null);
  const isDeleting = deletingTaskId === task.id;

  return (
    <div 
      ref={setNodeRef} 
      style={style} 
      {...attributes} 
      {...listeners}
      className={cn("touch-none", isDragging && "z-50 relative")}
    >
      <Card
        className={cn(
          "cursor-pointer transition-shadow hover:shadow-md",
          isDragging ? "cursor-grabbing shadow-lg" : "cursor-grab",
          isDeleting && "opacity-40 pointer-events-none"
        )}
        onClick={() => onCardClick(task)}
      >
        <CardContent className="flex flex-col gap-2 p-3">
          {/* Top row: drag handle + priority */}
          <div className="flex items-center justify-between gap-2">
            <div className="text-muted-foreground/30 pointer-events-none">
              <DragIcon />
            </div>
            <span
              className={cn(
                "rounded-full border px-2 py-0.5 text-[0.68rem] font-medium",
                priorityClasses[task.priority] ?? priorityClasses.medium
              )}
            >
              {labelForPriority(task.priority)}
            </span>
          </div>

          {/* Title */}
          <p className="line-clamp-2 text-sm font-medium leading-snug">{task.title}</p>

          {/* Bottom meta */}
          <div className="flex items-center gap-2 text-[0.72rem] text-muted-foreground">
            {assigneeName ? (
              <span
                className="grid size-5 shrink-0 place-items-center rounded-full bg-primary/15 text-[0.55rem] font-bold text-primary"
                title={assigneeName}
              >
                {initials(assigneeName)}
              </span>
            ) : (
              <span className="grid size-5 shrink-0 place-items-center rounded-full bg-muted text-[0.55rem] text-muted-foreground">
                —
              </span>
            )}
            {task.due_date ? (
              <span className="truncate">Due {formatDate(task.due_date)}</span>
            ) : (
              <span className="text-muted-foreground/40">No due date</span>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function DragIcon() {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 12 12"
      fill="currentColor"
      aria-hidden="true"
    >
      {[0, 1, 2].map((col) =>
        [0, 1].map((row) => (
          <circle key={`${col}-${row}`} cx={2 + col * 4} cy={3 + row * 6} r="1.25" />
        ))
      )}
    </svg>
  );
}
