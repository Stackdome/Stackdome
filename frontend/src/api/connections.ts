import api from "./client";
import type { components } from "./types/openapi";

export type StackConnection = components["schemas"]["StackConnection"];
export type StackConnectionList = components["schemas"]["StackConnectionList"];
export type ConnectionMapping = components["schemas"]["ConnectionMapping"];
export type TopologyNodeRef = components["schemas"]["TopologyNodeRef"];
export type StackConnectionConfig = components["schemas"]["StackConnectionConfig"];

// Connections are team-scoped, nested under a stack. The UI scopes to the default team.
const base = (orgId: string, teamName: string, stackId: string) =>
  `/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/connections`;

export async function listStackConnections(
  orgId: string,
  teamName: string,
  stackId: string,
): Promise<StackConnectionList> {
  const res = await api.get(base(orgId, teamName, stackId));
  return res.data as StackConnectionList;
}

export async function createStackConnection(
  orgId: string,
  teamName: string,
  stackId: string,
  connection: StackConnection,
): Promise<StackConnection> {
  const res = await api.post(base(orgId, teamName, stackId), connection);
  return res.data as StackConnection;
}

export async function updateStackConnection(
  orgId: string,
  teamName: string,
  stackId: string,
  connectionId: string,
  connection: StackConnection,
): Promise<StackConnection> {
  const res = await api.put(`${base(orgId, teamName, stackId)}/${connectionId}`, connection);
  return res.data as StackConnection;
}

export async function deleteStackConnection(
  orgId: string,
  teamName: string,
  stackId: string,
  connectionId: string,
): Promise<void> {
  await api.delete(`${base(orgId, teamName, stackId)}/${connectionId}`);
}
