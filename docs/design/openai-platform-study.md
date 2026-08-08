# OpenAI Platform — design study

> ## This document is EVIDENCE, not instruction
>
> **`DESIGN-PRODUCT.md` is the only authority.** Nothing here is a rule. These
> are measured findings; where they were considered and *rejected*, the
> rejection is recorded below and this file does not get to argue back.
>
> **Points overruled by `DESIGN-PRODUCT.md` — do not act on these:**
>
> | This file says | The rule says | Why |
> |---|---|---|
> | §1, §9 — "Keep **pill** as the default; add `flat`" | **`flat` is the default** | This recommendation is what shipped 166 of 219 buttons as pills against a rule of one per screen |
> | §7, §9 — "Adopt **uppercase** column headers" | **No uppercase, anywhere** | Caps strip the word-shape the eye reads by. Size and the muted tier already say "this is a label" |
> | §9 — "consider whether inputs should be taller than buttons" | **Inputs match their buttons** | Considered and closed. A 40px field pairs with a 40px button |
> | §3 — "our frame is lighter than the console's 4.7%" | Ours measured **5.0%** | The comparison was inverted. Grounds are being retuned into the 2–4% band |
>
> Everything else here stands as evidence and was folded into the rules.

**Source:** 417 screens of the OpenAI Platform console (web, Apr 2026), captured
via Mobbin. Every screen was reviewed; the numbers below were **measured off the
pixels**, not estimated.

**Why this doc exists.** `DESIGN-PRODUCT.md` names OpenAI as the reference for
the product UI. That reference was previously drawn from ChatGPT — a *consumer*
surface. This set is the *console*: projects, keys, logs, usage, billing,
permissions. It is the closer analogue to Stackdome, and on several points it
**contradicts** what we inferred from ChatGPT.

Findings that should change our rules are marked **▲**. Nothing here is a rule
until it lands in `DESIGN-PRODUCT.md`.

---

## 0. How the numbers were derived

The captures are 1920 × 1205 device px. Calibration:

| Probe | Device px | Implied scale |
|---|---|---|
| 1px hairline border | 2 | ~1.11 |
| Default button height | 40 | 36 CSS |
| Page title ink (cap→descender) | 23 | ~20px type |
| Body line ink | 16–17 | ~16px type |

Everything is consistent with a **1728 × 1084 CSS viewport captured at ~1.11×**
(a 16" MacBook Pro at default scaling). All CSS values below are device px ÷ 1.11.

Independently corroborated: `DESIGN-PRODUCT.md` §8 already records ChatGPT's
control step as **36px on a 16px type base**, measured separately in-browser.
The console lands on the same step.

---

## 1. ▲ Shape is a semantic, not a style

**The single most important finding.** OpenAI runs **two button shapes
simultaneously** and never mixes them inside one group.

| Shape | Radius | Height | What it means | Where it appears |
|---|---|---|---|---|
| **Pill** | height ÷ 2 | 37–56px | *This commits / ships / finishes* | Auth `Continue`, hero `+ Create` on an empty state, `Publish`, `Deploy` |
| **Flat** | **6px** | **36px** | *This is the everyday working control* | Toolbars, dialog footers, row actions, form `Save`, `Cancel` |

Measured instances:

| Screen | Button | Device | CSS | Radius | Shape |
|---|---|---|---|---|---|
| #3 | auth `Continue` | 432 × 66 | 389 × 59 | 28 | **pill** |
| #312 | hero `+ Create` | 130 × 51 | 117 × 46 | 21 | **pill** |
| #109 | `Publish` | 99 × 41 | 89 × 37 | 20 | **pill** |
| #225 | `+ Create new secret key` | 233 × 40 | 210 × 36 | 8 | flat |
| #276 | `+ Create` (top bar) | 113 × 40 | 102 × 36 | 7 | flat |
| #312 | dialog `Create` | 84 × 40 | 76 × 36 | 7 | flat |
| #59 | dialog `Rename` | 97 × 40 | 87 × 36 | 7 | flat |
| #148 | toolbar `Save` | 69 × 40 | 62 × 36 | 5 | flat |
| #87 | dense toolbar `Save` | 64 × 35 | 58 × 32 | 4 | flat |

**Reading:** the pill is *rationed*. A screen gets at most one, and it marks the
end of a flow. Everything you touch while working is a 6px rectangle.

**Conflict with our rules.** `DESIGN-PRODUCT.md` §2 and §6 currently say a
button gets `rounded-full` because "pill is what small things get". That is a
**size** rule. OpenAI's is a **meaning** rule. Adopting theirs means a `shape`
axis on the Button primitive, not a replacement of the pill.

**Corollary:** the radius tightens with the control — 6px at 36px, ~4px at 32px.
Radius tracks the control's own size, the same principle our §2 ladder already
encodes for panels and sheets.

---

## 2. Geometry

| Element | Device px | CSS px | Notes |
|---|---|---|---|
| Default control height | 40 | **36** | Buttons, inputs, selects |
| Dense control height | 35 | **32** | Toolbar row, in-table actions |
| Input field height | 46 | **41** | Taller than its button — inputs hold content |
| Sidebar width | 277 | **250** | |
| Sidebar nav row pitch | 45 | **40** | |
| Top bar height | 70 | **63** | Breadcrumb bar |
| Hairline | 2 | **1** | |
| Flat radius | 7 | **6** | |

**Scale translation.** Their step is 36px on a 16px base. Ours is 32px on a 13px
base (§3, §5). The *proportions* transfer; the *numbers* do not. A 6px radius on
a 36px control is 5.3px on our 32px control — our existing `rounded-sm` (6px)
rung is the honest landing spot, and no new token is needed.

---

## 3. Colour — recorded for reference, not for adoption

We are not changing our palette. Logged so the *structure* can be compared.

| Role | Value | Structural note |
|---|---|---|
| Sheet | `#FFFFFF` | |
| Frame / sidebar | `#F3F3F3` | ~4.7% ink |
| Hover / selected | `#ECECEC` | ~7% ink |
| Border | `#E0E0E0` | |
| Ink / primary fill | `#181818` | |
| Success | `#49B880` | Toasts only |
| Destructive | `#E12E2A` | |

**The structural lesson, which does transfer:** there is **no brand colour
anywhere in the product**. Black is the only action colour. Green and red appear
*only* to report state — never as decoration, never as a fill for a non-state
control. This is the same rubric as our §4 ("if a colour doesn't report
something, it's a bug"), independently arrived at.

Frame grey at 4.7% ink is slightly heavier than our §1 "2–4%" band. Worth a look
when we next audit surfaces, but it is a palette question and out of scope here.

---

## 4. ▲ Destructive actions escalate with the damage

Three distinct levels, chosen by blast radius — not one `destructive` variant
applied everywhere.

| Level | Gate | Example |
|---|---|---|
| 1 — Confirm | Red button in a dialog, enabled immediately | Delete app (#247), Leave without saving (#94) |
| 2 — Acknowledge | **Checkbox** must be ticked before the red button enables | Delete workflow (#139 → #140) |
| 3 — Retype | User must **type the resource name** before the red button enables | Archive project (#367 → #368) |

The red button sits at the **far right**, after `Cancel`. It is red *fill*, not a
red outline or red text. Until its gate is satisfied it renders disabled — the
control is present but visibly inert, so the cost is legible before you commit.

We currently have one `destructive` variant and no gating pattern. Level 2 and 3
are cheap and would suit stack deletion, cluster removal and addon destroy.

---

## 5. ▲ Toolbars: many ghosts, exactly one fill

Every content toolbar follows the same shape (#87, #281, #283):

```
↑ Upload   ↗ Export   ⊞ Columns   ▶ Generate output   ⚑ Grade        [ Save ]
└───────────── icon + label, no border, no fill ──────────────┘      └ filled ┘
```

Rules read off the set:

| | |
|---|---|
| Secondary actions | icon + label, **no border, no fill** |
| Primary | one filled dark button, far right |
| Disabled secondary | label greys out, icon greys with it, no box appears |
| Gap | consistent throughout; no dividers between groups |

This is stricter than what we do. Our toolbars use `outline` and `secondary`
buttons, which draw a box around things that are not the point of the screen.

---

## 6. Empty states carry the page

Every empty surface — Assistants, Batches, Evaluation, Storage, Logs, Webhooks,
Vector stores, Fine-tuning — uses the identical five-part block, centred:

1. Line-art glyph (~24px, `fg-muted`)
2. Headline, one line, sentence case
3. One sentence of explanation
4. The action — **as a pill**
5. Occasionally a `Learn more ↗` beside it

The explanation lives **here**, not in the page header. That is exactly our §8
rule ("you need the explanation when there is nothing yet") — independently
confirmed.

Two flavours worth copying:
- **Split empty state** (#193, #270): the list pane says "No batches found" *and*
  the detail pane says "Select a batch to view details." Both halves speak.
- **Blocked, not empty** (#181): "Payment method needed" with `Finish account
  setup →`. A blocked state is not an empty state and does not get the same block.

---

## 7. Lists and tables

| | |
|---|---|
| Table headers | **UPPERCASE**, small, `fg-muted`, no fill, 1px rule under |
| Row separation | 1px rule. No card, no shadow, no per-row box |
| Row actions | icons at the right, revealed on hover |
| Dataset grids | spreadsheet ruling **both** directions — but only for editable cells |
| Selected row | flat grey fill, small radius, no border |

Confirms our §7 ("hairlines, not cards"). The one addition: **uppercase column
headers**, which we do not do. It reads as data-table furniture and separates
the header from content without a fill.

---

## 8. Patterns we do not have

| Pattern | Where | Worth stealing? |
|---|---|---|
| **Segmented control** | Simple\|Advanced, ChatKit\|Agents SDK, All\|Restricted\|Read only | Yes — we fake this with tabs or radios |
| **Wizard stepper** | App submission (#237), New evaluation (#287) | Yes — multi-step flows currently have no chrome |
| **Removable token chips** | `product_type ×` in prompt variables | Yes — our multi-select is close |
| **Inline "Draft / Live" status by the title** | Agent Builder | Maybe — we put status in the header's right slot |
| **`Draft saved 4 seconds ago`** in the bar | App wizard | Yes — quiet autosave receipt |
| **Split detail layout** | Assistants, Batches, Storage, Fine-tuning | Yes — list left, detail right, no navigation |

---

## 9. What contradicts our current rules

| Our rule | What the console does | Recommendation |
|---|---|---|
| §2/§6 — buttons are always `rounded-full` | Pill is **rationed** to the one committing action; everything else is a 6px rect | **Add a `shape` axis.** Keep pill as the default; add `flat` |
| §5 — heights 28/32/40 | 32/36/41, and **inputs are taller than buttons** | Keep our ladder; consider whether our 40px input/40px button pairing should differ |
| §6 — one `destructive` | Three escalation levels gated by checkbox / retype | **Add gating**, not more variants |
| Toolbars use `outline`/`secondary` | Toolbars are **ghost only**, one fill | **Tighten.** Ghost for all secondary toolbar actions |
| — | Uppercase table column headers | Adopt for data tables |

---

## 10. What it confirms

- No brand colour in the product; black is the only action colour (§4)
- Filled is rare — one per screen (§6)
- Hairlines, not cards (§7)
- Explanation belongs in the empty state, not the header (§8)
- Grey means input or reference — code blocks, search fields, prompt wells (§1)
- Status is said once (§4)

---

## Appendix — screen index

Captures live outside the repo. Reference numbers used above:

| Range | Area |
|---|---|
| 0–12 | Marketing site, auth, sign-up |
| 13–18 | Onboarding — org creation, first key, credits |
| 19–99 | Chat prompts / Playground |
| 100–141 | Agent Builder |
| 142–164 | Audio — realtime, TTS |
| 165–192 | Images, Videos |
| 193–209 | Assistants (deprecated banner) |
| 210–223 | Usage & spend |
| 224–234 | API keys |
| 235–251 | ChatGPT Apps, Logs |
| 252–275 | Responses, Storage, Batches |
| 276–298 | Evaluation, datasets, graders |
| 299–314 | Fine-tuning, projects |
| 315–323 | Docs site, account menu, error state |
| 324–401 | Settings — org, members, roles, billing, limits, data controls |
| 402–416 | Auth edge cases — reset, invite, wrong password |
