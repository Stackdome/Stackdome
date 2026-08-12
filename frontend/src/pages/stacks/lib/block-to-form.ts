import type { BlockPreset } from "@/pages/stacks/data/blocks/types";
import { BlockId, DATA_BLOCK_CATEGORIES } from "@/pages/stacks/data/blocks/types";
import { PLACEHOLDER_PASSWORDS } from "@/pages/stacks/data/blocks/registry";
import { parseAndValidateDockerCompose } from "@/pages/stacks/lib/docker-compose-parser";
import { convertDockerComposeToStackData } from "@/pages/stacks/lib/docker-compose-converter";
import { renameResourceReferencesByMap } from "@/pages/stacks/lib/rename-references";
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

/** Generic blocks have no compose snippet — produce a minimal resource skeleton.
 *  Web gets the conventional public HTTP port 80; Custom starts portless. */
function genericResource(block: BlockPreset): FormStackResourceData {
  const base = {
    name: block.id,
    sourceType: "image" as const,
    source: { image: { ref: "" } },
    ports:
      block.id === BlockId.Web
        ? [{ name: "http-80", number: 80, protocol: "http", exposed_to_public: true }]
        : [],
  };
  return base as unknown as FormStackResourceData;
}

/** Random secret for data-store blocks whose images refuse to boot with an
 *  empty password (postgres/mysql/mariadb/mssql/couchdb). Alphanumeric with a
 *  guaranteed upper+lower+digit mix so even MSSQL's complexity check passes. */
function generatedPassword(): string {
  const charset = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789";
  const bytes = new Uint32Array(17);
  crypto.getRandomValues(bytes);
  const body = Array.from(bytes, (b) => charset[b % charset.length]).join("");
  return `Sd1${body}`;
}

type FormPort = { name?: string; number?: number; protocol?: string; exposed_to_public?: boolean };

/** Data stores are internal dependencies: their ports must never be published
 *  as public HTTP ingresses (the compose converter defaults every published
 *  port to http/public — right for web services, wrong for databases). */
function internalizePorts(resource: FormStackResourceData): FormStackResourceData {
  const ports = (resource.ports ?? []) as FormPort[];
  if (ports.length === 0) return resource;
  return {
    ...resource,
    ports: ports.map((p) => ({
      ...p,
      name: `tcp-${p.number}`,
      protocol: "tcp",
      exposed_to_public: false,
    })),
  } as FormStackResourceData;
}

/** Swap the registry's placeholder passwords (and any empty env value) for a
 *  generated secret so the container actually boots on first deploy. */
function fillPlaceholderPasswords(resource: FormStackResourceData): FormStackResourceData {
  const isPlaceholder = (v: string | undefined) => v === "" || (v !== undefined && PLACEHOLDER_PASSWORDS.has(v));
  const env = resource.execution_config?.environment_variables as
    | { name?: string; value?: string; from?: string }[]
    | undefined;
  if (!env?.some((r) => isPlaceholder(r.value))) return resource;
  return {
    ...resource,
    execution_config: {
      ...resource.execution_config,
      environment_variables: env.map((r) => (isPlaceholder(r.value) ? { ...r, value: generatedPassword() } : r)),
    },
  } as FormStackResourceData;
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
  const resources = (result.data.spec.stack_resources ?? []).map((r) =>
    DATA_BLOCK_CATEGORIES.has(block.category) ? fillPlaceholderPasswords(internalizePorts(r)) : r,
  );
  return {
    resources,
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

  // Build oldName → newName for resources renamed during de-duplication, so the
  // block's own depends_on and env refs follow its copy instead of silently
  // binding to the same-named resource already in the stack.
  const resourceNameMap = new Map<string, string>();
  const renamedResources = resources.map((r) => {
    const name = uniqueName(r.name, takenResources);
    takenResources.add(name);
    if (name !== r.name) resourceNameMap.set(r.name, name);
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

  const wiredResources = renameResourceReferencesByMap(
    rewiredResources,
    resourceNameMap,
  ) as FormStackResourceData[];

  return {
    ...stack,
    spec: {
      stack_resources: [...(stack.spec.stack_resources ?? []), ...wiredResources],
      volumes: [...(stack.spec.volumes ?? []), ...renamedVolumes],
    },
  };
}
