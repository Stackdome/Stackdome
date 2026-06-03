import api from "./client";
import type { components } from "./types/openapi";

export type ObjectStore = components["schemas"]["ObjectStore"];
export type ObjectStoreList = components["schemas"]["ObjectStoreList"];
export type ObjectStoreSpec = components["schemas"]["ObjectStoreSpec"];
export type ObjectStoreConfiguration = components["schemas"]["ObjectStoreConfiguration"];
export type S3Credentials = components["schemas"]["S3Credentials"];
export type AzureCredentials = components["schemas"]["AzureCredentials"];
export type GCSCredentials = components["schemas"]["GCSCredentials"];
export type SecretReference = components["schemas"]["SecretReference"];

export type ObjectStoreCreatePayload = Pick<ObjectStore, "name" | "spec">;

export async function getObjectStores(orgId: string): Promise<ObjectStoreList> {
  const res = await api.get(`/organizations/${orgId}/object-stores`);
  return res.data as ObjectStoreList;
}

export async function getObjectStore(orgId: string, id: string): Promise<ObjectStore> {
  const res = await api.get(`/organizations/${orgId}/object-stores/${id}`);
  return res.data as ObjectStore;
}

// Writes go through team-scoped endpoints (the org-scoped paths are GET-only).
export async function createObjectStore(
  orgId: string,
  teamName: string,
  payload: ObjectStoreCreatePayload,
): Promise<ObjectStore> {
  const res = await api.post(`/organizations/${orgId}/teams/${teamName}/object-stores`, payload);
  return res.data as ObjectStore;
}

export async function updateObjectStore(
  orgId: string,
  teamName: string,
  id: string,
  payload: ObjectStoreCreatePayload,
): Promise<ObjectStore> {
  const res = await api.put(`/organizations/${orgId}/teams/${teamName}/object-stores/${id}`, payload);
  return res.data as ObjectStore;
}

export async function deleteObjectStore(orgId: string, teamName: string, id: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/teams/${teamName}/object-stores/${id}`);
}
