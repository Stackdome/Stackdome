import type { z } from "zod";
import type { StackResource, Volume } from "@/pages/stacks/types";
import type { StackConnection } from "@/api/connections";
import {
  convertApiResourceToFormResource,
  convertApiVolumeToFormVolume,
  type FormStackResourceData,
  type FormVolumeExtendedData as VolumeFormData,
  type FormEnvVarData,
} from "@/pages/stacks/schemas/form-schema";
import type { ApiStackResourceSchema, ApiVolumeSchema } from "@/pages/stacks/schemas/api-schema";
import { connectionsToEnvRows, connectionsToMounts } from "@/pages/stacks/lib/connection-mapping";

export function mapStackResourceToFormData(resource: StackResource): FormStackResourceData {
  // Remove read-only fields before converting to form data
  const { id: _id, stack_id: _stackId, revision: _revision, ...writableResource } = resource;

  const cleanedVolumeMounts = writableResource.volume_mounts?.map((volumeMount) => {
    const { stack_resource_id: _stackResourceId, source_volume_type: _sourceVolumeType, ...cleanVolumeMount } = volumeMount;
    return cleanVolumeMount;
  });

  const resourceWithCleanedMounts = {
    ...writableResource,
    volume_mounts: cleanedVolumeMounts
  };

  return convertApiResourceToFormResource(resourceWithCleanedMounts as z.infer<typeof ApiStackResourceSchema>);
}

export function mapVolumeToFormData(volume: Volume): VolumeFormData {
  // Remove read-only fields before converting to form data
  const { id: _id, ...writableVolume } = volume;
  return convertApiVolumeToFormVolume(writableVolume as z.infer<typeof ApiVolumeSchema> & { status?: unknown });
}

/** Map a resource+connection set (live stack spec OR release snapshot — both use
 *  the same server shapes) into form data, folding connection-backed env rows
 *  and volume mounts into each resource. */
export function formResourcesFromSpec(
  resources: StackResource[] | undefined,
  connectionsIn: StackConnection[] | undefined,
): FormStackResourceData[] {
  const connections = connectionsIn ?? [];
  return (resources || []).map((r) => {
    const form = mapStackResourceToFormData(r);
    const connRows = connectionsToEnvRows(form.name ?? "", connections) as FormEnvVarData[];
    // Populate volume_mounts from volume_mount connections — the server always
    // returns resource.volume_mounts as [] since mounts are stored in connections.
    const mountRows = connectionsToMounts(form.name ?? "", connections);
    // connectionsToMounts only emits rows with all required fields present (it
    // skips malformed connections), so the cast to the strict form type is safe.
    const withMounts: FormStackResourceData = { ...form, volume_mounts: mountRows as FormStackResourceData["volume_mounts"] };
    if (connRows.length === 0) return withMounts;
    return {
      ...withMounts,
      execution_config: {
        ...(withMounts.execution_config ?? {}),
        environment_variables: [
          ...((withMounts.execution_config?.environment_variables ?? []) as FormEnvVarData[]),
          ...connRows,
        ],
      },
    };
  });
}
