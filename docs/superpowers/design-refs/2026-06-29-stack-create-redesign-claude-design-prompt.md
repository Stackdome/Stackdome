# Claude Design prompt — Stackdome "Create Stack" redesign

> Copy everything below the line into Claude Design, and attach the screenshots noted in
> the **Attachments** section. This brief asks Claude Design to explore **three distinct
> directions** for redesigning Stackdome's stack-creation experience.

---

## Context

**Stackdome** is a self-hosted PaaS (Heroku/Render/Railway-like) that deploys and manages
workloads across Kubernetes clusters. Users create **Stacks** — a stack is a set of
**resources** (containers/services), plus **volumes**, **addons** (managed Postgres etc.),
and the **connections** between them (depends-on ordering, env vars sourced from secrets /
addons / sibling resources).

The mental model maps 1:1 to a **docker-compose file**: a stack = several services, each
with an image-or-git source, ports, volumes, env vars, and depends-on links.

### The problem we're solving

The current "Create New Stack" form is **too form-heavy and dense**. It's one long scroll
of accordions, each resource expanding into **three tabs** (Configuration / Deployment /
Environment) packed with fields. Specifically, ranked by how much they hurt:

1. **Blank-slate paralysis (biggest pain)** — starting "from scratch" gives an empty,
   intimidating form. The user has nothing to start from and doesn't know what to add.
2. **Density** — too many fields visible at once; accordions feel cramped.
3. **Signal/noise** — hard to tell required fields from advanced/rarely-used ones.

There are four entry points that all currently dump the user into this same form:
**from a template**, **from a docker-compose file**, **from a GitHub repo** (not built yet),
and **from scratch**. Templates and compose **prefill** the form; "from scratch" does not.

### What good looks like (from peer PaaS — borrow the *interaction model*, not visuals)

- **Render**: you pick a **service type first** (Web Service / Background Worker / Cron /
  Postgres / Key-Value / Static Site / Blueprint), then connect a repo and Render
  **auto-detects** build & start commands, landing you in a tidy form with sensible defaults.
- **Railway**: an empty **canvas**; you add services as **nodes** (from repo / database /
  docker image / template) and wire them together with **reference variables**.
- **Heroku**: the create form is tiny (name + region); complexity is pushed out to
  git-push buildpack auto-detection and a one-click **add-ons marketplace** for managed
  services like Postgres/Redis.

None of them ever show a blank dense form. That's the bar.

---

## Brand & visual constraints (non-negotiable)

Match the existing Stackdome design system shown in the attached **current-state**
screenshots:

- **Dark theme**, near-black backgrounds, subtle card borders/elevation.
- **Orange** primary accent (buttons, active states, highlights).
- **Uppercase, letter-spaced, monospace** section labels (e.g. `STACK RESOURCES · 2`).
- Card-based sections, generous spacing, restrained iconography.
- Status dots (green = ready, amber = pending), `· ` separated metadata lines.
- No raw colors outside this palette; keep typography on-scale. Treat the screenshots as
  the source of truth for tokens, spacing, and component style.

---

## The three directions to explore

Produce these as **three separate, contrasting directions**. For each, show the key
screens (entry → build → configure) and call out how it kills blank-slate paralysis.

### Direction 1 — "Quick win": restructure the form in place (no new on-ramp)

Keep today's form but **re-tier it by importance, not topic**.

- Replace the three per-resource tabs with **one view, two zones**:
  - **Essentials (always visible):** Resource Name, Build From (Image / Git) + the source
    field, Ports, and Environment Variables (promoted out of its buried tab; keep the
    "paste .env" / "import file" affordances).
  - **`▸ Advanced` (collapsed by default):** Depends-on, image pull secret, volume mounts,
    init command/args, run command/args. (The entire "Deployment" tab folds in here — it's
    rarely used; it only overrides the container ENTRYPOINT.)
- Resources render **collapsed** by default with a one-line summary
  (`● postgres  postgres:16 · :5432 · pgdata`); expand to edit.
- Prefilled fields get a subtle **✓ "already set"** affordance.
- Page-level: **Stack Info** and **Resources** are primary; **Volumes** and **Addons**
  auto-collapse when empty (they show a `· 0` count today).

This is the **shared destination form** that Directions 2 and 3's flows also hand off to —
so design it well; the others reuse it.

### Direction 2 — "Pragmatic" (recommended): Block Composer → the Direction-1 form

A guided on-ramp that removes blank-slate by letting users **compose from recognizable
building blocks**, then drops them into the Direction-1 form to fine-tune.

Flow:

1. **Step 0 — Start chooser.** Replace today's split ("New Stack" button + "Import"
   dropdown) with one screen: *How do you want to start?* — cards for
   **Build from blocks** (the new default), **From template**, **Docker compose**,
   **GitHub repo**, and **Blank / advanced** (escape hatch straight to the empty form).
2. **Step 1 — Block composer.** Title: *"What's in your stack?"* A palette of block cards
   grouped as **Services** (Web service, Custom/empty), **Data** (Postgres, MySQL, Redis,
   MongoDB — each shows it auto-adds a volume where relevant), **Jobs** (Worker, Cron).
   Clicking a card adds a **pre-scaffolded resource** to a live **"Your stack so far"**
   list. Known-software blocks come essentially fully filled (e.g. Postgres → image
   `postgres:16`, port 5432, a `pgdata` volume, a `POSTGRES_PASSWORD` env). Generic blocks
   pre-fill the *shape* (Web → a port slot; Worker → a command field; Cron → a schedule
   field). Show **auto-wired links** inline (e.g. picking Web + Postgres auto-adds a
   `DATABASE_URL` on Web pointing at Postgres) and let the user toggle/remove them.
3. **Step 2 — Configure.** Hand off to the **Direction-1 form**, now prefilled, resources
   collapsed, advanced tucked away, prefilled fields ✓.

Design the **start chooser**, the **composer (palette + "stack so far" + linkage display)**,
and the **handoff into the prefilled form**.

### Direction 3 — "Ambitious": Canvas / node-graph (start + layout fused)

A visual, Railway-style board where the architecture *is* the editor.

- Empty **canvas** with a **palette** (drag/drop resource nodes: web, database, worker,
  cron, custom — same block catalog as Direction 2).
- Nodes show summary chips (image, port, volume); **draw edges between nodes** to express
  depends-on and env/reference wiring (an edge from Web → Postgres offers to inject a
  connection env var).
- Selecting a node opens a **side/inspector panel** containing the Direction-1
  Essentials + Advanced fields for that node (the panel replaces the long-scroll form).
- Volumes/addons appear as their own node types or attach to a node.
- Show: empty canvas with palette, a populated canvas (Web + Postgres + Redis with edges),
  and the node inspector panel open.

---

## Cross-cutting affordances (apply across all three where they fit)

- **Auto-detect from source** — when a user points at a Git repo / Docker image, infer and
  prefill build command, start command, and exposed port (Render-style). This also powers
  the not-yet-built **GitHub repo** entry point.
- **Auto-wiring** — when two blocks obviously connect (app + database), offer the
  connection env var pre-filled, clearly shown and dismissible.
- **Escape hatch** — never trap power users; always offer a path straight to the full
  advanced form.

---

## Deliverable

Three distinct, clickable-feeling design directions (1, 2, 3 above), each with its key
screens, all rendered in the Stackdome dark/orange brand from the attached screenshots.
Where a direction reuses the restructured form (Directions 2 & 3 lean on Direction 1's
form), show that handoff explicitly. Call out, per direction, how it defeats the
blank-slate pain. Feel free to propose variations or a fourth direction if you see a
stronger idea.

---

## Attachments

**Pile 1 — Stackdome current state (match this brand; this is what we're fixing):**
- Empty "Create New Stack" form (Stack Info + empty Resources + empty Volumes) — *the
  blank-slate problem*.
- Resource accordion, **Configuration** tab (dense fields) — *the density problem*.
- Resource **Deployment** tab and **Environment** tab — *the 3-tab structure being
  collapsed*.
- Stacks list page (New Stack button + Import dropdown) — *entry-point context*.
- Stack Volumes spec — *volumes styling*.

**Pile 2 — Inspiration (borrow the interaction model, NOT the visuals; 2–3 max):**
- Render "create a new service — pick a type" screen → informs Direction 2's block cards.
- Render web-service config form (auto-detected build/start, sectioned) → informs the
  Direction-1 form layout.
- Railway project canvas with a few service nodes + a Postgres → informs Direction 3.
- (Optional) Heroku add-ons marketplace or a Railway template → informs the block catalog.
