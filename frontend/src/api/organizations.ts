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

export type User = components["schemas"]["User"];
export type UserList = components["schemas"]["UserList"];
export type PromoteAdminRequest = components["schemas"]["PromoteAdminRequest"];
export type DemoteAdminRequest = components["schemas"]["DemoteAdminRequest"];

export async function listOrganizationUsers(orgId: string, page = 1, pageSize = 50): Promise<UserList> {
  const res = await api.get(`/organizations/${orgId}/users`, { params: { page, page_size: pageSize } });
  return res.data as UserList;
}

export async function promoteAdmin(orgId: string, body: PromoteAdminRequest): Promise<void> {
  await api.post(`/organizations/${orgId}/admins`, body);
}

export async function demoteAdmin(orgId: string, userId: string, body: DemoteAdminRequest): Promise<void> {
  await api.post(`/organizations/${orgId}/admins/${userId}/demote`, body);
}
