# Stack Canvas Editor — Autonomous Run Decisions Log

Running the plan (`docs/superpowers/plans/2026-07-01-stack-canvas-editor.md`) end-to-end without review pauses (user asleep). Every judgment call / plan deviation logged here with rationale, so it can be reviewed/reverted in the morning.

Format: `[slice] decision — why — impact/how-to-revert`.

---

## Slice 0 (already reviewed)
- **Skipped Task 0.2 token additions.** Reading the full `index.css` showed existing tokens already cover the design (radius `rounded-sm/md/lg` = 2/4/8px; 4-tone fg ramp `foreground`/`fg-2`/`muted-foreground`/`fg-muted`; brand bg/border; `.eyebrow`/`.mono-num`). Adding `--r-*`/`--s-*`/`--fg2/3` would have been duplicate/drift. Impact: zero `index.css` change in slice 0. React-Flow CSS-var theming deferred to slice 1.
- **Per-file `// @vitest-environment jsdom` pragma** required for the feature-flag test (vitest default env is node; DOM/localStorage tests opt in per-file — repo convention). Not a deviation, just matched the existing pattern.
- **Dev server on :5175** (5173/5174 held by sibling worktrees). Proxies `/api`→`:8000` (running backend).

---

## Slice 1 — read-only graph
- **Edges derive from typed env-var refs only** (`from:"addon"|"resource"`). Consequence on the live tooljet stack: `redis` shows **no edge** because tooljet references it via a plain env value, not a typed resource ref. This is correct given the data model, not a bug. If we later want redis wired, the connection must be a typed ref. (Verified visually.)
- **`kindLabel` simplified to SERVICE / POSTGRES.** The design uses per-image labels (WEB/REDIS/…). Deferred as polish — needs an image→label heuristic. Low risk.
- **Status dot is always green (success) for now.** Real status→colour mapping deferred to slice 3 (where node status/dirty marks land). `resource.status` is `unknown`-typed; needs a safe extractor.
- **Volume chips show name + mountPath, no fill bar.** The design's size/percent fill bar needs the joined `Volume.spec` size; deferred to slice 4 (volumes). Chip tooltip shows mountPath.
- **`StackCanvasTab` takes resolved view inputs** (`resources`/`linkedAddonIds`/`addonNameById`) rather than the whole session — keeps slice 1 read-only and the component a pure view. The `session` (for drawer/editing) is threaded back in at slice 2.
- **Two TS casts** added: volume-chip `name`/`mountPath` (source types are `unknown` from the zod passthrough) and `graph.nodes/edges as ResourceFlowNode[]/Edge[]` (CanvasNode/CanvasEdge are structurally compatible). Pragmatic, localized.

---

## Slice 2 — node drawer
- **Drawer is self-sufficient.** It sources its own `secrets`/`addons`/`addonNameById`/`allResources` (same hooks the form uses) and wires discard straight to the session (`discardResourceField`/`discardEnvRow`). So `detail/index.tsx` only passes `session` + baselines + `connectionAddonIds` + `errors` — minimal plumbing.
- **`useResourceTabProps` refactor verified safe.** Playwright-checked the flag-OFF accordion form (edit + env grouping render identically). DRY win: drawer and accordion share one wiring.
- **Addon-node click = no drawer (v1).** Addons are configured via a resource's Environment tab bindings; an addon node has no `resourceIdx`, so clicking it is a no-op. Logged as a deliberate v1 limitation.
- **Clicking a node lazily activates the session** (`session.start(baseline, …)`) so drawer edits land in a draft.
- **Verified no index bug:** editing tooljet's image updated ONLY tooljet (redis/mailhog untouched), drawer matched the selected node, Modified badge + "1 RESOURCE MODIFIED" appeared, node summary live-updated. (An earlier screenshot mismatch was a Playwright interaction artifact — selection had moved to redis — not a data bug.)
- **Known follow-up:** after "Discard all", the drawer does not auto-close (selectedIndex persists). Harmless (next click/reload fixes it); will close-on-discard when wiring deploy in slice 3.

---

## Slice 3 — node marks + layout stability + save (pragmatic re-scope)
- **No new `CanvasHeader` built.** The existing page header / tabs / sticky action bar / deploy lifecycle ALREADY work with the canvas (same session) — verified "N RESOURCE MODIFIED" + SAVE on canvas edits. Rebuilding a header would duplicate working chrome and risk the flag-off path. The bundle's header restyle (inline "N unsaved changes"+Deploy, per-tab count badge) is cosmetic → deferred to polish (slice 6) as a tweakable. Slice 3 instead delivers the real canvas gaps below.
- **Re-layout only on topology change (drag-stability fix).** Previously every edit changed `session.draft.resources` identity → re-memoised graph → re-ran dagre → reset positions (drag lost; layout re-run per keystroke). Fixed by splitting into (a) a topology-signature effect that re-lays-out only when the node/edge id-set changes, preserving existing drag positions by id, and (b) a data effect that updates node `data` (summary/dirty mark) in place without moving nodes. Deterministic dagre meant no *visible* reshuffle before, but drag was being lost and layout re-ran needlessly.
- Verified: edit → EDITED mark + amber border; Save → existing PUT persists and clears the mark (tested round-trip then restored the original image, no Deploy → backend stack untouched).

---

## Slice 4 — add / remove
- **Add reuses the wizard pipeline** (`addBlockToStack` + `getBlockById` + `blockCatalog` + `BlockPicker`) — DRY with the create flow, incl. unique-naming and volume rewiring. New node renders the NEW mark; verified adding MySQL + removing it.
- **Connections toggle moved to top-right** to make room for "+ Add resource" top-left (matches design).
- **One TS cast** at the volumes boundary: block-to-form yields base `FormVolumeData`, the session works in `FormVolumeExtendedData` (adds optional `sourceType`) — `as unknown as` at the single call site.
- **Known limitation:** removing a resource does NOT remove its orphaned volumes (volumes are managed separately, same as the form). Acceptable for v1; note for a future volumes pass.

---

## Slice 5 — ops views (deliberately light)
- **Did not rebuild Deployments / Logs / Metrics.** They already exist as sibling Radix tabs (`DeploymentsTab`/`StackLogsTab`/`StackMetricsTab`), work flag-independently, and are already on the brand design system. The bundle's `DeploymentsView`/`LogsView`/`MetricsView` are flagged "secondary, not part of the canvas flow" in the design-ref, and their detailed markup was NOT extracted into the design-ref. Rebuilding them pixel-faithfully would be guesswork and risks breaking working operational UIs.
- **Decision:** verify the three tabs render correctly with the canvas, apply only obvious brand-consistency fixes, and DEFER the pixel-faithful redesign as a tweakable follow-up that needs the bundle's `*View.dc.html` markup re-ingested. This matches the user's "we may tweak aspects later" and keeps the primary deliverable (canvas) the focus.
- **Verified:** Deployments (timeline + staged-changes diff + StageTracker + resource outcomes), Logs (streams), Metrics (Stack Overview + per-resource cards) all render correctly with the canvas and are already brand-consistent. No code change in this slice.

---

## Slice 6 — hand-off + polish
- **Wizard hand-off is flag-controlled — no composer change.** The canvas IS the flag-gated show page, so a stack created via the PR #129 composer lands on `/stacks/:id` and renders the canvas when the flag is on. Modifying the shipped composer to "navigate into a canvas" would exceed the show/edit-page scope (D1) and touch shipped code. So the hand-off works via the global flag; the composer is untouched.
- **Empty-state** added (centered prompt when a stack has 0 resources) and **snap-to-grid** (16px) for tidy drag. Verified the populated canvas still renders (4 nodes / 2 edges, empty-state hidden).
- **Perf:** `nodeTypes` module-scope ✓, `ResourceNode` memoized ✓, derivation/layout pure + memoised ✓, re-layout only on topology change ✓. `onlyRenderVisibleElements` intentionally NOT enabled — the research notes it hurts small/medium graphs (re-init cost); our stacks are small.

---

## Run summary (for morning review)
- **Slices 0–6 complete**, each committed and Playwright-verified on `:5175` against the running backend. Frontend: tsc clean (1 pre-existing unrelated `postgres-backups.ts` error), eslint clean on canvas code, **506 unit tests pass** (+24 new canvas/flag/hook tests).
- **Backend: zero changes** (D2 held throughout).
- **The tooljet-addon stack was left untouched** — every Save during verification was reverted; no Deploy was triggered.
- **NOT done (awaiting you):** no push / PR / merge (outward-facing — needs your go-ahead). Branch `stack-canvas-editor` holds all commits in the worktree.
- **Deferred/tweakable** (all logged above): pixel-faithful ops-view redesign (needs bundle markup), per-image kind labels (WEB/REDIS vs SERVICE), volume fill-bar + orphan-volume cleanup on remove, status→dot-colour mapping, draw-edges-on-canvas authoring, worker/cron nodes (backend), node-position persistence, header restyle to bundle (inline unsaved-count + per-tab badge).

---

# Visual-Fidelity Pass — Decisions Log

Executing `docs/superpowers/plans/2026-07-01-canvas-visual-fidelity.md` (slices FS-1..FS-8). Same running-log format.

## §3 open decisions (confirmed at kickoff)
1. **Editor takeover = full-bleed shell** (user-confirmed). Flag-ON + Configuration renders a dedicated shell replacing PageHeader + Radix tabs + sticky bar with the design's compact top bar + icon tab row; canvas edge-to-edge; app left sidebar + breadcrumb kept.
2. **Dark-force → OVERRIDDEN by user: follow the app theme.** No `.dark` wrapper. The shell uses brand tokens (`--brand*`, `--fg-*`, `--muted-foreground`, `--bg*`), which already carry both light and dark values in `index.css` — so it renders like the dark reference when the app is dark, and adapts to light otherwise. Cleaner than a forced scope and avoids inventing unspecified light-canvas values.
3. **Ops views = restyle existing** `deployments-tab`/`stack-logs-tab`/`stack-metrics-tab` (keep data hooks). (user-confirmed)
4. **Kind-badge/summary heuristic** = a pure, unit-tested mapper built in FS-2 — no user confirmation needed.

## FS-1 — Full-bleed editor shell
- **New `CanvasEditorShell.tsx`** (presentational; owns no stack state). Renders stack-title header (name 29px, status pill, subtitle, dirty summary, primary button, overflow menu) + icon tab row (Configuration count badge, amber active) + full-height mode body. Wired straight to the caller's `session.dirty` + `handleSave`/`onDeploy` — no save/deploy logic inside.
- **Full-bleed via `app-layout.tsx` conditional.** The centered `max-w-6xl p-6` wrapper is dropped (→ `h-full`) only when `isCanvasEnabled()` AND the route matches `/^\/stacks\/[^/]+$/` and isn't `/new`. Touches global chrome, but tightly gated: flag-OFF and every other route keep the exact original wrapper. Needed because the editor can't go edge-to-edge inside the max-width column.
- **Primary button is context-aware: Save when dirty, Deploy when staged/clean.** The bundle mock shows a single always-"Deploy" alongside "N unsaved changes", but this backend keeps *save* (PUT the draft) and *deploy* (create a release) separate, and the plan forbids rebuilding deploy logic. So: unsaved edits → **Save** + amber "N unsaved changes"; saved-but-undeployed (`lifecycle.phase==="staged"`) → **Deploy** + muted "draft saved · undeployed" hint; clean → **Deploy**. Preserves the two-step model exactly.
- **Amber CTA reuses `Button variant="default"`** (`bg-brand text-white`) rather than the mock's raw `#1a0e05` near-black-on-amber label. Stays on the established brand primitive / no raw hex (brand-system rule). Minor visual delta (white vs near-black label text).
- **Discard-all moved into the overflow (⋯) menu** with its own `AlertDialog` confirm (parity with the old sticky-bar confirm-threshold). Edit + Delete also live in that menu. The Configuration tab count badge = `session.dirty.dirtyResourceIdx.size`; the "N unsaved changes" total = resources + volumes + addon links (mirrors the sticky-bar segments).
- **DRY:** the three ops-view bodies (Deployments/Logs/Metrics) + the detach `AlertDialog` were extracted to local variables in `detail/index.tsx`, reused by both the flag-ON shell and the flag-OFF page (single place to restyle in FS-6/7/8). The dead `isCanvasEnabled()` ternary inside the flag-OFF Configuration `TabsContent` was removed (that path is now only reached when the flag is off).
- **`StackCanvasTab` box → `h-full w-full` edge-to-edge** (dropped `h-[calc(100vh-13rem)] rounded-md border`); its only consumer is the flag-ON canvas.
- In the shell, non-canvas modes get a temporary `overflow-auto px-7 py-6` wrapper; they are restyled to the bundle in FS-6/7/8.
- **Verified (dark, Playwright):** shell fills viewport & matches Image #4 (top bar + tab row + edge-to-edge dot-grid canvas); editing a node → "1 unsaved change" + **Save** + Configuration badge "1" + node EDITED mark + live summary; overflow → Discard all → confirm dialog reverts cleanly (no PUT, backend untouched); **flag-OFF reload is pristine** (PageHeader + sticky bar + Radix tabs + Panels + max-width column). tsc clean (1 pre-existing `postgres-backups.ts` error), eslint clean on touched files, **506 tests pass**.
- **Known (deferred to later FS slices, visible in FS-1 screenshots):** default React-Flow `<Controls>` renders as a light box in dark mode (→ FS-4 replaces it); "Hide connections" is still a top-right text button (→ FS-4 cluster); edges are solid grey (→ FS-3 dashed amber); node summaries are image-only + badge SERVICE/POSTGRES (→ FS-2); drawer still uses the accordion sub-tabs styling (→ FS-5).

## Perf check (chrome-devtools, between FS-1 and FS-2)
- Profiled the live canvas (dark, tooljet-addon). **Dev build** — absolute load numbers inflated by Vite dev + unminified + React dev-mode.
- **Load:** LCP 1114ms (97% render-delay = dev module eval, not app code), CLS 0.00 (no layout shift). **Mount:** one ForcedReflow 68ms inside `@xyflow/react`'s own `updateDimensions`/`getDimensions` — not our code, DevTools "estimated savings: none".
- **Edit (12 keystrokes):** 0 long tasks, and every node transform stayed byte-identical → editing does NOT re-run dagre. The topology-vs-data effect split holds (the core render-perf design). **Drag:** INP 55ms (good), CLS 0, only the dragged node moves, session stays clean (positions view-only, not persisted).
- Verdict: healthy for our small graphs; no hotspot in canvas code. `onlyRenderVisibleElements` correctly OFF. No changes made.

## FS-2 — Node card fidelity
- **New pure `node-presentation.ts`** — `nodePresentation({ isAddon, image, hasBuild, ports }) → { kindLabel, glyph, dotState, summary }` (9 unit tests). Heuristic (backend has no kind field): image-substring → kind (redis/postgres/mariadb|mysql/mongo/minio); a generic image with a **public** port reads as **Web**, otherwise **Service**; addon → **Postgres / managed postgres**. Summaries: `redis:6.2 · in-memory`, `postgres:16`, `minio · S3-compatible`, `web-api · :8080 · public` (base = registry/tag stripped), etc.
- **Computed in `deriveGraph` (data layer)**, stored on `ResourceNodeData` (`kindLabel`/`glyph`/`dotState`/`summary`); `ResourceNode` is a pure view. Follows GS calculation→data / PoSD.
- **`node-glyph` extended:** web=Globe, postgres/database=Database, redis=Zap, object=Cloud, worker=Cpu, service=Box. Signature changed `kind`→`glyph` (updated the drawer's call; still Box there — FS-5 makes the drawer glyph kind-aware).
- **`ResourceNode` restyled to spec:** 8px status dot coloured by `dotState` (ok/warn/err), 17px glyph (`text-fg-2`), kind badge (mono 9px uppercase tracked, `text-fg-muted`), indented mono summary, volume rows (drive glyph + name, hairline top border).
- **Dirty visual changed (deliberate):** dropped the NEW/EDITED/REMOVED **text label** (it occupied the badge slot). The kind badge is now **always** shown; unsaved state is signalled by a **left accent stripe** (amber for new/edited, crimson for removed) + tinted border, and selection by an amber border + ring halo — matching the design (counts still surface in the top bar + drawer). Verified: editing keeps the WEB badge and adds the amber stripe/border/halo.
- **Volume fill bar + % DEFERRED:** the design's usage bar needs *runtime* volume usage, which the editor draft doesn't have (only requested `volumes[].spec.size` exists; usage is a metrics-path value). Volume chip restyled to the design's row format (drive glyph + name); size/usage bar deferred. The test stack has 0 volumes, so the row is visually unverified.
- **Verified (dark, Playwright):** all four nodes match Image #4 — redis (Zap/REDIS/"redis:6.2 · in-memory"), tooljet-db (Database/POSTGRES/"managed postgres"), tooljet + mailhog (Globe/WEB/"…· :port · public"); dirty node shows the amber stripe+border+halo with badge retained. tsc clean (1 pre-existing), eslint clean, **515 tests pass** (+9 presentation).

## FS-3 — Dashed amber edges
- **New `ConnectionEdge` custom edge type.** `getBezierPath` with `curvature: 0.5` (control handles at the horizontal midpoint = the design's `mx` symmetric curve), `BaseEdge` styled `stroke:var(--brand); stroke-width:1.4; stroke-opacity:0.55; stroke-dasharray:"5 4"`, plus a filled `<circle r=3 fill=var(--brand)>` at the target = the end dot. No animation (per spec).
- Registered `edgeTypes = { connection }` at module scope and `defaultEdgeOptions = { type: "connection" }` on `<ReactFlow>` — presentation stays in the view; `deriveGraph` edges (`{id,source,target}`) are unchanged. Connections still toggle off via `showConnections` (edges → `[]`).
- **Verified (dark, Playwright):** 2 edges, computed stroke = `--brand`, width 1.4px, dasharray `5 4`, opacity 0.55, 2 end dots — dashed amber curved edges match Image #4.

## FS-4 — Custom controls + hint
- **New `CanvasControls`** (bottom-left `Panel`, drives the pane via `useReactFlow`): a zoom pill column (in `plus` / out `minus` / fit `maximize`, 40px cells, `bg-popover` = the design's `--bg-elev`, hairline dividers, hover `text-brand`) over a separate connections-toggle square (amber border+text when connections shown, neutral otherwise; `workflow` glyph). Replaces React Flow's default `<Controls>`, which rendered as a **light box** in dark mode.
- **Bottom-center hint** (`pointer-events-none` Panel, only when nodes>0): `move` glyph + `"drag to rearrange · click a node to configure · edges carry connection env vars"` (`text-fg-muted`, 11.5px).
- Removed the top-right "Hide connections" text button (the toggle now lives in the cluster). Background tuned to `Dots` gap 24 / size 1 (the design's 24px dot grid).
- **Simplified (logged):** the design's connections **popover** (a 2-option menu: variable references / hide) is not built — the toggle flips edges on/off directly, as before. Popover deferred as polish; low value.
- **Verified (dark, Playwright):** cluster + hint render; toggle hides/shows edges and flips amber↔neutral; zoom-out changed the viewport scale 2→1.667 (zoom-in was capped at fitView's maxZoom 2); the light Controls box is gone. tsc clean, eslint clean.

## FS-5 — Drawer chrome fidelity
- **Restyled the `ResourceDrawer` chrome to the bundle rail:** width 380 → **496px**, `bg-background` (design `--bg`). Header now: 9px status dot, **kind-aware amber glyph** (via `nodePresentation`), name (16px), **summary sub-line** (mono, e.g. "tooljet-ce · :80 · public"); when dirty a **"N changes" pill** (amber, with an `×` that calls `session.discardResource`) else a `.t-marker` **kind label**; close `×`. Sticky **mono-uppercase tab bar** (CONFIGURATION/DEPLOYMENT/ENVIRONMENT, 1.5px amber underline) with per-tab dirty dots; scroll body; footer View logs / Remove.
- **"N changes" = count of dirty sub-tabs** (config/deployment/environment, 1–3) — a proxy for the design's field-level count (a per-field diff count isn't readily available from `useResourceTabProps`). Logged.
- **Inner field layout kept as the shared sub-tabs (accordion-styled) — deliberate.** The bundle's compact ResourceConfig body (Build-from Image/Git toggle buttons, Repository+branch, Port+reset, Depends-on chips, volume fill bars, dashed addon-group boxes) lives in `stack-resource-{configuration,deployment,environment}-tab.tsx`, which **also drive the flag-OFF accordion**. A pixel restyle there is a separate, larger effort that risks the accordion, so FS-5 delivers the drawer **frame** fidelity and leaves the fields functional. Deferred (needs a shared-component pass or a canvas-specific fork).
- Only `ResourceDrawer.tsx` changed (canvas-only) + an import of `nodePresentation`; no shared sub-tab component touched, so the flag-OFF path is unaffected by construction.
- **Verified (dark, Playwright):** drawer chrome matches Image #4's right rail (amber kind glyph + summary sub-line + WEB label); editing → "1 change" pill + `×` + Configuration dirty dot; the pill `×` reverts cleanly (image back to `:latest`, top bar clean, no PUT). **Flag-OFF regression clean** (canvas off; STACK RESOURCES/VOLUMES/ADDONS panels + Radix tabs + resources render). tsc clean, eslint clean, full suite still 515.

## FS-6/7/8 — Ops views restyled to the bundle (Deployments / Logs / Metrics)
- **Shell ops-body padding removed** — each view now owns its max-width + padding per the design (Deployments 920px, Metrics 1000px, Logs 1100px, all `mx-auto`). These three components render in BOTH flag paths (same tabs), so the restyle applies to flag-OFF too (adapts via tokens).
- **FS-7 Metrics (largest rebuild):** header "Stack metrics" + LIVE/status pill + "updated <time>"; two summary cards (STACK CPU / STACK MEMORY — 30px number + unit + **sparkline from a rolling 16-sample window** accumulated off the stream = real data, amber/`fg-2` bars); PER RESOURCE auto-fill grid (dot + glyph + name + READY marker, CPU bar amber / Memory bar `fg-2` with labelled values). **Bars are relative to the peer max** — the metrics API gives raw usage with **no per-resource limit**, so a true % isn't derivable; the absolute value is always shown next to the bar. Per-resource status is hardcoded READY (metrics carry no state). Verified live: 3m / 385 MiB, relative bars, building sparklines.
- **FS-8 Logs:** header "Stack logs" + CONNECTED pill + restyled Resources / time-range filters; the stream is now a **near-black `#070a0f` terminal panel** (bordered, rounded) with LazyLog inside (muted mono 12px/1.7, line numbers + `[resource]` tags + built-in search/matches toolbar). **Kept LazyLog** (virtualised) rather than a bespoke renderer — so the design's per-line tag/message colour coding isn't applied (LazyLog renders raw text) and the panel is a fixed 560px (LazyLog needs an explicit height). Logged; both are acceptable trade-offs for stream perf. Verified live streaming.
- **FS-6 Deployments:** centered to `max-w-[920px]` with the design padding + `.t-marker` "DEPLOY TIMELINE" eyebrow. The timeline internals — rail, green/crimson dots, collapsed rows + status chips, expandable **step strip**, crimson **failure banner**, **RESOURCE OUTCOME** list, staged-diff **draft node** — were already structurally faithful + brand-consistent, so left as-is (a deeper `timeline/*` restyle is deferred; low marginal gain). Skipped the design's "Deploy now" button (Deploy lives in the top bar). Verified.
- tsc clean (1 pre-existing), eslint clean, **515 tests pass**. Backend untouched (the staged "redis changed" draft visible in Deployments is pre-existing stack state, not created this session).

## Post-review canvas UX tweaks (user feedback)
Three asks after reviewing the canvas: auto-layout, over-zoom, drawer-push.
- **Zoom cap.** New shared `canvas/fit-options.ts` = `{ maxZoom: 1, padding: 0.25 }`, applied to `<ReactFlow fitViewOptions>`, the Fit control, auto-layout, and the drawer refit. Small graphs no longer over-zoom — initial fit was `scale(2)` (React Flow's default maxZoom), now `scale(1)` (nodes at natural size).
- **Auto-layout button** (`Wand2` glyph, added to the controls cluster above the connections toggle). Re-runs the pure `layoutGraph` (dagre), resets every node to its computed position, then refits. Verified: a node dragged to (-160,192) snaps back to its dagre (24,24) on click.
- **Drawer is a side panel, not an overlay.** `StackCanvasFlow` now renders a flex row — canvas `flex-1 min-w-0`, drawer a `496px shrink-0` sibling (dropped its `absolute right-0`). Opening/closing the drawer refits the graph into the changed width (80ms settle for React Flow's ResizeObserver). Verified: nodes stay fully visible beside the drawer (Railway-style), and closing refits back to full width.
- All changes are canvas-only files (`StackCanvasTab`, `CanvasEditor`, `CanvasControls`, `ResourceDrawer`, new `fit-options.ts`) → flag-OFF unaffected. tsc/eslint clean, 515 tests pass.

## Post-review drawer tweaks (user feedback)
- **Remove resource is red by default** (footer button `text-danger` + red trash icon), was muted-until-hover. User preference over the design's muted default.
- **"View logs" wired to the Logs view** — new `onViewLogs` threaded page → `StackCanvasTab` → `ResourceDrawer`, calling `setActiveTab("logs")`. Button enabled (was a disabled "Coming soon"). Pre-filtering the Logs view to the clicked resource is deferred (matches the existing Deployments→logs behaviour, which also just switches tabs).
- **Drawer survives editor-tab switches.** The shell no longer swaps the mode body — it keeps the canvas (Configuration) **always mounted** (`absolute inset-0`) and renders the active ops view as an **opaque overlay** on top. So the open drawer + node selection (and drag positions) persist across Configuration↔Logs/Deployments/Metrics. Verified: open drawer on tooljet → View logs (Logs tab) → back to Configuration → drawer still open on tooljet. Bonus: React Flow is never `display:none`, so it stays measured (no re-fit glitch). Ops streams (logs/metrics) still mount only while their tab is active.
- Flag-ON-only surfaces (`CanvasEditorShell` + the canvas drawer) + one `onViewLogs` prop on the flag-ON `StackCanvasTab` render — flag-OFF unaffected. tsc/eslint clean, 515 tests pass.
