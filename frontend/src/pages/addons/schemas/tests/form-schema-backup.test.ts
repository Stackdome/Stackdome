// @vitest-environment node
import { describe, it, expect } from "vitest";
import { PostgresAddonFormSchema, defaultFormValues } from "../form-schema";

describe("form-schema backup", () => {
  it("defaults backup to disabled with a sane schedule", () => {
    const v = defaultFormValues("");
    expect(v.backup).toEqual({
      enabled: false,
      objectStoreId: "",
      schedule: "0 0 3 * * *",
      walArchiving: false,
    });
  });

  it("accepts a valid disabled backup", () => {
    const v = defaultFormValues("");
    v.name = "db1";
    expect(PostgresAddonFormSchema.safeParse(v).success).toBe(true);
  });

  it("rejects enabled backup with no object store, error on backup.objectStoreId", () => {
    const v = defaultFormValues("");
    v.name = "db1";
    v.backup = { enabled: true, objectStoreId: "", schedule: "0 0 3 * * *", walArchiving: false };
    const r = PostgresAddonFormSchema.safeParse(v);
    expect(r.success).toBe(false);
    if (!r.success) {
      expect(r.error.issues.some((i) => i.path.join(".") === "backup.objectStoreId")).toBe(true);
    }
  });

  it("rejects a non-6-field cron", () => {
    const v = defaultFormValues("");
    v.name = "db1";
    v.backup = { enabled: true, objectStoreId: "s1", schedule: "bad", walArchiving: false };
    const r = PostgresAddonFormSchema.safeParse(v);
    expect(r.success).toBe(false);
  });
});
