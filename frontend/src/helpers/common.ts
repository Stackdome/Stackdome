import type { User } from "@/api/users";
import { AUTH_SESSION_CHANGED } from "@/helpers/auth-events";

function notifyAuthSessionChanged() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(AUTH_SESSION_CHANGED));
  }
}

export function isUserLoggedIn(): boolean {
  return Boolean(localStorage.getItem('authToken'));
}

export function getCurrentUser(): User | null {
  const user = localStorage.getItem('currentUser');
  if (!user) return null;
  try {
    return JSON.parse(user) as User;
  } catch {
    console.error("Failed to parse current user from localStorage");
    return null;
  }
}

export function getCurrentOrganizationId(): string | null {
  const user = getCurrentUser();
  return user?.organisation_id ?? null;
}

export function getRefreshToken(): string | null {
  return localStorage.getItem('refreshToken');
}

export function setAuthSession(token: string, user: User, refreshToken?: string) {
  localStorage.setItem('authToken', token);
  localStorage.setItem('currentUser', JSON.stringify(user));
  if (refreshToken) {
    localStorage.setItem('refreshToken', refreshToken);
  }
  notifyAuthSessionChanged();
}

export function clearAuthSession() {
  localStorage.removeItem('authToken');
  localStorage.removeItem('currentUser');
  localStorage.removeItem('refreshToken');
  notifyAuthSessionChanged();
}

export function logoutAndRedirect(redirectTo: string = "/sign-in") {
  clearAuthSession();
  window.location.href = redirectTo;
}
