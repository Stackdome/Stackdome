// @vitest-environment node
import { describe, it, expect } from "vitest";
import { buildCreateInput, addonToFormValues } from "../payload";
import { defaultFormValues } from "../../schemas/form-schema";
import type { PostgresAddon } from "@/api/addons";

const base = { ...defaultFormValues("cluster-1"), name: "db1" };

const minimalAddon = (overrides?: Partial<PostgresAddon>): PostgresAddon => ({
  name: "my-pg",
  cluster_id: "cluster-1",
  spec: {
    version: {
      major: 17,
      enable_auto_minor_upgrade: true,
      enable_auto_major_upgrade: false,
    },
    instances: { count: 1 },
    storage: { size: "10Gi", storage_class: "standard" },
  },
  ...overrides,
});

describe("payload PITR mapping", () => {
  it("latest restore → no recovery_target_time", () => {
    const input = buildCreateInput({
      ...base,
      initialization: { type: "restore_from_object_store", sourceAddonId: "src", objectStoreId: "os1" },
    });
    expect(input.spec.initialization).toEqual({
      type: "restore_from_object_store",
      restore_from_object_store: { object_store_id: "os1", source_postgres_addon_id: "src" },
    });
  });

  it("specific time → recovery_target_time set", () => {
    const input = buildCreateInput({
      ...base,
      initialization: { type: "restore_from_object_store", sourceAddonId: "src", objectStoreId: "os1", recoveryTargetTime: "2026-05-17T03:00:00Z" },
    });
    expect(input.spec.initialization).toEqual({
      type: "restore_from_object_store",
      restore_from_object_store: {
        object_store_id: "os1",
        source_postgres_addon_id: "src",
        recovery_target_time: "2026-05-17T03:00:00Z",
      },
    });
  });

  it("new initialization unchanged", () => {
    const input = buildCreateInput({ ...base });
    expect(input.spec.initialization).toEqual({ type: "new" });
  });
});

describe("addonToFormValues PITR hydration", () => {
  it("hydrates restore_from_object_store without recovery_target_time", () => {
    const v = addonToFormValues(minimalAddon({
      spec: {
        ...minimalAddon().spec,
        initialization: {
          type: "restore_from_object_store",
          restore_from_object_store: { object_store_id: "os1", source_postgres_addon_id: "src" },
        },
      },
    }));
    expect(v.initialization).toEqual({
      type: "restore_from_object_store",
      sourceAddonId: "src",
      objectStoreId: "os1",
      recoveryTargetTime: undefined,
    });
  });

  it("hydrates restore_from_object_store with recovery_target_time", () => {
    const v = addonToFormValues(minimalAddon({
      spec: {
        ...minimalAddon().spec,
        initialization: {
          type: "restore_from_object_store",
          restore_from_object_store: {
            object_store_id: "os1",
            source_postgres_addon_id: "src",
            recovery_target_time: "2026-05-17T03:00:00Z",
          },
        },
      },
    }));
    expect(v.initialization).toEqual({
      type: "restore_from_object_store",
      sourceAddonId: "src",
      objectStoreId: "os1",
      recoveryTargetTime: "2026-05-17T03:00:00Z",
    });
  });
});

describe("payload restore_from_backup mapping", () => {
  it("build emits only backup_id (UI fields dropped)", () => {
    const input = buildCreateInput({
      ...base,
      initialization: { type: "restore_from_backup", backupId: "bk1", sourceAddonId: "src", objectStoreId: "os1" },
    });
    expect(input.spec.initialization).toEqual({
      type: "restore_from_backup",
      restore_from_backup: { backup_id: "bk1" },
    });
  });

  it("hydrate maps backup_id and fills placeholder UI fields", () => {
    const v = addonToFormValues(
      minimalAddon({
        spec: {
          ...minimalAddon().spec,
          initialization: { type: "restore_from_backup", restore_from_backup: { backup_id: "bk9" } },
        },
      }),
    );
    expect(v.initialization).toEqual({
      type: "restore_from_backup",
      backupId: "bk9",
      sourceAddonId: "",
      objectStoreId: undefined,
    });
  });
});

describe("buildCreateInput edit omits initialization", () => {
  it("create (default) keeps initialization", () => {
    const input = buildCreateInput({ ...base });
    expect(input.spec.initialization).toEqual({ type: "new" });
  });

  it("create with restore keeps initialization", () => {
    const input = buildCreateInput({
      ...base,
      initialization: { type: "restore_from_object_store", sourceAddonId: "s", objectStoreId: "o" },
    });
    expect(input.spec.initialization).toEqual({
      type: "restore_from_object_store",
      restore_from_object_store: { object_store_id: "o", source_postgres_addon_id: "s" },
    });
  });

  it("edit strips initialization entirely", () => {
    const input = buildCreateInput(
      { ...base, initialization: { type: "restore_from_object_store", sourceAddonId: "s", objectStoreId: "o" } },
      { isEdit: true },
    );
    expect(input.spec.initialization).toBeUndefined();
  });
});
