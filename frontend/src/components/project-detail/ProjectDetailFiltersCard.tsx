import { Alert, AlertDescription } from "../ui/alert";
import { Button } from "../ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { FormField } from "../ui/form-field";
import { Label } from "../ui/label";
import { Select } from "../ui/select";
import { statusOptions } from "../../lib/constants";
import { User } from "../../lib/types";

type ProjectDetailFiltersCardProps = {
  taskError: string;
  statusFilter: string;
  assigneeFilter: string;
  assigneeOptions: string[];
  currentUserId?: string;
  userMap: Map<string, User>;
  onStatusChange: (value: string) => void;
  onAssigneeChange: (value: string) => void;
  onCreateTask: () => void;
  onRefresh: () => void;
};

export function ProjectDetailFiltersCard({
  taskError,
  statusFilter,
  assigneeFilter,
  assigneeOptions,
  currentUserId,
  userMap,
  onStatusChange,
  onAssigneeChange,
  onCreateTask,
  onRefresh
}: ProjectDetailFiltersCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <CardTitle>Filters and actions</CardTitle>
          <CardDescription>
            Filter by task status or assignee. Status changes apply immediately and roll back on
            API failure.
          </CardDescription>
        </div>
        <div className="flex flex-wrap gap-3">
          <Button onClick={onCreateTask} type="button">
            New task
          </Button>
          <Button variant="outline" onClick={onRefresh} type="button">
            Refresh
          </Button>
        </div>
      </CardHeader>

      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-wrap gap-4">
          <FormField>
            <Label htmlFor="status-filter">Status</Label>
            <Select
              id="status-filter"
              value={statusFilter}
              onChange={(event) => onStatusChange(event.target.value)}
            >
              <option value="">All statuses</option>
              {statusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </Select>
          </FormField>

          <FormField>
            <Label htmlFor="assignee-filter">Assignee</Label>
            <Select
              id="assignee-filter"
              value={assigneeFilter}
              onChange={(event) => onAssigneeChange(event.target.value)}
            >
              <option value="">All assignees</option>
              {assigneeOptions.map((assigneeId) => (
                <option key={assigneeId} value={assigneeId}>
                  {userMap.get(assigneeId)?.name ||
                    (assigneeId === currentUserId ? "You" : assigneeId)}
                </option>
              ))}
            </Select>
          </FormField>
        </div>

        {taskError ? (
          <Alert variant="destructive">
            <AlertDescription>{taskError}</AlertDescription>
          </Alert>
        ) : null}
      </CardContent>
    </Card>
  );
}
