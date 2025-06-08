import api from "./client";

export function buildStackLogStreamUrl(
  organizationId: string,
  stackId: string
): string {
  const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
  return `${baseUrl}/organizations/${organizationId}/stacks/${stackId}/logs?follow=true`;
}

export async function getStackLogs(organizationId: string, stackId: string) {
  const res = await api.get(`/organizations/${organizationId}/stacks/${stackId}/logs`);
  return res.data;
}

export async function getStackMetrics(organizationId: string, stackId: string) {
  const res = await api.get(`/organizations/${organizationId}/stacks/${stackId}/metrics`);
  return res.data;
}

export async function getStackResourceMetrics(
  organizationId: string,
  stackId: string,
  resourceId: string
) {
  const res = await api.get(`/organizations/${organizationId}/stacks/${stackId}/resources/${resourceId}/metrics`);
  return res.data;
}
