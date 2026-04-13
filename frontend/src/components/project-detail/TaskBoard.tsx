import { Card, CardContent } from "../ui/card";
import { Task, TaskStatus, User } from "../../lib/types";
import { TaskColumn } from "./TaskColumn";

type TaskBoardProps = {
  loading: boolean;
  tasks: Task[];
  columns: Array<{ value: string; label: string; tasks: Task[] }>;
  deletingTaskId: string | null;
  statusSavingId: string | null;
  currentUserId?: string;
  userMap: Map<string, User>;
  onStatusChange: (task: Task, status: TaskStatus) => void;
  onEditTask: (task: Task) => void;
  onDeleteTask: (taskId: string) => void;
};

export function TaskBoard({
  loading,
  tasks,
  columns,
  deletingTaskId,
  statusSavingId,
  currentUserId,
  userMap,
  onStatusChange,
  onEditTask,
  onDeleteTask
}: TaskBoardProps) {
  if (loading) {
    return <div className="py-8 text-[var(--ink-soft)]">Refreshing tasks...</div>;
  }

  if (tasks.length === 0) {
    return (
      <Card>
        <CardContent className="px-8 py-8 text-center">
          <h3>No matching tasks</h3>
          <p className="text-[var(--ink-soft)]">Adjust the filters or add a new task to this project.</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <section className="grid grid-cols-1 items-start gap-4 xl:grid-cols-3">
      {columns.map((column) => (
        <TaskColumn
          column={column}
          currentUserId={currentUserId}
          deletingTaskId={deletingTaskId}
          key={column.value}
          onDeleteTask={onDeleteTask}
          onEditTask={onEditTask}
          onStatusChange={onStatusChange}
          statusSavingId={statusSavingId}
          userMap={userMap}
        />
      ))}
    </section>
  );
}
