import { ApiRequest } from "../request";
import { Project, ProjectStats, Task } from "../types";

export function listProjects(request: ApiRequest): Promise<{ projects: Project[] }> {
  return request<{ projects: Project[] }>("/projects");
}

export function createProject(
  request: ApiRequest,
  payload: { name: string; description: string }
): Promise<{ project: Project }> {
  return request<{ project: Project }>("/projects", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function getProject(
  request: ApiRequest,
  projectId: string
): Promise<{ project: Project; tasks: Task[] }> {
  return request<{ project: Project; tasks: Task[] }>(`/projects/${projectId}`);
}

export function getProjectStats(
  request: ApiRequest,
  projectId: string
): Promise<ProjectStats> {
  return request<ProjectStats>(`/projects/${projectId}/stats`);
}
