import type { StackResourceUpdateRequest, VolumeUpdateRequest } from "@/api/stacks";
import type { StackConnection } from "@/api/connections";
import { mountsToConnections, splitEnvRows } from "@/pages/stacks/lib/connection-mapping";
import type { CanonicalResource, CanonicalStack, CanonicalVolume } from "./canonical";

/**
 * Canonical back to the wire. `env` and `mounts` split into the resource's own
 * environment variables and the stack's connections — the storage shape the
 * canonical model exists to hide from everyone upstream of here.
 */
export function resourceToApi(r: CanonicalResource): StackResourceUpdateRequest {
  const { env, mounts, execution_config, ...rest } = r as CanonicalResource & Record<string, unknown>;
  void mounts;
  const { envVars } = splitEnvRows(r.name, env ?? []);
  return {
    ...rest,
    volume_mounts: undefined,
    execution_config: {
      ...execution_config,
      command: execution_config?.command ?? [],
      args: execution_config?.args ?? [],
      environment_variables: envVars,
    },
  } as StackResourceUpdateRequest;
}

export function volumeToApi(v: CanonicalVolume): VolumeUpdateRequest {
  return { ...v } as VolumeUpdateRequest;
}

/** Every connection the canonical stack implies: env references and mounts. */
export function connectionsOf(stack: CanonicalStack): StackConnection[] {
  return stack.resources.flatMap((r) => [
    ...splitEnvRows(r.name, r.env ?? []).connections,
    ...mountsToConnections(r.name, r.mounts ?? []),
  ]);
}
