# Releases

## What releases are

A release is an immutable, versioned deployment of a stack. Every time a user deploys or rolls back, the system creates a release that captures the exact state of the stack (resources, connections, env vars, volumes, git SHAs) and drives it to convergence on the target cluster.

Before releases, the stack worker applied changes directly to the cluster whenever the stack was modified. This made it impossible to answer "what changed?", "when was it deployed?", or "can I go back?" Releases solve this by separating **intent** (the stack spec in the DB) from **deployment** (the release manifest applied to the cluster).

## What releases enable

- **Rollback**: revert to any previous successful release in one API call. The manifest is reused — no re-rendering, no re-resolving connections. Rollbacks are instant.
- **Deploy history**: every deployment is a numbered, timestamped record with cause, duration, outcome, and the exact git SHAs/image digests that were deployed.
- **Audit trail**: who triggered it (user ID), why (manual, rollback, webhook push), and what happened (Released, Failed with message, Superseded by a newer release).
- **Safe concurrent deploys**: if a user triggers a new deploy while one is in progress, the system handles it cleanly — the older release detects it's been overtaken and marks itself Superseded.
- **Pin reproducibility**: each release records the resolved git SHA, volume hash, and image digest for every resource. A rollback to release #3 deploys the exact same artifacts, not "whatever main points to now."

## Lifecycle

```
User edits stack spec (resources, connections, env vars)
  |
  v
User creates a release (POST .../releases)
  |
  v
Pending ──> InProgress ──> Released
                |
                +──> Failed (timeout, apply error)
                |
                +──> Superseded (newer release took over)

Pending ──> Cancelled (user cancelled before worker picked it up)
        |
        +──> Superseded (newer release created)
```

### States

| State | Meaning | Terminal? |
|-------|---------|-----------|
| Pending | Created, waiting for the release worker to pick it up | No |
| InProgress | Worker is rendering, applying, or waiting for convergence | No |
| Released | All resources converged on the target cluster | Yes |
| Failed | Apply error or convergence timeout (15 min) | Yes |
| Superseded | A newer release took over (detected by gatekeeper check) | Yes |
| Cancelled | User cancelled via API (only from Pending) | Yes |

### Key timestamps

| Field | Set when |
|-------|----------|
| `created_at` | Release created |
| `rendered_at` | Manifest rendered (or copied from source on rollback) |
| `completed_at` | Terminal state reached (Released, Failed, Superseded, Cancelled) |

`outcome.duration` = time from `rendered_at` to `completed_at`.

## How it works internally

### 1. Create

`POST .../stacks/{id}/releases` with empty body (deploy) or `{"from_release_id": "..."}` (rollback).

**Deploy path:**
1. Deep-copy the entire stack (resources, connections, volumes, labels, annotations) into a `Snapshot`.
2. Resolve pins: for each resource with a build spec, resolve the git SHA (commit > tag > branch HEAD), volume hash, and image digest.
3. Assign `sequence = MAX(sequence) + 1` for this stack.
4. Insert with `state = Pending`, enqueue for background processing.

**Rollback path:**
1. Copy the source release's `Snapshot`, `Manifest`, `ManifestRevision`, `Pins`, and `RendererVersion`.
2. Set `rendered_at = now` (pre-rendered — no render step needed).
3. Insert with `state = Pending`, enqueue.

### 2. Worker pipeline

The release worker runs four sub-reconcilers in order. Each observes the environment and decides whether to act, skip, or requeue.

**Gatekeeper** — Is this release still the latest?
- Query the DB for the highest-sequence active release for this stack.
- If a newer release exists, mark self `Superseded` and stop.
- If state is `Pending`, CAS transition to `InProgress`. If CAS fails (another worker won), stop.

**Render** — Does this release have a manifest?
- If `manifest != nil`, skip (already rendered, or rollback with pre-copied manifest).
- Otherwise: reconstruct the stack from the snapshot, resolve all connections (postgres credentials, secrets, resource-to-resource env vars, volume mounts), build the Stack CR and StackResource CRs, compute per-resource revision hashes, save the manifest via CAS, requeue.

**Apply** — Are the cluster CRs up to date?
- Sync hub secrets (image pull/push, git credentials) and postgres credential secrets to the cluster namespace.
- Check volume readiness (all volumes must be `Ready` in DB).
- Apply Stack CR and StackResource CRs to the cluster using `DeepDerivative` comparison (only updates if our desired fields differ from existing). Sets ownerRefs on StackResource CRs pointing to the Stack CR.
- Prune orphaned StackResource CRs not in the current manifest.
- If no changes were needed, falls through to converge in the same loop iteration.

**Converge** — Has the cluster reached the desired state?
- Poll `stack.Status.LastConverged` from the DB (written by the stack controller watching the cluster).
- If `LastConverged.ReleaseID == release.ID && LastConverged.Revision == release.ManifestRevision`, mark `Released` with outcome (per-resource phase, replica counts, duration).
- If 15 minutes since `rendered_at`, mark `Failed` with timeout message.
- Otherwise, requeue after 15 seconds.

### 3. Convergence detection

The cluster-agent's Stack controller aggregates the status of all child StackResource CRs. When all children are converged (matching revision, available replicas), it writes `LastConverged = {Revision, ReleaseID, At}` to the Stack CR's status.

The API server's stack controller watches the Stack CR via controller-runtime. When the `StatusHash` changes, it maps the cluster status into the DB (`stacks.status` JSONB column) and re-enqueues the active release so the converge reconciler can observe the update.

## API reference

Base path: `/api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/releases`

### Create release

```
POST /releases
Content-Type: application/json

{}                                          # deploy current stack spec
{"from_release_id": "<release-id>"}         # rollback to a previous release
```

Response: `201 Created` with `StackRelease` object.

### List releases

```
GET /releases
```

Response: `200 OK` with `{"items": [...], "total": N}`. Ordered by sequence descending (newest first). The `manifest`, `snapshot`, and `outcome` JSONB fields are excluded from list responses for performance.

### Get release

```
GET /releases/{release_id}
```

Response: `200 OK` with full `StackRelease` object including `outcome` and `pins`.

### Cancel release

```
POST /releases/{release_id}/cancel
```

Response: `200 OK`. Only works on `Pending` releases. Returns 400 if `InProgress` or terminal.

## StackRelease object

```json
{
  "id": "uuid",
  "stack_id": "uuid",
  "sequence": 14,
  "state": "Released",
  "message": "",
  "cause": {
    "kind": "manual",
    "detail": "triggered by <user-id>"
  },
  "snapshot_revision": "sha256-of-snapshot",
  "manifest_revision": "fnv64a-of-manifest",
  "renderer_version": "2026-06-14.1",
  "pins": {
    "resources": {
      "app": {
        "git_sha": "20d73f323a4d95ff5a3847717892e1740a5a81b6",
        "image_digest": "sha256:abc123...",
        "volume_hash": ""
      }
    }
  },
  "outcome": {
    "duration": "32.17s",
    "resources": {
      "app": {
        "phase": "Ready",
        "ready_replicas": 1,
        "replicas": 1,
        "message": ""
      },
      "web": {
        "phase": "Ready",
        "ready_replicas": 1,
        "replicas": 1,
        "message": ""
      }
    }
  },
  "created_by": "user-uuid",
  "created_at": "2026-06-17T06:49:08Z",
  "rendered_at": "2026-06-17T06:49:08Z",
  "completed_at": "2026-06-17T06:49:40Z",
  "updated_at": "2026-06-17T06:49:40Z"
}
```

### Cause kinds

| Kind | When |
|------|------|
| `manual` | User clicked deploy / called the API |
| `rollback` | User rolled back to a previous release |
| `webhook_push` | Git push webhook triggered a deploy (future) |

### Pins

Pins capture the exact artifact versions resolved at release creation time. Only resources with build specs have pins. Image-only resources (e.g., `nginx:latest`) don't generate pins — the image tag is in the snapshot.

| Field | Source |
|-------|--------|
| `git_sha` | Resolved from commit, tag, or branch HEAD at creation time |
| `image_digest` | Populated after image build completes (may be empty during build) |
| `volume_hash` | Hash of volume source content |

### Outcome

Only present on terminal releases (`Released` or `Failed`). Contains per-resource status at the moment the release completed, plus total duration from render to completion.

## Frontend design guide

### Release list view

The release list is the deploy history for a stack. Each row should show:

- **Sequence number** (`#14`) — the deploy number, always incrementing
- **State badge** — color-coded: green (Released), red (Failed), yellow (InProgress), gray (Pending, Cancelled, Superseded)
- **Cause** — icon or label: manual deploy, rollback (with "from #N"), webhook push
- **Duration** — from `outcome.duration` (only on terminal releases)
- **Created by** — user who triggered it
- **Timestamps** — `created_at`, optionally `completed_at`
- **Message** — failure reason for Failed releases, rollback detail for rollbacks

The list endpoint excludes heavy JSONB fields (manifest, snapshot) so it's safe to poll.

### Release detail view

Clicking a release shows:

- Full state timeline (created -> rendered -> completed)
- **Resource outcomes table**: per-resource phase, replica counts, messages
- **Pins**: git SHA (linkable to GitHub commit), image digest, volume hash
- **Pin diff**: compare pins between this release and the previous one to show what changed (new git SHA, different image)
- **Snapshot revision**: identifies what stack config was deployed (same revision = same config)

### Deploy flow

1. User edits stack resources/connections in the UI
2. User clicks "Deploy" button
3. Frontend calls `POST /releases` with empty body
4. Show the release in `Pending` state, start polling `GET /releases/{id}`
5. Update UI as state transitions: `Pending` -> `InProgress` (show spinner) -> `Released` (show success) or `Failed` (show error message)

### Rollback flow

1. User views release history, clicks "Rollback" on a previous `Released` entry
2. Frontend calls `POST /releases` with `{"from_release_id": "<id>"}`
3. Same polling flow as deploy. Rollbacks are faster (no render step) — typically 1-2 seconds if images are cached.

### Cancel flow

User clicks "Cancel" on a `Pending` release. Calls `POST /releases/{id}/cancel`. Only works on Pending — if the release is already InProgress, show an error.

### Live status during deploy

While a release is `InProgress`, the stack's status reflects the deploy progress:

- `stack.status.state` transitions: `Pending` -> `Progressing` -> `Ready`
- `stack.status.resources[].phase`: per-resource status (Pending, Ready, Failed)
- `stack.status.resources[].message`: human-readable progress ("serving traffic but not fully converged", "not available: ImageBuildInProgress")

Poll `GET /stacks/{id}` alongside the release to show per-resource progress bars or status indicators.

### State machine for UI

```
Pending:      gray badge, "Queued" label, show cancel button
InProgress:   yellow badge, spinner, show resource progress from stack status
Released:     green badge, checkmark, show outcome summary and duration
Failed:       red badge, X icon, show message (e.g., "timed out after 15m")
Superseded:   gray badge, "Superseded" label, no action buttons
Cancelled:    gray badge, "Cancelled" label, no action buttons
```

### Diffing and change detection

There are three diff scenarios the frontend needs:

#### 1. "What will this deploy change?" — Current stack vs last deployed

Before deploying, show what changed since the last release. The left side is the current stack (from `GET /stacks/{id}`), the right side is the last `Released` release's snapshot.

The release's snapshot contains the full stack state at creation time — the same resources, connections, and volumes that the user sees in the stack editor.

> **Not yet implemented:** The snapshot is stored in the DB but not currently exposed in the release API response. The `GET /releases/{id}` endpoint returns `snapshot_revision` (hash) but not the snapshot contents. The backend needs a new field or sub-endpoint to expose the snapshot for the frontend to diff against. Design the UI assuming this data will be available — the API work is straightforward (the data exists, it just needs to be serialized in the response).

When available, the snapshot will look like:

```json
// GET /releases/{release_id}
{
  "snapshot": {
    "stack": {
      "id": "...", "name": "test-release", "namespace": "...",
      "labels": [...], "annotations": [...]
    },
    "resources": [
      {
        "name": "web",
        "image_spec": {"image": "nginx:1.25-alpine"},
        "ports": [{"name": "http", "number": 80, "exposed_to_public": false}],
        "execution_config": {"env": [{"name": "PG_HOST", "secret_key_ref": {...}}]},
        "volume_mounts": [{"mount_path": "/data", "volume_name": "uploads"}]
      },
      {
        "name": "worker",
        "image_spec": {"image": "nginx:1.25-alpine"},
        "depends_on": ["web"]
      }
    ],
    "volumes": [
      {"name": "uploads", "spec": {"size": "1Gi", "access_mode": "ReadWriteOnce"}}
    ],
    "connections": [
      {"from": {"type": "postgres_addon", "id": "..."}, "to": [...], "mappings": [...]}
    ],
    "captured_at": "2026-06-17T06:49:08Z"
  }
}
```

The frontend diffs by walking the snapshot's resources/connections/volumes against the current stack:

| What to diff | How to detect |
|---|---|
| Resource added | Name in current stack but not in snapshot.resources |
| Resource removed | Name in snapshot.resources but not in current stack |
| Image changed | Compare `image_spec.image` |
| Build source changed | Compare `build_spec.source_revision` and release `pins.resources[name].git_sha` |
| Ports changed | Compare port list (number, exposed_to_public, protocol) |
| Env vars changed | Compare execution_config.env lists |
| Connection added/removed | Compare connection lists by from/to |
| Volume added/removed | Compare volume lists by name |
| Volume config changed | Compare spec (size, access_mode) |

If `snapshot_revision` matches between the current stack and the last release, nothing changed — the Deploy button can be disabled.

#### 2. "What changed between two releases?" — Release vs release

Fetch both releases via `GET /releases/{id}` and diff their snapshots and pins:

**Quick indicators:**
- Same `snapshot_revision` = identical stack config (common for rollbacks)
- Different `pins.resources[name].git_sha` = new code version deployed
- Different `pins.resources[name].image_digest` = different image built

**Detailed diff:** walk both snapshots' resources, connections, volumes the same way as scenario 1. The snapshot is the source of truth for what was deployed — not the current stack state.

#### 3. "Is there an undeployed change?" — Banner/indicator

Compare the current stack against what's running on the cluster:

```json
// GET /stacks/{id}
{
  "status": {
    "last_converged": {
      "release_id": "f2c6b842-...",
      "revision": "56844845dddffcf4b696",
      "at": "2026-06-17T06:26:49Z"
    }
  }
}
```

If `last_converged` is null, the stack has never been deployed. If the current stack's snapshot would produce a different revision than `last_converged.revision`, there are undeployed changes. Show a banner: "You have undeployed changes" with a Deploy button.

> **Not yet implemented:** There's no API endpoint to compute the current stack's snapshot revision without creating a release. The backend needs either a `GET /stacks/{id}/pending-changes` endpoint that computes and compares the snapshot hash, or expose the snapshot on the release detail so the frontend can diff the two snapshots directly. For now, the frontend can use a simpler heuristic: if `stack.updated_at > last_converged.at`, show the banner.

The `last_converged.release_id` links to the release that's currently live. Fetch it to show "Currently running: Release #14, deployed 2 hours ago."

#### Data available for diffing

| Data | Source | Notes |
|---|---|---|
| Current stack state | `GET /stacks/{id}` | Live DB state with user's latest edits |
| Release snapshot | `GET /releases/{id}` → `snapshot` | Frozen state at release creation |
| Release pins | `GET /releases/{id}` → `pins` | Resolved git SHAs, image digests |
| Per-resource revision | `manifest_revision` on release | Hash of the rendered CRs |
| What's running | `stack.status.last_converged` | Release ID and revision on cluster |
| Release list | `GET /releases` | Excludes snapshot/manifest for perf |

Note: the list endpoint (`GET /releases`) excludes snapshot, manifest, and outcome for performance. Fetch the full release via `GET /releases/{id}` when the user drills into a release or when computing a diff.
