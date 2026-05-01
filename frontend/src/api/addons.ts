import api from "./client";
import type { components } from "./types/openapi";

export type PostgresAddon = components["schemas"]["PostgresAddon"];
export type PostgresAddonList = components["schemas"]["PostgresAddonList"];
export type PostgresAddonSpec = components["schemas"]["PostgresAddonSpec"];
export type PostgresAddonStatus = components["schemas"]["PostgresAddonStatus"];
export type PostgresAddonState = NonNullable<PostgresAddonStatus["state"]>;

export type PostgresAddonCreateInput = Pick<PostgresAddon, "name" | "spec"> &
  Partial<Pick<PostgresAddon, "labels" | "annotations" | "cluster_id">>;

export async function listPostgresAddons(orgId: string): Promise<PostgresAddonList> {
  const res = await api.get(`/organizations/${orgId}/addons/postgres`);
  return res.data as PostgresAddonList;
}

export async function getPostgresAddon(orgId: string, id: string): Promise<PostgresAddon> {
  const res = await api.get(`/organizations/${orgId}/addons/postgres/${id}`);
  return res.data as PostgresAddon;
}

export async function createPostgresAddon(
  orgId: string,
  input: PostgresAddonCreateInput,
): Promise<PostgresAddon> {
  const res = await api.post(`/organizations/${orgId}/addons/postgres`, input);
  return res.data as PostgresAddon;
}

export async function updatePostgresAddon(
  orgId: string,
  id: string,
  input: PostgresAddonCreateInput,
): Promise<PostgresAddon> {
  const res = await api.put(`/organizations/${orgId}/addons/postgres/${id}`, input);
  return res.data as PostgresAddon;
}

export async function deletePostgresAddon(orgId: string, id: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/addons/postgres/${id}`);
}
