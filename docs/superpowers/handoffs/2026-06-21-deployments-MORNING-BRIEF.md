# ☀️ Morning Brief — Deployments Tab

**TL;DR: The Deployments tab is fully implemented, reviewed, tested, and visually verified. 383/383 tests pass. Nothing is pushed — your call on PR. Two things want your eyeballs (below).**

Branch: `feat/stateful-deployments` (worktree), off `origin/main` @ `de222a8` (Releases #100). HEAD `cced91c`.

---

## What got built (Tier 1–3, all 18 plan tasks)

A Railway-style **Deployments** tab on the Stack detail page, on the Releases API:
- **Current deployment card** — status pill + `#sequence` + `config <hash>` + duration, a **Build→Deploy→Ready** stage tracker, live per-resource table, and a "recovered" note for healthy-but-previously-failed resources.
- **Release history** — newest-first list, state pills, cause labels (Manual / Rollback to #N / Webhook), git SHA + derived duration, state-aware ⋮ menu (View / Rollback to this / Cancel).
- **Failure visibility** — click-through accordion (one open at a time) of failing resources with `FailureCard` detail; release-level render/apply errors as an `AlertBanner`; best-effort crash-log snapshot (rendered only when non-empty).
- **Detail drawer** — full release fetch: lifecycle, resource outcomes, config-diff vs the previous release, durable "Why it failed" post-mortem.
- **Decoupled Save/Deploy** — the sticky bar's primary is now correctly **"Save"** (it was mislabeled "Deploy" but only ever PUT); **Deploy** is the explicit `POST /releases` action + an "Unreleased changes" drift banner.

## Quality gates (all green)
- **383/383** frontend tests (49 files). Every task built TDD (test-first, fail→pass), each behind an independent **subprocess review** (spec + quality). Reviews caught **3 real bugs** before they landed — see log.
- `tsc`: only the **pre-existing** unrelated `postgres-backups.ts` error (present on origin/main before this work).
- Lint: new code clean (1 pre-existing-style react-refresh warning).
- **Visual QA**: real components screenshotted across **8 success/error states (light theme)** via a throwaway harness (since reverted). All render correctly and design-system-faithful. (Dark theme not captured — quick follow-up.)

## 👀 Two things to look at
1. **Backend follow-up issues I filed** (outward-facing, done per the approved plan — close any you don't want):
   - #105 converge fast-fail on terminal resource failure
   - #106 missing-secret leaves release stuck `InProgress`
   - #107 expose current `snapshot_revision` (precise drift; kills the banner's post-deploy false-positive)
   - #108 snapshot `last_failure` into `release.outcome` (durable post-mortem)
   - comment on **#99** for the `last_validation_run` OpenAPI-hygiene point (folded in vs. a duplicate)
2. **The detail drawer was visually QA'd via unit tests + mockup only**, not against a live backend (it fetches `getRelease`). Worth one real-data pass when you next have `mage run` + a stack with releases up. Everything else was verified with real rendered components.

## Known honest limitations (by design, labelled in-UI + tracked)
- **Drift is heuristic** (`stack.updated_at` vs release `completed_at`) → brief false-positive right after a deploy. Labelled "(approximate)"; #107 fixes it precisely.
- **Crash log snapshot is best-effort** (often empty for crashing pods, backend #98) — shown only when non-empty; the structured `last_failure` is the reliable signal.
- **Post-mortem structured detail isn't durable yet** (#108) — live failures are full-detail; past failed releases show message + outcome summary.

## Your decision: how to integrate
Nothing pushed, no PR opened (deliberately, for your review). Options:
1. **Open a PR** — run `/create-pr` (it runs the full lint+test+tsc gate and pushes). Recommended once you've skimmed the diff.
2. **Keep local** — review the branch first, PR later.
3. **Adjust** — point me at anything in the visual QA or the filed issues you want changed.

## Where things are
- Plan + per-task trail: `docs/superpowers/plans/2026-06-21-stack-deployments-tab.md`
- Spec: `docs/superpowers/specs/2026-06-21-stack-deployments-tab-design.md`
- Mockup: `docs/superpowers/design-refs/deployments-tab-mockups.html`
- **Full decision/fix log (every ad-hoc call + the 3 caught bugs):** `docs/superpowers/handoffs/2026-06-21-deployments-overnight-log.md`
- Feature code: `frontend/src/pages/stacks/components/detail/deployments/` + salvaged primitives in `frontend/src/components/branded/` + `frontend/src/api/releases.ts`.
