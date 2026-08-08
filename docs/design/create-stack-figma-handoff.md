# Handoff — build the create-new-stack flow on the Figma board

**Paste this whole file into a fresh chat to start.**

Repo: `/Users/jaseem/Projects/Stackdome` · branch `graphite-pass-2`
Working with **Jaseem** — a designer with basic coding knowledge.

---

## Read these first, in this order

| # | Path | Why |
|---|---|---|
| 1 | `DESIGN-PRODUCT.md` | **The only authority.** If code and it disagree, the code is wrong |
| 2 | `docs/design/create-stack-figma-plan.md` | **Your plan.** Phases, components, frames, verification |
| 3 | `docs/design/prototypes/create-stack/README.md` | What the design settles, and why |
| 4 | The prototype (below) | The behavioural reference you are drawing |

**The prototype:** https://claude.ai/code/artifact/6d649912-8686-4915-b5e9-501eeb821e40
Or build it locally: `node docs/design/prototypes/create-stack/build.mjs`

The audit that produced it (24 issues, all measured):
https://claude.ai/code/artifact/9b332f32-91a3-4b06-aa7f-fe1ec2f61287

**Do not re-derive the design.** It went through five rounds with Jaseem and is
agreed. Your job is to put it on the board.

---

## The job

Figma file `2IcCJOgsROpgajjXlay1h9` — *Stackdome — Shape + Hierarchy Pass*.

Three phases, all specified in the plan (doc 2). Start at **Phase 1**.

Roughly: five new components → five Light frames → one Dark + one empty state.
Frame 00 carries most of the screen work; 01–04 are variations of one layout.

---

## Things the plan cannot tell you, that cost me time

**The canvas is out of scope.** Jaseem cut it in the last exchange — nodes,
wires, the node inspector, deploy, and alerts-in-place are **a separate project**.
The prototype still shows all of it. Do not model it.

Consequence: **switch** and **inline alert** are approved components with no home
on this board. They will sit on the components page unplaced. That is expected —
do not go looking for a screen to put them in.

**The illustration node** inside `glyph / architecture` (`201:1234`) is a visual
device for empty states only. Jaseem was explicit: it is *not* the basis for a
canvas UI component.

**Two rules break most designs here.** Both are in `DESIGN-PRODUCT.md` but easy
to skim past:
- **Orange is not in the product** (§7). It belongs to the website and to
  thresholds. Black is the only action colour. The current wizard is full of
  orange — that is the thing being fixed, not a pattern to copy.
- **No uppercase** (§11) — not for eyebrows, not for column headers, not for the
  block registry's own `SERVICES` / `DATABASES`.

**"Update the existing primitive rather than creating a new component — and ask
first"** is the rule broken most often. The seven asks in this flow are already
settled in the plan; if a new one appears, ask before building.

**`docs/superpowers/` is gitignored.** Anything that must persist goes in
`docs/design/`. That is why the plan lives there and not in `plans/`.

---

## How to work with Jaseem

- **Short lines, bullets, tables.** No long paragraphs. Bold what must not be missed.
- **Show, don't describe.** Screenshot every frame you build.
- **Measure, don't eyeball.** He will ask for numbers, and he is right to.
- Give a rough time estimate before starting something.
- Skip engineering detail unless he asks.
- He pushes back hard and usually correctly. When he does, **go and check the
  evidence** rather than defending the design — twice in this session his
  pushback exposed a real gap I had rationalised.

---

## State of the repo

Uncommitted, all additive, no product code touched:

```
 M .gitignore                                  # ignores the generated prototype.html
?? docs/design/create-stack-figma-plan.md
?? docs/design/prototypes/create-stack/        # README, source, build.mjs
```

Nothing in `frontend/` or `pkg/` was changed this session. Scratch Playwright
harnesses I created were deleted; the pre-existing `.probe-*` / `.shot-*` files
in `frontend/` are from earlier sessions — leave them.

---

## Suggested skills

| Skill | When |
|---|---|
| **`figma:figma-use`** | **Mandatory before every `use_figma` call.** Do not skip it — it is how components, variants and variable binding actually work |
| `figma:figma-generate-library` | Load alongside `figma-use` for Phase 1. It covers what to build and in what order for components, variants and token binding |
| `figma:figma-generate-design` | Phase 2/3, when assembling the screen frames |
| `superpowers:executing-plans` | If you want the plan driven step by step |

`get_metadata` and `get_variable_defs` need no skill and are the cheapest way to
read the board. `search_design_system` is **useless here** — these components are
local to the file, not published to a library, so it returns nothing.

---

## Definition of done for this piece

Phase 3 complete, and the verification section of the plan actually run:

- every fill, stroke and shadow bound to a variable — no raw hex
- no orange, no uppercase, radius matching height
- each frame screenshotted and compared to the prototype at 1440
- Jaseem has seen the screenshots and said yes

Then, and only then, code — Storybook-first per `CLAUDE.md`.
