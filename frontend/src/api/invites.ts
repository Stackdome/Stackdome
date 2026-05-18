import api from "./client";
import type { components } from "./types/openapi";

export type OrgInvite = components["schemas"]["OrgInvite"];
export type OrgInviteList = components["schemas"]["OrgInviteList"];
export type OrgInviteCreateRequest = components["schemas"]["OrgInviteCreateRequest"];
export type OrgInviteCreateResponse = components["schemas"]["OrgInviteCreateResponse"];
export type OrgInviteInfo = components["schemas"]["OrgInviteInfo"];
export type InviteStatus = components["schemas"]["InviteStatus"];

export async function createInvite(orgId: string, body: OrgInviteCreateRequest): Promise<OrgInviteCreateResponse> {
  const res = await api.post(`/organizations/${orgId}/invites`, body);
  return res.data as OrgInviteCreateResponse;
}

export async function listInvites(orgId: string, status?: InviteStatus): Promise<OrgInviteList> {
  const res = await api.get(`/organizations/${orgId}/invites`, {
    params: status ? { status } : undefined,
  });
  return res.data as OrgInviteList;
}

export async function revokeInvite(orgId: string, inviteId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/invites/${inviteId}`);
}

export async function resendInvite(orgId: string, inviteId: string): Promise<void> {
  await api.post(`/organizations/${orgId}/invites/${inviteId}/resend`);
}

export async function getPublicInviteInfo(token: string): Promise<OrgInviteInfo> {
  const res = await api.get(`/invites/${token}/info`);
  return res.data as OrgInviteInfo;
}
