import { apiRequest } from "./client";

export type User = { id: number; username: string; email: string };
type AuthResponse = { user: User };

export const authApi = {
  current: () => apiRequest<AuthResponse>("/me"),
  login: (identity: string, password: string) => apiRequest<AuthResponse>("/auth/login", { identity, password }),
  register: (username: string, email: string, password: string) =>
    apiRequest<AuthResponse>("/auth/register", { username, email, password }),
  logout: () => apiRequest<never>("/auth/logout", {}),
};
