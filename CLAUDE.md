# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Stackdome — self-hosted PaaS that deploys/manages workloads across multiple Kubernetes clusters (Heroku-like: REST API, web UI, managed Postgres, git builds). Go API server + embedded React SPA.

## Architecture

**Hub-and-spoke.** Single API Server (hub) holds state in PostgreSQL and exposes REST + UI. Each managed cluster runs a `stackdome-agent` operator (spoke, Kubebuilder) that reconciles Custom Resources into K8s workloads.

- **Write path:** User → API Server → writes a Kubernetes CR → Cluster Agent → K8s resources (Deployment/Service/Ingress).
- **Status path:** Agent updates CR status → `pkg/controllers/*` controller-runtime watchers observe → PostgreSQL updated → UI/API reflect state.

Backend layering (`pkg/`): `handlers` (HTTP) → `services` (business logic) → `stores`/`pgstore` (DB). `presenters` convert model↔API. `builders` construct K8s objects. `clustermanager` multiplexes per-cluster connections. `worker/` does async reconciliation (managed by `workermanager`). `resourceaccess` = Casbin RBAC. `auth` = JWT/cookie. `validator/` is per-resource-type input validation. Entrypoints in `cmd/` (`servecmd`, `migratecmd`, `server` = routes/middleware).

The REST contract is `config/openapi/stackdome_api.yaml`. Go client (`pkg/api/openapi`) and frontend types/zod clients are **generated** from it — edit the spec, not generated code.

Frontend (`frontend/`, React 19 + Vite + Tailwind v4 + Radix) builds into `pkg/web/dist/` and is `go:embed`-ed into the server binary — no separate frontend deploy.

## Commands

Build system is **Mage** (`magefile.go`, namespaced targets); `Makefile` wraps some. Use `mage` for dev.

| Task | Command |
|------|---------|
| Bootstrap local env (Postgres + Kind + RBAC) | `mage dev:setup` |
| DB migrations | `mage migrate` |
| Build + run API server | `mage run` |
| Tear down local env | `mage dev:teardown` |
| Backend unit tests | `mage test:unit` (or `make test`) |
| Single Go test | `go test ./pkg/<pkg>/ -run TestName` |
| Integration tests (needs Docker/Kind) | `make test-integration` |
| Focused integration test | `KEEP_CLUSTER=true KEEP_RESOURCES_ON_FAILURE=true FOCUS="Test Name" make test-integration` |
| Lint Go | `golangci-lint run ./...` (or `mage lint`) |
| Format Go | `mage fmt` |
| Regenerate mocks | `make mocks` (mockgen via `go:generate`) |
| Regenerate OpenAPI Go client | `make generate` |
| Build frontend only | `make frontend` |
| Frontend dev / test / lint | `pnpm --prefix frontend dev` / `test` / `lint` |
| Frontend OpenAPI types/zod | `pnpm --prefix frontend generate:openapi-types` / `:openapi-zod` |

Notes:
- `mage dev:setup` is idempotent; reads/writes `.env` (creates from `.env_template`), writes cluster creds to `dev_env.yaml`.
- `mage build` skips `tsc -b` so unrelated frontend type errors don't block the binary; frontend's own `pnpm build` does run `tsc -b`.
- **Running integration tests for new features or new tests:** Always use focused runs (`FOCUS="Test Name"`) for faster feedback instead of running the full suite. Use `KEEP_CLUSTER=true` to preserve the Kind cluster between runs (avoids 3-8min bootstrap). Use `KEEP_RESOURCES_ON_FAILURE=true` to skip resource cleanup on failure so you can inspect cluster state (`kubectl`) and DB to diagnose what went wrong. Recommended command:
  ```
  KEEP_CLUSTER=true KEEP_RESOURCES_ON_FAILURE=true FOCUS="My New Test" make test-integration
  ```
  `FOCUS` accepts a Ginkgo regex — use the `It()` description text or a `Context()` name to scope the run. Output is saved to `test/int/last-run.log`.
- New DB migration: scaffold with `hack/create_migration.sh`; ordered files live in `pkg/db/migrations/`.
- Full end-to-end demo env: `hack/run_local.sh [stack.json]`.

## Integration Test Patterns

When writing new integration tests, follow these conventions:

**File organization:**
- E2E tests that need a real cluster go in `test/int/*_e2e_test.go`
- API-only tests (no cluster interaction) go in `test/int/*_test.go`
- Fixtures (factory functions for OpenAPI objects) go in `test/int/shared/fixtures.go`
- CRUD helpers (create/get/update/delete via API client) go in `test/int/shared/helpers.go`
- Cluster inspection helpers (get deployments, services, CRs) go in `test/int/shared/cluster_helpers.go`

**Test structure:**
```go
var _ = Describe("Feature E2E", Ordered, func() {
    var client *openapi.APIClient
    var orgID string
    teamName := models.DefaultTeamName

    BeforeAll(func() {
        testEnv := GetEnvironment()
        client = testEnv.Client
        orgID = testEnv.OrgID
    })

    Context("Scenario", func() {
        It("should do something", func() {
            By("Creating a resource")
            // Use factory functions from shared/fixtures.go
            resource := shared.CreateSimpleStack("test-name")
            created := shared.CreateStack(client, orgID, teamName, resource)

            // Always register cleanup
            shared.DeferResourceCleanup(func() {
                shared.DeleteStack(client, orgID, teamName, created.GetId())
            })

            By("Verifying the result")
            // Assertions with gomega
        })
    })
})
```

**Key conventions:**
- Use `shared.DeferResourceCleanup()` (not raw `DeferCleanup`) for all resource cleanup — it respects `KEEP_RESOURCES_ON_FAILURE`
- Use `shared.ShouldSkipCleanup()` in custom `DeferCleanup` blocks that mix cleanup with debug logging
- Use `By("description")` to document each step — these appear in test output
- Use factory functions from `shared/fixtures.go` to create test objects — don't construct OpenAPI objects inline
- Add new CRUD helpers to `shared/helpers.go` following the existing pattern (call API, `Expect` no error, return result)
- Add new cluster inspection helpers to `shared/cluster_helpers.go`
- Use `GetEnvironment()` to access the shared test environment (client, cluster, org ID)
- Prefix test resource names with `test-` to make them identifiable in cluster output

**API-only tests (no cluster needed):**
For tests that only exercise the API layer (validation, CRUD, connections, topology) without needing real cluster reconciliation, use the `CreateSkipProvisioningStack()` fixtures. These add the `stack.stackdome.io/skip-cluster-provisioning` annotation which tells the stack worker to mark the stack as Ready immediately without creating CRs in the cluster. This makes tests much faster and avoids cluster dependencies.

**Running tests during development:**
```
KEEP_CLUSTER=true KEEP_RESOURCES_ON_FAILURE=true FOCUS="My New Test" make test-integration
```
Always use focused runs when developing. See `test/int/README.md` for full details.

## Agent skills

Superpowers is the workflow **spine** for all tasks: `superpowers:brainstorming` → spec (`docs/superpowers/specs/`), `superpowers:writing-plans` → plan (`docs/superpowers/plans/`), `superpowers:executing-plans`/TDD/debugging per superpowers. The repo-local skills in `.claude/skills/` are permitted as below; some gate on a superpowers artifact, others are off-spine utilities. `tdd`, `write-a-skill`, `caveman` are intentionally absent — the superpowers / global-plugin equivalents are authoritative.

| Skill | Use only | Gate |
|---|---|---|
| `ingest-design-bundle` | Claude Design URL → design-ref | none (inbound intake) |
| `grill-with-docs` | harden a spec/plan/design, sharpen glossary, extract ADRs | none (works best with a spec/plan) |
| `grill-me` | stress-test a plan/design in conversation | none (off-spine) |
| `handoff` | compact session → handoff doc | none (off-spine) |
| `to-prd` | spec → PRD issue | approved spec must exist |
| `to-issues` | plan → vertical-slice issues | plan must exist |
| `triage` | inbound external issue funnel | none (off-spine) |
| `zoom-out` | manual orientation | none (manual only) |

### Issue tracker

Issues and PRDs live in the `Stackdome/stackdome` GitHub repo (via the `gh` CLI). See `docs/agents/issue-tracker.md`.

### Triage labels

Canonical triage roles map 1:1 to label strings (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Multi-context: `CONTEXT-MAP.md` at root points to per-context `CONTEXT.md` files (backend root + `frontend/`; agent deferred). See `docs/agents/domain.md`.
