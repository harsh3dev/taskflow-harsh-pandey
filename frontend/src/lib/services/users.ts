import { ApiRequest } from "../request";
import { User } from "../types";

export function listUsers(
  request: ApiRequest,
  search?: string
): Promise<{ users: User[] }> {
  const params = new URLSearchParams();
  if (search?.trim()) {
    params.set("q", search.trim());
  }
  const query = params.toString();
  return request<{ users: User[] }>(`/users${query ? `?${query}` : ""}`);
}
