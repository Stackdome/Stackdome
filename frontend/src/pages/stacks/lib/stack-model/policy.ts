/**
 * What does not count as a change, as data.
 *
 * Every comparison in this module reads these rules, so a rule added here
 * applies to all of them at once.
 */

/** Written by the server, never by a user. Dropped before any comparison. */
export const SERVER_WRITTEN_RESOURCE_FIELDS = [
  "id",
  "stack_id",
  "revision",
  "outputs",
  "status",
] as const;

export const SERVER_WRITTEN_VOLUME_FIELDS = ["id", "project_id", "stack_id", "status"] as const;

/** Values the API fills in on write, so a spec that omits one and a spec that
 *  spells it out are the same config and must compare equal. */
export const DEFAULT_DOCKERFILE_PATH = "Dockerfile";
export const DEFAULT_BUILD_CONTEXT = ".";
export const DEFAULT_WORKLOAD_TYPE = "Service";

/**
 * Git revision keys the deploy-time pin resolver writes into a release snapshot.
 * A key the spec leaves unpinned carries a resolver fact on the baseline, not
 * user intent: a branch-tracking resource must not read as changed against the
 * commit that branch happened to resolve to. A key the spec DOES pin compares
 * strictly.
 */
export const REVISION_KEYS = ["branch", "tag", "commit"] as const;
export type RevisionKey = (typeof REVISION_KEYS)[number];

/** Which drawer tab a field belongs to — drives both the tab dirty-dots and the
 *  section grouping in the release diff cards. */
export type FieldSection = "configuration" | "deployment" | "environment";

const DEPLOYMENT_FIELDS = new Set(["init_spec", "lifecycle_config", "replicas", "schedule"]);

export function sectionForField(path: string): FieldSection {
  const head = path.split(".")[0];
  if (head === "env") return "environment";
  if (DEPLOYMENT_FIELDS.has(head)) return "deployment";
  return "configuration";
}

/**
 * Human labels for the fields worth naming in a diff card. Anything unlabelled
 * still surfaces — under its own path — so a new field is never silently
 * invisible, it is merely unpolished.
 */
export const FIELD_LABELS: Record<string, string> = {
  "source.image.ref": "image",
  "source.git.repo_url": "repo",
  "source.git.branch": "branch",
  "source.git.tag": "tag",
  "source.git.commit": "commit",
  "source.git.dockerfile_path": "dockerfile",
  "source.git.build_context": "build context",
  ports: "ports",
  "execution_config.command": "command",
  "execution_config.args": "args",
  depends_on: "depends on",
  workload_type: "workload type",
  replicas: "replicas",
  schedule: "schedule",
  mounts: "volume mounts",
  init_spec: "init",
  lifecycle_config: "lifecycle",
  "spec.size": "size",
  "spec.access_mode": "access mode",
  "spec.storage_class": "storage class",
  "spec.source": "source",
  "spec.needs_sync_before_use": "sync before use",
};

export function labelForField(path: string): string {
  if (FIELD_LABELS[path]) return FIELD_LABELS[path];
  // Per-row paths name the row: the env var, or the container path mounted.
  if (path.startsWith("env.")) return path.slice("env.".length);
  if (path.startsWith("mounts.")) return `mount ${path.slice("mounts.".length)}`;
  return path;
}
