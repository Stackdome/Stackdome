import { describe, it, expect } from "vitest";
import { alignBaselineToDraft, cloneJson, dirtyTabsForResource, getAddonLinkCount, isPathDirty, isResourceDirty, revertResource } from "../stack-diff";
import type { ResourceArr } from "../stack-diff";

describe("cloneJson", () => {
  it("passes undefined through (JSON.parse(JSON.stringify(undefined)) would throw)", () => {
    // Regression: the field-reset arrow used to crash the page with
    // `"undefined" is not valid JSON` when the path being reverted was
    // absent from baseline (e.g., resetting init_spec.command on a resource
    // that never had an init spec).
    expect(cloneJson(undefined)).toBe(undefined);
  });

  it("deep-clones plain objects, breaking referential identity", () => {
    const src = { a: 1, nested: { b: 2 } };
    const out = cloneJson(src);
    expect(out).toEqual(src);
    expect(out).not.toBe(src);
    expect(out.nested).not.toBe(src.nested);
  });

  it("deep-clones arrays", () => {
    const src = [1, [2, 3]];
    const out = cloneJson(src);
    expect(out).toEqual(src);
    expect(out).not.toBe(src);
    expect(out[1]).not.toBe(src[1]);
  });

  it("returns primitives as-is", () => {
    expect(cloneJson(42)).toBe(42);
    expect(cloneJson("hi")).toBe("hi");
    expect(cloneJson(null)).toBe(null);
    expect(cloneJson(false)).toBe(false);
  });
});

describe("isPathDirty — structurally-empty equivalence", () => {
  // Regression: clearing a comma-separated input (e.g., Init Command) leaves
  // the form with `{ init_spec: { command: [] } }` while baseline has
  // `init_spec: undefined`. These should diff as equal so the field stops
  // reading as dirty after the user reverts.

  it("treats empty array vs undefined as not-dirty", () => {
    expect(isPathDirty(
      { init_spec: { command: [] } },
      { init_spec: undefined },
      "init_spec.command",
    )).toBe(false);
  });

  it("treats empty object vs undefined as not-dirty", () => {
    expect(isPathDirty(
      { init_spec: {} },
      { init_spec: undefined },
      "init_spec",
    )).toBe(false);
  });

  it("treats empty string vs undefined as not-dirty", () => {
    expect(isPathDirty(
      { name: "" },
      { name: undefined },
      "name",
    )).toBe(false);
  });

  it("treats deeply-nested all-empty as not-dirty against undefined", () => {
    expect(isPathDirty(
      { init_spec: { command: [], args: [], image_spec: { image: "" } } },
      { init_spec: undefined },
      "init_spec",
    )).toBe(false);
  });

  it("flags as dirty when one side has actual content", () => {
    expect(isPathDirty(
      { init_spec: { command: ["sh"] } },
      { init_spec: undefined },
      "init_spec.command",
    )).toBe(true);
  });

  it("flags as dirty when value differs from a non-empty baseline", () => {
    expect(isPathDirty(
      { name: "" },
      { name: "redis" },
      "name",
    )).toBe(true);
    expect(isPathDirty(
      { ports: [] },
      { ports: [80] },
      "ports",
    )).toBe(true);
  });

  it("returns false when values are deep-equal", () => {
    expect(isPathDirty(
      { ports: [{ number: 80 }] },
      { ports: [{ number: 80 }] },
      "ports",
    )).toBe(false);
  });
});

describe("getAddonLinkCount", () => {
  // Env vars no longer carry addon-backed sources, so addon links come solely
  // from the explicit "addons in stack" panel. This helper just counts the
  // distinct panel-linked ids; resources never contribute.

  it("returns 0 with no linked addons and no resources", () => {
    expect(getAddonLinkCount(new Set(), [])).toBe(0);
  });

  it("counts addons explicitly linked via the panel", () => {
    expect(getAddonLinkCount(new Set(["pg-1", "pg-2"]), [])).toBe(2);
  });

  it("does not derive addon links from resource env vars", () => {
    const resources: ResourceArr = [
      {
        execution_config: {
          environment_variables: [
            { from: "stack", name: "FOO", value: "bar" },
          ],
        } as never,
      },
    ];
    expect(getAddonLinkCount(new Set(), resources)).toBe(0);
  });

  it("counts only panel-linked ids regardless of resources", () => {
    const resources: ResourceArr = [
      {
        execution_config: {
          environment_variables: [
            { from: "stack", name: "DATABASE_URL", value: "x" },
          ],
        } as never,
      },
    ];
    expect(getAddonLinkCount(new Set(["pg-1", "pg-2"]), resources)).toBe(2);
  });

  it("handles resources with no execution_config", () => {
    const resources: ResourceArr = [{ name: "svc-1" }, { name: "svc-2" }];
    expect(getAddonLinkCount(new Set(["pg-1"]), resources)).toBe(1);
  });
});

describe("alignBaselineToDraft", () => {
  it("reorders the baseline to the draft's name order", () => {
    const baseline = [{ name: "redis" }, { name: "web" }, { name: "mail" }];
    const draft = [{ name: "web" }, { name: "redis" }, { name: "mail" }];
    expect(alignBaselineToDraft(baseline, draft).map((r) => r?.name)).toEqual(["web", "redis", "mail"]);
  });

  it("leaves holes for draft-only entries and appends baseline-only ones", () => {
    const baseline = [{ name: "web" }, { name: "gone" }];
    const draft = [{ name: "web" }, { name: "brand-new" }];
    const aligned = alignBaselineToDraft(baseline, draft);
    expect(aligned[0]?.name).toBe("web");
    expect(aligned[1]).toBeUndefined();
    // deleted-from-draft baseline entry survives past the draft length so
    // positional diffing still flags the deletion
    expect(aligned[2]?.name).toBe("gone");
  });

  it("is the identity when baseline and draft share order", () => {
    const baseline = [{ name: "a" }, { name: "b" }];
    expect(alignBaselineToDraft(baseline, baseline).map((r) => r?.name)).toEqual(["a", "b"]);
  });
});

describe("status is server telemetry, never dirt", () => {
  const deployed = { name: "web", source: { image: { ref: "nginx:1" } }, status: { state: "Pending" } };
  const live = { name: "web", source: { image: { ref: "nginx:1" } }, status: { state: "Ready", public_ingress: [{}] } };

  it("isResourceDirty ignores status drift", () => {
    expect(isResourceDirty(live as never, deployed as never)).toBe(false);
  });

  it("dirtyTabsForResource does not light any tab for status drift", () => {
    const tabs = dirtyTabsForResource(live as never, deployed as never);
    expect(tabs).toEqual({ configuration: false, deployment: false, environment: false });
  });

  it("revertResource keeps the draft's live status", () => {
    const draft = { resources: [{ ...live, source: { image: { ref: "nginx:2" } } }], volumes: [] };
    const baseline = { resources: [deployed], volumes: [] };
    const next = revertResource(draft as never, baseline as never, 0);
    const r = next.resources[0] as typeof live;
    expect(r.source.image.ref).toBe("nginx:1");
    expect(r.status).toEqual(live.status);
  });
});

describe("revertResource dangling-mount guard", () => {
  it("drops restored mounts whose volume no longer exists in the draft", () => {
    const baseline = {
      resources: [
        { name: "web", volume_mounts: [{ source_volume_name: "data", source_sub_path: "", target_path: "/data" }] },
      ],
      volumes: [{ name: "data" }],
    };
    // Draft deleted the volume (cascade already removed the mount) and edited the resource.
    const draft = {
      resources: [{ name: "web", source: { image: { ref: "nginx:2" } }, volume_mounts: [] }],
      volumes: [],
    };
    const next = revertResource(draft as never, baseline as never, 0);
    expect(next.resources[0].volume_mounts).toEqual([]);
  });

  it("keeps restored mounts whose volume still exists", () => {
    const baseline = {
      resources: [
        { name: "web", volume_mounts: [{ source_volume_name: "data", source_sub_path: "", target_path: "/data" }] },
      ],
      volumes: [{ name: "data" }],
    };
    const draft = {
      resources: [{ name: "web", volume_mounts: [] }],
      volumes: [{ name: "data" }],
    };
    const next = revertResource(draft as never, baseline as never, 0);
    expect(next.resources[0].volume_mounts).toEqual([
      { source_volume_name: "data", source_sub_path: "", target_path: "/data" },
    ]);
  });
});
