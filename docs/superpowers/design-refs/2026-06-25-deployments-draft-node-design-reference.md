# Design reference — Deployments tab (draft → save → deploy)

> Inbound design-intake artifact (skill: `ingest-design-bundle`). Source: Claude
> Design project **"UI redesign alternatives"** (`b58cdd1a-3fde-4ef0-a5d0-d700cca92994`),
> handoff folder `design_handoff_deployments_draft/`, primary file
> `Deployments (Draft node).dc.html`. Fetched via the `DesignSync` MCP (the project
> is a `/p/` design project, not a `/v1/design/h/` gzip bundle, so the skill's
> tar-fetch recipe does not apply — the read-order + output discipline still do).
>
> **This design is already implemented** on `feat/stateful-deployments` (plan
> `docs/superpowers/plans/...` / `~/.claude/plans/pure-zooming-hickey.md`). This
> artifact records the canonical source for fidelity audit, not net-new work.

## Intent summary

Editing config no longer deploys. The lifecycle splits into three deliberate steps:

```
Edit config  →  Save (creates a DRAFT — staged, not deployed)  →  Deploy (creates a release)
```

Two facts that used to be conflated must stay separate:
- **Latest deploy** — the most recent deploy *attempt*; becomes the new release the instant Deploy is clicked (even while building, even if it later fails).
- **Live release** — what is *actually serving traffic*; does NOT change until a deploy converges. A failed/in-flight deploy must never hide what's live.

The old in-tab "Deployments / Deploy" header is removed — **deploy actions live only in the status bar**. The history becomes one unified `DEPLOY TIMELINE` (no Current/Earlier split), newest at top, every node a lean toggle row + a detail card directly below it. The saved-but-undeployed draft becomes a first-class **draft node** leading the rail.

No `chats/` or `acceptance.jsx` shipped in this handoff — the README is the distilled intent and the acceptance source.

## Component inventory (design → implementation)

| Design surface (README) | Implemented in |
|---|---|
| Status bar — 4 lifecycle states (editing/staged/deploying/clean) | `shared/sticky-action-bar.tsx` (tones) + page wiring `detail/index.tsx` |
| Lifecycle derivation (`phase`) | `deployments/use-deploy-lifecycle.ts` |
| Unified `DEPLOY TIMELINE` rail | `deployments/timeline/timeline-rail.tsx` |
| Universal node (lean row + card-below) | `deployments/timeline/timeline-node.tsx` |
| Rail dot (solid/ring/spinner/dashed) | `deployments/timeline/rail-node.tsx` |
| Latest-deploy body (live cluster status) | `deployments/timeline/live-release-body.tsx` |
| Earlier/historical body (stored outcome) | `deployments/timeline/release-post-mortem.tsx` |
| `RESOURCE OUTCOME` list + row | `deployments/timeline/resource-outcome-list.tsx`, `resource-row.tsx` |
| `CONFIG CHANGES · vs #N` toggle + diff | `deployments/timeline/config-changes-toggle.tsx`, `config-diff.tsx` |
| Draft node (dashed amber) | `deployments/timeline/draft-node.tsx` |
| Build→Deploy→Ready tracker | `components/branded/stage-tracker.tsx` |
| **Live anchor** (jump-to-live; *addition, not in design*) | `deployments/timeline/live-release-summary.tsx` |

## Token mapping (bundle `colors_and_type.css` ↔ live `frontend/src/index.css`)

All reused — no new colors, no forked tokens, no raw hex in components. Bundle hex is intent; live values are oklch exposed to Tailwind v4 as `--color-*`.

| Bundle token | Bundle value | Live token | Match |
|---|---|---|---|
| `--amber` | `#f97316` | `--brand` → `--color-brand` `oklch(0.72 0.20 40)` | match |
| `--amber-hover` | `#ea6a0e` | `--brand-hover` `oklch(0.67 0.20 40)` | match |
| `--amber-press` | `#c25809` | `--brand-press` `oklch(0.55 0.16 40)` | match |
| `--amber-soft` | `rgba(249,115,22,.12)` | `--brand-bg` / `--brand-border` | match |
| `--ok` | `#22c55e` | `--success` (+ `-bg`/`-border`) | match |
| `--warn` | `#eab308` | `--warn` (+ `-bg`/`-border`) | match |
| `--err` | `#d9223e` | `--danger` (+ `-bg`/`-border`) | match |
| `--info` | `#3b82f6` | `--info` (+ `-bg`/`-border`) | match |
| `--bg` | `#0a0e14` | `--background` / page bg | match |
| `--bg-card` | `#11161e` | `--card` / `bg-card` | match |
| `--bg-elev` | `#161c26` | `--muted` / elevated | match |
| `--border` | `#1f2937` (dark) / `#e8e4d6` (light) | `--border` → `--color-border` | match |
| `--fg2` | `#cbd5e1` | `--fg-2` → `--color-fg-2` | match |
| `--fg-muted` | `#64748b` | `--fg-muted` → `--color-fg-muted` | match |
| Geist / Geist Mono | brand fonts | `--font-sans` / `--font-mono` | match |

## Acceptance criteria (from the handoff README — fidelity: high)

- **Status bar owns deploy**, 4 states; amber `3px` left edge only when changes pending; clean state has none.
- **One unified `DEPLOY TIMELINE`** (mono marker), newest-first, no Current/Earlier sub-headers.
- Every node = lean row (`#num cause [chip] · duration   timestamp chevron`) + detail card **below** (does not morph the row).
- Detail card body order: stage tracker → failure banner (if failed) → `RESOURCE OUTCOME` (mono marker, `● name  state  note` + `▢ image`) → `CONFIG CHANGES · vs #N` (expands to diff).
- **Draft node**: dashed amber ring dot, `1px dashed` amber card border, changelog diff only (no tracker/resources), note "↑ Deploy from the bar above to ship as release #N".
- **Live release**: green `LIVE` chip + green dot; open card gets `0.5px solid green` hairline; collapsed row lean; **no glow/box-shadow** (explicitly rejected).
- **Failed**: red chip, red failure banner, red duration. **Deploying**: amber chip, spinner dot, amber card border.
- Amber is the **only** accent (active/primary CTA/focus/markers); green & red are status-only. Flat — no shadows/glow. Mono only for markers/labels/status/filenames/sha.
- Section markers: Geist Mono 11px / 500 / 1.5px tracking / UPPERCASE.
- Tweakable: `showStages` (default true), `highlightLive` (default true).

## Intentional divergences from the handoff

1. **Rail dots** — README maps **live *and* released → solid green dot**, failed → ring, deploying → spinner. Per a later explicit user instruction ("make all the left bullets hollow rings except the live one"), only the **live** release is solid; every other node (including released) is a hollow ring. User instruction overrides the handoff. (`timeline-rail.tsx::dotShape`)
2. **Live anchor** — a compact pinned "Live #N" summary leads the tab *only when the live release is buried* below newer deploys, with Jump (opens + scrolls to the live node). Not in the handoff; added per user request to avoid scrolling a long timeline to find what's live. (`live-release-summary.tsx`)

## Source-file map (handoff bundle)

| Bundle path | Defines |
|---|---|
| `design_handoff_deployments_draft/README.md` | The task + intent + acceptance + tokens (authoritative) |
| `design_handoff_deployments_draft/Deployments (Draft node).dc.html` | High-fidelity HTML prototype (visual source of truth; do not render) |
| `design_handoff_deployments_draft/_ds/.../colors_and_type.css` | Design-system tokens + type (= a snapshot of `frontend/src/index.css`) |
| `design_handoff_deployments_draft/support.js` | Prototype toolbar/runtime (not production) |
| `Deploy Timeline.dc.html`, `Deployments - three directions.dc.html`, `Draft & Deploy - three approaches.dc.html` | Earlier explorations (superseded by the Draft-node direction) |
| `screenshots/`, `uploads/` | Iteration screenshots (reference only) |
