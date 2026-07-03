# Design Reference: Stack Creation Redesign — Floating Drawers, Collapsible Header, Labels/Rename, Public Endpoints

**Date:** 2026-07-03
**Source:** Claude Design project `faed4868-7719-4ab1-9b21-ff8eae2933ff` ("Stackdome Stack Creation Redesign"), fetched fresh via DesignSync on 2026-07-03.
**Primary file:** `Stack Creation Redesign.dc.html` (116,206 bytes). **Companion:** `Panel Explorations.dc.html` (79,497 bytes).
**Raw copies:** session scratchpad `stack-creation-redesign.dc.html` / `panel-explorations.dc.html` (plus extracted `scr-main.jsx` = the interactive `text/x-dc` script, and `markup-main.html` = the static template).

> Supersedes-in-part: `2026-07-01-stack-creation-redesign-canvas-design-reference.md` (canvas/create-flow reference). This document covers the **new** features added to the design after July 1: floating stacked drawers (1a), collapsible header, inline stack rename + labels (1e), and the PUBLIC endpoint row. The July 1 doc remains valid for the create-wizard/composer/canvas fundamentals; where they overlap, THIS document is authoritative.

The design medium is an HTML/CSS/JS prototype. The `text/x-dc` script (a `DCLogic` class with `state` + `renderVals()`) is the authoritative interaction source; the markup template carries all geometry as inline styles. Preview viewport: **1440×900**.

---

## 1. Component Inventory

All locations are in `Stack Creation Redesign.dc.html` unless noted. Byte offsets are into the file; "script" = the `text/x-dc` block (lines are within that script).

| Component / section | Location |
|---|---|
| App sidebar (DEFAULT PLATFORM nav: Stacks, Secrets, Object Stores, Addons, Clusters, Domains, Admin) | markup `<!-- SIDEBAR -->`, ~bytes 400–6,300 |
| Expanded stack header (chevron, h1 name/rename, READY pill, label chips, dots menu, subtitle, PUBLIC row) | markup ~6,300–14,400 (`sc-if {{ showFullHeader }}`) |
| Tabs row (Configuration/Deployments/Logs/Metrics + dirty summary + Deploy) | markup ~13,100–14,400 |
| Collapsed header (44px slim bar) | markup `<!-- collapsed header -->` at 14,451 (`sc-if {{ showCompactHeader }}`) |
| Create-stack wizard overlay + chooser/template/compose phases | markup 17,200–41,000 (covered by July 1 doc) |
| Canvas (dotted grid, nodes, edges, Add resource, zoom stack, connections toggle, hint bar) | markup `<!-- CANVAS (configuration) -->` at 41,061 |
| Behind (peeking) panels | markup `<!-- stacked panels peeking behind the front one -->` at 48,105 |
| Front drawer panel (crumbs, title row, DEPENDS ON, ResourceConfig body, footer) | markup `<!-- front panel -->` at 49,104 |
| Deployments / Logs / Metrics mode views | markup 54,600–58,000 (`dc-import` DeploymentsView / LogsView / MetricsView) |
| Add-resource popover (search + BlockPicker) | markup `<!-- ADD POPOVER -->` ~58,000 |
| Select popover (root-level, never clips in drawer) | markup `<!-- SELECT POPOVER -->` end of file |
| Panel-stack logic (`openPanel`/`pushPanel`/`popPanel`/`closeAllPanels`/`gotoPanel`), header toggle | script lines 540–547 |
| Rename + labels logic (`startEditName`/`commitName`/`commitLabel`/…, `copyUrl`) | script lines 549–561 |
| Keyboard handling (⌘. toggle, Esc pop, ⇧Esc close-all) | script `componentDidMount` lines 563–574 |
| Derived view data (`behindPanels`, `crumbs`, `publicUrls`, `dirtySummary`, `editorModes`, `stackSubtitle`, `escHint`) | script `renderVals()` lines 576–713 |
| Animations (`sd-fade/scrim/pop/drawer/mode`) | `<style>` block in body head |
| Drawer explorations 1a–1d, name/label explorations 1e–1g | `Panel Explorations.dc.html` (offsets: 1a@2,130 · 1b@19,602 · 1c@37,040 · 1d@53,982 · 1e@68,004 · 1f@70,967 · 1g@75,092) |

Shared prototype components imported via `dc-import`: `Glyph` (icon set), `ResourceConfig`, `DeploymentsView`, `LogsView`, `MetricsView`, `BlockPicker`, `MiniSelect` — each has its own `.dc.html` in the project.

---

## 2. Token Mapping

Bundle tokens come from `_ds/stackdome-design-system-…/colors_and_type.css` (dark-first, flat vocabulary). Live source: `frontend/src/index.css` (oklch, light-first `:root` + `.dark`). Verified against live on 2026-07-03. The canvas/editor surface is dark-first; mapping uses dark values.

| Bundle token | Bundle value | Live token | Status |
|---|---|---|---|
| `--amber` | `#f97316` | `--brand` `oklch(0.72 0.20 40)` | MATCH (renamed) |
| `--amber-hover` | `#ea6a0e` | `--brand-hover` | MATCH |
| `--amber-press` | `#c25809` | `--brand-press` | MATCH |
| `--amber-soft` | `rgba(249,115,22,0.12)` | `--brand-bg` (α 0.10) / `--brand-bg-hover` (α 0.16) | DRIFT — 12% vs 10%/16% |
| `--border-amber` | `#f97316` solid | `--brand-border` (α 0.30) | DRIFT — solid vs translucent |
| `--bg` | `#0a0e14` | `.dark --background` | MATCH (dark) |
| `--bg-card` | `#11161e` | `.dark --card` | MATCH (dark) |
| `--bg-elev` | `#161c26` | `.dark --popover` | MATCH (dark) |
| `--bg-inverse` | `#f5f0e6` | — | NEW |
| `--border` | `#1f2937` | `.dark --border` | MATCH (dark) |
| `--border-soft` | `#161c26` | — | NEW (≈ `--bg-elev`) |
| `--fg1` | `#ffffff` | `.dark --foreground` | MATCH (renamed) |
| `--fg2` | `#cbd5e1` | — | NEW — heavily used (chips, hosts, tab-inactive) |
| `--fg3` | `#94a3b8` | — | NEW — heavily used (icons, subtitles) |
| `--fg-muted` | `#64748b` | `--fg-muted` `oklch(0.668 0 0)` | DRIFT — same name, different value |
| `--ok` / `--ok-soft` | `#22c55e` / α .14 | `--success` / `--success-bg` (α .10) | MATCH (renamed; soft α drift) |
| `--warn` / `--warn-soft` | `#eab308` / α .16 | `--warn` / `--warn-bg` (α .12) | DRIFT — live warn is `oklch(0.50 0.13 75)`, darker |
| `--err` / `--err-soft` | `#d9223e` / α .16 | `--destructive` / `--danger-bg` (α .10) | MATCH (renamed; α drift) |
| `--info` / `--info-soft` | `#3b82f6` / α .16 | `--info` / `--info-bg` (α .10) | DRIFT — live info `oklch(0.48 0.20 264)` darker |
| `--font-sans` / `--font-mono` | Geist / Geist Mono | `--font-sans` / `--font-mono` | MATCH |
| `--r-0…--r-4`, `--r-pill` | 0/2/4/6/8/999px | `--radius: 0.5rem` only | NEW — features use `--r-2`(4px), `--r-3`(6px), `--r-4`(8px), `--r-pill` |
| `--s-1…--s-10` | 4→128px | — | NEW (spacing scale; features mostly use literal px) |
| `--shadow-2` | `0 8px 24px rgba(0,0,0,0.40)` | `--shadow-lg`-ish | DRIFT — live shadow scale differs |
| `--shadow-amber` | `0 0 0 3px rgba(249,115,22,0.20)` | — | NEW (focus ring; ≈ `--ring` + offset) |
| `--hairline` | `0.5px solid var(--border)` | — | NEW (0.5px borders used everywhere) |
| `--ease` | `cubic-bezier(0.2,0.6,0.2,1)` | — | NEW |
| `--dur-1/2/3` | 120/240/400ms | — | NEW |
| `--t-*` type scale, `--tr-*` tracking, `--slab-*`, `--max-w`, `--pad-x` | various | — | NEW (marketing scale; editor uses literal px) |

**Raw values in the new features' markup/script (must not survive unmapped):**

| Raw value | Where | Suggested mapping |
|---|---|---|
| `#1a0e05` (×8) | Deploy button text on amber | `--primary-foreground` `oklch(0.137 0 0)` (accept slight drift) or new `--brand-foreground` |
| `rgba(10,14,20,0.72)` | wizard scrim | `--background` @ 72% (overlay token needed) |
| `rgba(10,14,20,0.55)` (×2) | behind-panel dim overlay; add-popover scrim | `--background` @ 55% (same overlay token family) |
| `rgba(249,115,22,0.55)` | dirty (unselected) canvas-node border | `--brand` @ 55% |
| `rgba(249,115,22,0.35)` / `0.32` / `0.06` | canvas edge/halo accents, dirty env-row bg | `--brand` alphas |
| `rgba(34,197,94,0.16)` | deploy dot halo | `--success-bg` |
| `rgba(217,34,62,0.16)` | error soft | `--danger-bg` |
| `#fff` | toggle knob | `--foreground` (dark) |
| `#08130b`, `#d9c36a` | check-icon on ok fill; warn log text | `--success-bg` contrast / `--warn` tint — flag for brand review |

**Decisions to make before implementing:** (1) add or alias `--fg2`/`--fg3` (the entire drawer/header hierarchy depends on them); (2) an overlay/scrim token; (3) a `--brand-foreground` for text-on-amber; (4) radius steps 4px/6px/8px vs single `--radius`.

---

## 3. Intent Summary (per feature)

### a. Floating stacked drawers (exploration 1a — the chosen direction)

From `Panel Explorations.dc.html` 1a note: *"Smallest change, biggest win. The panel becomes a floating surface inset 12px from the viewport edge — it no longer sits below the title + tab rows, so on 1080p it gains ~170px of height. Width 600px (was 496). Opening a sub-view pushes a new panel onto a stack; the previous one stays peeked behind with a breadcrumb at top. Esc pops one level."* (Railway-style). Rejected alternatives: 1b docked resizable split, 1c bottom dock, 1d focus mode. The main file implements 1a.

**Front panel geometry** (main file, authoritative):
- `position:fixed; top:12px; right:12px; bottom:12px; width:600px; z-index:200`
- `background: var(--bg)`; `border: 0.5px solid var(--border)`; `border-radius: var(--r-4)` (8px); `box-shadow: var(--shadow-2)` (0 8px 24px rgba(0,0,0,.40))
- Floats **over** header + tabs (full viewport height). **No page backdrop/scrim** — the canvas stays visible and interactive to the left.
- Open animation: `sd-drawer 260ms cubic-bezier(0.2,0.6,0.2,1)` = from `translateX(34px) + opacity:0` to rest. No explicit close animation (unmount).

**Behind (peeked) panels**, one per stacked entry except the front, depth `d` = distance from front (front-1 ⇒ d=1):
- `position:fixed; top:(12+10d)px; bottom:(12+10d)px; right:(12+16d)px; width:600px; z-index:(200−d)` — i.e. each deeper panel is inset 10px vertically and peeks 16px to the left.
- `background: var(--bg-elev)`, hairline border, `--r-4` radius, **full-card dim overlay** `rgba(10,14,20,0.55)` on top.
- Shows only its header row (padding 15px 16px): 17px Glyph icon (fg3) + name 14px/500 fg2. Whole card is clickable, `title="Back to {name}"`.
- Stack depth is unbounded in the model; visually each level costs 16px of peek.

**Front panel internal anatomy** (top→bottom):
1. Crumb row (only when depth > 1): padding 10px 14px 0; crumbs 10.5px mono, separator `›`, last = fg2, ancestors = fg-muted, hover amber, click = jump; right-aligned esc hint 10px mono fg-muted.
2. Title row: padding 11px 14px 13px, hairline bottom. 9px status dot (ok/warn/err), 19px type icon (amber), name 16px/500 fg1 + summary 11px mono fg-muted (ellipsis). If dirty: amber chip (border 0.5px amber, bg amber-soft, radius 4px, padding 2px 4px 2px 9px, 11px/500) with change-count label + 16×16 discard "x" (hover fills amber). Else: kind label t-marker fg-muted. Close "x" 18px (fg3, hover fg1 + bg-elev).
3. DEPENDS ON row (when deps exist): bg-card, padding 10px 14px, hairline bottom; label t-marker 9px; dep chips (pill, 0.5px border, padding 4px 10px, 12px fg2, icon 13px + arrowRight 12px, hover amber) — click **pushes** that resource's panel.
4. Body: `flex:1; overflow:auto` → `ResourceConfig` component.
5. Footer: hairline top, padding 11px 16px, space-between: "View logs" (12.5px fg3, terminal icon 14, hover amber + amber-soft bg) / "Remove resource" (fg-muted, trash 14, hover err + err-soft bg).

### b. Collapsible top section / header

Two mutually exclusive states, editor phase only (`showFullHeader = editor && !headerCollapsed`, `showCompactHeader = editor && headerCollapsed`).

**Expanded**: container padding 24px 28px 0.
- Row 1 (flex, gap 12, wrap): collapse chevron (Glyph `chevron` 18px rotated 180°, padding 6px, radius 4px, fg3, hover fg1+bg-elev, tooltip "Focus — full-screen canvas (⌘.)"), stack name h1 29px/500/−0.02em, READY status pill (6px ok dot, 0.5px ok border, ok-soft bg, pill radius, padding 3px 11px, 9.5px t-marker ok text), 0.5×22px vertical divider, label chips + "+ label" ghost, spacer, dots menu (18px, fg3).
- Row 2 (indent margin-left 38px): subtitle margin-top 7px, 13px fg3 — `"{n} services · {m} volumes"`.
- Row 3 (same indent, when public endpoints exist): PUBLIC endpoint row, margin-top 13px (see d).
- Tabs row: padding 18px 28px, hairline bottom. Tabs = Configuration (`grid`), Deployments (`rocket`), Logs (`terminal`), Metrics (`activity`): padding 8px 15px, 1px border (active: amber; inactive: transparent), radius 4px, 14px/500, active color amber + bg amber-soft, inactive fg3 hover fg1; transition color/border-color 150ms `--ease`. Configuration tab gets a badge = **count of dirty resources** (mono 9.5px, amber on amber-soft, pill, padding 1px 6px). Right side: unsaved summary (mono 11.5px amber, `"1 unsaved change"` / `"N unsaved changes"` — total field+env changes across resources) then **Deploy** button (amber bg, text `#1a0e05`, padding 8px 18px, radius 4px, 13px/600, letter-spacing .02em, rocket 15px, hover amber-hover; click → Deployments tab).

**Collapsed** ("focus mode" bar): single 44px-high bar, padding 0 16px, gap 10, hairline bottom, bg `--bg`. Contents: chevron 17px un-rotated (tooltip "Exit focus — expand header"), stack name 14px/500 (not editable here), 6px ok dot (tooltip "Ready"), 0.5×18px divider, all four tabs compact (padding 5px 10px, 12px/500, icons 13px, badge 9px/pad 1px 5px), spacer, unsaved summary 11px, Deploy compact (padding 6px 14px, 12px/600, rocket 13px). Everything else (labels, subtitle, PUBLIC row) is hidden — canvas gets the reclaimed height.

**Toggle**: chevron click, or **⌘. / Ctrl+.** globally (editor phase only). Plain state flip — no transition animation is defined on the swap.

### c. Stack labels + rename (exploration 1e — chosen; 1g popover and 1f stack-as-resource rejected for now)

1e note: *"Name becomes an inline-editable field (click or hover shows the pencil; enter commits, esc reverts — same dirty/reset semantics as resource fields). Labels are removable chips with a ghost + label chip that opens a tiny type-ahead. Zero new surfaces; always visible on every tab."* (1f — click empty canvas → stack panel — is flagged in the explorations "Try next" as pairing well with 1a; not in the main prototype.)

**Rename**: display = h1 29px + pencil Glyph 15px at 50% opacity, wrapper `cursor:text`, hover reveals `border-bottom:1px dashed var(--fg-muted)` (transparent when idle), tooltip "Rename stack". Click → replaced by `<input>`: same 29px/500/−0.02em type, bg-card, **1px solid amber** border, radius 4px, padding 1px 10px, `--shadow-amber` focus ring (0 0 0 3px rgba(249,115,22,.20)), min-width 220px, autofocus + select-all. Enter = commit, Escape = cancel, blur = commit. Empty/whitespace input keeps the previous name (no error UI).

**Labels**: chips are mono 11.5px fg2, 0.5px border, pill radius, padding 4px 11px, gap 7, hover border-fg3; each has a 12px "x" (fg-muted → err on hover), tooltip "Remove label". Ghost add-chip: `+ label` — 12px plus icon, fg-muted, **0.5px dashed** border, pill, same padding, hover amber border+text. Click → inline input: width 104px, mono 11.5px, bg-card, 0.5px amber border, pill, padding 4px 11px, placeholder `label…`, autofocus. Enter commits **and immediately reopens the input** if text was entered (chained adds); Escape cancels; blur commits. Normalization on commit: `trim → lowercase → whitespace→dash`; duplicates silently dropped; empty = no-op. **No max-label cap and no validation hints in the design.** Chips wrap (`flex-wrap`) beside the name.

### d. PUBLIC endpoint row

Rendered under the subtitle when the stack has ≥1 HTTP service (`CATALOG[type].kind === 'service' && portProto includes 'HTTP'`). Layout: flex row, gap 9, wrap, margin-top 13px.

- Prefix label: `PUBLIC` — t-marker, 9.5px, letter-spacing 1.5px, fg-muted.
- One **segmented pill per HTTP service** (container: 0.5px border, radius 4px, bg-card, overflow hidden, hover border-fg3), three segments:
  1. **Service chip**: bg-elev, padding 5px 10px, hairline right border; 12px type Glyph (fg-muted) + service name mono 11.5px fg3; tooltip `"Mapped to {name} · :{port}"` — this is the service→URL mapping affordance.
  2. **URL link** (`<a target="_blank">`): padding 5px 11px, gap 8; **5px ok-green status dot**, host mono 12px fg2, `externalLink` Glyph 12px fg-muted; hover turns text amber; tooltip `"Open {host}"`.
  3. **Copy button**: hairline left border, padding 0 9px; `copy` Glyph 12px fg-muted; hover amber + bg-elev; on click copies the full `https://` URL via `navigator.clipboard`, swaps icon to `check` in ok-green for **1400ms**, then reverts. Click is `preventDefault`/`stopPropagation` (doesn't follow the link).
- **URL scheme**: `https://{serviceName}.{slug}.stackdome.app` where `slug = stackName.trim().toLowerCase().replace(/[^a-z0-9]+/g,'-')` (trimmed of leading/trailing dashes, fallback `stack`). Renaming the stack live-updates every endpoint host.
- Multiple endpoints simply wrap as additional segmented pills after the single PUBLIC prefix. The status dot is static green in the prototype (no per-endpoint health logic).

---

## 4. Interaction Spec

**Drawer stack state machine** (`panelStack: string[]`, last = front):
- Canvas node click → `openPanel(iid)`: **replaces** the whole stack with `[iid]` (no stacking from canvas).
- DEPENDS ON chip → `pushPanel(iid)`: no-op if already front; otherwise removes `iid` from anywhere in the stack and appends it (dedupe — a resource appears at most once; re-pushing reorders it to front).
- Behind-panel click / crumb click → `gotoPanel(idx)`: truncates stack to `0..idx` (everything above pops at once).
- Front-panel "x" → `popPanel()` (pop one). "Remove resource" also pops any panels of the removed instance.
- **Esc** pops one; **Shift+Esc** closes all — only when: editor phase, canvas mode, stack non-empty, no select-popover open, and not editing name/label (inputs own their Escape). A popover consumes Esc first.
- Esc hint text in the drawer: `esc pops · ⇧esc closes all` when depth > 1, else `esc closes`.
- Selected (front) node on canvas gets amber border + `--shadow-amber`; dirty nodes get a 3px inset amber left bar.

**Collapse toggle**: `headerCollapsed` boolean; chevron button in both states, plus ⌘./Ctrl+. shortcut (editor phase only). Expanded chevron is rotated 180° ("point up/collapse"), collapsed is unrotated ("expand"). Tab selection, dirty count, and Deploy remain available in both states.

**Rename flow**: click name → input (autofocus, select-all) → Enter/blur commit (empty keeps old), Esc cancels. While editing, global Esc handling is suppressed.

**Label flow**: click `+ label` → input (autofocus) → Enter commits + reopens for the next label; blur commits; Esc cancels. Normalize lowercase/dash; dedupe. Remove via chip "x".

**Endpoint copy/open**: copy button → clipboard write + 1400ms `check`/ok feedback (per-URL: `copiedUrl` state holds the URL string); link opens `https://{host}` in a new tab. Chip tooltip documents the service/port mapping.

**Mode/overlay animations**: drawer `sd-drawer` 260ms; mode-view swap `sd-mode` 240ms (fade + 4px rise); popovers `sd-pop` 200ms (add) / 130ms (select); wizard scrim `sd-scrim` 240ms; generic `sd-fade` 300ms. All use `--ease` = `cubic-bezier(0.2,0.6,0.2,1)`.

**Z-index ladder**: canvas edges 1 · nodes 2 (selected 5) · add-popover overlay 60 · wizard 120 · behind panels 200−d · front panel 200 · select-popover scrim 300 / menu 301.

---

## 5. Other Notable Structures (main file)

- **Overall layout**: fixed left sidebar (app nav) + `<main>` column: header (one of two states) → tab row → `flex:1` content area (`position:relative; overflow:hidden`) hosting canvas/mode views, popovers, and drawers.
- **Canvas**: dotted-grid background `radial-gradient(var(--border) 0.7px, transparent 0.7px)` on 24px grid over `--bg`. Nodes: 216px wide cards (1px border, radius 6px, bg-card, `cursor:grab`), 8px status dot, 17px icon, 14px/500 name, 9px kind t-marker, 11px mono summary; optional volume chips footer (drive icon, mono name, 4px amber usage bar, pct). Edge math uses node W=216, H=58; edges are amber dashed beziers (strokeWidth 1.4, opacity .55, dash `5 4`) with 3px amber dot at target.
- **Add resource**: top-left floating button (16px inset; bg-elev, hairline border, padding 9px 14px, radius 4px, 13px/500, plus 16px, `--shadow-2`, hover amber). Opens centered popover (max-width 600px) with search + `BlockPicker` categories (SERVICES / DATA STORES / JOBS & WORKERS / ADDONS / NETWORKING) and a Done button.
- **Canvas controls** bottom-left: 40px-wide vertical zoom stack (+ / − / fit, 38px rows) and a 40×40 connections toggle opening a 280px menu ("Variable references" / "Hide connections").
- **Hint bar** bottom-center: 11.5px fg-muted, "drag to rearrange · click a node to configure · edges carry connection env vars".
- **Unsaved-changes model**: per-resource `mods` (config/deploy field lists) + env diff vs baseline → `totalChanges` drives header summary; `dirtyResources` count drives the Configuration tab badge; per-resource dirty chip in drawer title row with per-resource discard.
- **Select popover** is rendered at document root (fixed, positioned at trigger coords, width = trigger width, max-height 280px, bg-elev, radius 6px, `--shadow-2`) so it never clips inside the drawer — replicate this pattern (portal) in React.

---

## 6. Source-File Map

| Path (Claude Design project) | Defines |
|---|---|
| `Stack Creation Redesign.dc.html` | **Primary.** Full editor: sidebar, header (both states), tabs, canvas, drawers, wizard, popovers, all interaction logic (`text/x-dc` script) |
| `Panel Explorations.dc.html` | Static explorations: 1a floating overlay (chosen), 1b docked split, 1c bottom dock, 1d focus mode; 1e inline rename/labels (chosen), 1f stack-as-canvas-resource, 1g title popover. "Try next" suggests 1a+1f pairing |
| `_ds/stackdome-design-system-…/colors_and_type.css` | Bundle token set (see §2) + Geist font-faces |
| `ResourceConfig.dc.html` | Drawer body: Configuration/Deployment/Environment sections |
| `Glyph.dc.html` | Icon set (chevron, pencil, plus, x, dots, copy, check, externalLink, rocket, grid, terminal, activity, workflow, move, drive, database, globe, …) |
| `BlockPicker.dc.html`, `MiniSelect.dc.html`, `EnvRow.dc.html`, `Volume Options.dc.html`, `Connection Systems.dc.html` | Sub-components (picker grid, selects, env rows, volumes, connection UI) |
| `DeploymentsView.dc.html`, `LogsView.dc.html`, `MetricsView.dc.html`, `DeployStages.dc.html`, `Deployment States.dc.html`, `DiffView.dc.html`, `OutcomeRow.dc.html` | Non-canvas tab views (deploy pipeline, logs, metrics) — out of scope here |
| `support.js` | Prototype runtime (`DCLogic`, template binding) — do **not** port |

**Live implementation anchors**: `frontend/src/index.css` (tokens), stack canvas editor components on branch `stack-canvas-editor`.
