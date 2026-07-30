import api from "./client";
import type { components } from "./types/openapi";

export type PostgresAddon = components["schemas"]["PostgresAddon"];
export type PostgresAddonList = components["schemas"]["PostgresAddonList"];
export type PostgresAddonSpec = components["schemas"]["PostgresAddonSpec"];
export type PostgresAddonStatus = components["schemas"]["PostgresAddonStatus"];
export type PostgresAddonState = NonNullable<PostgresAddonStatus["state"]>;
export type PostgresConnectionInfo = components["schemas"]["PostgresConnectionInfo"];
export type PostgresCredentials = components["schemas"]["PostgresCredentials"];

export type PostgresAddonCreateInput = Pick<PostgresAddon, "name" | "spec"> &
  Partial<Pick<PostgresAddon, "labels" | "annotations" | "cluster_id">>;

// Addons are project-scoped (every addon belongs to a project); reads and writes both
// go through the project-scoped endpoints. The UI scopes everything to the default project.
export async function listPostgresAddons(orgId: string, projectName: string): Promise<PostgresAddonList> {
  const res = await api.get(`/organizations/${orgId}/projects/${projectName}/addons/postgres`);
  return res.data as PostgresAddonList;
}

export async function getPostgresAddon(orgId: string, projectName: string, id: string): Promise<PostgresAddon> {
  const res = await api.get(`/organizations/${orgId}/projects/${projectName}/addons/postgres/${id}`);
  return res.data as PostgresAddon;
}

export async function createPostgresAddon(
  orgId: string,
  projectName: string,
  input: PostgresAddonCreateInput,
): Promise<PostgresAddon> {
  const res = await api.post(`/organizations/${orgId}/projects/${projectName}/addons/postgres`, input);
  return res.data as PostgresAddon;
}

export async function updatePostgresAddon(
  orgId: string,
  projectName: string,
  id: string,
  input: PostgresAddonCreateInput,
): Promise<PostgresAddon> {
  const res = await api.put(`/organizations/${orgId}/projects/${projectName}/addons/postgres/${id}`, input);
  return res.data as PostgresAddon;
}

export async function deletePostgresAddon(orgId: string, projectName: string, id: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/projects/${projectName}/addons/postgres/${id}`);
}

// Credentials are minted on demand from the cluster secret — never persisted by the hub,
// so every reveal is a fresh read.
export async function getPostgresCredentials(
  orgId: string,
  projectName: string,
  id: string,
  database: string,
  superuser: boolean,
): Promise<PostgresCredentials> {
  const res = await api.get(
    `/organizations/${orgId}/projects/${projectName}/addons/postgres/${id}/credentials/${encodeURIComponent(database)}`,
    { params: { superuser } },
  );
  return res.data as PostgresCredentials;
}
