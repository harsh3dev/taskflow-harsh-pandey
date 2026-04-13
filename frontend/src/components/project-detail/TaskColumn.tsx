import { Badge } from "../ui/badge";
import { Card, CardContent } from "../ui/card";
import { Task, TaskStatus, User } from "../../lib/types";
import { TaskCard } from "./TaskCard";

type TaskColumnProps = {
  column: { value: string; label: string; tasks: Task[] };
  currentUserId?: string;
  userMap: Map<string, User>;
  deletingTaskId: string | null;
  statusSavingId: string | null;
  onStatusChange: (task: Task, status: TaskStatus) => void;
  onEditTask: (task: Task) => void;
  onDeleteTask: (taskId: string) => void;
};

export function TaskColumn({
  column,
  currentUserId,
  userMap,
  deletingTaskId,
  statusSavingId,
  onStatusChange,
  onEditTask,
  onDeleteTask
}: TaskColumnProps) {
  return (
    <div className="grid min-h-[240px] gap-4 rounded-[28px] border border-[var(--line)] bg-[rgba(255,253,248,0.64)] p-4">
      <div className="flex items-start justify-between gap-3">
        <h3 className="m-0 text-[1.35rem] font-semibold tracking-tight">{column.label}</h3>
        <Badge variant="outline">{column.tasks.length}</Badge>
      </div>

      {column.tasks.length === 0 ? (
        <Card>
          <CardContent>
            <p className="text-sm text-[var(--ink-soft)]">No tasks in this column.</p>
          </CardContent>
        </Card>
      ) : null}

      {column.tasks.map((task) => (
        <TaskCard
          currentUserId={currentUserId}
          deletingTaskId={deletingTaskId}
          key={task.id}
          onDeleteTask={onDeleteTask}
          onEditTask={onEditTask}
          onStatusChange={onStatusChange}
          statusSavingId={statusSavingId}
          task={task}
          userMap={userMap}
        />
      ))}
    </div>
  );
}
