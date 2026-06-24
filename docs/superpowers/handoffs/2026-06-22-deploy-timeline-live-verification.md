# Deploy Timeline — live verification (2026-06-22)

Env: fresh `mage dev:setup` + migrate, API `./bin/api-server serve` on :8000 (embeds branch frontend), k3d cluster registered to org DevOrg. Signed up admin@stackdome.dev, created stack `demo-timeline` (1 resource `web` = nginx:alpine), deployed twice.

## Verified working (happy path)
- Empty state: "No deployments yet" renders on the rail.
- Deploy → `POST /releases` 201; current node renders: StatusPill, `#seq`, meta (`config <rev7> · took 11s`), Build→Deploy→Ready StageTracker.
- 5s polling of `GET /releases` updates the node live (InProgress ✓→ Released) without reload.
- Resource row renders (`web | Ready | 1/1`) from `stack.status.resources`.
- Changelog expand → Resource outcomes table (from `outcome.resources`) ✓.
- Multi-release: current = #2, "Earlier releases" marker, HistoryRow #1 (`Manual deploy · 11s` + ⋮).
- State-gated ⋮ menu on a Released release: View details · Rollback to this · Copy release ID — **no Cancel** ✓.
- No console crashes through the flow.

## Findings (API ↔ UI mismatches)

### 1. IMPORTANT — Config diff is non-functional against the real API
- `StackRelease` OpenAPI schema and `pkg/presenters/stack_release.go` expose `snapshot_revision` + `outcome` but **never the `snapshot` object**.
- Frontend `release-post-mortem.tsx` reads `(data as unknown as {snapshot?}).snapshot` — an `as unknown` cast that is always `undefined` at runtime → `diffSnapshots(undefined, undefined) => []`.
- Live proof: #2's changelog shows "Config changes · vs #1" → body renders **"Initial release — nothing to compare."** even though #2 has a real previous (#1).
- Fix (backend): include `snapshot` (StackSnapshot) on the `StackRelease` detail (`getRelease`) response + schema, OR drop the config-diff UI. Unit tests pass only because they inject fake snapshots.

### 2. IMPORTANT — Live resource/failure state is stale until reload
- `DeploymentsTab` polls only `GET /releases` (via `useReleases`). The `stack` prop (carrying `stack.status.resources` + `last_failure`) comes from `index.tsx` and is **not** re-fetched.
- Live proof: deployed #1 from a freshly-created (Pending) stack; node went InProgress→Released via polling, but the `web` resource row never appeared and the stack header stayed "Pending". After a manual page reload: header "Ready" + `web | Ready | 1/1` row appeared.
- Impact: the headline goal — surfacing build/runtime failures *live* from status (not `release.state`) — does not actually update without a reload. Need to poll the stack (or its status) alongside releases.

### 3. MINOR — Rollback cause label is garbled
- Backend `stack_release_service.go:153` sets rollback `cause.detail = fmt.Sprintf("rollback to release #%d", seq)` → e.g. `"rollback to release #1"`.
- Frontend `causeLabel` renders rollback as `` `Rollback to #${cause.detail}` `` → **"Rollback to #rollback to release #1"**.
- Fix: backend should put the bare sequence in `detail` (e.g. `"1"`), or the frontend should not prefix `#`. (Manual deploy is unaffected — `causeLabel` ignores `detail` for `manual`, even though it is `"triggered by <uuid>"`.)

### 4. MINOR — ConfigDiff empty copy is misleading
- "Initial release — nothing to compare." is shown for *non-initial* releases when there is simply no diff data (see #1). Copy assumes empty == initial release.

## Not live-verified (need failure injection)
build_failure, runtime_crash (single/multi), drift banner, recovered note, release-level error block. Covered by unit tests; not reproduced live (would need a failing image / DB-seeded `last_failure` + release `message`).
