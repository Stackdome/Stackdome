// Generic axios API client setup
import axios, { AxiosError } from 'axios';
import type { components } from './types/openapi';
import { API_BASE_URL } from './base-url';
import { refreshAccessToken } from './auth-refresh';
import { clearAuthSession } from '@/helpers/common';

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
  baseURL: API_BASE_URL,
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

export interface AuthErrorDeps {
  refresh: () => Promise<string>;
  retry: (config: unknown) => Promise<unknown>;
  onAuthFailure: () => void;
  isAuthPage?: () => boolean;
}

function isOnAuthPage(): boolean {
  const path = window.location.pathname;
  return path === '/sign-in' || path === '/sign-up';
}

// Expired/invalid access tokens return 403 here (not 401); match the token reason
// so RBAC 403s ("unauthorized to perform…") are left alone.
const TOKEN_EXPIRY_REASON = /token parse error|token (is )?expired|expired by/i;

function reasonOf(err: { response?: { data?: { reason?: string; items?: { reason?: string }[] } } }): string {
  const data = err?.response?.data;
  return data?.reason ?? data?.items?.[0]?.reason ?? '';
}

function shouldRefresh(status: number | undefined, reason: string): boolean {
  return status === 401 || (status === 403 && TOKEN_EXPIRY_REASON.test(reason));
}

export async function handleResponseError(error: unknown, deps: AuthErrorDeps): Promise<unknown> {
  const err = error as {
    response?: { status?: number; data?: { reason?: string; items?: { reason?: string }[] } };
    config?: { headers?: Record<string, string>; _retry?: boolean };
  };
  const status = err?.response?.status;
  const onAuthPage = deps.isAuthPage ? deps.isAuthPage() : isOnAuthPage();
  const original = err?.config;

  if (shouldRefresh(status, reasonOf(err)) && !onAuthPage && original && !original._retry) {
    original._retry = true;
    let token: string;
    try {
      token = await deps.refresh();
    } catch {
      // Only a failed refresh logs out; a failed retry must propagate as-is.
      deps.onAuthFailure();
      return Promise.reject(error);
    }
    original.headers = original.headers ?? {};
    original.headers['Authorization'] = `Bearer ${token}`;
    return deps.retry(original);
  }
  return Promise.reject(error);
}

api.interceptors.response.use(
  (response) => response,
  (error) =>
    handleResponseError(error, {
      refresh: refreshAccessToken,
      retry: (config) => api(config as Parameters<typeof api>[0]),
      onAuthFailure: () => {
        clearAuthSession();
        window.location.href = '/sign-in';
      },
    }),
);

export default api;
