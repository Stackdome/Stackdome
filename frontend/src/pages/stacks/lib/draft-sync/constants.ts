/** Timing + status constants for the draft autosave engine. */
export const DEBOUNCE_IDLE_MS = 1200;
export const DEBOUNCE_MAX_WAIT_MS = 5000;
export const RETRY_BASE_MS = 1000;
export const RETRY_MAX_MS = 30000;
export const STICKY_FAILURE_THRESHOLD = 3;

export const SYNC_STATUS = {
  idle: "idle",
  saving: "saving",
  saved: "saved",
  error: "error",
} as const;
export type SyncStatus = (typeof SYNC_STATUS)[keyof typeof SYNC_STATUS];
