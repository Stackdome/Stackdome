import api from "./client";
import type { components } from "./types/openapi";

export type RegistryCredential = components["schemas"]["RegistryCredential"];
export type RegistryCredentialList = components["schemas"]["RegistryCredentialList"];
export type RegistryCredentialPurpose = components["schemas"]["RegistryCredentialPurpose"];
export type RegistryCredentialDeleteResponse = components["schemas"]["RegistryCredentialDeleteResponse"];
export type AffectedStackRef = components["schemas"]["AffectedStackRef"];

function base(orgId: string): string {
  return `/organizations/${orgId}/registry-credentials`;
}

export async function listRegistryCredentials(orgId: string): Promise<RegistryCredentialList> {
  const res = await api.get(base(orgId));
  return res.data as RegistryCredentialList;
}

export async function createRegistryCredential(orgId: string, body: RegistryCredential): Promise<RegistryCredential> {
  const res = await api.post(base(orgId), body);
  return res.data as RegistryCredential;
}

export async function updateRegistryCredential(orgId: string, id: string, body: RegistryCredential): Promise<RegistryCredential> {
  const res = await api.put(`${base(orgId)}/${id}`, body);
  return res.data as RegistryCredential;
}

export async function deleteRegistryCredential(orgId: string, id: string): Promise<RegistryCredentialDeleteResponse> {
  const res = await api.delete(`${base(orgId)}/${id}`);
  return (res.data ?? {}) as RegistryCredentialDeleteResponse;
}

export async function verifyRegistryCredential(
  orgId: string,
  id: string,
  repository: string,
  purpose?: RegistryCredentialPurpose,
): Promise<void> {
  await api.post(`${base(orgId)}/${id}/verify`, purpose ? { repository, purpose } : { repository });
}
