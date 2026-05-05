import { api, type User } from "./api";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("access_token");
}

export function setToken(token: string): void {
  localStorage.setItem("access_token", token);
}

export function clearToken(): void {
  localStorage.removeItem("access_token");
}

export async function login(email: string, password: string): Promise<User> {
  const data = await api.auth.login(email, password);
  setToken(data.token);
  return data.user;
}

export async function register(email: string, password: string): Promise<User> {
  const data = await api.auth.register(email, password);
  setToken(data.token);
  return data.user;
}

export async function logout(): Promise<void> {
  await api.auth.logout().catch(() => {});
  clearToken();
}
