import { DndContext, DragEndEvent, PointerSensor, useSensor, useSensors } from "@dnd-kit/core";
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
  onCardClick: (task: Task) => void;
  onStatusChange: (task: Task, status: TaskStatus) => void;
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
  onCardClick,
  onStatusChange,
  onDeleteTask
}: TaskBoardProps) {
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
  );

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (!over) return;
    const task = active.data.current?.task as Task | undefined;
    const newStatus = over.id as TaskStatus;
    if (task && task.status !== newStatus) {
      onStatusChange(task, newStatus);
    }
  }

  if (loading) {
    return (
      <div className="flex flex-1 items-center justify-center py-16 text-sm text-muted-foreground">
        Refreshing tasks…
      </div>
    );
  }

  if (tasks.length === 0) {
    return (
      <div className="flex flex-1 items-center justify-center py-16 text-center">
        <div>
          <p className="text-base font-medium">No matching tasks</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Adjust the filters or add a new task to this project.
          </p>
        </div>
      </div>
    );
  }

  return (
    <DndContext sensors={sensors} onDragEnd={handleDragEnd}>
      <div className="grid flex-1 grid-cols-1 divide-y divide-border md:grid-cols-3 md:divide-x md:divide-y-0">
        {columns.map((column) => (
          <TaskColumn
            column={column}
            currentUserId={currentUserId}
            deletingTaskId={deletingTaskId}
            key={column.value}
            onCardClick={onCardClick}
            onDeleteTask={onDeleteTask}
            onStatusChange={onStatusChange}
            statusSavingId={statusSavingId}
            userMap={userMap}
          />
        ))}
      </div>
    </DndContext>
  );
}
