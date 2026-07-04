import { describe, expect, it } from "vitest";
import {
  addMount,
  mountOwner,
  newVolume,
  removeMountsOf,
  suggestVolumeName,
  validateTargetPath,
  validateVolumeName,
} from "../volume-ops";

const vol = (name: string) => ({ name, spec: { size: "1Gi", access_mode: "ReadWriteOnce", needs_sync_before_use: false } });
const res = (name: string, mounts: Array<{ source_volume_name: string; target_path: string }> = []) => ({
  name,
  volume_mounts: mounts.map((m) => ({ ...m, source_sub_path: "" })),
});

describe("suggestVolumeName", () => {
  it("starts at 'volume' and skips taken names", () => {
    expect(suggestVolumeName([])).toBe("volume");
    expect(suggestVolumeName([vol("volume")])).toBe("volume-2");
    expect(suggestVolumeName([vol("volume"), vol("volume-2")])).toBe("volume-3");
  });
});

describe("newVolume", () => {
  it("builds the extended-form volume literal", () => {
    const v = newVolume({ name: "data", size: "2Gi" }) as Record<string, unknown>;
    expect(v.name).toBe("data");
    expect(v.sourceType).toBe("None");
    expect((v.spec as Record<string, unknown>).size).toBe("2Gi");
    expect((v.spec as Record<string, unknown>).access_mode).toBe("ReadWriteOnce");
  });
});

describe("addMount / removeMountsOf / mountOwner", () => {
  it("appends a mount to the target resource only", () => {
    const next = addMount([res("web"), res("api")], 1, { volumeName: "data", targetPath: "/var/data" });
    expect(next[0].volume_mounts).toEqual([]);
    expect(next[1].volume_mounts).toEqual([
      { source_volume_name: "data", source_sub_path: "", target_path: "/var/data" },
    ]);
  });

  it("removeMountsOf strips the volume's mounts from every resource", () => {
    const resources = [
      res("web", [{ source_volume_name: "data", target_path: "/a" }]),
      res("api", [
        { source_volume_name: "data", target_path: "/b" },
        { source_volume_name: "other", target_path: "/c" },
      ]),
    ];
    const next = removeMountsOf(resources, "data");
    expect(next[0].volume_mounts).toEqual([]);
    expect(next[1].volume_mounts).toEqual([{ source_volume_name: "other", source_sub_path: "", target_path: "/c" }]);
  });

  it("mountOwner returns the first mounting resource, or null when unmounted", () => {
    const resources = [res("web"), res("api", [{ source_volume_name: "data", target_path: "/d" }])];
    expect(mountOwner(resources, "data")).toEqual({ resourceIdx: 1, resourceName: "api", targetPath: "/d" });
    expect(mountOwner(resources, "ghost")).toBeNull();
  });
});

describe("validateVolumeName", () => {
  it("requires a value and rejects duplicates", () => {
    expect(validateVolumeName("", [])).toBeTruthy();
    expect(validateVolumeName("data", [vol("data")])).toBeTruthy();
    expect(validateVolumeName("fresh", [vol("data")])).toBeUndefined();
  });
});

describe("validateTargetPath", () => {
  it("requires an absolute path", () => {
    expect(validateTargetPath("", res("web"))).toBeTruthy();
    expect(validateTargetPath("data", res("web"))).toBeTruthy();
    expect(validateTargetPath("/data", res("web"))).toBeUndefined();
  });
  it("rejects a target path already mounted on the resource", () => {
    const r = res("web", [{ source_volume_name: "data", target_path: "/data" }]);
    expect(validateTargetPath("/data", r)).toBeTruthy();
    expect(validateTargetPath("/other", r)).toBeUndefined();
  });
});
