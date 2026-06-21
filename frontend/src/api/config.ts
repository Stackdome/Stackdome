import api from "./client";
import type { components } from "../api/types/openapi";

export type AppConfigResponse = components["schemas"]["AppConfigResponse"];

// Single-flight + value cache: the app config is immutable for the session, so
// fetch it at most once. Warmed eagerly at app entry (see main.tsx) so the
// "Continue with GitHub" gate is resolved before the auth screens paint, even
// on slow networks.
let configPromise: Promise<AppConfigResponse> | null = null;
let configValue: AppConfigResponse | null = null;

export function getAppConfig(): Promise<AppConfigResponse> {
  if (!configPromise) {
    configPromise = api.get("/config").then((res) => {
      configValue = res.data;
      return res.data;
    });
  }
  return configPromise;
}

// Synchronously returns the cached config if already resolved, else null.
// Lets consumers render the correct state on first paint after a warm cache.
export function getCachedAppConfig(): AppConfigResponse | null {
  return configValue;
}
