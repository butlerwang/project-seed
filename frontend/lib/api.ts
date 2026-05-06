const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = typeof window !== "undefined" ? localStorage.getItem("access_token") : null;
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string>),
  };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`${BASE}${path}`, { ...init, headers, credentials: "include" });

  if (res.status === 401) {
    const refreshed = await tryRefresh();
    if (refreshed) return request<T>(path, init);
    throw new ApiError(401, "unauthorized");
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, (body as Record<string, string>).error ?? res.statusText);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

async function tryRefresh(): Promise<boolean> {
  const res = await fetch(`${BASE}/api/v1/auth/refresh`, {
    method: "POST",
    credentials: "include",
  });
  if (!res.ok) {
    localStorage.removeItem("access_token");
    return false;
  }
  const data = (await res.json()) as { token: string };
  localStorage.setItem("access_token", data.token);
  return true;
}

export interface User {
  id: string;
  email: string;
  role: "user" | "admin";
  plan: "free" | "pro";
  created_at: string;
}

export const api = {
  auth: {
    register: (email: string, password: string) =>
      request<{ token: string; user: User }>("/api/v1/auth/register", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      }),
    login: (email: string, password: string) =>
      request<{ token: string; user: User }>("/api/v1/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      }),
    forgotPassword: (email: string) =>
      request<void>("/api/v1/auth/forgot-password", {
        method: "POST",
        body: JSON.stringify({ email }),
      }),
    resetPassword: (token: string, password: string) =>
      request<void>("/api/v1/auth/reset-password", {
        method: "POST",
        body: JSON.stringify({ token, password }),
      }),
    logout: () => request<void>("/api/v1/auth/logout", { method: "POST" }),
    me: () => request<User>("/api/v1/auth/me"),
  },
  admin: {
    listUsers: (limit = 50, offset = 0) =>
      request<{ users: User[]; limit: number; offset: number }>(
        `/api/v1/admin/users?limit=${limit}&offset=${offset}`
      ),
  },
};
