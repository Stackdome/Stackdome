import axios from "axios";
import { getRefreshToken } from "@/helpers/common";

// Bare axios (not the shared `api` instance) so refresh bypasses the interceptors.
const REFRESH_URL = `${import.meta.env.VITE_API_BASE_URL || "/api/v1"}/auth/refresh`;

// Single-flight: concurrent callers share one refresh request.
let refreshPromise: Promise<string> | null = null;

async function doRefresh(): Promise<string> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) {
    throw new Error("No refresh token available");
  }

  const res = await axios.post(REFRESH_URL, { refreshToken });
  const token: string | undefined = res.data?.token;
  if (!token) {
    throw new Error("Refresh response missing access token");
  }

  localStorage.setItem("authToken", token);
  const rotated: string | undefined = res.data?.refreshToken;
  if (rotated) {
    localStorage.setItem("refreshToken", rotated);
  }
  return token;
}

export function refreshAccessToken(): Promise<string> {
  if (!refreshPromise) {
    refreshPromise = doRefresh().finally(() => {
      refreshPromise = null;
    });
  }
  return refreshPromise;
}
