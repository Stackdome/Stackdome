# Deploy Timeline — Deployments Tab Redesign

**Status:** Design / spec
**Date:** 2026-06-21
**Branch:** `feat/stateful-deployments`
**Supersedes the view layer of:** `2026-06-21-stack-deployments-tab-design.md` (card + history + Sheet drawer). The data contract, the two-array principle, and `derive.ts` carry forward unchanged.
**Design source:** Claude Design project `b58cdd1a-…` → `Deploy Timeline.dc.html` (imported via claude_design MCP). Mockup screenshots: 12-scenario `Tweaks` switcher.

---

## 1. Problem & intent

The first Deployments tab shipped as three stacked blocks: a current-deployment card, a flat history list, and a Sheet drawer for per-release detail. The redesign reworks this into a single **vertical deploy timeline (rail)**: the active release is a large node at the top (live status), and every earlier release is a rail row that **expands inline** into its post-mortem (why it failed · resource outcomes · config diff vs the previous release). The Sheet drawer is removed — there is one interaction model (expand in place).

Goals:
- One coherent timeline, current → history, top to bottom.
- Per-release **changelog/post-mortem inline** (no drawer): outcomes table + resource-grouped config diff.
- Cover all **12 scenarios** faithfully (see §7).
- Be ready for **infinite-scroll pagination** later; today render all + a client-side "Show more" window (backend has no pagination — §6).
- Honor the Stackdome brand: `index.css` tokens + `branded/`/`ui/` primitives, **no raw hex** (the design's `--amber`/`--ok`/`--err` are mapped — §8).
- Remove the throwaway demo harness + screenshots.

Non-goals: backend pagination (deferred, filed as a follow-up issue); a Logs deep-link beyond switching to the Logs tab; redesigning the page header/tabs/theme (the real app already owns those — the design's `topBar` is design-tool chrome).

---

## 2. Architecture

```
DeploymentsTab                      (container — unchanged props/wiring in detail/index.tsx)
  useReleases()  → releases (newest-first), activeRelease, loading, error, refetch  [REUSE]
  deploy / rollback / cancel via run() + toast                                       [REUSE]
  drift heuristic (stack.updated_at vs activeRelease.completed_at)                    [REUSE]
  └─ TimelineRail
       ├─ DriftBanner | ReleaseErrorBanner            (top, conditional — reuse AlertBanner)
       ├─ CurrentReleaseNode            (big rail dot; pulse when active non-terminal)
       │    ├─ StatusPill + #seq + title + meta(config/took/elapsed)   [reuse StatusPill]
       │    ├─ StageTracker {build,deploy,ready}                         [reuse branded]
       │    ├─ ResourceRow[]   (from stack.status.resources + two-array failure join)
       │    │     └─ expand (failing only) → FailureCard + LogSnapshot + "Open in Logs →"
       │    ├─ RecoveredNote                                            (amber)
       │    ├─ ReleaseErrorBlock     (render/apply message; FIXES the release-level-error gap)
       │    └─ "View changelog" → ReleasePostMortem (active release's own diff/outcomes)
       ├─ "Earlier releases" rail marker
       ├─ HistoryRow[]          (releases excluding active; click row → toggle post-mortem)
       │    ├─ ReleaseMenu (⋮): View details · Rollback to this · Cancel · Copy release ID
       │    └─ ReleasePostMortem (lazy)  → WhyItFailed · OutcomesTable · ConfigDiff vs prev
       └─ ShowMore (client window)  |  EmptyState "No deployments yet"
```

**Reused as-is:** `use-releases.ts`, `derive.ts` (deriveStages, deriveFailingResources, deriveRecovered, releaseGitSha, causeLabel, formatDuration, humanizeFailureType), branded `StatusPill`/`StageTracker`/`FailureCard`/`LogSnapshot`/`AlertBanner`/`EmptyState`/`Panel`, `ui/dropdown-menu`.

**New logic:** `release-snapshot-diff.ts` (resource-grouped diff), `use-release-detail.ts` (lazy getRelease + cache), small `derive.ts` additions (`phaseTone`, `releaseTitle`, `releaseMeta`).

**Removed:** `current-deployment-card.tsx`, `release-history.tsx`, `release-row.tsx`, `unreleased-changes-banner.tsx`, `failing-resources-accordion.tsx`, `release-detail-drawer.tsx`, `release-diff.ts` (+ their test files), `src/pages/__demo/`, the `App.tsx` demo route, `deploy-*.jpeg`, and the demo-tied verification handoff doc.

### Component responsibilities (isolation)

| Component | Does | Input | Depends on |
|---|---|---|---|
| `DeploymentsTab` | orchestrates data + actions + drift; lays out the rail | orgId, teamName, stackId, stack, canDeploy | useReleases, TimelineRail |
| `TimelineRail` | renders rail dots/connectors, markers, current node, history rows, show-more, empty | vm (releases, activeRelease, stack, handlers, drift) | RailNode, CurrentReleaseNode, HistoryRow, banners |
| `RailNode` | one dot + vertical connector + content slot (presentational) | tone, big, pulse, isLast, children | tokens only |
| `CurrentReleaseNode` | live status of the active release | release, stack, logContext | derive, StageTracker, StatusPill, ResourceRow |
| `ResourceRow` | one live resource; expand failing → detail | resource summary + joined failure, logContext | FailureCard, LogSnapshot |
| `HistoryRow` | one release line; click → toggle post-mortem; ⋮ menu | release, isActive, prevRelease, handlers | ReleaseMenu, ReleasePostMortem |
| `ReleasePostMortem` | lazy detail: why-failed, outcomes, diff | orgId, teamName, stackId, releaseId, prevReleaseId | useReleaseDetail, OutcomesTable, ConfigDiff |
| `OutcomesTable` | resource outcomes grid | ReleaseOutcome.resources | tokens |
| `ConfigDiff` | resource-grouped added/removed/modified | prevSnapshot, snapshot | release-snapshot-diff |
| `ReleaseMenu` | state-gated ⋮ actions | release, handlers | ui/dropdown-menu |
| `DriftBanner`/`ReleaseErrorBanner` | top banners | drift/error props | AlertBanner |

---

## 3. Data flow

- `useReleases` polls `listReleases` every 5s while the active release is non-terminal (reused). Returns newest-first `releases`, `activeRelease`, `loading`, `error`, `refetch`.
- **Current node** reads the *live* picture from `stack.status.resources[]` (progress) joined with `stack.spec.stack_resources[].status.last_failure` (structured failure) — the two-array principle, already in `derive.ts`. Release-level errors (no per-resource failure) render `release.message` in `ReleaseErrorBlock`.
- **History rows** render from the *list* payload (light — no snapshot/outcome). Expanding a row lazily fetches the heavy detail:
  - `useReleaseDetail(id)` → `getRelease` (carries `snapshot` + `outcome`), cached in a `Map<id, {data|error|loading}>`.
  - Config diff needs the **previous** release's snapshot too (the list omits it), so the post-mortem also fetches `getRelease(prevId)` where `prevId` = the next-older release by sequence (newest-first ⇒ `releases[idx+1]`). Cached the same way; reused across rows.
- Expansion state: one resource open per current node; one history row open at a time (accordion semantics), reset on `activeRelease.id` change.

### History inclusion (deviates from the literal mock — confirm)

The mock duplicates the active release in some scenarios (shown as both the current node *and* the first "Earlier release") but not others. Production rule chosen here = **no duplication**:

- Current node = the active release (live status), **lifted out** of the Earlier-releases list.
- "Earlier releases" = `releases` excluding the active release.
- The active release's own changelog (config diff vs previous + outcomes) stays reachable: the **current node is itself expandable** into the same `ReleasePostMortem` (a "View changelog" affordance), so nothing is lost and nothing is shown twice.
- When `activeRelease` is null (no live deploy — scenarios 11/12), every release renders in the Earlier-releases / history list.

This is a deliberate, reviewable deviation from the mock's literal duplication.

---

## 4. Config diff (resource-grouped)

`release-snapshot-diff.ts` replaces the generic `release-diff.ts`. Snapshot is `StackSnapshot { stack, resources[], volumes, connections }` (untyped JSONB → cast to a local `SnapshotShape` interface). Algorithm:

1. Key `prev.resources` and `cur.resources` by resource name.
2. Classify each name: **added** (cur only), **removed** (prev only), **modified** (in both, fields differ), unchanged (omit).
3. For modified/added resources, split fields into two sections:
   - **configuration** — non-env scalar fields (e.g. `image_spec.image`, `ports[].number`, `command`, `replicas`).
   - **environment** — env var map; each var is added / removed / changed.
4. Each leaf row carries `{ key, from?, to?, kind: added|removed|changed }`.

`ConfigDiff` renders per-resource cards: header (dot + name + `ADDED`/`REMOVED` tag + optional note), then sections with `from → to` (strikethrough old in danger, new in success), `— → new` for added, removal note for removed resources. Exact snapshot field paths are pinned during implementation against a real release snapshot (TDD fixture).

Edge: first release (no previous) → "Initial release — nothing to compare." Missing/oversized snapshot → graceful "Diff unavailable."

---

## 5. Error & empty handling

| Condition | UI |
|---|---|
| list fetch error | `EmptyState` "Could not load deployments" + message |
| no releases, no active | `EmptyState` "No deployments yet" rail node |
| release-level failure (Failed, no per-resource) | `ReleaseErrorBlock` in current node (release.message) — closes the prior gap |
| post-mortem getRelease error | inline error line in the expanded row; outcomes/diff omitted |
| diff: no previous | "Initial release — nothing to compare." |
| log snapshot empty (best-effort, #98) | hidden |
| action (deploy/rollback/cancel) failure | destructive toast (reused `run()`) |

---

## 6. Pagination (deferred)

Backend `GET /releases` has **no** pagination today: `ListByStackID` returns every release `ORDER BY sequence DESC`, `StackReleaseList = {items, total}`, no cursor. So:

- **Now:** render all returned releases; show the first window (e.g. 15) of *history* rows with a **"Show more"** button revealing +15 each click. No extra requests.
- **Later (filed as a backend issue):** add `limit` + cursor (e.g. `before_sequence`) to `listReleases` (OpenAPI + handler + service + store), regenerate clients, and swap the client window for fetch-more / IntersectionObserver infinite scroll. `TimelineRail` is built so only the "load more" source changes.

---

## 7. The 12 scenarios

| # | scenario | Current node | History | Notes |
|---|---|---|---|---|
| 1 | released | READY #14, Build✓Deploy✓Ready✓, 3 resources Ready; expandable changelog | earlier (#13,#12,#11) | active lifted, not duplicated |
| 2 | recovered | as released + `RECOVERED` tag + amber recovered note | earlier | resource Ready but had last_failure |
| 3 | drift | released + top **DriftBanner** (Deploy changes) | earlier | stack.updated_at > completed_at |
| 4 | pending | PENDING #15, Build **active**, resource Building | minimal | cancellable |
| 5 | inprogress | INPROGRESS, Build✓ Deploy **active**, Progressing 0/1 | minimal | polling |
| 6 | build_fail | INPROGRESS, Build **✗**, resource Build-failed → detail (builder reason, no logs) + top error banner + tab alert | minimal | live-status-first |
| 7 | crash_single | INPROGRESS, Deploy **✗**, tooljet CrashLoopBackOff → detail (reason, empty log) | minimal | |
| 8 | crash_multi | FAILED, Ready **✗**, 2 failing (Crash + OOMKilled, with log snapshot) + top error banner | 2 rows | Rollback + Redeploy actions |
| 9 | release_err | FAILED, Build✗, **ReleaseErrorBlock** (apply: unknown postgres addon) | minimal | no per-resource status |
| 10 | pre_cluster | FAILED, Build✗, **ReleaseErrorBlock** (render: unresolved DATABASE_URL) | row w/ err meta | failed before cluster |
| 11 | history | no current node; "Release history · N" | all states + causes + ⋮; #9 expands → why-failed (apply quota) | View/Rollback/Cancel/Copy gated by state |
| 12 | empty | none | EmptyState rail node | |

---

## 8. Brand token mapping (design → ours)

| design token | ours (`index.css` / Tailwind) |
|---|---|
| `--amber`, `--amber-soft`, `--amber-hover` | `--primary` (amber), `--warn-bg`, primary hover; classes `text-primary`/`bg-warn-bg` |
| `--ok` / `--ok-soft` (green) | existing success/`ready` token used by `StatusPill`'s `ready` variant — reuse the variant; for bare text use that token (verify exact name in impl) |
| `--err` / `--err-soft` | `--danger` / `--danger-bg` / `--danger-border` |
| `--warn` / `--warn-soft` | `--warn` / `--warn-bg` |
| `--fg1/2/3`, `--fg-muted` | `--foreground`, `--fg-2`, `--muted-foreground`, `--fg-muted` |
| `--bg`, `--bg-card`, `--bg-elev` | `--background`, `--card`, `--muted`/`--accent` |
| `--border` | `--border` |
| `--font-sans`, `--font-mono` | `--font-sans` (Geist), `--font-mono` (Geist Mono) |
| `--shadow-2` | `--shadow-md`/`--shadow-lg` |

Status/stage/failure/log/banner visuals come from the **branded primitives** (already tokenized), so most color decisions are inherited, not re-specified. New chrome (rail dots, connectors, phase text, diff add/remove) uses the mapped tokens above. Dark + light both ship (tokens already define both).

---

## 9. Testing (Vitest + RTL)

- `release-snapshot-diff.test.ts` — added/removed/modified, config vs env split, no-previous, malformed snapshot.
- `use-release-detail.test.tsx` — fetch + cache hit + error.
- `current-release-node.test.tsx` — stages per state, resource expand (failing only), recovered note, release-error block.
- `resource-row.test.tsx` — expand gated on failing; FailureCard + LogSnapshot; replica display.
- `history-row.test.tsx` — toggle post-mortem; ⋮ items gated by state; copy id.
- `release-post-mortem.test.tsx` — loading → outcomes + diff; why-failed; error path; initial-release diff.
- `outcomes-table.test.tsx`, `config-diff.test.tsx` — rendering of each row kind.
- `timeline-rail.test.tsx` — current+history composition, markers, show-more window, empty.
- `deployments-tab.test.tsx` — deploy/rollback/cancel handlers fire + refetch; drift banner; release-error banner.
- Coverage spans all 12 scenarios via fixtures. TDD per task (test-first, fail→pass, commit).

---

## 10. Cleanup (explicit)

Delete: `src/pages/__demo/` (DeploymentsDemo + fixtures), the two `// DEMO-ONLY` lines + import in `App.tsx`, `deploy-*.jpeg` (worktree root), and the demo-tied verification handoff doc. Remove the 7 superseded components/modules + their tests (§2). The release-level-error finding from that doc is now resolved by `ReleaseErrorBlock`.

---

## 11. Follow-ups (filed separately)

- Backend: `listReleases` pagination (`limit` + cursor) → enables true infinite scroll.
- Backend: rollback + cancel handler/service tests (no e2e coverage today).
- (Optional) include `outcome` in the list payload to avoid the per-row getRelease for outcomes — weighed against payload size; current lazy approach preferred.
