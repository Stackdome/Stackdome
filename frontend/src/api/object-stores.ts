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

// Single object-store read is project-scoped (only the org-level list is aggregated).
export async function getObjectStore(orgId: string, projectName: string, id: string): Promise<ObjectStore> {
  const res = await api.get(`/organizations/${orgId}/projects/${projectName}/object-stores/${id}`);
  return res.data as ObjectStore;
}

// Writes go through project-scoped endpoints (the org-scoped paths are GET-only).
export async function createObjectStore(
  orgId: string,
  projectName: string,
  payload: ObjectStoreCreatePayload,
): Promise<ObjectStore> {
  const res = await api.post(`/organizations/${orgId}/projects/${projectName}/object-stores`, payload);
  return res.data as ObjectStore;
}

export async function updateObjectStore(
  orgId: string,
  projectName: string,
  id: string,
  payload: ObjectStoreCreatePayload,
): Promise<ObjectStore> {
  const res = await api.put(`/organizations/${orgId}/projects/${projectName}/object-stores/${id}`, payload);
  return res.data as ObjectStore;
}

export async function deleteObjectStore(orgId: string, projectName: string, id: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/projects/${projectName}/object-stores/${id}`);
}
