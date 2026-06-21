# Deployments Tab — Overnight Execution Log

> Autonomous subagent-driven execution of `docs/superpowers/plans/2026-06-21-stack-deployments-tab.md`
> on branch `feat/stateful-deployments` (worktree). Started 2026-06-21.
> Methodology: fresh implementer subagent per task + independent review subagent(s). Two-stage
> (spec then quality) for logic-bearing tasks; combined single review for thin/dictated/verbatim tasks.

## Ad-hoc decisions & deviations (with rationale)

1. **Task 1 (types regen) executed inline, no implementer subagent.** Pure code generation
   (`pnpm generate:openapi-types`/`:zod`); no implementer judgment to review. Verified the only
   tsc error is the pre-existing unrelated `postgres-backups.ts` one (present on pristine origin/main
   before regen — confirmed by stashing the regen and re-running tsc).

2. **Task 3 (salvage 5 branded primitives): skipped a separate quality-review subagent.** Files were
   copied byte-identical from the already-QA'd parked branch (`feat/stack-activity-tab`); spec review
   confirmed verbatim copies + that the only new lines (barrel re-exports) match the real export names.
   No new logic to quality-review.

3. **Task 4 (stage tracker) reviewed spec-only.** Implementation was dictated verbatim in the plan;
   spec compliance subsumes quality for dictated code.

4. **Task 7 (useReleases): accepted an implementer DEVIATION from the dictated polling code.** The
   dictated `shouldPoll`-gated `setInterval` was flaky under vitest fake timers (interval registered
   only after a state-driven re-render). Implementer switched to an always-on interval (while enabled)
   that gates the fetch on an `activeIsNonTerminal` ref. Trade-off: a perpetual 5s no-op timer while the
   tab is open (negligible) for deterministic tests. All required semantics preserved (verified by
   review + an added "stops polling when terminal" test).

5. **Task 8 (release row): implementer extended `variantFromState` in the shared `status-pill.tsx`**
   to map the 6 release states, and used `@testing-library/user-event` + the existing jsdom
   pointer-capture shim for the Radix dropdown. All new variant returns are existing `StatusVariant`
   members (no invented tokens). Accepted.

6. **Task 10 (current deployment card): REAL BUG caught by quality review and fixed.** The implementer's
   phase+replica row merge collapsed the replica display into a single expression
   `{r.available_replicas ?? 0 / r.replicas ?? 0}` — `/` binds tighter than `??`, so it rendered
   "Ready · 0" instead of "Ready · 1/1". Fixed by splitting into two brace-expressions with a literal
   slash; added a guard test that fails pre-fix / passes post-fix. Also fixed optional-key fallback
   (`key={r.name ?? idx}`) and "1 restart"/"N restarts" pluralization.

7. **Task 11 (wiring): reviewer mis-flagged `previousRelease` as an inverted off-by-one.** Trace showed
   the dictated `find((_r,i) => releases[i-1]?.id === open)` is CORRECT given the newest-first sort
   (returns the next-older release); the reviewer's suggested `i+1` would have returned the newer one
   (wrong). Did NOT apply the reviewer's fix; instead refactored to an explicit
   `findIndex(...) + releases[openIdx + 1]` form (identical behavior) to remove the ambiguity that
   misled the reviewer.

## Process notes
- After Task 11, switched to **one combined spec+quality review per task** (full two-stage reserved
  for the detail drawer, the most complex remaining task) to finish the remaining tasks reliably
  overnight while keeping an independent review on every change.
- Every task: TDD (test first, observed fail→pass), committed individually. Pre-existing
  `postgres-backups.ts` tsc error tolerated throughout (documented in the plan).

## Per-task commit trail
- Task 1 types regen: `1ce4176`
- Task 2 release API client: `86111b8` → fixes `b95a294`
- Task 3 salvage primitives: `68335cb`
- Task 4 stage tracker: `bf70a9b`
- Task 5 view-model pt1: `88cbeb5` → tests `35b11c3`
- Task 6 stage derivation: `07dae82` → docs+tests `6861961`
- Task 7 useReleases hook: `536668b` → test `be26377`
- Task 8 release row: `ea07827` → fixes `fdd91a7`
- Task 9 history list: `826c039`
- Task 10 current deployment card: `5a364d3` → bugfix `abadca7`
- Task 11 wire tab + Save relabel: `75c70ef` → refactor `ca83932`
- Task 12 failing-resources accordion: `47954eb` → fix (reason in header + single-open) `f3fd45f`
- Task 13 generic snapshot diff: `646e4b6` (also exported `deepEqual` from stack-diff.ts)
- Task 14 release detail drawer: `3e61461` → fix (stale-data/loading/outcome-msg) `be92ad0`
- Task 15 best-effort crash log: `5454848` → fix (refetch-loop dep) `f513e05`
- Task 16 unreleased-changes drift banner: `cced91c`
- Task 17 backend follow-ups: filed issues **#105, #106, #107, #108**; `last_validation_run` point added as a comment on existing **#99** (its root-cause bug) to avoid a duplicate.
- Task 18 final review + visual QA: see below (no code commit; demo harness reverted).

## Ad-hoc decisions (continued)

8. **Task 12 (accordion): restored a UX regression.** First pass dropped the failure reason from the collapsed accordion header (to dodge a duplicate-text test match), defeating the click-through "scan without expanding" intent. Restored reason in the header; fixed the test to assert on the card's unique `message`; added a single-open test.

9. **Task 14 (drawer): REAL bug caught — stale data on `releaseId` change.** The drawer stays mounted while `openReleaseId` switches, so without resetting state it showed the previous release's data during the next fetch. Fixed by `setRelease(null)/setError(null)` at the top of the effect; added a loading line + rendered the previously-dropped `outcome.message`; added error-path + non-failed tests.

10. **Task 15 (crash logs): REAL bug caught — refetch loop.** `CrashLog`'s effect depended on the inline `logContext` object (new ref every render); with 5s polling it would refetch logs every render. Fixed to depend on primitive fields (`ctx.orgId, ctx.teamName, ctx.stackId, resourceName`).

11. **Task 16 (drift banner): accepted a known heuristic limitation.** Drift compares `stack.updated_at` vs the active release's `completed_at`, which can show a brief false-positive right after a deploy (write-ordering). Honestly labelled "(approximate)" in the UI and tracked by backend follow-up #107 (expose `snapshot_revision`).

12. **Task 17 (filing issues): outward-facing action taken autonomously.** Filing GitHub issues notifies the team, but it is explicitly part of the approved plan (durable authorization), the user said "everything should be done," and issues are trivially reversible (close). Filed 4 (#105–#108) + a comment on #99. All reference the spec sections. **If any are unwanted, close them.**

## Visual QA (Task 18)

Built a throwaway demo harness (`__demo__/DeploymentsDemo.tsx` + a temporary unauthenticated `/__demo/deployments` route) rendering the REAL presentational components with fixtures for 8 states, served via `vite --port 5180`, screenshotted in **light and dark** via Playwright, reviewed, then **fully reverted** (route + harness + screenshots removed; `git status` clean; tsc still only the pre-existing error).

States verified rendering correctly + design-system faithful in both themes:
1. Drift banner (Unreleased changes, Deploy affordance)
2. Healthy/Released — stage tracker Build✓→Deploy✓→Ready✓, 3 resources Ready 1/1
3. Recovered — Ready resource + amber "recovered after 5 restarts — last failure CrashLoopBackOff" note
4. Build failure — INPROGRESS, Build✗ (red) + downstream greyed, `api ErrImagePull` accordion
5. Runtime crash (multiple) — Build✓→Deploy✗, 2-item click-through accordion (tooljet/worker)
6. Release-level error — AlertBanner "apply error: …forbidden: exceeded quota"
7. Standalone failures accordion
8. History list — correct state pills + cause labels (Manual deploy / Rollback to #11 / Webhook push), Failed message in danger red, ⋮ menus

**Note:** the detail DRAWER (Sheet) was not screenshotted (it fetches `getRelease`, needs a backend); it is covered by 4 unit tests + the design mockup. Worth a real-data pass once the backend is reachable (see brief).

## Final state
- HEAD `cced91c` on `feat/stateful-deployments` (off `origin/main` @ `de222a8` = Releases #100).
- Full suite: **383/383 passed** (49 files). tsc: only the pre-existing unrelated `postgres-backups.ts` error. New-code lint: clean (1 pre-existing-style react-refresh warning on status-pill.tsx).
- Not pushed; no PR opened (left for morning review).
