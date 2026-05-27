# Stack Topology Next Phases Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Continue the stack topology / explicit connections redesign from the current branch state, focusing next on making the new model operationally meaningful beyond persistence and read APIs.

**Architecture:** `Stack` remains the aggregate root. Explicit user-authored wiring lives in `stack_connections` and is projected into topology. Internal usage views should read connection-backed usage from `stack_connections` and persist `ResourceUsage` only for direct config references. The next phases should extend runtime resolution for non-`env` connection kinds and continue shifting toward independently mutable child resources validated against a materialized desired aggregate.

**Tech Stack:** Go, GORM/Postgres, OpenAPI-generated Go client, existing `handlers -> services -> stores` backend layering, worker reconciliation under `pkg/worker/stack`.

---

## References

- Primary spec: [docs/superpowers/specs/2026-05-22-stack-topology-connections-api-design.md](/Users/asnaraya/projects/skysync/api-server/api-server/docs/superpowers/specs/2026-05-22-stack-topology-connections-api-design.md)
- Broader foundation: [docs/superpowers/specs/2026-05-15-stack-foundational-primitive-design.md](/Users/asnaraya/projects/skysync/api-server/api-server/docs/superpowers/specs/2026-05-15-stack-foundational-primitive-design.md)
- Current branch: `stack-topology-connections-api`
- Relevant commits already on this branch:
  - `7d1043b` — named ports, connection models, config validation
  - `284b2d7` — output metadata and `self_output`
  - `0bb51e1` — topology endpoint, table-backed connection persistence, connection CRUD

## Current State

The branch already includes:

- Named ports on `StackResource`, with validation and output addressing like `port.http`, `url.http`, `public.http.url`
- Declared outputs for:
  - `StackResource`
  - `PostgresAddon`
  - `Secret`
- Explicit connection model:
  - `Stack.connections` in the aggregate
  - persisted in `stack_connections` table
  - surfaced via independent `/stacks/{id}/connections` CRUD
- Topology read model:
  - `GET /stacks/{id}/topology`
  - explicit edges from `stack_connections`
  - derived `depends_on` edges from `StackResource.DependsOn`
- Runtime env resolution for:
  - Postgres addon -> stack resource
  - stack resource -> stack resource
  - secret -> stack resource
- Inline `self_output` env var support

Important current design decisions:

- No backward-compatibility work is required for unreleased product state.
- Connection IDs are DB-generated, not app-generated.
- `depends_on` remains on `StackResource` for now and is projected into topology as a derived edge.
- Not every reference should become a connection. Build-time secrets, pull secrets, and similar config remain direct fields and should still be tracked internally as usage/dependency metadata where appropriate.

## Invariants For Next Work

- `Stack` must exist before child resources.
- Long-term direction is independent mutation of child resources like `StackResource` and `StackConnection`.
- Validation should be against the **materialized desired aggregate after mutation**, not only the request body.
- `stack_connections` are the source of truth for explicit wiring.
- `/topology` is a projection, not an editable source of truth.

## Phase Order

The next agent should implement these in order.

### Phase 1: Read Connection Usage From `stack_connections`

This is the highest-value next step. It makes explicit connections matter for delete protection, impact analysis, and dependency introspection.

Scope:

- Query explicit `stack_connections` directly for connection-backed usage.
- Cover at least:
  - `secret` -> consumer usage
  - `addon/postgres` -> consumer usage
  - `volume` -> consumer usage where connection kind warrants it
- Continue persisting `ResourceUsage` for direct config that intentionally remains outside connections:
  - build git secrets
  - registry push secrets
  - image pull secrets
  - Postgres backup object-store config
- Do not duplicate connection-backed usage into `resource_usages`.

Suggested file areas:

- [pkg/models/addon_usage.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/models/addon_usage.go)
- [pkg/models/secret_usage.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/models/secret_usage.go)
- [pkg/services/secret_service.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/services/secret_service.go)
- [pkg/services/addon_usage_service.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/services/addon_usage_service.go)
- [pkg/services/stack_service.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/services/stack_service.go)
- [pkg/stores/pgstore/addon_usage.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/stores/pgstore/addon_usage.go)
- [pkg/stores/pgstore/secret_usage.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/stores/pgstore/secret_usage.go)

Expected outcome:

- Delete protection and reverse lookup should use the new explicit wiring model.
- No dependency should exist only “in topology” without corresponding internal usage state where applicable.

### Phase 2: Runtime Resolution For Non-`env` Connection Kinds

The model and validation already know about more than `env`, but runtime behavior is still incomplete.

Implement:

- `volume_mount`
- `build_artifact_source`

Suggested file areas:

- [pkg/worker/stack/addon_env_reconciler.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/worker/stack/addon_env_reconciler.go)
- [pkg/models/stack_connection.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/models/stack_connection.go)
- [pkg/builders](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/builders)
- any stack CR builder code that currently consumes legacy mount/build-source fields

Important rule:

- Do not force unrelated direct config into connections if the spec explicitly kept it local. Only move behavior that the public API now models as a connection kind.

### Phase 3: Topology Completeness And Read Model Hardening

The topology endpoint exists, but it should become the dependable canvas read surface.

Improve:

- node coverage consistency across resource types
- stable labels/refs for canvas identity
- node metadata useful to the UI
- explicit handling for references that remain direct config but should still appear as dependency metadata

Suggested file areas:

- [pkg/models/stack_topology.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/models/stack_topology.go)
- [pkg/services/stack_topology.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/services/stack_topology.go)
- [pkg/presenters/stack_topology.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/presenters/stack_topology.go)

### Phase 4: Aggregate-Merge Validation For Future Child Resource CRUD

Connection CRUD already nudges us in this direction. The next agent should make this explicit and reusable for future `StackResource` child-resource mutations.

Implement/refactor toward:

- a service-layer materialization pattern:
  - load current persisted stack aggregate
  - apply requested child mutation in memory
  - validate the resulting aggregate
  - persist the minimal row-level change(s)

This matters most before independent `StackResource` CRUD is added.

Suggested file areas:

- [pkg/services/stack_service.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/services/stack_service.go)
- [pkg/validator/stack/stack_validator.go](/Users/asnaraya/projects/skysync/api-server/api-server/pkg/validator/stack/stack_validator.go)

### Phase 5: Independent `StackResource` CRUD

Once connection child-resource behavior is stable, the next larger structural move is to give `StackResource` the same treatment.

Not all of this needs to land in one go, but this is the direction:

- create/update/delete stack resources independently
- validate against post-merge desired aggregate
- enforce reference checks from connections to resources and from resources to volumes/secrets/etc

This phase should happen only after the validation/materialization approach is clean enough to support it.

## Testing Expectations

The next agent should preserve the current verification bar:

- `go test ./pkg/...`
- `go test ./test/int/... -run TestDoesNotExist`

And add focused tests for each phase:

- service/store tests for usage derivation
- worker tests for mount/build-source resolution
- topology service tests for projection behavior
- integration tests once new runtime or CRUD behavior is exposed publicly

## Known Constraints / Things Not To Revisit

- Do not reintroduce backward-compatibility layers for legacy interpolation unless the user explicitly asks.
- Do not move `depends_on` into `stack_connections` yet.
- Do not model build-time secrets or backup object-store refs as canvas-authored connections.
- Do not move connection ID generation back into app code.

## Suggested First Task For Next Agent

Start with **Phase 1: read connection usage from explicit connections**.

Reason:

- it operationalizes the new source of truth
- it improves deletion safety immediately
- it reduces divergence between topology, validation, and internal dependency tracking
- it does not require reopening the public API shape

## Recommended Skills For Next Session

- `superpowers:subagent-driven-development`
- `superpowers:verification-before-completion`
- `superpowers:receiving-code-review`
