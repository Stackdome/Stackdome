import type { BlockPreset } from "@/pages/stacks/data/blocks/types";
import { BlockId } from "@/pages/stacks/data/blocks/types";
import { parseAndValidateDockerCompose } from "@/lib/docker-compose-parser";
import { convertDockerComposeToStackData } from "@/lib/docker-compose-converter";
import type {
  FormStackData,
  FormStackResourceData,
  FormVolumeData,
} from "@/pages/stacks/schemas/form-schema";
import type { DockerComposeFile } from "@/types/docker-compose";

type WorkingStack = Pick<FormStackData, "name" | "labels" | "spec">;

export function emptyStack(): WorkingStack {
  return { name: "", labels: [], spec: { stack_resources: [], volumes: [] } };
}

/** Generic blocks have no compose snippet — produce a minimal resource skeleton. */
function genericResource(block: BlockPreset): FormStackResourceData {
  const base = {
    name: block.id,
    sourceType: "image" as const,
    image_spec: { image: "" },
    ports: block.id === BlockId.Web ? [{ container_port: 8080, protocol: "TCP" }] : [],
  };
  return base as unknown as FormStackResourceData;
}

export function blockToResources(block: BlockPreset): {
  resources: FormStackResourceData[];
  volumes: FormVolumeData[];
} {
  if (!block.compose) {
    return { resources: [genericResource(block)], volumes: [] };
  }
  const parsed = parseAndValidateDockerCompose(block.compose);
  const result = convertDockerComposeToStackData(parsed as DockerComposeFile);
  if (!result.success || !result.data) {
    throw new Error(
      `Block "${block.id}" failed to convert: ${result.errors?.[0]?.message ?? "unknown"}`
    );
  }
  return {
    resources: result.data.spec.stack_resources ?? [],
    volumes: (result.data.spec.volumes ?? []) as FormVolumeData[],
  };
}

function uniqueName(base: string, taken: Set<string>): string {
  if (!taken.has(base)) return base;
  let i = 2;
  while (taken.has(`${base}-${i}`)) i += 1;
  return `${base}-${i}`;
}

export function addBlockToStack(stack: WorkingStack, block: BlockPreset): WorkingStack {
  const { resources, volumes } = blockToResources(block);
  const takenResources = new Set((stack.spec.stack_resources ?? []).map((r) => r.name));
  const takenVolumes = new Set((stack.spec.volumes ?? []).map((v) => v.name));

  const renamedResources = resources.map((r) => {
    const name = uniqueName(r.name, takenResources);
    takenResources.add(name);
    return { ...r, name };
  });

  // Build oldName → newName map for volumes that were renamed during de-duplication
  const volumeNameMap = new Map<string, string>();
  const renamedVolumes = volumes.map((v) => {
    const name = uniqueName(v.name, takenVolumes);
    takenVolumes.add(name);
    if (name !== v.name) volumeNameMap.set(v.name, name);
    return { ...v, name };
  });

  // Rewrite volume_mounts in the new resources to reflect any renamed volumes
  const rewiredResources = renamedResources.map((r) => {
    if (volumeNameMap.size === 0 || !r.volume_mounts?.length) return r;
    return {
      ...r,
      volume_mounts: r.volume_mounts.map((vm) => {
        const newSourceName = volumeNameMap.get(vm.source_volume_name);
        return newSourceName ? { ...vm, source_volume_name: newSourceName } : vm;
      }),
    };
  });

  return {
    ...stack,
    spec: {
      stack_resources: [...(stack.spec.stack_resources ?? []), ...rewiredResources],
      volumes: [...(stack.spec.volumes ?? []), ...renamedVolumes],
    },
  };
}
