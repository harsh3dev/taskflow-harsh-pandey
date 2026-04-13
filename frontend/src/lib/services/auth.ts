import { apiRequest } from "../api";
import { AuthResponse } from "../types";

export function login(payload: { email: string; password: string }): Promise<AuthResponse> {
  return apiRequest<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function register(payload: {
  name: string;
  email: string;
  password: string;
}): Promise<AuthResponse> {
  return apiRequest<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
