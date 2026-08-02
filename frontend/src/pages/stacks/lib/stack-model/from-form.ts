import type { z } from "zod";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";
import {
  FormStackResourceSchema,
  prepareFormResourceForApi,
  convertFormVolumeToApiVolume,
  type FormStackResourceData,
  type FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema";
import {
  connectionsToEnvRows,
  connectionsToMounts,
  mountsToConnections,
  splitEnvRows,
  type FormEnvRow,
  type FormMountRow,
} from "@/pages/stacks/lib/connection-mapping";
import {
  sortEnv,
  sortMounts,
  type CanonicalResource,
  type CanonicalStack,
  type CanonicalVolume,
} from "./canonical";
import { normalizeSource, normalizeWorkloadType, omitServerWrittenResourceFields } from "./normalize";

export interface CanonicalDraft extends CanonicalStack {
  /** Resources that fail validation but carry a name. They produce no ops and
   *  exempt their server counterparts from deletion — a half-typed resource
   *  must never read as deleted. */
  held: Set<string>;
  /** zod issues per draft-resource index, for live drawer errors. */
  issues: Map<number, z.ZodIssue[]>;
  /** Draft array index per resource name, for the index-keyed drawer UI. */
  indexByName: Map<string, number>;
}

/**
 * Env and mount rows are round-tripped through the same persistence split the
 * save path uses, then read back the way `from-api` reads them. Rows that would
 * never survive a save (half-typed references, mounts of a deleted volume) are
 * dropped by that round trip, not by a second set of rules here.
 */
function persistedEnvAndMounts(
  name: string,
  rows: FormEnvRow[],
  mounts: FormMountRow[],
  liveVolumeNames: Set<string> | undefined,
): { env: FormEnvRow[]; mounts: FormMountRow[] } {
  const { envVars, connections } = splitEnvRows(name, rows);
  // No volume set means "not filtering" — the caller is asking about one
  // resource in isolation and has no stack to check mounts against.
  const liveMounts = liveVolumeNames
    ? mounts.filter((m) => liveVolumeNames.has(m.source_volume_name ?? ""))
    : mounts;
  const mountConnections = mountsToConnections(name, liveMounts);
  const literalRows: FormEnvRow[] = envVars.map((v) =>
    v.self_output
      ? { from: "self", name: v.name ?? "", selfOutput: v.self_output }
      : { from: "stack", name: v.name ?? "", value: v.value ?? "" },
  );
  return {
    env: sortEnv([...literalRows, ...connectionsToEnvRows(name, connections)]),
    mounts: sortMounts(connectionsToMounts(name, mountConnections)),
  };
}

export function canonicalResourceFromForm(
  data: FormStackResourceData,
  liveVolumeNames?: Set<string>,
): CanonicalResource {
  const name = data.name!.trim();
  const api = prepareFormResourceForApi(data) as Record<string, unknown>;
  const rest = omitServerWrittenResourceFields(api);
  delete rest.volume_mounts;
  delete rest.execution_config;
  const { env, mounts } = persistedEnvAndMounts(
    name,
    (data.execution_config?.environment_variables ?? []) as FormEnvRow[],
    (data.volume_mounts ?? []) as FormMountRow[],
    liveVolumeNames,
  );
  const execution = (api.execution_config ?? {}) as { command?: string[]; args?: string[] };
  return {
    ...rest,
    name,
    workload_type: normalizeWorkloadType(data.workload_type),
    source: normalizeSource(api.source as never),
    env,
    mounts,
    execution_config: {
      ...execution,
      environment_variables: undefined,
      command: execution.command ?? [],
      args: execution.args ?? [],
    },
  } as CanonicalResource;
}

export function canonicalVolumeFromForm(raw: FormVolumeExtendedData): CanonicalVolume {
  return { ...convertFormVolumeToApiVolume(raw), name: raw.name!.trim() } as CanonicalVolume;
}

export function canonicalFromDraft(draft: EditSessionDraft): CanonicalDraft {
  const volumes: CanonicalVolume[] = [];
  for (const raw of draft.volumes) {
    const name = (raw as Partial<FormVolumeExtendedData>).name?.trim();
    if (!name) continue;
    volumes.push(canonicalVolumeFromForm(raw as FormVolumeExtendedData));
  }
  const liveVolumeNames = new Set(volumes.map((v) => v.name));

  const resources: CanonicalResource[] = [];
  const held = new Set<string>();
  const issues = new Map<number, z.ZodIssue[]>();
  const indexByName = new Map<string, number>();

  draft.resources.forEach((raw, idx) => {
    // Validate user intent only. `status` and `outputs` ride along on form data
    // for display; validating them treats server writes as user input, and one
    // unrecognized subfield would hold the whole resource out of every sync.
    const parsed = FormStackResourceSchema.safeParse(omitServerWrittenResourceFields(raw));
    if (!parsed.success) {
      issues.set(idx, parsed.error.issues);
      const name = (raw as Partial<FormStackResourceData>).name?.trim();
      if (name) {
        held.add(name);
        indexByName.set(name, idx);
      }
      return;
    }
    const data = parsed.data as FormStackResourceData;
    if (!data.name?.trim()) return;
    indexByName.set(data.name.trim(), idx);
    resources.push(canonicalResourceFromForm(data, liveVolumeNames));
  });

  return { resources, volumes, held, issues, indexByName };
}
