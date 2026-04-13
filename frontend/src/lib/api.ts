import { API_BASE_URL } from "./constants";
import { ApiErrorShape } from "./types";

export class ApiError extends Error {
  status: number;
  fields?: Record<string, string>;

  constructor(status: number, message: string, fields?: Record<string, string>) {
    super(message);
    this.status = status;
    this.fields = fields;
  }
}

export async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
  token?: string | null
): Promise<T> {
  const headers = new Headers(options.headers ?? {});
  if (options.body !== undefined && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  headers.set("Accept", "application/json");
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers
  });

  const text = await response.text();
  const trimmed = text.trim();
  const payload = trimmed ? (JSON.parse(trimmed) as ApiErrorShape & T) : ({} as T);

  if (!response.ok) {
    const apiError = payload as ApiErrorShape;
    throw new ApiError(
      response.status,
      apiError.error || "Request failed",
      apiError.fields
    );
  }

  return payload as T;
}
