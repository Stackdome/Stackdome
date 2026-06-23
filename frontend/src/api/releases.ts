import api from "./client";
import type { components } from "./types/openapi";

export type StackRelease = components["schemas"]["StackRelease"];
export type StackReleaseDetail = components["schemas"]["StackReleaseDetail"];
export type StackReleaseSnapshot = components["schemas"]["StackReleaseSnapshot"];
export type StackReleaseList = components["schemas"]["StackReleaseList"];
export type CreateReleaseRequest = components["schemas"]["CreateReleaseRequest"];

function releasesPath(orgId: string, teamName: string, stackId: string): string {
  return `/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/releases`;
}

export async function listReleases(orgId: string, teamName: string, stackId: string): Promise<StackReleaseList> {
  const response = await api.get<StackReleaseList>(releasesPath(orgId, teamName, stackId));
  return response.data;
}

export async function getRelease(orgId: string, teamName: string, stackId: string, releaseId: string): Promise<StackReleaseDetail> {
  const response = await api.get<StackReleaseDetail>(`${releasesPath(orgId, teamName, stackId)}/${releaseId}`);
  return response.data;
}

// POST with no from_release_id triggers a fresh deploy of the current saved config.
export async function createRelease(orgId: string, teamName: string, stackId: string): Promise<StackRelease> {
  const response = await api.post<StackRelease>(releasesPath(orgId, teamName, stackId), {});
  return response.data;
}

// POST with from_release_id re-deploys that release's snapshot+pins (rollback).
export async function rollbackRelease(orgId: string, teamName: string, stackId: string, fromReleaseId: string): Promise<StackRelease> {
  const body: CreateReleaseRequest = { from_release_id: fromReleaseId };
  const response = await api.post<StackRelease>(releasesPath(orgId, teamName, stackId), body);
  return response.data;
}

export async function cancelRelease(orgId: string, teamName: string, stackId: string, releaseId: string): Promise<void> {
  await api.post<void>(`${releasesPath(orgId, teamName, stackId)}/${releaseId}/cancel`);
}
