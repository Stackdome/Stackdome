import api from "./client";
import type { components } from "./types/openapi";

export type Secret = components["schemas"]["Secret"];
export type SecretList = components["schemas"]["SecretList"];
export type SecretType = components["schemas"]["SecretType"];
export type SecretData = components["schemas"]["SecretData"];

export async function getSecrets(orgId: string): Promise<SecretList> {
  const res = await api.get(`/organizations/${orgId}/secrets`);
  return res.data as SecretList;
}

// Single-secret read is team-scoped (only the org-level list is aggregated).
export async function getSecret(orgId: string, teamName: string, secretId: string): Promise<Secret> {
  const res = await api.get(`/organizations/${orgId}/teams/${teamName}/secrets/${secretId}`);
  return res.data as Secret;
}

// Writes go through team-scoped endpoints (the org-scoped paths are GET-only).
export async function createSecret(orgId: string, teamName: string, secret: Omit<Secret, "id" | "organisation_id" | "created_at" | "updated_at">): Promise<Secret> {
  const res = await api.post(`/organizations/${orgId}/teams/${teamName}/secrets`, secret);
  return res.data as Secret;
}

export async function updateSecret(orgId: string, teamName: string, secretId: string, secret: Omit<Secret, "id" | "organisation_id" | "created_at" | "updated_at">): Promise<Secret> {
  const res = await api.put(`/organizations/${orgId}/teams/${teamName}/secrets/${secretId}`, secret);
  return res.data as Secret;
}

export async function deleteSecret(orgId: string, teamName: string, secretId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/teams/${teamName}/secrets/${secretId}`);
}
