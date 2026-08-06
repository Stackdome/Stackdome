import api from "./client";
import type { components } from "./types/openapi";

export type APIToken = components["schemas"]["APIToken"];
export type APITokenList = components["schemas"]["APITokenList"];
export type APITokenCreateRequest = components["schemas"]["APITokenCreateRequest"];
export type APITokenCreateResponse = components["schemas"]["APITokenCreateResponse"];
export type ScopeList = components["schemas"]["ScopeList"];

export async function listApiTokens(): Promise<APITokenList> {
  const res = await api.get(`/api-tokens`);
  return res.data as APITokenList;
}

export async function createApiToken(req: APITokenCreateRequest): Promise<APITokenCreateResponse> {
  const res = await api.post(`/api-tokens`, req);
  return res.data as APITokenCreateResponse;
}

export async function revokeApiToken(id: string): Promise<void> {
  await api.delete(`/api-tokens/${id}`);
}

export async function getApiTokenScopes(): Promise<ScopeList> {
  const res = await api.get(`/api-tokens/scopes`);
  return res.data as ScopeList;
}
