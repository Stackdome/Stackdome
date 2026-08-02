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
      {
        name: "cache",
        change: "removed",
        rows: [
          { key: "access mode", from: "ReadWriteOnce", kind: "removed" },
          { key: "size", from: "500Mi", kind: "removed" },
        ],
        note: "Volume removed from this release.",
      },
      {
        name: "logs",
        change: "added",
        rows: [
          { key: "access mode", to: "ReadWriteOnce", kind: "added" },
          { key: "size", to: "1Gi", kind: "added" },
        ],
      },
    ]);
  });

  it("returns no volume diff when volumes are unchanged", () => {
    const v = [vol("data", "1Gi")];
    expect(diffSnapshots(mk({ resources: [], volumes: v }), mk({ resources: [], volumes: v })).volumes).toEqual([]);
  });
});

/**
 * References and mounts are stored as connections but surface on the resource
 * that reads them, never as a category of their own.
 */
describe("diffSnapshots references", () => {
  const api = (over: Record<string, unknown> = {}) => ({
    name: "api",
    source: { image: { ref: "api:1" } },
    ports: [],
    execution_config: { command: [], args: [], environment_variables: [] },
    ...over,
  });
  const dbConn = (envName: string, output: string, to = "api") => ({
    kind: "env",
    from: { type: "addon/postgres", id: "addon-1" },
    to: { type: "stack_resource", name: to },
    config: { database: "app", superuser: false },
    mappings: [{ target: { type: "env", name: envName }, value: { output } }],
  });
  const mount = (to = "api", volume = "data") => ({
    kind: "volume_mount",
    from: { type: "volume", name: volume },
    to: { type: "stack_resource", name: to },
    config: { mount_path: "/data" },
  });

  it("reports a changed reference under the resource that reads it", () => {
    const prev = mk({ resources: [api()], connections: [dbConn("DATABASE_URL", "url")] });
    const cur = mk({ resources: [api()], connections: [dbConn("DATABASE_URL", "host")] });
    const out = diffSnapshots(prev, cur).resources;
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "api", change: "modified" });
    const env = out[0].sections.find((s) => s.kind === "environment")!;
    expect(env.rows).toEqual([
      { key: "DATABASE_URL", from: "addon · url", to: "addon · host", kind: "changed" },
    ]);
  });

  it("reports an added reference", () => {
    const out = diffSnapshots(
      mk({ resources: [api()], connections: [] }),
      mk({ resources: [api()], connections: [dbConn("DATABASE_URL", "url")] }),
    ).resources;
    const env = out[0].sections.find((s) => s.kind === "environment")!;
    expect(env.rows).toEqual([{ key: "DATABASE_URL", to: "addon · url", kind: "added" }]);
  });

  it("reports a disconnected volume mount", () => {
    const out = diffSnapshots(
      mk({ resources: [api()], connections: [mount()] }),
      mk({ resources: [api()], connections: [] }),
    ).resources;
    expect(out).toHaveLength(1);
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toEqual([{ key: "mount /data", from: "data", to: undefined, kind: "removed" }]);
  });

  it("reports no churn when a resource endpoint was only renamed", () => {
    const prev = mk({ resources: [api({ name: "mysql" })], connections: [mount("mysql", "mysql-data")] });
    const cur = mk({ resources: [api({ name: "mysqla" })], connections: [mount("mysqla", "mysql-data")] });
    expect(diffSnapshots(prev, cur).resources).toEqual([
      { name: "mysqla", fromName: "mysql", change: "renamed", sections: [] },
    ]);
  });

  it("leaves a volume named like a renamed resource alone", () => {
    const prev = mk({ resources: [api({ name: "web" }), api()], connections: [mount("api", "web")] });
    const cur = mk({ resources: [api({ name: "web2" }), api()], connections: [mount("api", "web")] });
    const out = diffSnapshots(prev, cur).resources;
    expect(out).toEqual([{ name: "web2", fromName: "web", change: "renamed", sections: [] }]);
  });
});

describe("diffSnapshots surfaces fields nobody projected", () => {
  it("reports an init_spec change rather than staying silent", () => {
    const out = diffSnapshots(
      snap([web({ init_spec: { command: ["migrate"] } })]),
      snap([web({ init_spec: { command: ["migrate", "--seed"] } })]),
    ).resources;
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "web", change: "modified" });
    const deployment = out[0].sections.find((s) => s.kind === "deployment")!;
    expect(deployment.rows).toEqual([
      {
        key: "init",
        from: JSON.stringify({ command: ["migrate"] }),
        to: JSON.stringify({ command: ["migrate", "--seed"] }),
        kind: "changed",
      },
    ]);
  });

  it("reports an unprojected volume field change", () => {
    const prev = mk({ resources: [], volumes: [{ name: "data", spec: { size: "1Gi" }, labels: { tier: "hot" } }] });
    const cur = mk({ resources: [], volumes: [{ name: "data", spec: { size: "1Gi" }, labels: { tier: "cold" } }] });
    const out = diffSnapshots(prev, cur).volumes;
    expect(out).toHaveLength(1);
    expect(out[0].rows).toEqual([
      { key: "labels", from: JSON.stringify({ tier: "hot" }), to: JSON.stringify({ tier: "cold" }), kind: "changed" },
    ]);
  });
});

/**
 * Covers the PRESENTATION of git revisions — which rows a card shows. The
 * model-level tests in one-edit-one-change.test.ts do not reach this.
 */
describe("diffSnapshots git revisions", () => {
  const gitWeb = (git: Record<string, unknown>) => ({
    name: "web",
    source: { git: { repo_url: "https://github.com/acme/demo.git", dockerfile_path: "Dockerfile", build_context: ".", ...git } },
    ports: [],
    execution_config: { command: [], args: [], environment_variables: [] },
  });

  it("flags only the branch change, with no phantom commit row", () => {
    const prev = snap([gitWeb({ branch: "main", commit: "resolved-by-the-pin-resolver" })]);
    const cur = snap([gitWeb({ branch: "next" })]);
    const out = diffSnapshots(prev, cur).resources;
    expect(out).toHaveLength(1);
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toEqual([{ key: "branch", from: "main", to: "next", kind: "changed" }]);
  });

  it("flags a commit pin change beside an unchanged branch as a single commit row", () => {
    const prev = snap([gitWeb({ branch: "main", commit: "aaaaaaa" })]);
    const cur = snap([gitWeb({ branch: "main", commit: "bbbbbbb" })]);
    const out = diffSnapshots(prev, cur).resources;
    expect(out).toHaveLength(1);
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toEqual([{ key: "commit", from: "aaaaaaa", to: "bbbbbbb", kind: "changed" }]);
  });

  it("ignores the resolver-written commit when the spec tracks a branch", () => {
    const prev = snap([gitWeb({ branch: "main", commit: "9f1c2b7" })]);
    const cur = snap([gitWeb({ branch: "main" })]);
    expect(diffSnapshots(prev, cur).resources).toEqual([]);
  });
});

/**
 * The deploy pill counts what this module returns and the changes modal renders
 * it, so no entry may survive without a legible row. Values format independently
 * of the rule that found them, so the guard belongs on the output.
 */
describe("diffSnapshots never reports a change it cannot show", () => {
  const cases: Array<[string, unknown[], unknown[]]> = [
    ["an empty array replacing an absent one", [web({ depends_on: undefined })], [web({ depends_on: [] })]],
    ["an empty command replacing an absent one", [web({ execution_config: {} })], [web({ execution_config: { command: [] } })]],
    ["ports emptied to an empty array", [web({ ports: [] })], [web({ ports: [] })]],
    ["an absent env list against an empty one", [web({ execution_config: { environment_variables: [] } })], [web({ execution_config: {} })]],
  ];

  for (const [label, prev, cur] of cases) {
    it(`stays silent about ${label}`, () => {
      expect(diffSnapshots(snap(prev), snap(cur)).resources).toEqual([]);
    });
  }

  /**
   * A port carrying a protocol but no number has no friendly phrasing, and the
   * model still calls it a change. It falls back to the raw value: dropping the
   * row would hide the entry from the count it explains and disable Deploy,
   * which only enables on a non-empty diff.
   */
  it("falls back to the raw value rather than dropping a row it cannot phrase", () => {
    const prev = snap([web({ ports: [{ protocol: "TCP" }] })]);
    const cur = snap([web({ ports: [{ protocol: "UDP" }] })]);
    const [entry] = diffSnapshots(prev, cur).resources;
    expect(entry).toMatchObject({ name: "web", change: "modified" });
    const rows = entry.sections.flatMap((s) => s.rows);
    expect(rows).toHaveLength(1);
    expect(rows[0].from).toContain("TCP");
    expect(rows[0].to).toContain("UDP");
  });

  it("still prefers the friendly phrasing when it has one", () => {
    const prev = snap([web({ ports: [{ number: 3000 }] })]);
    const cur = snap([web({ ports: [{ number: 8080 }] })]);
    const rows = diffSnapshots(prev, cur).resources[0].sections.flatMap((s) => s.rows);
    expect(rows[0]).toMatchObject({ from: "3000", to: "8080" });
  });

  it("gives every reported entry at least one row to show", () => {
    const prev = snap([web({ ports: [{ protocol: "TCP" }], execution_config: { command: ["a"] } })]);
    const cur = snap([web({ ports: [{ protocol: "UDP" }], execution_config: { command: ["b"] } })]);
    for (const entry of diffSnapshots(prev, cur).resources) {
      const rows = entry.sections.flatMap((s) => s.rows);
      expect(rows.length).toBeGreaterThan(0);
      for (const row of rows) expect(row.from !== undefined || row.to !== undefined).toBe(true);
    }
  });
});
