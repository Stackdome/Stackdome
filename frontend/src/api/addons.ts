import api from "./client";
import type { components } from "./types/openapi";

export type PostgresAddon = components["schemas"]["PostgresAddon"];
export type PostgresAddonList = components["schemas"]["PostgresAddonList"];
export type PostgresAddonSpec = components["schemas"]["PostgresAddonSpec"];
export type PostgresAddonStatus = components["schemas"]["PostgresAddonStatus"];
export type PostgresAddonState = NonNullable<PostgresAddonStatus["state"]>;

export type PostgresAddonCreateInput = Pick<PostgresAddon, "name" | "spec"> &
  Partial<Pick<PostgresAddon, "labels" | "annotations" | "cluster_id">>;

// Addons are team-scoped (every addon belongs to a team); reads and writes both
// go through the team-scoped endpoints. The UI scopes everything to the default team.
export async function listPostgresAddons(orgId: string, teamName: string): Promise<PostgresAddonList> {
  const res = await api.get(`/organizations/${orgId}/teams/${teamName}/addons/postgres`);
  return res.data as PostgresAddonList;
}

export async function getPostgresAddon(orgId: string, teamName: string, id: string): Promise<PostgresAddon> {
  const res = await api.get(`/organizations/${orgId}/teams/${teamName}/addons/postgres/${id}`);
  return res.data as PostgresAddon;
}

export async function createPostgresAddon(
  orgId: string,
  teamName: string,
  input: PostgresAddonCreateInput,
): Promise<PostgresAddon> {
  const res = await api.post(`/organizations/${orgId}/teams/${teamName}/addons/postgres`, input);
  return res.data as PostgresAddon;
}

export async function updatePostgresAddon(
  orgId: string,
  teamName: string,
  id: string,
  input: PostgresAddonCreateInput,
): Promise<PostgresAddon> {
  const res = await api.put(`/organizations/${orgId}/teams/${teamName}/addons/postgres/${id}`, input);
  return res.data as PostgresAddon;
}

export async function deletePostgresAddon(orgId: string, teamName: string, id: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/teams/${teamName}/addons/postgres/${id}`);
}
