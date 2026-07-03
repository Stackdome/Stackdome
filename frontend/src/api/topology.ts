// Read-side client for the stack topology graph endpoint.
import api from "./client";
import type { components } from "./types/openapi";

export type StackTopology = components["schemas"]["StackTopology"];
export type TopologyNode = components["schemas"]["TopologyNode"];
export type TopologyEdge = components["schemas"]["TopologyEdge"];
export type TopologyNodeRef = components["schemas"]["TopologyNodeRef"];

export async function getStackTopology(orgId: string, teamName: string, stackId: string): Promise<StackTopology> {
  const response = await api.get<StackTopology>(
    `/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/topology`,
  );
  return response.data;
}
