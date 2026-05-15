// @vitest-environment node
import { describe, it, expect } from "vitest";
import { buildCreateInput, addonToFormValues } from "../payload";
import { defaultFormValues } from "../../schemas/form-schema";
import type { PostgresAddon } from "@/api/addons";

function base() {
  const v = defaultFormValues("");
  v.name = "db1";
  return v;
}

describe("payload backup mapping", () => {
  it("omits spec.backup when backup disabled", () => {
    const input = buildCreateInput(base());
    expect(input.spec.backup).toBeUndefined();
  });

  it("writes spec.backup when backup enabled", () => {
    const v = base();
    v.backup = { enabled: true, objectStoreId: "store-1", schedule: "0 0 2 * * *", walArchiving: true };
    const input = buildCreateInput(v);
    expect(input.spec.backup).toEqual({
      enabled: true,
      schedule: "0 0 2 * * *",
      wal_archiving: true,
      object_store_id: "store-1",
    });
  });

  it("hydrates backup from an existing addon (round-trip)", () => {
    const addon = {
      id: "a1",
      name: "db1",
      spec: {
        version: { major: 16 },
        instances: { count: 1 },
        storage: { size: "20Gi", storage_class: "standard" },
        resources: {},
        configuration: {},
        initialization: { type: "new" },
        backup: { enabled: true, object_store_id: "store-1", schedule: "0 0 2 * * *", wal_archiving: true },
      },
    } as unknown as PostgresAddon;
    const v = addonToFormValues(addon);
    expect(v.backup).toEqual({
      enabled: true,
      objectStoreId: "store-1",
      schedule: "0 0 2 * * *",
      walArchiving: true,
    });
    // And it must NOT leak into the advanced JSON area.
    expect(v.advancedJson.includes("backup")).toBe(false);
  });
});
