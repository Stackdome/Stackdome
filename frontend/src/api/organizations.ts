import api from "./client";
import type { components } from "../api/types/openapi";

export type Organization = components["schemas"]["Organisation"];
export type DomainName = components["schemas"]["DomainName"];

export async function getOrganization(orgId: string): Promise<Organization> {
  const res = await api.get(`/organizations/${orgId}`);
  return res.data as Organization;
}

export async function updateOrganization(orgId: string, input: Organization): Promise<Organization> {
  const res = await api.put(`/organizations/${orgId}`, input);
  return res.data as Organization;
}
