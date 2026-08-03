/**
 * Fired on the window whenever the auth session changes (login, signup, logout).
 * The current-user context listens for it so it can re-hydrate without a full
 * page reload — client-side navigation after auth does not remount the provider.
 */
export const AUTH_SESSION_CHANGED = "auth-session-changed";
