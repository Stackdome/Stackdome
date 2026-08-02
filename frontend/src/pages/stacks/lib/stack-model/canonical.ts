import type { StackResourceUpdateRequest, VolumeUpdateRequest } from "@/api/stacks";
import type { FormEnvRow, FormMountRow } from "@/pages/stacks/lib/connection-mapping";

/**
 * The one shape every "what changed?" question is asked in.
 *
 * It is the API resource shape with three departures, each where the API asks
 * the same question twice:
 *
 *  - `env` holds every environment row — literal values, self-references, and
 *    references to secrets/addons/resources alike. The API splits those across
 *    `execution_config.environment_variables` and `connections`, which is a
 *    storage decision; to a reader of the stack they are one list.
 *  - `mounts` holds volume mounts, which the API also stores as connections.
 *  - `command`/`args` are argv arrays, never the drawer's shell-ish text.
 *
 * Everything else passes through under its API name, unnormalized fields
 * included, so a field added to the OpenAPI spec is diffed without anyone
 * teaching this module about it.
 */
export interface CanonicalResource
  extends Omit<StackResourceUpdateRequest, "volume_mounts" | "execution_config"> {
  name: string;
  /** Sorted by name, so storage order and authoring order never read as a change. */
  env: FormEnvRow[];
  /** Sorted by target path, for the same reason. */
  mounts: FormMountRow[];
  execution_config: { command: string[]; args: string[] };
}

export interface CanonicalVolume extends VolumeUpdateRequest {
  name: string;
}

export interface CanonicalStack {
  resources: CanonicalResource[];
  volumes: CanonicalVolume[];
}

export const EMPTY_CANONICAL_STACK: CanonicalStack = { resources: [], volumes: [] };

export function resourcesByName(stack: CanonicalStack): Map<string, CanonicalResource> {
  return new Map(stack.resources.map((r) => [r.name, r]));
}

export function volumesByName(stack: CanonicalStack): Map<string, CanonicalVolume> {
  return new Map(stack.volumes.map((v) => [v.name, v]));
}

export function sortEnv(rows: FormEnvRow[]): FormEnvRow[] {
  return [...rows].sort((a, b) => (a.name ?? "").localeCompare(b.name ?? ""));
}

export function sortMounts(rows: FormMountRow[]): FormMountRow[] {
  return [...rows].sort((a, b) => (a.target_path ?? "").localeCompare(b.target_path ?? ""));
}
