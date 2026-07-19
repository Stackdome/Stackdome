import api from "./client";
import { API_BASE_URL } from "./base-url";
import type { components } from "./types/openapi";

export type StackRelease = components["schemas"]["StackRelease"];
export type StackReleaseDetail = components["schemas"]["StackReleaseDetail"];
export type StackReleaseSnapshot = components["schemas"]["StackReleaseSnapshot"];
export type StackReleaseList = components["schemas"]["StackReleaseList"];
export type CreateReleaseRequest = components["schemas"]["CreateReleaseRequest"];
export type ReleaseLiveStatus = components["schemas"]["ReleaseLiveStatus"];
export type ReleaseEvent = components["schemas"]["ReleaseEvent"];
export type ReleaseEventList = components["schemas"]["ReleaseEventList"];
export type ReleaseSummary = components["schemas"]["ReleaseSummary"];

/** The generated `ReleaseEvent.scope` is a types-only union; this is its runtime mirror. */
export type ReleaseEventScopeValue = NonNullable<ReleaseEvent["scope"]>;
export const ReleaseEventScope = {
  Release: "release",
  Resource: "resource",
} as const satisfies Record<string, ReleaseEventScopeValue>;

/**
 * The spec types `ReleaseEvent.type` as a plain string; these values mirror the
 * backend enum in pkg/models/release_event.go.
 */
export const ReleaseEventType = {
  ResourceWaiting: "resource_waiting",
  ResourceDeploying: "resource_deploying",
  ResourceReady: "resource_ready",
  ResourceFailed: "resource_failed",
} as const;

function releasesPath(orgId: string, projectName: string, stackId: string): string {
  return `/organizations/${orgId}/projects/${projectName}/stacks/${stackId}/releases`;
}

export async function listReleases(orgId: string, projectName: string, stackId: string): Promise<StackReleaseList> {
  const response = await api.get<StackReleaseList>(releasesPath(orgId, projectName, stackId));
  return response.data;
}

export async function getRelease(orgId: string, projectName: string, stackId: string, releaseId: string): Promise<StackReleaseDetail> {
  const response = await api.get<StackReleaseDetail>(`${releasesPath(orgId, projectName, stackId)}/${releaseId}`);
  return response.data;
}

// POST with no from_release_id triggers a fresh deploy of the current saved config.
export async function createRelease(orgId: string, projectName: string, stackId: string): Promise<StackRelease> {
  const response = await api.post<StackRelease>(releasesPath(orgId, projectName, stackId), {});
  return response.data;
}

// POST with from_release_id re-deploys that release's snapshot+pins (rollback).
export async function rollbackRelease(orgId: string, projectName: string, stackId: string, fromReleaseId: string): Promise<StackRelease> {
  const body: CreateReleaseRequest = { from_release_id: fromReleaseId };
  const response = await api.post<StackRelease>(releasesPath(orgId, projectName, stackId), body);
  return response.data;
}

export async function cancelRelease(orgId: string, projectName: string, stackId: string, releaseId: string): Promise<void> {
  await api.post<void>(`${releasesPath(orgId, projectName, stackId)}/${releaseId}/cancel`);
}

export async function listReleaseEvents(
  orgId: string, projectName: string, stackId: string, releaseId: string, afterSequence?: number,
): Promise<ReleaseEventList> {
  const params = afterSequence !== undefined ? { after_sequence: afterSequence } : undefined;
  const response = await api.get<ReleaseEventList>(
    `${releasesPath(orgId, projectName, stackId)}/${releaseId}/events`, { params },
  );
  return response.data;
}

// EventSource cannot set headers; base-URL handling mirrors buildStackLogStreamUrl in api/observability.ts.
export function buildReleaseEventStreamUrl(
  orgId: string, projectName: string, stackId: string, releaseId: string, afterSequence?: number,
): string {
  const baseUrl = API_BASE_URL;
  const path = `${baseUrl}${releasesPath(orgId, projectName, stackId)}/${releaseId}/events/stream`;
  const suffix = afterSequence !== undefined ? `?after_sequence=${afterSequence}` : "";
  return `${path}${suffix}`;
}
