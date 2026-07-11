import api from "./client";
import type { components } from "./types/openapi";

export type PreviewStack = components["schemas"]["PreviewStack"];
export type PreviewStackCreate = components["schemas"]["PreviewStackCreate"];
export type PreviewStackList = components["schemas"]["PreviewStackList"];
export type PreviewStackSync = components["schemas"]["PreviewStackSync"];

export type PreviewPhase = NonNullable<NonNullable<PreviewStack["status"]>["phase"]>;

/** Phases where the backend has finished reconciling; polling can stop. */
export const TERMINAL_PHASES: PreviewPhase[] = ["Ready", "Failed"];

// Labels the backend stamps on stacks created by preview environments
// (pkg/models/preview_stack.go). Used to detect preview ownership in the UI.
export const PREVIEW_STACK_LABEL = "preview.stackdome.io/preview-stack";
export const PREVIEW_ID_LABEL = "preview.stackdome.io/preview-id";

function base(orgId: string, teamName: string): string {
  return `/organizations/${orgId}/teams/${teamName}/preview-stacks`;
}

export interface ListPreviewEnvOpts {
  configId?: string;
  page?: number;
  pageSize?: number;
}

export async function listPreviewEnvs(
  orgId: string,
  teamName: string,
  opts: ListPreviewEnvOpts = {},
): Promise<PreviewStackList> {
  const params: Record<string, string | number> = {};
  if (opts.configId) params.config_id = opts.configId;
  if (opts.page) params.page = opts.page;
  if (opts.pageSize) params.page_size = opts.pageSize;
  const res = await api.get(base(orgId, teamName), { params });
  return res.data as PreviewStackList;
}

/** Fetches every page so callers see the complete env set, not the first 20. */
export async function listAllPreviewEnvs(orgId: string, teamName: string, configId?: string): Promise<PreviewStack[]> {
  const pageSize = 100;
  const items: PreviewStack[] = [];
  for (let page = 1; ; page++) {
    const res = await listPreviewEnvs(orgId, teamName, { configId, page, pageSize });
    const batch = res.items ?? [];
    items.push(...batch);
    const total = res.total ?? items.length;
    if (batch.length === 0 || items.length >= total) break;
  }
  return items;
}

export async function getPreviewEnv(
  orgId: string,
  teamName: string,
  id: string,
): Promise<PreviewStack> {
  const res = await api.get(`${base(orgId, teamName)}/${id}`);
  return res.data as PreviewStack;
}

export async function createPreviewEnv(
  orgId: string,
  teamName: string,
  input: PreviewStackCreate,
): Promise<PreviewStack> {
  const res = await api.post(base(orgId, teamName), input);
  return res.data as PreviewStack;
}

export async function deletePreviewEnv(
  orgId: string,
  teamName: string,
  id: string,
): Promise<PreviewStack> {
  const res = await api.delete(`${base(orgId, teamName)}/${id}`);
  return res.data as PreviewStack;
}

export async function syncPreviewEnv(
  orgId: string,
  teamName: string,
  id: string,
  input: PreviewStackSync = {},
): Promise<PreviewStack> {
  const res = await api.post(`${base(orgId, teamName)}/${id}/sync`, input);
  return res.data as PreviewStack;
}
