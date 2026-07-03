# Canvas Connections From API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Render stack-canvas edges and nodes from real `StackConnection` semantics (plus the server topology endpoint) instead of inferring edges from env rows.

**Architecture:** Two producers, one merge. A pure local projector turns the edit-session draft into a graph whose edges come from `buildDesiredConnections`/`mountsToConnections` plus `depends_on` (all data already client-side). A topology fetch (`GET .../stacks/{id}/topology`) contributes server-derived edges and node runtime state after each autosave refresh. A pure merge combines them; dagre layout and React Flow rendering are unchanged.

**Tech Stack:** React 19, @xyflow/react (React Flow), dagre, vitest, generated OpenAPI types (`frontend/src/api/types/openapi.d.ts` — already up to date, no regen).

**Spec:** `docs/superpowers/specs/2026-07-04-canvas-connections-api-design.md`

## Global Constraints

- All paths relative to worktree root `/Users/akshaysasidharan/code/stackdome/.claude/worktrees/stack-canvas-editor` unless absolute.
- Brand design system only: CSS tokens from `frontend/src/index.css` (`var(--brand)`, semantic Tailwind classes like `text-fg-muted`, `bg-card`, `border-border`). **No raw hex.**
- No magic strings — use the `NODE_KIND`, `EDGE_KIND` style const objects defined in tasks below.
- Frontend tests: `pnpm --prefix frontend exec vitest run <path>` (the `run` flag matters — bare `vitest` watches).
- Type check: `pnpm --prefix frontend exec tsc -b`. Lint: `pnpm --prefix frontend lint`.
- Commit after every task (conventional commits, `fix(stacks):`/`feat(stacks):` style as in recent history).
- Edge direction convention: **source = producing end** (`StackConnection.from`), **target = consuming end** (`StackConnection.to`). This flips today's resource→addon arrows to addon→resource; that is intentional (matches `TopologyEdge.source` semantics).
- Node id prefixes: `resource:<name>`, `addon:<id>`, `secret:<id>`, `volume:<name>`, `objectstore:<name-or-id>`.

---

### Task 1: Topology API client

**Files:**
- Modify: `frontend/src/api/topology.ts` (create)

**Interfaces:**
- Produces: `StackTopology`, `TopologyNode`, `TopologyEdge` type re-exports; `getStackTopology(orgId: string, teamName: string, stackId: string): Promise<StackTopology>`.
- Pattern source: `frontend/src/api/connections.ts` (same `api` client from `./client`; baseURL already includes `/api/v1`).

Thin generated-type client — no unit test (matches `connections.ts` precedent).

- [ ] **Step 1: Write the client**

```ts
// frontend/src/api/topology.ts
// Read-side client for the stack topology graph endpoint.
import api from "./client";
import type { components } from "./types/openapi";

export type StackTopology = components["schemas"]["StackTopology"];
export type TopologyNode = components["schemas"]["TopologyNode"];
export type TopologyEdge = components["schemas"]["TopologyEdge"];
export type TopologyNodeRef = components["schemas"]["TopologyNodeRef"];

export async function getStackTopology(orgId: string, teamName: string, stackId: string): Promise<StackTopology> {
  const response = await api.get<StackTopology>(
    `/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/topology`,
  );
  return response.data;
}
```

- [ ] **Step 2: Type check**

Run: `pnpm --prefix frontend exec tsc -b`
Expected: clean (StackTopology exists in generated types).

- [ ] **Step 3: Commit**

```bash
git add frontend/src/api/topology.ts
git commit -m "feat(stacks): add stack topology API client"
```

---

### Task 2: Connection-based graph projector (replaces env-row inference)

**Files:**
- Create: `frontend/src/pages/stacks/lib/canvas/graph-from-connections.ts`
- Test: `frontend/src/pages/stacks/lib/canvas/tests/graph-from-connections.test.ts`
- Reference (read only): `frontend/src/pages/stacks/lib/canvas/derive-graph.ts` (node-building code migrates verbatim; file deleted in Task 6), `frontend/src/pages/stacks/lib/connection-mapping.ts`

**Interfaces:**
- Consumes: `splitEnvRows(resourceName, rows)`, `mountsToConnections(resourceName, mounts)` from `@/pages/stacks/lib/connection-mapping` (existing, unchanged); `nodePresentation` from `./node-presentation`; `FormStackResourceData`, `FormEnvVarData` from `@/pages/stacks/schemas/form-schema`.
- Produces (used by Tasks 3–7):

```ts
export const NODE_KIND = {
  service: "service", addon: "addon",
  secret: "secret", volume: "volume", objectStore: "objectStore",
} as const;
export type NodeKind = (typeof NODE_KIND)[keyof typeof NODE_KIND];
export type AttachmentKind = typeof NODE_KIND.secret | typeof NODE_KIND.volume | typeof NODE_KIND.objectStore;

export const EDGE_KIND = {
  env: "env", volumeMount: "volume_mount",
  buildArtifactSource: "build_artifact_source", dependsOn: "depends_on",
} as const;
export type EdgeKind = (typeof EDGE_KIND)[keyof typeof EDGE_KIND];

export const EDGE_SOURCE_OF_TRUTH = { connection: "connection", derived: "derived" } as const;
export type EdgeSourceOfTruth = (typeof EDGE_SOURCE_OF_TRUTH)[keyof typeof EDGE_SOURCE_OF_TRUTH];

export interface ConnectionEdgeData {
  kind: EdgeKind;
  sourceOfTruth: EdgeSourceOfTruth;
  [key: string]: unknown;
}

export interface AttachmentNodeData {
  kind: AttachmentKind;
  name: string;
  kindLabel: string;      // "SECRET" | "VOLUME" | "OBJECT STORE"
  [key: string]: unknown;
}

// ResourceNodeData, DirtyState, VolumeChip, DirtyInput: migrated unchanged from derive-graph.ts
export interface CanvasNode {
  id: string;
  type: "resource" | "attachment";
  data: ResourceNodeData | AttachmentNodeData;
  position: { x: number; y: number };
}
export interface CanvasEdge {
  id: string;
  source: string;
  target: string;
  type: "connection";
  data: ConnectionEdgeData;
}
export interface CanvasGraph { nodes: CanvasNode[]; edges: CanvasEdge[] }

export interface DeriveGraphInput {
  resources: Partial<FormStackResourceData>[];
  linkedAddonIds: ReadonlySet<string>;
  addonNameById: ReadonlyMap<string, string>;
  secretNameById: ReadonlyMap<string, string>;
  dirty?: DirtyInput;
}
export function deriveGraph(input: DeriveGraphInput): CanvasGraph;
export function edgeKey(kind: string, source: string, target: string): string; // `${kind}:${source}->${target}`
export const NODE_ID_PREFIX = {
  resource: "resource:", addon: "addon:", secret: "secret:", volume: "volume:", objectstore: "objectstore:",
} as const;
```

**Algorithm for `deriveGraph`:**
1. Service nodes per resource and addon nodes per linked addon id — migrate the existing builders (`servicePresentation`, `volumeChips`, dirty-state helpers) from `derive-graph.ts` verbatim, including volume chips on the card.
2. Per resource, compute authored connections: `splitEnvRows(name, envRows).connections` concat `mountsToConnections(name, (resource.volume_mounts ?? []) as FormMountRow[])`.
3. Project each connection to an edge: `source = nodeIdOfConnRef(conn.from)`, `target = nodeIdOfConnRef(conn.to)`, `kind = conn.kind`, `sourceOfTruth: "connection"`, `id = edgeKey(kind, source, target)`.
   - `nodeIdOfConnRef` maps `{type:"stack_resource",name}` → `resource:<name>`, `{type:"addon/postgres",id}` → `addon:<id>`, `{type:"secret",id}` → `secret:<id>`, `{type:"volume",name}` → `volume:<name>`, `{type:"object_store",...}` → `objectstore:<name ?? id>`; unknown → null (edge skipped).
   - **Secret and volume endpoint nodes are created on demand** (compact attachment nodes; secret label from `secretNameById`, falling back to the id; volume label = volume name). Resource/addon endpoints must already exist or the edge is skipped (same guard as today).
4. `depends_on` derived edges: for each resource `r` and each `dep` of `r.depends_on ?? []`, edge `source = resource:<dep>`, `target = resource:<r.name>`, `kind: "depends_on"`, `sourceOfTruth: "derived"`, skipped when the dep node doesn't exist.
5. Dedupe all edges by `edgeKey` (kind + endpoints); self-edges skipped.

- [ ] **Step 1: Write the failing tests**

```ts
// frontend/src/pages/stacks/lib/canvas/tests/graph-from-connections.test.ts
import { describe, expect, it } from "vitest";
import { deriveGraph, EDGE_KIND, NODE_KIND } from "../graph-from-connections";
import type { FormStackResourceData } from "@/pages/stacks/schemas/form-schema";

const web = (envRows: unknown[] = [], extra: Partial<FormStackResourceData> = {}): Partial<FormStackResourceData> =>
  ({
    name: "web",
    image_spec: { image: "nginx:1" },
    execution_config: { environment_variables: envRows },
    ...extra,
  }) as Partial<FormStackResourceData>;

const base = {
  linkedAddonIds: new Set<string>(),
  addonNameById: new Map<string, string>(),
  secretNameById: new Map<string, string>(),
};

describe("deriveGraph (connection projection)", () => {
  it("projects addon env rows into an addon→resource env edge (from = producer)", () => {
    const g = deriveGraph({
      ...base,
      resources: [web([{ from: "addon", name: "DB_URL", addonId: "a1", superuser: false, database: "app", credField: "url" }])],
      linkedAddonIds: new Set(["a1"]),
      addonNameById: new Map([["a1", "tooljet-db"]]),
    });
    expect(g.edges).toEqual([
      expect.objectContaining({
        id: "env:addon:a1->resource:web",
        source: "addon:a1",
        target: "resource:web",
        type: "connection",
        data: { kind: EDGE_KIND.env, sourceOfTruth: "connection" },
      }),
    ]);
  });

  it("creates a secret attachment node on demand and labels it from secretNameById", () => {
    const g = deriveGraph({
      ...base,
      resources: [web([{ from: "secret", name: "TOKEN", secretId: "s1", secretKey: "token" }])],
      secretNameById: new Map([["s1", "api-creds"]]),
    });
    const secretNode = g.nodes.find((n) => n.id === "secret:s1");
    expect(secretNode?.type).toBe("attachment");
    expect(secretNode?.data).toMatchObject({ kind: NODE_KIND.secret, name: "api-creds" });
    expect(g.edges.map((e) => e.id)).toContain("env:secret:s1->resource:web");
  });

  it("projects volume mounts into a volume node plus volume_mount edge", () => {
    const g = deriveGraph({
      ...base,
      resources: [web([], { volume_mounts: [{ name: "data", source_volume_name: "data", target_path: "/var/data" }] })],
    });
    expect(g.nodes.find((n) => n.id === "volume:data")?.data).toMatchObject({ kind: NODE_KIND.volume, name: "data" });
    expect(g.edges.map((e) => e.id)).toContain("volume_mount:volume:data->resource:web");
  });

  it("emits depends_on edges as derived and skips unknown deps", () => {
    const g = deriveGraph({
      ...base,
      resources: [web(), { name: "worker", depends_on: ["web", "ghost"] } as Partial<FormStackResourceData>],
    });
    const dep = g.edges.find((e) => e.data.kind === EDGE_KIND.dependsOn);
    expect(dep).toMatchObject({ source: "resource:web", target: "resource:worker", data: { sourceOfTruth: "derived" } });
    expect(g.edges.filter((e) => e.data.kind === EDGE_KIND.dependsOn)).toHaveLength(1);
  });

  it("dedupes edges with the same kind and endpoints (two rows, one connection group)", () => {
    const rows = [
      { from: "resource", name: "API_URL", resourceName: "api", output: "url" },
      { from: "resource", name: "API_HOST", resourceName: "api", output: "host" },
    ];
    const g = deriveGraph({ ...base, resources: [web(rows), { name: "api" } as Partial<FormStackResourceData>] });
    expect(g.edges.filter((e) => e.data.kind === EDGE_KIND.env)).toHaveLength(1);
  });

  it("skips edges whose resource endpoint does not exist and skips in-progress rows", () => {
    const g = deriveGraph({
      ...base,
      resources: [web([
        { from: "resource", name: "X", resourceName: "missing", output: "url" },
        { from: "secret", name: "Y", secretId: "", secretKey: "" }, // in-progress: dropped by splitEnvRows
      ])],
    });
    expect(g.edges).toHaveLength(0);
    expect(g.nodes.filter((n) => n.type === "attachment")).toHaveLength(0);
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/lib/canvas/tests/graph-from-connections.test.ts`
Expected: FAIL — module `../graph-from-connections` not found.

- [ ] **Step 3: Implement `graph-from-connections.ts`**

Start by copying from `derive-graph.ts`: the doc-comment style, `ResourceNodeData`, `CanvasNode`-related types, `DirtyInput`, dirty-state helpers, `servicePresentation`, `volumeChips`, and the service/addon node loops. Then add the interfaces exactly as in the **Produces** block above and implement the algorithm:

```ts
import { splitEnvRows, mountsToConnections, type FormMountRow, type FormEnvRow } from "@/pages/stacks/lib/connection-mapping";
import type { StackConnection } from "@/api/connections";

const ATTACHMENT_LABEL: Record<AttachmentKind, string> = {
  [NODE_KIND.secret]: "SECRET",
  [NODE_KIND.volume]: "VOLUME",
  [NODE_KIND.objectStore]: "OBJECT STORE",
};

export function edgeKey(kind: string, source: string, target: string): string {
  return `${kind}:${source}->${target}`;
}

function nodeIdOfConnRef(ref: StackConnection["from"] | undefined): string | null {
  if (!ref) return null;
  switch (ref.type) {
    case "stack_resource": return ref.name ? NODE_ID_PREFIX.resource + ref.name : null;
    case "addon/postgres": return ref.id ? NODE_ID_PREFIX.addon + ref.id : null;
    case "secret":         return ref.id ? NODE_ID_PREFIX.secret + ref.id : null;
    case "volume":         return ref.name ? NODE_ID_PREFIX.volume + ref.name : null;
    case "object_store":   return (ref.name ?? ref.id) ? NODE_ID_PREFIX.objectstore + (ref.name ?? ref.id) : null;
    default: return null;
  }
}
```

In `deriveGraph`, after building service/addon nodes:

```ts
const nodeById = new Map(nodes.map((n) => [n.id, n]));
const edges: CanvasEdge[] = [];
const seen = new Set<string>();

const ensureAttachment = (id: string, kind: AttachmentKind, name: string) => {
  if (nodeById.has(id)) return;
  const node: CanvasNode = {
    id, type: "attachment", position: { x: 0, y: 0 },
    data: { kind, name, kindLabel: ATTACHMENT_LABEL[kind] },
  };
  nodeById.set(id, node);
  nodes.push(node);
};

const addEdge = (kind: EdgeKind, sourceOfTruth: EdgeSourceOfTruth, source: string | null, target: string | null) => {
  if (!source || !target || source === target) return;
  if (!nodeById.has(source) || !nodeById.has(target)) return;
  const id = edgeKey(kind, source, target);
  if (seen.has(id)) return;
  seen.add(id);
  edges.push({ id, source, target, type: "connection", data: { kind, sourceOfTruth } });
};

for (const resource of input.resources) {
  const name = resource.name ?? "";
  const conns: StackConnection[] = [
    ...splitEnvRows(name, envVarsOf(resource) as FormEnvRow[]).connections,
    ...mountsToConnections(name, (resource.volume_mounts ?? []) as FormMountRow[]),
  ];
  for (const conn of conns) {
    // Secret/volume producers materialize as attachment nodes on demand.
    if (conn.from?.type === "secret" && conn.from.id) {
      ensureAttachment(NODE_ID_PREFIX.secret + conn.from.id, NODE_KIND.secret,
        input.secretNameById.get(conn.from.id) ?? conn.from.id);
    } else if (conn.from?.type === "volume" && conn.from.name) {
      ensureAttachment(NODE_ID_PREFIX.volume + conn.from.name, NODE_KIND.volume, conn.from.name);
    }
    addEdge(conn.kind as EdgeKind, EDGE_SOURCE_OF_TRUTH.connection, nodeIdOfConnRef(conn.from), nodeIdOfConnRef(conn.to));
  }
}

for (const resource of input.resources) {
  for (const dep of resource.depends_on ?? []) {
    addEdge(EDGE_KIND.dependsOn, EDGE_SOURCE_OF_TRUTH.derived,
      NODE_ID_PREFIX.resource + dep, NODE_ID_PREFIX.resource + (resource.name ?? ""));
  }
}

return { nodes, edges };
```

Note on `volume_mounts` typing: the form's mount rows carry `name`/`source_volume_name`/`target_path`; `mountsToConnections` accepts the loose `FormMountRow` and skips incomplete rows. If the form field names differ from `FormMountRow` (check `form-schema.ts` at implementation time), adapt the cast at this one call site — do not change `connection-mapping.ts`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/lib/canvas/tests/graph-from-connections.test.ts`
Expected: PASS (6 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/lib/canvas/graph-from-connections.ts frontend/src/pages/stacks/lib/canvas/tests/graph-from-connections.test.ts
git commit -m "feat(stacks): project canvas graph from connection semantics"
```

---

### Task 3: Topology merge (server derived edges + node state)

**Files:**
- Create: `frontend/src/pages/stacks/lib/canvas/merge-topology.ts`
- Test: `frontend/src/pages/stacks/lib/canvas/tests/merge-topology.test.ts`

**Interfaces:**
- Consumes: `CanvasGraph`, `CanvasNode`, `edgeKey`, `NODE_ID_PREFIX`, `EDGE_SOURCE_OF_TRUTH`, `NODE_KIND`, attachment types from `./graph-from-connections`; `StackTopology`, `TopologyNodeRef` from `@/api/topology`.
- Produces: `mergeTopology(local: CanvasGraph, server: StackTopology | null | undefined): CanvasGraph` — pure, copy-on-write (returns `local` untouched when `server` is nullish or contributes nothing).

**Merge rules (from spec):**
- Authored edges: local set is authoritative — server `source_of_truth: "connection"` edges are ignored.
- Server `source_of_truth: "derived"` edges: unioned in, deduped by `edgeKey(kind, source, target)` against local edges.
- Endpoint mapping via the same ref→id scheme (`stack_resource`→`resource:`, `addon/postgres`→`addon:`, `secret`→`secret:`, `volume`→`volume:`, `object_store`→`objectstore:`). A derived edge whose **resource/addon** endpoint is missing locally is skipped (deleted locally wins); a missing **secret/volume/object_store** endpoint is materialized as an attachment node using the server `TopologyNode.label`.
- Node `state`: for server nodes with `state` set and a matching local **resource-type** node, overlay `data.status = state`. (No status→dotState mapping exists in the codebase; do not invent one — dot stays presentation-driven. Deviation from spec noted and accepted.)

- [ ] **Step 1: Write the failing tests**

```ts
// frontend/src/pages/stacks/lib/canvas/tests/merge-topology.test.ts
import { describe, expect, it } from "vitest";
import { mergeTopology } from "../merge-topology";
import type { CanvasGraph } from "../graph-from-connections";
import type { StackTopology } from "@/api/topology";

const localGraph = (): CanvasGraph => ({
  nodes: [
    { id: "resource:web", type: "resource", position: { x: 0, y: 0 }, data: { kind: "service", name: "web", kindLabel: "WEB", glyph: "web", dotState: "ok", summary: "", volumes: [] } },
    { id: "resource:api", type: "resource", position: { x: 0, y: 0 }, data: { kind: "service", name: "api", kindLabel: "WEB", glyph: "web", dotState: "ok", summary: "", volumes: [] } },
  ],
  edges: [
    { id: "env:resource:api->resource:web", source: "resource:api", target: "resource:web", type: "connection", data: { kind: "env", sourceOfTruth: "connection" } },
  ],
});

describe("mergeTopology", () => {
  it("returns local unchanged when server is null", () => {
    const local = localGraph();
    expect(mergeTopology(local, null)).toBe(local);
  });

  it("ignores server authored edges (local wins) and does not duplicate", () => {
    const server = {
      nodes: [],
      edges: [{ kind: "env", source: { type: "stack_resource", name: "api" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "connection" }],
    } as unknown as StackTopology;
    expect(mergeTopology(localGraph(), server).edges).toHaveLength(1);
  });

  it("unions server derived edges, deduped against local", () => {
    const server = {
      nodes: [],
      edges: [
        { kind: "depends_on", source: { type: "stack_resource", name: "api" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "derived" },
        { kind: "env", source: { type: "stack_resource", name: "api" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "derived" }, // same key as local authored edge
      ],
    } as unknown as StackTopology;
    const merged = mergeTopology(localGraph(), server);
    expect(merged.edges.map((e) => e.id)).toEqual([
      "env:resource:api->resource:web",
      "depends_on:resource:api->resource:web",
    ]);
  });

  it("skips derived edges whose resource endpoint is gone locally, materializes missing secret endpoints", () => {
    const server = {
      nodes: [{ ref: { type: "secret", id: "s9" }, label: "legacy-creds" }],
      edges: [
        { kind: "depends_on", source: { type: "stack_resource", name: "ghost" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "derived" },
        { kind: "env", source: { type: "secret", id: "s9" }, target: { type: "stack_resource", name: "web" }, source_of_truth: "derived" },
      ],
    } as unknown as StackTopology;
    const merged = mergeTopology(localGraph(), server);
    expect(merged.edges.map((e) => e.id)).not.toContain("depends_on:resource:ghost->resource:web");
    expect(merged.nodes.find((n) => n.id === "secret:s9")?.data).toMatchObject({ kind: "secret", name: "legacy-creds" });
    expect(merged.edges.map((e) => e.id)).toContain("env:secret:s9->resource:web");
  });

  it("overlays server node state onto matching resource nodes as status", () => {
    const server = {
      nodes: [{ ref: { type: "stack_resource", name: "web" }, label: "web", state: "Degraded" }],
      edges: [],
    } as unknown as StackTopology;
    const merged = mergeTopology(localGraph(), server);
    expect(merged.nodes.find((n) => n.id === "resource:web")?.data.status).toBe("Degraded");
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/lib/canvas/tests/merge-topology.test.ts`
Expected: FAIL — module `../merge-topology` not found.

- [ ] **Step 3: Implement `merge-topology.ts`**

```ts
// Pure merge of the local (instant, authored) graph with server topology
// (derived edges + runtime state). Local is never mutated.
import type { StackTopology, TopologyNodeRef } from "@/api/topology";
import {
  EDGE_SOURCE_OF_TRUTH, NODE_ID_PREFIX, NODE_KIND, edgeKey,
  type AttachmentKind, type CanvasEdge, type CanvasGraph, type CanvasNode, type EdgeKind,
} from "./graph-from-connections";

const ATTACHMENT_BY_REF_TYPE: Partial<Record<string, { kind: AttachmentKind; prefix: string }>> = {
  secret: { kind: NODE_KIND.secret, prefix: NODE_ID_PREFIX.secret },
  volume: { kind: NODE_KIND.volume, prefix: NODE_ID_PREFIX.volume },
  object_store: { kind: NODE_KIND.objectStore, prefix: NODE_ID_PREFIX.objectstore },
};

function nodeIdOfRef(ref: TopologyNodeRef | undefined): string | null {
  if (!ref) return null;
  switch (ref.type) {
    case "stack_resource": return ref.name ? NODE_ID_PREFIX.resource + ref.name : null;
    case "addon/postgres": return ref.id ? NODE_ID_PREFIX.addon + ref.id : null;
    case "secret":         return ref.id ? NODE_ID_PREFIX.secret + ref.id : null;
    case "volume":         return ref.name ? NODE_ID_PREFIX.volume + ref.name : null;
    case "object_store":   return (ref.name ?? ref.id) ? NODE_ID_PREFIX.objectstore + (ref.name ?? ref.id) : null;
    default: return null;
  }
}
```

Full merge:

```ts
export function mergeTopology(local: CanvasGraph, server: StackTopology | null | undefined): CanvasGraph {
  if (!server) return local;

  const nodes = [...local.nodes];
  const nodeById = new Map(nodes.map((n) => [n.id, n]));
  const serverNodeByListId = new Map(
    (server.nodes ?? []).map((n) => [nodeIdOfRef(n.ref), n] as const).filter(([id]) => id !== null),
  );
  let changed = false;

  // Runtime-state overlay onto matching local resource nodes.
  for (const [id, serverNode] of serverNodeByListId) {
    if (!serverNode.state) continue;
    const idx = nodes.findIndex((n) => n.id === id && n.type === "resource");
    if (idx === -1) continue;
    const node = nodes[idx];
    if ((node.data as { status?: string }).status === serverNode.state) continue;
    const next: CanvasNode = { ...node, data: { ...node.data, status: serverNode.state } };
    nodes[idx] = next;
    nodeById.set(id!, next);
    changed = true;
  }

  // Union server-derived edges; local authored set is authoritative for the rest.
  const edges = [...local.edges];
  const seen = new Set(edges.map((e) => e.id));
  const ensureAttachment = (id: string, refType: string, label: string): boolean => {
    if (nodeById.has(id)) return true;
    const meta = ATTACHMENT_BY_REF_TYPE[refType];
    if (!meta) return false;
    const node: CanvasNode = {
      id, type: "attachment", position: { x: 0, y: 0 },
      data: { kind: meta.kind, name: label, kindLabel: ATTACHMENT_LABEL[meta.kind] },
    };
    nodes.push(node);
    nodeById.set(id, node);
    changed = true;
    return true;
  };

  for (const edge of server.edges ?? []) {
    if (edge.source_of_truth !== EDGE_SOURCE_OF_TRUTH.derived) continue;
    const source = nodeIdOfRef(edge.source);
    const target = nodeIdOfRef(edge.target);
    if (!source || !target || source === target) continue;
    const endpointOk = (id: string, ref: TopologyNodeRef) =>
      nodeById.has(id) ||
      ensureAttachment(id, ref.type, serverNodeByListId.get(id)?.label ?? ref.name ?? ref.id ?? id);
    if (!endpointOk(source, edge.source) || !endpointOk(target, edge.target)) continue;
    const id = edgeKey(edge.kind, source, target);
    if (seen.has(id)) continue;
    seen.add(id);
    edges.push({ id, source, target, type: "connection", data: { kind: edge.kind as EdgeKind, sourceOfTruth: EDGE_SOURCE_OF_TRUTH.derived } });
    changed = true;
  }

  return changed ? { nodes, edges } : local;
}
```

`ATTACHMENT_LABEL` is exported from `graph-from-connections.ts` (add `export` there in Task 2).

- [ ] **Step 4: Run tests to verify they pass**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/lib/canvas/tests/merge-topology.test.ts`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/lib/canvas/merge-topology.ts frontend/src/pages/stacks/lib/canvas/tests/merge-topology.test.ts
git commit -m "feat(stacks): merge server topology into local canvas graph"
```

---

### Task 4: Topology query hook

**Files:**
- Create: `frontend/src/pages/stacks/hooks/use-stack-topology.ts`
- Pattern source: `frontend/src/pages/stacks/hooks/use-secrets.ts` (plain useState/useEffect fetch hooks — this repo does not use react-query)

**Interfaces:**
- Consumes: `getStackTopology`, `StackTopology` from `@/api/topology`.
- Produces: `useStackTopology(args: { ids: { orgId: string; teamName: string; stackId: string } | null; refreshKey: number }): { topology: StackTopology | null }`.

Behavior: fetch when `ids` is non-null, refetch when `refreshKey` changes, `null` while loading/on failure (silent fallback per spec — `console.debug` only, no toast), stale responses discarded via a cancelled flag. Hook is trivial glue over tested pure parts — no unit test (matches `use-secrets.ts` precedent).

- [ ] **Step 1: Write the hook**

```ts
// frontend/src/pages/stacks/hooks/use-stack-topology.ts
// Server topology feed for the canvas. Silent enhancement layer: failures fall
// back to null and the canvas renders the local graph alone.
import { useEffect, useState } from "react";
import { getStackTopology, type StackTopology } from "@/api/topology";

export interface UseStackTopologyArgs {
  /** Null disables the fetch (draft stacks have no server topology). */
  ids: { orgId: string; teamName: string; stackId: string } | null;
  /** Bump to refetch — wired to autosave refreshes. */
  refreshKey: number;
}

export function useStackTopology({ ids, refreshKey }: UseStackTopologyArgs): { topology: StackTopology | null } {
  const [topology, setTopology] = useState<StackTopology | null>(null);

  useEffect(() => {
    if (!ids) {
      setTopology(null);
      return;
    }
    let cancelled = false;
    getStackTopology(ids.orgId, ids.teamName, ids.stackId)
      .then((t) => {
        if (!cancelled) setTopology(t);
      })
      .catch((error) => {
        console.debug("stack topology fetch failed; canvas falls back to local graph", error);
        if (!cancelled) setTopology(null);
      });
    return () => {
      cancelled = true;
    };
  }, [ids?.orgId, ids?.teamName, ids?.stackId, refreshKey]); // eslint-disable-line react-hooks/exhaustive-deps -- keyed on primitive id parts

  return { topology };
}
```

- [ ] **Step 2: Type check**

Run: `pnpm --prefix frontend exec tsc -b`
Expected: clean.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/stacks/hooks/use-stack-topology.ts
git commit -m "feat(stacks): add stack topology hook with silent fallback"
```

---

### Task 5: Edge and attachment-node rendering

**Files:**
- Modify: `frontend/src/pages/stacks/components/canvas/edges/ConnectionEdge.tsx`
- Create: `frontend/src/pages/stacks/components/canvas/nodes/AttachmentNode.tsx`
- Modify: `frontend/src/pages/stacks/components/canvas/CanvasEditor.tsx` (register node type, widen node prop type, update hint copy)

**Interfaces:**
- Consumes: `ConnectionEdgeData`, `AttachmentNodeData`, `EDGE_KIND`, `EDGE_SOURCE_OF_TRUTH` from `@/pages/stacks/lib/canvas/graph-from-connections`.
- Produces: `AttachmentFlowNode = Node<AttachmentNodeData, "attachment">`; `CanvasFlowNode = ResourceFlowNode | AttachmentFlowNode` (exported from `CanvasEditor.tsx`, used by Task 6).

Visual components — verified by tsc/lint here and by the Playwright pass in Task 8 (screenshots), not by unit tests.

- [ ] **Step 1: Rewrite `ConnectionEdge.tsx`**

```tsx
import { BaseEdge, EdgeLabelRenderer, getBezierPath, type EdgeProps } from "@xyflow/react";
import {
  EDGE_KIND, EDGE_SOURCE_OF_TRUTH,
  type ConnectionEdgeData, type EdgeKind,
} from "@/pages/stacks/lib/canvas/graph-from-connections";

/** Short chip label per connection kind. */
const KIND_LABEL: Record<EdgeKind, string> = {
  [EDGE_KIND.env]: "env",
  [EDGE_KIND.volumeMount]: "mount",
  [EDGE_KIND.buildArtifactSource]: "build",
  [EDGE_KIND.dependsOn]: "deps",
};

/**
 * Connection edge — an explicit StackConnection (solid brand) or a derived
 * relationship such as depends_on (dashed, muted). A mid-edge chip names the
 * connection kind; a filled dot marks the consuming end.
 */
export function ConnectionEdge({
  id, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, data,
}: EdgeProps) {
  const [path, labelX, labelY] = getBezierPath({
    sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, curvature: 0.5,
  });
  const edgeData = (data ?? { kind: EDGE_KIND.env, sourceOfTruth: EDGE_SOURCE_OF_TRUTH.connection }) as ConnectionEdgeData;
  const derived = edgeData.sourceOfTruth === EDGE_SOURCE_OF_TRUTH.derived;

  return (
    <>
      <BaseEdge
        id={id}
        path={path}
        style={
          derived
            ? { stroke: "var(--fg-muted)", strokeWidth: 1.2, strokeOpacity: 0.5, strokeDasharray: "4 4" }
            : { stroke: "var(--brand)", strokeWidth: 1.4, strokeOpacity: 0.7 }
        }
      />
      <circle cx={targetX} cy={targetY} r={3} fill={derived ? "var(--fg-muted)" : "var(--brand)"} />
      <EdgeLabelRenderer>
        <span
          style={{ transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)` }}
          className="pointer-events-none absolute rounded border border-border bg-card px-1 py-px font-mono text-[9px] uppercase tracking-[0.1em] text-fg-muted"
        >
          {KIND_LABEL[edgeData.kind] ?? edgeData.kind}
        </span>
      </EdgeLabelRenderer>
    </>
  );
}
```

Token check at implementation time: confirm `--fg-muted` is defined in `frontend/src/index.css` (the `text-fg-muted` utility implies it). If the CSS var has a different name (e.g. only exists as a Tailwind theme color), use the closest defined muted token — never a hex literal.

- [ ] **Step 2: Create `AttachmentNode.tsx`**

```tsx
import { memo } from "react";
import { Handle, Position, type Node, type NodeProps } from "@xyflow/react";
import { Archive, HardDrive, KeyRound, type LucideIcon } from "lucide-react";
import { NODE_KIND, type AttachmentKind, type AttachmentNodeData } from "@/pages/stacks/lib/canvas/graph-from-connections";

export type AttachmentFlowNode = Node<AttachmentNodeData, "attachment">;

/** Edges are derived, not hand-drawn, so the handles are anchors only. */
const HIDDEN_HANDLE = { opacity: 0, pointerEvents: "none" as const };

const ICON: Record<AttachmentKind, LucideIcon> = {
  [NODE_KIND.secret]: KeyRound,
  [NODE_KIND.volume]: HardDrive,
  [NODE_KIND.objectStore]: Archive,
};

/**
 * Compact display-only node for connection endpoints that aren't workloads —
 * secrets, volumes, object stores. Same card language as ResourceNode, one
 * line, visually lighter. Not clickable (no drawer).
 */
function AttachmentNodeImpl({ data }: NodeProps<AttachmentFlowNode>) {
  const Icon = ICON[data.kind];
  return (
    <div className="w-[180px] cursor-default rounded-lg border border-border bg-card px-[13px] py-2.5 shadow-xs">
      <Handle type="target" position={Position.Left} style={HIDDEN_HANDLE} isConnectable={false} />
      <Handle type="source" position={Position.Right} style={HIDDEN_HANDLE} isConnectable={false} />
      <div className="flex items-center gap-2.5">
        <Icon className="size-[15px] shrink-0 text-fg-muted" aria-hidden />
        <span className="flex-1 truncate text-[13px] font-medium text-fg-2">{data.name}</span>
        <span className="shrink-0 font-mono text-[9px] uppercase tracking-[0.12em] text-fg-muted">{data.kindLabel}</span>
      </div>
    </div>
  );
}

export const AttachmentNode = memo(AttachmentNodeImpl);
```

- [ ] **Step 3: Register in `CanvasEditor.tsx`**

- Import `AttachmentNode, type AttachmentFlowNode` and export `export type CanvasFlowNode = ResourceFlowNode | AttachmentFlowNode;`
- `const nodeTypes = { resource: ResourceNode, attachment: AttachmentNode };`
- Change `nodes: ResourceFlowNode[]` → `nodes: CanvasFlowNode[]`, `onNodesChange: OnNodesChange<ResourceFlowNode>` → `OnNodesChange<CanvasFlowNode>`, `onNodeClick?: NodeMouseHandler<ResourceFlowNode>` → `NodeMouseHandler<CanvasFlowNode>`.
- Hint copy (bottom panel): `edges carry connection env vars` → `edges show stack connections`.

- [ ] **Step 4: Type check + lint**

Run: `pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint`
Expected: `StackCanvasTab.tsx` may now fail on `ResourceFlowNode` vs `CanvasFlowNode` — acceptable interim ONLY if Task 6 lands in the same session before commit; otherwise adapt `StackCanvasTab`'s `useNodesState` generic in this task. Prefer: fold that one-line generic change (`useNodesState<CanvasFlowNode>`, import from `./CanvasEditor`, and cast updates) into this task so the tree stays green.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/canvas/edges/ConnectionEdge.tsx frontend/src/pages/stacks/components/canvas/nodes/AttachmentNode.tsx frontend/src/pages/stacks/components/canvas/CanvasEditor.tsx frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx
git commit -m "feat(stacks): render connection-kind edges and attachment nodes"
```

---

### Task 6: Wire canvas to projector + topology; retire derive-graph.ts

**Files:**
- Modify: `frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx`
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx` (~L358 draftSync, ~L744 StackCanvasTab)
- Modify: `frontend/src/pages/stacks/components/canvas/nodes/ResourceNode.tsx:5` (import `ResourceNodeData` from `graph-from-connections`)
- Modify: `frontend/src/pages/stacks/lib/canvas/layout-graph.ts:2` (import `CanvasGraph` from `./graph-from-connections`)
- Delete: `frontend/src/pages/stacks/lib/canvas/derive-graph.ts`, `frontend/src/pages/stacks/lib/canvas/tests/derive-graph.test.ts` (superseded by graph-from-connections tests; port any old test case not already covered — dirty-state marking, addon node naming — into `graph-from-connections.test.ts` first)
- Check: `frontend/src/pages/stacks/lib/canvas/tests/layout-graph.test.ts` compiles against the new `CanvasGraph` (edges now require `type`/`data` — update fixtures)

**Interfaces:**
- Consumes: `deriveGraph` (new signature with `secretNameById`), `mergeTopology`, `useStackTopology`, `useSecrets` (`frontend/src/pages/stacks/hooks/use-secrets.ts`, returns `{ secrets: Secret[] }` with `id`/`name`), `CanvasFlowNode` from `./CanvasEditor`.
- Produces: `StackCanvasTabProps` gains `topologyIds: { orgId: string; teamName: string; stackId: string } | null; topologyRefreshKey: number;`.

- [ ] **Step 1: StackCanvasTab.tsx**

In `StackCanvasFlow`:

```tsx
const { secrets } = useSecrets();
const secretNameById = useMemo(
  () => new Map(secrets.filter((s) => s.id && s.name).map((s) => [s.id!, s.name!])),
  [secrets],
);

const { topology } = useStackTopology({ ids: topologyIds, refreshKey: topologyRefreshKey });

const dataGraph = useMemo(
  () => deriveGraph({ resources, linkedAddonIds, addonNameById, secretNameById, dirty }),
  [resources, linkedAddonIds, addonNameById, secretNameById, dirty],
);
const mergedGraph = useMemo(() => mergeTopology(dataGraph, topology), [dataGraph, topology]);
```

Then replace every downstream use of `dataGraph` (`topologySignature`, layout effect, data-update effect, `autoLayout`) with `mergedGraph`. Node state generic: `useNodesState<CanvasFlowNode>` (if not already done in Task 5). The `onNodeClick` guard (`node.data.resourceIdx == null → return`) already ignores attachment nodes — keep as is; add `resourceIdx` narrowing if tsc complains: `const idx = (node.data as { resourceIdx?: number }).resourceIdx;`.

- [ ] **Step 2: detail/index.tsx**

Near the `draftSync` block (~L358):

```tsx
const [topologyRefreshKey, setTopologyRefreshKey] = useState(0);
```

Inside the existing `onStackRefreshed` callback body, after `setStacks(...)`: `setTopologyRefreshKey((k) => k + 1);`

At the `<StackCanvasTab>` site (~L744) add:

```tsx
topologyIds={!isDraft && deployIds.stackId ? deployIds : null}
topologyRefreshKey={topologyRefreshKey}
```

(`deployIds` is the same `{ orgId, teamName, stackId }` object handed to `useDraftSync` — reuse it, do not rebuild.)

- [ ] **Step 3: Update importers, delete old module**

- `ResourceNode.tsx`: `import type { ResourceNodeData } from "@/pages/stacks/lib/canvas/graph-from-connections";`
- `layout-graph.ts`: `import type { CanvasGraph } from "./graph-from-connections";`
- Port still-relevant dirty-state/addon-label test cases from `tests/derive-graph.test.ts` into `tests/graph-from-connections.test.ts`, then `git rm` the old module + test. Update `tests/layout-graph.test.ts` edge fixtures to `{ id, source, target, type: "connection", data: { kind: "env", sourceOfTruth: "connection" } }`.

- [ ] **Step 4: Full frontend gate**

Run: `pnpm --prefix frontend exec vitest run && pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint`
Expected: all green, no remaining references to `derive-graph` (`grep -rn "derive-graph" frontend/src` → empty).

- [ ] **Step 5: Commit**

```bash
git add -A frontend/src
git commit -m "feat(stacks): drive canvas from connections and server topology"
```

---

### Task 7: Playwright MCP E2E pass (spec T1–T8)

**Files:** none (verification). Dev env: backend `mage run` (:8000) + `pnpm --prefix frontend dev` (:5173), Kind cluster from `mage dev:setup`.

Execute the spec's Playwright plan via Playwright MCP tools against `http://localhost:5173`; use `browser_snapshot` for structure, `browser_take_screenshot` for styling, `browser_network_requests` for fetch assertions:

- [ ] **T1** Existing stack (ToolJet template): edge `addon:tooljet-db → resource:tooljet` present with `env` chip, solid brand stroke; all resource + addon nodes render.
- [ ] **T2** Resource with `depends_on`: `deps` chip, dashed muted stroke, visually distinct from authored edges.
- [ ] **T3** Add addon env row in drawer: edge appears immediately, before any topology response.
- [ ] **T4** After autosave flush: exactly one new topology GET; edge count unchanged (dedupe holds); status content intact.
- [ ] **T5** Add volume mount + secret env row: `volume` node + `mount` edge, `secret` node + `env` edge; attachment nodes compact, click opens no drawer; unreferenced secrets/volumes get no nodes.
- [ ] **T6** Remove the T3 env row: edge (and orphaned attachment node) disappears locally and stays gone after the next refetch.
- [ ] **T7** New unsaved draft with two resources + a binding: graph renders; zero `/topology` requests.
- [ ] **T8** Topology failure (block the route via Playwright or use an id that 404s): canvas renders the full local graph, no error UI.
- [ ] Screenshots reviewed against brand tokens (chips legible in dark theme). Fix visual issues, re-run affected checks, then commit any tweaks: `git commit -m "fix(stacks): canvas connection rendering polish"`.
