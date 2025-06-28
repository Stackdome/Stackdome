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

export async function getSecret(orgId: string, secretId: string): Promise<Secret> {
  const res = await api.get(`/organizations/${orgId}/secrets/${secretId}`);
  return res.data as Secret;
}

export async function createSecret(orgId: string, secret: Omit<Secret, "id" | "organisation_id" | "created_at" | "updated_at">): Promise<Secret> {
  const res = await api.post(`/organizations/${orgId}/secrets`, secret);
  return res.data as Secret;
}

export async function updateSecret(orgId: string, secretId: string, secret: Omit<Secret, "id" | "organisation_id" | "created_at" | "updated_at">): Promise<Secret> {
  const res = await api.put(`/organizations/${orgId}/secrets/${secretId}`, secret);
  return res.data as Secret;
}

export async function deleteSecret(orgId: string, secretId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/secrets/${secretId}`);
}
