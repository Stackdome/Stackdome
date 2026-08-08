# Plan — build the create-a-stack flow in code

## Status — BUILT, August 2026

All ten steps are done. The flow lives at `/stacks/new`; the canvas on an
unsaved draft moved to `/stacks/draft`; the old dialog and its seven files are
deleted.

**One premise in this plan was wrong and was corrected in the build:** a draft
has **no id** until it is saved (`isNewStack = !id` in the editor), so
`Create stack` cannot open `/stacks/:id`. Jaseem chose to give the unsaved
canvas its own path rather than write a half-made stack to the database. Whether
a draft should ever be persisted is still open.

Two more corrections the build made, both rules the code had been breaking:

| | |
|---|---|
| `eyebrow-muted` was set in **mono** | §6 — mono is never for a human label. Now sans. Also fixes Panel and the node ledger |
| The block registry shipped `SERVICES` / `DATABASES` | §11 — no uppercase. Now sentence case |

---

**Paste this whole file into a fresh chat to start.**

Repo: `/Users/jaseem/Projects/Stackdome` · branch `graphite-pass-2`
Working with **Jaseem** — a designer with basic coding knowledge. Short lines,
bullets, tables. Show, don't describe. Measure, don't eyeball.

---

## Read these first, in this order

| # | Path | Why |
|---|---|---|
| 1 | `DESIGN-PRODUCT.md` | **The only authority.** If code and it disagree, the code is wrong |
| 2 | `frontend/CONTEXT.md` | The SPA's glossary — canonical nouns |
| 3 | `CLAUDE.md` → *Frontend UI work starts in Storybook* | How work is built and judged here |
| 4 | `docs/design/prototypes/create-stack/README.md` | What the design settles, and why |
| 5 | The Figma section `create a stack` | The visual reference |

**Figma:** file `2IcCJOgsROpgajjXlay1h9`, section **`create a stack`** — five
1440×900 Light frames, `00`–`04`.
**Prototype:** https://claude.ai/code/artifact/6d649912-8686-4915-b5e9-501eeb821e40
**Audit it came from:** https://claude.ai/code/artifact/9b332f32-91a3-4b06-aa7f-fe1ec2f61287

**Do not re-derive the design.** It went through the audit, five rounds of
prototype review, and a full Figma pass. Your job is to build it.

---

## Settle this before writing anything

**`/stacks/new` is already taken.** It routes to `CanvasEditorPage`
(`src/App.tsx:55`). The redesigned chooser wants the same path.

The prototype's own answer resolves it: **`Create stack` makes a draft**, and
the canvas is where that draft opens. A draft has an id, so:

| Route | Today | After |
|---|---|---|
| `/stacks/new` | Canvas editor | **The chooser page** (this work) |
| `/stacks/:id` | Canvas editor | Unchanged — where a draft opens |
| `/stacks/create` | Redirects to `/stacks/new` | Unchanged |

So the blank-canvas path stops being "navigate to the canvas" and becomes
"create an empty draft, then open it" — the same ending as the other four. That
is exactly what the prototype specifies: *"Picking one never traps you… the flow
ends on the canvas."*

**Confirm this with Jaseem before building.** Everything else depends on it.

---

## What is already built — do not redo it

| Thing | Where | State |
|---|---|---|
| **`StartingPointTabs`** | `pages/stacks/components/wizard/starting-point-tabs.tsx` | **Done.** 7 story tests green |
| **Journey header** — back arrow, no trail | `components/sheet-header.tsx` | **Done.** 7 story tests green |
| `useJourney(origin)` | `hooks/use-journey.ts` | **Done** — one line makes a page a journey |
| `journeyOrigin` in context | `contexts/breadcrumb-context.tsx` | **Done** |
| `DESIGN-PRODUCT.md` §12a — journeys vs nested | doc | **Amended** |

### Primitives that already exist — use them, do not rebuild

**This is the rule that gets broken most often here.** Three of the five
components we drew in Figma turned out to need **zero code**:

| Figma component | Use this | Note |
|---|---|---|
| Switch | `ui/switch.tsx` | 36×20, ink when checked. Already correct |
| Segmented — label | `ui/segmented-control.tsx` | Already takes labels via `showLabel`. The "labelled sibling" **is** this primitive |
| `Start from` eyebrow | `branded/eyebrow-label.tsx` | `tone="muted"` |
| Inline alert | `branded/alert-banner.tsx` | **Needs a `tone` prop** — hardcoded danger-only today |
| Picker row | — | **New.** The only genuinely new primitive left |

---

## What this replaces

The current flow is a **Dialog** opened from the stacks list:

| File | Fate |
|---|---|
| `wizard/stack-create-wizard.tsx` (Dialog shell) | **Replaced by a page** |
| `wizard/wizard-chooser.tsx` | **Replaced** by `starting-point-tabs.tsx` (done) |
| `wizard/git-source-panel.tsx` | Rebuilt as a tab body |
| `wizard/templates-browser-panel.tsx` | Rebuilt as a tab body |
| `wizard/docker-compose-import-panel.tsx` | Rebuilt as a tab body |
| `wizard/block-composer.tsx`, `block-picker.tsx` | Rebuilt as a tab body |
| `pages/stacks/components/list/index.tsx:434` | `+ New stack` **navigates** instead of opening a dialog |

Their hooks — `use-template-import`, `use-docker-compose-import` — are the good
part and should survive. Read them before rewriting anything.

---

## Build order

Storybook-first at every step (`pnpm --prefix frontend storybook`). Each
component gets a colocated `.stories.tsx` covering the happy path **plus the
states that break in production**: empty, error, long text, overflow.

| # | Step | Notes |
|---|---|---|
| 1 | **`PickerRow`** | Two sizes (56 / 40), states default/hover/selected, trailing none/tick/count/plus/remove. The 40px rung is the "In this stack" panel |
| 2 | **`AlertBanner` `tone` prop** | `blocking` (warn) / `info`. Keep danger as-is for existing callers |
| 3 | **The page shell** at `/stacks/new` | Sheet header + `useJourney("/stacks")` + tab strip + body + footer |
| 4 | **Tab body: repository** | Segmented (provider / public URL), 32px search, repo rows |
| 5 | **Tab body: ready-made app** | List + the template detail side panel |
| 6 | **Tab body: building blocks** | Category headings, count badge, "In this stack" at 40px with remove |
| 7 | **Tab body: compose file** | Paste area, parse preview, chips |
| 8 | **Tab body: blank canvas** | Explainer rows with brand-logo chips |
| 9 | **Wire `Create stack`** | Draft creation → navigate to `/stacks/:id` |
| 10 | **Retire the dialog** | Delete the old wizard, update its tests |

---

## Rules this flow breaks most easily

| Rule | |
|---|---|
| **A brand logo is never tinted** | Lucide glyphs take `ink/fg-2`; a Postgres or Redis logo keeps its own colour. This was got wrong once already on the board and had to be undone across 9 logos |
| **No orange** | §7. Orange belongs to the website and to thresholds. The old wizard is full of it — that is the thing being fixed |
| **No uppercase** | §11 — including the block registry's own `SERVICES` / `DATABASES` |
| **Selection is `--ring`** | The blue. One colour, and it already matches the board's `accent/primary` |
| **32px controls, not 40** | Search and filters are working controls (§11). There are no true form fields in this scope |
| **One filled button per screen** | `Create stack` in the footer. Everything else is ghost or outline |
| **Update the primitive, don't clone it** | And **ask first**. See the table above — most of the work is already done for you |

---

## Not designed yet — two real gaps

| Gap | Risk | Suggestion |
|---|---|---|
| **Dark mode** | **Low.** Every value is a token and both themes are already defined. But it is *unverified* — no dark frame was ever drawn | Build in light, then check every story in Storybook's dark theme. Fix tokens, never call sites |
| **Empty org — no provider connected** | **Real.** A brand-new org has no Git provider, so the repository tab's first state is an empty state, and it was never drawn | The prototype implements it (`prototype.source.html`, search `no provider`). Follow it, and use `branded/empty-state.tsx` |

Two smaller ones, both behaviour rather than pixels:

- **`Clear` is disabled until there is something to clear**; `Paste an example`
  arguably becomes `Replace` once content exists. Decided in code, not drawn.
- **Per-tab no-results states** — each search needs one. `SearchGlyph` in
  `branded/empty-state.tsx` is the existing pattern.

### Out of scope — say no to these

- **The canvas** — nodes, wires, the node inspector, deploy. Separate project.
  This flow **ends** by creating a draft and opening it.
- The **inline alert** has no home here. It belongs to the canvas.
- Motion. §14 is still "best practices for now".

---

## Verification

| | |
|---|---|
| Types | `npx tsc -b --noEmit` clean |
| Stories are tests | `pnpm --prefix frontend test:run` — jsdom suite **then** every story in headless Chromium. Both run in CI |
| Baseline today | **1975 tests, 278 files, all passing.** Do not regress it |
| The real app | `pnpm --prefix frontend dev:mock` (:5273), and `dev:mock:empty` (:5274) for the first-run states |
| Against the board | Screenshot each screen at 1440 and compare to the Figma frame |

**Judge it in the running app, not in isolation.** §2: *"A component judged in
isolation has no basis to be judged."*

---

## Effort

| Step | |
|---|---|
| 1–2 — `PickerRow`, alert tone | ~half a day |
| 3 — page shell + routing change | ~half a day |
| 4–8 — the five tab bodies | ~2 days; repository and blocks are the big ones |
| 9–10 — wire up, retire the dialog | ~half a day |

**Roughly 3–4 days.** The riskiest step is 9 — draft creation touches the API
and the edit session, and it is the one part with no prototype behaviour to copy.
