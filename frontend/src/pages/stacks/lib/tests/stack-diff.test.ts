import { describe, it, expect } from "vitest";
import {
  alignBaselineToDraft,
  cloneJson,
  resourceRenameFingerprint,
  revertResource,
  revertResourceField,
} from "../stack-diff";
import { pairByFingerprint } from "../stack-model/equal";

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

  it("pairs a renamed entry with its content-identical baseline when given a fingerprint", () => {
    const baseline = [
      { name: "mysql", image_spec: { image: "mysql:8" }, status: { state: "Ready" } },
      { name: "postgres", image_spec: { image: "postgres:16" } },
    ];
    const draft = [
      { name: "mysqla", image_spec: { image: "mysql:8" }, status: { state: "Pending" } },
      { name: "postgres", image_spec: { image: "postgres:16" } },
    ];
    const aligned = alignBaselineToDraft(baseline, draft, resourceRenameFingerprint);
    // The renamed baseline slots into the draft's position instead of leaving a
    // hole + an appended deletion — so positional diffing reads ONE change.
    expect(aligned.map((r) => r?.name)).toEqual(["mysql", "postgres"]);
    expect(aligned).toHaveLength(2);
  });

  it("does not pair when content also changed (stays add + remove)", () => {
    const baseline = [{ name: "mysql", image_spec: { image: "mysql:8" } }];
    const draft = [{ name: "mysqla", image_spec: { image: "mysql:9" } }];
    const aligned = alignBaselineToDraft(baseline, draft, resourceRenameFingerprint);
    expect(aligned[0]).toBeUndefined();
    expect(aligned[1]?.name).toBe("mysql");
  });

  it("without a fingerprint keeps the old hole + append behavior", () => {
    const baseline = [{ name: "mysql", image_spec: { image: "mysql:8" } }];
    const draft = [{ name: "mysqla", image_spec: { image: "mysql:8" } }];
    const aligned = alignBaselineToDraft(baseline, draft);
    expect(aligned[0]).toBeUndefined();
    expect(aligned[1]?.name).toBe("mysql");
  });
});

describe("pairByFingerprint", () => {
  const fp = (s: { v: string }) => s.v;
  it("greedily pairs matching fingerprints, each entry used once", () => {
    const pairs = pairByFingerprint([{ v: "x" }, { v: "y" }], [{ v: "y" }, { v: "x" }, { v: "x" }], fp, fp);
    expect(pairs).toHaveLength(2);
    expect(pairs.map(([a, b]) => [a.v, b.v])).toEqual([["x", "x"], ["y", "y"]]);
  });
  it("leaves non-matching entries unpaired", () => {
    expect(pairByFingerprint([{ v: "x" }], [{ v: "z" }], fp, fp)).toEqual([]);
  });
});

describe("revert keeps live telemetry", () => {
  const deployed = { name: "web", source: { image: { ref: "nginx:1" } }, status: { state: "Pending" } };
  const live = { name: "web", source: { image: { ref: "nginx:1" } }, status: { state: "Ready", public_ingress: [{}] } };

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

/**
 * The forward rename carries every sibling to the new name. Undoing the rename
 * has to bring them back, or the discard leaves the siblings pointing at a name
 * nothing has any more.
 */
describe("reverting a rename carries the references back", () => {
  const baseline = {
    resources: [{ name: "redis" }, { name: "web", depends_on: ["redis"] }],
    volumes: [],
  };
  // The rename to "cache" already moved the sibling with it.
  const draft = {
    resources: [{ name: "cache" }, { name: "web", depends_on: ["cache"] }],
    volumes: [],
  };

  it("repoints siblings when the whole resource is reverted", () => {
    const next = revertResource(draft as never, baseline as never, 0);
    expect(next.resources[0].name).toBe("redis");
    expect(next.resources[1].depends_on).toEqual(["redis"]);
  });

  it("repoints siblings when only the name field is discarded", () => {
    const next = revertResourceField(draft as never, baseline as never, 0, "name");
    expect(next.resources[0].name).toBe("redis");
    expect(next.resources[1].depends_on).toEqual(["redis"]);
  });

  /** Reverting an added resource is a delete, so it has to prune like one. */
  it("prunes references when reverting a resource that only exists in the draft", () => {
    const emptyBaseline = { resources: [], volumes: [] };
    const withAdded = {
      resources: [{ name: "cache" }, { name: "web", depends_on: ["cache"] }],
      volumes: [],
    };
    const next = revertResource(withAdded as never, emptyBaseline as never, 0);
    expect(next.resources.map((r) => r?.name)).toEqual(["web"]);
    expect(next.resources[0].depends_on).toEqual([]);
  });

  /** The sibling's baseline predates the rename, so restoring it names a
   *  resource that no longer exists. Same shape as the dangling-mount guard. */
  it("drops a restored reference to a sibling that has since been renamed", () => {
    const next = revertResource(draft as never, baseline as never, 1);
    expect(next.resources[1].depends_on).toEqual([]);
  });

  it("keeps a restored reference when the sibling still exists", () => {
    const stillThere = {
      resources: [{ name: "redis" }, { name: "web", depends_on: [] }],
      volumes: [],
    };
    const next = revertResource(stillThere as never, baseline as never, 1);
    expect(next.resources[1].depends_on).toEqual(["redis"]);
  });

  /** A draft-only resource has no baseline name, so discarding the field
   *  clears it and leaves siblings naming something that is gone. */
  it("prunes references when the name field is discarded on a draft-only resource", () => {
    const holedBaseline = { resources: [undefined, { name: "web", depends_on: [] }], volumes: [] };
    const withAdded = {
      resources: [{ name: "cache" }, { name: "web", depends_on: ["cache"] }],
      volumes: [],
    };
    const next = revertResourceField(withAdded as never, holedBaseline as never, 0, "name");
    expect(next.resources[1].depends_on).toEqual([]);
  });

  it("leaves siblings alone when the reverted field is not the name", () => {
    const withImage = {
      resources: [{ name: "cache", source: { image: { ref: "redis:8" } } }, draft.resources[1]],
      volumes: [],
    };
    const withBaseImage = {
      resources: [{ name: "cache", source: { image: { ref: "redis:7" } } }, baseline.resources[1]],
      volumes: [],
    };
    const next = revertResourceField(withImage as never, withBaseImage as never, 0, "source.image.ref");
    expect(next.resources[1].depends_on).toEqual(["cache"]);
  });
});
