import api from "./client";
import type { Volume, VolumeUpdateRequest } from "./stacks";

/** Thin stack-scoped create: the backend creates AND associates in one tx. */
export async function createStackVolume(orgId: string, projectName: string, stackId: string, body: VolumeUpdateRequest): Promise<Volume> {
  const response = await api.post<Volume>(`/organizations/${orgId}/projects/${projectName}/stacks/${stackId}/volumes`, body);
  return response.data;
}

/** Destroys the cluster volume synchronously — confirm-gated callers only (revert, canvas delete). */
export async function deleteVolume(orgId: string, projectName: string, volumeId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/projects/${projectName}/volumes/${volumeId}`);
}
