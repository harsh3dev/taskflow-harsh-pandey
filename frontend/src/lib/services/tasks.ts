import { ApiRequest } from "../request";
import { Task, TaskPriority, TaskStatus } from "../types";

type TaskPayload = {
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  assignee_id: string | null;
  due_date: string | null;
};

export function listProjectTasks(
  request: ApiRequest,
  projectId: string,
  filters?: { status?: string; assignee?: string }
): Promise<{ tasks: Task[] }> {
  const params = new URLSearchParams();
  if (filters?.status) {
    params.set("status", filters.status);
  }
  if (filters?.assignee) {
    params.set("assignee", filters.assignee);
  }
  const query = params.toString();
  return request<{ tasks: Task[] }>(
    `/projects/${projectId}/tasks${query ? `?${query}` : ""}`
  );
}

export function createTask(
  request: ApiRequest,
  projectId: string,
  payload: TaskPayload
): Promise<{ task: Task }> {
  return request<{ task: Task }>(`/projects/${projectId}/tasks`, {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function updateTask(
  request: ApiRequest,
  taskId: string,
  payload: Partial<TaskPayload>
): Promise<{ task: Task }> {
  return request<{ task: Task }>(`/tasks/${taskId}`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}

export function deleteTask(request: ApiRequest, taskId: string): Promise<{ status: string }> {
  return request<{ status: string }>(`/tasks/${taskId}`, { method: "DELETE" });
}
