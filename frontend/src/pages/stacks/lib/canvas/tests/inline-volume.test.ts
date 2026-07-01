import { describe, it, expect } from "vitest";
import { addInlineVolume } from "../inline-volume";

describe("addInlineVolume", () => {
  it("appends a stack volume and a matching resource mount", () => {
    const { volumes, mounts } = addInlineVolume([], [], { name: "data", size: "2Gi", targetPath: "/var/lib/data" });
    expect(volumes).toEqual([
      {
        name: "data",
        sourceType: "None",
        labels: [],
        spec: { size: "2Gi", access_mode: "ReadWriteOnce", needs_sync_before_use: false },
      },
    ]);
    expect(mounts).toEqual([{ source_volume_name: "data", source_sub_path: "", target_path: "/var/lib/data" }]);
  });

  it("does not clobber existing volumes/mounts", () => {
    const { volumes, mounts } = addInlineVolume(
      [{ name: "a", sourceType: "None", labels: [], spec: { size: "1Gi", access_mode: "ReadWriteOnce", needs_sync_before_use: false } }] as never,
      [{ source_volume_name: "a", source_sub_path: "", target_path: "/a" }] as never,
      { name: "b", size: "5Gi", targetPath: "/b" },
    );
    expect(volumes.map((v) => v.name)).toEqual(["a", "b"]);
    expect(mounts.map((m) => m.target_path)).toEqual(["/a", "/b"]);
  });
});
