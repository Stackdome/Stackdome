// Single source of truth for the API base path. Defaults to the Vite dev-proxy
// prefix; VITE_API_BASE_URL overrides it (e.g. a non-proxied backend origin).
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api/v1";
