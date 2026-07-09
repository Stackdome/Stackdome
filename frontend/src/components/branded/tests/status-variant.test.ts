import { describe, it, expect } from "vitest";
import { statusVariant, type StatusDomain, type StatusVariant } from "../status-variant";

// The table IS the vocabulary contract. Sources:
// stack: pkg/models/stack.go:75 · resource: pkg/models/stack_resource.go:61
// release: pkg/models/stack_release.go:14 · rollout: cluster-agent stack_resource_types.go:21
// volume: cluster-agent volume_types.go:15 · addon: pkg/models/postgres_addon.go:21
// registry: pkg/models/cluster_image_registry.go:15 · storage: pkg/models/stack_storage.go:17
// build: cluster-agent imagebuild_types.go:38
const CASES: [StatusDomain, string, StatusVariant][] = [
  // stack
  ["stack", "Pending", "pending"],
  ["stack", "Progressing", "pending"],
  ["stack", "Deleting", "pending"],
  ["stack", "Ready", "ready"],
  ["stack", "Failed", "error"],
  ["stack", "Degraded", "error"],
  ["stack", "Error", "error"],
  // resource
  ["resource", "Pending", "pending"],
  ["resource", "Ready", "ready"],
  ["resource", "Failed", "error"],
  ["resource", "Error", "error"],
  // release
  ["release", "Pending", "pending"],
  ["release", "InProgress", "pending"],
  ["release", "Released", "ready"],
  ["release", "Failed", "error"],
  ["release", "Superseded", "neutral"],
  ["release", "Cancelled", "neutral"],
  // rollout
  ["rollout", "Pending", "pending"],
  ["rollout", "Ready", "ready"],
  ["rollout", "Degraded", "error"],
  ["rollout", "Failed", "error"],
  // volume
  ["volume", "Pending", "pending"],
  ["volume", "Ready", "ready"],
  // addon
  ["addon", "Pending", "pending"],
  ["addon", "Creating", "pending"],
  ["addon", "Initializing", "pending"],
  ["addon", "Updating", "pending"],
  ["addon", "Backing Up", "pending"],
  ["addon", "Restoring", "pending"],
  ["addon", "Deleting", "pending"],
  ["addon", "Ready", "ready"],
  ["addon", "Error", "error"],
  ["addon", "Hibernated", "neutral"],
  ["addon", "Fenced", "neutral"],
  // registry
  ["registry", "Pending", "pending"],
  ["registry", "Running", "ready"],
  ["registry", "Error", "error"],
  // storage
  ["storage", "Pending", "pending"],
  ["storage", "Creating", "pending"],
  ["storage", "Deleting", "pending"],
  ["storage", "Ready", "ready"],
  ["storage", "Created", "ready"],
  ["storage", "Failed", "error"],
  // build
  ["build", "Pending", "pending"],
  ["build", "Success", "ready"],
  ["build", "Failed", "error"],
  ["build", "Cancelled", "neutral"],
];

describe("statusVariant vocabulary", () => {
  it.each(CASES)("%s: %s → %s", (domain, state, expected) => {
    expect(statusVariant(domain, state)).toBe(expected);
  });

  it("is case-insensitive and trims", () => {
    expect(statusVariant("stack", "  READY ")).toBe("ready");
    expect(statusVariant("release", "inprogress")).toBe("pending");
  });

  it("unknown non-empty word → info (honest 'unrecognized')", () => {
    expect(statusVariant("volume", "running")).toBe("info");
    expect(statusVariant("stack", "Converged")).toBe("info");
  });

  it("empty/null/undefined → neutral", () => {
    expect(statusVariant("stack", "")).toBe("neutral");
    expect(statusVariant("stack", null)).toBe("neutral");
    expect(statusVariant("stack", undefined)).toBe("neutral");
  });

  it("generic keeps the legacy sets plus failure enums", () => {
    expect(statusVariant("generic", "running")).toBe("ready");
    expect(statusVariant("generic", "deploying")).toBe("pending");
    expect(statusVariant("generic", "crashloopbackoff")).toBe("error");
    expect(statusVariant("generic", "superseded")).toBe("neutral");
    expect(statusVariant("generic", "out_of_memory")).toBe("error");
    expect(statusVariant("generic", "image_pull_failed")).toBe("error");
    expect(statusVariant("generic", "mystery")).toBe("info");
  });
});
