# Stack Deployments Tab — Design (Releases-backed)

> Status: **DRAFT — awaiting user review**
> Date: 2026-06-21
> Area: Frontend (Stacks detail page) on the new Releases backend (#100). Plus a short list of backend follow-ups.
> Supersedes: `2026-06-11-stack-activity-tab-design.md` (the parked Activity-tab design, built on the pre-Releases data model). Mockup: `docs/superpowers/design-refs/deployments-tab-mockups.html`.

---

## 1. Problem & context

When a deploy does not reach Ready, the UI cannot tell the user **what stage failed, which resource, why, or what to do next** — and there is no deploy history, rollback, or audit trail.

Two things changed since the parked Activity-tab work:

1. **The Releases backend shipped (#100).** A deploy is now an immutable, versioned **release**: `POST /releases` renders a manifest, applies it, and drives convergence. Releases give us numbered history, rollback, cause/audit, per-resource outcome, and pinned artifacts — a real deploy timeline. This obsoletes the parked "builds-as-timeline" synthesis.
2. **Deploy is now explicit and decoupled from Save.** `PUT /stacks/{id}` only persists the spec (intent); `POST /releases` deploys. The frontend currently has **no** release wiring, so the UI cannot deploy at all — wiring deploy is load-bearing in this work.

This design replaces the parked **Activity** tab with a **Deployments** tab (Railway-style): a current-deployment surface + release history + per-release detail, with first-class failure visibility layered on top.

## 2. Goals / Non-goals

**Goals**
- Rename/replace the stack-detail **Activity → Deployments** tab, backed by Releases.
- Surface the **current deployment** (active release + live stack status) with a stage tracker and per-resource state.
- Show **release history** (`GET /releases`) with state, cause (deploy vs rollback vs webhook), derived duration, git SHA, and a state-aware ⋮ menu.
- **Deploy / Rollback / Cancel** wired to the Releases API.
- **Failure visibility**: what stage failed, which resource(s), why — live (the instant a resource fails) and durably (post-mortem on a past failed release).
- **Per-release detail drawer**: lifecycle, resource outcomes, pinned artifacts + pin-diff, and a config-change diff vs the previous release.
- Reuse the existing diff/draft machinery (`stack-diff.ts`, `StickyActionBar`, `DirtyField`, the resource accordion).

**Non-goals (this iteration)**
- A build-log stream or a builds-history page (no backend endpoint / no route exists).
- `created_by` → name/email resolution (skipped; show no attribution for v1).
- Webhook-push deploys (backend `cause` exists; no trigger UI yet).
- Topology/metrics changes.

## 3. Scope tiers

- **Tier 1** — Current deployment block (active release + live `stack.status`) + history list (read-only) + failure visibility (live). Deploy button.
- **Tier 2** — Detail drawer (lifecycle, outcomes, pins, pin-diff, config-change diff) + Rollback + Cancel + post-mortem view.
- **Tier 3** — Drift / "Unreleased changes" banner (saved≠deployed) reusing the draft diff; polling lifecycle refinements.

## 4. Architecture

Two-zone tab (same shape as Railway's Deployments and the parked Activity), plus a right-side drawer:

```
Deployments tab
├─ Current deployment   ← active release (GET /releases[0]) + live stack.status (GET /stacks/{id})
│   ├─ headline + StatusPill + sequence
│   ├─ stage tracker: Build → Deploy → Ready
│   ├─ live per-resource progress (status.resources[] — StackResourceSummary)
│   └─ failure surface — joins spec.stack_resources[].status.last_failure
│        (StackResourceStatus, a DIFFERENT array, by resource name) — click-through accordion
├─ History              ← GET /releases (newest first; list excludes heavy JSONB)
│   └─ release rows (state, cause, derived duration, git SHA, ⋮)
└─ Detail drawer        ← GET /releases/{id} (full: outcome, pins, snapshot)
    ├─ lifecycle timeline (created→rendered→completed)
    ├─ resource outcomes table
    ├─ pinned artifacts + pin-diff vs previous
    └─ config changes vs previous (stack-diff on snapshots)
```

**Data sources (authoritative):**

| Need | Source |
|---|---|
| Release history rows | `GET /releases` → `StackReleaseList.items` |
| Active/current release | `items[0]` (highest sequence) |
| Full release (drawer) | `GET /releases/{id}` |
| Live per-resource progress | `GET /stacks/{id}` → `status.resources[]` (`StackResourceSummary`: phase, replicas, message) |
| Convergence → release link | `status.last_converged.{release_id, revision, at}` |
| Structured failure detail (live) | `GET /stacks/{id}` → `spec.stack_resources[].status.last_failure` (`type` build_failure\|runtime_crash) |
| Deploy / Rollback / Cancel | `POST /releases`, `POST /releases {from_release_id}`, `POST /releases/{id}/cancel` |

## 5. Data availability — list vs detail (audited against the spec + pgstore)

The release **list** query (`ListByStackID`) selects only:
`id, stack_id, sequence, state, message, cause, snapshot_revision, manifest_revision, pins, renderer_version, created_by, created_at, updated_at, rendered_at, completed_at`.

- **In the list:** state, cause(kind+detail), **pins**, snapshot_revision, manifest_revision, message, all timestamps.
- **NOT in the list:** `outcome`, `snapshot`, `manifest` (heavy JSONB, excluded for poll-safety).
- **In the detail** (`GetByID` = `First(&release)`, all columns): everything incl. `outcome`, `snapshot`, `manifest`.

Consequences (data-honest rules):
- History row **duration** is **derived** from `completed_at − rendered_at` (outcome.duration is detail-only).
- History row "all resources Ready" / per-resource outcome is **drawer-only**.
- History row **git SHA** comes from `pins` (present in list).
- **Pin-diff** can compare two releases' `pins` from the list (no extra fetch); **config-change diff** uses the two releases' `snapshot` (detail fetch).

## 6. Surfaces

### 6.1 Current deployment block
Headline + `StatusPill` + `#sequence` from the active release; meta = derived duration + short `snapshot_revision` ("config <hash>").

**Stage tracker — `Build → Deploy → Ready`** (user-chosen, friendly). Derivation from real fields:
- **Build** — gate on contract fields, **not** free-form phase strings. `StackResourceSummary.phase`/`message` are unstructured (no enum guarantees a "Building" value), so do not key the tracker on a phase substring. Shown only when the stack has build-spec resources (resource has a build spec / `git_sha` pins present). Build is **failed** on `last_failure.type=build_failure` (structured, reliable); otherwise treat Build as in-progress until the resource leaves the not-available state. Image-only stacks (no build spec) skip Build and start at Deploy. *(If a phase substring must be matched for an "active" shimmer, treat it as best-effort cosmetic only — never as state truth.)*
- **Deploy** — release `InProgress` with `rendered_at` set (render+apply done, rolling out).
- **Ready** — `status.last_converged.release_id == activeRelease.id` (or release `Released`).
- **Pre-cluster errors** (render/apply) fail at the **first node (Build ✕)** with the release-level message card disambiguating — see 6.3. *(Open nuance — see §13.)*

**Live per-resource table** from `status.resources[]` (`StackResourceSummary`): dot + name + phase + replicas (`available/replicas`) + message (e.g. "not available: ImageBuildInProgress"). **Two arrays, joined by resource name:** progress/phase/replicas come from `status.resources[]` (summaries — these have **no** `last_failure`); structured failure comes from `spec.stack_resources[].status.last_failure` (`StackResourceStatus`). The view-model joins them on `name`. An implementer must not look for `last_failure` on the summary array — it is not there.

**Recovered / last-error (even when currently healthy).** `StackResourceStatus` exposes `state` and `last_failure` as **independent** fields, so a resource can be `Ready` *and* still carry its `last_failure`. When a resource is currently Ready but `last_failure` is present, show a quiet **"recovered — last failure X, N restarts (~time)"** note (warn-tinted, not an alarm) with a "View last error" affordance that opens the same failure card read-only. Persistence: runtime `last_failure` survives via K8s `lastState.terminated` until the next termination; build `last_failure` clears on a successful rebuild (`imagebuild controller`). Timing ≈ `conditions[].last_transition_time` / `last_restart_request_processed_at`. This is distinct from the post-mortem (§6.5): the recovered note reflects the **current live resource**; the post-mortem reflects a **past release**.

### 6.2 History list
`GET /releases`, newest first. Row: state badge (6 states), `#sequence`, **cause** (`manual` / `rollback to #N` from `cause.detail` / `webhook_push` — deploy vs rollback always discernible), sub-line (Released → `git <sha> · <derived duration>`; Failed → `message`; Superseded → neutral copy), timestamp, **⋮ menu (state-aware)**:
- Released → **View details**, **Rollback to this** (`POST /releases {from_release_id}`)
- Pending → **Cancel** (`POST /releases/{id}/cancel`)
- Others → **View details**

No `created_by` column (resolution skipped). No "hide superseded" toggle (rare; render Superseded greyed inline).

### 6.3 Failure visibility (the core)

**Two error shapes**, by where the failure occurs:

1. **Resource-level** (build / runtime) — from live `spec.stack_resources[].status.last_failure`:
   - `type=build_failure` → BUILD badge; `.build` detail (reason, message, exit, restarts). Same field/path as runtime, written by `imagebuild controller`; cleared on recovery.
   - `type=runtime_crash` → RUNTIME badge; `.container`/`.init_container` detail + **log snapshot (best-effort)** (`/resources/{name}/logs?follow=false&tail=N`) + deep-link to Logs. **The log snapshot is best-effort and often empty for a crashing pod** — per-resource log streaming refuses non-Ready resources (#98, §12.5). The structured `last_failure` (reason/exit/restarts) is the reliable signal; render the log block only when non-empty, never as a guaranteed surface.
   - **Click-through accordion** when multiple resources fail — one open at a time (reuses the Configuration-tab accordion). Scales to N failures without a wall of cards.

2. **Release-level** (render / apply) — from `release.message` (set fast by `failRelease()`), a plain actionable message card, no per-resource detail (`outcome` is null).

**Load-bearing principle — show failure from LIVE status, not from `release.state`.** Build/runtime errors leave the release `InProgress` until the **15-min converge timeout**, then fail with a generic `"timed out…"`. The real error is in live `last_failure` within seconds. So the current-deployment block is failure-aware while the release pill still reads "Deploying" — a broken build never looks healthy for 15 minutes.

### 6.4 Detail drawer (`GET /releases/{id}`)
Right-side Radix Sheet over the history list. Sections:
- **Lifecycle timeline** — created → rendered → completed.
- **Resource outcomes** — `outcome.resources[]` (phase, replicas, message).
- **Pinned artifacts + pin-diff vs previous** — `pins.resources[]` (git_sha → commit link only on a known Git host; image_digest; volume_hash); changed pins highlighted.
- **Config changes vs previous** — reuse `stack-diff.ts` to diff this release's `snapshot` against the previous release's; read-only old→new rows in the `DirtyField` visual language (amber edge, ADDED tags).

### 6.5 Post-mortem (Failed release drawer)
Opened from history after recovery / a later deploy. Shows the **durable** record:
- **Why it failed** — `release.message` (durable).
- **Resource outcomes at failure** — `outcome.resources[]` phase + summary message (durable).
- **Failure detail [PROPOSED]** — structured `last_failure` (exit/restarts/k8s message). **Not durable today** (live-only, clears on recovery); becomes durable with the backend snapshot in §12. Log snapshot is not re-fetchable post-mortem (pod gone).

## 7. Deploy flow — Save and Deploy decoupled

- **Save** = `PUT /stacks/{id}` (persist spec). On `origin/main`, `PUT` **already** persists-only — `stackService.UpdateStack` has no `CreateRelease`/reconcile call; the only release touch is on Delete. **But the existing `StickyActionBar` primary button is currently labelled "Deploy" (`detail/index.tsx`) while it only calls `updateStack` (PUT) — a live mislabel.** This work relabels that button **"Deploy" → "Save"**; there is no "deploy" behaviour to remove from PUT. Comparison: draft vs saved.
- **Deploy** = `POST /releases` — explicit, from the Deployments tab Deploy button or the **"Unreleased changes" drift banner**. Comparison: saved spec vs the active release's snapshot.
- Same component + diff engine for both banners; two phases: **Draft** (unsaved → Save) → **Unreleased** (saved≠deployed → Deploy).
- **Rollback** = `POST /releases {from_release_id}` (re-deploys that release's snapshot+pins). **Cancel** = `POST /releases/{id}/cancel` (Pending only; disabled once InProgress).

**Drift signal** = heuristic: a `stack-diff` between the saved spec and the active release's snapshot (or `updated_at` vs release time), labelled "approximate". Precise drift needs the backend to expose the current `snapshot_revision` (§12).

## 8. Polling lifecycle
- Poll `GET /releases` + `GET /stacks/{id}` every **5s while the active release is non-terminal** (`Pending`/`InProgress`) or any resource is non-Ready.
- Back off / stop when `Released`; keep a short forced window after a Deploy click (status-transition lag).
- In-flight guard; stop on unmount.

## 9. Error taxonomy → surface (complete)

| Error class | Release.state behaviour | Surface | Mockup state |
|---|---|---|---|
| Build failure | InProgress → Failed @15m (generic) | live `last_failure` build card (BUILD), shown immediately | 4 |
| Runtime crash | InProgress → Failed @15m (generic) | live `last_failure` runtime card (reliable) + log snapshot (best-effort, may be empty — #98); click-through accordion if many | 5 |
| Render / config error | Failed fast (real message) | release-level message card | 6 |
| Apply error | Failed fast (real message) | release-level message card | 6 |
| Convergence timeout | Failed @15m | message + outcome table (which resource never Ready) | 8 |
| Missing-secret stuck | stuck InProgress, **no signal** | **no UI** (filed backend gap — not faked) | — |
| Post-mortem (any) | terminal | drawer: message + outcome (durable); structured detail [proposed] | 8 |

## 10. Component inventory

**Reuse (existing, unchanged):**
- `pages/stacks/lib/stack-diff.ts` — diff engine for the config-change drawer section and the drift banner.
- `components/shared/dirty-field.tsx` (`DirtyField`, `hideReset`) — read-only diff rows visual language.
- `StickyActionBar` — Draft (Save) + Unreleased (Deploy) banners.
- `pages/stacks/hooks/use-stack-edit-session.ts` — baseline↔draft session backing the Draft comparison (note the path is under `pages/stacks/hooks/`, not top-level `hooks/`).
- Resource accordion pattern — the failing-resources click-through.
- `hooks/use-current-user.ts` — (not used for attribution in v1; available later).

**Salvage from the parked Activity branch (`feat/stack-activity-tab`):**
- `components/branded/`: `stage-tracker`, `failure-card`, `stage-badge`, `alert-banner`, `log-snapshot`, `event-row`.
- Log deep-link plumbing: `api/observability.ts` `fetchLogSnapshot` + `buildStackResourceLogStreamUrl`; the Logs `LogsPreset` wiring.

**New:**
- `api/releases.ts` — `getStackReleases`, `getStackRelease`, `createRelease`, `cancelRelease`.
- `pages/stacks/components/detail/deployments/` — `deployments-tab.tsx`, `current-deployment-card.tsx`, `release-history.tsx`, `release-row.tsx`, `release-detail-drawer.tsx`, `failing-resources-accordion.tsx`, `use-releases.ts` (fetch + poll), `derive.ts` (view-model: stage derivation, headline, failure composition), `release-diff.ts` (snapshot diff via `stack-diff.ts`), `tests/`.

## 11. API wiring & types
- Regenerate frontend types/zod from the updated spec: `pnpm --prefix frontend generate:openapi-types` + `generate:openapi-zod` (StackRelease, ReleaseOutcome, ReleasePins, ResourcePins, ResourceOutcome, ReleaseCause, StackResourceSummary, StackConvergenceRecord are absent from `openapi.d.ts` today).
- New client fns in `api/releases.ts` against the 4 endpoints under `/stacks/{id}/releases`.

## 12. Backend follow-ups (file as issues)
1. **Fast-fail on terminal resource failure** — `converge.go` should detect a resource `last_failure` (build_failure/runtime_crash) and mark the release `Failed` with that reason, instead of waiting the 15-min timeout with a generic message.
2. **Snapshot `last_failure` into `release.outcome`** at fail time. Note `buildOutcome` (converge.go) currently walks the **summary** array `status.resources[]` (`StackResourceSummary`), which carries `phase`/`message`/`replicas` but **no** structured `last_failure`. To make the post-mortem durable, also read each resource's `last_failure` from the **other** array — `spec.stack_resources[].status.last_failure` (`StackResourceStatus`) — and copy it into the outcome. (Same two-array distinction as §6.1.)
3. **Missing-secret stuck `InProgress`** — apply requeues forever with no surfaced reason; should fail (or surface a precondition message) instead of hanging silently.
4. **Expose current `snapshot_revision`** on the stack response so drift detection is precise (not a heuristic).
5. **(Existing) #98** — per-resource log streaming refuses non-Ready resources, so crash-pod log snapshots are often empty.
6. **(Spec hygiene)** — `last_validation_run` is absent from the OpenAPI spec; with Releases, validation failures surface as `Failed` releases, so the parked validation-stage synthesis is dropped.

## 13. Resolved decisions & open nuance

**Locked:** Full Deployments tab · fresh branch off `origin/main` salvaging the branded primitives · decoupled Save/Deploy · stage labels `Build→Deploy→Ready` · detail in a side drawer · no `created_by` · click-through failure accordion · backend-data-only (no faked surfaces).

**Open nuance — stage labels vs pre-cluster errors.** The real pipeline is render→apply→build→run, but the friendly tracker is `Build→Deploy→Ready`, so render/apply errors (which precede build) don't have a natural node. Current resolution: mark pre-cluster errors as **Build ✕** ("failed at the first stage") with the message card disambiguating. Acceptable for v1; alternative is an order-honest `Prepare→Build→Ready`. Decide during implementation against real transitions.

## 14. Out of scope
Build-log streaming; a builds-history page; `created_by` resolution; webhook-push trigger UI; topology/metrics.
