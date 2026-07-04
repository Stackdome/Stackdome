import type { ResourceArr, VolumeArr } from "@/pages/stacks/lib/stack-diff";

type Resource = ResourceArr[number];
type Mount = NonNullable<Resource["volume_mounts"]>[number];

const BASE_NAME = "volume";

/** Pure calculations for canvas volume gestures. No React, unit-testable. */

export function suggestVolumeName(volumes: VolumeArr): string {
  const taken = new Set(volumes.map((v) => v.name).filter(Boolean));
  if (!taken.has(BASE_NAME)) return BASE_NAME;
  for (let i = 2; ; i++) {
    const candidate = `${BASE_NAME}-${i}`;
    if (!taken.has(candidate)) return candidate;
  }
}

/** Volume literal in the extended form shape (mirrors addInlineVolume's). */
export function newVolume(input: { name: string; size: string }): VolumeArr[number] {
  return {
    name: input.name,
    sourceType: "None",
    labels: [],
    spec: { size: input.size, access_mode: "ReadWriteOnce", needs_sync_before_use: false },
  } as unknown as VolumeArr[number];
}

export function addMount(
  resources: ResourceArr,
  resourceIdx: number,
  mount: { volumeName: string; targetPath: string; subPath?: string },
): ResourceArr {
  return resources.map((r, i) =>
    i === resourceIdx
      ? {
        ...r,
        volume_mounts: [
          ...(r.volume_mounts ?? []),
            {
              source_volume_name: mount.volumeName,
              source_sub_path: mount.subPath ?? "",
              target_path: mount.targetPath,
            } as Mount,
        ],
      }
      : r,
  );
}

export function removeMountsOf(resources: ResourceArr, volumeName: string): ResourceArr {
  return resources.map((r) =>
    (r.volume_mounts ?? []).some((m) => m.source_volume_name === volumeName)
      ? { ...r, volume_mounts: (r.volume_mounts ?? []).filter((m) => m.source_volume_name !== volumeName) }
      : r,
  );
}

export function mountOwner(
  resources: ResourceArr,
  volumeName: string,
): { resourceIdx: number; resourceName: string; targetPath: string } | null {
  for (let i = 0; i < resources.length; i++) {
    const m = (resources[i].volume_mounts ?? []).find((vm) => vm.source_volume_name === volumeName);
    if (m) return { resourceIdx: i, resourceName: resources[i].name ?? "", targetPath: m.target_path };
  }
  return null;
}

export function validateVolumeName(name: string, volumes: VolumeArr): string | undefined {
  if (!name.trim()) return "Required";
  if (volumes.some((v) => v.name === name)) return "Volume name must be unique";
  return undefined;
}

export function validateTargetPath(
  path: string,
  resource: Resource | undefined,
): string | undefined {
  if (!path.trim()) return "Required";
  if (!path.startsWith("/")) return "Must be an absolute path (start with /)";
  if ((resource?.volume_mounts ?? []).some((m) => m.target_path === path)) {
    return "This path is already mounted on the service";
  }
  return undefined;
}
