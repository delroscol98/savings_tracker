import type { User } from "./schemas";

export function setStoredUser(user: User): void {
  localStorage.setItem("user", JSON.stringify(user));
}

export function getStoredUser(): User {
  const user = localStorage.getItem("user");
  if (user == null) {
    return {} as User;
  }
  return JSON.parse(user);
}

export function clearStoredUser(): void {
  localStorage.removeItem("user");
}

export function getStoredToken(): string {
  const token = localStorage.getItem("token");
  if (token == null) {
    return "";
  }

  return token;
}

export function setStoredToken(token: string): void {
  localStorage.setItem("token", token);
}

export function clearStoredToken(): void {
  localStorage.removeItem("token");
}
