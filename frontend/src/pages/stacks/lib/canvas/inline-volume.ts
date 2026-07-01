import type {
  FormStackResourceData,
  FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema";

type VolumeMount = NonNullable<FormStackResourceData["volume_mounts"]>[number];

export interface InlineVolumeInput {
  name: string;
  size: string;
  targetPath: string;
}

/** Create a stack-level volume and the resource mount that references it, in one step. */
export function addInlineVolume(
  volumes: FormVolumeExtendedData[],
  mounts: VolumeMount[],
  input: InlineVolumeInput,
): { volumes: FormVolumeExtendedData[]; mounts: VolumeMount[] } {
  // Structural cast: the literal satisfies the `sourceType: "None"` branch of FormVolumeExtendedData's discriminated union; the double-cast bypasses the union narrowing that TypeScript cannot verify from an object literal.
  const volume = {
    name: input.name,
    sourceType: "None" as const,
    labels: [],
    spec: { size: input.size, access_mode: "ReadWriteOnce", needs_sync_before_use: false },
  } as unknown as FormVolumeExtendedData;
  const mount = { source_volume_name: input.name, source_sub_path: "", target_path: input.targetPath } as VolumeMount;
  return { volumes: [...volumes, volume], mounts: [...mounts, mount] };
}
