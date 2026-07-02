import type { z } from "zod";
import type { StackResourceUpdateRequest, VolumeUpdateRequest } from "@/api/stacks";
import type { StackConnection } from "@/api/connections";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";
import {
  FormStackResourceSchema,
  prepareFormResourceForApi,
  convertFormVolumeToApiVolume,
  type FormStackResourceData,
  type FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema";
import { buildDesiredConnections, type FormEnvRow } from "@/pages/stacks/lib/connection-mapping";
import { connectionIdentityKey } from "./server-state";

/**
 * What the server SHOULD hold, derived from the edit-session draft.
 * Invalid-but-named resources are "held": they produce no ops and exempt
 * their server-side counterparts (and connections) from deletion — a
 * half-typed resource must never read as deleted.
 */
export interface DesiredStackState {
  resources: Map<string, StackResourceUpdateRequest>;
  held: Set<string>;
  volumes: Map<string, VolumeUpdateRequest>;
  connections: Map<string, StackConnection>;
  /** zod issues per draft-resource index, for live drawer errors. */
  resourceIssues: Map<number, z.ZodIssue[]>;
}

export function buildDesiredState(draft: EditSessionDraft): DesiredStackState {
  const resources = new Map<string, StackResourceUpdateRequest>();
  const held = new Set<string>();
  const resourceIssues = new Map<number, z.ZodIssue[]>();
  const validForConnections: { name: string; rows: FormEnvRow[] }[] = [];

  draft.resources.forEach((raw, idx) => {
    const parsed = FormStackResourceSchema.safeParse(raw);
    if (parsed.success) {
      const data = parsed.data as FormStackResourceData;
      if (!data.name?.trim()) return; // unnamed: nothing to sync
      resources.set(data.name, prepareFormResourceForApi(data));
      validForConnections.push({
        name: data.name,
        rows: (data.execution_config?.environment_variables ?? []) as FormEnvRow[],
      });
      return;
    }
    resourceIssues.set(idx, parsed.error.issues);
    const name = (raw as Partial<FormStackResourceData>).name?.trim();
    if (name) held.add(name);
  });

  const volumes = new Map<string, VolumeUpdateRequest>();
  for (const raw of draft.volumes) {
    const name = (raw as Partial<FormVolumeExtendedData>).name?.trim();
    if (!name) continue;
    volumes.set(name, convertFormVolumeToApiVolume(raw as FormVolumeExtendedData));
  }

  const connections = new Map<string, StackConnection>();
  for (const conn of buildDesiredConnections(validForConnections)) {
    connections.set(connectionIdentityKey(conn), conn);
  }

  return { resources, held, volumes, connections, resourceIssues };
}
