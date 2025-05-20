import api from "./client";
import type { components } from "../api/types/openapi";

export type Cluster = components["schemas"]["Cluster"];
export type ClusterList = components["schemas"]["ClusterList"];

export async function getClusters(orgId: string): Promise<ClusterList> {
  const res = await api.get(`/organizations/${orgId}/clusters`);
  return res.data as ClusterList;
}

export async function createCluster(orgId: string, input: Omit<Cluster, "id">): Promise<Cluster> {
  const res = await api.post(`/organizations/${orgId}/clusters`, input);
  return res.data as Cluster;
}

export async function deleteCluster(orgId: string, clusterId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/clusters/${clusterId}`);
}

export async function getCluster(orgId: string, clusterId: string): Promise<Cluster> {
  const res = await api.get(`/organizations/${orgId}/clusters/${clusterId}`);
  return res.data as Cluster;
}


