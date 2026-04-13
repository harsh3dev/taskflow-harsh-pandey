import { Badge } from "../ui/badge";
import { Button } from "../ui/button";
import { Card, CardContent } from "../ui/card";
import { FormField } from "../ui/form-field";
import { Label } from "../ui/label";
import { Select } from "../ui/select";
import { statusOptions } from "../../lib/constants";
import { cn } from "../../lib/utils";
import { formatDate, labelForPriority, labelForStatus } from "../../lib/utils";
import { Task, TaskStatus, User } from "../../lib/types";

type TaskCardProps = {
  task: Task;
  currentUserId?: string;
  userMap: Map<string, User>;
  deletingTaskId: string | null;
  statusSavingId: string | null;
  onStatusChange: (task: Task, status: TaskStatus) => void;
  onEditTask: (task: Task) => void;
  onDeleteTask: (taskId: string) => void;
};

export function TaskCard({
  task,
  currentUserId,
  userMap,
  deletingTaskId,
  statusSavingId,
  onStatusChange,
  onEditTask,
  onDeleteTask
}: TaskCardProps) {
  const statusChipClassName = cn(
    "inline-flex items-center gap-[0.35rem] rounded-full px-[0.65rem] py-[0.36rem] text-[0.78rem] font-medium text-white",
    task.status === "done" && "bg-[var(--success)]",
    task.status === "in_progress" && "bg-[var(--warn)]",
    task.status === "todo" && "bg-[var(--panel)]"
  );

  const priorityBadgeClassName = cn(
    task.priority === "high" && "border-[rgba(177,69,62,0.28)] bg-[rgba(177,69,62,0.12)] text-[var(--danger)]",
    task.priority === "low" && "border-[rgba(47,124,88,0.28)] bg-[rgba(47,124,88,0.12)] text-[var(--success)]"
  );

  return (
    <Card>
      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <span className={statusChipClassName}>
            {labelForStatus(task.status)}
          </span>
          <Badge className={priorityBadgeClassName} variant="outline">
            {labelForPriority(task.priority)}
          </Badge>
        </div>

        <div className="flex flex-col gap-2">
          <h4 className="m-0 text-xl font-semibold tracking-tight">{task.title}</h4>
          <p className="text-sm text-[var(--ink-soft)]">
            {task.description || "No task description yet."}
          </p>
        </div>

        <div className="flex flex-wrap gap-[0.55rem] text-[0.88rem] text-[var(--ink-soft)]">
          <span>
            Assignee:{" "}
            {task.assignee_id
              ? userMap.get(task.assignee_id)?.name ||
                (task.assignee_id === currentUserId ? "You" : task.assignee_id)
              : "Unassigned"}
          </span>
          <span>Due: {task.due_date ? formatDate(task.due_date) : "No date"}</span>
        </div>

        <div className="flex flex-wrap gap-3">
          <FormField>
            <Label htmlFor={`status-${task.id}`}>Move task</Label>
            <Select
              id={`status-${task.id}`}
              disabled={statusSavingId === task.id}
              value={task.status}
              onChange={(event) => onStatusChange(task, event.target.value as TaskStatus)}
            >
              {statusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </Select>
          </FormField>
        </div>

        <div className="flex flex-wrap gap-3">
          <Button variant="outline" onClick={() => onEditTask(task)} type="button">
            Edit
          </Button>
          <Button
            variant="destructive"
            disabled={deletingTaskId === task.id}
            onClick={() => onDeleteTask(task.id)}
            type="button"
          >
            {deletingTaskId === task.id ? "Deleting..." : "Delete"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
