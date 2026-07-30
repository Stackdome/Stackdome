import { describe, it, expect } from "vitest";
import {
  addonToFormValues,
  buildCreateInput,
  deepMerge,
  parseAdvancedJson,
  JsonAreaParseError,
} from "../payload";
import {
  defaultFormValues,
  type PostgresAddonFormValues,
} from "../../schemas/form-schema";
import type { PostgresAddon } from "@/api/addons";

const baseValues = (overrides: Partial<PostgresAddonFormValues> = {}): PostgresAddonFormValues => ({
  ...defaultFormValues("cluster-1"),
  name: "my-pg",
  ...overrides,
});

describe("deepMerge", () => {
  it("override wins on scalar conflicts", () => {
    expect(deepMerge({ a: 1 }, { a: 2 })).toEqual({ a: 2 });
  });

  it("recurses into nested objects", () => {
    expect(deepMerge({ a: { b: 1, c: 1 } }, { a: { b: 2 } })).toEqual({ a: { b: 2, c: 1 } });
  });

  it("replaces arrays wholesale", () => {
    expect(deepMerge({ arr: [1, 2] }, { arr: [9] })).toEqual({ arr: [9] });
  });

  it("ignores undefined override values", () => {
    expect(deepMerge({ a: 1 }, { a: undefined })).toEqual({ a: 1 });
  });
});

describe("parseAdvancedJson", () => {
  it("returns empty object for blank input", () => {
    expect(parseAdvancedJson("")).toEqual({});
    expect(parseAdvancedJson("   \n  ")).toEqual({});
  });

  it("parses valid JSON", () => {
    expect(parseAdvancedJson('{"a":1}')).toEqual({ a: 1 });
  });

  it("throws JsonAreaParseError on invalid JSON", () => {
    expect(() => parseAdvancedJson("{not json")).toThrow(JsonAreaParseError);
  });

  it("throws when top-level is not an object", () => {
    expect(() => parseAdvancedJson("[1,2,3]")).toThrow(JsonAreaParseError);
  });
});

describe("buildCreateInput", () => {
  it("builds a minimal payload from defaults + name", () => {
    const input = buildCreateInput(baseValues({ advancedJson: "" }));

    expect(input.name).toBe("my-pg");
    // cluster_id is readonly on the API; not sent on create.
    expect(input.cluster_id).toBeUndefined();
    expect(input.spec.version.major).toBe(17);
    expect(input.spec.version.enable_auto_minor_upgrade).toBe(true);
    expect(input.spec.version.enable_auto_major_upgrade).toBe(false);
    expect(input.spec.instances.count).toBe(1);
    expect(input.spec.storage.size).toBe("10Gi");
    expect(input.spec.resources?.cpu).toEqual({ request: "250m", limit: "500m" });
    expect(input.spec.resources?.memory).toEqual({ request: "1Gi", limit: "1Gi" });
    expect(input.spec.initialization?.type).toBe("new");
    expect(input.spec.configuration?.enable_superuser_access).toBe(false);
  });

  it("HA toggle bumps replica count to 2", () => {
    const input = buildCreateInput(baseValues({ highAvailability: true, advancedJson: "" }));
    expect(input.spec.instances.count).toBe(2);
  });

  it("replicas override beats HA toggle", () => {
    const input = buildCreateInput(
      baseValues({ highAvailability: true, replicasOverride: 5, advancedJson: "" }),
    );
    expect(input.spec.instances.count).toBe(5);
  });

  it("plan changes resource preset", () => {
    const input = buildCreateInput(baseValues({ plan: "launch", advancedJson: "" }));
    expect(input.spec.resources?.cpu).toEqual({ request: "1", limit: "2" });
    expect(input.spec.resources?.memory?.request).toBe("8Gi");
  });

  it("custom plan emits raw resource fields", () => {
    const input = buildCreateInput(
      baseValues({
        plan: "custom",
        customCpuRequest: "750m",
        customCpuLimit: "1500m",
        customMemoryRequest: "3Gi",
        customMemoryLimit: "6Gi",
        advancedJson: "",
      }),
    );
    expect(input.spec.resources).toEqual({
      cpu: { request: "750m", limit: "1500m" },
      memory: { request: "3Gi", limit: "6Gi" },
    });
  });

  it("storageGB is mapped to Gi suffix", () => {
    const input = buildCreateInput(baseValues({ storageGB: 100, advancedJson: "" }));
    expect(input.spec.storage.size).toBe("100Gi");
  });

  it("databases with extensions land in spec", () => {
    const input = buildCreateInput(
      baseValues({
        databases: [
          { name: "appdb", extensions: ["vector"] },
          { name: "logs", extensions: [] },
        ],
        advancedJson: "",
      }),
    );
    expect(input.spec.databases).toEqual([
      { name: "appdb", extensions: ["vector"] },
      { name: "logs" },
    ]);
  });

  it("restore_from_backup propagates backup_id", () => {
    const input = buildCreateInput(
      baseValues({
        initialization: { type: "restore_from_backup", backupId: "bk-123" },
        advancedJson: "",
      }),
    );
    expect(input.spec.initialization).toEqual({
      type: "restore_from_backup",
      restore_from_backup: { backup_id: "bk-123" },
    });
  });

  it("merges advanced JSON spec sub-objects when no conflict", () => {
    const advancedJson = JSON.stringify({
      configuration: { parameters: { max_connections: "200" } },
      instances: { placement: { topology_key: "topology.kubernetes.io/zone" } },
    });
    const input = buildCreateInput(baseValues({ advancedJson }));
    expect(input.spec.configuration?.parameters).toEqual({ max_connections: "200" });
    // Form-driven fields preserved
    expect(input.spec.configuration?.enable_superuser_access).toBe(false);
    // Advanced-only field preserved
    const placement = (input.spec.instances as unknown as { placement?: { topology_key?: string } }).placement;
    expect(placement?.topology_key).toBe("topology.kubernetes.io/zone");
    // Form's instance count survives
    expect(input.spec.instances.count).toBe(1);
  });

  it("form values win on conflict with advanced JSON; JSON storage_class wins", () => {
    const advancedJson = JSON.stringify({
      storage: { size: "999Gi", storage_class: "io2" },
      instances: { count: 99 },
    });
    const input = buildCreateInput(
      baseValues({ storageGB: 25, highAvailability: true, advancedJson }),
    );
    // Form wins for size and count
    expect(input.spec.storage.size).toBe("25Gi");
    expect(input.spec.instances.count).toBe(2);
    // storage_class isn't form-driven, so the JSON value passes through
    expect(input.spec.storage.storage_class).toBe("io2");
  });

  it("omits storage_class when not set in JSON", () => {
    const input = buildCreateInput(baseValues({ advancedJson: "" }));
    expect(input.spec.storage.storage_class).toBeUndefined();
  });

  it("omits storage_class when the JSON value is blank", () => {
    const advancedJson = JSON.stringify({ storage: { storage_class: "" } });
    const input = buildCreateInput(baseValues({ advancedJson }));
    expect(input.spec.storage.storage_class).toBeUndefined();
  });

  it("labels and annotations from advanced JSON appear at top level", () => {
    const advancedJson = JSON.stringify({
      labels: [{ key: "env", value: "prod" }],
      annotations: [{ key: "project", value: "platform" }],
    });
    const input = buildCreateInput(baseValues({ advancedJson }));
    expect(input.labels).toEqual([{ key: "env", value: "prod" }]);
    expect(input.annotations).toEqual([{ key: "project", value: "platform" }]);
  });

  it("empty labels/annotations arrays are omitted", () => {
    const advancedJson = JSON.stringify({ labels: [], annotations: [] });
    const input = buildCreateInput(baseValues({ advancedJson }));
    expect(input.labels).toBeUndefined();
    expect(input.annotations).toBeUndefined();
  });

  it("invalid advanced JSON propagates as JsonAreaParseError", () => {
    expect(() =>
      buildCreateInput(baseValues({ advancedJson: "{not valid" })),
    ).toThrow(JsonAreaParseError);
  });
});

describe("addonToFormValues", () => {
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

  it("hydrates basic fields from minimal addon", () => {
    const v = addonToFormValues(minimalAddon());
    expect(v.name).toBe("my-pg");
    expect(v.clusterId).toBe("cluster-1");
    expect(v.versionMajor).toBe(17);
    expect(v.storageGB).toBe(10);
    expect(v.highAvailability).toBe(false);
    expect(v.replicasOverride).toBeUndefined();
    expect(v.plan).toBe("basic");
    expect(v.initialization).toEqual({ type: "new" });
  });

  it("HA derives from instances.count", () => {
    expect(addonToFormValues(minimalAddon({
      spec: {
        ...minimalAddon().spec,
        instances: { count: 2 },
      },
    })).highAvailability).toBe(true);

    const tri = addonToFormValues(minimalAddon({
      spec: {
        ...minimalAddon().spec,
        instances: { count: 3 },
      },
    }));
    expect(tri.highAvailability).toBe(true);
    expect(tri.replicasOverride).toBe(3);
  });

  it("detects matching plan preset", () => {
    const v = addonToFormValues(minimalAddon({
      spec: {
        ...minimalAddon().spec,
        resources: {
          cpu: { request: "1", limit: "2" },
          memory: { request: "8Gi", limit: "8Gi" },
        },
      },
    }));
    expect(v.plan).toBe("launch");
    expect(v.customCpuRequest).toBeUndefined();
  });

  it("falls back to custom when resources don't match a preset", () => {
    const v = addonToFormValues(minimalAddon({
      spec: {
        ...minimalAddon().spec,
        resources: {
          cpu: { request: "750m", limit: "1500m" },
          memory: { request: "3Gi", limit: "6Gi" },
        },
      },
    }));
    expect(v.plan).toBe("custom");
    expect(v.customCpuRequest).toBe("750m");
    expect(v.customCpuLimit).toBe("1500m");
    expect(v.customMemoryRequest).toBe("3Gi");
    expect(v.customMemoryLimit).toBe("6Gi");
  });

  it("parses storage size in Gi and Ti", () => {
    expect(addonToFormValues(minimalAddon({
      spec: { ...minimalAddon().spec, storage: { size: "100Gi", storage_class: "" } },
    })).storageGB).toBe(100);

    expect(addonToFormValues(minimalAddon({
      spec: { ...minimalAddon().spec, storage: { size: "2Ti", storage_class: "" } },
    })).storageGB).toBe(2048);
  });

  it("preserves restore_from_backup initialization", () => {
    const v = addonToFormValues(minimalAddon({
      spec: {
        ...minimalAddon().spec,
        initialization: {
          type: "restore_from_backup",
          restore_from_backup: { backup_id: "bk-99" },
        },
      },
    }));
    expect(v.initialization).toEqual({ type: "restore_from_backup", backupId: "bk-99", sourceAddonId: "", objectStoreId: undefined });
  });

  it("dumps placement / parameters / labels into advancedJson", () => {
    const v = addonToFormValues(minimalAddon({
      labels: [{ key: "env", value: "prod" }],
      spec: {
        ...minimalAddon().spec,
        instances: {
          count: 1,
          placement: { topology_key: "kubernetes.io/hostname", policy: "preferred" },
        },
        configuration: {
          enable_superuser_access: false,
          parameters: { max_connections: "200" },
        },
      },
    }));
    const parsed = JSON.parse(v.advancedJson);
    expect(parsed.instances.placement.policy).toBe("preferred");
    expect(parsed.configuration.parameters.max_connections).toBe("200");
    expect(parsed.labels).toEqual([{ key: "env", value: "prod" }]);
    // Stored storage_class is server-resolved and truthful, so it surfaces too.
    expect(parsed.storage.storage_class).toBe("standard");
  });

  it("hydrates as empty when addon has no meaningful overrides", () => {
    const v = addonToFormValues(minimalAddon({
      spec: { ...minimalAddon().spec, storage: { size: "10Gi" } },
    }));
    expect(v.advancedJson).toBe("");
  });

  it("hydrates storage_class into JSON when set", () => {
    const v = addonToFormValues(minimalAddon({
      spec: {
        ...minimalAddon().spec,
        storage: { size: "10Gi", storage_class: "io2" },
      },
    }));
    const parsed = JSON.parse(v.advancedJson);
    expect(parsed.storage.storage_class).toBe("io2");
    // Other empty groups still omitted.
    expect(parsed.instances).toBeUndefined();
    expect(parsed.configuration).toBeUndefined();
  });

  it("round-trips form → API → form for a typical create", () => {
    const original = baseValues({
      name: "round-trip",
      plan: "launch",
      storageGB: 50,
      versionMajor: 16,
      highAvailability: true,
      advancedJson: "",
    });
    const apiInput = buildCreateInput(original);
    // Simulate server response: assume server echoes the spec back
    const serverAddon: PostgresAddon = {
      name: apiInput.name,
      cluster_id: "cluster-1",
      spec: apiInput.spec,
      labels: apiInput.labels,
      annotations: apiInput.annotations,
    };
    const rehydrated = addonToFormValues(serverAddon);
    expect(rehydrated.name).toBe("round-trip");
    expect(rehydrated.plan).toBe("launch");
    expect(rehydrated.storageGB).toBe(50);
    expect(rehydrated.versionMajor).toBe(16);
    expect(rehydrated.highAvailability).toBe(true);
  });
});
