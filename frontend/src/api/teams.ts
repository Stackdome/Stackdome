import api from "./client";
import type { components } from "./types/openapi";

export type Team = components["schemas"]["Team"];
export type TeamList = components["schemas"]["TeamList"];
export type TeamCreateRequest = components["schemas"]["TeamCreateRequest"];
export type TeamUpdateRequest = components["schemas"]["TeamUpdateRequest"];
export type TeamMembership = components["schemas"]["TeamMembership"];
export type TeamMembershipList = components["schemas"]["TeamMembershipList"];
export type AddTeamMemberRequest = components["schemas"]["AddTeamMemberRequest"];
export type UpdateTeamMemberRoleRequest = components["schemas"]["UpdateTeamMemberRoleRequest"];
export type TeamRoleList = components["schemas"]["TeamRoleList"];

export async function listTeams(orgId: string): Promise<TeamList> {
  const res = await api.get(`/organizations/${orgId}/teams`);
  return res.data as TeamList;
}

export async function getTeam(orgId: string, teamName: string): Promise<Team> {
  const res = await api.get(`/organizations/${orgId}/teams/${teamName}`);
  return res.data as Team;
}

export async function createTeam(orgId: string, body: TeamCreateRequest): Promise<Team> {
  const res = await api.post(`/organizations/${orgId}/teams`, body);
  return res.data as Team;
}

export async function renameTeam(orgId: string, teamName: string, body: TeamUpdateRequest): Promise<Team> {
  const res = await api.put(`/organizations/${orgId}/teams/${teamName}`, body);
  return res.data as Team;
}

export async function deleteTeam(orgId: string, teamName: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/teams/${teamName}`);
}

export async function listTeamMembers(orgId: string, teamName: string): Promise<TeamMembershipList> {
  const res = await api.get(`/organizations/${orgId}/teams/${teamName}/members`);
  return res.data as TeamMembershipList;
}

export async function addTeamMember(orgId: string, teamName: string, body: AddTeamMemberRequest): Promise<TeamMembership> {
  const res = await api.post(`/organizations/${orgId}/teams/${teamName}/members`, body);
  return res.data as TeamMembership;
}

export async function updateTeamMemberRole(orgId: string, teamName: string, membershipId: string, body: UpdateTeamMemberRoleRequest): Promise<TeamMembership> {
  const res = await api.put(`/organizations/${orgId}/teams/${teamName}/members/${membershipId}`, body);
  return res.data as TeamMembership;
}

export async function removeTeamMember(orgId: string, teamName: string, membershipId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/teams/${teamName}/members/${membershipId}`);
}

export async function listTeamRoles(): Promise<TeamRoleList> {
  const res = await api.get(`/team-roles`);
  return res.data as TeamRoleList;
}
