import { ReactNode, createContext, useContext, useEffect, useState } from "react";
import { apiRequest, ApiError } from "../lib/api";
import { AUTH_STORAGE_KEY } from "../lib/constants";
import { User } from "../lib/types";
import * as authService from "../lib/services/auth";

type AuthContextValue = {
  token: string | null;
  user: User | null;
  login: (payload: { email: string; password: string }) => Promise<void>;
  register: (payload: {
    name: string;
    email: string;
    password: string;
  }) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

function readStoredAuth() {
  const fallback = { token: null as string | null, user: null as User | null };
  const raw = window.localStorage.getItem(AUTH_STORAGE_KEY);
  if (!raw) {
    return fallback;
  }

  try {
    const parsed = JSON.parse(raw) as { token?: string; user?: User };
    return {
      token: parsed.token ?? null,
      user: parsed.user ?? null
    };
  } catch {
    window.localStorage.removeItem(AUTH_STORAGE_KEY);
    return fallback;
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => readStoredAuth().token);
  const [user, setUser] = useState<User | null>(() => readStoredAuth().user);

  useEffect(() => {
    if (token && user) {
      window.localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify({ token, user }));
      return;
    }
    window.localStorage.removeItem(AUTH_STORAGE_KEY);
  }, [token, user]);

  async function handleAuth<T extends { token: string; user: User }, P extends object>(
    action: (payload: P) => Promise<T>,
    body: P
  ) {
    const response = await action(body);
    setToken(response.token);
    setUser(response.user);
  }

  const value: AuthContextValue = {
    token,
    user,
    login: (payload) => handleAuth(authService.login, payload),
    register: (payload) => handleAuth(authService.register, payload),
    logout: () => {
      setToken(null);
      setUser(null);
    }
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("Auth context is unavailable");
  }
  return context;
}

export function useApi() {
  const { token, logout } = useAuth();

  return async function request<T>(path: string, options: RequestInit = {}) {
    try {
      return await apiRequest<T>(path, options, token);
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        logout();
      }
      throw error;
    }
  };
}
