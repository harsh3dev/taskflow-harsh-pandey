import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { priorityOptions, statusOptions } from "./constants";
import { ApiError } from "./api";
import { TaskPriority, TaskStatus } from "./types";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function getErrorMessage(error: unknown, fallback: string) {
  if (error instanceof ApiError) {
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}

export function labelForStatus(status: TaskStatus) {
  return statusOptions.find((option) => option.value === status)?.label ?? status;
}

export function labelForPriority(priority: TaskPriority) {
  return priorityOptions.find((option) => option.value === priority)?.label ?? priority;
}

export function formatDateTime(value: string) {
  return new Date(value).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric"
  });
}

export function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric"
  });
}

export function abbreviateId(value: string) {
  if (value.length <= 10) {
    return value;
  }
  return `${value.slice(0, 8)}…${value.slice(-4)}`;
}

export function toDateInputValue(value: string) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return date.toISOString().slice(0, 10);
}
