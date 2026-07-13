import api from "./client";
import type { components } from "./types/openapi";

export type Project = components["schemas"]["Project"];
export type ProjectList = components["schemas"]["ProjectList"];
export type ProjectCreateRequest = components["schemas"]["ProjectCreateRequest"];
export type ProjectUpdateRequest = components["schemas"]["ProjectUpdateRequest"];
export type ProjectMembership = components["schemas"]["ProjectMembership"];
export type ProjectMembershipList = components["schemas"]["ProjectMembershipList"];
export type AddProjectMemberRequest = components["schemas"]["AddProjectMemberRequest"];
export type UpdateProjectMemberRoleRequest = components["schemas"]["UpdateProjectMemberRoleRequest"];
export type ProjectRoleList = components["schemas"]["ProjectRoleList"];

export async function listProjects(orgId: string): Promise<ProjectList> {
  const res = await api.get(`/organizations/${orgId}/projects`);
  return res.data as ProjectList;
}

export async function getProject(orgId: string, projectName: string): Promise<Project> {
  const res = await api.get(`/organizations/${orgId}/projects/${projectName}`);
  return res.data as Project;
}

export async function createProject(orgId: string, body: ProjectCreateRequest): Promise<Project> {
  const res = await api.post(`/organizations/${orgId}/projects`, body);
  return res.data as Project;
}

export async function renameProject(orgId: string, projectName: string, body: ProjectUpdateRequest): Promise<Project> {
  const res = await api.put(`/organizations/${orgId}/projects/${projectName}`, body);
  return res.data as Project;
}

export async function deleteProject(orgId: string, projectName: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/projects/${projectName}`);
}

export async function listProjectMembers(orgId: string, projectName: string): Promise<ProjectMembershipList> {
  const res = await api.get(`/organizations/${orgId}/projects/${projectName}/members`);
  return res.data as ProjectMembershipList;
}

export async function addProjectMember(orgId: string, projectName: string, body: AddProjectMemberRequest): Promise<ProjectMembership> {
  const res = await api.post(`/organizations/${orgId}/projects/${projectName}/members`, body);
  return res.data as ProjectMembership;
}

export async function updateProjectMemberRole(orgId: string, projectName: string, membershipId: string, body: UpdateProjectMemberRoleRequest): Promise<ProjectMembership> {
  const res = await api.put(`/organizations/${orgId}/projects/${projectName}/members/${membershipId}`, body);
  return res.data as ProjectMembership;
}

export async function removeProjectMember(orgId: string, projectName: string, membershipId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/projects/${projectName}/members/${membershipId}`);
}

export async function listProjectRoles(): Promise<ProjectRoleList> {
  const res = await api.get(`/project-roles`);
  return res.data as ProjectRoleList;
}
