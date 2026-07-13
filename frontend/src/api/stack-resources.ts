import api from "./client";
import type { StackResource, StackResourceUpdateRequest } from "./stacks";

function resourcesPath(orgId: string, projectName: string, stackId: string): string {
  return `/organizations/${orgId}/projects/${projectName}/stacks/${stackId}/resources`;
}

export async function createStackResource(orgId: string, projectName: string, stackId: string, body: StackResourceUpdateRequest): Promise<StackResource> {
  const response = await api.post<StackResource>(resourcesPath(orgId, projectName, stackId), body);
  return response.data;
}

export async function updateStackResource(orgId: string, projectName: string, stackId: string, resourceName: string, body: StackResourceUpdateRequest): Promise<StackResource> {
  const response = await api.put<StackResource>(`${resourcesPath(orgId, projectName, stackId)}/${encodeURIComponent(resourceName)}`, body);
  return response.data;
}

export async function deleteStackResource(orgId: string, projectName: string, stackId: string, resourceName: string): Promise<void> {
  await api.delete(`${resourcesPath(orgId, projectName, stackId)}/${encodeURIComponent(resourceName)}`);
}
