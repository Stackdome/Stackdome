# Plan — Create-new-stack on the Figma board

## Context

The create-new-stack flow has been audited, redesigned and agreed as a working
prototype. Figma is the remaining step before code.

| | |
|---|---|
| **Prototype (the reference)** | https://claude.ai/code/artifact/6d649912-8686-4915-b5e9-501eeb821e40 |
| **Source, in-repo** | `docs/design/prototypes/create-stack/` |
| **Audit it came from** | https://claude.ai/code/artifact/9b332f32-91a3-4b06-aa7f-fe1ec2f61287 |
| **File** | `2IcCJOgsROpgajjXlay1h9` — *Stackdome — Shape + Hierarchy Pass* |
| **Rules** | `DESIGN-PRODUCT.md` — the only authority |

`DESIGN-PRODUCT.md` §2 says Figma lays the **foundation** — *"usually not
exhaustive, just enough to fix the direction, then expanded in code"*. The
direction is already fixed by the prototype, so the board's job is narrow: **make
the new primitives real, and lay down the starting-point screens.**

### Scope: the canvas is a separate project

Everything after `Create stack` — the draft canvas, its nodes and wires, the node
inspector, deploy — **is out of scope and gets its own piece of work.** The
prototype still demonstrates it; the board does not model it.

The illustration node inside `glyph / architecture` (`201:1234`) is a **visual
device for empty states only**. It is not a UI component and must not be used to
design one.

---

## What is already on the board

Read from the file. Reuse all of it as-is.

| Component | Node | Shape |
|---|---|---|
| `Button` | `3:14` | Shape (pill/flat) × Tone (primary/secondary/ghost) × State ×5 — 30 variants |
| `icon button` | `50:1822` | Same matrix at 32×32 |
| `Field` | `21:191` | Shape × State ×4 — 300 × **32** |
| `Select` | `21:204` | Shape × State ×5 — 104 × 32 |
| `View toggle` | `27:358` | Icon-only segmented, 64 × 32 |
| `Sidebar` | `50:2246` / `50:2473` | 240 expanded / 56 collapsed |
| `Nav item` | `16:12` | 12 variants |
| `Account block` | `50:1389` | 8 variants |
| `Icons — lucide` | `17:4` | 20 icons |
| `Icons — brand` | `132:950` | postgres · redis · elasticsearch · mysql · mongo |
| Empty-state patterns | `199:2386`, `199:2399` | **Dark only** |

Variable collections `Colour` and `Elevation` stay the only source of values.

---

## Decisions — settled

| # | Need | Decision |
|---|---|---|
| 1 | 40px form field | **No.** Stay on the existing 32px `Field` and judge it on screen first |
| 2 | Segmented control with labels | **Add a sibling.** Leave `View toggle` alone |
| 3 | Starting-point card | **New component** |
| 4 | Picker row | **New component** |
| 5 | Switch | **New component** |
| 6 | Canvas node | **Dropped** — canvas is a separate project |
| 7 | Inline alert | **New component** |

### Why 32px turns out to be the right call anyway

§8 reserves 40px for *"form fields and their primary button"*. With the canvas
deferred, **there are no true form fields left in scope** — the only inputs on
these screens are a search box and a URL paste, and §11 classes search and
filters as `flat` working controls at 32px. So 32 is not a compromise here; it is
the correct rung. The 40px question returns with the canvas project, where stack
name and project genuinely are form fields.

### Two primitives with no home on this board

**Switch** and **inline alert** were both used on the canvas in the prototype.
They are approved and worth building now — forms and warnings will need them —
but expect them to sit on the components page **without appearing in any screen
frame** until the canvas project runs. Nothing is wrong if they look unused.

---

## Phase 1 — components

Each gets the states the board already conventions — `default / hover / focus /
disabled` where interactive — plus dark. Variants named `Property=Value`,
labels sentence case.

| Component | Size | Properties |
|---|---|---|
| **Starting-point card** | ≈189 × 124 | `State = default / hover / active` · `Style = solid / dashed` |
| **Picker row** | **`Size = 56 / 40`** | `State = default / hover / selected` · `Trailing = none / tick / count / plus` |
| **Segmented — label** | h 32, hugs | `Segments = 2` · `Selected = 1 / 2` · `State` |
| **Switch** | 36 × 20 | `Checked = true / false` · `State` |
| **Inline alert** | fills | `Tone = blocking / info` · `Action = none / button` |

The **40px picker row** is the "In this stack" item — same object, denser. One
component, two sizes; not a second component.

---

## Phase 2 — screens, Light, 1440 × 900

A new section `create a stack`, beside the existing `app shell` section. Each
frame is the app shell with the sheet swapped.

| # | Frame | What it proves |
|---|---|---|
| 00 | `new stack — repository` | The landing state. Cards on top, one active, work below. Segmented tabs, picker rows, 32px search |
| 01 | `new stack — ready-made app` | The side panel as **template detail** — badge, category, version, blurb, Website/Docs, then "Includes · n" |
| 02 | `new stack — blocks` | Three categories, a `×3` count badge, "In this stack" at 40px with per-instance remove |
| 03 | `new stack — compose file` | Paste area and the parse preview — chips and the found-summary line |
| 04 | `new stack — blank canvas` | The dashed card active, and the explainer rows in place of a list |

All five share one layout, so 00 carries the real work and 01–04 are its
variations. Build 00 properly and the rest are fast.

## Phase 3 — dark and empty

| # | Frame | Why |
|---|---|---|
| 05 | `new stack — repository · Dark` | Proves the card, row and segmented tokens flip |
| 06 | `empty org — no provider · Light` | The two existing empty states are **Dark only**. Reuses their pattern and closes the gap |

---

## Values the board must match

From the prototype; all of them are already `DESIGN-PRODUCT.md` rules.

| | |
|---|---|
| Card grid | 5 columns, 12px gap, 20px below, then a `border-subtle` divider |
| Starting-point card | radius 12, hairline. Active = **ink border + ink-filled 30px chip**. Blank = dashed edge, dashed chip, and the chip still fills when active |
| Picker row | 56px, radius 8, **no rule between rows**. Hover 4% · selected 6% · selected-hover 9% |
| Side panel | 300px, `border-subtle` on the left, 24px inset. **Only appears once something is selected** |
| Category headings | `text-label` 11/16, `fg-muted`, **sentence case** — never the registry's `SERVICES` |
| Buttons | One filled per screen. The empty-state action is a **pill** (§9) |
| Colour | **No orange.** §7 — it belongs to the website and to thresholds |

---

## Verification

Figma has no test runner, so this is a read-back, not a feeling.

1. **Every fill, stroke and shadow is a bound variable.** No raw hex — spot-check
   each new component with `get_variable_defs`.
2. **No orange** anywhere in the section.
3. **No uppercase** in any label, including category headings.
4. **Radius matches height** — 32px controls at 8, cards at 12.
5. **Screenshot each frame** and compare against the prototype at 1440.
6. Re-read §8 and §9 against the built buttons and the 32px fields — the point of
   holding 32 is to judge it, so look at it deliberately.

---

## Effort

| Phase | |
|---|---|
| Phase 1 — five components | ~half a day, mostly variant matrices |
| Phase 2 — five Light frames | ~half a day; 00 is most of it |
| Phase 3 — dark + empty | ~1–2 hours |

## Then

Code, Storybook-first per `CLAUDE.md`, following `frontend/CONTEXT.md`. The
prototype stays the behavioural reference; the board stays the visual one. Where
they disagree, `DESIGN-PRODUCT.md` wins and both get corrected.

## Not doing

- **The canvas** — nodes, wires, inspector, deploy, alerts in place. Separate project.
- Rebuilding anything already on the board.
- The three rejected chooser layouts (rail, gallery, tabs).
- A 40px `Field` variant. Revisit with the canvas project.
- Motion. §14 is still "best practices for now"; the prototype's 150ms
  directional transition is a proposal, not a rule to enshrine.
