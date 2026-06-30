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
