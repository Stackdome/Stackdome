# E2E Test Gap Analysis: Stacks & PostgreSQL Addon

**Date:** 2026-04-22

## Test Files

| File | Scope | Test Count |
|------|-------|------------|
| `test/int/stack_test.go` | Stack API CRUD + validation | 11 |
| `test/int/stack_e2e_test.go` | Stack cluster integration | 9 (1 skipped, 1 conditional) |
| `test/int/postgres_addon_test.go` | Postgres addon CRUD + validation | 11 |
| `test/int/postgres_addon_e2e_test.go` | Postgres addon cluster integration | 6 |
| `test/int/object_store_test.go` | Object store CRUD + validation | 21 |

---

## Stack API Endpoint Coverage

### Routes (from `cmd/server/routes.go:165-183`)

| # | Method | Endpoint | Unit | E2E | Notes |
|---|--------|----------|------|-----|-------|
| 1 | POST | `/stacks` | Yes | Yes | |
| 2 | GET | `/stacks` | Yes | - | |
| 3 | GET | `/stacks/current` | Yes | - | |
| 4 | GET | `/stacks/{id}` | Yes | Yes (polling) | |
| 5 | PUT | `/stacks/{id}` | Yes | Yes | |
| 6 | DELETE | `/stacks/{id}` | Yes | Yes | |
| 7 | GET | `/stacks/{id}/logs` | **No** | **No** | SSE streaming |
| 8 | GET | `/stacks/{id}/metrics` | **No** | **No** | Supports SSE |
| 9 | GET | `/stacks/{id}/resources` | **No** | **No** | |
| 10 | GET | `/stacks/{id}/resources/{name}` | **No** | **No** | |
| 11 | GET | `/stacks/{id}/resources/{name}/logs` | **No** | **No** | SSE streaming |
| 12 | GET | `/stacks/{id}/resources/{name}/metrics` | **No** | **No** | Supports SSE |
| 13 | GET | `/stacks/{id}/resources/{name}/builds` | **No** | **No** | |
| 14 | GET | `/stacks/{id}/builds` | **No** | Partial | Only in build-from-source test |
| 15 | GET | `/stacks/{id}/builds/{build_id}` | **No** | **No** | |

**Endpoint coverage: 6/15 (40%)**

### What's Tested (Stack)

**Unit tests (`stack_test.go`):**
- Create minimal stack, get by ID, list by org, list by user, update, delete
- Validation: empty name, no resources, missing build/image spec, duplicate resource names, duplicate stack names (409)

**E2E tests (`stack_e2e_test.go`):**
- Create stack -> CR appears in cluster -> reaches Ready -> StackResource Available -> Deployment + Service verified
- Delete stack -> CR removed from cluster -> 404 from API
- Multi-resource stack (backend + frontend) with env var interpolation
- Depends-on ordering (creates both resources, but doesn't verify actual ordering)
- Multiple ports and env vars -> verified on Deployment and Service
- Init container (skipped - cluster-agent bug)
- Update propagation -> CR spec updated -> stack returns to Ready
- Stack with PostgresAddon -> postgres env vars injected into Deployment
- Build from private git repo (conditional on GITHUB_TOKEN) -> image build, registry, ingress verified

### What's Missing (Stack)

**Untested endpoints:**

1. **Stack Resources API** (`GET /resources`, `GET /resources/{name}`) - Resources are verified via cluster-side CR checks but the API itself is never called. No test validates the API returns correct resource data, status, or conditions.

2. **Logs streaming** (`GET /stacks/{id}/logs`, `GET /resources/{name}/logs`) - Core observability feature with SSE support. Completely untested.

3. **Metrics** (`GET /stacks/{id}/metrics`, `GET /resources/{name}/metrics`) - Both stack-level and resource-level metrics. Completely untested.

4. **Image Build API** - `GET /builds/{build_id}` never tested. `ListByResourceName` never tested. `ListByStackID` only partially exercised in build-from-source test.

**Untested features:**

5. **Stateful resources** - No test uses `stateful: true` to create a StatefulSet instead of a Deployment.

6. **Volume mounts** - No test creates a stack with volume mounts. The Volume API (`POST/GET/DELETE /volumes`) is also untested in the stack context.

7. **Init containers** - Test exists but is skipped due to known cluster-agent bug (uses main image instead of InitSpec.ImageSpec).

8. **Public port exposure from image-based stack** - `expose_to_public` is only tested in build-from-source. No test verifies ingress creation from a simple image-based stack with exposed ports.

9. **Depends-on ordering verification** - Test creates a depends_on stack but only asserts both resources exist. No verification of actual ordering (timestamps, readiness sequence).

10. **Deployment failure scenarios** - No test for bad image, OOM, crash loop, or any failure that would leave a stack in Error/Failed state.

11. **Secrets as environment variables** - Secrets API is used for git credentials in build-from-source, but no test injects secrets as env vars into a normal stack.

12. **Custom command/args** - No test uses `command` or `args` fields to override container entrypoint.

13. **Resource limits (CPU/memory)** - No test sets or verifies resource requests/limits on stack resources.

---

## PostgreSQL Addon API Endpoint Coverage

### Routes (from `cmd/server/routes.go:188-203`)

| # | Method | Endpoint | Unit | E2E | Notes |
|---|--------|----------|------|-----|-------|
| 1 | POST | `/addons/postgres` | Yes | Yes | |
| 2 | GET | `/addons/postgres` | Yes | - | |
| 3 | GET | `/addons/postgres/{id}` | Yes | Yes (polling) | |
| 4 | PUT | `/addons/postgres/{id}` | Yes | Yes | |
| 5 | DELETE | `/addons/postgres/{id}` | Yes | Yes | |
| 6 | POST | `/addons/postgres/{id}/actions/backup` | - | Yes | |
| 7 | POST | `/addons/postgres/{id}/actions/fence` | **No** | **No** | |
| 8 | POST | `/addons/postgres/{id}/actions/hibernate` | **No** | **No** | |
| 9 | GET | `/addons/postgres/{id}/backups` | - | Yes | |
| 10 | GET | `/addons/postgres/{id}/credentials/{db}` | - | Yes | |

**Endpoint coverage: 8/10 (80%)**

### What's Tested (PostgreSQL Addon)

**Unit tests (`postgres_addon_test.go`):**
- Create minimal addon, create with resources + database, get by ID, list, update (instances + databases), delete
- Validation: invalid storage size, zero instances, invalid postgres version
- Backup configuration (creates addon with backup config, asserts `HasBackup()`)
- HA config (3 instances with placement topology key)

**E2E tests (`postgres_addon_e2e_test.go`):**
- Create addon -> PostgresCluster CR appears in cluster with correct spec + labels
- Full lifecycle: create -> Ready -> ConnectionInfo populated -> ClusterReady condition True -> CR verified -> JIT credentials returned
- Update propagation: change instances from 1 to 3 -> CR updated -> addon returns to Ready
- Deletion cleanup: delete via API -> CR removed from cluster -> 404 from API
- Failure reporting: invalid storage class -> status reflects non-Ready state
- Backup + WAL archiving: create S3 secret + ObjectStore + addon with backup -> verify WAL archiving condition -> trigger backup -> verify backup reaches "completed" phase

### What's Missing (PostgreSQL Addon)

**Untested endpoints:**

1. **Fence action** (`POST /actions/fence`) - Completely untested. No verification that fencing prevents database access or that unfencing restores it.

2. **Hibernate action** (`POST /actions/hibernate`) - Completely untested. No verification of hibernation/wake cycle or that resources are freed during hibernation.

**Untested features:**

3. **Actual database connectivity** - `ConnectToPostgres()` helper exists in `test/int/shared/cluster_helpers.go` but is unused. No test verifies you can connect and run queries with the returned credentials.

4. **Superuser credentials** - `GetCredentials` supports `?superuser=true` query param but this is never tested.

5. **Database extensions** - Fixture support exists for extensions (e.g., pgvector) but no test verifies extension installation in the cluster.

6. **Multiple databases in E2E** - Update unit test adds 2 databases, but no E2E test verifies multiple databases are actually created and accessible in the cluster.

7. **Backup restore** - Backup creation and recording is tested, but restoring from a backup (BootstrapSpec RecoverySpec) is not tested.

8. **Scheduled backups** - No test verifies cron-based scheduled backups fire automatically.

9. **Storage expansion** - No test for increasing storage size on an existing addon.

10. **PostgreSQL version upgrade** - No test for major/minor version changes.

11. **Replica configuration** - No test verifies synchronous replica behavior.

12. **Addon deletion side effects** - No test for what happens to stacks that reference a deleted addon.

---

## Object Store API Endpoint Coverage

### Routes (from `cmd/server/routes.go:208-212`)

| # | Method | Endpoint | Unit | E2E | Notes |
|---|--------|----------|------|-----|-------|
| 1 | POST | `/object-stores` | Yes | - | |
| 2 | GET | `/object-stores` | Yes | - | |
| 3 | GET | `/object-stores/{id}` | Yes | - | |
| 4 | PUT | `/object-stores/{id}` | Yes | - | |
| 5 | DELETE | `/object-stores/{id}` | Yes | - | |

**Endpoint coverage: 5/5 (100%)**

Tested: S3, Azure, and GCS credential types. Custom retention policies. Extensive validation (21 tests covering invalid names, bad retention format, non-existent secret refs, invalid key refs, missing credentials, name change on update, invalid destination paths for each provider).

Missing: No test for deleting an object store that is in use by a postgres addon (should fail with 409 Conflict).

---

## Coverage Summary

| Area | Endpoints Tested | Total | Coverage | Functional Depth |
|------|-----------------|-------|----------|-----------------|
| Stack Core CRUD | 6 | 6 | 100% | Good |
| Stack Resources API | 0 | 2 | **0%** | None |
| Stack Observability | 0 | 4 | **0%** | None |
| Stack Builds API | 1 | 3 | 33% | Shallow |
| Postgres Addon CRUD | 5 | 5 | 100% | Good |
| Postgres Actions | 1 | 3 | **33%** | Backup only |
| Postgres Sub-resources | 2 | 2 | 100% | Shallow |
| Object Store | 5 | 5 | 100% | Good |

**Overall: 20/30 endpoints have at least one test (67%).**

Endpoint coverage alone overstates readiness. Streaming (logs/metrics), stateful workloads, volume mounts, fencing, hibernation, and database connectivity are all untested functional areas.

---

## Priority Recommendations

### P0 - High value, moderate effort
- Stack Resources API (list + get by name)
- Fence and Hibernate actions for postgres addon
- Database connectivity verification using existing `ConnectToPostgres()` helper

### P1 - High value, higher effort
- Logs streaming (at least basic SSE connection test)
- Stateful stack resources (StatefulSet creation)
- Volume mounts in stacks
- Superuser credentials

### P2 - Important but complex
- Backup restore (requires existing backup + new addon with RecoverySpec)
- Public port exposure + ingress from image-based stacks
- Deployment failure scenarios
- Object store deletion while in use (409)
- Multiple databases with extensions in E2E
