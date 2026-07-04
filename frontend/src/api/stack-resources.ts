import api from "./client";
import type { StackResource, StackResourceUpdateRequest } from "./stacks";

function resourcesPath(orgId: string, teamName: string, stackId: string): string {
  return `/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/resources`;
}

export async function createStackResource(orgId: string, teamName: string, stackId: string, body: StackResourceUpdateRequest): Promise<StackResource> {
  const response = await api.post<StackResource>(resourcesPath(orgId, teamName, stackId), body);
  return response.data;
}

export async function updateStackResource(orgId: string, teamName: string, stackId: string, resourceName: string, body: StackResourceUpdateRequest): Promise<StackResource> {
  const response = await api.put<StackResource>(`${resourcesPath(orgId, teamName, stackId)}/${encodeURIComponent(resourceName)}`, body);
  return response.data;
}

export async function deleteStackResource(orgId: string, teamName: string, stackId: string, resourceName: string): Promise<void> {
  await api.delete(`${resourcesPath(orgId, teamName, stackId)}/${encodeURIComponent(resourceName)}`);
}
