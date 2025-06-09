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

export default api;
