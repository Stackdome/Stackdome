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
