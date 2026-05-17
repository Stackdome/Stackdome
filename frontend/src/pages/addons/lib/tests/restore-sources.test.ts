// @vitest-environment node
import { describe, it, expect } from "vitest";
import { eligibleRestoreSources } from "../restore-sources";
import type { PostgresAddon } from "@/api/addons";

function addon(p: Partial<PostgresAddon> & { id: string }): PostgresAddon {
  return {
    id: p.id,
    name: p.name ?? p.id,
    spec: p.spec ?? ({} as PostgresAddon["spec"]),
  } as PostgresAddon;
}

describe("eligibleRestoreSources", () => {
  it("keeps any addon with an object store (WAL on OR off)", () => {
    const list = [
      addon({ id: "a", spec: { backup: { object_store_id: "os1", wal_archiving: true } } as PostgresAddon["spec"] }),
      addon({ id: "b", spec: { backup: { object_store_id: "os2", wal_archiving: false } } as PostgresAddon["spec"] }),
      addon({ id: "c", spec: { backup: { wal_archiving: true } } as PostgresAddon["spec"] }),
      addon({ id: "d", spec: {} as PostgresAddon["spec"] }),
    ];
    expect(eligibleRestoreSources(list).map((a) => a.id)).toEqual(["a", "b"]);
  });

  it("excludes a given addon id (cannot restore from self)", () => {
    const list = [
      addon({ id: "self", spec: { backup: { object_store_id: "os1" } } as PostgresAddon["spec"] }),
      addon({ id: "other", spec: { backup: { object_store_id: "os2" } } as PostgresAddon["spec"] }),
    ];
    expect(eligibleRestoreSources(list, "self").map((a) => a.id)).toEqual(["other"]);
  });
});
