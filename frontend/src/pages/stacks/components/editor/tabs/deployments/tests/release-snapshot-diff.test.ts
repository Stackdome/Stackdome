// @vitest-environment jsdom
import { describe, it, expect } from "vitest";
import { diffSnapshots } from "../release-snapshot-diff";

const snap = (resources: unknown[]) => ({ resources });
const mk = (o: Record<string, unknown>) => o as unknown as Parameters<typeof diffSnapshots>[0];
const web = (over: Record<string, unknown> = {}) => ({
  name: "web",
  source: { image: { ref: "web:1" } },
  ports: [{ number: 3000 }],
  execution_config: { command: ["node", "a.js"], environment_variables: [{ name: "LOG", value: "info" }] },
  ...over,
});

describe("diffSnapshots", () => {
  it("returns [] when nothing changed", () => {
    expect(diffSnapshots(snap([web()]), snap([web()])).resources).toEqual([]);
  });

  it("flags a modified image as a changed configuration row", () => {
    const out = diffSnapshots(snap([web()]), snap([web({ source: { image: { ref: "web:2" } } })])).resources;
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "web", change: "modified" });
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toContainEqual({ key: "image", from: "web:1", to: "web:2", kind: "changed" });
  });

  it("splits env changes into an environment section", () => {
    const out = diffSnapshots(
      snap([web()]),
      snap([web({ execution_config: { command: ["node", "a.js"], environment_variables: [{ name: "LOG", value: "debug" }, { name: "NEW", value: "1" }] } })]),
    ).resources;
    const env = out[0].sections.find((s) => s.kind === "environment")!;
    expect(env.rows).toContainEqual({ key: "LOG", from: "info", to: "debug", kind: "changed" });
    expect(env.rows).toContainEqual({ key: "NEW", to: "1", kind: "added" });
  });

  it("marks an added resource with all rows as added", () => {
    const out = diffSnapshots(snap([]), snap([web()])).resources;
    expect(out[0]).toMatchObject({ name: "web", change: "added" });
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows.every((r) => r.kind === "added")).toBe(true);
  });

  it("enumerates a removed resource's full config as removed rows", () => {
    const out = diffSnapshots(snap([web()]), snap([])).resources;
    expect(out[0]).toMatchObject({ name: "web", change: "removed" });
    expect(out[0].note).toMatch(/removed/i);
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toContainEqual({ key: "image", from: "web:1", kind: "removed" });
    const env = out[0].sections.find((s) => s.kind === "environment")!;
    expect(env.rows).toContainEqual({ key: "LOG", from: "info", kind: "removed" });
  });

  it("collapses a removed + added pair with identical config into one rename", () => {
    const out = diffSnapshots(snap([web()]), snap([web({ name: "web1" })])).resources;
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "web1", fromName: "web", change: "renamed" });
    expect(out[0].sections).toEqual([]);
  });

  it("does NOT pair a rename when the config also changed (stays remove + add)", () => {
    const out = diffSnapshots(snap([web()]), snap([web({ name: "web1", source: { image: { ref: "web:2" } } })])).resources;
    const changes = out.map((r) => r.change).sort();
    expect(changes).toEqual(["added", "removed"]);
  });

  it("returns [] when there is no previous snapshot", () => {
    expect(diffSnapshots(undefined, snap([web()])).resources).toEqual([]);
  });
});

describe("diffSnapshots volumes", () => {
  const vol = (name: string, size: string) => ({ name, spec: { size, access_mode: "ReadWriteOnce" } });

  it("flags an added, a removed, and a resized volume", () => {
    const prev = mk({ resources: [], volumes: [vol("data", "1Gi"), vol("cache", "500Mi")] });
    const cur = mk({ resources: [], volumes: [vol("data", "2Gi"), vol("logs", "1Gi")] });
    const out = diffSnapshots(prev, cur);
    expect(out.volumes).toEqual([
      { name: "data", change: "modified", rows: [{ key: "size", from: "1Gi", to: "2Gi", kind: "changed" }] },
      { name: "cache", change: "removed", rows: [{ key: "size", from: "500Mi", kind: "removed" }, { key: "access_mode", from: "ReadWriteOnce", kind: "removed" }], note: "Volume removed from this release." },
      { name: "logs", change: "added", rows: [{ key: "size", to: "1Gi", kind: "added" }, { key: "access_mode", to: "ReadWriteOnce", kind: "added" }] },
    ]);
  });

  it("returns no volume diff when volumes are unchanged", () => {
    const v = [vol("data", "1Gi")];
    expect(diffSnapshots(mk({ resources: [], volumes: v }), mk({ resources: [], volumes: v })).volumes).toEqual([]);
  });
});

describe("diffSnapshots connections", () => {
  const conn = (env: string, output: string) => ({
    kind: "env",
    from: { type: "addon/postgres", name: "db" },
    to: { type: "stack_resource", name: "api" },
    mappings: [{ target: { type: "env", name: env }, value: { output } }],
  });

  it("flags a changed mapping value on an existing connection", () => {
    const prev = mk({ resources: [], connections: [conn("DATABASE_URL", "url")] });
    const cur = mk({ resources: [], connections: [conn("DATABASE_URL", "public_url")] });
    expect(diffSnapshots(prev, cur).connections).toEqual([
      { name: "env · db → api", change: "modified", rows: [{ key: "DATABASE_URL", from: "url", to: "public_url", kind: "changed" }] },
    ]);
  });

  it("flags an added connection", () => {
    const out = diffSnapshots(mk({ resources: [], connections: [] }), mk({ resources: [], connections: [conn("DATABASE_URL", "url")] }));
    expect(out.connections).toEqual([
      { name: "env · db → api", change: "added", rows: [{ key: "DATABASE_URL", to: "url", kind: "added" }] },
    ]);
  });

  it("flags a disconnected volume mount as a removed connection", () => {
    const mount = {
      kind: "volume_mount",
      from: { type: "volume", name: "data" },
      to: { type: "stack_resource", name: "api" },
    };
    const out = diffSnapshots(mk({ resources: [], connections: [mount] }), mk({ resources: [], connections: [] }));
    expect(out.connections).toHaveLength(1);
    expect(out.connections[0]).toMatchObject({ name: "volume_mount · data → api", change: "removed" });
    expect(out.connections[0].note).toMatch(/removed/i);
  });

  it("does not report phantom connection changes when a resource endpoint was only renamed", () => {
    const mount = (resource: string) => ({
      kind: "volume_mount",
      from: { type: "volume", name: "mysql-data" },
      to: { type: "stack_resource", name: resource },
    });
    const prev = mk({ resources: [web({ name: "mysql" })], connections: [mount("mysql")] });
    const cur = mk({ resources: [web({ name: "mysqla" })], connections: [mount("mysqla")] });
    const out = diffSnapshots(prev, cur);
    expect(out.resources).toEqual([{ name: "mysqla", fromName: "mysql", change: "renamed", sections: [] }]);
    expect(out.connections).toEqual([]);
  });

  it("still surfaces a real mapping change on a renamed resource's connection, under the new name", () => {
    const prev = mk({
      resources: [web({ name: "api" })],
      connections: [{ ...conn("DATABASE_URL", "url"), to: { type: "stack_resource", name: "api" } }],
    });
    const cur = mk({
      resources: [web({ name: "api2" })],
      connections: [{ ...conn("DATABASE_URL", "public_url"), to: { type: "stack_resource", name: "api2" } }],
    });
    const out = diffSnapshots(prev, cur);
    expect(out.connections).toEqual([
      { name: "env · db → api2", change: "modified", rows: [{ key: "DATABASE_URL", from: "url", to: "public_url", kind: "changed" }] },
    ]);
  });

  it("leaves connections to non-resource endpoints with a colliding name untouched by the rename map", () => {
    const volMount = {
      kind: "volume_mount",
      from: { type: "volume", name: "web" }, // volume named like the renamed resource
      to: { type: "stack_resource", name: "api" },
    };
    const prev = mk({ resources: [web({ name: "web" }), web({ name: "api", image_spec: { image: "api:1" } })], connections: [volMount] });
    const cur = mk({ resources: [web({ name: "web2" }), web({ name: "api", image_spec: { image: "api:1" } })], connections: [volMount] });
    const out = diffSnapshots(prev, cur);
    expect(out.connections).toEqual([]);
  });
});

describe("diffSnapshots catch-all", () => {
  it("surfaces an unprojected resource field change as one generic row", () => {
    // init_spec is real config but outside the projected scalar set — before
    // the catch-all this diffed as "no changes" (the Bug-3 class).
    const out = diffSnapshots(
      snap([web({ init_spec: { command: ["migrate"] } })]),
      snap([web({ init_spec: { command: ["migrate", "--seed"] } })]),
    ).resources;
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "web", change: "modified" });
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toEqual([{ key: "other configuration", kind: "changed" }]);
  });

  it("surfaces an unprojected volume field change as one generic row", () => {
    const vol = (labels: unknown[]) => ({ name: "data", spec: { size: "1Gi" }, labels });
    const out = diffSnapshots(
      mk({ resources: [], volumes: [vol([{ key: "tier", value: "hot" }])] }),
      mk({ resources: [], volumes: [vol([{ key: "tier", value: "cold" }])] }),
    ).volumes;
    expect(out).toEqual([
      { name: "data", change: "modified", rows: [{ key: "other configuration", kind: "changed" }] },
    ]);
  });

  it("ignores server-owned resource fields (id, stack_id, revision, outputs)", () => {
    const out = diffSnapshots(
      snap([web({ id: "a1", stack_id: "s1", revision: "3", outputs: [{ name: "url.port-87" }] })]),
      snap([web({ id: "b2", stack_id: "s1", revision: "4", outputs: [{ name: "url.port-88" }] })]),
    ).resources;
    expect(out).toEqual([]);
  });

  it("ignores server-owned volume fields (id, project_id, status)", () => {
    const vol = (over: Record<string, unknown>) => ({ name: "data", spec: { size: "1Gi" }, ...over });
    const out = diffSnapshots(
      mk({ resources: [], volumes: [vol({ id: "v1", project_id: "p1", status: { phase: "Bound" } })] }),
      mk({ resources: [], volumes: [vol({ id: "v2", project_id: "p1", status: { phase: "Pending" } })] }),
    ).volumes;
    expect(out).toEqual([]);
  });

  it("does not emit a generic row for resolver-written revisions on an unpinned spec", () => {
    const gitWeb = (git: Record<string, unknown>) => ({
      name: "web",
      source: { git: { repo_url: "https://github.com/acme/app", ...git } },
    });
    const out = diffSnapshots(
      snap([gitWeb({ branch: "master", commit: "b1eff14" })]),
      snap([gitWeb({})]),
    ).resources;
    expect(out).toEqual([]);
  });

  it("treats permuted env vars as unchanged", () => {
    const w = (envs: { name: string; value: string }[]) =>
      web({ execution_config: { command: ["node", "a.js"], environment_variables: envs } });
    const out = diffSnapshots(
      snap([w([{ name: "A", value: "1" }, { name: "B", value: "2" }])]),
      snap([w([{ name: "B", value: "2" }, { name: "A", value: "1" }])]),
    ).resources;
    expect(out).toEqual([]);
  });
});

describe("diffSnapshots git sources", () => {
  const gitWeb = (git: Record<string, unknown>) => ({
    name: "web",
    source: { git: { repo_url: "https://github.com/acme/app", dockerfile_path: "Dockerfile", build_context: ".", ...git } },
    ports: [{ number: 3000 }],
  });

  it("flags a branch change as a changed configuration row", () => {
    // A saved branch edit was invisible in the staged diff (only image/ports/
    // command/args were projected) — the pill said "1 change" from session
    // dirt while the modal showed nothing.
    const out = diffSnapshots(snap([gitWeb({ branch: "master" })]), snap([gitWeb({ branch: "masterd" })])).resources;
    expect(out).toHaveLength(1);
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toContainEqual({ key: "branch", from: "master", to: "masterd", kind: "changed" });
  });

  it("flags a repo change", () => {
    const out = diffSnapshots(
      snap([gitWeb({})]),
      snap([{ ...gitWeb({}), source: { git: { repo_url: "https://github.com/acme/other", dockerfile_path: "Dockerfile", build_context: "." } } }]),
    ).resources;
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toContainEqual({ key: "repo", from: "https://github.com/acme/app", to: "https://github.com/acme/other", kind: "changed" });
  });

  it("ignores resolver-written revisions when the spec side pins none", () => {
    // Deploy resolves the revision and writes branch+commit into the snapshot.
    // An unpinned spec differing only by those is NOT drift.
    const deployed = gitWeb({ branch: "master", commit: "b1eff1415a6b30ff3d476c2e907862f61a98a70d" });
    const unpinnedSpec = gitWeb({});
    expect(diffSnapshots(snap([deployed]), snap([unpinnedSpec])).resources).toEqual([]);
  });

  it("ignores the resolver-written commit when the spec tracks a branch", () => {
    // Branch-tracking spec: branch set, no commit. Deploy resolves the branch
    // and writes the commit into the snapshot — comparing that against the
    // spec's empty commit read as MODIFIED forever after every deploy.
    const deployed = gitWeb({ branch: "main", commit: "0f5c32589c37115508af0f48fe6d566925493700" });
    const branchTrackingSpec = gitWeb({ branch: "main" });
    expect(diffSnapshots(snap([deployed]), snap([branchTrackingSpec])).resources).toEqual([]);
  });

  it("flags only the branch change on a branch-tracking spec, without a phantom commit row", () => {
    const deployed = gitWeb({ branch: "master", commit: "abc123" });
    const branchTrackingSpec = gitWeb({ branch: "dev" });
    const out = diffSnapshots(snap([deployed]), snap([branchTrackingSpec])).resources;
    expect(out).toHaveLength(1);
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toEqual([{ key: "branch", from: "master", to: "dev", kind: "changed" }]);
  });

  it("flags a commit pin change beside an unchanged branch as a single commit row", () => {
    const out = diffSnapshots(
      snap([gitWeb({ branch: "main", commit: "aaa111" })]),
      snap([gitWeb({ branch: "main", commit: "bbb222" })]),
    ).resources;
    expect(out).toHaveLength(1);
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toEqual([{ key: "commit", from: "aaa111", to: "bbb222", kind: "changed" }]);
  });

  it("still pairs a renamed unpinned git resource instead of add+remove", () => {
    const deployed = gitWeb({ branch: "master", commit: "abc123" });
    const renamed = { ...gitWeb({}), name: "web2" };
    const out = diffSnapshots(snap([deployed]), snap([renamed])).resources;
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "web2", change: "renamed", fromName: "web" });
  });
});
