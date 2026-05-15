// @vitest-environment node
import { describe, it, expect } from "vitest";
import { filterAndSortAddons, type AddonStatusFilter } from "../addon-list-filter";
import type { PostgresAddon } from "@/api/addons";

const mk = (name: string, state: string, created: string) =>
  ({ id: name, name, status: { state }, created_at: created, spec: {} } as unknown as PostgresAddon);

const list = [
  mk("alpha", "Ready", "2026-05-01T00:00:00Z"),
  mk("beta", "Error", "2026-05-03T00:00:00Z"),
  mk("gamma", "Creating", "2026-05-02T00:00:00Z"),
];

describe("filterAndSortAddons", () => {
  it("filters by query (case-insensitive substring on name)", () => {
    expect(filterAndSortAddons(list, "ALP", "all", "name").map((a) => a.name)).toEqual(["alpha"]);
  });
  it("filters by status bucket", () => {
    const f: AddonStatusFilter = "error";
    expect(filterAndSortAddons(list, "", f, "name").map((a) => a.name)).toEqual(["beta"]);
  });
  it("sorts by name A–Z", () => {
    expect(filterAndSortAddons(list, "", "all", "name").map((a) => a.name)).toEqual(["alpha", "beta", "gamma"]);
  });
  it("sorts by recently created (desc)", () => {
    expect(filterAndSortAddons(list, "", "all", "created").map((a) => a.name)).toEqual(["beta", "gamma", "alpha"]);
  });
});
