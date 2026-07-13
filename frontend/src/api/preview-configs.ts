import api from "./client";
import type { components } from "./types/openapi";

export type StackPreviewConfig = components["schemas"]["StackPreviewConfig"];
export type StackPreviewConfigCreate = components["schemas"]["StackPreviewConfigCreate"];
export type StackPreviewConfigUpdate = components["schemas"]["StackPreviewConfigUpdate"];
export type StackPreviewConfigList = components["schemas"]["StackPreviewConfigList"];
export type PreviewGitRepository = components["schemas"]["PreviewGitRepository"];

function base(orgId: string, projectName: string): string {
  return `/organizations/${orgId}/projects/${projectName}/stack-preview-configs`;
}

export async function listPreviewConfigs(orgId: string, projectName: string, page = 1, pageSize = 20): Promise<StackPreviewConfigList> {
  const res = await api.get(base(orgId, projectName), { params: { page, page_size: pageSize } });
  return res.data as StackPreviewConfigList;
}

/** Fetches every page so callers see the complete config set, not the first 20. */
export async function listAllPreviewConfigs(orgId: string, projectName: string): Promise<StackPreviewConfig[]> {
  const pageSize = 100;
  const items: StackPreviewConfig[] = [];
  for (let page = 1; ; page++) {
    const res = await listPreviewConfigs(orgId, projectName, page, pageSize);
    const batch = res.items ?? [];
    items.push(...batch);
    const total = res.total ?? items.length;
    if (batch.length === 0 || items.length >= total) break;
  }
  return items;
}

export async function getPreviewConfig(orgId: string, projectName: string, id: string): Promise<StackPreviewConfig> {
  const res = await api.get(`${base(orgId, projectName)}/${id}`);
  return res.data as StackPreviewConfig;
}

export async function createPreviewConfig(orgId: string, projectName: string, input: StackPreviewConfigCreate): Promise<StackPreviewConfig> {
  const res = await api.post(base(orgId, projectName), input);
  return res.data as StackPreviewConfig;
}

export async function updatePreviewConfig(orgId: string, projectName: string, id: string, input: StackPreviewConfigUpdate): Promise<StackPreviewConfig> {
  const res = await api.put(`${base(orgId, projectName)}/${id}`, input);
  return res.data as StackPreviewConfig;
}

export async function deletePreviewConfig(orgId: string, projectName: string, id: string): Promise<void> {
  await api.delete(`${base(orgId, projectName)}/${id}`);
}
