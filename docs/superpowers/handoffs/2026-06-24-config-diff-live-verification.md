# Per-release config diff — live verification

**Date:** 2026-06-24
**Plan:** `docs/superpowers/plans/2026-06-23-per-release-config-diff.md`

## What was verified

The snapshot-powered config diff renders real backend data end-to-end.

Required rebuilding the dev backend first: the running `bin/api-server` predated
the `main` merge, so `GET /releases/{id}` returned no `snapshot` and the release
worker stored none. Steps taken:

1. `go build -o bin/api-server ./cmd` (merged code).
2. `mage migrate` (applied the merge's migrations: resource_references, stack settings).
3. Restarted the server; release-worker now stores `StackSnapshot` per release.

Then drove a throwaway `diff-demo` stack through two differing releases:

- #3 — `nginx:1.27`, env `FOO=baz`, `BAZ=qux`
- #4 — `nginx:1.28`, env `FOO=prod` (BAZ removed)

On the active card (#4 Released), "View config changes · vs #3" renders:

- **CONFIGURATION** — `image: nginx:1.27 → nginx:1.28`
- **ENVIRONMENT** — `FOO: baz → prod`, `BAZ: qux` (removed, struck through)

Screenshot: `diff-verified.jpeg` (worktree root).

## Coverage notes

- Resource (image) + env diff: verified live.
- Volume diff: release snapshots don't re-version volume size (data stayed 1Gi
  across releases despite spec PUTs), so the volume group stayed empty live;
  covered by unit tests (`release-snapshot-diff.test.ts`).
- Connection diff: covered by unit tests; not exercised live (needs an addon
  connection — heavier seed).

## Follow-ups

- Active-card diff toggle (`showDiff`) appears to collapse on each 5s poll while
  the release is non-terminal (deploying). Stable once terminal; history-row
  diffs unaffected. Investigate whether the poll re-render remounts
  `CurrentReleaseNode` and persist the open state if so.
