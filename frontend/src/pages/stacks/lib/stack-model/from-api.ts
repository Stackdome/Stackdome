import type { Stack, StackResource, Volume } from "@/api/stacks";
import type { StackConnection } from "@/api/connections";
import type { StackReleaseSnapshot } from "@/api/releases";
import { connectionsToEnvRows, connectionsToMounts, type FormEnvRow } from "@/pages/stacks/lib/connection-mapping";
import {
  sortEnv,
  sortMounts,
  type CanonicalResource,
  type CanonicalStack,
  type CanonicalVolume,
} from "./canonical";
import {
  normalizeSource,
  normalizeWorkloadType,
  omitServerWrittenResourceFields,
  omitServerWrittenVolumeFields,
} from "./normalize";

/** Literal and self-referencing rows, the two kinds the API keeps on the
 *  resource itself. Everything else arrives as a connection. */
function envRowsFromApi(r: StackResource): FormEnvRow[] {
  return (r.execution_config?.environment_variables ?? []).map((v) =>
    v.self_output
      ? { from: "self" as const, name: v.name ?? "", selfOutput: v.self_output }
      : { from: "stack" as const, name: v.name ?? "", value: v.value ?? "" },
  );
}

export function canonicalResourceFromApi(
  r: StackResource,
  connections: StackConnection[],
): CanonicalResource {
  const name = r.name ?? "";
  const rest = omitServerWrittenResourceFields(r) as Record<string, unknown>;
  delete rest.volume_mounts;
  delete rest.execution_config;
  return {
    ...rest,
    name,
    workload_type: normalizeWorkloadType(r.workload_type),
    source: normalizeSource(r.source),
    env: sortEnv([...envRowsFromApi(r), ...connectionsToEnvRows(name, connections)]),
    mounts: sortMounts(connectionsToMounts(name, connections)),
    execution_config: {
      command: r.execution_config?.command ?? [],
      args: r.execution_config?.args ?? [],
    },
  } as CanonicalResource;
}

export function canonicalVolumeFromApi(v: Volume): CanonicalVolume {
  return { ...omitServerWrittenVolumeFields(v), name: v.name ?? "" } as CanonicalVolume;
}

function canonicalStackFrom(
  resources: StackResource[] | undefined,
  volumes: Volume[] | undefined,
  connections: StackConnection[] | undefined,
): CanonicalStack {
  const conns = connections ?? [];
  return {
    resources: (resources ?? []).filter((r) => r.name).map((r) => canonicalResourceFromApi(r, conns)),
    volumes: (volumes ?? []).filter((v) => v.name).map(canonicalVolumeFromApi),
  };
}

/** The saved server spec. */
export function canonicalFromStack(stack: Stack | undefined): CanonicalStack {
  return canonicalStackFrom(
    stack?.spec?.stack_resources,
    stack?.spec?.volumes,
    stack?.spec?.connections as StackConnection[] | undefined,
  );
}

/** A release snapshot — the same server shapes under different key names. */
export function canonicalFromSnapshot(snap: StackReleaseSnapshot | undefined): CanonicalStack {
  return canonicalStackFrom(
    snap?.resources as StackResource[] | undefined,
    snap?.volumes as Volume[] | undefined,
    snap?.connections as StackConnection[] | undefined,
  );
}
