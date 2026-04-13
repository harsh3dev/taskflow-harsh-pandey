import { TaskPriority, TaskStatus } from "./types";

export const AUTH_STORAGE_KEY = "taskflow.auth";
export const API_BASE_URL = (
  import.meta.env.VITE_API_BASE_URL?.trim() || "/api"
).replace(/\/$/, "");

export const statusOptions: Array<{ value: TaskStatus; label: string }> = [
  { value: "todo", label: "To do" },
  { value: "in_progress", label: "In progress" },
  { value: "done", label: "Done" }
];

export const priorityOptions: Array<{ value: TaskPriority; label: string }> = [
  { value: "low", label: "Low" },
  { value: "medium", label: "Medium" },
  { value: "high", label: "High" }
];
