import api from "./client";
import type { components } from "./types/openapi";

export type APIToken = components["schemas"]["APIToken"];
export type APITokenList = components["schemas"]["APITokenList"];
export type APITokenCreateRequest = components["schemas"]["APITokenCreateRequest"];
export type APITokenCreateResponse = components["schemas"]["APITokenCreateResponse"];
export type ScopeList = components["schemas"]["ScopeList"];

const BASE = "/api-tokens";

export async function listApiTokens(): Promise<APITokenList> {
  const res = await api.get(BASE);
  return res.data as APITokenList;
}

export async function createApiToken(req: APITokenCreateRequest): Promise<APITokenCreateResponse> {
  const res = await api.post(BASE, req);
  return res.data as APITokenCreateResponse;
}

export async function revokeApiToken(id: string): Promise<void> {
  await api.delete(`${BASE}/${id}`);
}

export async function getApiTokenScopes(): Promise<ScopeList> {
  const res = await api.get(`${BASE}/scopes`);
  return res.data as ScopeList;
}
