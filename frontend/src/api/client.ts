// Generic axios API client setup
import axios, { AxiosError } from 'axios';
import type { components } from './types/openapi';

// OpenAPI Error types
export type ApiError = components["schemas"]["Error"];
export type ApiErrorList = components["schemas"]["ErrorList"];

// Combined error type that represents any API error
export type AppError = AxiosError<ApiError> | AxiosError<ApiErrorList> | Error;

export function isAxiosApiError(error: unknown): error is AxiosError<ApiError> {
  return error instanceof AxiosError && error.response?.data != null;
}

export function isAxiosApiErrorList(error: unknown): error is AxiosError<ApiErrorList> {
  return error instanceof AxiosError &&
    error.response?.data != null &&
    'items' in error.response.data;
}

export function isAxiosError(error: unknown): error is AxiosError {
  return error instanceof AxiosError;
}

export function getErrorMessage(error: unknown): string {
  if (isAxiosApiError(error)) {
    return error.response?.data?.reason ||
      error.message ||
      "An API error occurred";
  }

  if (isAxiosApiErrorList(error)) {
    const firstError = error.response?.data?.items?.[0];
    return firstError?.reason ||
      error.message ||
      "An API error occurred";
  }

  if (isAxiosError(error)) {
    return error.response?.statusText ||
      error.message ||
      "A network error occurred";
  }

  if (error instanceof Error) {
    return error.message;
  }

  return "An unknown error occurred";
}

export function getErrorStatus(error: unknown): number | undefined {
  if (isAxiosError(error)) {
    return error.response?.status;
  }
  return undefined;
}

export function isErrorStatus(error: unknown, status: number): boolean {
  return getErrorStatus(error) === status;
}

export function isNotFoundError(error: unknown): boolean {
  return isErrorStatus(error, 404);
}

export function isUnauthorizedError(error: unknown): boolean {
  return isErrorStatus(error, 401);
}

export function isForbiddenError(error: unknown): boolean {
  return isErrorStatus(error, 403);
}

export function isBadRequestError(error: unknown): boolean {
  return isErrorStatus(error, 400);
}

export function isServerError(error: unknown): boolean {
  const status = getErrorStatus(error);
  return status != null && status >= 500;
}

const api = axios.create({
  baseURL: (import.meta.env.VITE_API_BASE_URL || '/api/v1'),
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add interceptor to include Authorization header and cookie if token exists
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('authToken');
  if (token) {
    config.headers = config.headers || {};
    config.headers['Authorization'] = `Bearer ${token}`;
    document.cookie = `auth_token=${token}; path=/; secure; samesite=strict`;
  }
  return config;
});

// Only 401 (unauthenticated — bad/expired token) resets the session. A 403 means
// the session is valid but the caller lacks permission for that resource (RBAC),
// e.g. an OrgMember hitting an admin-only endpoint. Logging out on 403 would lock
// members out of the whole app, so we let the calling code surface it in-page.
// Skip on auth pages so a wrong-password 401 shows inline instead of refresh-looping.
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status;
    const path = window.location.pathname;
    const onAuthPage = path === '/sign-in' || path === '/sign-up';
    if (status === 401 && !onAuthPage) {
      localStorage.removeItem('authToken');
      localStorage.removeItem('currentUser');
      window.location.href = '/sign-in';
    }
    return Promise.reject(error);
  }
);

export default api;
