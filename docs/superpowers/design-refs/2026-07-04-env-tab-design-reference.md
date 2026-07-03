# Design Reference: ENVIRONMENT Tab (Stack Editor Drawer)

**Date:** 2026-07-04
**Sources read (verbatim, no rendering):**
- `/Users/akshaysasidharan/.claude/jobs/3558ca04/tmp/stack-creation-redesign.dc.html` (1,244 lines) — full-page prototype
- `/Users/akshaysasidharan/.claude/jobs/3558ca04/tmp/env-row.dc.html` (42 lines) — extracted `EnvRow` component
- `/Users/akshaysasidharan/code/stackdome/.claude/worktrees/stack-canvas-editor/frontend/src/index.css` (304 lines) — live design tokens

**Cross-reference:** `docs/superpowers/design-refs/2026-07-03-stack-creation-redesign-design-reference.md` — an earlier session's design-ref for the *same* Claude Design job, produced with access to the bundle's `colors_and_type.css` (not available to this task). Its verified §2 Token Mapping table is reused below for tokens this task's two files reference but don't resolve (e.g. exact hex/alpha of `--fg2`, `--fg3`, `--amber-soft`, `--r-2`).

## 0. Critical scope gap — read this first

The ENVIRONMENT tab's actual rendered markup (tab header row, "paste .env"/"raw editor" buttons, the "+ Add variable" ghost row, the ADDON group container chrome, and the CONFIGURATION/ENVIRONMENT tab strip itself) lives inside a **`dc-import name="ResourceConfig"`** component (main file line 447: `<dc-import name="ResourceConfig" r="{{ frontRes }}" hint-size="100%,360px"></dc-import>`). That component has its own `.dc.html` file in the Claude Design project (confirmed by the 2026-07-03 doc: *"each has its own `.dc.html` in the project"* — `ResourceConfig.dc.html`), but **it was not included in either file given for this task**. I grepped exhaustively (case-insensitive) for `raw`, `paste`, `.env`, `ADDON` (as UI text), `Add binding`, `Add variable`, `variables`, `EnvRow`, `AddonGroup` across the full 1,244-line main file — zero literal markup hits outside of JS identifiers and data-shaping code.

What **is** available and fully authoritative:
1. `env-row.dc.html` — the single-row anatomy (§b below), complete and exact.
2. The main file's JS state/data-shaping layer (`buildEnvRow`, `buildEnvRows`, `envStatus`, `envRowEqual`, lines 700–844) — this reveals the **exact color logic** for every row state, the **exact grouping key** for ADDON binding groups, and the handlers (`onAddVariable`, `onAddBinding`, `onReset`, `onRemove`) that any header/ghost-row/group-header markup would wire to. This is reliable even though the DOM it feeds isn't visible to me.
3. The drawer chrome around `ResourceConfig` (crumbs, header, DEPENDS ON chips, esc hint) — fully present, lines 417–453.

Sections **a, c, g, h** below are therefore marked **NOT FOUND IN SOURCE** for markup/CSS, with whatever indirect JS evidence exists called out separately. Sections **b, d (partial), e, f** are fully specified. **To complete this reference, `ResourceConfig.dc.html` needs to be pulled from the same Claude Design job/project and read the same way.**

---

## a. Tab header row — "Environment · N variables" / paste .env / raw editor

**NOT FOUND IN SOURCE.** No literal text "Environment", "variables", "paste .env", or "raw editor" appears anywhere in `stack-creation-redesign.dc.html` or `env-row.dc.html`.

Indirect evidence only: `buildEnvRows(iid)` (main file, line 843) returns `{ count, hasPlain, plainRows, addonGroups, onAddVariable }`. `count = rows.length` (line 822) is almost certainly the "N" a header would interpolate into "Environment · N variables", and `onAddVariable` is almost certainly what a "+ Add variable" affordance (wherever it lives) calls. But the header's typography, layout, and the paste-.env / raw-editor buttons (icons, whether it's a modal or inline textarea toggle) have no markup in either provided file — cannot be specified without `ResourceConfig.dc.html`.

## b. Env row card — FULLY SPECIFIED

Source: `env-row.dc.html` lines 10–39 (structure) + main file `buildEnvRow`/`buildEnvRows` lines 771–844 (state → color/prop bindings).

**Outer card** (`env-row.dc.html:10`):
```
position:relative; border-radius:var(--r-2); background:{{ row.rowBg }};
border:0.5px solid {{ row.border }}; padding:9px 10px 9px 13px;
transition:background .18s var(--ease), border-color .18s var(--ease)
```
Note asymmetric padding — left is 13px (3px extra) to clear the 2px accent bar + 3px gap from it.

**Accent bar** (`env-row.dc.html:11`): `position:absolute; left:0; top:6px; bottom:6px; width:2px; border-radius:2px; background:{{ row.accent }}` — a 2px vertical bar inset 6px from top/bottom of the row.

**Row states — bound in `buildEnvRow` (main file lines 771–775):**
```js
const dirty = st !== 'unchanged';                       // st ∈ {'unchanged','modified','added'}
const accent = dirty ? 'var(--amber)' : 'transparent';
const rowBg  = dirty ? 'rgba(249,115,22,0.06)' : 'transparent';
const border = dirty ? 'rgba(249,115,22,0.35)' : 'var(--border)';
```
`st` comes from `envStatus()` (line 756–760): `'added'` if no row in `baseEnv` shares the row's `name`; `'modified'`/`'unchanged'` from a field-by-field `envRowEqual()` diff otherwise (per-`from`-type: value / secretId+secretKey / addonId+database+credField / resourceName+output / selfOutput).

**Important finding — only 2 visual row states, not 3:** "added" (new) and "modified" (dirty) rows render **identically** — same amber accent bar, same 6%-amber row tint, same 35%-amber border. There is no distinct color treatment for "new" vs "dirty" at the row-surface level. The only place the two diverge is the trailing action icon (see below): modified rows get a Reset icon, added (and unchanged) rows get a Remove icon, because a newly-added row has no prior value to reset to.

**Inner flex row** (`env-row.dc.html:12`): `display:flex; align-items:flex-start; gap:8px`.

**KEY input** (`env-row.dc.html:13`):
```
width:130px; flex:none; padding:6px 9px; background:var(--bg);
border:0.5px solid var(--border); border-radius:var(--r-2); color:var(--fg1);
font-size:12px; class="mono"; placeholder="KEY"; spellcheck=false; outline:none;
transition:border-color .15s var(--ease)
focus → border-color:var(--amber)
```

**Value column** (`env-row.dc.html:14`): `flex:1; min-width:0; display:flex; flex-direction:column; gap:6px`. Children are mutually exclusive on `row.from` (`row.isStack|isSecret|isAddon|isResource|isSelf` — computed in `buildEnvRow` lines 787–817):

| `from` | Stacked controls | Height each |
|---|---|---|
| `stack` (Plain text) | 1× text `<input>`, width 100%, same padding/border/radius/font as KEY but `color:var(--fg2)` (dimmer than KEY's `--fg1`), placeholder "value", mono, focus border amber | matches KEY row |
| `secret` | `MiniSelect` (secretTrigger) always; **+** conditional `MiniSelect` (keyTrigger) if `row.hasKey` (a secret is picked) | 32px each (`hint-size="100%,32px"`) |
| `addon` | 1× `MiniSelect` (credTrigger) — picks the credential field | 32px |
| `resource` | `MiniSelect` (resTrigger) always; **+** conditional `MiniSelect` (outTrigger) if `row.hasOut` (a sibling resource is picked) | 32px each |
| `self` | 1× `MiniSelect` (selfTrigger) | 32px |

All `MiniSelect` triggers are built via the shared `trigger()` helper (main file lines 674–687): clicking opens a floating popover anchored under the trigger (`openSelect`, lines 663–672) — **not** a native `<select>`. Trigger border is `var(--border)` normally, `var(--amber)` while its popover is open; the chevron rotates 180° when open; unselected/placeholder text renders `var(--fg-muted)`, selected text renders the caller-supplied color (`var(--fg1)` almost everywhere, `var(--fg2)` for the From trigger).

**From select** (`env-row.dc.html:33`): `width:106px; flex:none`, one `MiniSelect` (`fromTrigger`, 106×32). Options (`fromDefs`, line 776): `stack`→"Plain text"/`code` icon, `secret`→"Secret"/`lock`, `addon`→"Addon"/`plug`, `resource`→"Resource"/`workflow`, `self`→"Self"/`layers`. Icon color `var(--fg3)`; trigger text color `var(--fg2)`.

**Action cell** (`env-row.dc.html:34–37`): `width:24px; flex:none; flex center; height:32px`.
- `row.isModified` (st==='modified'): 24×24 button, `border-radius:var(--r-2)`, icon `rotateccw` size 14, color `var(--amber)`, hover `background:var(--amber-soft)`, title "Reset to original value", `onClick={{ row.onReset }}` → `envReset()` (line 728) restores the row from `baseEnv` by matching `name`.
- `row.notModified` (st==='unchanged' OR 'added' — default true): 24×24 button, icon `x` size 14, color `var(--fg-muted)`, hover `background:var(--err-soft); color:var(--err)`, title "Remove variable", `onClick={{ row.onRemove }}` → `envRemove()` (line 727).

### Token map — §b

| Prototype token / literal | Used for | Live `index.css` token | Status |
|---|---|---|---|
| `var(--amber)` | accent bar (dirty), Reset icon color, KEY/value focus border | `--brand` `oklch(0.72 0.20 40)` (= `#f97316`, matches index.css comment) | MATCH (renamed) |
| `rgba(249,115,22,0.06)` | dirty row bg | no exact live token — nearest is `--brand-bg` at **10%**, not 6% | **UNMAPPED — bespoke 6% tint**, drift vs `--brand-bg` |
| `rgba(249,115,22,0.35)` | dirty row border | nearest is `--brand-border` at **30%**, not 35% | **UNMAPPED (close)** — 35% vs 30%, drift |
| `var(--amber-soft)` | Reset-icon hover bg | `--brand-bg` (10%) — prior verified doc resolves bundle `--amber-soft` to `rgba(249,115,22,0.12)` (12%) | DRIFT — 12% (bundle) vs 10% (live) |
| `var(--border)` | default row border, all input borders | `--border` | MATCH |
| `var(--bg)` | KEY/value input fill | `--background` | MATCH |
| `var(--fg1)` | KEY input text, selected-option text | `--foreground` | MATCH (renamed) |
| `var(--fg2)` | value-input text, From-trigger text | `--fg-2` | MATCH (name literally matches: `fg2`→`fg-2`) |
| `var(--fg3)` | From-option icon color | **no live equivalent** — index.css light mode explicitly documents only 2 ink tones (fg, muted); dark mode's `--muted-foreground` happens to equal bundle "fg3" but isn't exposed under that name | **UNMAPPED — 3rd-tier ink tone missing from live token set** |
| `var(--fg-muted)` | placeholder text, unchanged-row Remove icon | `--fg-muted` | MATCH (identical name) |
| `var(--err)` / `var(--err-soft)` | Remove-icon hover color/bg | `--danger` / `--danger-bg` (task hint confirms err→danger); index.css also has a separate shadcn `--destructive` token — do not conflate | MATCH (renamed), alpha likely drifts (bundle soft ≈16% per prior doc vs live `--danger-bg` 10%) |
| `var(--r-2)` | card radius, accent-bar radius, input radius, action-button radius | `--radius-sm`? or `--radius-md`? — prior verified doc resolves bundle `--r-2 = 4px`, which matches live `--radius-md: 4px` (index.css comment: "controls 'md' = 4px") | MATCH → **`--radius-md`** |
| `var(--ease)` | background/border-color transitions | prior doc resolves to `cubic-bezier(0.2,0.6,0.2,1)` — **no live `--ease` token exists** | UNMAPPED — needs to be added or hardcoded |
| `class="mono"` | KEY/value input font | `var(--font-mono)` → live `--font-mono` (Geist Mono) | MATCH |

## c. "+ Add variable" ghost row

**NOT FOUND IN SOURCE.** No markup for this affordance exists in either file. Only the handler is confirmed: `onAddVariable: ()=>this.envAdd(iid)` (main file line 843), where `envAdd()` (line 726) appends `{ id, name:'', from:'stack', value:'' }` — i.e. a brand-new row always defaults to the "Plain text" (`stack`) source type with an empty value.

For reference only (not a confirmed match — flagged as analogy, not fact): the prototype's one other "ghost add" affordance in this file is the label-add pill (main file line 92): `border:0.5px dashed var(--border); border-radius:var(--r-pill); color:var(--fg-muted)`, hover → `border-color:var(--amber); color:var(--amber)`. If `ResourceConfig.dc.html` follows the same design language, a dashed-border/muted-to-amber-on-hover treatment is plausible for "+ Add variable" — but this must be verified against the actual file, not assumed.

## d. ADDON binding group container — PARTIALLY SPECIFIED (logic yes, DOM/CSS no)

DOM/CSS **not found** (same `ResourceConfig.dc.html` gap). Grouping logic and data shape are fully specified from `buildEnvRows` (main file lines 821–843):

- **Grouping key confirms "per addon-connection":** rows with `from==='addon'` are grouped by `key = (row.addonId||'') + '|' + (row.database||'')` (line 825) — i.e. one group per **(addon, database) pair**, not one group per addon type. An addon with two databases bound (e.g. `tooljet-db`/`tooljet` and `tooljet-db`/`analytics`) would render as **two separate group boxes**, not one.
- **Group header data** (lines 829–839): `addonTrigger` (MiniSelect, picks which addon this group points at, from `ADDONS` catalog — icon `database`, iconColor `var(--fg3)`, trigger icon `database`/`var(--amber)`); `hasDbSelect` (true only if the addon has >1 database) gating between a `dbTrigger` MiniSelect or a static `dbLabel` text fallback (`db || dbList[0] || '—'`).
- **Binding rows inside the group** (line 840, `rows: groupsMap.get(key).map(mk)`): built via the **same** `buildEnvRow()` as plain rows — meaning each binding row still renders its full KEY input + From-select (labeled "Addon", icon `plug`) on the right, exactly per §b's anatomy. This confirms the task's expectation: "binding rows inside (KEY → field select + 'Addon' from select)". The addon-row's stacked value column is a single `credTrigger` MiniSelect (list from `CRED_FIELDS = ['host','port','database','username','password','sslmode','ca_certificate','url']`, with a "CLUSTER" sub-label badge on `CLUSTER_FIELDS = ['host','port','sslmode','ca_certificate']`).
- **"+ Add binding"**: `onAddBinding: ()=>this.groupAddBinding(iid, addonId, db)` (line 840), calling `groupAddBinding()` (line 744) which appends `{ id, name:'', from:'addon', addonId, database:db, credField:'' }` — pinned to the same addon+database as the group it was added from.
- **Switching a group's addon/database is destructive-rebind, not per-row**: `groupSetAddon`/`groupSetDb` (lines 734–743) rewrite `addonId`/`database` on **every row currently in that group**, so changing the group header's addon changes it for all bindings inside at once.

Because none of the container's border/background/spacing exists in the provided files, I cannot confirm "dashed border color, bg" as stated in the task's framing — that description may be based on knowledge of `ResourceConfig.dc.html` the task author has that wasn't attached here.

## e. DEPENDS ON chips strip — FULLY SPECIFIED

Source: main file lines 438–445, gated on `frontRes.hasDeps` (true when `depChips.length>0`, built at lines 872–873 from the resource catalog's `dependsOn` string, e.g. `'db · cache'`).

```
Container: flex:none; display:flex; align-items:center; gap:8px;
  padding:10px 14px; border-bottom:0.5px solid var(--border); background:var(--bg-card)
Label: class="t-marker"; color:var(--fg-muted); font-size:9px; text "DEPENDS ON"
```
Each chip (`sc-for depChips`):
```
display:inline-flex; align-items:center; gap:6px; padding:4px 10px;
border:0.5px solid var(--border); border-radius:var(--r-pill); font-size:12px;
color:var(--fg2); cursor:pointer
hover → border-color:var(--amber); color:var(--amber)
```
Chip content: leading `Glyph` icon (`d.icon`, size 13) + `d.name` + trailing `arrowRight` glyph (size 12, `var(--fg-muted)`). **Click behavior:** `onClick={{ d.onPush }}` → `pushPanel(d.iid)` (wired at line 1147) — clicking a dependency chip **pushes a new stacked drawer panel** for that dependency resource (the "behind panels" peeking-panel stack), it does not navigate away or open elsewhere.

## f. Drawer header changes chip / esc hint — FULLY SPECIFIED

Source: main file lines 417–437.

**"N changes ×" chip** (line 434), gated `frontRes.isDirty` (`changeCount>0`, line 899):
```
display:inline-flex; align-items:center; gap:5px; border:0.5px solid var(--amber);
background:var(--amber-soft); border-radius:var(--r-2); padding:2px 4px 2px 9px;
font-size:11px; font-weight:500; color:var(--amber)
```
Text = `frontRes.dirtyLabel` = `changeCount===1 ? '1 change' : changeCount+' changes'` (line 900) — note: singular is "1 change" not "1 changes", and there's no literal "×" glyph text; the close affordance is a 16×16 `x` Glyph (size 12) button nested at the chip's trailing edge:
```
width:16px; height:16px; border-radius:3px; cursor:pointer
hover → background:var(--amber)
onClick → frontRes.onDiscard (discardResource(iid), line 901) — discards ALL changes for this resource (config + deploy + env), not just env
```
When not dirty (`frontRes.notDirty`, line 435): shows `frontRes.kindLabel` (e.g. "Postgres") styled as `class="t-marker"; color:var(--fg-muted)` instead of the chip.

`changeCount` is the sum of `cfgDirty` fields + `deployDirty` fields + `envChangeCount(iid)` (line 860) — so the ENVIRONMENT tab's dirty count folds into the *same* aggregate number shown at the drawer-header level; there's no separate "N env changes" counter at this location (only the per-resource total).

**"esc closes" hint** (line 428): `class="mono"; font-size:10px; color:var(--fg-muted)`, text = `escHint` (line 1212): `stackArr.length>1 ? 'esc pops · ⇧esc closes all' : 'esc closes'` — i.e. the hint text changes based on how many panels are stacked (single panel → "esc closes"; 2+ stacked panels, e.g. after clicking a DEPENDS ON chip → "esc pops · ⇧esc closes all"). Positioned top-right of the drawer, right-aligned via `<span style="flex:1"></span>` spacer preceding it (line 427), same row as the breadcrumb trail (`hasCrumbs`).

## g. Tab row dirty dots (CONFIGURATION • / ENVIRONMENT •)

**NOT FOUND IN SOURCE** as the literal two-tab strip described in the task. The `editorModes` tabs found in this file (lines 1159–1167: Configuration/Deployments/Logs/Metrics, labels `m.label`, `m.tabBorder`/`m.tabBg` = amber when active) are the **top-level page tabs** (canvas vs deploy vs logs vs metrics), a different UI element from a CONFIGURATION/ENVIRONMENT sub-tab strip *inside* the resource drawer.

Closest related evidence — `modTabs` (main file lines 875–879), three single-letter dot/pill indicators per resource:
```js
const modTabs = [
  { label:'C', on:cfgDirty, title:'Configuration changed' },
  { label:'D', on:depDirty, title:'Deployment changed' },
  { label:'E', on:envDirty, title:'Environment changed' },
].map(t => ({
  label:t.label, title:t.title,
  color:  t.on ? 'var(--amber)'      : 'var(--fg-muted)',
  bg:     t.on ? 'var(--amber-soft)' : 'transparent',
  border: t.on ? 'var(--amber)'      : 'var(--border)',
}));
```
This is data only (`buildRes()` returns `modTabs` at line 899 but I found no markup consuming it in the two provided files) — it's very plausibly what feeds a CONFIGURATION•/ENVIRONMENT•-style dirty-dot tab strip inside `ResourceConfig.dc.html`, but that is an inference, not a confirmed match (note it's 3 letters C/D/E, not 2 tabs). If/when `ResourceConfig.dc.html` is available, verify against this exact color logic: dirty = amber text + amber-soft bg + amber border; clean = fg-muted text + transparent bg + default border.

## h. Raw-editor / paste-.env flow markup

**Confirmed absent.** Exhaustive case-insensitive grep for `raw`, `paste`, `\.env`, `textarea` (env-context) across the full main file found: zero matches for "raw" or "paste" anywhere; the only `<textarea>` in the entire file is the Docker Compose YAML importer (line 250, unrelated — `composeYaml`/`onComposeYaml`, in the wizard's "Import from Docker Compose" phase, not the env tab). There is no modal, no inline-textarea toggle, no bulk-paste handler for env vars anywhere in the two provided files. This flow, if it exists at all in the design, is entirely inside `ResourceConfig.dc.html`.

---

## Interaction notes

- **MiniSelect popover pattern** applies uniformly to every stacked select in the row (secret/key, addon cred, resource/output, self, from): click opens a floating menu (`openSelect`) positioned via `getBoundingClientRect()` with viewport-edge flipping (lines 663–671), not a native `<select>`. Selecting an item calls `closePopover()` immediately (single-select, no explicit confirm).
- **Reset vs Remove is state-driven, not row-type-driven**: any row can show either icon depending purely on `st` (dirty-modified → Reset; unchanged or newly-added → Remove). This is a common trap to get backwards when re-implementing — it is *not* "existing rows get Reset, new rows get Remove" as a static rule; it's keyed off equality-to-baseline.
- **`envReset()` (line 728) falls back to `envRemove()`** if no baseline row matches by name — i.e. clicking Reset on a row whose name was changed away from anything in `baseEnv` just deletes it (there's no baseline to restore to).
- **Addon group add/remove semantics differ from plain rows**: `groupAddBinding` always seeds the new row with the group's current `addonId`/`database` — a binding row cannot be created "unbound"; it's always born inside a specific addon-connection group. There's no visible mechanism in the read JS for moving a binding row between groups other than editing its own `credTrigger`/using the row's own From-select to change type away from `addon` entirely (which would re-bucket it out of the group via `buildEnvRows`' grouping-by-`from` pass).
- **Dirty aggregation**: the drawer header's "N changes" chip (§f) counts config+deploy+env changes together; there is no isolated "N environment changes" figure exposed anywhere in the read markup — if the ENVIRONMENT tab header needs its own count, it would use `env.count` (total row count, not diff count) per §a's inference, not `envChangeCount()`.

## Source line references

| Item | File | Lines |
|---|---|---|
| EnvRow full markup | `env-row.dc.html` | 10–39 |
| `buildEnvRow` (row state → accent/rowBg/border/icons) | main file | 771–819 |
| `buildEnvRows` (grouping, plain vs addon, count, onAddVariable) | main file | 821–844 |
| `envStatus` / `envRowEqual` (dirty/added/unchanged classification) | main file | 746–760 |
| `envAdd` / `envRemove` / `envReset` | main file | 726–732 |
| `groupSetAddon` / `groupSetDb` / `groupAddBinding` | main file | 734–744 |
| `trigger()` / `openSelect()` (MiniSelect popover mechanics) | main file | 662–687 |
| `SECRETS` / `ADDONS` / `CRED_FIELDS` / `CLUSTER_FIELDS` / `OUTPUTS` catalogs | main file | 603–613 |
| Drawer front-panel container, crumbs, esc hint | main file | 417–429 |
| Drawer title row + dirty "N changes ×" chip | main file | 430–437 |
| DEPENDS ON chips strip | main file | 438–445 |
| `ResourceConfig` mount point (body of drawer) | main file | 447 |
| `modTabs` (C/D/E dirty-dot data) | main file | 875–879 |
| `buildRes()` (assembles `frontRes`, incl. `env: buildEnvRows(iid)`) | main file | 847–904 |
| `editorModes` (top-level Configuration/Deployments/Logs/Metrics tabs — NOT the drawer's inner tabs) | main file | 1159–1167 |
| `escHint` definition | main file | 1212 |
| Docker-compose textarea (only `<textarea>` in file; unrelated to env tab) | main file | 250 |

## Gaps to close

To fully answer items **a, c, g, h** and the DOM/CSS half of **d**, fetch `ResourceConfig.dc.html` from the same Claude Design job/project (job `3558ca04`, or the project referenced by the 2026-07-03 doc as `faed4868-7719-4ab1-9b21-ff8eae2933ff`) and re-run this same read-only extraction against it.
