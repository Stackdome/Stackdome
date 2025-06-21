import api from "./client";

interface LogStreamParams {
  follow?: boolean;
  since?: string;
  tail?: number;
}

export function buildStackLogStreamUrl(
  organizationId: string,
  stackId: string,
  params?: LogStreamParams
): string {
  const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
  const path = `${baseUrl}/organizations/${organizationId}/stacks/${stackId}/logs`;

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

export function buildStackResourceLogStreamUrl(
  organizationId: string,
  stackId: string,
  resourceName: string,
  params?: LogStreamParams
): string {
  const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
  const path = `${baseUrl}/organizations/${organizationId}/stacks/${stackId}/resources/${resourceName}/logs`;

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

export async function getStackLogs(organizationId: string, stackId: string) {
  const res = await api.get(`/organizations/${organizationId}/stacks/${stackId}/logs`);
  return res.data;
}

export async function getStackMetrics(organizationId: string, stackId: string) {
  const res = await api.get(`/organizations/${organizationId}/stacks/${stackId}/metrics`);
  return res.data;
}

export function buildStackMetricsStreamUrl(
  organizationId: string,
  stackId: string
): string {
  const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
  return `${baseUrl}/organizations/${organizationId}/stacks/${stackId}/metrics?stream=true`;
}

export async function getStackResourceMetrics(
  organizationId: string,
  stackId: string,
  resourceId: string
) {
  const res = await api.get(`/organizations/${organizationId}/stacks/${stackId}/resources/${resourceId}/metrics`);
  return res.data;
}

export function buildStackResourceMetricsStreamUrl(
  organizationId: string,
  stackId: string,
  resourceName: string
): string {
  const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
  return `${baseUrl}/organizations/${organizationId}/stacks/${stackId}/resources/${resourceName}/metrics?stream=true`;
}
