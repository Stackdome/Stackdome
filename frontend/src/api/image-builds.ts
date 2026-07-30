import api from "./client";
import { API_BASE_URL } from "./base-url";
import type { components } from "./types/openapi";

export type ImageBuild = components["schemas"]["ImageBuild"];

/** Mirrors cluster-agent buildsv1alpha1.BuildJobCreated — the backend refuses
 *  log streaming (409) until this condition is True. */
export const BUILD_JOB_CREATED_CONDITION = "BuildJobCreated";
const CONDITION_TRUE = "True";

export async function getImageBuild(
  orgId: string,
  projectName: string,
  stackId: string,
  buildId: string,
): Promise<ImageBuild> {
  const res = await api.get(
    `/organizations/${orgId}/projects/${projectName}/stacks/${stackId}/builds/${buildId}`,
  );
  return res.data;
}

export function isBuildJobCreated(build: ImageBuild): boolean {
  return (build.status?.conditions ?? []).some(
    (c) => c.type === BUILD_JOB_CREATED_CONDITION && c.status === CONDITION_TRUE,
  );
}

interface BuildLogStreamParams {
  follow?: boolean;
  tail?: number;
  since?: string;
}

export function buildImageBuildLogStreamUrl(
  orgId: string,
  projectName: string,
  stackId: string,
  buildId: string,
  params?: BuildLogStreamParams,
): string {
  const path = `${API_BASE_URL}/organizations/${orgId}/projects/${projectName}/stacks/${stackId}/builds/${buildId}/logs`;
  const searchParams = new URLSearchParams();
  if (params?.follow !== undefined) searchParams.set("follow", String(params.follow));
  if (params?.since) searchParams.set("since", params.since);
  if (params?.tail !== undefined) searchParams.set("tail", String(params.tail));
  const qs = searchParams.toString();
  return qs ? `${path}?${qs}` : path;
}
