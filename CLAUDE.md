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
| Integration tests (needs Docker/Kind) | `mage test:integration` |
| Lint Go | `golangci-lint run ./...` (or `mage lint`) |
| Format Go | `mage fmt` |
| Regenerate mocks | `make mocks` (mockgen via `go:generate`; add new packages to the target) |
| Regenerate OpenAPI Go client | `make generate` |
| Build frontend only | `make frontend` |
| Frontend dev / test / lint | `pnpm --prefix frontend dev` / `test` / `lint` |
| Frontend OpenAPI types/zod | `pnpm --prefix frontend generate:openapi-types` / `:openapi-zod` |

Notes:
- `mage dev:setup` is idempotent; reads/writes `.env` (creates from `.env_template`), writes cluster creds to `dev_env.yaml`.
- `mage build` skips `tsc -b` so unrelated frontend type errors don't block the binary; frontend's own `pnpm build` does run `tsc -b`.
- Integration tests need `TEST_KUBECONFIG` pointing at a Mage-created Kind cluster (`mage cluster:create` / `cluster:delete`); state cached in `~/.cache/stackdome-api-server/clusters/`.
- New DB migration: scaffold with `hack/create_migration.sh`; ordered files live in `pkg/db/migrations/`.
- Full end-to-end demo env: `hack/run_local.sh [stack.json]`.
- **Tests use Ginkgo.** All Go tests are written with Ginkgo v2 + Gomega (`Describe`/`It`/`Expect`), never bare `testing.T` test functions. Ginkgo allows exactly one `RunSpecs` per package — if a suite bootstrap (`*_suite_test.go` or any `RunSpecs` call) already exists, add `var _ = Describe(...)` blocks to new files instead of a second suite. Existing bare-`testing.T` tests may stay, but new tests never add to them.
- **Mocks:** Always use `go.uber.org/mock/gomock` + `mockgen`. Never hand-roll mock structs. For package-private interfaces, generate mocks in-package with `mockgen -source=<file>.go -destination=<file>_mock.go -package=<pkg>` and add a `//go:generate` directive to the source file. Generated mock files do NOT take a `_test` suffix — `<file>_mock.go`, never `<file>_mock_test.go`. Existing generated mocks live in `pkg/mocks/` (for exported interfaces) and `*_mock.go` files (for in-package interfaces); older `*_mock_test.go` files may be renamed when their package is next touched.
- **No magic strings.** Always use defined constants (model enums, annotation keys, state values, error codes) instead of raw string literals. This applies to both production code and tests — e.g. use `models.ReleaseStateReleased` not `"Released"`, `models.SkipClusterProvisioningAnnotation` not `"stack.stackdome.io/skip-cluster-provisioning"`. If no constant exists and one is needed, define it first. If a raw string is truly unavoidable, get explicit approval before using it.
- **No defensive programming.** Never write `if s.someValidator != nil { ... }` around work that must always happen (validation, authorization, persistence), and do NOT add explicit nil-check-and-panic guards in constructors either (`if spec.X == nil { panic(...) }`). A missing required dependency is a wiring bug: just use the dependency unconditionally and let the natural nil-pointer panic surface it — e2e tests catch mis-wiring. Nil-guards are only acceptable for genuinely optional behavior where nil substitutes a default implementation (e.g. `RegistryClients` → real client) — never where nil silently skips the work. Tests wire every seam with mocks instead of relying on skip-when-nil.
- **Comments: simple, short, useful.** Write them in plain language a tired reader can parse. Use bullet points when the comment covers more than one thing. Add a short example when the code calls for it (a tricky format, a non-obvious input). Be brief — nobody reads a wall of LLM-generated prose, and a long comment is usually a sign the code should be clearer instead.

  ```go
  // Bad: one dense paragraph restating the code.
  // recordResourceEvent records the observed resource state onto the active
  // release timeline. It runs only inside the StatusHash-change gate, so each
  // state transition is recorded at most once. A missing active release or a
  // recorder failure is non-fatal: the status update has already been persisted...

  // Good: says what matters, in bullets.
  //   - Runs only when StatusHash changed, so each state is recorded once.
  //   - TLS events fall back to the latest release: a cert can finish issuing
  //     after the release is done.
  //   - No release, or a failed record: log and move on. Status is already saved.
  ```
- **Don't comment for the sake of it. Code is the best comment.** A comment that restates the line below it is noise. Before writing one, try to make the comment unnecessary: a clearer name, a smaller function, an extracted helper, an early return. Comment only what the code genuinely cannot say — why a non-obvious choice was made, a constraint imposed from outside (an agent's behaviour, a k8s rule), a surprising ordering. Always strive for simplicity and readability.

  ```go
  // Noise — the code already says this.
  // increment the retry count
  retries++

  // Better — no comment needed, the name carries it.
  func currentGenCondition(cr *corev1alpha1.StackResource, condType string) *metav1.Condition

  // Worth keeping — the reason is invisible from the code.
  // Pre-0.6.11 agents set True on issuer discovery alone, so key on the reason.
  case cond.Status == metav1.ConditionTrue && cond.Reason == controllers.ReasonTLSReady:
  ```

## Agent skills

Superpowers is the workflow **spine** for all tasks: `superpowers:brainstorming` → spec (`dev-docs/specs/`), `superpowers:writing-plans` → plan (`dev-docs/plans/`), which is git-ignored so neither lands in this repo; `superpowers:executing-plans`/TDD/debugging per superpowers. The repo-local skills in `.claude/skills/` are permitted as below; some gate on a superpowers artifact, others are off-spine utilities. `tdd`, `write-a-skill`, `caveman` are intentionally absent — the superpowers / global-plugin equivalents are authoritative.

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
