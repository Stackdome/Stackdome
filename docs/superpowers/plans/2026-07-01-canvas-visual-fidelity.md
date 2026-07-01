# Plan: Stack Canvas Editor — Visual Fidelity Pass

**Purpose:** The canvas editor is functionally complete (slices 0–6) but visually diverges from the design bundle. This plan closes the gap so the **whole editor + Deployments/Logs/Metrics** match the design. Written to be self-contained after conversation compaction.

**Status:** planning. Predecessor plan (functional build, DONE): `docs/superpowers/plans/2026-07-01-stack-canvas-editor.md`.

---

## 0. Resume Context (read this first after compaction)

- **Worktree / branch:** `.claude/worktrees/stack-canvas-editor`, branch `stack-canvas-editor` (session is already inside this worktree). Branched from `stack-creation-redesign` (has PR #129 wizard + templates). 13 commits already landed for the functional build.
- **Feature flag:** the canvas replaces the Configuration tab only when enabled — `localStorage.setItem('stackCanvas','1')` (per-browser) or `VITE_STACK_CANVAS=true` (build). File: `frontend/src/lib/feature-flags.ts` (`isCanvasEnabled()`). Flag OFF → original form (must stay working).
- **Dev server:** `pnpm --prefix frontend dev` (has been landing on :5174/:5175 — 5173 is a sibling worktree). Proxies `/api`→`:8000` (backend already running).
- **Login:** `admin@stackdome.io` / `welcome@123`. Test stack: **tooljet-addon**, id `d3e497e8-2ec4-4f24-866c-dc6152dee9fa` (3 services: mailhog, redis, tooljet + 1 postgres addon tooljet-db). URL: `http://localhost:<port>/stacks/d3e497e8-2ec4-4f24-866c-dc6152dee9fa`.
- **Commands:** unit tests `pnpm --prefix frontend test:run` (Vitest; DOM tests need a `// @vitest-environment jsdom` pragma). Types `cd frontend && pnpm exec tsc -b --noEmit` (1 PRE-EXISTING unrelated error in `postgres-backups.ts` is expected — `mage build` skips tsc for this reason). Lint `pnpm --prefix frontend exec eslint <paths>`. Baseline: **506 tests pass**.
- **Verify workflow:** drive the app with Playwright MCP on the dev-server port; screenshots into the worktree root (gitignored). Toggle theme to DARK to match the design (the app defaults to light; the design is dark).
- **Design sources:**
  - **Visual spec (authoritative detail):** `docs/superpowers/design-refs/2026-07-01-canvas-editor-visual-spec.md` (extracted from the bundle for this plan).
  - Design-ref (overview + token map): `docs/superpowers/design-refs/2026-07-01-stack-creation-redesign-canvas-design-reference.md`.
  - Bundle zip: `/Users/akshaysasidharan/Downloads/Stackdome Stack Creation Redesign-handoff.zip` (unpack to read `.dc.html`). Key files: `project/Stack Creation Redesign.dc.html` (shell), `ResourceConfig.dc.html` (drawer), `DeploymentsView/MetricsView/LogsView.dc.html`, `EnvRow.dc.html`, `MiniSelect.dc.html`, `_ds/.../colors_and_type.css`.
  - Design screenshots the user compared: **Image #4 = target**, **Image #5 = current**.
- **Decisions log (running):** `docs/superpowers/2026-07-01-stack-canvas-editor-decisions-log.md`.
- **Rules:** brand design system only — `index.css` tokens (`--brand*`, `--fg-2`, `--muted-foreground`, `rounded-sm/md/lg`), `.eyebrow`/`.mono-num` utilities, no raw hex/px in a way that drifts; no magic strings; Grokking-Simplicity + PoSD (pure calcs vs actions). No mention of comparable third-party PaaS in code/commits.

## 1. Current-state file map (what exists)

| File | Role |
|---|---|
| `frontend/src/lib/feature-flags.ts` | `isCanvasEnabled()` |
| `frontend/src/pages/stacks/lib/canvas/derive-graph.ts` | **pure** `deriveGraph(input)→{nodes,edges}`; `NODE_KIND`, `ResourceNodeData` (has `dirtyState`, `volumes`, `summary`, `kind`, `resourceIdx`), `DirtyInput` |
| `frontend/src/pages/stacks/lib/canvas/layout-graph.ts` | **pure** `layoutGraph` (dagre LR, node 216×88) |
| `frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx` | flag-gated entry; derive+layout, topology-vs-data effects (drag-stable), selection, add/remove, drawer host |
| `frontend/src/pages/stacks/components/canvas/CanvasEditor.tsx` | `<ReactFlow>` shell: Background, default `<Controls>`, nodeTypes, snapGrid, Add-resource Panel (top-left), connections toggle (top-right), empty-state |
| `frontend/src/pages/stacks/components/canvas/nodes/ResourceNode.tsx` | memo node card (status dot, glyph, name, kind badge SERVICE/POSTGRES, mono summary, volume chips NAME-only, dirty mark) |
| `frontend/src/pages/stacks/components/canvas/nodes/node-glyph.tsx` | kind→lucide (Box/Database) |
| `frontend/src/pages/stacks/components/canvas/ResourceDrawer.tsx` | slide-in drawer; wraps existing sub-tabs via `useResourceTabProps`; header (dot/glyph/name/status/Modified pill/close); footer (View logs/Remove) |
| `frontend/src/pages/stacks/components/canvas/AddResourcePopover.tsx` | "+ Add resource" → wizard `BlockPicker` |
| `frontend/src/pages/stacks/components/shared/hooks/use-resource-tab-props.ts` | assembles the 3 sub-tab prop sets (shared by accordion + drawer) |
| `frontend/src/pages/stacks/components/detail/index.tsx` | page shell; Configuration `TabsContent` (~683) flag-gates form ↔ `StackCanvasTab`; owns session, PageHeader, top sticky DRAFT/SAVE bar, deploy lifecycle |
| existing ops tabs | `detail/deployments/deployments-tab.tsx`, `detail/logs/stack-logs-tab.tsx`, `detail/metrics/stack-metrics-tab.tsx` |

## 2. Gap analysis (Image #4 design vs Image #5 current)

| # | Area | Design (target) | Current | 
|---|---|---|---|
| A | **Editor shell** | Full-bleed dark editor fills the viewport. Compact top bar: name + READY + "3 services · 1 volume" (left), "3 unsaved changes" + **Deploy** + app-grid icon (right). Canvas edge-to-edge. | Standard app `PageHeader` (big title) + standard Radix tabs + **separate** top sticky "DRAFT / DISCARD ALL / SAVE" bar; canvas in a **bordered box** `h-[calc(100vh-13rem)]`. |
| B | **Tab row** | Configuration **[2]** / Deployments / Logs / Metrics — compact, icons, **count badge** on Configuration, amber active. | Standard tabs, no count badge, no leading icons. |
| C | **Node card** | Kind badge = **WEB / POSTGRES / REDIS** (role/tech); rich mono summary ("node · :8080 · public", "postgres:16 · 10Gi volume", "redis:7 · in-memory"); **volume chip with fill bar + %** (db "pgdata ▓▓ 62%"). | Badge = SERVICE/POSTGRES; summary = image ref only; volume chip = name only (no fill bar). |
| D | **Edges** | **Dashed amber** curved edges. | Solid default grey edges. |
| E | **Controls** | Custom **bottom-left** vertical cluster: +, −, fit, connections-toggle icon. | Default React Flow `<Controls>` bottom-left; connections toggle is a top-right text button. |
| F | **Bottom hint** | "drag to rearrange · click a node to configure · edges carry connection env vars" centered bottom. | none |
| G | **Drawer** | Full-height right **rail**; header sub-line + **"2 changes" pill**; tabs w/ **dirty dots**; **Build from Image/Git toggle** buttons; Repository + branch; **Port + reset**; Depends-on **chips**; Volumes + Add; footer View logs/Remove. | Floating panel inside the boxed canvas; header Modified badge; has fields via reused sub-tabs but styling/spacing differ; no prominent Image/Git toggle; no "N changes" pill. |
| H | **Ops views** | `DeploymentsView` / `LogsView` / `MetricsView` per bundle, inside the full-bleed dark shell. | Existing tabs (already brand-consistent, but not bundle-faithful and inside the standard page chrome). |
| I | **Theme** | Editor is **dark** (cartographic). | Adapts to app theme (shows light when app is light). |

## 3. Open decisions (resolve at kickoff — recommendation first)

1. **Editor takeover extent.** Recommendation: when flag ON and on the Configuration tab (canvas), render a dedicated **full-bleed editor shell** that replaces `PageHeader` + standard tabs + sticky bar with the design's compact top bar + tab row, and makes the canvas edge-to-edge. Keep the app left sidebar. Deployments/Logs/Metrics render inside the same shell. (Alt: keep current chrome, only restyle — rejected; won't match "fit all the dashboard".)
2. **Dark-force.** Recommendation: the **canvas editor surface is always dark** (wrap in `.dark` scope), matching the bundle's cartographic aesthetic, regardless of the app's light/dark toggle. Rest of app unchanged. (Alt: respect theme — but design is dark-first; user's images are dark.)
3. **Ops views approach.** Recommendation: **restyle the existing** `deployments-tab`/`stack-logs-tab`/`stack-metrics-tab` to the bundle's `*View` designs (keep their data hooks), rather than new components — preserves working data wiring. Read the bundle `*View.dc.html` for exact layout.
4. **Kind-badge + summary heuristic.** Need a small pure mapper `image/kind → { kindLabel, summary }` (WEB for web/node services, REDIS/POSTGRES/etc for known images, port/public/volume/in-memory hints). Unit-test it.

## 4. Vertical slices (each Playwright-verified in DARK mode)

> Order by visual impact. FS-1 first (the shell is the biggest gap).

- **FS-1 — Full-bleed dark editor shell.** New `CanvasEditorShell` (or restructure `detail/index.tsx` under the flag): compact top bar (name, status pill, "N services · M volumes", "N unsaved changes" + Deploy, app-grid icon), tab row with Configuration **count badge** + icons + amber active, canvas edge-to-edge, dark-scoped. Wire "N unsaved changes"/Deploy to existing `session.dirty` + `useDeployLifecycle`/`handleSave` (do NOT rebuild deploy logic). Keep flag-OFF path pristine. **Demo:** editor fills viewport, dark, matches Image #4 header/tabs.
- **FS-2 — Node card fidelity.** Pure `nodePresentation(resource|addon) → { kindLabel, summary, volumeFill }` (TDD). Update `ResourceNode`: kind badge (WEB/REDIS/POSTGRES…), rich summary, **volume chip fill bar + %**, status-dot color by state. **Demo:** nodes match Image #4 (web/db/cache).
- **FS-3 — Dashed amber edges.** Style edges (custom edge type or `style`/`animated`): dashed, `var(--brand)`, curved. **Demo:** edges match.
- **FS-4 — Custom controls + hint.** Replace default `<Controls>` with a bottom-left cluster (zoom in/out/fit/connections-toggle) per spec; add centered bottom hint text. **Demo:** controls + hint match.
- **FS-5 — Drawer fidelity.** Full-height rail styling; header sub-line + "N changes" pill (from `session.dirty.perResourceDirty`); tab dirty dots (already have via `useResourceTabProps.dirtyTabs`); ensure Build-from Image/Git toggle, Repository+branch, Port+reset, Depends-on chips render per spec (these live in the reused sub-tabs — restyle there if needed, but keep the accordion working). **Demo:** drawer matches Image #4 right rail.
- **FS-6 — Deployments view** to bundle `DeploymentsView.dc.html`.
- **FS-7 — Metrics view** to bundle `MetricsView.dc.html`.
- **FS-8 — Logs view** to bundle `LogsView.dc.html`.

## 5. Guardrails
- Reuse the working engine: `useStackEditSession`, `stack-diff`, `useDeployLifecycle`, `performSave`/`handleSave`, `useResourceTabProps`, `deriveGraph`/`layoutGraph`. This pass is **presentation only** — no changes to the save/PUT contract or session model.
- Flag-OFF must stay identical to today. After each slice: `tsc` (expect the 1 pre-existing error), `eslint` clean on touched files, `test:run` green (506+), Playwright screenshot in dark mode.
- Keep pure calcs pure (node presentation mapper, any layout tweaks) and unit-tested.
- Append every judgment call to the decisions log.

## 6. Not in this plan
Node-position persistence, draw-edges-on-canvas authoring, worker/cron nodes (backend), orphan-volume cleanup on remove — all remain deferred (see functional plan + decisions log).
