// Type re-exports for the connections schema — consumed by connection-mapping.ts.
// The CRUD client (listStackConnections, createStackConnection, etc.) was removed
// after the save path moved to a full connection set in the stack PUT spec.
import api from "./client";
import type { components } from "./types/openapi";

export type StackConnection = components["schemas"]["StackConnection"];
export type StackConnectionList = components["schemas"]["StackConnectionList"];
export type ConnectionMapping = components["schemas"]["ConnectionMapping"];
export type TopologyNodeRef = components["schemas"]["TopologyNodeRef"];
export type StackConnectionConfig = components["schemas"]["StackConnectionConfig"];

function connectionsPath(orgId: string, teamName: string, stackId: string): string {
  return `/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/connections`;
}

export async function createStackConnection(orgId: string, teamName: string, stackId: string, body: StackConnection): Promise<StackConnection> {
  const response = await api.post<StackConnection>(connectionsPath(orgId, teamName, stackId), body);
  return response.data;
}

export async function updateStackConnection(orgId: string, teamName: string, stackId: string, connectionId: string, body: StackConnection): Promise<StackConnection> {
  const response = await api.put<StackConnection>(`${connectionsPath(orgId, teamName, stackId)}/${connectionId}`, body);
  return response.data;
}

export async function deleteStackConnection(orgId: string, teamName: string, stackId: string, connectionId: string): Promise<void> {
  await api.delete(`${connectionsPath(orgId, teamName, stackId)}/${connectionId}`);
}
