/**
 * The single word→variant brain. Every status string the backend can emit is
 * mapped here, categorized by resource domain, each case listing that
 * resource's complete closed vocabulary (source file noted per case).
 *
 * Rules: unknown non-empty word → "info" (visibly unrecognized, never a
 * silent green); empty/missing → "neutral". Matching is trimmed and
 * case-insensitive. Do NOT add regexes or share word lists across domains —
 * one case per resource is the point.
 */

export type StatusVariant = "ready" | "pending" | "error" | "info" | "neutral";

/**
 * Display word per variant — cards, filter pills, and any other status
 * surface must speak the same language, so derive labels from here rather
 * than hardcoding words per component.
 */
export const statusVariantLabel: Record<StatusVariant, string> = {
  ready: "ready",
  pending: "pending",
  error: "failed",
  info: "unknown",
  neutral: "unknown",
};

/**
 * Visual tone per variant for the "Status Strip" cards: `deploying` renders
 * the animated in-flight rail; settled tones color the status word only.
 */
export type StatusTone = "success" | "brand" | "danger" | "deploying";

export const statusVariantTone: Record<StatusVariant, StatusTone> = {
  ready: "success",
  pending: "deploying",
  error: "danger",
  info: "brand",
  neutral: "brand",
};

/**
 * Preview envs with no reported phase are still provisioning — treat a
 * missing phase as pending, not neutral. All preview status surfaces (cards,
 * filter buckets) must go through this so they agree.
 */
export function previewStatusVariant(phase?: string | null): StatusVariant {
  if (!(phase ?? "").trim()) return "pending";
  return statusVariant("preview", phase);
}

export type StatusDomain =
  | "stack"
  | "resource"
  | "release"
  | "rollout"
  | "volume"
  | "addon"
  | "registry"
  | "storage"
  | "build"
  | "preview"
  | "generic";

export function statusVariant(domain: StatusDomain, state?: string | null): StatusVariant {
  const s = (state ?? "").trim().toLowerCase();
  if (!s) return "neutral";

  switch (domain) {
    // pkg/models/stack.go:75 (Deleting/Error declared but not emitted today — mapped anyway)
    case "stack":
      switch (s) {
        case "pending":
        case "progressing":
        case "deleting":
          return "pending";
        case "ready":
          return "ready";
        case "failed":
        case "degraded":
        case "error":
          return "error";
        default:
          return "info";
      }

    // pkg/models/stack_resource.go:61
    case "resource":
      switch (s) {
        case "pending":
          return "pending";
        case "ready":
          return "ready";
        case "failed":
        case "error":
          return "error";
        default:
          return "info";
      }

    // pkg/models/stack_release.go:14 — the one real OpenAPI enum
    case "release":
      switch (s) {
        case "pending":
        case "inprogress":
          return "pending";
        case "released":
          return "ready";
        case "failed":
          return "error";
        case "superseded":
        case "cancelled":
          return "neutral";
        default:
          return "info";
      }

    // cluster-agent api/core/v1alpha1/stack_resource_types.go:21 — exactly 4 words
    case "rollout":
      switch (s) {
        case "pending":
          return "pending";
        case "ready":
          return "ready";
        case "degraded":
        case "failed":
          return "error";
        default:
          return "info";
      }

    // cluster-agent api/storage/v1alpha1/volume_types.go:15 — exactly 2 words
    case "volume":
      switch (s) {
        case "pending":
          return "pending";
        case "ready":
          return "ready";
        default:
          return "info";
      }

    // pkg/models/postgres_addon.go:21-31
    case "addon":
      switch (s) {
        case "pending":
        case "creating":
        case "initializing":
        case "updating":
        case "backing up":
        case "restoring":
        case "deleting":
          return "pending";
        case "ready":
          return "ready";
        case "error":
          return "error";
        case "hibernated":
        case "fenced":
          return "neutral";
        default:
          return "info";
      }

    // pkg/models/cluster_image_registry.go:15-17
    case "registry":
      switch (s) {
        case "pending":
          return "pending";
        case "running":
          return "ready";
        case "error":
          return "error";
        default:
          return "info";
      }

    // pkg/models/stack_storage.go:17-22 (not rendered yet; mapped for when it is)
    case "storage":
      switch (s) {
        case "pending":
        case "creating":
        case "deleting":
          return "pending";
        case "ready":
        case "created":
          return "ready";
        case "failed":
          return "error";
        default:
          return "info";
      }

    // cluster-agent api/builds/v1alpha1/imagebuild_types.go:38 (flows as raw string)
    case "build":
      switch (s) {
        case "pending":
          return "pending";
        case "success":
          return "ready";
        case "failed":
          return "error";
        case "cancelled":
          return "neutral";
        default:
          return "info";
      }

    // pkg/models/preview_stack.go — PreviewStackPhase vocabulary
    case "preview":
      switch (s) {
        case "provisioning":
        case "deploying":
        case "deleting":
          return "pending";
        case "ready":
          return "ready";
        case "failed":
          return "error";
        default:
          return "info";
      }

    // Legacy variantFromState behavior + the failure-detail enums
    // (openapi StackResourceFailure / ContainerFailureDetail).
    case "generic":
      switch (s) {
        case "ready":
        case "running":
        case "active":
        case "succeeded":
        case "exposed":
        case "released":
          return "ready";
        case "pending":
        case "deploying":
        case "creating":
        case "updating":
        case "provisioning":
        case "inprogress":
          return "pending";
        case "error":
        case "failed":
        case "crash":
        case "crashloopbackoff":
        case "unhealthy":
        case "runtime_crash":
        case "build_failure":
        case "crash_loop":
        case "out_of_memory":
        case "image_pull_failed":
        case "create_container_error":
        case "exit_error":
          return "error";
        case "superseded":
        case "cancelled":
          return "neutral";
        default:
          return "info";
      }
  }
}
