import { FormEvent, useState } from "react";
import { useApi, useAuth } from "../../app/auth";
import { Alert, AlertDescription } from "../ui/alert";
import { Button } from "../ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { ApiError } from "../../lib/api";
import { priorityOptions, statusOptions } from "../../lib/constants";
import { createTask, updateTask } from "../../lib/services/tasks";
import { getErrorMessage, toDateInputValue } from "../../lib/utils";
import { Task, TaskPriority, TaskStatus, User } from "../../lib/types";
import { FormDescription, FormField, FormMessage } from "../ui/form-field";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Select } from "../ui/select";
import { Textarea } from "../ui/textarea";

type TaskModalProps = {
  mode: "create" | "edit";
  projectId: string;
  task: Task | null;
  users: User[];
  onClose: () => void;
  onSaved: () => void;
  onDelete?: () => void;
};

export function TaskModal({
  mode,
  projectId,
  task,
  users,
  onClose,
  onSaved,
  onDelete
}: TaskModalProps) {
  const api = useApi();
  const { user } = useAuth();
  const [form, setForm] = useState({
    title: task?.title ?? "",
    description: task?.description ?? "",
    status: task?.status ?? "todo",
    priority: task?.priority ?? "medium",
    assignee_id: task?.assignee_id ?? "",
    due_date: toDateInputValue(task?.due_date ?? "")
  });
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [errorMessage, setErrorMessage] = useState("");
  const [submitting, setSubmitting] = useState(false);

  function validate() {
    const nextErrors: Record<string, string> = {};
    if (!form.title.trim()) {
      nextErrors.title = "Title is required";
    }
    if (form.due_date && !/^\d{4}-\d{2}-\d{2}$/.test(form.due_date)) {
      nextErrors.due_date = "Use YYYY-MM-DD";
    }
    return nextErrors;
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const errors = validate();
    setFieldErrors(errors);
    setErrorMessage("");
    if (Object.keys(errors).length > 0) {
      return;
    }

    setSubmitting(true);
    const payload = {
      title: form.title.trim(),
      description: form.description.trim(),
      status: form.status,
      priority: form.priority,
      assignee_id: form.assignee_id.trim() || null,
      due_date: form.due_date || null
    };

    try {
      if (mode === "create") {
        await createTask(api, projectId, payload);
      } else if (task) {
        await updateTask(api, task.id, payload);
      }
      onSaved();
    } catch (error) {
      if (error instanceof ApiError && error.fields) {
        setFieldErrors(error.fields);
      }
      setErrorMessage(getErrorMessage(error, "Unable to save the task."));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-20 grid place-items-center bg-black/50 p-4"
      role="presentation"
    >
      <Card className="max-h-[calc(100vh-2rem)] w-full max-w-[680px] overflow-auto rounded-xl shadow-2xl sm:rounded-2xl">
        <CardHeader className="flex flex-row items-start justify-between gap-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.24em] text-primary">
              {mode === "create" ? "Create task" : "Edit task"}
            </p>
            <CardTitle>{mode === "create" ? "Add a new task" : "Update task details"}</CardTitle>
            <CardDescription>
              Pick an assignee from the backend user directory, or leave the task unassigned.
            </CardDescription>
          </div>
          <div className="flex shrink-0 items-center gap-2">
            {mode === "edit" && onDelete ? (
              <Button variant="destructive" size="sm" onClick={onDelete} type="button">
                Delete
              </Button>
            ) : null}
            <Button variant="ghost" size="sm" onClick={onClose} type="button">
              Close
            </Button>
          </div>
        </CardHeader>

        <CardContent>
          <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
            {errorMessage ? (
              <Alert variant="destructive">
                <AlertDescription>{errorMessage}</AlertDescription>
              </Alert>
            ) : null}

            <FormField>
              <Label htmlFor="task-title">Title</Label>
              <Input
                id="task-title"
                value={form.title}
                onChange={(event) =>
                  setForm((current) => ({ ...current, title: event.target.value }))
                }
                placeholder="Draft API release notes"
              />
              {fieldErrors.title ? <FormMessage>{fieldErrors.title}</FormMessage> : null}
            </FormField>

            <FormField>
              <Label htmlFor="task-description">Description</Label>
              <Textarea
                id="task-description"
                value={form.description}
                onChange={(event) =>
                  setForm((current) => ({ ...current, description: event.target.value }))
                }
                placeholder="Context, expected output, and review notes."
              />
            </FormField>

            <div className="flex flex-col gap-4 sm:flex-row">
              <FormField className="flex-1">
                <Label htmlFor="task-status">Status</Label>
                <Select
                  id="task-status"
                  value={form.status}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      status: event.target.value as TaskStatus
                    }))
                  }
                >
                  {statusOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </Select>
              </FormField>

              <FormField className="flex-1">
                <Label htmlFor="task-priority">Priority</Label>
                <Select
                  id="task-priority"
                  value={form.priority}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      priority: event.target.value as TaskPriority
                    }))
                  }
                >
                  {priorityOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </Select>
              </FormField>
            </div>

            <FormField>
              <Label htmlFor="task-assignee">Assignee</Label>
              <Select
                id="task-assignee"
                value={form.assignee_id}
                onChange={(event) =>
                  setForm((current) => ({ ...current, assignee_id: event.target.value }))
                }
              >
                <option value="">Unassigned</option>
                {users.map((candidate) => (
                  <option key={candidate.id} value={candidate.id}>
                    {candidate.name}
                    {candidate.id === user?.id ? " (You)" : ""}
                    {" · "}
                    {candidate.email}
                  </option>
                ))}
              </Select>
              <div className="flex flex-wrap gap-3">
                <Button
                  variant="outline"
                  onClick={() =>
                    setForm((current) => ({ ...current, assignee_id: user?.id ?? "" }))
                  }
                  type="button"
                >
                  Assign to me
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setForm((current) => ({ ...current, assignee_id: "" }))}
                  type="button"
                >
                  Clear assignee
                </Button>
              </div>
              {fieldErrors.assignee_id ? (
                <FormMessage>{fieldErrors.assignee_id}</FormMessage>
              ) : (
                <FormDescription>
                  Choose from users returned by the authenticated backend directory.
                </FormDescription>
              )}
            </FormField>

            <FormField>
              <Label htmlFor="task-due-date">Due date</Label>
              <Input
                id="task-due-date"
                type="date"
                value={form.due_date}
                onChange={(event) =>
                  setForm((current) => ({ ...current, due_date: event.target.value }))
                }
              />
              {fieldErrors.due_date ? (
                <FormMessage>{fieldErrors.due_date}</FormMessage>
              ) : (
                <FormDescription>Optional. Stored as a date-only value.</FormDescription>
              )}
            </FormField>

            <div className="flex flex-wrap gap-3">
              <Button disabled={submitting} type="submit">
                {submitting ? "Saving..." : mode === "create" ? "Create task" : "Save changes"}
              </Button>
              <Button variant="outline" onClick={onClose} type="button">
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
