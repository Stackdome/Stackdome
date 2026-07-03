# Canvas connections from Connections API — Design

**Date:** 2026-07-04
**Status:** Approved (brainstormed with user)
**Scope:** Stack canvas editor edge/node rendering (frontend)

## Context

The stack canvas editor derives edges indirectly today: `spec.connections` is hydrated into form env rows (`connectionsToEnvRows`), then `derive-graph.ts` walks env rows (`from === "addon" | "resource"`) to infer edges. Connection identity, kind, mappings, and config are lost in that translation, and non-env connection kinds (volume mounts, build artifact sources) plus non-resource node types (secrets, volumes, object stores) never reach the canvas.

The backend has a first-class connections API (`GET/POST/PUT/DELETE .../stacks/{id}/connections`; schemas `StackConnection`, `ConnectionMapping`, `StackConnectionConfig`) **and** a topology endpoint (`GET .../stacks/{id}/topology`, OpenAPI L1535) returning `TopologyNode[]` / `TopologyEdge[]` with `source_of_truth: connection | derived` and edge kind `env | volume_mount | build_artifact_source | depends_on`. The frontend already calls connection CRUD via draft-sync; generated types (`openapi.d.ts`, `zod-schemas.ts`) are up to date — no regeneration needed.

**Goal:** the canvas renders the full topology faithfully from connection semantics.

## Decisions

1. **Hybrid, two producers:** local projector for instant authored edges + topology fetch for server truth.
2. **Full topology scope:** all node types (`stack_resource`, `addon/postgres`, `secret`, `volume`, `object_store`), all edge kinds (`env`, `volume_mount`, `build_artifact_source`, `depends_on`).
3. **Render-only edges:** kind + source_of_truth styling; no click-to-inspect, no on-canvas authoring.

## Architecture

```
form state ──▶ desired connections ──▶ local projector ──▶ authored edges + nodes
(instant)      (buildDesiredConnections,                    + depends_on edges (derived)
                mountsToConnections)

autosave flush ──▶ GET /stacks/{id}/topology ──▶ server nodes/edges
(use-draft-sync)                                   │
                                                   ▼
                                    merge ──▶ canvas graph ──▶ dagre layout ──▶ render
```

Merge rules (each datum has one owner):

- Authored edges (`source_of_truth: connection`): **local wins** — instant while typing.
- Derived edges: **union** keyed `(kind, source, target)` — local computes `depends_on` from `spec.stack_resources[].depends_on` (already in frontend data); server-derived edges merged in, dupes dropped.
- Node runtime `state`: server topology overlays onto matching local nodes by `ref`.
- Unsaved draft (no stack id): local producer alone, fetch disabled.
- Topology refetch: after each draft-sync flush completes.

Env-row edge inference in `derive-graph.ts` is removed, replaced by connection projection.

## Components

New:

- `frontend/src/api/topology.ts` — `getStackTopology` client over the generated zod endpoint; re-exports `TopologyNode` / `TopologyEdge` types.
- `frontend/src/pages/stacks/hooks/use-stack-topology.ts` — query hook; enabled only for persisted stacks; refetch keyed on draft-sync flush completion.
- `frontend/src/pages/stacks/lib/canvas/graph-from-connections.ts` — pure projector `(formResources, desiredConnections, dependsOn) → {nodes, edges}` in `TopologyNode`/`TopologyEdge` shape. Reuses `buildDesiredConnections` / `mountsToConnections` from `connection-mapping.ts`.
- `frontend/src/pages/stacks/lib/canvas/merge-topology.ts` — pure merge per rules above.

Changed:

- `derive-graph.ts` — env-row walking removed; delegates to projector (or file retired into it).
- `StackCanvasTab.tsx` — wires hook + merge; `topologySignature` computed from merged graph.
- `canvas/nodes/` — new compact display-only node components for `secret` / `volume` / `object_store` (icon + name + type tag; brand card language, lighter than addon node; not clickable; rendered only when referenced by an edge). `ResourceNode.tsx` unchanged except server `state` overlay on status dot.
- `edges/ConnectionEdge.tsx` — `connection` → solid, `derived` → dashed muted; kind chip label mid-edge (`env` / `mount` / `build` / `deps`); colors via `index.css` tokens only; edge id from connection id when present for stable React Flow keys.

Untouched: draft-sync CRUD path (`use-draft-sync.ts`, `draft-sync/ops.ts`, `draft-sync/desired-state.ts`) — already connection-native.

## Error handling

Topology fetch failure → silent fallback to local graph; canvas never blocks on it; retried on next flush. No error banner (enhancement layer).

## Testing

### Unit (vitest)

`graph-from-connections.ts` + `merge-topology.ts` (pure fns): authored/derived precedence, dedupe key `(kind, source, target)`, draft-without-id, state overlay, referenced-only node inclusion, stable edge ids from connection ids.

### Playwright MCP E2E plan (`http://localhost:5173`)

Use `browser_snapshot` for structure, `browser_take_screenshot` for styling, `browser_network_requests` for fetch assertions.

- **T1 — Existing stack renders authored edges.** ToolJet-template stack, canvas tab: edge tooljet→tooljet-db with `env` chip, solid stroke; all resource + addon nodes present.
- **T2 — depends_on renders as derived.** Resource with `depends_on`: edge with `deps` chip, dashed/muted, distinct from authored.
- **T3 — Instant local edge.** Add env binding in drawer: edge appears before any topology response (local producer).
- **T4 — Refetch after autosave, no dupes.** After flush: exactly one topology GET; edge count unchanged; status dots correct.
- **T5 — Volume/secret/object-store nodes.** Mount + secret binding: `volume` node + `mount` edge, `secret` node + edge; compact display-only variant; unreferenced secrets/volumes get no nodes.
- **T6 — Edge removal.** Delete binding: edge + orphaned node disappear locally and stay gone after refetch.
- **T7 — Unsaved draft local-only.** New draft, two resources + binding: canvas renders; zero `/topology` calls.
- **T8 — Topology failure fallback.** Simulated 500: full local graph renders, no error banner, no blank canvas.

Pass criteria: all assertions green + screenshots reviewed against brand tokens (no raw hex, chips legible in dark theme).
