# Design Reference: Stack Creation Redesign — Canvas Editor + Full Create Flow

**Bundle:** `stackdome-stack-creation-redesign`
**Ingested:** 2026-07-01
**Primary file:** `project/Stack Creation Redesign.dc.html` (98 KB)
**Related PR:** #129 (`stack-creation-redesign` branch) — the pre-canvas wizard already shipped.

---

## 1. Component Inventory

| Component | Source path | Description |
|---|---|---|
| **Stack Creation Redesign** | `project/Stack Creation Redesign.dc.html` | Primary prototype — full create-to-editor shell. Contains SIDEBAR, wizard overlay (CHOOSER → COMPOSER/TEMPLATE/DOCKER-COMPOSE), CANVAS mode, DEPLOYMENTS/LOGS/METRICS modes, ADD POPOVER, SELECT POPOVER. Imports all other components below. |
| **BlockPicker** | `project/BlockPicker.dc.html` | Scrollable 2-column grid of block cards grouped into categories (e.g. "SERVICES", "DATA STORES"). Each card shows icon + name + mono summary. Amber `+` icon to add; amber count badge if already added. Receives `cats` prop (array of `{ label, note, blocks }`). |
| **ResourceConfig** | `project/ResourceConfig.dc.html` | Slide-in right drawer for a selected canvas node. Three tab bar items rendered as mono-uppercase labels with amber underline active state: **CONFIGURATION** (name, build-from toggle image/git, port, depends-on, volumes), **DEPLOYMENT** (build cmd, start cmd), **ENVIRONMENT** (env var list via EnvRow, addon groups via MiniSelect). Amber dot indicator on dirty tabs. External `openTab` prop lets the parent jump to a specific tab. |
| **Connection Systems** | `project/Connection Systems.dc.html` | Design exploration (not a final implementation component). Presents **three** wiring UX options — 1a (service accordion: collapsed per-service, open to wire), 1b, 1c — so the user can pick one. Each option is a fully working interactive mock on the same stack. |
| **Volume Options** | `project/Volume Options.dc.html` | Design exploration (not a final implementation component). Presents **three** volume attachment patterns for canvas nodes — 1a (attached slab below node), 1b, 1c (compact chip + inline fill bar inside node). Option 1c is already implemented in the canvas (`volChips`). Full management (size, mount, snapshots) lives in the drawer's Volumes section. |
| **EnvRow** | `project/EnvRow.dc.html` | Single env-variable row: KEY input (Geist Mono, 130 px) + value input, left accent-color bar (2 px, indicates source type), dirty/addon-binding states via `row.accent`, `row.rowBg`. Used inside ResourceConfig's ENVIRONMENT tab and in addon binding groups. |
| **MiniSelect** | `project/MiniSelect.dc.html` | Compact inline dropdown trigger used inside addon binding groups in the ENVIRONMENT tab (addon type selector + DB selector). Receives trigger state via prop `t`. |
| **Glyph** | `project/Glyph.dc.html` | Thin wrapper over the Lucide icon set. Used everywhere in the prototype via `<dc-import name="Glyph" icon="..." size="N">`. No novel icons — all substituted from Lucide. |
| **DeploymentsView** | `project/DeploymentsView.dc.html` | Secondary view mounted under `isModeDeploy`. Deployment history list for the selected stack. Not part of the create/canvas flow. |
| **MetricsView** | `project/MetricsView.dc.html` | Secondary view mounted under `isModeMetrics`. Charts/metrics for the selected stack. Not part of the create/canvas flow. |
| **LogsView** | `project/LogsView.dc.html` | Secondary view mounted under `isModeLogs`. Streaming logs for the selected stack. Not part of the create/canvas flow. |

---

## 2. Token Mapping

The bundle's design system (`colors_and_type.css`) uses a flat, dark-first custom-property vocabulary. The live codebase (`frontend/src/index.css`) uses Shadcn/Tailwind conventions with light-first `:root` and `.dark` overrides. The underlying color values align well; the naming convention diverges significantly.

**Mapping uses dark-mode values** since the canvas/editor surface is always dark.

| Bundle token | Bundle value (dark) | Live token name | Status |
|---|---|---|---|
| `--amber` | `#f97316` | `--brand` | **MATCH** (same hue, different name) |
| `--amber-hover` | `#ea6a0e` | `--brand-hover` | **MATCH** |
| `--amber-press` | `#c25809` | `--brand-press` | **MATCH** |
| `--amber-soft` | `rgba(249,115,22,0.12)` | `--brand-bg` (0.10 alpha) | **DRIFT** — 12 % vs 10 % opacity |
| `--border-amber` | `#f97316` | `--brand-border` (0.30 alpha) | **DRIFT** — bundle is solid; live is translucent |
| `--bg` | `#0a0e14` | `.dark --background` ≈ `oklch(0.162 0.014 258)` | **MATCH** (dark) |
| `--bg-card` | `#11161e` | `.dark --card` ≈ `oklch(0.199 0.018 260)` | **MATCH** (dark) |
| `--bg-elev` | `#161c26` | `.dark --popover` ≈ `oklch(0.225 0.022 260)` | **MATCH** (dark) |
| `--bg-inverse` | `#f5f0e6` | — | **NEW** — no live equivalent |
| `--border` | `#1f2937` | `.dark --border` ≈ `oklch(0.278 0.030 257)` | **MATCH** (dark) |
| `--border-soft` | `#161c26` | — | **NEW** — no live equivalent (close to `--bg-elev`) |
| `--fg1` | `#ffffff` | `.dark --foreground` ≈ `oklch(0.98 0 0)` | **MATCH** (different name) |
| `--fg2` | `#cbd5e1` | — | **DRIFT** — live has no `--fg2`; closest is removed |
| `--fg3` | `#94a3b8` | — | **NEW** — live has no `--fg3` |
| `--fg-muted` | `#64748b` | `--fg-muted` ≈ `oklch(0.668 0 0)` (#9c9c9c light) | **DRIFT** — same name, different value (dark #64748b vs light #9c9c9c) |
| `--font-sans` | `'Geist', system-ui` | `--font-sans: Geist, sans-serif` | **MATCH** |
| `--font-mono` | `'Geist Mono', ...` | `--font-mono: Geist Mono, monospace` | **MATCH** |
| `--ok` / `--ok-soft` | `#22c55e` / rgba(34,197,94,0.14) | `--success` / `--success-bg` | **MATCH** (renamed) |
| `--warn` / `--warn-soft` | `#eab308` / rgba(234,179,8,0.16) | `--warn` / `--warn-bg` | **MATCH** |
| `--err` / `--err-soft` | `#d9223e` / rgba(217,34,62,0.16) | `--destructive` / `--danger-bg` | **MATCH** (different names) |
| `--info` / `--info-soft` | `#3b82f6` / rgba(59,130,246,0.16) | `--info` / `--info-bg` | **MATCH** |
| `--r-0` … `--r-pill` | 0px/2px/4px/6px/8px/999px | — | **NEW** — live only has `--radius: 0.5rem` |
| `--s-1` … `--s-10` | 4px/8px/12px/16px/24px/32px/48px/64px/96px/128px | — | **NEW** — live has no spacing-scale tokens |
| `--shadow-amber` | `0 0 0 3px rgba(249,115,22,0.20)` | — | **NEW** — live has no amber focus ring token |
| `--shadow-0/1/2` | none/border/box-shadow | `--shadow-2xs` … `--shadow-2xl` | **DRIFT** — different scale names |
| `--hairline` | `0.5px solid var(--border)` | — | **NEW** — shorthand not in live |
| `--dur-1/2/3`, `--ease`, `--transition` | animation durations/easing | — | **NEW** — live has no animation tokens |
| `--t-display/h1/h2/h3/body/body-sm/caption/mono` | type scale 80px → 13px | — | **NEW** — live has no type-scale tokens |
| `--tr-loose/normal/tight/mono` | letter-spacing values | — | **NEW** — no tracking tokens in live |
| `--slab-1/2/3` | #fff / #cbd5e1 / #94a3b8 | — | **NEW** — brand slab system not in live |
| `--max-w`, `--pad-x` | layout constraints | — | **NEW** — not in live |

**Notable drifts requiring decision before implementation:**
1. **Naming convention** — every `--amber` usage in the bundle must map to `--brand` in live code (or `--amber` aliases must be added to `index.css`).
2. **fg hierarchy** — `--fg2` / `--fg3` are absent from live CSS; the drawer and canvas rely on them heavily. Either add them or map to Tailwind muted classes.
3. **Radius scale** — `--r-2` (4px) … `--r-4` (8px) are used throughout components. The live `--radius: 0.5rem` (8px) only covers one step.
4. **`--amber-soft` opacity** — bundle at 12%, live `--brand-bg` at 10%. Minor but visible on amber halos/backgrounds.
5. **Dark-mode-first assumption** — the canvas surface is always dark; the live app can be light. Confirm whether canvas always forces `.dark` class or adapts.

---

## 3. Intent Summary

### What this design is

This bundle is the **deferred canvas editor** — confirmed from source. It is broader than just the canvas node graph: it redesigns the entire create-and-edit flow from the first "Create Stack" click through to the ongoing editor. Specifically it contains:

**Phase 1 — Wizard overlay** (phases that PR #129 already partially shipped):
- **CHOOSER**: "How do you want to start?" — three starting points: block composer, template browser, docker-compose import.
- **TEMPLATE CHOOSER**: Searchable grid of curated templates (imports `BlockPicker` categories for template cards).
- **DOCKER COMPOSE IMPORT**: Paste/upload compose file → parse into canvas nodes.
- **COMPOSER**: "What's in your stack?" — left column: searchable block picker with categories (uses `BlockPicker`); right column: connections summary panel with "Drag, wire, and configure on the canvas" CTA.

**Phase 2 — Canvas editor** (the deferred deliverable):
- A **dotted-grid canvas** (`radial-gradient` dot pattern, `var(--bg)` fill) with `position:absolute` draggable nodes.
- **Canvas nodes** are 216 px-wide cards: status dot (ok/warn), Glyph icon, name, kindLabel badge, summary line (mono), optional volume chips (compact fill bar, option 1c from `Volume Options.dc.html`).
- **SVG edges** (`edgesSvg`) rendered as a layer behind nodes to show connections.
- **"+ Add resource" button** (top-left of canvas) opens the ADD POPOVER → BlockPicker filtered list.
- **Right-side drawer** opens when a node is clicked → `ResourceConfig` with CONFIGURATION / DEPLOYMENT / ENVIRONMENT tabs. Drawer has save/discard footer with dirty-count badge.
- **Mode tab bar** (top of editor shell) switches between: `canvas` / `deploy` / `logs` / `metrics`. The canvas tab is the main configuration surface; the others show live operational views (DeploymentsView, LogsView, MetricsView).
- **"Show variable connections" toggle** on canvas controls SVG edge visibility.

**Design explorations embedded in the bundle (not yet a final implementation decision):**
- `Connection Systems.dc.html` — three options (1a/1b/1c) for how the Composer "Connections" panel works. User must pick one before the canvas connection-wiring can be implemented.
- `Volume Options.dc.html` — three attachment patterns (1a/1b/1c). The canvas already uses option 1c (compact chip inside node); this file is an audit/decision artifact.

### Relationship to the just-built wizard (PR #129)

PR #129 built a **modal chooser → composer/templates/docker-compose flow** that hands off to the existing `/stacks/create` form route. This design bundle **replaces that handoff target** — instead of routing to the existing form, the COMPOSER step flows into the new CANVAS editor. The wizard phases in this bundle are also a refinement of what PR #129 shipped (same screens, tighter design, animation tokens). Implementation will need to decide whether to:
(a) re-do the PR #129 wizard phases with these refined designs, or
(b) keep PR #129's wizard as-is and only build the new canvas editor as the handoff target.

---

## 4. Acceptance Criteria

Extracted from `_adherence.oxlintrc.json` (treat as the authoritative token-lint ruleset):

**Hard rules (lint warnings on violation):**
1. **No raw hex colors** — any `Literal` matching `/#[0-9a-fA-F]{3,8}/` must be replaced with `var(--<token>)`.
2. **No raw px values** — any `Literal` matching `/\b\d+px\b/` must be replaced with a spacing token via `var()`.
3. **Font families** — only `Geist` (sans) and `Geist Mono` (mono) are permitted. Any other `font-family` literal is a violation.
4. **No direct component-internal imports** — import from the design system's `index.js` entry, not from `ui_kits/dashboard/**` or `ui_kits/marketing-site/**` internals.

**Canonical token allowlist** (all `var()` references in components must draw from this set — any token not in this list is a violation):
```
--amber, --amber-hover, --amber-press, --amber-soft
--bg, --bg-card, --bg-elev, --bg-inverse
--border, --border-amber, --border-soft
--dur-1, --dur-2, --dur-3, --ease, --transition
--err, --err-soft, --fg-muted, --fg1, --fg2, --fg3
--font-mono, --font-sans, --hairline
--info, --info-soft, --max-w, --ok, --ok-soft, --pad-x
--r-0, --r-1, --r-2, --r-3, --r-4, --r-pill
--s-1 … --s-10
--shadow-0, --shadow-1, --shadow-2, --shadow-amber
--slab-1, --slab-2, --slab-3
--t-body, --t-body-sm, --t-caption, --t-display, --t-h1, --t-h2, --t-h3, --t-mono
--tr-loose, --tr-mono, --tr-normal, --tr-tight
--warn, --warn-soft
```

**Implied visual constraints from the design system README:**
- Single brand color only (`--amber`); no secondary palette.
- Flat is the default — shadows used very rarely (`--shadow-2` at most).
- Radii kept tight (4–8 px max for cards); no large rounded cards.
- Icons: Lucide only (via `Glyph` component); no other icon libraries.
- Cartographic / infrastructure aesthetic — no consumer-app softness.

---

## 5. Source-File Map

| Bundle path | What it defines |
|---|---|
| `README.md` | Handoff preamble: what coding agents should do, notes on the DC component format (`x-dc` / `dc-import` / `sc-if` / `sc-for`), bundle structure. |
| `project/Stack Creation Redesign.dc.html` | Full create-to-editor shell: all screens/phases (CHOOSER, TEMPLATE CHOOSER, DOCKER COMPOSE IMPORT, COMPOSER, CANVAS, DEPLOYMENTS, LOGS, METRICS), mode state machine, drawer, popovers. Primary implementation target. |
| `project/BlockPicker.dc.html` | Reusable block-picker grid component (categories → block cards). Used in COMPOSER and ADD POPOVER. |
| `project/ResourceConfig.dc.html` | Right drawer for a selected canvas node. Three tabs: CONFIGURATION / DEPLOYMENT / ENVIRONMENT. Dirty-state tracking per tab. |
| `project/Connection Systems.dc.html` | Design exploration: 3 wiring UX options for the Connections panel. **Decision artifact — not directly implementable until option is chosen.** |
| `project/Volume Options.dc.html` | Design exploration: 3 volume attachment patterns for canvas nodes. **Decision artifact — option 1c is already shown in canvas.** |
| `project/EnvRow.dc.html` | Env-variable KEY=VALUE row with source accent. Used inside ResourceConfig ENVIRONMENT tab. |
| `project/MiniSelect.dc.html` | Compact inline dropdown for addon/db binding. Used inside ResourceConfig ENVIRONMENT tab's addon groups. |
| `project/Glyph.dc.html` | Lucide icon wrapper. Used everywhere via `<dc-import name="Glyph" icon="...">`. |
| `project/DeploymentsView.dc.html` | Deployments mode view (secondary — operational, not create/edit flow). |
| `project/MetricsView.dc.html` | Metrics mode view (secondary). |
| `project/LogsView.dc.html` | Logs mode view (secondary). |
| `project/support.js` | DC runtime bootstrap (component system). Not part of the React implementation. |
| `project/_ds/.../colors_and_type.css` | Design system token source-of-truth: colors, typography scale, spacing, radii, shadows, animation. |
| `project/_ds/.../styles.css` | Thin entry sheet — imports `colors_and_type.css`. |
| `project/_ds/.../_adherence.oxlintrc.json` | Token-adherence lint rules and canonical token allowlist. **Treat as acceptance criteria source.** |
| `project/_ds/.../README.md` | Brand, voice, visual, iconography guidance for the Stackdome design system. |
| `project/screenshots/` | Reference PNGs: `editor-canvas.png`, `composer.png`, `env-tab.png`, `volopts.png`, `volnode.png`, `popover.png`, `01/02/03-mode(s).png`, `01/02-v2.png`, `cs-*.png` (connection system options), `dash.png`. Do not open — filenames used as flow hints only. |

---

## Appendix: Canvas Node Data Shape

From the main file's state section — the workload kinds present in the design prototype:

| `kind` | `kindLabel` | `group` | Notes |
|---|---|---|---|
| `service` | `Web` | — | Standard web service |
| `worker` | `Worker` | `job` | **Backend prerequisite**: WorkloadType CRD is ready in cluster-agent v0.6.2 but unwired in api-server model/builder/form |
| `postgres` | `Postgres` | — | Managed addon |
| `redis` | `Redis` | — | Managed addon |
| `minio` | `Object` | — | Object storage addon (MinIO) |

No `CronJob` or `schedule` kind is present in the design data.

## Appendix: Backend Prerequisites

1. **WorkloadType / Worker** — `kind:'worker'` appears as a canvas node. The CRD is ready (cluster-agent v0.6.2) but api-server model, builder, and form are unwired. Canvas and ResourceConfig DEPLOYMENT tab must be extended to support worker-type resources before the Worker node can be fully configured. This was flagged as a known gap.
2. **Connection auto-wiring** — The design shows SVG edges between nodes, but the specific connection UX (which of the 3 options in `Connection Systems.dc.html`) is unresolved. No backend connection-injection API is shown in the bundle; whether `spec.connections` is set by the canvas save or auto-wired from addon discovery is an open question.
3. **Canvas node persistence** — Node positions (x/y) and the canvas state will need to be persisted somewhere (stack spec? separate endpoint?). No backend API for this appears in the bundle.
