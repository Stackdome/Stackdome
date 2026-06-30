# Stack Creation Wizard — Phase Refinement Design

**Date:** 2026-07-01
**Status:** Draft — awaiting user review
**Branch:** `stack-creation-redesign` (extends open PR #129)
**Type:** Frontend styling re-skin. No backend. No new routes.

## Goal

Refine the visual design of the four existing stack-creation **wizard phases**
(Chooser, Composer, Template, Docker-Compose) to match the Claude Design
handoff bundle "Stack Creation Redesign". Styling only — structure, navigation,
and behavior already match the bundle and stay as shipped in PR #129.

## Scope

### In scope
- Visual refinement of the 4 wizard-phase components to match the bundle's
  polished design (type scale, a few sizing/decorative elements).

### Explicitly out of scope (deferred)
- **Canvas / node-graph editor** — the bundle's headline screen. Separate spec.
- **Composer "Connections" wiring sub-panel** — needs the connection/auto-wiring
  data model the canvas effort owns. Deferred (user-confirmed 2026-07-01).
- **ResourceConfig drawer, Volume Options, Connection Systems, EnvRow** — canvas-era
  components. Out.
- **Worker / WorkloadType** — backend CRD-ready but unwired in api-server. Out.
- **Operational views** (Deployments / Logs / Metrics) — unrelated. Out.
- Any change to the `/stacks/create` form route or the wizard's phase machine /
  navigation / modal frame (already correct).

## Design reference

- **Design-ref artifact:** `docs/superpowers/design-refs/2026-07-01-stack-creation-redesign-canvas-design-reference.md`
  (full component inventory + token mapping, authored via `ingest-design-bundle`).
- **Bundle source (authoritative for exact values):**
  `scratchpad/handoff/stackdome-stack-creation-redesign/project/`
  - `Stack Creation Redesign.dc.html` — CHOOSER / COMPOSER / TEMPLATE / DOCKER-COMPOSE phase markup.
  - `BlockPicker.dc.html`, `Glyph.dc.html` — palette + icons.
  - `_ds/.../colors_and_type.css`, `styles.css` — tokens, exact spacing/type values.
- The bundle README forbids rendering/screenshotting the bundle; read the source
  HTML/CSS directly for exact values.

## Token reconciliation — CRITICAL

The bundle names its accent the `--amber` family. **Live `--primary` and `--brand`
both resolve to `oklch(0.72 0.20 40)` = amber `#f97316`** (verified in
`frontend/src/index.css:15,31`). Therefore `text-primary` / `bg-primary/10` /
`border-primary` in the live wizard **already render the bundle's amber** — they
are NOT a color defect.

Implication: any delta phrased as "uses primary/muted/default instead of amber"
must be **re-verified against the bundle source per element** before acting. The
upstream delta analysis (`scratchpad/wizard-phase-deltas.md`) already produced one
false positive on exactly this point (it called the entire Chooser primary card
"wrong-colored" when it is already amber). **Do not trust the delta doc blind** —
treat the bundle `.dc.html` + `colors_and_type.css` as the source of truth and
confirm each change against it.

Use existing semantic tokens only (`bg-card`, `border`, `text-muted-foreground`,
`bg-primary`/`text-primary`/`border-primary`, `bg-brand-bg`, `border-brand-border`,
`text-success`). No raw hex, no off-scale type. Per brand-system guidance.

## Components and per-phase deltas

Each delta below is a *candidate* confirmed at design level; the implementer
verifies the exact target value against the bundle source before editing.

### 1. Chooser — `frontend/src/pages/stacks/components/wizard/wizard-chooser.tsx`
- **Title type scale:** bump heading to the bundle's larger size (live ≈`text-2xl`/24px
  → bundle ≈30px). Map to nearest brand/Tailwind scale class (`text-3xl`); no `text-[30px]`.
- **Primary icon container:** 44px → 54px (verify exact in bundle).
- **RECOMMENDED badge:** filled → outlined (border + transparent bg + amber text).
- **Content centering:** wrap chooser content in a max-width box (~780px) centered.
- **Alt-start cards:** add trailing `arrowRight` icon to each of the 4 alt cards.
- **No change:** primary card amber coloring (already correct via `primary`).

### 2. Composer — `frontend/src/pages/stacks/components/wizard/block-composer.tsx`
- **Title type scale:** bump to bundle size (verify exact).
- **"Your stack so far" rows:** add 8px leading status dot per row.
- **No change:** subtitle copy — we intentionally dropped the auto-wire/"Connections"
  line when wiring was deferred; keep current copy.
- **Out:** Connections sub-panel (deferred).

### 3. Template — `frontend/src/pages/stacks/components/wizard/templates-browser-panel.tsx`
- **Title type scale:** bump to bundle size (verify exact).
- **Count label:** add "N TEMPLATES" marker (mono uppercase tracked, `text-muted-foreground`).
- **Right-detail icon container:** `bg-muted` → amber-soft bg + amber border
  (`bg-brand-bg` + `border-brand-border`). This is a real delta (muted ≠ amber).
- **Verify per element:** active-selected template + search-focus accent — likely
  already amber via `primary`; confirm against bundle before changing.

### 4. Docker-Compose — `frontend/src/pages/stacks/components/wizard/docker-compose-import-panel.tsx`
- **Title type scale:** bump only if it is short vs bundle; otherwise leave.
- Structure/layout/content already sound — no other changes.

### Shell / modal frame — `stack-create-wizard.tsx`
- No changes. Fixed 1000×620 frame, phase machine, Back/Continue footer all correct.

## Data flow / error handling

No data-flow or error-handling changes — styling only. The phases' existing
import/navigation behavior (handoff to `/stacks/create` with `importedData`) is
untouched.

## Testing

- **Unit:** existing wizard test suite stays green (currently 481 total across the
  frontend suite; wizard subset 12). Update only assertions that pin changed copy
  or structure (e.g. a new "N TEMPLATES" label or status-dot element). Do not weaken
  assertions to pass.
- **Visual:** Playwright pass on all 4 phases in **both** light and dark themes;
  confirm against the bundle source intent (no rendering of the bundle itself).
- **Gates:** `tsc` clean (the pre-existing `api/postgres-backups.ts` error is the
  only acceptable one), `lint` 0 errors.

## Acceptance criteria

1. Each in-scope delta matches the bundle source value (verified element-by-element).
2. No raw hex; no off-scale type classes; only existing semantic tokens used.
3. No false-positive "fixes" — color deltas that were already amber via `primary`
   are left unchanged.
4. Connections panel and all canvas-era components remain absent.
5. Full frontend suite green; tsc + lint clean per gates above.
6. Both-theme Playwright verification of all 4 phases.

## Risks

- **False deltas:** the delta analysis already contained one. Mitigation: per-element
  verification against bundle source is a hard requirement, baked into acceptance.
- **Type-scale mapping:** bundle uses raw px; brand forbids off-scale type. Mitigation:
  map each to the nearest brand/Tailwind scale class, not arbitrary px.
- **Branch concurrency:** `stack-creation-redesign` was also receiving background-agent
  polish. Ensure no concurrent writes to the same wizard files during implementation.
