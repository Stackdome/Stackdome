# DESIGN-PRODUCT.md

**The rules for the Stackdome product UI. This file is the only authority.**

Read it before changing anything visual. If a rule and the code disagree, the
code is wrong. If another document and this file disagree, **this file wins** —
and the other document gets corrected.

| Document | Role |
|---|---|
| **`DESIGN-PRODUCT.md`** (this file) | **Authority.** The rules |
| `docs/design/openai-platform-study.md` | **Evidence.** Measured findings. Never an instruction |
| `docs/design/redesign-log.md` | **History.** How we got here. Never an instruction |
| `stackdome-website/DESIGN.md`, `DESIGN-PROMPT.md` | The **website's** rules. They govern the website, not the product |

Working surface: `pnpm --prefix frontend storybook`.

---

## 1. Who this is for

Three people, from the product's own positioning:

| Who | What they want |
|---|---|
| **Platform team** — 1–3 people running infra | Developers self-serving without giving up control |
| **Developer** — the primary user | *"Doesn't want to learn infra, write YAML, or file a ticket to get a URL."* Push, deploy, get a URL |
| **CTO** | Not to pay PaaS prices, and not to hire a platform team |

**The primary user is deliberately not an infrastructure expert.** That is the
most useful fact in this file, and it settles more arguments than any token:

- **Comfortable over dense.** This is not a trading terminal. Room to read wins.
  A section may go dense where the work genuinely demands it — density is a
  default, not a cage.
- **Screens explain themselves.** Someone who avoided infra on purpose should
  not need documentation open beside the product.
- **The canvas is the headline.** *"Topology you can see beats YAML you can't"*
  is the product's central promise. It gets designed like the hero, not like a
  utility.

## 2. How we work

| Rule | |
|---|---|
| **Artifact → Figma → code** | An artifact to understand the flow and settle the solution. Figma to lay the foundation — usually not exhaustive, just enough to fix the direction, then expanded in code using these rules. When the picture isn't clear, Figma scopes it **fully** before any code |
| **Storybook is the working surface** | Build and judge here, on mocked data. Never hand-roll a copy of a component inside a story |
| **Judge against a screen** | A component judged in isolation has no basis to be judged |
| **When to make a new component** | Only when **no Storybook component exists**, or the existing one's look and feel **cannot be altered** to fit. Otherwise **add a variant** — `pill` and `flat` are variants of Button, not two components |
| **Done** | When we are happy with what is on the screen and the components on it |
| **Order of work** | Structure, nav and page shell first, revamping components as needed. Then page by page, journey by journey. Component changes rippling into other pages is expected and fine — we reach every page |
| **Fix the rule before the pages** | A page-by-page sweep cannot fix a rule; it can only spread it. If a change touches more than two screens, the rule lands here first |
| **Measure, don't eyeball** | Read computed styles in the browser. Two rounds of button padding were wrong because the glyph's safe area was assumed |
| **Research the mechanic before building it** | Copying a look without its mechanism produces the wrong thing convincingly |

---

## 3. Surfaces — white floats, grey recedes

The surface a thing sits on **is** a statement about its scope.

| Plane | What it holds |
|---|---|
| **Paper frame** — sidebar, page background | Navigation, global helpers, the account block |
| **White sheet** — content plane + its top bar | Only what is scoped to the page you are on |

| Rule | |
|---|---|
| Higher elevation = **whiter** | Modals, popovers and the content plane are white |
| Grey never means "a card" | Grey means **pushed back into the frame** |
| Grey wells are for **input** and **reference** | Search fields, code blocks. Never a list of things you click |
| **Grounds sit 2–4% below white** | Measured as **lightness drop**, not channel alpha — see below. Governs **grounds only**; washes are marks (§4) |

The sheet's top bar is part of the sheet, so it may hold page-scoped things —
and only those. See §12.

### How to measure a ground — lightness, never channel alpha

**Measure the L\* drop from white.** Do not express a warm grey as "% of the text
ink": text ink is near-neutral, so a warm ground's blue channel — which is doing
*warmth* work, not *depth* work — inflates the number and makes a correct grey
look too heavy. This is the same trap as the line ink in §4, one level up.

| Ground | Value | L\* drop | |
|---|---|---|---|
| `--secondary` | `#FAF9F6` | ~2.0% | ✓ |
| `--control` | `#F7F6F3` | ~3.1% | ✓ |
| `--background` — the paper frame | `#F5F4F1` | ~3.8% | ✓ |

All three are in band. `--muted` sits at ~5.9% and is **not a ground** — it is the
hover wash, and it becomes an alpha ink tint per §4.

### Grounds are solid, marks are alpha

This is the rule underneath the whole token layer.

> A **ground** is something you stand things on — the paper frame, the sheet, a
> control fill, the terminal. It is **opaque**, because it must be a known
> colour that everything else is measured against.
>
> A **mark** is laid over a ground it does not control — a hairline, a hover
> wash, a selection tint, a grid dot, a press recess, a scrim. It is **alpha**,
> because the same mark must work on three grounds and still read as one thing.

| | |
|---|---|
| **Solid** | `background` · `card` · `secondary` · `control` · `code-bg` · every ink tier · the state hues |
| **Alpha** | borders · hover and selection washes · grid dots · press shadows · state fills and borders · scrims · the focus ring |

**Consequence:** a hover wash is a *mark*, so it is not bound by the 2–4% ground
band. It is an alpha ink tint, the way `--sidebar-accent` already is.

### Dark grounds — one hue, one chroma

Light measures its grounds by **lightness drop from white**. Dark cannot: the
drops are tiny and the eye reads *colour* long before it reads level. So dark
grounds are governed by a second number.

> **Every ground in dark sits at chroma ~4, hue ~85 (OKLCH).** One warmth for
> the whole screen. Level is carried by lightness alone.

| Ground | Value | L | C |
|---|---|---|---|
| `--background` — the frame | `#131210` | 18.3 | 4.3 |
| `--card` — the sheet | `#1C1B19` | 22.2 | 4.1 |
| `--popover` | `#1E1D1B` | 23.2 | 4.1 |
| `--control` / `--input` — the control fill | `#232220` | 25.2 | 4.0 |
| `--muted` / `--accent` — hover | `#292826` | 27.8 | 4.2 |

The controls had escaped this. `--input`, `--control`, `--popover` and
`--accent` all ran at **chroma 10–12.5** — two to three times the ladder — so on
a neutral warm-grey screen every field, select and menu read **khaki**. They are
back on the ladder.

### A field and a control share one fill

`--input` and `--control` are the **same value in both themes.** Light has
always had them equal (`#F7F6F3`); dark briefly split them — input one rung
below the card, control one rung above — which meant a search field and the
select beside it in the same toolbar row were two greys in dark and one grey in
light. **The theme was changing the relationship, not just the values.**

| | Light | Dark |
|---|---|---|
| `--input` — you type into it | `#F7F6F3` | `#232220` |
| `--control` — you press it | `#F7F6F3` | `#232220` |

**A field is still a well and a select is still a face** — that difference is
carried by the **line** and the **press mechanic** (§9), the way light has
always carried it, not by a second fill. A field's hover moves the border
(`border-strong`); a control's hover moves the fill (`--control-hover`).

**Measure any new dark ground in OKLCH before it lands.** A hand-picked hex that
looks right on its own will carry the wrong warmth next to the ladder.

### OKLCH in the code, hex in Figma

**The product authors colour in OKLCH. Figma has no OKLCH, so the board stores
the resolved sRGB hex.** That is a deliberate asymmetry, not drift.

| | |
|---|---|
| **Code is the source** | OKLCH is what makes the ladder *tunable* — one chroma dial for every ground, lightness carrying the tier. Hex can only express the output, not the relationship |
| **Figma is a resolved snapshot** | One direction, code → board. Turning `--ground-c` or `--ink-c` does **not** propagate; the Figma variables must be re-derived |
| **Never downgrade the code to match** | Storing hex in `index.css` so the two look alike would throw away the mechanic to save a conversion step |

When adding a Figma colour variable, **compute** the sRGB from the OKLCH value —
never eyeball a hex — and record the OKLCH it came from. Revisit the whole
arrangement when Figma ships OKLCH; the conversion is the only reason the two
can ever disagree.

## 4. Alpha — lines, washes, dots, states

One line has to sit on the **white sheet**, the **paper frame** and a **control
fill** and still read as the same line. An opaque hex cannot: it is right on one
surface and wrong on the other two.

**The ink is not `--foreground`.** That was the trap. Text ink `#191714` is very
nearly neutral, so it composited to `#E6E5E5` and `#D6D5D5` — grey lines on a
warm palette, going *colder* as they got stronger. Figma's board had been
compensating by hand-picking a warmer solid, which is why the two never matched.

The line ink is **derived**: solve for the colour that, at the smallest alpha
which keeps every channel in range, composites to the target line on the sheet.

```
a   = max((bg − target) / bg)        per channel
ink = (target − bg × (1 − a)) / a
```

For the board's `#E8E6E2` hairline on white that gives **`rgb(53 35 0)`** at
11.37%, and the round trip is exact.

| Token | Alpha | On the sheet | For |
|---|---|---|---|
| `--border-subtle` | 6% | `#F3F2F0` | A line **inside** a control — the divider between segments |
| **`--border`** | **11%** | **`#E9E7E3`** | **The hairline.** Rows, cards, inputs, the default |
| `--border-strong` | 18% | `#DBD7D1` | Hover and emphasis |

Never hand-write a line colour. If a new rung is genuinely needed, derive it
from `--line-ink` and add it here.

### Alpha is a function of the shape, not the token's name

A 1px line and a 1.5px dot at the same alpha are **not** equally visible — a dot
has far less area to accumulate contrast, so at hairline alpha it disappears.
Grid dots therefore run at **16% / 24%**, roughly double the line. Any new mark
gets its alpha chosen for its own shape.

### Dark is not light inverted — and the alphas are not either

The ink flips to `rgb(255 253 247)`, already warm and needing no correction. But
**the alphas do not carry over.** The same 11% is a bigger step in dark: light's
hairline lifts white by **ΔL 0.072** in OKLCH, while 11% over the dark card
lifted it by **0.102** — about 40% louder. That is what read as "the borders are
more intense in dark".

Each rung is solved so its **ΔL off the dark card matches the light rung's ΔL
off the sheet.** WCAG contrast ratio agrees with OKLCH ΔL to within 0.005 at
every rung, so two independent measures pick the same numbers.

| Rung | Light | Dark | ΔL light | ΔL dark |
|---|---|---|---|---|
| `--border-subtle` | 6% | **6.5%** | 0.039 | 0.062 † |
| **`--border`** | **11%** | **7.5%** | **0.072** | **0.071** |
| `--border-strong` | 18% | **13%** | 0.119 | 0.120 |

Do not round the dark alphas back up to match light's numbers. **They differ
precisely so the line does not.** A new rung is solved the same way, not
guessed — the maths is a ~20-line script, not a judgement call.

† **`subtle` is the one deliberate exception.** Parity put it at 4%, and at that
value the divider inside a control — the only thing this rung draws — could not
be seen in dark. It is the shortest line in the system and it has a **control
fill on both sides** rather than open sheet, so it loses more to its
surroundings than a parity calculation on the card can predict. Raised to 6.5%
against the real control and held there.

**The exception is the rule doing its job, not a failure of it.** Parity gives
you the answer for a line in open field; a rung that draws somewhere else gets
checked in the place it actually lands. §4 already says this for dots — alpha is
a function of the *shape and its setting*, not of the token's name.

Dots follow the same logic in the other direction: white marks on a near-black
field separate *less* at a 1.5px disc, so they run firmer at **20% / 30%**.

**Dark mode is fixed as we go.** The commitment is that values live in tokens so
the theme stays flexible, not that every dark value is final today.

### Washes and selection are ink tints, never picked greys

A selected sidebar row is **6% of the ink**. This is what makes "selected" mean
one thing product-wide and survive a theme flip. A hand-picked grey needs a
second hand-picked grey for dark, and the two drift.

### The interaction ladder — one ink, four rungs

Every interactive surface — a nav row, a ghost button, a segment — uses the same
four rungs. Hover, selected and pressed once all shared the 6% tint, which made a
hovered row indistinguishable from the selected one and made pressing show
nothing at all.

| Rung | Light | Dark | Token |
|---|---|---|---|
| **hover** | 4% | 5% | `--wash-hover` |
| **selected** | 6% | 7% | `--wash-selected` |
| **selected + hover** | 9% | 10% | `--wash-selected-hover` |
| **pressed** | 12% | 13% | `--wash-pressed` |

**Hover sits below selected on purpose** — selection has to stay the stronger
signal. Dark runs a point firmer at every rung, for the reason above.

A button with its own face does not use the ladder: `primary` darkens its own ink
(`primary-hover` → `primary-press`), `secondary` lifts to `control-hover`. Only
faces that are *transparent at rest* borrow the wash.

**Do not express these as stacked Tailwind variants.** `hover:` and
`data-[active=true]:hover:` both match a selected row being hovered, and which
one wins is not reliably predictable. Branch on the state in the component and
emit one set of classes — `washes(isActive)` in `sidebar.tsx` is the pattern.

### A state colour is one hue at three alphas

Text is the opaque hue · fill ~10–12% · border ~40–50%. One hue, three jobs, so
a state can be a word, a tint or an edge without ever needing a second value.

**Using all three at once is the bug.** That is what makes a status pill say one
fact four times.

## 5. Elevation — four shadows, and content gets none of them

Content is **flat**. Shadow is what tells you something is floating *over* the
page, so anything that isn't floating must not have one — including the inset
white "raised" highlight, which is a bevel and reads as one.

There are **four** shadows. Tailwind offers eight names and the stylesheet used
to spell all eight out, with three holding one identical value and two more
holding another. A name that promises a choice it cannot deliver is worse than
no name: it reads as a decision someone made. The four are defined; `2xs`, `xs`,
`shadow` and `xl` alias them.

| Rung | For |
|---|---|
| `shadow-sm` | The one step between flat and an overlay |
| `shadow-md` | Small overlays |
| `shadow-lg` | Popovers, dropdowns, select panels |
| `shadow-2xl` | Modals and dialogs — the only thing allowed to float this far |

### The one raised piece of content: a selected segment

**"Content is flat" has exactly one exception, and it is a well, not a float.**
The selected segment of a segmented control carries `shadow-sm`.

It does not break the rule, because the rule is about *floating over the page*.
A selected segment is a **raised face inside a recessed track** — the same idea
as a key sitting proud of a keyboard — and its track's `overflow-hidden` clips
the shadow on three sides. What survives is a lift along the **divider** and
nowhere else.

| Measured off the board | |
|---|---|
| Divider, before | `#E2DFD8` — the hairline over the control fill |
| Divider, with the lift | **`#DEDAD3`** |
| The two pixels beyond it | one level darker, then nothing |

That is the entire effect. If it is doing more than this, it is wrong.

**Anything else that is content still gets no shadow.** This exception does not
generalise to rows, cards, chips or nav items.

Press is separate and **inset**: `--btn-press-soft` · `--btn-press-mid` ·
`--btn-press-strong`, chosen by how light the face is, never by how important
the button is (§9). Figma carries all of these as effect styles named to match.

### Focus is a shadow, and it is not elevation

**The focus ring is drawn as a `box-shadow`, never an `outline`.** It is the one
shadow on the list that does not mean "this is floating" — it means *the
keyboard is here* — so it is exempt from "content is flat" and it is not a fifth
elevation rung.

An outline paints one colour at one offset and cannot stack. `box-shadow` can —
which is what lets the ring **compose with the element's own elevation** (a
focused popover is `var(--focus-ring), var(--shadow-lg)`) and with its press
recess, and what lets it carry a gap when a gap is needed.

**The ring is a SOLID accent blue.** Not a tint. It is the one mark in the
product that has to be unmissable, and at full strength it can never be misread
as a border — which is the only thing the gap was ever protecting against.

### The ring is blue, and it also carries selection

It was solid **ink** until August 2026, and ink had two problems. A black ring on
a near-black button was invisible, so the gap had to do the entire job. And when
the create-stack tabs needed a *selection* mark, ink could not give one — the
page is already ink, so "selected" had nothing left to say it with.

One value now carries **selection and focus** product-wide:

| | |
|---|---|
| Code | `--ring` → `#3B6FE0` light · `#6E9BFF` dark |
| Figma | `accent/primary`, same hexes. **`line/ring` and `focus/ink` are aliases of it**, so there is one dial for every blue mark on the board |

**This is not a brand colour and not an action colour.** §7 still holds: black is
the only action colour, and orange still belongs to the website and to
thresholds. Blue is a **state** mark — *this one, and the keyboard is here* —
which is a different job from *press me*. A filled primary button stays ink.

**Consequence, still open:** the gap on dark faces was justified by ink being
invisible there. A blue ring is not. The gap may now be redundant on
`default` / `destructive` / `inverse` — worth testing before the next journey.

| Token | Value | |
|---|---|---|
| `--ring` | **`#3B6FE0`** · `#6E9BFF` in dark | Solid accent. The one mark that is not alpha |
| `--ring-width` | **2px** | |
| `--ring-offset` | **2px** | The gap |
| `--ring-offset-color` | `var(--card)` | The gap's fill. A **ground**, so solid |
| `--focus-ring-edge` | one stop, no gap | **The default.** `shadow-focus-edge` |
| `--focus-ring` | two stops, with the gap | `shadow-focus` |
| `--focus-ring-inset` | two stops, inward | `shadow-focus-inset` |

### Flush is the default; the gap is for dark faces only

| The element | Ring | |
|---|---|---|
| **Anything** — card, field, select, nav row, ghost or secondary button | **`shadow-focus-edge`** | The ring lands on the element's own edge |
| **A dark face** — `default`, `destructive`, `inverse` buttons | **`shadow-focus`** | Held from when the ring was ink and vanished here. The blue reads on its own, so the gap is now a refinement rather than the signal — see the open note above |
| **Flush to its container** — a full-bleed row, a segment in a track | **`shadow-focus-inset`** | An outside ring is clipped by the container and loses a side |

The gap used to be the default, and it was wrong twice over: it is painted in the
*surface* colour, so on a grey field it showed up as a **bright white band prying
the ring off the control**, and on a white card it did nothing at all.

| Rule | |
|---|---|
| **The gap follows the face, not taste** | Dark face → gap. Everything else → flush |
| **A flush ring replaces the hairline** | The ring lands directly outside the element's own 1px line and the two stack into a doubled edge. On focus the border goes **transparent** — it keeps its 1px so nothing shifts, it just stops being visible. An `aria-invalid` border is the exception: that line is reporting something |
| **Press must not eat the ring** | Press and focus are both `box-shadow`, so `active:` overwrites `focus-visible:` and a pressed button loses its ring mid-click. Every variant spells the combined state out — `focus-visible:not-disabled:active:` beats `active:` on specificity, never on declaration order |
| **The gap follows the surface, the geometry never moves** | On the paper frame, re-point `--ring-offset-color` (the rail carries `--sidebar-ring-offset-color`). Never change the width or the offset to compensate |
| **`focus-visible`, not `focus`** | A mouse click must not ring |
| **Never the brand colour** | `--ring` is the accent blue and nothing else may be. Orange stays on the website and on thresholds (§7) |

**In Figma** the three rings are effect styles — `focus/ring-edge` (one drop
shadow, spread 2), `focus/ring` (spread 2 in the surface colour, then spread 4 in
the ring ink) and `focus/ring-inset`. A drop shadow does **not render on a
`COMPONENT` or an `INSTANCE`** node, only on a plain `FRAME`, so each focus
variant carries a child frame matching the host exactly — same size, same fill,
same hairline — whose only job is to cast. Give the host `clipsContent = false`
or the ring is trimmed off.

## 6. Type — named by job, anchored on 13px

Use the token, never `text-[13px]` or `text-sm`.

| Token | Size / line | Job |
|---|---|---|
| `text-label` | 11 / 16 | Group labels, avatar initials |
| `text-meta` | 12 / 16 | Row data — branch, counts, status, timestamps |
| **`text-body`** | **13 / 20** | **The base.** Nav, buttons, breadcrumbs, prose, inputs |
| `text-name` | 14 / 20 | The thing you scan a list for, and **the page title** |
| `text-title` | 16 / 24 | Section titles, and a card's own name |
| `text-head` | 20 / 28 | Dialog and empty-state headlines |

Every line-height is a multiple of 4. **Three weights only: 400 / 500 / 600.**

**20px is the ceiling for now** — not forever. If a screen genuinely demands
larger type we add a rung here deliberately, rather than reaching for an
arbitrary value at the call site.

### The page title is 14/20 at weight 500 — it is a label, not a headline

It has been 16/24 at 600 and then 20/28 at 500. Both made the sheet header the
loudest thing on screen, above content it only introduces. Settled at `name/500`
on the Shape + Hierarchy board (node `110:4030`).

| | |
|---|---|
| **The title is chrome** | It says which section you are in. The sidebar already said it, the trail already said it — it does not need to be announced a third time in 20px |
| **One rung above the trail** | 14 against the trail's 13, and weight 500 against 400. That gap is enough to read as "you are here" |
| **The content is louder than the frame** | A stack card's own name at 16/600 now clearly outranks the header. Correct: the cards are the page |
| **Same rung as the one fact** | "8 stacks" opposite it is `name/400`. Same size, weight separates them — the title row reads as one band of chrome |

**Tracking is the token's, not the board's.** `--text-name` carries the scale's
own tracking; the board sits at 0. Held the token rather than fragment the scale
for a sub-pixel difference across one word.

Before adding a size, ask whether **weight**, **colour** or **position** can do
the job. They were already carrying most of the hierarchy.

Before adding a size, ask whether **weight**, **colour** or **position** can do
the job. They were already carrying most of the hierarchy.

### Typeface

**Geist** for the interface. **JetBrains Mono** for code, strings, IDs and
machine values. Both settled.

`font-mono` means **a machine produced this and a machine will read it back** —
IDs, keys, hashes, branch names, env-var names, code, JSON, log lines.

**Never for a human label.** A role called `Developer`, a status word like
`Ready`, a plan name — those are words, and mono makes them look like values the
user is not allowed to change.

## 7. Colour — three text tiers, one job each

| Token | Job |
|---|---|
| `text-foreground` `#191714` | What you came to find — names, **and every nav label** |
| `text-fg-2` `#5C574E` | Data you read — project, branch, counts, status |
| `text-fg-muted` `#726C63` | Furniture and time — group labels, ages, separators |
| `text-fg-ghost` | Placeholder and disabled only |

| Rule | |
|---|---|
| **If a colour doesn't report something, it's a bug** | No decorative colour |
| **Black is the only action colour** | Both the reference study and our own audit found no brand colour anywhere in a product of this kind. A filled primary is ink |
| **Blue is a STATE, not an action** | `--ring` marks *selection and focus* — the live tab, the focused field (§5). It never says *press me*, so it never fills a button. One value, one job |
| **Orange is not in the product** | It belongs to the website and to **thresholds** — sign-in, 404. Never an action, never a chip, never a label tint on a working surface. The canvas wire is the one load-bearing exception, and it reports connection |
| Nav labels are **ink at rest** | The grey is carried by the **icon**, never the word |
| Status says it **once** | One *colour channel*, not four — never a coloured word plus a dot plus a fill plus a border. A **per-state glyph** is not a repeat; see below |
| All tiers pass AA at 12px | `fg-muted` was darkened for exactly this reason |

### A glyph earns its place by making a distinction the word cannot

"Says it once" was read for a while as "no icon on a status," and both stack
views shipped with icons switched off. That was the right call for the set they
had and the wrong rule to draw from it.

The set was **per family** — three glyphs for `ready` / `pending` / `error` — so
`Degraded`, `Unavailable` and `Failed` all drew one triangle. That is a mark
that adds a symbol without adding a fact, and the distinction it flattened is
exactly the one that changes what you do next.

| | |
|---|---|
| **A glyph per STATE** | Earns its place. `CircleCheck` · `TriangleAlert` · `CircleX` · `CircleOff` · `CircleDashed` · `Trash2` · a spinning `Loader2` for in-flight |
| **A glyph per FAMILY** | Does not. It repeats the colour and flattens the word |
| The glyph is **derived**, never passed | Same rule as the colour — an icon that disagrees with its word has to be unbuildable, not merely discouraged |
| One tone | The glyph inherits the word's colour. It is not a second channel |
| **Inline, never inline-flex** | An inline-flex box takes its baseline from its first flex item, so the glyph drags the word off a shared baseline (§8). Inline-block plus an optical nudge holds it |
| In-flight is the only thing that **moves** | And it is `motion-safe:`, so reduced motion still gets the mark — just still |

What "says it once" still forbids is unchanged: a coloured word **plus** a dot
**plus** a fill **plus** a border. One channel, however many marks it takes to
be specific.

## 8. Geometry

### Radius scales with the size of the element

| Token | Value | For |
|---|---|---|
| `rounded-sm` | 6px | Chips, badges, icon hit-areas |
| `rounded-md` | **8px** | **Default** — list rows, menu items, nav items, inputs |
| `rounded-lg` | 12px | Cards, panels, the content sheet |
| `rounded-xl` | 16px | Modals, dialogs, sheets |

**Radius is a function of HEIGHT.** Anything the same height takes the same
radius — a 32px button, a 32px input and a 32px select sit in a row together,
and if their corners disagree the row reads as three unrelated things.

| Control height | Radius | Token |
|---|---|---|
| **20px** | 6px | `rounded-sm` — **chips only**, see below |
| 28px | 6px | `rounded-sm` |
| **32px** | **8px** | **`rounded-md`** — the default step |
| 40px | 12px | `rounded-lg` |

**20px is a rung for chips, and only chips.** A chip is not interactive — the
row or the card around it is the target — so it is not bound by a hit-target
minimum, and 28px chips would eat a fifth of a card. Nothing you can click gets
this rung.

**Inputs match their buttons.** A 40px field pairs with a 40px button. (The
reference runs inputs taller than buttons; we deliberately do not.)

`rounded-full` is **not a rung on this ladder.** It is the Button's `pill`
variant declaring itself — and §9 is guidance about *which button to reach for*,
not a rule about radius.

### Control heights — 28 / 32 / 40

| Height | For |
|---|---|
| 28px | Chips, in-row actions, anything inside a dense row |
| **32px** | **Default** — page toolbars, dialog footers, sidebar rows |
| 40px | Form fields and their primary button |

**Height follows density, never importance.** An important button gets
**filled**, not taller. Chrome: topnav and sidebar header are 52px.

## 9. Buttons

| | |
|---|---|
| Variants | `default · destructive · outline · secondary · ghost · link · inverse` |
| Shape | `flat` (**the default**) · `pill` |
| Filled is **rare** | One `default` per screen. Everything else is a hairline or a ghost |
| At rest | **Flat.** No highlight, no bevel, no drop shadow |

### The board's "secondary" is code's `outline` — read this before reaching for a variant

**`secondary` and `outline` are the same fill. Only `outline` has the hairline.**
`secondary` drops it deliberately, because side by side in a story the two read
as one variant drawn twice.

| Board tone | Code variant | Renders |
|---|---|---|
| primary | `default` | Ink fill, inverse label |
| **secondary** | **`outline`** | **`control` fill + `border` hairline** |
| — | `secondary` | `control` fill, **no** hairline |
| ghost | `ghost` | Transparent until hover |

**The name in Figma does not map to the name in code**, and picking `secondary`
off the board's label gives you an edgeless blob on the sheet — `control` is
only 2% off white, so without the hairline there is no button there at all.
When a Figma button shows a fill *and* a border, it is `outline`.

### Shape says what kind of action this is

This is **guidance on which button to reach for**, not a geometry rule.

| Shape | Means | Where |
|---|---|---|
| **`flat`** | *This is a working control* | Toolbars, dialog footers, row actions, filters, `Save`, `Cancel` — the bulk of every screen |
| **`pill`** | *This commits. It finishes a flow or ships something* | Auth `Continue`, the action in an empty state, `Deploy`, `Publish`, a wizard's final step |

**At most one pill per screen.** None is often correct — most screens are work,
not commitment. `flat` defers to the height ladder in §8; `pill` is the one
variant that overrides it.

**Enforcement.** The component default is `flat`, so new work is right without
thinking about it. Existing screens are corrected **as we work each journey** —
deliberate pills are restored screen by screen, never retrofitted in a sweep.

### Disabled

**Dimmed, plus the not-allowed cursor.** This is the rule for **every disabled
control in the product**, not just buttons. The dim and the cursor together are
the whole signal.

The control must still receive the pointer for the cursor to show, so
`pointer-events-none` is wrong — the native `disabled` attribute already blocks
the click.

`loading` is **not** `disabled`: contrast stays at 100%, the content swaps for a
spinner, and `loadingText` says **what is happening** — `"Deploying…"` — not that
something is.

### Press

Depth exists **only while you are touching it**, and what ships today is final:
the face recesses, label and icon travel **1px down**, the fill darkens. One
geometry for every variant — 1px side walls, 2px top shadow, 2px blur — with
intensity set by how light the face is, never by importance (§5).

### Optical padding

The eye aligns on **centre of mass**, not bounding boxes. A 14px glyph carries
~2.9px of invisible safe area each side; a letterform meets the edge with almost
none.

```
icon side = base − 3      label side = base + 2
```

| Size | Base | Icon side | Label side |
|---|---|---|---|
| sm | 10 | 7 | 12 |
| default | 12 | 9 | 14 |
| lg | 15 | 12 | 17 |

Symmetric where there is nothing to correct against: text-only, icon-only, and
icons on both sides.

## 10. Destructive actions escalate with the blast radius

One `destructive` variant applied to everything trains people to click through
it. **The friction has to be proportional to the damage, or it stops being read.**

| Level | Gate | Use for |
|---|---|---|
| **1 — Confirm** | Red button in a dialog, live immediately | Reversible or cheap: leave without saving, remove a row you can re-add |
| **2 — Acknowledge** | A **checkbox** must be ticked before the red button goes live | Destroys something rebuildable: an addon, a preview env, a secret |
| **3 — Retype** | The user must **type the resource name** | Destroys something with dependents or data: a stack, a cluster, a project |

| Rule | |
|---|---|
| Position | Red button **last**, after `Cancel` |
| Treatment | Red **fill**. Not a red outline, not red text |
| Before the gate is met | Rendered **disabled, not hidden** — the cost must be visible before it is payable |
| Shape | `flat`. Destroying is work, not a commitment |
| Never in the sheet header | The bar is on screen the entire time you scroll |

Say what will break, in plain words — *"All requests using this key will start
failing"* — not *"This action cannot be undone."*

## 11. Lists — space, not lines, and never cards

No box, no shadow, no card per item. **The list is not boxed either** — a border
around the whole table is the same mistake at a larger scale.

### There is no rule between rows

**A separator has to earn its place by GROUPING**, and in a 64px row it groups
nothing. Measured on the stacks list:

| Gap | |
|---|---|
| **Inside** a row — name → branch line | **0px** |
| **Between** rows | **28px** |

Space was already doing all of the work. The rule was decorating a boundary that
was never ambiguous, and eight of them stacked down the page read as a grid the
data does not have.

| What stays | |
|---|---|
| **The rule under the column header** | Not a separator — it is the chrome/content boundary. Labels above, data below |
| **The hover wash** | Row extent on approach was always its job, and it reads better with nothing competing |
| **The sheet edge** | The list's only outer boundary |

**Bring the rule back if a compact row mode lands.** At ~8px between rows instead
of 28 the grouping argument reverses, and then the line is doing real work. The
condition is the density, not the taste.

Row actions appear on **hover**; a kebab on every row at rest is chrome
competing with content.

### Columns are labelled

| Rule | |
|---|---|
| Every column gets a header | An unlabelled column makes the reader infer what `2 services · 1 vol` is |
| Headers are **sentence case**, `text-label` (11/16, weight 400), `fg-muted` | Size and colour already separate the header from the data |
| One 1px rule under the header | Nothing else |

**The header sits in a uniform 8px inset — 8 above, 8 each side, 8 below.**

| | |
|---|---|
| Above | **8px**, not the sheet's 16px content inset. A column header is **chrome**; it sits tighter to the band than content does |
| Sides | **8px**, the same inset the rows carry, so label and data share one left edge |
| Below | **8px**, then the rule, then the first row with **no gap** |
| Between columns | **20px** |

It ran 16 above and 6 below — the one piece of chrome on the page was the only
thing not square with itself. Settled on the board (node `121:885`).

### No uppercase

**Do not set type in all caps** — not for column headers, not for section
labels, not for eyebrows. Caps strip the word-shape the eye reads by and measure
wider for the same information.

**The only exceptions** are acronyms genuinely uppercase in the world — `API`,
`URL`, `CPU`, `TLS`, `JSON`, `ID` — and machine values that are literally
uppercase, like an env-var name. Never `text-transform`; if it is uppercase, it
is uppercase because the word is.

### Status is said once, and this is where it gets broken

A coloured dot at the left of the row **and** the status word in a column is
saying it twice. **Pick the column.** The dot survives only where there is no
room for a word.

### Cards are a second view, never a replacement

| View | Answers | Use when |
|---|---|---|
| **List** (default) | *Which one, and how do they compare?* | Scanning many, sorting, finding the odd one out |
| **Cards** | *What is going on with each one?* | Fewer items, more per item |

**Both views show the same rows, the same filters and the same sort.** The
toggle lives in the content toolbar, right side, last: a segmented control,
icon-only, two options. Persistence is per page, per user.

| Property | Value |
|---|---|
| Fill | **White** — same as the sheet |
| Border | 1px hairline |
| Radius | `rounded-lg` 12px |
| Shadow | **None.** Content is flat |
| Hover | The hairline strengthens. No lift, no scale |

### How the segmented control is built

Both the track and the selected segment carry a hairline, and **the track has no
padding.** That is the whole mechanic.

A bordered segment *inset* inside a bordered track puts two edges 2px apart, and
at that distance the eye reads a doubled line rather than a raised surface. The
fix is not to drop the segment's border — it is to drop the track's **padding**.
The segment runs flush, its outer edges land on the same pixel as the track's,
and the only line you see between them is the divider. `overflow-hidden` trims
the segment's square corners to the track's radius, so a selected segment's
corners *are* the track's corners.

The divider belongs to the **selected** segment, drawn as its own left/right
edge. A separate rule would eventually sit next to that edge and double it.

| Property | Value |
|---|---|
| Track | `--control` + hairline, radius by height (§8), **no padding** |
| Segment | The sheet's own white when selected |
| **Selected segment — lift** | **`shadow-sm`.** The one raised piece of content in the product — see §5. Clipped by the track on three sides, so it reads along the divider only |
| **Selected segment — radius** | **The track's rung, on the outer corners it owns.** The clip would round them anyway; the segment carries its own so the *shadow* follows the curve instead of cutting a square corner |
| Selection | **Ink vs `fg-muted`**, never opacity — a dimmed icon reads as disabled |

The track is **not** a grey well. With a hairline around it the control already
reads as a control, and a darker fill under the white segment makes the
unselected side look switched off.

### Toolbars — many ghosts, exactly one fill

| Rule | |
|---|---|
| Secondary actions | `ghost` — icon + label, **no border, no fill** |
| Primary | **One** filled button, last |
| Filters, sorts, searches | `flat`. Working controls — never a pill, never `outline` |
| Disabled secondary | Label and icon grey together. **No box appears** |

Eight actions with eight boxes is a wall with no primary.

## 12. The shell — frame, sidebar and sheet

**Source of truth: the `app shell` board in Figma.** Every number below is taken
from it. If code and the board disagree, the board wins and the code is fixed.

| Property | Value |
|---|---|
| Sidebar — expanded | **240px** |
| Sidebar — collapsed | **56px** |
| Gap, rail → sheet | **2px** |
| Sheet inset | **12px** on the other three edges |
| **Brand band** — the sidebar's lockup block | **64px**, lockup 32px centred, inset **16px** left |
| **Sheet content inset** | **16px** — the same edge the header uses |
| **Brand lockup ↔ page title** | **Same centreline.** See below |
| Sheet radius | **12px** (`rounded-lg`) |
| Sheet fill | `surface/sheet` — white |
| Sheet edge | 1px `border-subtle`, **as an `outline`** — see below |
| Sheet elevation | **`shadow-md`** = the board's `elevation/md` |
| Frame fill | `surface/frame` |
| Divider between sidebar and sheet | **None.** The sheet is a card and carries its own edge |

### The sheet's edge is an outline, never a border

A `border` participates in layout. The sheet is specified at an exact size and
its header row has to share a centreline with the sidebar's lockup — a 1px
border pushes that row down by 1px and breaks both.

**Use `outline`** (Figma: *stroke align — outside*), or a `box-shadow` ring. It
paints outside the box, follows the radius, and costs the layout nothing.

Two consequences worth knowing:

- **The 2px gap to the rail is load-bearing, not decorative.** An outline paints
  *outside* the box, so with the two columns flush its left edge lands under the
  `fixed` sidebar and is clipped away — the card loses one of its four sides.
- **The edge is `border-subtle` (6%), not the 11% hairline.** Once the sheet
  carries a shadow, the shadow does the separating and the line only has to
  describe the shape.

### Collapsing the rail is choreographed, not switched

The rail animates its width over 200ms. Anything inside it that arrives or
leaves travels with that width — `display` cannot be transitioned, so anything
that has to vacate space animates `max-width` or `max-height` to zero and is
clipped by its own overflow. The utilities are `rail-x` / `rail-x-in` /
`rail-y` / `rail-y-in` in `index.css`.

**The two directions are deliberately not symmetric:**

| | |
|---|---|
| **Leaving** | Immediate, ~100ms. The column is already closing on the label; holding it there just means watching it get guillotined |
| **Arriving** | ~90–110ms delay, then ~140ms. Text faded in early lands in a column too narrow to hold it and reflows while you are reading it |

Two traps, both of which put the collapsed glyphs off the 28px centre column:
**a flex `gap` still reserves space around zero-width children**, and **`ml-auto`
absorbs the free space** the centring needed.

### The two planes line up across the seam

The brand lockup and the page title **share a centreline**. That is the entire
reason the sidebar head and the sheet header are related at all — get it wrong
and the two columns read as two unrelated screens.

Both columns carry the shell's 12px gutter and then **16px** of their own
padding, so both 32px rows centre on **44**. The sidebar is `fixed` and ignores the
wrapper's padding, so it pays its gutter internally instead.

```
title  row centre = 12 gutter + 16 padding + 16 = 44
lockup row centre = 28 lead-in             + 16 = 44
```

**The seam carries no border.** That is not only §3 — it is also what makes this
arithmetic work. A 1px border on the sheet pushed its row down by exactly the
1px the two columns were out by.

Verified in the browser, not eyeballed. If the gutter or the header padding
changes, this number changes with it.

### The sidebar

| Group | Holds |
|---|---|
| *(no label)* | **Stacks · Previews** — these are the product; a label above the first item is furniture |
| **Platform** | **Addons · Secrets · Object Stores** — what you attach to a stack |
| **Infrastructure** | **Clusters · Domains · Git Integrations · Image Registries** — what an admin configures once. This is already the admin-gated set |

Grouped by **who touches it** (§1), not by what kind of object it is. Footer holds
**Appearance** and the account block (name over organisation, with a switcher) —
global helpers live in the frame, never on the sheet.

Nav labels are **ink at rest**; the grey is carried by the icon (§7). Selected is
`surface/selected` — a 6% ink tint (§4), never a picked grey.

## 12a. The sheet header — two parts

**The top of the sheet is the header.** Not a bar sitting on it: no divider,
content passes under it and dissolves into it. There is exactly **one** per
screen, and it has **two parts**:

| Part | Job | Holds |
|---|---|---|
| **1 — the title row** | *Identifies the section* | The collapse toggle, the page title, the one fact, the page's actions |
| **2 — the toolbar row** | *The tools for that section* | Search, filters, sort, view toggle — whatever the section needs |

**The band used to carry no divider at all**, on the reasoning that the sheet's
own edge is the boundary (§3). That was reversed on the board: at 108px the band
is tall enough that content sliding under it needed a stated edge. The fade went
with it — the two mechanisms answer the same question and cancel each other.

**Part 2 is conditional.** A page with nothing to filter renders the title row
alone. Both are the same component; the second row appears when the page gives it
something.

| Property | Single row | Double row |
|---|---|---|
| Band height | **64px** | **108px** |
| Row height | 32px | 32px each |
| Padding | 16px top and bottom | 16px top, **12px between rows**, 16px bottom |
| Horizontal inset | **16px** both ends | **16px** both ends |

**The outer padding is 16 and the row gap is 12, on purpose.** They are
different measurements doing different jobs: 16 is the sheet's own margin, 12 is
the distance between two rows of controls. Making them equal is what made the
band read as cluttered.

| Property | Value |
|---|---|
| Background | **Opaque** `surface/sheet` — content scrolls **under** the bar |
| Divider | **1px `--border` hairline along the bottom**, full sheet width — not inset to the padding |
| Fade under the bar | **None.** A dissolve under a crisp line is two answers to the same question. Content is **cut** by the hairline as it scrolls under |

### The dissolve — parked, not deleted

The header used to carry no divider and instead let content **dissolve** into it:
8px of the bar's own colour fading to transparent on its underside, so a row
sliding under was never sliced. It lost to the hairline here, and only here —
the two mechanisms answer the same question and cancel each other.

**It is a good technique and it keeps.** Reach for it wherever content scrolls
under something that should not announce an edge — a canvas, a log tail, a
metrics pane, the top of a long dialog body.

```html
<!-- directly under the bar; `-mb-2` cancels its own height so it costs no flow -->
<div aria-hidden class="pointer-events-none -mb-2 h-2 bg-gradient-to-b from-card to-transparent" />
```

| Use | |
|---|---|
| **Dissolve** | The boundary should be felt, not seen. Nothing below is a separate region |
| **Hairline** | The boundary is structural — chrome above, content below |

Never both.
| Position | `sticky top-0`, inside the scroll container |

**Title row — left**

| Element | Spec |
|---|---|
| Collapse toggle | 32×32 icon button, 16px glyph, `fg-2` — chrome, not content |
| Toggle → title | **6px** |
| **Page title** | **`name/500`** — 14px medium, ink |
| Trail, when nested | 13px weight 400 `fg-2`, `/` separator at `fg-2`/50% |

**Title row — right**

| Element | Spec |
|---|---|
| The one fact | **`name/400`** 14px, `fg-muted`, **tabular numbers** |
| Fact → first action | **12px** |
| Actions | 32px on the control ladder |
| Between actions | **8px** |
| Kebab | 32×32 icon button, always last |

**Toolbar row**

| Element | Spec |
|---|---|
| Search field | **300px**, 32px tall |
| Right cluster | Filters, sort, then the view toggle — **6px** between controls |
| Everything here | `flat` (§9). Working controls — never a pill, never `outline` |

Content is illustrative: a section brings whatever tools it needs. The **slots and
the spacing** are the rule, not the specific controls.

### Budget — the title row only

**One fact · one primary · one secondary · one kebab.** Everything past that goes
in the kebab. The toolbar row has no budget — it holds what the section needs —
but §11's rule still binds it: **many ghosts, exactly one fill.**

### What the header never holds

| Not this | Why | Where it goes |
|---|---|---|
| Docs, theme, account, product-wide search | About the **product**, not the page | The paper frame |
| Explanatory sentences | You need the explanation when there is **nothing yet** | The empty state |
| Entity metadata — repo URL, created date | Reference, not orientation | The page body |
| Back links **on a nested page** (`← Previews`, `← All addons`) | The breadcrumb **is** the way up | Deleted — but see journeys below |
| Destructive actions | The bar is on screen the entire time you scroll | The kebab |

**Filters, sort and search are no longer on this list.** They are about the page,
and they survive the scroll — they pass both tests below. They belong in part 2.

### Two tests before anything goes in

| Test | |
|---|---|
| **Does it survive the scroll?** | A count that updates with the filter passes. Anything tied to a scroll position fails |
| **Is it about the page, or the product?** | About the product → it belongs in the grey frame |

### Scaling to a new page

| Page type | The one fact | Actions |
|---|---|---|
| **List** — Stacks, Secrets, Domains, Users | Item count (`20 stacks`) | `+ New <thing>` |
| **Detail** — a stack, an addon, a preview config | Status | One primary verb, rest in the kebab |
| **Form / wizard** | — | The wizard footer owns its buttons |
| **Empty or errored** | — | The recovery action lives in the empty state |

**An empty right side is correct**, not unfinished.

**There is no `Home` crumb.** Every top-level destination is one click away in
the sidebar. A top-level page shows its title alone; the trail appears only once
you are actually nested. Consequence: **any new page must be reachable at a named
path** — a bare `/` route has nothing to title itself with.

### A journey has an exit; a nested page has a trail

**The header once banned back links outright. That was the rule written flat** —
it assumed every header belongs to a nested *place*, where the crumb genuinely is
the way up. A **journey** is not a place. It is a task you launched from a main
screen and will either finish or abandon, and the thing it needs is an exit.

| Page type | Header left | The test |
|---|---|---|
| **Journey** — `New stack`, `New cluster` | **Back arrow, then the title alone. No trail** | You are *leaving a task* |
| **Nested page** — `Stacks / acme-web / Environment` | **Trail, no back arrow** | You are *climbing a hierarchy* |

**Why the trail goes on a journey.** It would be the third way back. The sidebar
is always on screen with `Stacks` highlighted, so `Stacks /` repeats a permanent
nav item and pushes the page title into third position. The two marks are not
doing the same job, and only one job was unmet:

| | Says | Already covered by |
|---|---|---|
| Crumb | *where you are* | The sidebar, permanently |
| Back arrow | **this is a task you can leave** | Nothing else |

A crumb is a location; an arrow is an exit. On a wizard people look for the exit,
and a crumb is not read as one.

| Rule | |
|---|---|
| **Order** | Collapse toggle · **hairline divider** · back · title. The divider separates chrome that belongs to the **shell** from chrome that belongs to **this journey** |
| Divider | 1px `border`, **20px tall**, centred in the 32px row |
| Back | 32×32 icon button, 16px `ArrowLeft`, `fg-2` — same rung as the collapse toggle |
| **Where it goes** | **Back through history**, so it returns you the way you came. Deep-linked with no history, it falls back to the journey's origin |
| Not a journey | A nested page keeps the trail and gets **no** arrow. Never both |

**The cost, stated:** on the 56px collapsed rail (§12b) the word `Stacks` is not
visible — only the highlighted icon. Accepted; the arrow carries the exit.

## 12b. Widths — desktop only

**We are not optimising for mobile.** The target is desktop screen sizes; phone
and tablet layouts are out of scope until we say otherwise. Do not spend effort
on sub-1024 behaviour, and do not let it constrain a desktop decision.

| Width | Sidebar | Sheet |
|---|---|---|
| **≥ 1280** | Expanded, **240px** | Fills the remainder, 12px inset |
| **1024 – 1280** | Collapses to the **56px** rail | Fills the remainder — the rail buys the content 184px |
| **< 1024** | Out of scope | Out of scope |

The collapse is also **manual** at any width — the toggle in the title row (§12a)
is the user's, not just the breakpoint's.

## 13. Overlays

Shadow belongs to things that float (§5); content is flat. A menu opening from a
32px row does not cast a dialog's shadow.

**Reach for a drawer first, a dialog second.** Drawers hold work you return
from; dialogs interrupt. Refined per journey as the UI demands.

## 14. Motion

**Best practices for now; decided properly as we go.** Today only the button
press has a rule behind it (§9). Until motion is settled:

- Respect `prefers-reduced-motion` on everything.
- Keep durations short and consistent with the 150ms transition already in use.
- No ambient movement on a working surface.

---

## 15. Open — decided as we go

Not omissions. Each is settled when a journey needs it, and lands in this file
when it is.

| Open | Notes |
|---|---|
| Loading and skeletons | Define per journey |
| Failure and error loudness | Per journey, by severity |
| Toasts | Best practices for now |
| Multi-step / wizard chrome | To be decided |
| The canvas, logs, metrics, the deployment timeline | Define as we design each |
| Split detail layout | Add if a journey demands it |
| Content column at very wide widths | Figma specifies 1440. Whether the sheet's content caps or spans at 2560 is unanswered |
| Hover, focus and pressed states for the shell | Figma carries rest states only — these get derived in code from §4 and §9 |
| The list/cards toggle | Whether a second view earns its keep |
| The stack editor conversion | To be scheduled |
| The auth screens | To be decided — they currently follow the website's rules |

---

## Gotchas that cost real time

| Gotcha | |
|---|---|
| **Tailwind silently drops `shadow-[…]` behind a variant** | `focus-visible:shadow-[var(--x)]` is read as a shadow **colour**, not a shadow, so the geometry vanishes and it renders transparent. The `shadow-[shadow:var(--x)]` type hint fixes the ambiguity but is then not extracted at all in some builds — the class lands on the element, `:focus-visible` matches, the variable resolves, and **no rule is ever generated**. It worked in the dev server and failed in the Storybook build. The focus rings are therefore plain unlayered CSS classes, not utilities |
| **Unlayered CSS beats `@layer utilities`** | Which is what stops `active:shadow-*` overwriting the focus ring and eating it mid-press. Anything that must survive a utility goes outside the layers |
| **A `border-b` participates in layout** | It pushed the sheet header from 108 to 109 and moved every row below it. Figma's stroke is align=INSIDE; the CSS equivalent is `box-shadow: inset 0 -1px 0` (`.sheet-edge-b`), which costs the layout nothing — the same reason the sheet's own edge is an `outline` (§12a) |
| **tailwind-merge and named sizes** | It classifies any unfamiliar `text-*` as a **colour**, so `cn("text-body","text-fg-2")` silently drops the size and the element falls back to 16px. `cn` registers the scale as a `font-size` group in `frontend/src/lib/utils.ts`. **Any new named utility that shadows a Tailwind prefix must be registered there too** |
| **Tailwind misses new files** | A file created while the dev server is running is skipped by the scan — arbitrary classes produce no CSS at all. `touch src/index.css` to force a rescan |
| **`:active` fires while `:hover` is true** | A press that only sets a shadow will still inherit the hover fill and get *lighter*. Set the pressed fill explicitly |
| **Text nodes are not elements** | They cannot be transformed or matched by `:first-child`. The Button wraps its children for exactly this reason |
| **`max-height: 0` cannot shrink a box below its own padding** | A collapsed group label kept 12+4px in the flow and pushed every group below it out of rhythm. Collapse the padding with the height |
| **A flex `gap` still reserves space around zero-width children** | And `ml-auto` absorbs the free space centring needs. Both knocked collapsed glyphs off the icon column |
| **shadcn's sidebar hides a collapsed group label with `-mt-8`** | It stays in flow and is yanked upward. Stacking your own collapse on top of that double-counts. Removed — collapsing the label is the consumer's job |

## Gotchas — writing to Figma

| Gotcha | |
|---|---|
| **A bound variable does not drive the render; the raw paint colour does** | `setBoundVariableForPaint` on a paint built from some placeholder hex leaves that hex rendering, while the data reads correctly as `var:surface/control`. **Build the paint from the variable's own value, then bind.** This is why a screenshot can disagree with an inspection |
| **A `DROP_SHADOW` follows the node's alpha** | So a transparent `ghost` variant casts no focus ring at all. Use a 2px OUTSIDE stroke where the node has no other stroke, and the shadow-ring only where a hairline must survive (a node has one `strokes` array) |
| **Component sets are often auto-layout** | Explicit `x`/`y` on variants is silently ignored, and resizing the set **stretches any FILL-sized child** — it grew the Account block's variants from 216 to 320. Set `layoutMode = 'NONE'` before authoring a variant grid |
| **Figma has no stroke offset** | `outline-offset: 2px` cannot be drawn. Ring it at spread/weight 2 and note that the offset lives in code |
| **A drop shadow does not reliably render on a `COMPONENT` or `INSTANCE` node** | Two side-by-side probes had it paint on a plain `FRAME` and not at all on either; a later variant contradicted that, so treat it as unreliable rather than impossible — either way, do not depend on it. Phase 1 drew every focus ring as a spread-2 shadow, so **12 focus variants were invisible on the board and nobody could see why**. The fix is a **child frame** matching the host exactly — same size, radius, fill and hairline — whose only job is to cast. Give the host `clipsContent = false` or the ring is trimmed |
| **Figma paints the LAST effect on top — the reverse of CSS** | A two-stop ring authored in CSS order (`gap, ring`) draws the spread-4 ring **over its own gap**. On a light face you cannot see the difference; on a **dark** one it reads as no ring at all, which is exactly where the gap is the only signal. Author `[ring, gap]` in Figma and `gap, ring` in CSS |
| **`node.effects` returns a fresh array every read** | So `effects.map(e => e === effects[0] ? … : …)` never matches — the identity check compares against a new clone. Author the whole array; do not patch it by reference |
| **An attribute selector cannot hold a space** | `query('TEXT[name=last change]')` silently matches **nothing** — no error, no result. Rename the layer without spaces, or walk `children` |
| **`visible = false` inside auto-layout removes the node from the flow** | So a sibling set to FILL swallows its width and the row reflows. Where the slot must stay reserved — a hover-only action, for instance — use `opacity = 0` |
| **Clone before you clear** | `const t = kids[0]; kids.forEach(k => k.remove()); t.clone()` throws *"node does not exist"*. Take the clones first, then empty the container |
