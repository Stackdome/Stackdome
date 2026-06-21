import axios from "axios";
import type { components } from "../api/types/openapi";
import { API_BASE_URL } from "./base-url";

export type RefreshTokenResponse = components["schemas"]["RefreshTokenResponse"];

export function githubOAuthUrl(inviteToken?: string): string {
  const base = `${API_BASE_URL}/auth/github`;
  return inviteToken ? `${base}?invite_token=${encodeURIComponent(inviteToken)}` : base;
}

// Bare axios (not the shared `api` instance) so this pre-session call bypasses
// the auth interceptors — same rationale as api/auth-refresh.ts.
export async function completeGitHubOAuth(
  code: string,
  state: string,
): Promise<RefreshTokenResponse> {
  const res = await axios.get(`${API_BASE_URL}/auth/github/callback`, { params: { code, state } });
  return res.data;
}
