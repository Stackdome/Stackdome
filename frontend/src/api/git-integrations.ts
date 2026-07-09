import api from "./client";
import type { components } from "./types/openapi";

export type GitIntegration = components["schemas"]["GitIntegration"];
export type GitIntegrationList = components["schemas"]["GitIntegrationList"];
export type GitHubAppManifestFlow = components["schemas"]["GitHubAppManifestFlow"];
export type GitInstallation = components["schemas"]["GitInstallation"];
export type GitInstallationList = components["schemas"]["GitInstallationList"];
export type GitRepository = components["schemas"]["GitRepository"];
export type GitRepositoryPage = components["schemas"]["GitRepositoryPage"];
export type GitBranchList = components["schemas"]["GitBranchList"];

function base(orgId: string): string {
  return `/organizations/${orgId}/git-integrations`;
}

export async function listGitIntegrations(orgId: string): Promise<GitIntegrationList> {
  const res = await api.get(base(orgId));
  return res.data as GitIntegrationList;
}

export async function getGitIntegration(orgId: string, id: string): Promise<GitIntegration> {
  const res = await api.get(`${base(orgId)}/${id}`);
  return res.data as GitIntegration;
}

export async function deleteGitIntegration(orgId: string, id: string): Promise<void> {
  await api.delete(`${base(orgId)}/${id}`);
}

export async function verifyGitIntegration(orgId: string, id: string, repoUrl: string): Promise<void> {
  await api.post(`${base(orgId)}/${id}/verify`, { repo_url: repoUrl });
}

export async function createGitHubAppManifest(orgId: string): Promise<GitHubAppManifestFlow> {
  const res = await api.post(`${base(orgId)}/github/manifest`);
  return res.data as GitHubAppManifestFlow;
}

export async function listInstallations(orgId: string, integrationId: string, refresh = false): Promise<GitInstallationList> {
  const res = await api.get(`${base(orgId)}/${integrationId}/installations`, { params: { refresh } });
  return res.data as GitInstallationList;
}

export interface RepoSearchOpts {
  query?: string;
  page?: number;
  installationId?: number;
}

export async function searchRepositories(orgId: string, integrationId: string, opts: RepoSearchOpts = {}): Promise<GitRepositoryPage> {
  const params: Record<string, string | number> = {};
  if (opts.query) params.query = opts.query;
  if (opts.page) params.page = opts.page;
  if (opts.installationId) params.installation_id = opts.installationId;
  const res = await api.get(`${base(orgId)}/${integrationId}/repositories`, { params });
  return res.data as GitRepositoryPage;
}

export async function getRepository(orgId: string, integrationId: string, owner: string, repo: string): Promise<GitRepository> {
  const res = await api.get(`${base(orgId)}/${integrationId}/repositories/${owner}/${repo}`);
  return res.data as GitRepository;
}

export async function listRepositoryBranches(orgId: string, integrationId: string, owner: string, repo: string): Promise<GitBranchList> {
  const res = await api.get(`${base(orgId)}/${integrationId}/repositories/${owner}/${repo}/branches`);
  return res.data as GitBranchList;
}
