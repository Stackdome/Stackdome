# Product redesign — running log

> ## This document is HISTORY, not instruction
>
> **`DESIGN-PRODUCT.md` is the only authority. Read that first, not this.**
>
> This log records how the system got here — what was tried, what was reverted,
> and why. It is useful for understanding a decision's reasoning. It is **never**
> a source of rules, and where it disagrees with `DESIGN-PRODUCT.md` it is simply
> out of date.
>
> **Two known defects in this record, kept deliberately:**
>
> - Its goal statement — *"bring the website's design language into the
>   product"* — was **superseded**. The product follows the console reference;
>   the website's language is confined to thresholds. This is now §1 and §7 of
>   the rules.
> - The section headed *"Promoted — the rules are now the real components"*
>   claims 457 replacements across 129 files. **Git shows those tokens first
>   appearing in a single commit two days later.** Treat that section as
>   narrative, not as a record of what shipped when.

**Historical handoff notes follow.** They describe where things stood at the time
of writing, not where they stand now.

Last updated: 2026-08-04 (promoted to real components; Button settled)

---

## The goal

Bring the **website's** design language into the **product**, and write it down
as rules so that anyone prompting Claude gets the same result.

The team likes the website. The product is stock shadcn. PR #211 already
attempted the port; it got the tokens right and the translation wrong.

---

## Where things live

| What | Where |
|---|---|
| Product repo | `~/Projects/Stackdome` (this repo) |
| Website repo | `~/Projects/stackdome-website` |
| Stale local website copy — **ignore**, `src/` is empty | `~/Projects/stackdome-react` |
| Audit screenshots | `~/Projects/audit/` |
| Website token source of truth | `stackdome-website/src/directions/graphite.ts` |
| Website design reasoning | `stackdome-website/DESIGN.md`, `DESIGN-PROMPT.md` |

**There is no `DESIGN-LANGUAGE.md`.** It was believed to exist; it does not,
anywhere. The website language lives in `DESIGN.md` + `DESIGN-PROMPT.md`.

### Branches

```
main
 └── graphite-redesign      ← PR #211, Akshay's redesign (do not edit)
      └── graphite-pass-2   ← ours, local only, nothing pushed yet
```

PR #211 already contains PR #200's Storybook — one branch has everything.
When we push, the PR targets `graphite-redesign`, **never `main`**.

### Working surface

`pnpm --prefix frontend storybook` → <http://localhost:6006> → **Shell / Platform**

`frontend/src/stories/shell/platform-shell.stories.tsx` renders the **real**
sidebar + topnav + Stacks page together on mocked data — real components, not a
mock-up. **All design decisions get made against this**, never against an
isolated component.

**The rules themselves live in `DESIGN-PRODUCT.md` at the repo root.** Read that
first; this log is the reasoning and the history behind it.

---

## How we work (the user's explicit instruction)

**Never present a component decision in isolation.** Asked "how should the input
look?" with no screen around it, he has no basis to answer. Build the real
screen, show whole-screen variants as images, let the reactions become component
rules. Component-by-component coverage is the goal, reached *through* screens.

**Screen order** — chosen so the component surface is covered fastest:

1. App shell + Stacks — nav, page header, card, filters, status, buttons
2. Stack detail — tabs, logs, timeline, metrics, empty states
3. Users — tables, chips, menus, dialogs
4. Create flow — inputs, selects, validation, wizard
5. Auth — threshold screens

**Also:** flag context-window cost *before* image-heavy steps. Screenshots are
~95% of token burn. Save to `~/Projects/audit/` and let him open the folder
rather than pulling every image into the conversation.

---

## The reference: OpenAI

The website was based on OpenAI's design. The component styles come from there.
Studied ChatGPT web UI on Mobbin. **Three systems make it feel crafted:**

### 1. Radius scales with the size of the element

| Element | Radius |
|---|---|
| Small buttons, chips, toggles | full pill |
| Sidebar items, list rows, menu items | ~8px |
| Modals, composer, panels | ~16–20px |

Pill is **not a style** — it is what small things get. The current product makes
everything a pill regardless of size. That is the single biggest craft failure.

### 2. Filled buttons are rare

Nearly every button is a **white pill with a hairline border** (`Manage`,
`Archive all`, `Log out`, `Share`). Filled appears twice only: the send button
(black circle) and destructive (red pill).

The current product fills black for every primary action.

### 3. Separation is hairlines, not cards

Settings rows are label-left / control-right with a 1px line between. No boxes,
no shadows, no card per item. The product wraps everything in bordered,
shadowed cards.

### Three more

- **Colour is nearly absent.** Greyscale throughout; colour only for destructive
  red and app logos. No decorative brand colour anywhere.
- **Active state** is a soft grey rounded rect — no colour, no border, no pill.
- **Nothing is large.** Body ~13–14px, section titles ~16px semibold. **There is
  no display type in the product at all.**

**Conclusion:** the website copied OpenAI's *marketing* scale. The product needs
OpenAI's *product* scale. Same family, different system. That gap is the missing
translation layer.

---

## Audit of PR #211 — what's right, what isn't

**Right, leave alone:**

- **Tokens are exact.** Every value matches `graphite.ts` — `#f5f4f1` paper,
  `#191714` ink, `#ff6007` brand, status greens. No drift.
- **Tables read well** — Users and Projects have sensible product density.
- Black-as-action is applied consistently (though see OpenAI §2 — it may be
  applied too often).

**Wrong — one root cause: marketing rules copied literally instead of translated.**

1. **Page headers are marketing-scale.** `Stacks`, `Projects` at hero size with
   an orange eyebrow (`Platform`, `Settings`) above.
2. **The eyebrow is the worst offender** — it breaks the project's own rule,
   *"if a colour doesn't report something, it's a bug."* "Settings" reports
   nothing, in the colour reserved for signal.
3. **Air became emptiness.** On Stacks, one card occupied ~12% of a 1440px
   screen. Cards are ~180px tall to show a name, two numbers and a status.
4. **Everything is a pill** — search, filters, dropdowns, chips, badges, buttons,
   sidebar active item. Nothing signals "this is the action."
5. **Status shouts three times** — mono uppercase coloured text + coloured dot +
   coloured card edge, for one fact.
6. **Sidebar doesn't hold.** Same background as the page, active item is a
   floating grey pill, ~44px row pitch. Reads as links, not structure.
7. **Mono type is scattered** — email mono, name sans, `INVITED` mono,
   `Developer` mono. No rule behind it.
8. **"Manage members"** sits in a table column as plain grey text — reads as
   data, not a link. Identical on every row.
9. **Dropdown menus are oversized** relative to the tables they open from.

**Real bug, not styling:** the sidebar avatar renders `CN` while the user is
`Ada Lovelace` — initials are hardcoded. `frontend/src/components/nav-user.tsx`.

---

## Decisions made

- **Fix on top of PR #211, do not restart it.** Foundation is sound; the
  translation layer is missing.
- **Work on `graphite-pass-2`, branched off `graphite-redesign`.** Our PR will
  target their branch. Nothing merges to `main` until both agree.
- **Screen-driven redesign**, order above.
- Rules land in **`DESIGN-PRODUCT.md`** — **written**, at the repo root. That is
  the file to read before changing anything visual; this log is the reasoning
  and the history behind it.

## Settled since

| Question | Answer |
|---|---|
| Radius scale | **6 / 8 / 12 / 16**, pill for small controls only |
| Card vs row for Stacks | **Rows.** Cards stay as a second view; toggle planned |
| Eyebrow on page headers | **Gone**, everywhere |
| Surfaces | White sheet on a paper frame |
| Type | Named scale anchored on 13px |
| Control heights | 28 / 32 / 40 |

## Still not decided

- **Motion.** Inconsistent across the app; the user named this explicitly. The
  button press is the only motion with a rule behind it so far.
- **List/cards toggle** for Stacks — agreed in principle, not built.
- **The card itself** still says status three times (rail + mono word + dot).
  Needs its own pass before the toggle ships.

---

## Stacks pass 2 — the exploration (file since deleted)

**Where it was:** Storybook → **Shell / Stacks pass 2**
(`frontend/src/stories/shell/stacks-pass2.stories.tsx`)

Self-contained on purpose — it borrows nothing from the live shell, so the three
systems can be judged without the current components arguing back. Nothing is
wired to the app. What survives the reaction gets promoted into the real
components.

Two stories: **White sheet** (adopted) and **Flat paper** (comparison only).

**Images** (`~/Projects/audit/`): `p2-sheet-light.png`, `p2-sheet-dark.png`
(adopted); `p2-stacks-light.png`, `p2-stacks-dark.png` (flat paper);
before = `p2-before-stacks-light.png`.

What changed, against the audit list:

| Audit item | Applied |
|---|---|
| 1 · marketing-scale header | Title is 16px semibold, count beside it, no subtitle |
| 2 · orange eyebrow | Gone |
| 3 · air became emptiness | 8 stacks now occupy the space 3 cards used |
| 4 · everything is a pill | Pill = small buttons only. Rows, nav items, text field = 8px |
| 5 · status shouts three times | One 7px dot + one neutral word. No rail, no mono, no colour text |
| 6 · sidebar doesn't hold | Own surface + right hairline, 32px rows, 8px soft-grey active |
| 7 · mono scattered | No mono on the screen at all |
| 9 · oversized menus | List rows carry the actions; kebab appears on hover |

Filled buttons: exactly one on the screen (`New stack`). Everything else is a
hairline control.

### Finding — the radius tokens are themselves marketing-scale

`--radius-sm: 9px`, `--radius-md: 14px`, `--radius-lg: 22px` (`frontend/src/index.css`).
So `rounded-lg` on a 32px control produces a near-pill. **The "everything is a
pill" symptom is partly a token bug, not just call-site misuse** — pass 2 had to
hard-code `rounded-[8px]` to escape it.

Proposed product ladder (not yet applied): `sm 6 / md 8 / lg 12 / xl 16`, pill
reserved as an explicit `rounded-full`.

### Real bug still open

`frontend/src/components/nav-user.tsx` hardcodes the avatar initials (`CN`).

---

## Surfaces — the OpenAI system (from Mobbin, ChatGPT web)

| Surface | Colour |
|---|---|
| Sidebar (the frame) | grey |
| Main content plane | white |
| Composer / search field | grey, **no border** |
| Code blocks, plan-benefit panels | grey rounded box |
| User's own message | grey pill |
| Assistant's reply | no container at all |
| Modals, dialogs, popovers | white |
| Buttons (`Manage`, `Export`) | white + hairline |
| Active nav item, hover | grey rounded rect |

**The rule — and it inverts the common instinct:**

> **White floats. Grey recedes.** Higher elevation = whiter. Grey never means
> "a card"; grey means *pushed back into the frame*.

Corollaries: the greys are 2–4% ink, never mid-grey. Grey boxes are for things
you **type into** or **read but don't act on** — never for a list of things you
click.

**We already own the tokens.** `graphite.ts` defines `surface: #FFFFFF`,
`surface2: #FAF9F6`, `bg: #F5F4F1`. Pure white *is* the website's main surface.
No new value needed — `--card` is already `#FFFFFF`.

**Decided (user, this session): white is the main surface.** The screen is
`Shell / Stacks pass 2 → White sheet`; `Flat paper` is kept only as the
comparison.

---

## Grouping — what lives in the frame vs the sheet

The surface a thing sits on **is** a statement about its scope. OpenAI splits it
cleanly:

| Plane | Holds |
|---|---|
| **Grey frame** (sidebar) | Navigation, global helpers (`View plans`), the account block |
| **White sheet** (content + its top bar) | Only what is scoped to the thing you are looking at (`Share`, the kebab) |

**Applied:** `Docs` moved out of the sheet's top bar into the sidebar bottom.
Docs is about the product, not about Stacks. The sheet's top bar is now
breadcrumb-only — **an empty right side is correct, not unfinished.**

### The org section

ChatGPT's sidebar head is the quietest thing on the screen — no workspace name.
The switchable context sits in the **account block at the bottom** (`Alex Smith`
/ `Free`).

**Applied:** sidebar head is `stackdome` alone, one line. The org moved to the
account block's second line (`Ada Lovelace` / `Acme Corp`), replacing the email.
The org is the thing you *switch*; the account menu is where switching lives. An
email is not switchable and does not earn a permanent line.

**Reported, not applied:** the alternative is an org switcher in the sheet's top
bar (ChatGPT's `ChatGPT 4o ⌄` slot). Rejected for now — the breadcrumb already
owns that position.

### Type scale — anchored on 13px body (**decided**)

**13px is the base.** Named by job, not by size, so a call site declares what
the text *is* and cannot drift. Shipped as Tailwind tokens in
`frontend/src/index.css` — use `text-body`, never `text-[13px]`.

| Token | Size / line | Job |
|---|---|---|
| `text-label` | 11 / 16 | Group labels, avatar initials, overline |
| `text-meta` | 12 / 16 | Row data — branch, counts, status, timestamps |
| **`text-body`** | **13 / 20** | **BASE.** Nav, buttons, breadcrumbs, prose, inputs |
| `text-name` | 14 / 20 | The thing you scan a list for |
| `text-title` | 16 / 24 | Page and section titles — the largest type we ship |
| `text-head` | 20 / 28 | Dialogs and empty-state headlines only |

**Every line-height is a multiple of 4** (user's rule), so text lands on the same
rhythm as the 4px spacing grid and never pushes a row off by a pixel.

Nothing above 20px exists in the product — display type belongs to the website.
**Three weights only:** 400 default, 500 interactive/emphasis, 600 titles.

#### Spend the scale sparingly (**user's rule**)

> Having the tier ≠ using the tier. **13px carries the screen.** A size step has
> to be *earned* — and in a product this complex, exceptions will happen. Break
> the rule knowingly, not by default.

| Tier | Budget per screen |
|---|---|
| `text-head` 20 | Dialogs only. Never on a page. |
| `text-title` 16 | **Exactly one** — the page title. |
| `text-name` 14 | Only where a list has *one* scannable subject. Prefer weight. |
| **`text-body` 13** | **Everything else.** The default; no justification needed. |
| `text-meta` 12 | Pure furniture — group labels, avatar initials. |
| `text-label` 11 | Reserve. Not currently earned by any screen. |

**Test it before adding a size:** can weight (400/500/600) or colour
(ink/fg-2/fg-muted) do this instead? If yes, use those — they were already
carrying most of the hierarchy.

**Proved on screen.** `Shell / Stacks pass 2 → Restrained type` drops the list
from 5 sizes to 3 (16 title · 13 everything · 12 furniture). Images:
`p2-restrained-light.png` / `p2-restrained-dark.png`.

| | Scaled | Restrained |
|---|---|---|
| Sizes on screen | 5 | **3** |
| Name vs data | size + weight + colour | weight + colour |
| Row height | same | same |
| Cost | — | long image refs truncate; 13px is wider than 12px |

#### Why these steps

- **13 is the base because it is the interaction size.** The smallest size that
  stays legible without focusing. Everything you click lives here.
- **Row data is 12 because it is scanned, not read.** A stack row has five
  fields; at 13px they all have equal claim on the eye and the row reads as a
  sentence instead of a record. 12 makes the name the *subject* and the rest the
  *predicate*, and buys the horizontal room five columns need.
  **Not 11** — branch, status and counts are load-bearing facts and must never
  read as fine print. The 13→12 step is ~8%: enough to rank, not enough to
  demote.
- **Names are 14 — one step above base.** Enough to lead their row, not enough
  to tie with the page title.
- **Titles are 16, and we stop there.** In a product the page title is the least
  useful text on screen: you already know what page you are on, you clicked to
  get here. It is orientation, not reading. On the website the title *is* the
  content — that is why the website runs to 64px+. Same family, opposite job.
- **Five sizes over a 5px range, because size is one of four channels.** Colour
  (ink / fg-2 / fg-muted), weight (400/500/600) and position carry the hierarchy
  with it. If size worked alone we would need eight steps and a 2× range — the
  reason the screen reads calm is that **nothing has to be big to be important.**
  A textbook 1.25× modular scale gives 13/16/20/25: too coarse to get three
  distinguishable sizes inside one table row.
- **Line heights follow the job.** 13/20 is generous — the base carries prose and
  stacked nav and needs air. 12/16 is tight — row data is always one line, and
  extra leading only inflates the row. 14/20 deliberately *shares* 20 with body
  so a name and a body line sit on the same rhythm. 11 and 12 share 16 so a
  label and a meta value swap without relayout.

### Text colour — three tiers, one job each

Audit of the rendered screen found **four** tiers, with jobs overlapping:
inactive nav labels and row data both `#5C574E`; project name, group label and
timestamp all `#8B8579`.

| Tier | Token | Job |
|---|---|---|
| **ink** | `#191714` | What you came to find — stack names, **all nav labels** |
| **fg-2** | `#5C574E` | Data you read — project, branch, counts, status word |
| **fg-muted** | `#726C63` | Furniture and time — group labels, ages, breadcrumb `/` |
| ghost | `#BEB9AF` | Placeholder text only |

**`fg-muted` was darkened this session.** At `#8B8579` it measured **3.67:1** on
white and **3.33:1** on the paper frame — below the 4.5:1 AA needs for text under
18px. That was survivable while metadata sat at 13px; the moment the type scale
put row data at **12px** it became a real failure in a list people scan. Now
`#726C63` (5.2:1 white / 4.73:1 paper) and `#8A8479` in dark (4.64:1 card /
5.04:1 page) — **dark was failing too.** Both are the lightest greys that clear
their worst-case background, so the tier stays as quiet as the rule allows.
`--fg-ghost` is untouched: WCAG exempts placeholder and disabled text.

**The correction that mattered most:** OpenAI's nav labels are **near-black at
rest**. The grey is carried by the *icon*, never the word. Greying the labels
made the entire sidebar read as disabled.

---

## Control heights — proposed ladder

Primitives currently use **seven** heights (28/32/36/40/44/48/64). Proposed:

| Height | For |
|---|---|
| 28px | Chips, tags, in-row actions |
| **32px** | **Default** — buttons, search, filters, selects on a page toolbar |
| 40px | Form fields in a create/edit flow, and that form's primary button |

Nothing else. Rule: **height follows density, not importance** — an important
button gets *filled*, not taller. Structure: sidebar rows 32px (aligns with the
control default), topnav + sidebar header 52px.

User's reaction: "that's good." Not yet proven on a form-heavy screen.

---

## Promoted — the rules are now the real components

Pass 2 stopped being a picture. Everything below is in the shipped app, not a
story file. **1844 tests pass; lint and tsc clean.**

Images: `promoted-stacks-light.png`, `promoted-stacks-dark.png`.

| File | Change |
|---|---|
| `frontend/src/index.css` | Radius ladder **6 / 8 / 12 / 16** (was 9/14/22/22); `--radius` 14→12 |
| `components/ui/button.tsx` | Height ladder **28 / 32 / 40**; `icon` 40→32, new `icon-sm`; `text-body` |
| `components/ui/input · select · textarea` | Fields take the **8px step, not the pill**, and the 32px height |
| `components/ui/sidebar.tsx` | 32px rows, body-size **ink** labels, `fg-2` icons |
| `components/app-layout.tsx` | White sheet (`bg-card`, 12px, hairline) on the paper frame; 52px topnav; theme toggle out |
| `components/app-sidebar.tsx` | 52px head, `stackdome` alone; theme toggle moved into the frame as an "Appearance" row |
| `components/nav-user.tsx` | **`CN` bug fixed** — real initials; second line is the org, not the email |
| `components/theme-toggle.tsx` | New `presentation="row"` for the frame placement |
| `components/branded/page-header.tsx` | Eyebrow **ignored**, 16px title, no bottom rule |
| `pages/stacks/components/list/stack-row.tsx` | **New.** Hairline row replacing the 210px card |
| `pages/stacks/components/list/index.tsx` | Grid of cards → `divide-y` rows; count in the header |
| `pages/stacks/components/list/stack-card.tsx` | `HEALTH_LABEL` — `"ok"` → `Healthy`, `"progressing"` → `Deploying` |

### Two things the real screen exposed that the story could not

| Found | Fix |
|---|---|
| Search/select were **pills** — a field is not a small control | Inputs, selects, textareas moved to `rounded-md` (8px) |
| Status showed raw API strings (`ok`, `progressing`) beside `Not deployed` — one column, two voices | `HEALTH_LABEL`, typed by `ReleaseHealth` so a new value fails the build |

### The eyebrow

`PageHeader` accepts `eyebrow` and **ignores** it, so all twelve call sites still
compile and every page loses its eyebrow at once. `EyebrowLabel` itself stays —
`panel.tsx` and the canvas nodes still use it legitimately.

## Storybook is the source of truth (**user's rule**)

> **No new component if one already exists** — update the primitive and use it.
> Ask first, every time. A design story that hand-rolls its own button or row is
> a standalone picture, and a standalone picture is useless: it proves nothing
> and gets built twice.

**Applied:** `stacks-pass2.stories.tsx` is **deleted**. Everything it
hand-rolled — QuietButton, PrimaryButton, NavItem, NavGroup, its own sidebar and
its own row — now lives in the real primitives. `Shell / Platform` renders the
same screen out of `AppLayout` + `AppSidebar` + `StacksPage`, so it *is* the
design. Also removed: a `ViewToggle` started without permission.

### Full primitive sweep

The scale is now enforced everywhere, not just where Stacks touched it.

| Was | Now | Scope |
|---|---|---|
| `text-sm` (14) | `text-body` (13) | 18 primitives + 111 files |
| `text-xs` (12) | `text-meta` | same |
| `text-lg` (18) | `text-title` (16) | dialog + alert-dialog titles |
| `text-xl / 2xl / 3xl` | `text-head` (20) | error states, wizard headlines — **`head` finally earns its slot** |
| `font-bold` (700) | `font-semibold` (600) | 3 weights only |
| `text-[10.5px]` mono uppercase | `text-meta` | badge — mono had no rule behind it |
| `text-[11.5px]` mono | `text-meta` | toast description is copy, not terminal output |
| `size-9`, `h-9`, `h-12` | the 28 / 32 / 40 ladder | breadcrumb, tabs, command |

**457 replacements across 129 files. Nothing above 20px exists in the product.**

### The bug that broke every merged component — read this before adding a token

Naming the type scale by job (`text-body`, `text-meta`) collided with
**tailwind-merge**: it classifies any `text-*` utility it does not recognise as a
text **colour**. So `cn("text-body", "text-fg-2")` dropped the size entirely and
the element fell back to the **inherited 16px**.

Silent, and only in components that merge through `cn()` — which is every
primitive. The sidebar group labels rendering at 16px were the visible symptom;
the cause was in `frontend/src/lib/utils.ts`.

**Fixed at the root:** `cn` now uses `extendTailwindMerge` with the scale
declared as a `font-size` class group.

> **Any new named utility that shadows a Tailwind prefix must be registered
> there too, or it will be silently dropped wherever `cn` merges it.**

### Half-sizes removed

162 further replacements: `15px` → `name`, `12.5px` → `meta`, `11.5 / 10.5 / 10px`
→ `label`, `29 / 30px` → `head`. **The scale has no half-steps.**

Verified by rendering: every text node on Stacks, Users, Secrets, Tabs now
computes to one of **11 / 12 / 13 / 14 / 16 / 20**.

Two deliberate exceptions, both flagged rather than swept:

| Exception | Why it stays |
|---|---|
| `9 / 9.5px` mono micro-labels in the stack editor | Whole editor is unconverted; belongs to the Stack-detail pass |
| `120–200px` numeral on the 404 page | A threshold screen, not a product screen |

### Primary button — the white rim

The visible "border" on the black button was **`--edge`**:
`inset 0 1px 0 rgba(255,255,255,0.7)`. That token is tuned for *light* controls
on cream, where 70% white is almost invisible. On a black fill it is a hard
white rim.

New token **`--edge-fill`** — 12% white in light, 35% in dark (where the primary
button inverts to a light fill and needs an ink-side lift instead). Applied to
`default`, `destructive`, `inverse`. `--edge` stays on the light controls it was
designed for.

| | Was | Now |
|---|---|---|
| Top highlight | 70% white — reads as a border | **12%** — reads as a lift |
| Icon | 16px, out-weighing 13px text | **14px** |
| Icon → label gap | 8px | **6px** |
| Glyph | `PlusCircle` — a circle inside a pill | **`Plus`** — no competing curve |
| Size | 109 × 32 | 105 × 32 |

**Rule:** a highlight tuned for one surface is not a highlight on its inverse.
Any inset-light token needs a fill-side counterpart.

### Press — the Polaris model (**user's ask: "tactile, like Shopify"**)

Sources: [Polaris `Button.module.css`](https://github.com/Shopify/polaris/blob/main/polaris-react/src/components/Button/Button.module.css),
[shadow tokens](https://polaris-react.shopify.com/tokens/shadow),
[Uplifting Shopify Polaris](https://halfool.medium.com/uplifting-shopify-polaris-7c54fc6564d9).

**The button does not move. Its content does.** Polaris applies
`transform: translate3d(0, 1px, 0)` to the button's *children*; the pressable
element stays put. Three things fire together and none of them works alone:

| On press | |
|---|---|
| 1. Highlight | **disappears** |
| 2. Inner shadow | **appears** — 3px, hard, zero blur |
| 3. Label + icon | **shift 1px down** |
| 4. Background | **darkens** |

The rest highlight is offset **and** spread — `inset 0 0.5px 0 1.5px` — not a
flat 1px line. That inset ring is what reads as moulded plastic rather than a
drawn border.

**They took 10 iterations on the primary button alone**, and the article names
the blocker: *"the primary button initially lacked darker values to make it feel
three dimensional… after making the default button a touch lighter, we landed on
the level of juiciness."* Our face is ink at rest, so the press gets its own
darker step: **`--primary-press` `#0D0C0A`**.

**Implementation:** `Button` wraps its children in a content span (skipped for
`asChild`, which already supplies one element), and the base class carries
`active:[&>*]:translate-y-px`. Either path gives the button exactly one element
child to move — a bare text node cannot be transformed, which is why the wrapper
is needed at all. `has-[>svg]:` became `has-[svg]:` since the icon is now a
grandchild.

**Three wrong turns, recorded so they are not repeated:**

| Wrong | Why it failed |
|---|---|
| Blurred inner shadows | Reads as a dent, not a key |
| 1px *spread* on the bottom edge | Haloed the whole pill |
| Translating the button itself | The body must stay planted; only the content travels |

Also fixed: press inherited `hover:bg-primary-hover` and got **lighter** when
pushed — `:active` always fires while `:hover` is true.

### Then: keep the mechanic, drop the material (**user's call**)

> *"Imitating exactly was a bad idea, it does not look great. Keep the recess,
> the inner shadow and the label moving. Remove everything else — the glow, the
> box shadow."*

Copying Polaris's full material wholesale fought the graphite language: the
shine and the bevel read as someone else's product sitting inside ours.

| Kept | Dropped |
|---|---|
| Inner-shadow recess on press | Rest highlight / "shine" |
| Label + icon travelling 1px | Bottom-edge thickness |
| Fill darkening on press | Ground shadow |

**Buttons are now flat at rest — `box-shadow: none`, verified in the browser.**
Depth exists only while you are touching the button.

### Optical padding — icon side tightens

A glyph carries its own whitespace inside its bounding box, so equal padding
*measures* equal and *reads* lopsided. The side with the icon gets less.

**Anchored by tightening the icon side, not widening the label side**, so the
label sits the same distance from the edge on every button and a column of
buttons reads aligned.

**Measure the glyph, do not guess it.** Two rounds were wrong because the
correction was assumed:

| Attempt | Icon-side inset | Optical result |
|---|---|---|
| 1 | widened label side by 4px | ~2× over-corrected |
| 2 | tightened icon side to 10px | left still ~1px heavy |
| **3** | **9px** | **11.9 vs 12.0 — level** |

A lucide 14px icon carries **2.92px of air each side** (`getBBox()` against the
24-unit viewBox), not the 2px assumed. Icon inset = label inset − glyph air.

**The gap mattered more than the padding.** At 6px box (≈8.9px optical) it was
nearly as wide as the 12px side inset, so the icon read as a *separate object*
beside the label rather than part of it. Now 4px box ≈ 6.9px optical.

### Variants: 11 → 7

Four rendered **identically** to something that already existed — confirmed by
comparing computed style, not by reading the class strings.

| Removed | Was identical to | Call sites |
|---|---|---|
| `mono` | `ghost` | 0 |
| `railGhost` | `ghost` | 1 |
| `railPrimary` | `secondary` | 1 |
| `railDanger` | — (distinct, but existed only to pair with the rail set) | 0 |
| `size="rail"` | `size="default"` | 2 |

All of it lived in **one file**, `sticky-action-bar.tsx`, which also
hand-rolled its own spinner — now migrated onto `loading` / `loadingText`, so
the removal deleted code rather than moving it.

Safe because the Button styles children with **descendant** selectors
(`[&_svg]`), not direct-child ones, so the content wrapper does not affect them.

**Remaining: `default · destructive · outline · secondary · ghost · link ·
inverse`.** Anyone picking a variant now gets one obvious answer instead of four
ways to spell the same button.

### Loading state

`loading` swaps the content for a spinner and makes the button inert.
`loadingText` says **what is happening** — `"Creating…"`, `"Deploying…"` — not
that something is; the spinner already reports *that*.

| | |
|---|---|
| `loadingText` given | Replaces `children` wholesale, so a leading icon disappears on its own — no doubled glyph |
| `loadingText` omitted | The idle label stays beside the spinner. Fine for a verb that reads the same either way (`Save`) |
| Contrast | **Stays at 100%.** A request in flight is not a disabled control, and the label must stay readable |
| Accessibility | `disabled` + `aria-busy="true"` |

The dim is suppressed for this case only —
`disabled:[&:not([data-loading])]:opacity-50` — so a genuinely disabled button
still greys out.

Before this, call sites hand-rolled `{isLoading && <Loader2 className="h-4 w-4
animate-spin" />}` alongside their own `disabled` prop, each slightly
differently. Those can now collapse onto the primitive.

### Optical correction — the discipline, not a vibe

This has a name and a body of practice: **optical alignment**. Mathematical
alignment is exact and based on numerical rules; optical alignment is
perceptual, and the two do not agree because the eye reads **centre of mass**,
not bounding boxes. The same principle produces overshoot in type (a round `O`
is drawn taller than a flat `H` so the two *appear* the same height) and the
nudge that shifts a play triangle right of a circle's true centre.

**We aim for optical correction. The measurement is an input to it, not the
answer.** Three things carry invisible space that geometry cannot see:

| Invisible space | Effect |
|---|---|
| Icon safe area | The glyph sits inside built-in padding, so the icon side reads roomier than it measures |
| Text bounding box | Ascender/descender space above and below the ink |
| Rounded corners | A pill's curve pulls the visual edge inward near the ends |

**One rule, derived from each size's base inset** — not hand-set per size, which
is exactly how `sm` drifted to only +0.5px of correction while `default` had
+2.1:

```
icon side  = base − 3      (pays back the glyph's safe area)
label side = base + 2      (the optical margin dense letterforms need)
```

| Size | Base | Icon side | Label side |
|---|---|---|---|
| `sm` | 10 | 7 | 12 |
| `default` | 12 | **9** | **14** |
| `lg` | 15 | 12 | 17 |
| `rail` | 12 | 9 | 14 | 

**Verified optically, ink to ink:**

| Button | Icon side | Label side |
|---|---|---|
| `New stack` — leading `+` (2.9px safe area) | 11.9 | 14.0 |
| `Sort ⌄` — trailing chevron (4.5px safe area) | 11.5 | 13.0 |

The two glyphs carry different safe areas, so the correction lands slightly
differently on each — correct behaviour, not drift. **Chasing per-glyph values
would be over-fitting**; the rule holds the family together.

**Vertical needed nothing.** Measured against the font's real ink extents
(`actualBoundingBoxAscent/Descent` via canvas, not the line box): **11.27 above,
11.34 below — 0.04px drift.** Geist centres cleanly at 13/20. Descenders are
left to hang, which is the convention.

Sources: [Optical alignment — Gonçalo Dias](https://zalodias.com/notes/optical-alignment) ·
[Optical vs mathematical alignment in UI](https://railsdesigner.com/mathematical-optical-alignment-design/) ·
["Eyeballing" or optical alignment in design](https://medium.com/ringcentral-ux/eyeballing-or-optical-alignment-in-design-4ef5ab2d326f)

Final, `default` size: **11.9 | 6.9 | 14.0** (optical, ink to ink).

| Size | Text only | Icon side | Label side | Gap |
|---|---|---|---|---|
| `sm` 28px | 10 / 10 | 8 | **12** | 4 |
| `default` 32px | 12 / 12 | **9** | **14** | **4** |
| `lg` 40px | 15 / 15 | 13 | **17** | 4 |

Symmetric is kept where there is nothing to balance against: **icon-only**
buttons and **icons on both sides** (`not-has-` guards).

**One thing this needed:** the label is now wrapped in a `<span>`. Without it an
icon+label button has exactly one *element* child — a text node is not an
element — so the icon matches both `:first-child` and `:last-child` and CSS
cannot tell which side it is on.

### One geometry, intensity by face lightness

The recess was zero-blur on `destructive` and soft everywhere else — two
different components. Now every variant shares **1px side walls · 2px top
shadow · 2px blur**, and only the alpha moves:

| Token | Faces | Colour |
|---|---|---|
| `--btn-press-soft` | Light — `outline`, `secondary`, `ghost`, rail | ink @ .10/.18 |
| `--btn-press-mid` | Mid-tone — `destructive` | black @ .18/.34 |
| `--btn-press-strong` | Near-black — `default`, `inverse` | black @ .38/.62 |

> **Intensity is set by how light or dark the face is — never by how important
> the button is.** Importance is carried by the fill.

Dark theme flips the recess colour (light faces recess with black, the
near-white primary recesses with ink) but keeps the identical geometry. `--edge`
and `--press` are untouched — they belong to cards and non-button controls.

### Per-primitive design pass

Tokens alone were not enough — these needed a decision, not a find-and-replace.

| Primitive | Was | Now | Why |
|---|---|---|---|
| `card` | 8px, 24px padding | **12px (panel step), 16px padding** | A card is a panel, not a row. 24px of inset on a card holding a name and two numbers is what made the old grid read as empty. Keeps the `--edge` inset per D13/D14. |
| `dropdown-menu` | 8px, `shadow-lg` | **12px, `shadow-md`** | A menu is a floating panel. One opening from a 32px row should not cast a 24px shadow (audit item 9). |
| `tabs` | Filled grey trough, `bg-muted` | **Transparent track, soft-ink active chip** | The trough was a second surface competing with the sheet. Active is now the same wash sidebar rows use — "selected" means one thing product-wide. |
| `badge` | Tinted fill **+** coloured border **+** coloured text | **Fill + text, no border** | Three channels for one fact. Same de-shout as the status column. |
| `table` | Header 40px, body-size ink | **Header 32px, meta-size `fg-2`; cells body size** | A header is a row of labels; the data is the point. |

### Known debt

| Item | Note |
|---|---|
| `stack-card.tsx` | **Not debt — a second view.** The card stays; a list/cards toggle is planned. The card still says status three times (rail + mono word + dot) and needs its own pass. |
| Every other page | Inherits the new tokens (corners, heights, type) but keeps its old layout until its own pass. Expected, not a regression. |

---

## Next step

Stack detail — tabs, logs, timeline, metrics, empty states. Biggest screen,
most new components, and it now sits inside a shell that actually exists.

Open questions the screen is meant to settle:

1. **Row vs card** for the Stacks list — pass 2 commits to rows.
2. **Title at 16px** — is that too quiet for a page header?
3. **Status = dot + neutral word** — enough signal, or does failure need more?
4. **Sidebar on its own surface** — right call, or should chrome stay flat?
5. **Radius ladder** — adopt `6/8/12/16` as tokens?

After that: promote the survivors into the real components, then Stack detail.

---

## Housekeeping done this session

- Status line added showing live context usage:
  `~/.claude/statusline-command.sh`, wired in `~/.claude/settings.json`.
  Reads the session transcript and reports percent used and tokens left.
- Screenshot harness: `frontend/.shot.mjs` (Playwright, 1440×900 @2x, drives
  Storybook's `?globals=theme:` for light/dark). Untracked scratch file.
- **Tailwind gotcha:** new files added while Storybook is already running are
  missed by the Tailwind scan — arbitrary classes silently produce no CSS.
  `touch src/index.css` to force a rescan.

---

## The sheet header — pass 3

Screenshots: `tn-a…tn-d` (band count), `sc-d/sc-e/sc-f` (scale), `oai-header.png`
(ChatGPT, measured live at 1440px).

### What was wrong

The sheet stacked **three bars** before a single stack appeared — breadcrumb,
page header, filter toolbar — and said "Stacks" twice while the topnav's right
side sat empty. Separately, the sidebar wordmark sat **9px** above the
breadcrumb: the code claimed the two lined up across the seam, and they did not.
The sheet is inset 8px and draws a 1px hairline, so its band starts 9px down.

### Four options, judged on the real screen

| | Topnav right | Bands | Data starts |
|---|---|---|---|
| A · today | empty | 3 | 195px |
| B · action up | `+ New stack` | 3 | 187px |
| C · search ⌘K | search | 3 | 195px |
| **D · header absorbed** | fact + action | **2** | **143px** |

B saves 8px — not worth a move. C needs a product-wide search that does not
exist. **D chosen.**

### The bar is the top of the sheet, not a band on it

User's correction, checked against ChatGPT: their bar has **no divider at all**
and content slides under it. Ours now does the same — sticky inside the scroll
container, opaque `bg-card`, 8px fade on its underside so rows dissolve instead
of being sliced.

Three defects this introduced and the measurement that caught each:

| Defect | Found by | Fix |
|---|---|---|
| Fade at z-40, above `#page-sticky-bar` (z-30) | reading computed z-index | fade → z-20 |
| Fade cost **7px** of flow — a phantom scrollbar on the one full-bleed page | summing `offsetHeight + marginTop` per child | `-mt-2` cancels its own height → 0 |
| Padding drifted to 20/28 | same pass | back to 28/28 |

### Scale — measured, then translated

ChatGPT at 1440px: band **52px** (8px padding around a 36px row), title
**18/28 w600**, buttons **36px** pill at **14/20 w500**, icon buttons 36×36 with
a 20px glyph, 8px between actions.

Two candidates built and shot side by side:

| | E · their numbers | **F · chosen** | was |
|---|---|---|---|
| Title | 18 / 600 | **16 / 600** | 13 / 400 |
| Button | 36px, 14px | **32px, 14px** | 28px, 13px |
| Title ÷ body | 1.38× | **1.23×** | 1.0× |

Their 18px sits on a **16px** base; ours sits on **13px**. Copying the number
would have made our title *louder relative to the page* than ChatGPT's own is —
and added an 18px step to a scale that stops at 16 for titles. F takes the
ratio, not the number.

**The move worth stealing at any scale:** their button text is a step *up* from
the trail. Ours had trail, title, fact and button all at 13px — four things on
one line at one size, which is why the bar read flat.

### `Home` deleted

Every top-level destination is one click away in the sidebar, so a Home hop said
nothing the frame wasn't already saying — and it pushed the real title out of
first position. Consequence: `/` had no segment to name itself with, so it now
redirects to `/stacks` instead of rendering the page directly.

### Still open

- All 15 `PageHeader` call sites are untouched; the fact and action live in the
  topnav **in the story only**.
- The canvas stack editor is the only full-bleed page and has no shell story —
  its `h-[calc(100%-52px)]` fix has not been seen running.

---

## Alpha lines, and why Figma and the code never matched

The board drew its hairline as `#E8E6E2`. The code drew it as
`rgba(25, 23, 20, 0.11)`. Both looked defensible; neither was the other.

Composited over the sheet, the code's line lands on `#E6E5E5` — **flat neutral
grey**. The reason is the ink: `--border` was an alpha of `--foreground`, the
*text* ink, and `#191714` is very nearly achromatic. So the line got colder as
it got stronger — `#E6E5E5` at the hairline, `#D6D5D5` at the strong rung — on
a palette that is warm everywhere else. The designer had compensated by hand,
picking a warmer solid, which is exactly why the two sides drifted.

The fix is to derive the ink instead of borrowing it. Alphredo's method: solve
for the colour that, at the smallest alpha keeping every channel in range,
composites to the target on the background.

```
a   = max((bg − target) / bg)      per channel
ink = (target − bg × (1 − a)) / a
```

For `#E8E6E2` on white: **`rgb(53 35 0)` at 11.37%**, round-tripping exactly.
Three rungs off it (6 / 11 / 18%) now render `#F3F2F0`, `#E9E7E3`, `#DBD7D1` —
warm all the way up, and within one unit of *both* solids the board had chosen
independently. That agreement is the tell that the derived ink was the right
one: two hand-picked values, months apart, sitting on the same ramp.

Alpha rather than a solid because one line has to work on the white sheet, the
paper frame and a control fill. A solid is correct on one and wrong on two.

### Alphredo

<https://alphredo.app/> — generates translucent colours that match their opaque
counterparts against a known background. We use the method, not the tool.

## The toggle: the fix was the padding, not the border

The first segmented control avoided a doubled edge by giving the thumb no
border at all — a white block floating in a grey well. The updated board keeps
the border and removes the **track's padding** instead.

That is the better answer. The problem was never that the thumb had a border;
it was that an inset bordered thumb puts two edges 2px apart, and at that
distance the eye reads a doubled line rather than a raised surface. Run the
segment flush and both edges land on the same pixel. `overflow-hidden` gives
the segment the track's radius, so its corners *are* the track's corners.

The divider is drawn by the selected segment's own edge rather than by a
separate rule — a rule would eventually sit beside that edge and double it.

## The bevel, and the shadow scale that wasn't

`--edge` — an inset white top highlight — was applied at 14 places while §6 said
content is flat at rest. It is gone. The two worth naming were the switch thumb
and the radio dot, which are physical controls rather than content; neither was
relying on it, since both separate from their track by fill in every state.

The shadow scale advertised eight rungs and held four values: `2xs`/`xs`/`sm`
were one shadow, `shadow`/`md` a second, `lg`/`xl` a third. Four remain defined
and the rest alias them, so a call site still renders and now says out loud
which rung it lands on.

### Figma had no effect styles at all

Eight now exist — four elevations, four press insets — built from values
**measured in the running app**, not transcribed from the stylesheet. The
elevation styles carry the aliasing in their descriptions so the duplication
cannot quietly come back.

### The audit that prompted all of this

| | Figma | Code |
|---|---|---|
| Colour tokens | 16 | 64 |
| Radius | 5 | 5 ✓ |
| Shadows | 0 | 8 names, 4 values |
| Themes | 1 mode | light + dark |

Two variables had drifted outright: `surface/control` held `#FAF9F6`, which is
code's `--control-hover`, so a control at rest on the board was drawn as a
hovered one; and `surface/selected` had no code counterpart at all. The first is
corrected, the second deleted.

**Still open:** Figma has no dark mode, so dark is designed nowhere and exists
only in code. 48 code colour tokens — the status tints especially — have no
Figma counterpart, which means a designer cannot mock an alert or a badge
truthfully.

---

# Session — 2026-08-05 · the restart, and Phase 1 (shell)

The session that reset the project. Read `DESIGN-PRODUCT.md` for the rules; this
records what happened and why, including the things that went wrong.

## What triggered it

The concern raised: old and new tokens coexisting, principles not respected, the
direction drifting. **Audited — the concern was right, but the cause was not
drift.** Five documents claimed authority across two repos:

| Doc | Claimed |
|---|---|
| `DESIGN-PRODUCT.md` | The rules |
| `docs/design/redesign-log.md` | "Bring the *website's* language into the product" |
| `docs/design/openai-platform-study.md` | Console: "no brand colour anywhere", "no display type at all" |
| `stackdome-website/DESIGN.md`, `DESIGN-PROMPT.md` | The website's rules — which the auth screens were following |

The stated goal and the stated reference were **opposites**. Where research and
rules disagreed, nobody updated the losing document, so the code followed the
study on some points (`shape` defaulted to `pill` because the study said to keep
it) and the rules on others.

## Decisions taken

| # | Decision |
|---|---|
| 1 | **Split by surface.** Thresholds carry the website language; working surfaces are pure console. Test: *is there work in front of the user right now?* |
| 2 | Threshold = **auth · 404 · empty states · onboarding** |
| 3 | Tokens: **alias now, migrate later** |
| 4 | `DESIGN-PRODUCT.md` is the **only** authority. The study is evidence, the log is history — both now carry banners saying so |
| 5 | **Not optimising for mobile.** Desktop widths only |
| 6 | New component only when none exists in Storybook, or the existing look and feel cannot be altered. **Otherwise add a variant** |
| 7 | Done = when we are happy with what is on screen |
| 8 | Order: structure/nav/shell first, then page by page, journey by journey. Routing gets fixed per section as we reach it |
| 9 | Workflow: **artifact → Figma → code.** Figma lays the foundation; it is usually not exhaustive, and the design is extended in code using the rules. When the picture is unclear, Figma scopes it fully first |

Answered along the way: the product is for a **platform team, a developer who
explicitly does not want to learn infra, and a CTO** — from the website's own
copy. That settles density (comfortable) and makes the **canvas the headline**,
since "topology you can see beats YAML you can't" is the marketing promise.

## Phase 1 — the shell

Built to the `app shell` Figma board, every number verified in the browser.

- **10 `nav-*.tsx` files → `nav-items.ts` + `nav-item.tsx`.** 312 lines → ~90, and
  the nav became a list you can reorder rather than ten imports in a fixed order
- Grouping by **who touches it**: Stacks + Previews unlabelled, then Platform,
  then Infrastructure (already the admin-gated set)
- **Two-part sheet header** — title row identifies, toolbar row serves it. The
  toolbar row is conditional and collapses itself via `empty:hidden`, so the band
  is 56px or 100px with no height math anywhere
- **Every hardcoded `52px` offset removed.** Header, form save-bar and fade now
  travel as one sticky block
- Sheet: 2px gap to the rail, `outline` not `border`, `shadow-md`
- Collapse choreography — `rail-x` / `rail-y` / `rail-y-in` / `rail-logo`
- **The wash ladder** (§4) — hover/selected/pressed had all been 6%
- Figma: states built for Nav item (12 variants), Account block (8), Button (30),
  icon button (30), View toggle (6)

## Things that were wrong, and what they cost

Worth keeping — each one cost real time and each has a rule now.

| What | Why it mattered |
|---|---|
| **My verification measured the wrong things** | Positions and centres were asserted, but never the *gaps between* them — so a uniform-looking column with the wrong pitch passed three times |
| **A test that clicked a nav link** | It navigated, moved the selection, then measured. Reported two false failures, and I chased the code instead of the harness |
| **A probe testing classes Tailwind had never scanned** | "The utility resolves to transparent" was an artefact of the test |
| **Greys measured as channel alpha** | Overstated every warm grey. Measure the **L\* drop**, never "% of the text ink" — the same trap as the line ink, one level up |
| **A Figma paint bound to a variable but built from a placeholder hex** | Renders the placeholder. Screenshot disagreed with inspection |
| **`-mt-8` still on the collapsed group label** | Stacked on top of the new height collapse and knocked every group out of rhythm |

## Open

- **Auto-collapse at 1024–1280** — §12b states it; it is not built. Parked
- **Field + Select states** — they appear only in the toolbar row, which is Stacks
  content in the shell's slot. They land with that work

**Closed since this was written:** the Stacks toolbar now portals into
`#sheet-toolbar` via `PageHeader`'s `toolbar` prop, and the sheet's content edge
came in from 32px to the header's 12px (`max-w-6xl` removed with it — it capped
the body at 1152 while the header spanned the sheet, so the two planes drifted
apart above 1280).
- `--muted` → alpha ink tint (77 call sites)
- Users / Projects fully built and unrouted; `/settings/*` redirects away
- Below 1024 the rail becomes a drawer — inherited shadcn, undesigned
- Account component is 40px tall in Figma while every instance is 48px

---

# Session — 2026-08-05 (cont.) · Phase 2, the Stacks list

Rules live in `DESIGN-PRODUCT.md`; the implementation handoff is
`docs/superpowers/plans/stacks-list-page.md`. This records only what was learned.

## The job came first, and it cut more than it added

The artifact settled one sentence — *"tell me where things stand, take me to
whatever needs me"* — and then one test: **an item earns a place on the list only
if it does something a click cannot.** Help you choose a row, compare across
rows, or finish without leaving. Anything else belongs on the stack page.

That test killed four things, three of which were mine:

| Cut | Why |
|---|---|
| The public URL | Fails all three. It was also going to need an API field |
| The topology thumbnail | Four unlabelled boxes need notation the reader does not have. It performed *"this stack is complex"* without saying what is in it |
| The 14-day sparkline | Fourteen bars, no axis, no unit — nobody can tell eleven deploys from four |
| The `Services` count | The stack's shape written as text, unrecognisable at a glance |

The lesson worth keeping: **all three drew the *shape* of substance instead of
supplying it.** What replaced them was the same data as words — the components,
named — which anyone can read.

## The second line is a mechanic, not a field

A row grows to two lines **only when a human is needed**. Healthy rows stay one
line, so trouble is found by the *shape* of the list before a word is read or a
colour registers — which also means it survives for anyone who cannot see the red.

The first implementation printed any message the release carried, and an
in-flight release carries progress messages, so `Deploying` rendered a second
line and read exactly like a failure. The fix was a predicate that already
existed: `needsAttention()`. Now the header's *"N need attention"*, the default
sort and the two-line rows are **the same set**.

## Two claims that did not survive checking

Both were mine, both were stated confidently, and both were wrong in a way that
would have changed what someone built.

1. *"The resource-scoped reason costs a presenter change and nothing else."*
   Half right. The per-resource status **is** loaded in the Go process on the list
   path — `Status` is a jsonb column, so `Omit(clause.Associations)` does not drop
   it — but `PresentStackResource` never puts it in the API contract. It needs an
   OpenAPI change and a regenerate.
2. *"The card shows the reason in full."* True while the card could grow. Once the
   height was fixed it gets one line at 344px, where the row's status column has
   **504px** — so the row shows *more* of a failure than the card does.

## Figma renders some things and silently does not render others

**A drop shadow does not paint on a `COMPONENT` node.** Verified against an
identical `FRAME`, which paints it fine. Phase 1 had drawn every focus ring as a
spread-2 shadow, so **twelve focus variants across Button and icon button were
invisible on the board** and had been for a session. Focus is a 2px OUTSIDE
stroke now.

Three more, all in §16 of `DESIGN-PRODUCT.md`: an attribute selector cannot hold
a space and fails silently; `visible = false` inside auto-layout removes a node
from the flow, so a FILL sibling swallows its width; clone before you clear.

## Process

The **artifact → Figma → code** gate was broken once — card changes went into
code before the board settled, and a new idea (component icons on the chips) was
built in code "to show him" rather than drawn. Called out, and correctly. The
gate applies to **every** change, not just the first one, and a question is not
approval.

## Dark mode had two hue families, and nobody had measured them

Reported as *"I don't like the fill of the controls on dark mode."* Eyeballing
would have produced a nudged hex. Measuring found a structural fault.

The dark surface ladder runs at **chroma ~4** (OKLCH) — one warmth, level
carried by lightness. Four tokens had drifted off it:

| Token | Was | Chroma | Ladder |
|---|---|---|---|
| `--input` | `#1D1A15` | 10.5 | 4 |
| `--control` | `#1D1A15` | 10.5 | 4 |
| `--control-hover` | `#26221C` | 12.5 | 4 |
| `--popover` | `#201D18` | 10.4 | 4 |
| `--accent` | `#26221C` | 12.5 | 4 |

Two to three times the ladder's chroma. Every field, select, menu and hovered
menu row was **khaki on a neutral screen** — the only saturated thing in view,
which is exactly what "muddy" was naming.

**The second fault was worse and invisible until measured.** `--input` and
`--control` were the *same* hex, at L 21.9 — the card's own lightness (22.2).
So on a card a field had no fill at all, only a hairline; and the product's own
rule that *an input is a well and a select is a face* did not exist in dark.
Light had it (`--input` and `--control` both recess below the white sheet, and
the press mechanic separates them); dark had collapsed it into one colour.

Fixed by putting them back on the ladder **on opposite sides of the card**:
`--input` `#181715` (below), `--control` `#232220` (above), and the two washes
onto the ladder's existing rungs. `--surface-node` followed `--card`, the way
light already had it.

**The lesson is the measurement, not the values.** Light measures grounds by
lightness drop from white, and that number is useless in dark — the drops are
tiny and the eye reads colour first. Dark needs a chroma budget, so it now has
one, written into §3.

## The status glyph was switched off for a good reason, and the reason was fixable

Reported as *"can't really understand the status easily."* The icons already
existed in `StatusText` and were switched off at both call sites, each with a
written justification — so the first job was working out whether the old
decision was wrong or the old implementation was.

It was the implementation. The set was **per family**: three glyphs for
`ready` / `pending` / `error`, which meant `Degraded`, `Unavailable` and
`Failed` all drew the same triangle. The card's own comment named the defect
exactly — *"the family triangle cannot tell Degraded from Failed, which is the
distinction that changes what you do next"* — and then drew the wrong
conclusion from it: it removed the glyph rather than fixing the set.

Per state, the mark carries its own information and the objection evaporates:

| Healthy | Deploying | Degraded | Unavailable | Failed | NotDeployed | Deleting |
|---|---|---|---|---|---|---|
| `CircleCheck` | `Loader2` spinning | `TriangleAlert` | `CircleOff` | `CircleX` | `CircleDashed` | `Trash2` |

**Two things the change nearly broke, both caught by measuring.**

1. *Baseline.* The glyph shipped as `inline-flex items-center`, which takes its
   baseline from the first flex item — so the status word would have slid off
   the shared baseline the card is built on. Inline-block with `align-[-0.22em]`
   holds it: measured delta stayed at exactly **3px**, and the glyph sits
   **0.38px** off the text box's optical centre.
2. *Reduced motion.* The spinner is `motion-safe:`, so reduced-motion gets a
   still mark rather than no mark. The shape still reports "in flight".

**Four existing tests failed, and that was the system working.** Three asserted
`querySelector('svg')` was null; one asserted the baseline. They were not
obstacles — they were the previous decision defending itself, which is exactly
what a test for a design rule is for. Rewriting them meant stating the new rule
in the same place the old one lived.

The rule in §7 was also wrong as written. "A dot **or** a word" reads as "no
icon"; what it actually protects against is a second *colour channel*. Restated.

## The page title came down twice, and the second time it went below the content

Started at **16/24 weight 600**. Moved to **20/28 weight 500** because 16/600
put it on exactly the same rung as a stack card's own name — the page and the
cards inside it were tied. Now settled at **14/20 weight 500** on the board
(node `110:4023`).

The 20px version fixed the tie by making the header win. That was the wrong
winner. The title says *which section you are in* — the sidebar already said it
and the trail already said it, so a third statement in 20px is the loudest thing
on screen doing the least work. At 14/500 it sits one rung above the trail
(13/400), level with the one fact opposite it (`8 stacks`, 14/400), and the
whole title row reads as **one band of chrome**. A card's name at 16/600 now
clearly outranks it, which is correct: the cards are the page.

Worth noting the intermediate step was not wasted. It broke the tie, and the tie
had to break before it was obvious which direction was right.

## Dark mode's input fill was a well; the well was a theme inconsistency

Reported as *"the input dark-mode bg is different from the surface colour of
controls, but in light mode they're the same."*

The previous session had deliberately split them — `--input` one rung *below*
the card, `--control` one rung *above* — to make "a field is a well, a select is
a face" literal in dark. Light could not do that (nothing sits above white), so
light kept them equal at `#F7F6F3`.

The result was that **the theme changed the relationship, not just the values**:
a search field and the select beside it in the same toolbar row were one grey in
light and two greys in dark. A user flipping themes sees the layout re-group
itself.

Collapsed back to one fill in both themes (`--input: var(--secondary)`). The
well/face distinction is still there — it is carried by the **line** and the
**press mechanic**, the way light has always carried it. A field's hover moves
the border; a control's hover moves the fill. That was always the more reliable
signal anyway; a 3% lightness step between two adjacent 32px controls was not
doing the teaching it was credited with.

## The dark hairline was 40% louder than the light one, at the same alpha

Reported as *"in dark mode the border seems a bit more intense."* Correct, and
measurable.

The alphas had been copied across themes on the assumption that the same number
gives the same line. It does not — a white ink lifting a near-black ground is a
much bigger perceptual step than a dark ink dropping white:

| | Ink | Ground | ΔL (OKLCH) |
|---|---|---|---|
| Light `--border` 11% | `rgb(53 35 0)` | `#FFFFFF` | **0.072** |
| Dark `--border` 11% | `rgb(255 253 247)` | card `L 22.2` | **0.102** |

Solved each rung for the alpha whose ΔL off the dark card matches the light
rung's ΔL off the sheet. **WCAG contrast ratio agrees with OKLCH ΔL to within
0.005 at every rung** — two independent measures landing on the same numbers is
the check that the answer is not an artefact of the metric.

| Rung | Light | Dark | ΔL, both |
|---|---|---|---|
| `--border-subtle` | 6% | 4% | 0.039 |
| `--border` | 11% | **7.5%** | 0.071 |
| `--border-strong` | 18% | 13% | 0.120 |

**The general lesson, and it now applies to every alpha in the system:** an
alpha is not a theme-portable value. It is a *recipe* for a value, and the
recipe reads its ground. §4 already knew this for dots — a 1.5px disc needs
roughly double a line's alpha because it has less area to accumulate contrast,
and dark dots run firmer still. Lines needed the same treatment in the other
direction and had not been given it.

## The row rules came out, and the measurement made it a short argument

Asked as *"what if we remove the border completely from the list?"*

A separator earns its place by **grouping** — telling you the branch line belongs
to the name above it and not the name below. So that is the thing to measure,
rather than argue about taste:

| Gap | |
|---|---|
| Inside a row (name → branch) | **0px** |
| Between rows | **28px** |

Space was already doing 100% of the grouping. Eight hairlines down the page were
drawing a grid the data does not have, and the hover wash — which is what
actually tells you what you are about to click — had to compete with them.

Out. **Two lines stayed, for reasons that are not "it looks better":**

- The rule **under the column header** is not a row separator. It is the
  chrome/content boundary — labels above, data below — which is the same job
  §12a's header hairline does.
- The **skeleton** dropped its rules too. Its whole purpose is that nothing
  moves or appears when the data lands, so it has to match the loaded row.

**The condition for reversing it is written down**: if a compact row mode lands,
28px becomes ~8px, the grouping argument flips, and the rule earns its place
again. That is worth more than the decision — a rule with no stated failure
condition gets re-litigated every six months.

§11 was retitled from *"hairlines, not cards"* to *"space, not lines, and never
cards"*, because the old title now taught the wrong half of the idea.

## Dark mode landed on the board, and the flip found 162 hand-painted fills

The Shape + Hierarchy board had one variable mode — 39 colour variables, light
values only. Added a **Dark** mode and set all 39 from the product's dark tokens,
renaming the existing mode `Mode 1` → `Light` so the two are named, not implied.

The OKLCH grounds had to be resolved to sRGB for Figma, which is worth stating
because it is a one-way step: the board stores `#232220`, the code stores
`oklch(25.2% var(--ground-c) var(--ground-h))`. **The code is the source.** Turn
the `--ground-c` dial and the board does not follow — it has to be re-derived.

**The valuable part was not the mode, it was what the mode exposed.** Flipping to
dark rendered the board broken: card titles vanished, the brand lockup stayed
black on a near-black rail, every nav glyph held its light-mode grey. 67 paints
inside the stacks frame alone were **hand-painted hexes with no variable bound**,
and page-wide it was 162 nodes:

| Hex | Count | Should have been |
|---|---|---|
| `#191714` | 61 | `ink/primary` |
| `#5C574E` | 73 | `ink/fg-2` |
| `#FFFFFF` | 29 | `ink/on-primary` |

Every one of those hexes **is** the exact value of the token it should have been
bound to, so the mapping was mechanical rather than a judgement call. They had
been invisible for the whole redesign because in a single-mode file a hardcoded
hex and a bound variable render identically. **A second mode is the only thing
that can tell them apart** — which is an argument for adding dark early, before
the board grows, not after.

Fixed page-wide rather than on the one board, so the main components carry the
binding and future instances inherit it.

### Still open — effect styles are not mode-aware

`elevation/md` is an effect **style**, and a style holds one value. Dark's
shadows in code are a different set entirely (`rgba(0,0,0,0.45)` where light runs
`rgba(16,20,26,0.035)` — an order of magnitude apart, because a shadow on a
near-black ground has almost nothing to darken). The board will show light
shadows in dark mode until those are rebuilt as variables or as a second style.

## Shadows flip too — effect styles can't hold modes, but effects can hold variables

The gap left by the dark-mode pass. A Figma effect **style** holds one value, so
`elevation/md` would have shown light shadows in dark mode forever. The way out
is that an individual **effect** can bind variables — `color`, `radius`,
`spread`, `offsetX`, `offsetY` — and a *variable* is mode-aware. So the style
stays a style, and every field inside it points at a two-mode variable.

New `Elevation` collection, Light/Dark, 25 variables covering all 11 effect
styles: the four elevation rungs, the four press mechanics, the three focus
rings.

**Colour carries most of the difference, and the gap is enormous:**

| Rung | Light | Dark |
|---|---|---|
| `elevation/sm` | `#10141A` @ **3.5%** | black @ **45%** |
| `elevation/lg-1` | `#10141A` @ **11%** | black @ **70%** |

Roughly 13×. A shadow works by darkening what is behind it, and on a near-black
ground there is almost nothing left to darken — so dark also needs **more travel
and more blur**, not just more alpha (`lg` goes 20px/50 blur → 28px/64).

**Geometry variables were created only where the two themes actually differ** —
nine floats, all on `md` and `lg`. The press styles got colour bindings only,
because their geometry is deliberately identical across every variant and both
themes: one geometry, intensity is the only thing that moves.

### Two things this turned up that nobody was looking for

- **`line/ring` on the board still held `#191714`** — the value from before the
  focus ring became the palette blue. The board had drifted from the code
  silently, because a ring only paints on a focused node and nothing on the
  board was focused. Now `#3B6FE0` light / `#6E9BFF` dark.
- **`figma.createAutoLayout()` frames default to a white fill.** The first proof
  strip rendered as a white band swallowing every shadow. Layout containers need
  `fills = []` explicitly — worth remembering, it looks like a shadow bug and is
  not.

### The asymmetry is now written down (§3)

**OKLCH in the code, hex in Figma** — Jaseem's call, until Figma ships OKLCH.
Code stays the source; the board is a resolved snapshot that has to be
re-derived whenever a dial turns. The risk this creates is silent divergence,
which is exactly what `line/ring` had already done.

## The column header was the only thing on the page not square with itself

Jaseem fixed the spacing on the board first (node `121:885`) and asked for it in
code. Measuring the live header against the board found two numbers off, not one:

| | Board | Was in code |
|---|---|---|
| Above the label | **8px** | **16px** |
| Each side | 8px | 8px ✓ |
| Below the label | **8px** | **6px** |
| Between columns | 20px | 20px ✓ |
| Type | 11/16 weight 400 | ✓ |

So the header ran **16 / 8 / 6** — no two insets the same, on the one piece of
chrome the page has. Now 8 all round.

**The 16 above was not the header's own padding** — it was the sheet's content
inset (`px-4 py-4`) landing on it. That is why it read as a header problem and
was actually a *nesting* problem: content gets 16, and the column header is not
content. Fixed with `-mt-2` on the header, which gives back half the inset
locally rather than changing the page padding every other screen depends on.

Worth stating because it will recur: **when a spacing value looks wrong on a
component, check whether the component owns it.** Half the time it belongs to
the container, and "fixing" it on the component either does nothing or breaks
the sibling that was correct.

The card grid keeps the full 16 — the boards differ deliberately (`Frame 9` at
y=16 on the cards board, `column headers` at y=8 on the list board), because
cards *are* the content and a column label is not.

## Empty state and no-results, designed on the board

Two states, deliberately siblings rather than one component with a prop.

### The glyph is built from the product's own shapes

A stock magnifier says "search". **Ghosted list rows with a lens over them say
"your list came back empty"** — which is the actual message. Same for the empty
state: a deck of the product's own cards, the front one carrying a name line and
two service chips, is a glyph that could not belong to another app.

Both are assembled from existing colour variables (`surface/selected`,
`surface/pressed`, `line/hairline`, `surface/sheet`), so they flip with the
theme for free. Nothing is a flat hex and nothing is an imported asset.

### What separates the two

| | No results | Empty |
|---|---|---|
| Cause | Your filter excluded everything | You have not made one yet |
| Title | `title/600` — 16px | **`head/600` — 20px** |
| Copy | One line, teaches nothing | **Says what a stack IS** |
| Action | `flat/secondary` — "Clear filters" | `flat/primary` — "New stack" |

**The size difference is the signal.** A filter mistake is not the same moment
as first run, and the type rung says so before the words do. The empty state is
also the one place in the product where the core noun gets defined — a
no-results state has no business teaching, and this one has to.

### Two things that took three attempts

- **A white card on a white sheet has no silhouette.** The stack glyph fused
  into one mass twice. Separation had to be *drawn*: each card carries a 3px
  OUTSIDE stroke in `surface/sheet` — a gap, not a line — which is the standard
  stacked-card read and the only one that survives a theme flip, because the
  stroke is a variable and follows the surface.
- **Width steps of 24px read as shoulders**, turning the deck into a machine.
  At 8px the step is depth and the silhouette stays a rectangle.

### The box is gone

The current `EmptyState` primitive draws a **dashed bordered box**. Dropped it:
§11 already says the list is not boxed, and a bordered panel here
re-introduces exactly the frame that rule removes. The state sits on the sheet,
like the rows it replaces.

**This is a primitive change, not a page change** — `EmptyState` has seven call
sites (git integrations, domains, metrics, stacks ×2). Not implemented in code
yet; the board is the agreed reference first.

## The empty-state glyph, rebuilt from the website instead of invented

Rejected: *"I don't like the stacks visual, I don't like the transparency."*
Correct on both counts, and the second explains the first.

The ghost deck was three translucent rectangles. **Transparency was doing the
work that shape should have done** — the cards had no substance, so they read as
a smudge rather than as objects, which is why it took three attempts to make the
layering legible at all. No amount of tuning fixes a glyph whose whole idea is
"faint".

Rebuilt from `stackdome.com`, which had the answer already:

| Website | In the glyph |
|---|---|
| Solid white cards, hairline, soft elevation | Two opaque `surface/sheet` nodes with `line/hairline` and `elevation/sm` |
| Dot-grid canvas | 54 dots on the ladder, new `canvas/grid` + `canvas/grid-bold` variables |
| Orange dashed connector with an arrowhead | 22px dashed `brand/orange` wire, drawn arrowhead |
| Green status dot per node | `state/success` dot, said once |

**Nothing in it is transparent now**, and it is no longer a generic "empty box"
illustration — it is the product's own architecture canvas in miniature, which
is the one image Stackdome owns.

The grid tokens were missing from the board entirely; the code has had `--grid`
and `--grid-bold` all along. Added rather than hand-picked, so they follow the
theme like everything else.

### Figma mechanics that cost time

- **A `LINE`'s `strokeCap` applies to BOTH ends.** `ARROW_EQUILATERAL` gives a
  double-headed wire, which says "these talk to each other" rather than "this
  depends on that". One direction means flat caps plus a drawn head.
- **Rotating a `POLYGON` moves it about its own origin**, so the shape never
  lands where the coordinates say. An explicit `vectorPaths` triangle is
  deterministic and shorter.

### The two states now speak different visual languages, deliberately

No-results keeps the flat ghost bars — Jaseem picked that style from a reference
and likes it. Empty state is the solid architecture canvas. **They are no longer
siblings by construction, only by layout and type.** Worth revisiting if that
starts to read as inconsistency rather than as two different moments.

## Both states in code, and a second preview server to review the one you cannot reach

Decision on the two glyphs, and the reasoning is worth keeping: **the decorated
architecture glyph goes to the empty state because that is the first screen a
new user ever sees. A filter that matched nothing gets a 34px lens.** Decoration
is spent where it buys something.

### The primitive lost its box

`EmptyState` drew a dashed bordered panel. Removed — §11 already says a list is
not boxed, and the state sits on the sheet like the rows it replaces. The
`dashed` prop went with it (two call sites, both just passing `false`).

Type is deliberately quiet: `text-body` medium over `text-meta` muted. An empty
state is not a headline, and a page with nothing on it does not get to shout.

### The glyph settled at ONE card, not two

The two-card-plus-wire version drew a **closed system**. The final board has one
card with its wires sweeping off the canvas edges and fading out — which says
the stack connects to more than fits, and is the calmer image. Rejected mine
before it shipped; the geometry here is the board's (node `201:1177`).

Mostly divs, so every colour stays a real token and the elevation is the real
`shadow-md`. The two wires are **curves**, so they need real SVG paths — but the
gradient stops use `stopColor="var(--brand)"` rather than a hardcoded hex, and
they fade to **transparent brand** where the board fades to white. White is only
invisible in one of the two themes.

### The empty state is unreachable in the normal preview

`pnpm dev:mock` serves eight stacks. **No amount of clicking gets you to first
run**, so the state that matters most for a new user was the one nobody could
look at.

    pnpm dev:mock         → :5273  the review dataset, every status
    pnpm dev:mock:empty   → :5274  a brand-new org, nothing in it

Driven by `VITE_PREVIEW_SCENARIO`, not a URL param — the scenario has to be
chosen before the service worker boots. **Separate ports on purpose**, so the
two sit side by side in a browser instead of being toggled back and forth.

Only `/stacks` is blanked. A new org still has a project and a cluster; emptying
those would put a different screen on trial.

### One test defended the old copy, correctly

`stacks-page.stories.tsx` asserted `'No stacks deployed yet'`. Rewritten to
assert what the state now has to do: the definition of a stack is present, and
the action is offered **twice** — the header's and the empty state's own,
because on first run the centre of the page is where the eye is.

## `secondary` in Figma is `outline` in code, and the names actively mislead

*"What button did you use for no results, does not look secondary?"* — a real
mismatch, and the trap is worth naming because it will happen again.

| Board tone | Code variant | Renders |
|---|---|---|
| secondary | **`outline`** | `control` fill **+ hairline** |
| — | `secondary` | `control` fill, **no** hairline |

`secondary` drops the border deliberately — side by side in a story it and
`outline` read as one variant drawn twice. But `--control` is **2% off white**,
so on the sheet, without the hairline, there is no button there at all: it reads
as a soft blob rather than a control.

**Picking a variant by matching the board's layer name gives the wrong one.**
The rule is now in §9: when a Figma button shows a fill *and* a border, it is
`outline`.

## Closing the Stacks page — what the verification pass turned up

Signed off. Context and guidelines written to the files that own them:

| Went into | What |
|---|---|
| `DESIGN-PRODUCT.md` | §3 OKLCH-vs-hex, §4 the dark alpha ladder + the `subtle` exception, §9 the `secondary`/`outline` trap, §11 no row rules + header inset + empty-state spec, §12a the 14/20 page title |
| `frontend/CONTEXT.md` | `EmptyState`, `SearchGlyph`, `StackArchitectureGlyph`, **board units**, `SheetHeader`, preview scenario |
| `CLAUDE.md` | `dev:mock:empty` in the commands table, and why the scenario cannot be a URL param |
| `frontend/.gitignore` | Scratch harnesses by **pattern** — four had already escaped into `git status` |

### The sweep found a page that was already broken

Walking every sidebar destination for console errors — not just the ones this
work touched — `/object-stores` was dying in the router's error boundary on
`store.spec.configuration`.

**Pre-existing, and the cause is instructive.** Its preview fixture was a
hand-written `{ provider: 's3', bucket: … }` literal — fields the API does not
have. Every other fixture comes off a factory typed against the generated
schemas, which is precisely why none of them could drift. This one bypassed
that.

Fixed by **annotating the type**, not by guarding the component: a nil-check in
`providerLabel` would have hidden a wrong fixture rather than fixing it, and the
repo rules rule out defensive programming for exactly this reason. The
annotation immediately rejected my own first guess at `SecretReference`
(`{name, key}` — it is `{secret_id, key}`), which is the guarantee doing its job
within seconds of being restored.

**The lesson for the next fixture:** an untyped literal in a mock is not a
shortcut, it is a page that works until someone opens it.

### Verified, not assumed

`tsc -b` · `lint` (0 errors) · **1961 tests** · a real `vite build` · and a
browser sweep of all nine destinations plus both preview servers, listening for
`pageerror` and `console.error`. Backend untouched — `git status` over `pkg/`,
`cmd/`, `config/` is empty.

Next: the create-new-stack flow. Plan at
`docs/superpowers/plans/create-new-stack-redesign.md`.
