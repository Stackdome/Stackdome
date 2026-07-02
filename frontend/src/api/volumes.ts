import api from "./client";
import type { Volume, VolumeUpdateRequest } from "./stacks";

/** Thin stack-scoped create: the backend creates AND associates in one tx. */
export async function createStackVolume(orgId: string, teamName: string, stackId: string, body: VolumeUpdateRequest): Promise<Volume> {
  const response = await api.post<Volume>(`/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/volumes`, body);
  return response.data;
}

/** Destroys the cluster volume synchronously — confirm-gated callers only (revert). */
export async function deleteVolume(orgId: string, teamName: string, volumeId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/teams/${teamName}/volumes/${volumeId}`);
}
