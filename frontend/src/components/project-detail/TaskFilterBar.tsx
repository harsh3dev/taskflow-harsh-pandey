import { Button } from "../ui/button";
import { Select } from "../ui/select";
import { statusOptions } from "../../lib/constants";
import { User } from "../../lib/types";

type TaskFilterBarProps = {
  statusFilter: string;
  assigneeFilter: string;
  assigneeOptions: string[];
  currentUserId?: string;
  userMap: Map<string, User>;
  taskError: string;
  onStatusChange: (value: string) => void;
  onAssigneeChange: (value: string) => void;
  onCreateTask: () => void;
  onRefresh: () => void;
};

export function TaskFilterBar({
  statusFilter,
  assigneeFilter,
  assigneeOptions,
  currentUserId,
  userMap,
  onStatusChange,
  onAssigneeChange,
  onCreateTask,
  onRefresh
}: TaskFilterBarProps) {
  return (
    <div className="sticky top-[60px] z-4 flex flex-wrap items-center gap-2 border-b border-border bg-background px-4 py-2">
      <Select
        className="h-7 w-full rounded-md px-2 py-0 text-sm sm:w-36"
        value={statusFilter}
        onChange={(e) => onStatusChange(e.target.value)}
      >
        <option value="">All statuses</option>
        {statusOptions.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </Select>

      <Select
        className="h-7 w-full rounded-md px-2 py-0 text-sm sm:w-40"
        value={assigneeFilter}
        onChange={(e) => onAssigneeChange(e.target.value)}
      >
        <option value="">All assignees</option>
        {assigneeOptions.map((id) => (
          <option key={id} value={id}>
            {userMap.get(id)?.name || (id === currentUserId ? "You" : id)}
          </option>
        ))}
      </Select>

      <div className="ml-auto flex items-center gap-2">
        <Button size="sm" onClick={onCreateTask} type="button">
          + New task
        </Button>
        <Button
          size="icon-sm"
          variant="ghost"
          onClick={onRefresh}
          type="button"
          title="Refresh tasks"
          aria-label="Refresh tasks"
        >
          ↻
        </Button>
      </div>
    </div>
  );
}
