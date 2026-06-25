import api from "./client";
import { API_BASE_URL } from "./base-url";

// Logs/metrics (incl. SSE) served from team-scoped stack endpoints; UI scopes to the org's default team.

interface LogStreamParams {
  follow?: boolean;
  since?: string;
  tail?: number;
}

function withLogParams(path: string, params?: LogStreamParams): string {
  const searchParams = new URLSearchParams();

  if (params?.follow !== undefined) {
    searchParams.set('follow', params.follow.toString());
  }

  if (params?.since) {
    searchParams.set('since', params.since);
  }

  if (params?.tail !== undefined) {
    searchParams.set('tail', params.tail.toString());
  }

  const queryString = searchParams.toString();
  return queryString ? `${path}?${queryString}` : path;
}

export function buildStackLogStreamUrl(
  organizationId: string,
  teamName: string,
  stackId: string,
  params?: LogStreamParams
): string {
  const baseUrl = API_BASE_URL;
  const path = `${baseUrl}/organizations/${organizationId}/teams/${teamName}/stacks/${stackId}/logs`;
  return withLogParams(path, params);
}

export function buildStackResourceLogStreamUrl(
  organizationId: string,
  teamName: string,
  stackId: string,
  resourceName: string,
  params?: LogStreamParams
): string {
  const baseUrl = API_BASE_URL;
  const path = `${baseUrl}/organizations/${organizationId}/teams/${teamName}/stacks/${stackId}/resources/${resourceName}/logs`;
  return withLogParams(path, params);
}

export async function getStackLogs(organizationId: string, teamName: string, stackId: string) {
  const res = await api.get(`/organizations/${organizationId}/teams/${teamName}/stacks/${stackId}/logs`);
  return res.data;
}

export async function getStackMetrics(organizationId: string, teamName: string, stackId: string) {
  const res = await api.get(`/organizations/${organizationId}/teams/${teamName}/stacks/${stackId}/metrics`);
  return res.data;
}

export function buildStackMetricsStreamUrl(
  organizationId: string,
  teamName: string,
  stackId: string
): string {
  const baseUrl = API_BASE_URL;
  return `${baseUrl}/organizations/${organizationId}/teams/${teamName}/stacks/${stackId}/metrics?stream=true`;
}

export async function getStackResourceMetrics(
  organizationId: string,
  teamName: string,
  stackId: string,
  resourceId: string
) {
  const res = await api.get(`/organizations/${organizationId}/teams/${teamName}/stacks/${stackId}/resources/${resourceId}/metrics`);
  return res.data;
}

export function buildStackResourceMetricsStreamUrl(
  organizationId: string,
  teamName: string,
  stackId: string,
  resourceName: string
): string {
  const baseUrl = API_BASE_URL;
  return `${baseUrl}/organizations/${organizationId}/teams/${teamName}/stacks/${stackId}/resources/${resourceName}/metrics?stream=true`;
}

/** One-shot read of a resource log endpoint with follow=false — returns plain lines.
 *  Best-effort: returns [] on any error (pod may be unreachable, issue #98). */
export async function fetchLogSnapshot(
  organizationId: string,
  teamName: string,
  stackId: string,
  resourceName: string,
  tail = 50,
): Promise<string[]> {
  const url = buildStackResourceLogStreamUrl(organizationId, teamName, stackId, resourceName, { follow: false, tail });
  try {
    const res = await fetch(url, { credentials: "include" });
    if (!res.ok) return [];
    const text = await res.text();
    return text.split("\n").map((l) => l.replace(/^data:\s?/, "").trim()).filter((l) => l.length > 0);
  } catch { return []; }
}
