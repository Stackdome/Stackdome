# Canvas Editor — Visual Implementation Spec

Source: Claude Design bundle `stackdome-stack-creation-redesign` (DC component format).
This spec transcribes the exact px / color / spacing values from the bundle so a
React + Tailwind engineer can rebuild the stack-editor UI faithfully **without opening
the bundle**.

## Token translation (bundle → live code)

The bundle CSS uses the design-system variable names below. Where the live Stackdome
codebase uses a different token name, use the **live** name. All raw values are also
given in the Token Values section so nothing is ambiguous.

| Bundle var | Live-code var | Value (dark) |
|---|---|---|
| `--amber` | `--brand` | `#f97316` |
| `--amber-hover` | `--brand` hover | `#ea6a0e` |
| `--amber-soft` | `--brand-bg` | `rgba(249,115,22,0.12)` |
| `--bg` | `--bg` | `#0a0e14` |
| `--bg-card` | `--bg-card` | `#11161e` |
| `--bg-elev` | `--bg-elev` | `#161c26` |
| `--fg1` | `--fg` | `#ffffff` |
| `--fg2` | `--fg-2` | `#cbd5e1` |
| `--fg3` | `--muted-foreground` | `#94a3b8` |
| `--fg-muted` | `--fg-muted` | `#64748b` |
| `--border` | `--border` | `#1f2937` |
| `--ok` | (state ok) | `#22c55e` |
| `--warn` | (state warn) | `#eab308` |
| `--err` | (state err) | `#d9223e` |
| `--r-2` | `rounded-sm` | `4px` |
| `--r-3` | `rounded-md` | `6px` |
| `--r-4` | `rounded-lg` | `8px` |
| `--r-pill` | `rounded-full` | `999px` |

Notes:
- The bundle's "amber" **is** the brand orange — there is only one brand color. Semantic
  green/yellow/crimson (`--ok`/`--warn`/`--err`) are used **only** for status dots, status
  pills, and deploy outcomes, never as decoration.
- Almost every border in this UI is a **hairline: `0.5px solid var(--border)`** (not 1px).
  The node card and canvas dot grid are the notable `1px`/sub-px exceptions (see below).
- Text on amber-filled buttons is **`#1a0e05`** (near-black warm), not white.
- Fonts: sans = **Geist**, mono = **Geist Mono**. Mono is used for summaries, ports,
  breadcrumbs, env keys/values, kind badges, log lines, metric numbers subtitles.

---

## Editor Shell

**Full-bleed, not boxed.** The entire app fills the viewport with a flex row; there is no
max-width container around the editor. Structure:

```
<div h:100vh w:100% display:flex overflow:hidden bg:--bg color:--fg1 font:Geist>
  <aside 232px sidebar/>          ← fixed-width left nav
  <main flex:1 minw:0 column>     ← everything else
    <topbar 52px/>                ← breadcrumb bar
    <editor region flex:1 column>
      <stack-title header/>       ← only when inEditor
      <tab row/>                  ← only when inEditor
      <mode body flex:1/>         ← canvas | deployments | logs | metrics
    </editor region>
  </main>
</div>
```

### Left sidebar (232px)
- `width:232px; flex:none; height:100%; border-right:0.5px solid var(--border); background:var(--bg)`, column flex.
- **Brand header** (clickable, opens create-stack wizard): `padding:18px; gap:11px; border-bottom:0.5px solid var(--border); cursor:pointer`, hover `background:var(--bg-card)`.
  - Glyph tile: `32×32; border-radius:7px; background:var(--amber-soft); border:1px solid var(--amber); color:var(--amber)`, centered `layers` glyph size 17.
  - Text block: wordmark `stackdome` (`.wordmark`, 15px, `--fg1`) over `.t-marker` `DEFAULT` (10px, `--fg-muted`).
- **Nav** (`padding:18px 12px; flex:1`): section marker `.t-marker` `PLATFORM` (`padding:0 10px 12px`), then items col `gap:2px`.
  - Item: `gap:11px; padding:9px 10px; border-radius:var(--r-2); font-size:14px`.
  - Active item (Stacks): `background:var(--amber-soft); color:var(--amber); font-weight:500`. Inactive: `color:var(--fg3)`, hover `color:var(--fg1); background:var(--bg-card)`.
  - Items + icons: Stacks(`layers`), Secrets(`key`), Object Stores(`cloud`), Addons(`grid`), Clusters(`boxes`), Domains(`globe`), all glyph size 17.
- **User footer**: `gap:11px; padding:14px 16px; border-top:0.5px solid var(--border)`. Avatar `30×30; rounded-sm; bg:--bg-elev; border:0.5px --border; font 12px/600 --fg2` (initials "CN"); name `13px/500 --fg1`; email `.mono 10.5px --fg-muted` (ellipsis).

### Top bar (52px breadcrumb)
- `height:52px; flex:none; display:flex; align-items:center; justify-content:space-between; padding:0 20px; border-bottom:0.5px solid var(--border)`.
- Left cluster `gap:14px`: `panelLeft` glyph (18, `--fg3`, hover `--fg1`); a `0.5px×20px` vertical divider (`background:--border`); breadcrumb `.mono 13px`: `Home / Stacks / acme-platform`, slashes `--fg-muted`, segments `--fg3` with the **last** segment `--fg1`.
- Right: `moon` glyph (17, `--fg3`) in a `6px`-padded rounded-sm hit target, hover `--fg1` on `--bg-elev`.

### Stack-title header (only when in editor)
`flex:none; padding:24px 28px 0`. Row `gap:14px; align-items:center`:
- `<h1>` stack name — `font-size:29px; font-weight:500; letter-spacing:-0.02em; color:--fg1; margin:0`.
- **Status pill** (READY): inline-flex, `gap:7px; border:0.5px solid var(--ok); background:var(--ok-soft); border-radius:var(--r-pill); padding:3px 11px`. Dot `6×6 round bg:--ok`; label `.t-marker` `READY` (`--ok`, 9.5px). (Pill color swaps to `--warn`/`--err` per state.)
- Spacer `flex:1`.
- **Dirty summary** (only if `dirtyAny`): `.mono 11.5px color:--amber`, text = `"1 unsaved change"` or `"N unsaved changes"`.
- **Deploy button**: `background:var(--amber); border:none; color:#1a0e05; padding:8px 18px; border-radius:var(--r-2); font-size:13px; font-weight:600; letter-spacing:.02em; gap:7px`; leading `rocket` glyph size 15; hover `background:var(--amber-hover)`.
- Overflow `dots` glyph (18, `--fg3`) in a 6px rounded-sm hit target.
- Subtitle row under the title: `margin-top:7px; font-size:13px; color:--fg3` — text like `"3 services · 2 volumes"` (the `stackSubtitle`).

---

## Tab Row

Sits directly under the stack title, only when in the editor.
`flex:none; display:flex; align-items:center; gap:8px; padding:18px 28px; border-bottom:0.5px solid var(--border)`.

Four tabs from `editorModes`: **Configuration** (`grid`), **Deployments** (`rocket`),
**Logs** (`terminal`), **Metrics** (`activity`). Each tab:

- `display:flex; align-items:center; gap:8px; padding:8px 15px; border:1px solid <tabBorder>; border-radius:var(--r-2); font-size:14px; font-weight:500; color:<color>; background:<tabBg>`, transition color+border-color `.15s`.
- Leading glyph size 15.
- **Active** (`mode===id`): `color:var(--amber)`, `border:1px solid var(--amber)`, `background:var(--amber-soft)`.
- **Inactive**: `color:var(--fg3)`, `border-color:transparent`, `background:transparent`, hover `color:var(--fg1)`.
- **Count badge** — only on **Configuration**, only when `dirtyResources > 0`: a pill after the label showing the count (e.g. `2`). `.mono 9.5px; color:var(--amber); background:var(--amber-soft); border-radius:var(--r-pill); padding:1px 6px; margin-left:1px`. Value = number of resources with pending config/deploy/env changes.

(Contextual meta string `editorMeta` — e.g. `"3 resources · 2 connections"`, `"release #12 · live"` — is computed but rendered as the header subtitle, not in the tab row.)

---

## Canvas Surface (Configuration mode)

Edge-to-edge, fills the whole mode body. Two nested layers:

- Outer: `height:100%; position:relative; overflow:hidden`.
- Canvas plane: `position:absolute; inset:0; overflow:hidden` with the **dot grid**:
  - `background-image: radial-gradient(var(--border) 0.7px, transparent 0.7px);`
  - `background-size: 24px 24px;`
  - `background-color: var(--bg);`
  - i.e. a `0.7px` border-colored dot every `24px` in both axes, on the base bg. Dots are faint (`--border` = `#1f2937`).

Children of the plane: the edges `<svg>` (behind nodes, see Edges), then the absolutely-positioned node cards (see Node Card). Node positions come from `posStyle = position:absolute; left:Xpx; top:Ypx; z-index:(selected?5:2)`.

**Add-resource button** — top-left overlay: `position:absolute; left:16px; top:16px`. Button:
`display:flex; gap:8px; background:var(--bg-elev); border:0.5px solid var(--border); color:var(--fg1); padding:9px 14px; border-radius:var(--r-2); font-size:13px; font-weight:500; box-shadow:var(--shadow-2)`; leading `plus` glyph 16; hover `border-color:var(--amber); color:var(--amber)`. Opens a centered "Add a resource" modal (600px max, search + BlockPicker, "Done" amber button).

**Bottom hint** — centered near bottom: `position:absolute; left:50%; transform:translateX(-50%); bottom:18px; gap:8px; font-size:11.5px; color:var(--fg-muted); pointer-events:none`; leading `move` glyph 13; text:
`"drag to rearrange · click a node to configure · edges carry connection env vars"`.

---

## Node Card

Absolutely positioned on the canvas; the outer positioned wrapper carries `onMouseDown` for drag. The card itself:

- `width:216px; border:1px solid var(--border); border-radius:var(--r-3); background:var(--bg-card); overflow:hidden; cursor:grab;` + a state-dependent `cardStyle` (below).
- Fixed layout width **216px**; canvas edge math assumes card **W=216, H≈58**.

**Header block** (`padding:12px 13px`), a row `gap:10px; align-items:center`:
1. **Status dot** `8×8; flex:none; border-radius:50%` with `dotStyle` = one of
   `background:var(--ok)` / `var(--warn)` / `var(--err)` (mapped from the catalog `dot` field; e.g. web/db/cache = ok, worker/minio = warn).
2. **Glyph** — `color:var(--fg2)`, size 17. Icon per kind (web=`globe`, postgres=`database`, redis=`zap`, worker=`cpu`, object/minio=`cloud`).
3. **Name** — `font-size:14px; color:var(--fg1); font-weight:500; flex:1`.
4. **Kind badge** — `.t-marker` (mono uppercase tracked), `font-size:9px; color:var(--fg-muted)`, right-aligned. Text = catalog `kindLabel`: **Web / Postgres / Redis / Worker / Object** (title-case, rendered uppercase by `.t-marker`). Derived from the resource's kind, not free text.

**Summary line** — under the header row, `.mono; font-size:11px; color:var(--fg-muted); margin-top:7px; padding-left:18px` (indented to align past the dot), single-line ellipsis. Examples (verbatim from catalog):
- `node · :8080 · public`
- `postgres:16 · 10Gi volume`
- `redis:7 · in-memory`
- `background queue consumer`
- `minio · S3-compatible`
- `managed postgres · backups on`

**Volume chip(s)** — zero or more rows appended below the header, each:
`display:flex; align-items:center; gap:8px; padding:8px 13px; border-top:0.5px solid var(--border); background:var(--bg)`:
- `drive` glyph (13, `--fg-muted`, flex:none).
- Volume name `.mono 10.5px color:--fg3` (flex:none).
- **Fill bar**: track `flex:1; height:4px; border-radius:3px; background:var(--bg-elev); overflow:hidden`; fill `display:block; height:100%; width:<pct>%; background:var(--amber); border-radius:3px`.
- Percent label `.mono 10px color:--fg-muted` (flex:none), e.g. `62%`. `pct = round(used/size*100)`, clamped 0–100.

**Card states** (`cardStyle`, applied on top of the base border):
- Selected + dirty: `border-color:var(--amber); box-shadow:var(--shadow-amber), inset 3px 0 0 var(--amber)`.
- Selected (clean): `border-color:var(--amber); box-shadow:var(--shadow-amber)`.
- Dirty (not selected): `border-color:rgba(249,115,22,0.55); box-shadow:inset 3px 0 0 var(--amber)`.
- Default: base `1px solid var(--border)` only.
- `--shadow-amber = 0 0 0 3px rgba(249,115,22,0.20)` (amber focus halo). The `inset 3px 0 0 var(--amber)` is a **left accent stripe** marking unsaved changes.

---

## Edges

Rendered as one absolutely-positioned SVG behind the nodes, only when `showConn` is true:
`position:absolute; inset:0; width:100%; height:100%; pointer-events:none; z-index:1`.

Per wire (`buildEdges`, with `W=216, H=58`):
- Start point = right-middle of source card: `x1 = a.x + 216`, `y1 = a.y + 29`.
- End point = left-middle of dest card: `x2 = b.x`, `y2 = b.y + 29`.
- **Path**: cubic Bézier with horizontal control handles at the midpoint x:
  `d = "M x1 y1 C mx y1, mx y2, x2 y2"` where `mx = (x1+x2)/2`.
- **Stroke**: `stroke:var(--amber); stroke-width:1.4; stroke-opacity:0.55; stroke-dasharray:"5 4"; fill:none`. → **dashed amber, thin, semi-transparent.**
- **End cap**: a filled `<circle cx=x2 cy=y2 r=3 fill:var(--amber)>` dot at the destination.
- No animation on the edges themselves.

Toggling connections on/off is done via the connections control (below); when off, the SVG is not rendered.

---

## Canvas Controls

A vertical control cluster pinned bottom-left: `position:absolute; left:16px; bottom:16px; display:flex; flex-direction:column; gap:8px`. Two grouped pieces stacked, with an optional popover.

**Zoom group** (a single rounded pill column):
`display:flex; flex-direction:column; background:var(--bg-elev); border:0.5px solid var(--border); border-radius:var(--r-3); box-shadow:var(--shadow-2); overflow:hidden`. Three `40px`-wide cells, each `width:40px; height:38px; center; color:var(--fg3); cursor:pointer`, hover `color:var(--amber)`, separated by `0.5px` `--border` divider spans:
1. Zoom in — `plus` glyph 17.
2. Zoom out — `minus` glyph 17.
3. Fit to view (`resetCanvas`, title "Fit to view") — `maximize` glyph 16.

**Connections toggle** (separate square button below the group):
`width:40px; height:40px; center; background:var(--bg-elev); border:0.5px solid <connBtnBorder>; border-radius:var(--r-3); box-shadow:var(--shadow-2); color:<connBtnColor>`; `workflow` glyph 18. When connections are shown, border/color go amber; otherwise `--border`/`--fg3`. Clicking toggles a small menu.

**Connections popover** (`connMenu`, opens to the right of the cluster):
`position:absolute; left:52px; bottom:0; width:280px; background:var(--bg-elev); border:0.5px solid var(--border); border-radius:var(--r-3); box-shadow:var(--shadow-2); padding:6px`. Two options, each a `padding:11px 12px; border-radius:var(--r-2)` row (hover `--bg-card`):
- **Variable references** (`workflow` glyph) — title `13.5px/500 --fg1`, sub `11.5px --fg3` "Show variable connections". Active row gets amber icon/emphasis (`connOnStyle`).
- **Hide connections** (`x` glyph) — sub "Clear the canvas of edges".

`--shadow-2 = 0 8px 24px rgba(0,0,0,0.40)` (the only real elevation shadow in the UI; used on floating controls, popovers, drawer, modals).

---

## Drawer (ResourceConfig rail)

Opens when a node is selected (`drawerRes`). Full-height right rail, **496px wide**, overlays the canvas:
`position:absolute; top:0; right:0; bottom:0; width:496px; z-index:20; background:var(--bg); border-left:0.5px solid var(--border); box-shadow:var(--shadow-2); display:flex; flex-direction:column; overflow:hidden`. Entry animation `sd-drawer` (slide-in from +34px x, `.26s`).

### Drawer header (`flex:none; gap:12px; padding:15px 16px; border-bottom:0.5px solid var(--border)`)
- Status dot `9×9 round` with the resource `dotStyle` (ok/warn/err).
- Glyph — `color:var(--amber)`, size 19 (kind icon).
- Title block (`flex:1; min-width:0; line-height:1.25`): name `16px/500 --fg1`; sub `.mono 11px --fg-muted` ellipsis (the same summary string, e.g. `node · :8080 · public`).
- **Changes pill** (only if `isDirty`): `inline-flex; gap:5px; border:0.5px solid var(--amber); background:var(--amber-soft); border-radius:var(--r-2); padding:2px 4px 2px 9px; font-size:11px; font-weight:500; color:var(--amber)`. Label text = e.g. `"2 changes"`, followed by an `×` discard control: `16×16; border-radius:3px`, hover `background:var(--amber)`, `x` glyph 12.
- When **not** dirty, the pill is replaced by a `.t-marker` kind label (`--fg-muted`) in the same slot.
- Close `×` — `x` glyph 18, `--fg3` in a 5px rounded-sm hit target, hover `--fg1`/`--bg-elev`.

### Drawer body (`flex:1; overflow:auto`) → `ResourceConfig`

**Tab bar** (`display:flex; align-items:stretch; gap:2px; border-bottom:0.5px solid var(--border); padding:0 4px; flex:none`). Three tabs, mono uppercase:
`CONFIGURATION`, `DEPLOYMENT`, `ENVIRONMENT`. Each: `padding:12px 13px; font-family:var(--font-mono); font-size:11px; letter-spacing:1.5px; text-transform:uppercase; border-bottom:1.5px solid <underline>; margin-bottom:-0.5px` (overlaps the bar's hairline).
- Active: `color:var(--fg1)`, `underline = var(--amber)`.
- Inactive: `color:var(--fg-muted)`, `underline = transparent`.
- **Dirty dot per tab** (`cfgDirty`/`depDirty`/`envDirty`): trailing `6×6 round background:var(--amber)`.

#### CONFIGURATION tab (`padding:18px; column gap:18px`)
- **Row 1** — 2-col grid `gap:16px`:
  - *Resource name*: label `12.5px --fg2` with a **reset** affordance on the right when `nameDirty` (`rotateccw` glyph 12 + "reset", `--amber` 11px). Value shown in a `nameBox` (clean box vs dirty box, see below).
  - *Build from*: label, then two toggle buttons in a `gap:7px` row, each `flex:1; center; gap:7px; padding:9px; border-radius:var(--r-2); font-size:13px; border:0.5px solid <b>`:
    - **Image** (`package` glyph 15) and **Git** (`gitBranch` glyph 15).
    - Active source: `border:var(--amber); color:var(--amber); background:var(--amber-soft)`. Inactive: `border:var(--border); color:var(--fg3); background:transparent`.
- **Repository/Image field** (label switches: `sourceFieldLabel` = "Repository" for git, "Image" for image): value box `background:var(--bg); border:0.5px solid var(--border); border-radius:var(--r-2); padding:9px 12px; font-size:13px; color:--fg1; gap:9px`. Leading source glyph (`gitBranch`/`package`/`cloud`, `--fg-muted` 15) + `.mono` source label (e.g. `github.com/acme/web`, `postgres:16-alpine`). For git, a right-aligned `.mono 11px --fg-muted` branch label (e.g. `branch: main`).
- **Row 3** — 2-col grid `gap:16px`:
  - *Port*: label + reset (when `portDirty`) **or** a `.mono 11px --fg-muted` protocol hint when clean (e.g. `TCP · HTTP`, `TCP · internal`). Value `.mono` in a port box; `portDisplay` = the port or `—`.
  - *Depends on*: `.mono` box, value = e.g. `db · cache` (color `--fg1` when present, `--fg-muted` when `—`).
- **Volumes** section: header row — `Volumes · N` (label `12.5px --fg2`, count `.mono --fg-muted`) and right caption `.mono 10.5px --fg-muted` "persistent storage". Then a col `gap:8px` of volume editors, plus an **Add volume** dashed row.
  - Volume editor card: `border:0.5px solid var(--border); border-radius:var(--r-2); background:var(--bg); padding:11px 12px; column gap:9px`:
    - Row: `drive` glyph (`--amber` 15) + name input (`.mono`, `bg:--bg-card`, `border:0.5px --border`, `12px`, focus `border:--amber`) + remove `×` (24×24, hover `--err`/`--err-soft`).
    - Row `gap:8px`: mount-path input (flex:1, leading `arrowRight` glyph 13, placeholder `/data`, `.mono` color `--fg2`) + size input (`width:106px`, right-aligned, trailing `.mono "GB"` suffix).
    - Usage: fill bar `height:5px; rounded 3px; track --bg-elev; fill --amber width:pct`; below it `used / size` (`.mono 10.5px --fg-muted`) on the left and `pct` (`.mono 10.5px --amber`) on the right.
  - **Add volume**: `center; gap:8px; padding:9px 12px; font-size:12.5px; color:--fg-muted; border:0.5px dashed var(--border); border-radius:var(--r-2)`, hover `color:--amber; border-color:--amber`; `plus` glyph 14.
- **Clean vs dirty field boxes** (reused for name/port/build/start/deploy values):
  - clean = `background:var(--bg); border:0.5px solid var(--border); border-radius:var(--r-2); padding:9px 12px; font-size:13px; color:--fg1`.
  - dirty = same but `background:var(--amber-soft); border:0.5px solid var(--amber)`.

#### DEPLOYMENT tab (`padding:18px; column gap:16px`)
- If the resource has a build (`hasBuild`): **Build command** and **Start command**, each a label(+reset when dirty) over a `.mono` clean/dirty box (`build` e.g. `npm ci && npm run build`, `start` e.g. `npm start`).
- Then per `deployRows` (managed resources): label + reset + `.mono` box, e.g. **Plan** → `standard-1`, **Image pull policy** → value.

#### ENVIRONMENT tab (`padding:16px 16px 20px; column gap:12px`)
- Header: `Environment · N variables` (label `--fg2`, count `.mono --fg-muted`) and two right-side pill buttons `.mono 11px --fg3; border:0.5px --border; rounded-sm; padding:5px 9px`, hover amber: **paste .env** (`copy` glyph 13) and **raw editor** (`filter` glyph 13).
- **Plain rows** group (col `gap:6px`) rendering `EnvRow` per variable (see Env Row).
- **Add variable**: dashed row identical in style to "Add volume" (`plus` glyph 14, hover amber).
- **Addon groups** — a boxed, dashed-amber region per addon binding group:
  `position:relative; border:0.5px dashed var(--amber); border-radius:var(--r-3); background:var(--amber-soft); padding:20px 11px 11px; margin-top:8px`.
  - Floating label chip straddling the top border (`position:absolute; top:-12px; left:11px; background:var(--bg); padding:2px 4px; rounded-sm`): a `.t-marker amber` `ADDON` badge (`plug` glyph 12, 9.5px), a `172px` **MiniSelect** for the addon, a `·` separator, and either a `130px` MiniSelect (db) or a `.mono 11px --fg3` static `db: <label>`.
  - The group's `EnvRow`s (col `gap:6px`), then an **Add binding** link right-aligned (`plus` glyph 13, `--amber` 12px).

### Drawer footer (`flex:none; border-top:0.5px solid var(--border); padding:11px 16px; space-between`)
- Left: **View logs** — `inline-flex; gap:7px; font-size:12.5px; color:--fg3; padding:6px 10px; rounded-sm`; `terminal` glyph 14; hover `color:--amber; background:--amber-soft`.
- Right: **Remove resource** — same layout; `trash` glyph 14; `color:--fg-muted`; hover `color:--err; background:--err-soft`.

---

## Env Row + MiniSelect

### EnvRow (~52px per row)
Container: `position:relative; border-radius:var(--r-2); background:<rowBg>; border:0.5px solid <border>; padding:9px 10px 9px 13px`, transition bg+border `.18s`.
- **Left accent bar**: `position:absolute; left:0; top:6px; bottom:6px; width:2px; border-radius:2px; background:<accent>`.
- **Dirty coloring** (when the var differs from baseline): `accent = var(--amber)`, `rowBg = rgba(249,115,22,0.06)`, `border = rgba(249,115,22,0.35)`. Clean: `accent = transparent`, `rowBg = transparent`, `border = var(--border)`.
- Inner row `align-items:flex-start; gap:8px`:
  1. **KEY input** — `width:130px; flex:none; .mono; bg:--bg; border:0.5px --border; rounded-sm; font-size:12px`, placeholder `KEY`, focus `border:--amber`.
  2. **Value column** `flex:1; column gap:6px` — the control depends on the variable's **source** (`from`):
     - `stack` (Plain text) → a plain value input (`.mono`, color `--fg2`, placeholder `value`).
     - `secret` → a `secret` MiniSelect + (optional) a `key` MiniSelect.
     - `addon` → a credential-field MiniSelect (`host`/`port`/`database`/`username`/`password`/`sslmode`/`ca_certificate`/`url`; cluster-owned fields tagged `CLUSTER`).
     - `resource` → a resource MiniSelect + (optional) an output MiniSelect.
     - `self` → a self-output MiniSelect.
  3. **Source ("from") MiniSelect** — `width:106px; flex:none`. Options with icons: **Plain text** (`code`), **Secret** (`lock`), **Addon** (`plug`), **Resource** (`workflow`), **Self** (`layers`).
  4. **Trailing action** (`width:24px; height:32px; center`): when modified → a **reset** (`rotateccw` 14, `--amber`, hover `--amber-soft`); otherwise → **remove** (`x` 14, `--fg-muted`, hover `--err`/`--err-soft`).

### MiniSelect (dropdown trigger, 32px tall)
A pill trigger (the menu itself renders via the shared root popover, below):
`display:flex; align-items:center; gap:7px; width:100%; padding:6px 8px 6px 10px; background:var(--bg); border:0.5px solid <border>; border-radius:var(--r-2); min-width:0`, hover `border-color:var(--fg3)`, open `border-color:var(--amber)`.
- Optional leading glyph size 13 (colored per `iconColor`).
- Label `flex:1; font-size:12.5px; color:<color>; ellipsis` (may be `.mono`).
- Trailing `chevron` glyph 13, `--fg-muted`, rotates on open (`chevTransform`), transition `.18s`.

**Shared select popover** (rendered at root, `position:fixed` at trigger x/y/w, `max-height:280px`): `background:var(--bg-elev); border:0.5px solid var(--border); border-radius:var(--r-3); box-shadow:var(--shadow-2); padding:5px`. Items `padding:8px 9px; rounded-sm; gap:9px`, hover `--bg-card`; optional leading icon, label `12.5px`, optional `.t-marker` right sub-tag (8.5px), and an amber `check` glyph on the active item.

---

## DeploymentsView

Scroll region `height:100%; overflow:auto; padding:26px 30px 40px`, inner `max-width:920px; margin:0 auto`. Entry anim `sd-mode` (from parent) and `sd-anim` on content.

- **Header row**: `.t-marker --fg3` `DEPLOY TIMELINE` on the left; **Deploy now** button on the right (`transparent bg; border:0.5px --border; padding:7px 13px; rounded-sm; 12.5px/500 --fg2`; `rocket` glyph 14; hover amber).
- **Timeline** (`position:relative; column`): a vertical rail `position:absolute; left:5.5px; top:14px; bottom:14px; width:0.5px; background:var(--border)`. Then one entry per deploy:
  - Entry `position:relative; padding-left:30px; margin-bottom:14px`.
  - **Timeline dot**: `position:absolute; left:0; top:13px; 12×12 round; border:2px solid var(--bg); background:<dotColor>; box-shadow:0 0 0 3px <dotHalo>` — green for live (`--ok`, halo `rgba(34,197,94,0.16)`), crimson for failed (`--err`, halo `rgba(217,34,62,0.16)`).
  - **Collapsed row** (clickable): `gap:11px; padding:9px 12px; border:0.5px solid <rowBorder>; border-radius:var(--r-3); background:var(--bg-card)`, hover `border-color:--fg3`. Contains: `#<seq>` (`13.5px/600 --fg1`), cause text (`13px --fg2`, e.g. "Manual deploy"/"Config change"), a **status chip** (`inline-flex; gap:6px; border:0.5px solid <chipBorder>; background:<chipBg>; color:<chipColor>; rounded-full; padding:2px 9px; mono 9.5px; letter-spacing:1px; uppercase` — `LIVE` = ok with a leading dot, `FAILED` = err), a flex subline (`12.5px`, err-colored for failures, ellipsis), a `.mono 11px --fg-muted` timestamp, and a `chevron` glyph 15 that rotates when open.
  - **Expanded card** (`n.open`): `margin-top:8px; border:0.5px solid <cardBorder>; border-radius:var(--r-3); background:var(--bg); padding:18px 18px 6px`:
    - **Step strip**: a horizontal sequence of stages, each = a `22×22` circular status node (`border:1.5px solid <ringColor>; background:<fillColor>; color:<iconColor>`, glyph 12 — e.g. `check` for done) + a `13px/500` label + a connecting `1.5px` bar (`barColor`). Done stages ring/bar = `--ok`; failed = `--err`.
    - **Failure banner** (failed only): `border:0.5px solid var(--err); background:var(--err-soft); rounded-md; padding:13px 15px`; `alert` glyph 15 + `Deploy failed` (`13.5px/600 --err`) over a `.mono 12px --fg2` fail message.
    - **RESOURCE OUTCOME** (`.t-marker --fg-muted`) then a list; each outcome row `gap:11px; padding:11px 2px; border-top:0.5px solid var(--border)`: `8×8` status dot, name (`13.5px/600 --fg1; min-width:90px`), status label (`12.5px/500`, ok green / err crimson), a right-aligned `.mono 11.5px --fg-muted` image string, and a small **logs** button (`terminal` glyph 12, `border:0.5px --border; rounded-sm; padding:4px 8px`, hover amber).

---

## MetricsView

Scroll region `padding:26px 30px 40px`, inner `max-width:1000px; margin:0 auto`.

- **Header**: `Stack metrics` (`18px/500 --fg1; letter-spacing:-0.01em`) + a **LIVE** status pill (same green pill pattern as READY, `.t-marker --ok 9.5px`), and a right-aligned `.mono 11px --fg-muted` `updated <time>`.
- **Two summary cards** (2-col grid `gap:14px; margin-bottom:18px`), each `border:0.5px solid var(--border); border-radius:var(--r-4); background:var(--bg-card); padding:18px`:
  - **STACK CPU** (`cpu` glyph 16, amber) / **STACK MEMORY** (`memory` glyph 16, amber), each labeled with a `.t-marker --fg3`.
  - Body: big number `font-size:30px; font-weight:500; color:--fg1; letter-spacing:-0.02em; line-height:1` + `.mono 11px --fg-muted` sub, next to a **mini bar sparkline** (`display:flex; align-items:flex-end; gap:3px; height:40px`; each bar `width:5px; height:<h>; background:<color>; border-radius:1px`). ~16 bars.
- **PER-RESOURCE** section (`.t-marker --fg-muted`) then an auto-fill grid `repeat(auto-fill, minmax(280px, 1fr)); gap:12px`. Each resource card `border:0.5px solid var(--border); border-radius:var(--r-3); background:var(--bg-card); padding:14px 15px`:
  - Header row: `8×8` status dot + kind glyph (15, `--fg3`) + name (`14px/500 --fg1; flex:1`) + `.t-marker` status (`READY`, 9px, colored).
  - Two metrics (**CPU**, **Memory**), each: a label row (`11.5px --fg3` label + `.mono 12px --fg1` value like `180m` / `312 MiB`) over a fill bar `height:5px; rounded 3px; track --bg-elev`. **CPU fill = `var(--amber)`; Memory fill = `var(--fg2)`** (deliberately different so the two bars read distinctly). Widths are the `cpuPct`/`memPct` strings.

---

## LogsView

`height:100%; display:flex; flex-direction:column; padding:24px 30px 26px`, inner `max-width:1100px; width:100%; margin:0 auto`, column, `min-height:0; flex:1`.

- **Header** (`flex:none; margin-bottom:16px; space-between`): left = `Stack logs` (`18px/500 --fg1`) + a **CONNECTED** green pill; right = two dropdown triggers (`gap:9px`):
  - **Resource filter** — `border:0.5px solid <border>; background:<bg>; color:<color>; rounded-sm; padding:7px 11px; 12.5px/500`; `layers` glyph 14 + label + `chevron` 13. When a resource is filtered, it goes amber (`border/bg/color` = amber set).
  - **Time range** — same shape; `clock` glyph 14 + label + chevron; neutral `--border`/`--fg2`, hover `--fg3`.
- **Terminal panel** (`flex:1; min-height:0; column; border:0.5px solid var(--border); border-radius:var(--r-3); overflow:hidden`) with a **distinct near-black background `#070a0f`** (darker than `--bg`):
  - **Toolbar** (`flex:none; gap:12px; padding:9px 12px; border-bottom:0.5px solid var(--border); background:var(--bg-card)`): a search input (`flex:1; max-width:340px`, leading `search` glyph 14, `.mono 12px`, placeholder `Filter log lines…`, focus `border:--amber`), a spacer, and a `.mono 11px --fg-muted` match counter with a `filter` glyph 13.
  - **Log stream** (`flex:1; min-height:0; overflow:auto; padding:8px 0`): one row per line, `display:flex; align-items:baseline; gap:14px; padding:1.5px 16px; font-family:var(--font-mono); font-size:12px; line-height:1.7`, hover `background:rgba(255,255,255,0.025)`:
    - Line number — `flex:none; width:30px; text-align:right; color:#3a4453; user-select:none`.
    - Resource tag `[res]` — `flex:none; color:<tagColor>` (err = `--err`, warn = `--warn`, else `--fg-muted`).
    - Message — `flex:1; color:<color>; white-space:pre-wrap; word-break:break-all` (err = `--err`, warn = `#d9c36a` muted-gold, else `--fg3`).
  - Empty state: centered `13px --fg-muted` "No log lines match this filter."

---

## Token Values

Exact values from `_ds/.../colors_and_type.css` (dark theme `:root` is default; light overrides noted where relevant). Use the **live-code** equivalents from the translation table above.

**Brand (the only color):**
- `--amber: #f97316` (→ `--brand`), `--amber-hover: #ea6a0e`, `--amber-press: #c25809`
- `--amber-soft: rgba(249,115,22,0.12)` (→ `--brand-bg`)

**Surfaces (dark):** `--bg:#0a0e14` · `--bg-card:#11161e` · `--bg-elev:#161c26` · `--bg-inverse:#f5f0e6`
(light theme: `--bg:#fdfcf9` · `--bg-card:#ffffff` · `--bg-elev:#ffffff`)

**Text (dark):** `--fg1:#ffffff` (→`--fg`) · `--fg2:#cbd5e1` (→`--fg-2`) · `--fg3:#94a3b8` (→`--muted-foreground`) · `--fg-muted:#64748b`
(light: `--fg1:#0d0d0d` · `--fg2:#2a2a2a` · `--fg3:#6b6b6b` · `--fg-muted:#9c9c9c`)

**Borders:** `--border:#1f2937` (dark) / `#e8e4d6` (light) · `--border-soft:#161c26` · `--hairline:0.5px solid var(--border)` · `--border-amber:#f97316`

**State (used sparingly — status only):**
- `--ok:#22c55e` · `--ok-soft:rgba(34,197,94,0.14)`
- `--warn:#eab308` · `--warn-soft:rgba(234,179,8,0.16)`
- `--err:#d9223e` · `--err-soft:rgba(217,34,62,0.16)`
- `--info:#3b82f6` · `--info-soft:rgba(59,130,246,0.16)`

**Radii (tight — reads as infrastructure):** `--r-0:0` · `--r-1:2px` · `--r-2:4px` (inputs/buttons → `rounded-sm`) · `--r-3:6px` (→`rounded-md`) · `--r-4:8px` (cards → `rounded-lg`) · `--r-pill:999px`

**Spacing scale:** `--s-1:4` · `--s-2:8` · `--s-3:12` · `--s-4:16` · `--s-5:24` · `--s-6:32` · `--s-7:48` · `--s-8:64` · `--s-9:96` · `--s-10:128` (px). (Editor uses many off-scale values directly, e.g. 7/9/11/13/14/18/20/26/28px paddings — transcribe the literals given per component above.)

**Shadows (rare — flat is the point):**
- `--shadow-0:none` · `--shadow-1:0 1px 0 0 var(--border)`
- `--shadow-2:0 8px 24px rgba(0,0,0,0.40)` — floating controls, popovers, drawer, modals
- `--shadow-amber:0 0 0 3px rgba(249,115,22,0.20)` — selected/focus halo

**Type families:** `--font-sans: 'Geist', -apple-system, …` · `--font-mono: 'Geist Mono', 'JetBrains Mono', monospace`

**Type scale:** `--t-display:80` · `--t-h1:56` · `--t-h2:32` · `--t-h3:22` · `--t-body:18` · `--t-body-sm:15` · `--t-caption:13` · `--t-mono:11` (px). The `.t-marker`/`.marker` class = mono, 11px, `font-weight:500`, `letter-spacing:1.5px` (`--tr-mono`), uppercase, `color:--fg3` (add `.amber` variant for brand color). `.mono`/`code` = mono with `tabular-nums` + `ss01`.

**Tracking:** `--tr-tight:-0.03em` · `--tr-normal:-0.02em` · `--tr-loose:0.04em` (wordmark) · `--tr-mono:1.5px`

**Motion:** `--ease: cubic-bezier(0.2,0.6,0.2,1)` · `--dur-1:120ms` (hover) · `--dur-2:240ms` · `--dur-3:400ms`. Named keyframes used here: `sd-anim`/`sd-fade` (content fade-up), `sd-scrim`, `sd-pop` (modal), `sd-drawer` (drawer slide from +34px), `sd-mode` (mode-switch fade-up).

**Misc:** `::selection { background:var(--amber-soft); color:var(--amber) }`. Scrollbars: 10px, thumb `--border` with 3px transparent inset. Amber-filled buttons/tiles use text color `#1a0e05`.
