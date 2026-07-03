# Stack Canvas Editor: Floating Drawer Stack, Collapsible Header, Labels, PUBLIC Endpoint Row

**Date:** 2026-07-03
**Status:** Approved
**Design reference:** `docs/superpowers/design-refs/2026-07-03-stack-creation-redesign-design-reference.md` (distilled from Claude Design project `Stack Creation Redesign.dc.html` + `Panel Explorations.dc.html`)

## Goal

Bring the stack canvas editor in line with the Stack Creation Redesign: resource editing moves into floating, stackable drawers; the page header becomes collapsible; stack labels become editable at any lifecycle stage; and the header surfaces each service's public endpoint.

## Scope decisions (from brainstorming)

| Decision | Outcome | Why |
|---|---|---|
| Stack rename | **Dropped** | Backend rejects post-deploy rename (`pkg/validator/stack/stack_validator.go:92`); CR name and namespace derive from `stack.Name` (`pkg/builders/cluster_resource_builder.go:148`, `pkg/services/namespace_service.go:55`) — rename would orphan CRs. Draft rename already exists and stays as-is. |
| Labels post-deploy | **Always editable** | `Stack.labels` touches nothing derived; persists via existing autosave path. |
| PUBLIC row data | **Real status only** | URLs come from `StackResourceStatus.public_ingress`; no client-computed hosts. Generated FQDNs are hash-based (`pkg/services/exposed_port_domain.go`) and stable across rename and rolling deploys. |
| Endpoint pill density | **One pill per service, best URL** | Priority: custom domain > subdomain-prefix FQDN > generated FQDN. Full URL list remains in the resource drawer. |
| Status dot semantics | **Binary** | Green when `public_ingress` present; no deploy-state mapping (deferred). |
| Drawer push triggers | **Volume mount rows only** | Addon/secret bindings and service connection chips deferred. |
| Exploration 1f (stack-as-drawer) | **Deferred** | Stack meta stays in the header per the main prototype. |
| Drawer state architecture | **State array in `StackCanvasTab`** | Smallest evolution of existing `selectedIndex` model; context provider not justified yet. |

## 1. Floating drawer stack

### State

In `StackCanvasTab` (`frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx`), replace `selectedIndex: number | null` with:

```ts
type DrawerEntry =
  | { kind: 'resource'; index: number }   // stack resource by draft index
  | { kind: 'volume'; name: string };     // draft volume by name

drawerStack: DrawerEntry[]
```

Handlers (pure reducer, unit-tested):

- `replace(entry)` — canvas node click; stack becomes `[entry]`.
- `push(entry)` — reference click inside an open drawer. Dedupe: if `entry` already in the stack, truncate the stack to that entry instead of pushing a duplicate.
- `truncateTo(depth)` — click on a behind panel or breadcrumb.
- `pop()` — Esc.
- `closeAll()` — ⇧Esc or explicit close of the last panel.

### Geometry & layering (from design reference)

- Front panel: `position: fixed; top/right/bottom: 12px; width: 600px; z-index: 200`, border-radius 8px, `--shadow-2`, slide-in 260ms from `translateX(34px)` (design `sd-drawer` keyframes).
- Behind panel at depth `d` (1 = directly behind front): `top/bottom: 12 + 10d px; right: 12 + 16d px; z-index: 200 − d`, content covered by a scrim of `rgba(10,14,20,0.55)` — mapped to an overlay token, not raw rgba (see Tokens). Behind panels render header only (resource icon + name); clicking one truncates the stack to it.
- **No page backdrop.** Canvas remains visible and interactive; a canvas node click replaces the stack.
- Unbounded depth (practically limited by dedupe).

### Behavior changes

- Remove the canvas-resize-on-drawer-open logic (`StackCanvasTab.tsx:101-110`); drawers float above the canvas.
- `ResourceDrawer` (`frontend/src/pages/stacks/components/canvas/ResourceDrawer.tsx`) becomes the panel body for `kind: 'resource'` entries; a volume drawer body (reusing existing volume editing UI) serves `kind: 'volume'`.
- Volume mount rows inside a service drawer become clickable → `push({ kind: 'volume', name })`.
- Keyboard: Esc pops one; Shift+Esc closes all. Listener active only while the stack is non-empty; must not swallow Esc when a nested popover/select is open.
- Select/popover menus inside drawers portal to document root at z-index 300/301 so they never clip inside the panel.

### New component

`DrawerStack` — presentational; receives `drawerStack` + handlers, renders panels with depth offsets and the front panel's body. Lives beside `ResourceDrawer` in `frontend/src/pages/stacks/components/canvas/`.

## 2. Collapsible header

In `CanvasEditorShell` (`frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx`):

- `headerCollapsed: boolean`, persisted to localStorage keyed by stack ID; default expanded.
- Toggle: chevron button at header left, and ⌘. (mac) / Ctrl+. keyboard shortcut.
- **Expanded:** current header content — name h1 (29px), status pill, label chips, counts subtitle, plus new PUBLIC row, then the tab row with autosave status and Deploy button.
- **Collapsed:** single 44px bar — chevron, stack name at 14px, status dot, compact tab row, unsaved-changes summary, compact Deploy button. Labels, subtitle, and PUBLIC row hidden.
- Transition: fade/translate per design `sd-fade` (300ms `--ease`); no layout jank in the canvas below (canvas area grows to absorb the freed height).

## 3. Labels: always editable

- Remove the draft-only gating around the label chips UI (`CanvasEditorShell.tsx:227-254` renders; gating in detail page state).
- For deployed stacks, add/remove persists through the existing autosave path (full-payload PUT — the established autosave flow already sends complete connections/fields; labels ride along).
- Chip behavior (from design): mono 11.5px pills with ×-remove; dashed "+ label" ghost opens a 104px inline input; commit normalizes lowercase, whitespace→dash; dedupes; Enter commits and chains a new add; blur commits; Esc cancels. No max count, no validation hints.
- Save failure surfaces through the existing `AutosaveStatus` error state.

## 4. PUBLIC endpoint row

Rendered in the expanded header only, below the subtitle.

### Data

- Source: each stack resource's `StackResourceStatus.public_ingress` (array of `{ url, target_port }`) — already fetched with stack detail.
- Org domains from `Organisation.domains` (`DomainName.fqdn`) for URL classification.
- Row absent when no resource has a non-empty `public_ingress` (drafts, undeployed stacks, nothing exposed).

### Best-URL heuristic (pure function, unit-tested)

For a service's `public_ingress` URLs, pick by priority:

1. **Custom domain** — host does not end with any org domain.
2. **Prefix FQDN** — host under an org domain whose first label is not a generated token.
3. **Generated FQDN** — first label matches the generated pattern (16-char lowercase base32, per `EncodeStackResourceSubdomainPrefix` in `pkg/services/exposed_port_domain.go:32`).

Ties broken by array order.

### Pill anatomy (per design)

`PUBLIC` micro-label, then one 3-segment pill per exposed service:

1. Service chip — icon + service name; tooltip "Mapped to {name} · :{port}".
2. Link segment — 5px green dot + host; opens URL in new tab (external-link icon).
3. Copy button — copies URL; icon swaps to green check for 1400ms. Clipboard API with fallback.

Multiple pills wrap to additional lines.

## Out of scope

- Stack rename (any form, including display-name field).
- Exploration 1f (stack-as-resource drawer on empty-canvas click).
- Drawer push triggers for addon/secret bindings and service connection chips.
- Deploy-state → dot mapping and replica-progress tooltips.
- Custom domain management UI.

## Tokens

Per the design reference token audit: 6 same-name drifts and 13 raw hex/rgba values appear in the design source. Implementation maps every value onto live `frontend/src/index.css` tokens; where a needed step is missing (overlay/scrim, `--fg2`/`--fg3` equivalents, radius steps), reuse the nearest existing token rather than introducing new raw values. Any genuinely new token requires explicit sign-off before addition.

## Testing

- **Unit:** drawer-stack reducer (replace/push-dedupe-truncate/pop/closeAll), best-URL heuristic (custom vs prefix vs generated classification), label normalization (case, whitespace, dedupe).
- **Component:** DrawerStack renders correct offsets/z-indices for depth ≥ 2; collapsed header hides labels/PUBLIC row; PUBLIC row absent without ingress.
- **Manual:** Playwright pass against localhost:5173 — open service → push volume drawer → Esc pop → ⇧Esc close; collapse toggle + ⌘.; label add/remove on a deployed stack; endpoint copy/open.

## Error handling

- Label persistence failures: existing `AutosaveStatus` error surface; no bespoke toasts.
- Clipboard copy failure: fallback to `document.execCommand`/manual selection; no crash.
- Drawer for a resource/volume deleted from the draft while open: entry drops from the stack (stack reconciles against draft on change).
