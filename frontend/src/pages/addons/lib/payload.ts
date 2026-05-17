import type { PostgresAddon, PostgresAddonCreateInput, PostgresAddonSpec } from "@/api/addons";
import type { PostgresAddonFormValues } from "../schemas/form-schema";
import { DEFAULT_ADVANCED_JSON } from "../schemas/form-schema";
import { PLAN_PRESETS, resourcesForPlan, type PlanId } from "./plan-presets";
import { toApiBackupConfig } from "../schemas/backup-config-schema";

type Json = Record<string, unknown>;

function isPlainObject(v: unknown): v is Json {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}

export function deepMerge<T extends Json>(base: T, override: Json): T {
  const out: Json = { ...base };
  for (const [key, value] of Object.entries(override)) {
    if (value === undefined) continue;
    const existing = out[key];
    if (isPlainObject(existing) && isPlainObject(value)) {
      out[key] = deepMerge(existing, value);
    } else {
      out[key] = value;
    }
  }
  return out as T;
}

export class JsonAreaParseError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "JsonAreaParseError";
  }
}

export function parseAdvancedJson(raw: string): Json {
  const trimmed = raw.trim();
  if (!trimmed) return {};
  try {
    const parsed = JSON.parse(trimmed);
    if (!isPlainObject(parsed)) {
      throw new JsonAreaParseError("Advanced configuration must be a JSON object");
    }
    return parsed;
  } catch (e) {
    if (e instanceof JsonAreaParseError) throw e;
    throw new JsonAreaParseError(
      e instanceof Error ? `Invalid JSON: ${e.message}` : "Invalid JSON",
    );
  }
}

function buildFormSpec(values: PostgresAddonFormValues): PostgresAddonSpec {
  const presetResources = resourcesForPlan(values.plan);
  const customResources = {
    ...(values.customCpuRequest || values.customCpuLimit
      ? {
        cpu: {
          ...(values.customCpuRequest ? { request: values.customCpuRequest } : {}),
          ...(values.customCpuLimit ? { limit: values.customCpuLimit } : {}),
        },
      }
      : {}),
    ...(values.customMemoryRequest || values.customMemoryLimit
      ? {
        memory: {
          ...(values.customMemoryRequest ? { request: values.customMemoryRequest } : {}),
          ...(values.customMemoryLimit ? { limit: values.customMemoryLimit } : {}),
        },
      }
      : {}),
  };

  const replicas = values.replicasOverride ?? (values.highAvailability ? 2 : 1);

  const spec: PostgresAddonSpec = {
    version: {
      major: values.versionMajor,
      enable_auto_minor_upgrade: values.autoMinorUpgrade,
      enable_auto_major_upgrade: values.autoMajorUpgrade,
    },
    instances: { count: replicas },
    storage: {
      size: `${values.storageGB}Gi`,
      // Default to "standard" — works for Kind/AKS/GKE defaults. Power users
      // override via the Advanced JSON area's storage.storage_class field.
      storage_class: "standard",
    },
    resources: presetResources ?? customResources,
    configuration: { enable_superuser_access: values.superuserAccess },
    initialization:
      values.initialization.type === "restore_from_backup"
        ? {
          type: "restore_from_backup",
          restore_from_backup: { backup_id: values.initialization.backupId },
        }
        : values.initialization.type === "restore_from_object_store"
          ? {
            type: "restore_from_object_store",
            restore_from_object_store: {
              object_store_id: values.initialization.objectStoreId,
              source_postgres_addon_id: values.initialization.sourceAddonId,
              ...(values.initialization.recoveryTargetTime
                ? { recovery_target_time: values.initialization.recoveryTargetTime }
                : {}),
            },
          }
          : { type: "new" },
  };

  if (values.backup.enabled) {
    spec.backup = toApiBackupConfig(values.backup);
  }

  if (values.databases.length > 0) {
    spec.databases = values.databases.map((db) => ({
      name: db.name,
      ...(db.extensions.length > 0 ? { extensions: db.extensions } : {}),
    }));
  }

  return spec;
}

const SPEC_KEYS = new Set([
  "version",
  "instances",
  "storage",
  "resources",
  "backup",
  "initialization",
  "databases",
  "configuration",
]);

interface SplitAdvanced {
  spec: Json;
  topLevel: Pick<PostgresAddonCreateInput, "labels" | "annotations">;
}

function splitAdvanced(advanced: Json): SplitAdvanced {
  const spec: Json = {};
  const topLevel: SplitAdvanced["topLevel"] = {};

  if (isPlainObject(advanced.spec)) {
    Object.assign(spec, advanced.spec);
  }

  for (const [key, value] of Object.entries(advanced)) {
    if (key === "spec") continue;
    if (key === "labels" && Array.isArray(value) && value.length > 0) {
      topLevel.labels = value as PostgresAddonCreateInput["labels"];
      continue;
    }
    if (key === "annotations" && Array.isArray(value) && value.length > 0) {
      topLevel.annotations = value as PostgresAddonCreateInput["annotations"];
      continue;
    }
    if (SPEC_KEYS.has(key)) {
      const existing = spec[key];
      spec[key] = isPlainObject(existing) && isPlainObject(value)
        ? deepMerge(existing, value)
        : value;
    }
  }

  return { spec, topLevel };
}

export function detectPlan(resources?: PostgresAddonSpec["resources"]): PlanId {
  const cpuReq = resources?.cpu?.request;
  const cpuLim = resources?.cpu?.limit;
  const memReq = resources?.memory?.request;
  const memLim = resources?.memory?.limit;
  if (!cpuReq && !cpuLim && !memReq && !memLim) return "basic";

  for (const preset of PLAN_PRESETS) {
    if (preset.id === "custom") continue;
    const expected = resourcesForPlan(preset.id);
    if (!expected) continue;
    if (
      expected.cpu.request === cpuReq &&
      expected.cpu.limit === cpuLim &&
      expected.memory.request === memReq &&
      expected.memory.limit === memLim
    ) {
      return preset.id as PlanId;
    }
  }
  return "custom";
}

function parseStorageGB(size?: string): number {
  if (!size) return 10;
  const m = /^(\d+)\s*Gi$/i.exec(size.trim());
  if (m) return Number(m[1]);
  const t = /^(\d+)\s*Ti$/i.exec(size.trim());
  if (t) return Number(t[1]) * 1024;
  const n = parseInt(size, 10);
  return Number.isFinite(n) && n > 0 ? n : 10;
}

// Default value the form sends for storage_class. We strip this from the
// hydrated JSON so the user doesn't see a "set" field they didn't actually set.
const DEFAULT_STORAGE_CLASS = "standard";

function buildAdvancedJson(addon: PostgresAddon): string {
  // Surface only fields the user has *meaningfully* configured. Empty objects,
  // empty arrays, defaults, and the no-op "new" initialization are omitted so
  // the JSON area genuinely reflects user-set overrides — not boilerplate.
  const obj: Record<string, unknown> = {};

  const placement = addon.spec.instances.placement;
  if (placement && Object.keys(placement).length > 0) {
    obj.instances = { placement };
  }

  const parameters = addon.spec.configuration?.parameters;
  if (parameters && Object.keys(parameters).length > 0) {
    obj.configuration = { parameters };
  }

  const storageClass = addon.spec.storage.storage_class;
  if (storageClass && storageClass !== DEFAULT_STORAGE_CLASS) {
    obj.storage = { storage_class: storageClass };
  }

  if (addon.spec.initialization && addon.spec.initialization.type !== "new") {
    obj.initialization = addon.spec.initialization;
  }

  if (addon.labels && addon.labels.length > 0) {
    obj.labels = addon.labels;
  }

  if (addon.annotations && addon.annotations.length > 0) {
    obj.annotations = addon.annotations;
  }

  if (Object.keys(obj).length === 0) return DEFAULT_ADVANCED_JSON;
  return `${JSON.stringify(obj, null, 2)}\n`;
}

export function addonToFormValues(addon: PostgresAddon): PostgresAddonFormValues {
  const plan = detectPlan(addon.spec.resources);
  const isCustom = plan === "custom";
  const replicaCount = addon.spec.instances.count;
  const init = addon.spec.initialization;

  return {
    name: addon.name,
    clusterId: addon.cluster_id ?? "",
    plan,
    customCpuRequest: isCustom ? addon.spec.resources?.cpu?.request : undefined,
    customCpuLimit: isCustom ? addon.spec.resources?.cpu?.limit : undefined,
    customMemoryRequest: isCustom ? addon.spec.resources?.memory?.request : undefined,
    customMemoryLimit: isCustom ? addon.spec.resources?.memory?.limit : undefined,
    storageGB: parseStorageGB(addon.spec.storage.size),
    versionMajor: addon.spec.version.major,
    highAvailability: replicaCount >= 2,
    replicasOverride: replicaCount > 2 ? replicaCount : undefined,
    autoMinorUpgrade: addon.spec.version.enable_auto_minor_upgrade ?? true,
    autoMajorUpgrade: addon.spec.version.enable_auto_major_upgrade ?? false,
    superuserAccess: addon.spec.configuration?.enable_superuser_access ?? false,
    databases: (addon.spec.databases ?? []).map((db) => ({
      name: db.name,
      extensions: (db.extensions ?? []) as ("vector")[],
    })),
    initialization:
      init?.type === "restore_from_backup"
        ? {
          type: "restore_from_backup",
          backupId: init.restore_from_backup?.backup_id ?? "",
        }
        : init?.type === "restore_from_object_store"
          ? {
            type: "restore_from_object_store",
            sourceAddonId: init.restore_from_object_store?.source_postgres_addon_id ?? "",
            objectStoreId: init.restore_from_object_store?.object_store_id ?? "",
            recoveryTargetTime: init.restore_from_object_store?.recovery_target_time,
          }
          : { type: "new" },
    backup: {
      // Backend stores spec.backup only when backups are enabled and drops the
      // `enabled` flag from responses, so presence of the object means enabled.
      enabled: addon.spec.backup ? addon.spec.backup.enabled ?? true : false,
      objectStoreId: addon.spec.backup?.object_store_id ?? "",
      schedule: addon.spec.backup?.schedule || "0 0 3 * * *",
      walArchiving: addon.spec.backup?.wal_archiving ?? false,
    },
    advancedJson: buildAdvancedJson(addon),
  };
}

export function buildCreateInput(values: PostgresAddonFormValues): PostgresAddonCreateInput {
  const formSpec = buildFormSpec(values);
  const { spec: advancedSpec, topLevel } = splitAdvanced(parseAdvancedJson(values.advancedJson));

  // Form-driven spec wins on conflict.
  const mergedSpec = deepMerge(
    advancedSpec,
    formSpec as unknown as Json,
  ) as unknown as PostgresAddonSpec;

  // Special case: storage_class isn't surfaced in the form, so any value the
  // user typed in the JSON area should win over the form's "standard" fallback.
  const advancedStorage = (advancedSpec.storage as Json | undefined);
  const advancedStorageClass = typeof advancedStorage?.storage_class === "string"
    ? advancedStorage.storage_class
    : undefined;
  if (advancedStorageClass) {
    mergedSpec.storage = { ...mergedSpec.storage, storage_class: advancedStorageClass };
  }

  // cluster_id is readonly on the API — backend auto-picks the org's cluster.
  // Once the backend accepts it on create, restore: cluster_id: values.clusterId.
  void values.clusterId;

  return {
    name: values.name,
    spec: mergedSpec,
    ...topLevel,
  };
}
