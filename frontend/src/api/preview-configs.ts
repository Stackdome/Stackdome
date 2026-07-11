import api from "./client";
import type { components } from "./types/openapi";

export type StackPreviewConfig = components["schemas"]["StackPreviewConfig"];
export type StackPreviewConfigCreate = components["schemas"]["StackPreviewConfigCreate"];
export type StackPreviewConfigUpdate = components["schemas"]["StackPreviewConfigUpdate"];
export type StackPreviewConfigList = components["schemas"]["StackPreviewConfigList"];
export type PreviewGitRepository = components["schemas"]["PreviewGitRepository"];

function base(orgId: string, teamName: string): string {
  return `/organizations/${orgId}/teams/${teamName}/stack-preview-configs`;
}

export async function listPreviewConfigs(orgId: string, teamName: string, page = 1, pageSize = 20): Promise<StackPreviewConfigList> {
  const res = await api.get(base(orgId, teamName), { params: { page, page_size: pageSize } });
  return res.data as StackPreviewConfigList;
}

/** Fetches every page so callers see the complete config set, not the first 20. */
export async function listAllPreviewConfigs(orgId: string, teamName: string): Promise<StackPreviewConfig[]> {
  const pageSize = 100;
  const items: StackPreviewConfig[] = [];
  for (let page = 1; ; page++) {
    const res = await listPreviewConfigs(orgId, teamName, page, pageSize);
    const batch = res.items ?? [];
    items.push(...batch);
    const total = res.total ?? items.length;
    if (batch.length === 0 || items.length >= total) break;
  }
  return items;
}

export async function getPreviewConfig(orgId: string, teamName: string, id: string): Promise<StackPreviewConfig> {
  const res = await api.get(`${base(orgId, teamName)}/${id}`);
  return res.data as StackPreviewConfig;
}

export async function createPreviewConfig(orgId: string, teamName: string, input: StackPreviewConfigCreate): Promise<StackPreviewConfig> {
  const res = await api.post(base(orgId, teamName), input);
  return res.data as StackPreviewConfig;
}

export async function updatePreviewConfig(orgId: string, teamName: string, id: string, input: StackPreviewConfigUpdate): Promise<StackPreviewConfig> {
  const res = await api.put(`${base(orgId, teamName)}/${id}`, input);
  return res.data as StackPreviewConfig;
}

export async function deletePreviewConfig(orgId: string, teamName: string, id: string): Promise<void> {
  await api.delete(`${base(orgId, teamName)}/${id}`);
}
