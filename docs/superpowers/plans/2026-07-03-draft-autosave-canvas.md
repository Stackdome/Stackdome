# Draft Concept + Debounced Autosave Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Canvas edits autosave with a debounce via thin per-entity endpoints; the UI reframes around the persistent draft (`staged` lifecycle phase) with Deploy as the primary action, plus revert-to-deployed and delete-stack flows.

**Architecture:** Three layers. (1) Pure calculations in `frontend/src/pages/stacks/lib/draft-sync/` translate the edit-session draft and the server's stack into an ordered list of thin-endpoint operations. (2) A `useDraftSync` engine hook debounces, executes ops single-flight, refetches, and rebases the session baseline. (3) `CanvasEditorShell`/detail-page rework removes Save, adds an autosave indicator, DRAFT pill, Deploy-flush, revert, and delete. One new backend endpoint (`POST /stacks/{id}/volumes`) closes the volume-association gap.

**Tech Stack:** React 19 + Vite + Vitest (jsdom) + zod; Go (gorilla/mux, gomock) for the backend slice.

**Spec:** `docs/superpowers/specs/2026-07-02-draft-autosave-canvas-design.md`

## Global Constraints

- Brand design system only: `index.css` tokens + `branded/`/`ui/` primitives. No raw hex, no off-scale type.
- No third-party-PaaS names ("Railway" etc.) in code, copy, or commits.
- No magic strings — use defined constants (backend: model enums/error helpers; frontend: `draft-sync/constants.ts`).
- Mocks: `go.uber.org/mock/gomock` + `mockgen` only. Never hand-roll mock structs. New exported-interface mocks go in `pkg/mocks/` and get added to the `make mocks` target.
- Pure calculations stay side-effect-free and unit-tested; components stay views; the engine hook is the only new action layer.
- Whole-stack `PUT` is used ONLY by revert (Task 6). Autosave cycles must never issue it.
- Op ordering invariant (correctness, not style — backend does not cascade connections on resource delete): createVolume → createResource → updateResource → deleteConnection → updateConnection → createConnection → deleteResource.
- Existing frontend suite (518 tests) stays green: `pnpm --prefix frontend test -- --run`. Backend: `mage test:unit`, `golangci-lint run ./...`.
- Frontend type/lint gates: `pnpm --prefix frontend exec tsc -b` and `pnpm --prefix frontend lint`.
- Commit after every task (small, conventional-commit messages).

---

### Task 0: Backend thin volume endpoint — `POST /stacks/{id}/volumes`

`POST /volumes` creates a volume but never associates it with a stack (association only happens inside the whole-stack PUT via `InternalCreateWithTx`). Add a stack-scoped create that does both in one transaction, mirroring the `CreateStackConnection` thin-endpoint pattern.

**Files:**
- Modify: `pkg/services/stack_service.go` (interface + method)
- Modify: `pkg/handlers/stack_handler.go` (new `CreateVolume` handler)
- Modify: `cmd/server/routes.go` (route next to line 257 `GET /stacks/{id}/volumes`)
- Modify: `config/openapi/stackdome_api.yaml` (add POST to a new `/stacks/{id}/volumes` path entry)
- Modify: `Makefile` (add `VolumeService` to the `mocks` target if not present)
- Create: `pkg/mocks/mock_volume_service.go` (generated)
- Test: `pkg/services/stack_service_test.go` (or a new `stack_volume_create_test.go` in package `services`)

**Interfaces:**
- Consumes: `VolumeService.InternalCreateWithTx(ctx, stack *models.Stack, volume *models.Volume)` (exists, `pkg/services/volume_service.go:98`), `BackgroundJobEnqueuer.Enqueue` (exists — see `UpdateStack`'s new-volume enqueue at `stack_service.go:261-272`).
- Produces: `StackService.CreateStackVolume(ctx context.Context, stackID string, volume *models.Volume) (*models.Volume, *errors.ServiceError)`; HTTP `POST /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/volumes` → 201 `Volume`.

- [ ] **Step 1: Generate the VolumeService mock**

VolumeService is an exported interface consumed across packages → mock belongs in `pkg/mocks`:

```bash
mockgen -source=pkg/services/volume_service.go -destination=pkg/mocks/mock_volume_service.go -package=mocks
```

Add the same line to the `mocks` target in `Makefile` (match the existing entries' style).

- [ ] **Step 2: Write the failing service test**

In package `services` (in-package test, struct-literal service with mocks — mirror `pkg/services/stack_resource_service_test.go`'s `TestStackResourceService_Restart` harness):

```go
func TestStackService_CreateStackVolume(t *testing.T) {
	ctx := context.Background()
	stackID := "stack-123"
	teamID := "team-456"
	stack := &models.Stack{ID: stackID, TeamID: teamID, Volumes: []*models.Volume{{ID: "v-0", Name: "existing-data"}}}
	newVolume := &models.Volume{Name: "web-data"}

	t.Run("creates, associates and enqueues the volume", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockVolumeService := mocks.NewMockVolumeService(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)
		mockEnqueuer := mocks.NewMockBackgroundJobEnqueuer(ctrl)

		svc := &stackService{
			stackStore:            mockStackStore,
			volumeService:         mockVolumeService,
			permissions:           mockPermissions,
			BackgroundJobEnqueuer: mockEnqueuer,
		}

		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		mockPermissions.EXPECT().Check(ctx, teamID, auth.ResourceStacks, stackID, auth.ActionRead).Return(nil)
		mockPermissions.EXPECT().Check(ctx, teamID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)
		mockStackStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})
		created := &models.Volume{ID: "v-1", Name: "web-data"}
		mockVolumeService.EXPECT().InternalCreateWithTx(ctx, stack, newVolume).Return(created, nil)
		mockEnqueuer.EXPECT().Enqueue(&models.Volume{ID: "v-1"}).Return(nil)

		got, serr := svc.CreateStackVolume(ctx, stackID, newVolume)
		assert.Nil(t, serr)
		assert.Equal(t, "v-1", got.ID)
	})

	t.Run("rejects a duplicate volume name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)
		svc := &stackService{stackStore: mockStackStore, permissions: mockPermissions}

		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		mockPermissions.EXPECT().Check(ctx, teamID, auth.ResourceStacks, stackID, auth.ActionRead).Return(nil)
		mockPermissions.EXPECT().Check(ctx, teamID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)

		_, serr := svc.CreateStackVolume(ctx, stackID, &models.Volume{Name: "existing-data"})
		assert.NotNil(t, serr)
		assert.Equal(t, errors.ErrorConflict, serr.Code)
	})
}
```

Adapt exact field names / permission-check call shapes to what `GetStack` and `UpdateStack` actually do in `stack_service.go:225-235` (read check lives inside `GetStack`; write check is explicit). If `GetStack` performs the read check itself, drop the duplicate `ActionRead` expectation and mirror reality. If the `errors` package uses a different conflict discriminator than `errors.ErrorConflict`, use the real one (see `errors.Conflict` usage at `stack_service.go:337`).

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./pkg/services/ -run TestStackService_CreateStackVolume -v`
Expected: FAIL — `svc.CreateStackVolume undefined`.

- [ ] **Step 4: Implement the service method**

Add to the `StackService` interface (next to the connection methods):

```go
CreateStackVolume(ctx context.Context, stackID string, volume *models.Volume) (*models.Volume, *errors.ServiceError)
```

Implementation (place next to `CreateStackConnection`):

```go
// CreateStackVolume creates a volume and associates it with the stack in one
// transaction — the thin counterpart of the whole-stack PUT's volume sync.
func (s *stackService) CreateStackVolume(ctx context.Context, stackID string, volume *models.Volume) (*models.Volume, *errors.ServiceError) {
	stack, err := s.GetStack(ctx, stackID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	for _, existing := range stack.Volumes {
		if existing.Name == volume.Name {
			return nil, errors.Conflict("a volume named '%s' already exists in this stack", volume.Name)
		}
	}

	var created *models.Volume
	txErr := s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		var serr *errors.ServiceError
		created, serr = s.volumeService.InternalCreateWithTx(ctx, stack, volume)
		return serr
	})
	if txErr != nil {
		return nil, txErr
	}

	// Mirror UpdateStack's behavior for volumes new to the stack: enqueue the
	// background job that provisions the volume in the cluster.
	if enqErr := s.BackgroundJobEnqueuer.Enqueue(&models.Volume{ID: created.ID}); enqErr != nil {
		return nil, errors.GeneralError("failed to enqueue volume '%s': %s", created.ID, enqErr.Error())
	}
	return created, nil
}
```

Match the receiver's actual field names (`BackgroundJobEnqueuer` embedding/casing) to the struct definition.

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./pkg/services/ -run TestStackService_CreateStackVolume -v`
Expected: PASS (both subtests).

- [ ] **Step 6: Add the handler**

In `pkg/handlers/stack_handler.go`, next to `CreateConnection` (line 94) — same `handlerConfig` idiom, with volume validation like `volumeHandler.Create` (`pkg/handlers/volume_handler.go:67-97`):

```go
func (h *stackHandler) CreateVolume(w http.ResponseWriter, r *http.Request) {
	var volume openapi.Volume
	cfg := &handlerConfig{
		MarshalInto: &volume,
		Validate:    validation.ValidateVolume(&volume),
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]
			obj, err := h.stackService.CreateStackVolume(r.Context(), stackID, presenters.ConvertVolume(&volume))
			if err != nil {
				return nil, err
			}
			return presenters.PresentVolume(obj, true), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}
```

Check `handlerConfig`'s field names for validation (the connection handlers skip validation; `volumeHandler.Create` uses positional struct literal `{&ws, validation.ValidateVolume(&ws), func...}` — use whichever form compiles against the struct definition in `pkg/handlers/`).

- [ ] **Step 7: Register the route**

In `cmd/server/routes.go`, directly above the existing stack-scoped volumes GET (line 257):

```go
teamResourceRouter.HandleFunc("/stacks/{id}/volumes", stackHandler.CreateVolume).Methods(http.MethodPost)
```

- [ ] **Step 8: OpenAPI spec + regenerate clients**

In `config/openapi/stackdome_api.yaml`, add a new path entry (place near the connection paths for stacks; copy the response envelope style of the team-scoped volume POST at line 1930):

```yaml
  /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/volumes:
    post:
      summary: Create a volume and associate it with the stack
      security:
        - Bearer: []
      parameters:
        - $ref: '#/components/parameters/org_id'
        - $ref: '#/components/parameters/team_name'
        - $ref: '#/components/parameters/id'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Volume'
      responses:
        '201':
          description: Volume created and associated with the stack
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Volume'
        '400':
          description: Invalid request payload
        '401':
          description: Unauthorized
        '403':
          description: Forbidden
        '404':
          description: Stack not found
        '409':
          description: A volume with this name already exists in the stack
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '500':
          description: Internal server error
```

Regenerate:

```bash
make generate
make mocks
pnpm --prefix frontend generate:openapi-types
pnpm --prefix frontend generate:openapi-zod
```

- [ ] **Step 9: Full backend gates**

Run: `mage test:unit && golangci-lint run ./...`
Expected: PASS / no new findings.

- [ ] **Step 10: Commit**

```bash
git add pkg/services pkg/handlers pkg/mocks cmd/server config/openapi Makefile frontend/src/api/types frontend/src/api/zod-schemas.ts
git commit -m "feat(api): add thin stack-scoped volume create endpoint"
```

---

### Task 1: Pure draft-sync core (calculations + tests)

**Files:**
- Create: `frontend/src/pages/stacks/lib/draft-sync/constants.ts`
- Create: `frontend/src/pages/stacks/lib/draft-sync/server-state.ts`
- Create: `frontend/src/pages/stacks/lib/draft-sync/desired-state.ts`
- Create: `frontend/src/pages/stacks/lib/draft-sync/ops.ts`
- Modify: `frontend/src/pages/stacks/schemas/form-schema.ts` (export `FormStackResourceSchema`, `convertFormVolumeToApiVolume`; extract `prepareFormResourceForApi`)
- Test: `frontend/src/pages/stacks/lib/draft-sync/tests/server-state.test.ts`
- Test: `frontend/src/pages/stacks/lib/draft-sync/tests/desired-state.test.ts`
- Test: `frontend/src/pages/stacks/lib/draft-sync/tests/ops.test.ts`

**Interfaces:**
- Consumes: `EditSessionDraft` (`hooks/use-stack-edit-session.ts:16`), `deepEqual` + `ResourceArr`/`VolumeArr` (`lib/stack-diff.ts`), `splitEnvRows`/`buildDesiredConnections`/`FormEnvRow` (`lib/connection-mapping.ts`), `StackResourceUpdateRequest`/`VolumeUpdateRequest`/`Stack` (`api/stacks.ts`), `StackConnection` (`api/connections.ts`).
- Produces (later tasks rely on these EXACT names):
  - `constants.ts`: `DEBOUNCE_IDLE_MS`, `DEBOUNCE_MAX_WAIT_MS`, `RETRY_BASE_MS`, `RETRY_MAX_MS`, `STICKY_FAILURE_THRESHOLD`, `SYNC_STATUS`, `type SyncStatus`
  - `server-state.ts`: `interface ServerStackState`, `connectionIdentityKey(c: StackConnection): string`, `serverStateFromStack(stack: Stack): ServerStackState`, `cleanServerResource(r: StackResource): StackResourceUpdateRequest`
  - `desired-state.ts`: `interface DesiredStackState`, `buildDesiredState(draft: EditSessionDraft): DesiredStackState`
  - `ops.ts`: `type SyncOp`, `computeSyncOps(server: ServerStackState, desired: DesiredStackState): SyncOp[]`

- [ ] **Step 1: form-schema exports + extraction**

In `form-schema.ts`:
1. Extract the per-resource pre-processing block from `convertFormStackToApiStack` (lines 418-451: `source_revision` normalization + secret-ref injection + `convertFormResourceToApiResource`) into:

```ts
// Prepare one form resource for the API: normalize git source_revision, attach
// selected secrets, then strip UI-only fields.
function prepareFormResourceForApi(resource: FormStackResourceData): StackResourceUpdateRequest {
  if (resource.build_spec) {
    const gitRepoRev = resource.build_spec.source_revision?.git_repo_revision;
    resource.build_spec.source_revision = {
      volume_source_revision: undefined,
      git_repo_revision: gitRepoRev,
    };
  }
  const resourceWithSecrets = { ...resource };
  if (resource.sourceType === 'image' && resource.useImageSecret && resource.selectedImageSecretId) {
    resourceWithSecrets.image_spec = {
      image: resourceWithSecrets.image_spec?.image || '',
      ...resourceWithSecrets.image_spec,
      pull_secret: { secret_id: resource.selectedImageSecretId },
    };
  }
  if (resource.sourceType === 'git' && resource.useGitSecret && resource.selectedGitSecretId) {
    if (resourceWithSecrets.build_spec?.source_context?.git_repo) {
      resourceWithSecrets.build_spec.source_context.git_repo.git_secret = { secret_id: resource.selectedGitSecretId };
    }
  }
  return convertFormResourceToApiResource(resourceWithSecrets);
}
```

2. `convertFormStackToApiStack` calls `prepareFormResourceForApi` per valid resource (behavior identical).
3. Add `FormStackResourceSchema`, `convertFormVolumeToApiVolume`, `prepareFormResourceForApi` to the export block (lines 497-505).

Run: `pnpm --prefix frontend test -- --run` → all existing tests still green (pure refactor).

- [ ] **Step 2: constants.ts**

```ts
/** Timing + status constants for the draft autosave engine. */
export const DEBOUNCE_IDLE_MS = 1200;
export const DEBOUNCE_MAX_WAIT_MS = 5000;
export const RETRY_BASE_MS = 1000;
export const RETRY_MAX_MS = 30000;
export const STICKY_FAILURE_THRESHOLD = 3;

export const SYNC_STATUS = {
  idle: "idle",
  saving: "saving",
  saved: "saved",
  error: "error",
} as const;
export type SyncStatus = (typeof SYNC_STATUS)[keyof typeof SYNC_STATUS];
```

- [ ] **Step 3: Write failing tests for server-state.ts**

```ts
import { describe, it, expect } from "vitest";
import { connectionIdentityKey, serverStateFromStack } from "../server-state";
import type { Stack } from "@/api/stacks";

describe("connectionIdentityKey", () => {
  it("keys an addon connection by kind, endpoints and config", () => {
    const key = connectionIdentityKey({
      kind: "env",
      from: { type: "addon/postgres", id: "ad-1" },
      to: { type: "stack_resource", name: "web" },
      config: { database: "app", superuser: false },
    });
    expect(key).toBe("env|addon/postgres:ad-1|stack_resource:web|db:app");
  });

  it("distinguishes superuser from database-scoped addon connections", () => {
    const base = {
      kind: "env" as const,
      from: { type: "addon/postgres", id: "ad-1" },
      to: { type: "stack_resource", name: "web" },
    };
    expect(connectionIdentityKey({ ...base, config: { superuser: true } })).not.toBe(
      connectionIdentityKey({ ...base, config: { database: "app", superuser: false } }),
    );
  });

  it("keys secret and resource sources without config", () => {
    expect(
      connectionIdentityKey({ kind: "env", from: { type: "secret", id: "s-1" }, to: { type: "stack_resource", name: "web" } }),
    ).toBe("env|secret:s-1|stack_resource:web|");
  });

  it("ignores mappings — same identity with different mappings collides", () => {
    const a = { kind: "env" as const, from: { type: "secret", id: "s-1" }, to: { type: "stack_resource", name: "web" }, mappings: [{ target: { type: "env", name: "A" }, value: { output: "k" } }] };
    const b = { ...a, mappings: [{ target: { type: "env", name: "B" }, value: { output: "k2" } }] };
    expect(connectionIdentityKey(a)).toBe(connectionIdentityKey(b));
  });
});

describe("serverStateFromStack", () => {
  const stack = {
    id: "st-1",
    name: "demo",
    spec: {
      stack_resources: [
        { id: "r-1", stack_id: "st-1", revision: 3, name: "web", image_spec: { image: "nginx:1" }, status: { state: "Ready" }, volume_mounts: [{ source_volume_name: "web-data", target_path: "/data", stack_resource_id: "r-1", source_volume_type: "pvc" }] },
      ],
      volumes: [{ id: "v-1", name: "web-data", spec: { size: "1Gi", access_mode: "ReadWriteOnce" }, status: {} }],
      connections: [{ id: "c-1", kind: "env", from: { type: "secret", id: "s-1" }, to: { type: "stack_resource", name: "web" }, mappings: [{ target: { type: "env", name: "TOKEN" }, value: { output: "token" } }] }],
    },
  } as unknown as Stack;

  it("indexes resources by name with read-only fields stripped", () => {
    const s = serverStateFromStack(stack);
    const web = s.resourcesByName.get("web")!;
    expect(web).toBeDefined();
    expect((web as Record<string, unknown>).id).toBeUndefined();
    expect((web as Record<string, unknown>).status).toBeUndefined();
    expect(web.volume_mounts?.[0]).toEqual({ source_volume_name: "web-data", target_path: "/data" });
  });

  it("indexes volumes by name and maps ids", () => {
    const s = serverStateFromStack(stack);
    expect(s.volumeIdByName.get("web-data")).toBe("v-1");
    expect((s.volumesByName.get("web-data") as Record<string, unknown>).id).toBeUndefined();
  });

  it("indexes connections by identity key, retaining the server id", () => {
    const s = serverStateFromStack(stack);
    const entry = s.connections.get("env|secret:s-1|stack_resource:web|")!;
    expect(entry.id).toBe("c-1");
    expect(entry.conn.mappings).toHaveLength(1);
  });
});
```

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/lib/draft-sync/tests/server-state.test.ts`
Expected: FAIL — module not found.

- [ ] **Step 4: Implement server-state.ts**

```ts
import type { Stack, StackResource, StackResourceUpdateRequest, Volume, VolumeUpdateRequest } from "@/api/stacks";
import type { StackConnection } from "@/api/connections";

/**
 * The engine's mirror of what the server currently holds for a stack, indexed
 * for diffing. Server ids live here and ONLY here — form state stays id-free.
 */
export interface ServerConnectionEntry {
  id?: string;
  conn: StackConnection;
}

export interface ServerStackState {
  resourcesByName: Map<string, StackResourceUpdateRequest>;
  volumeIdByName: Map<string, string>;
  volumesByName: Map<string, VolumeUpdateRequest>;
  connections: Map<string, ServerConnectionEntry>;
}

type NodeRef = { type?: string; id?: string; name?: string } | undefined;

function nodeKey(n: NodeRef): string {
  if (!n) return "";
  return `${n.type ?? ""}:${n.id ?? n.name ?? ""}`;
}

/**
 * Content identity of a connection — everything except its mappings. Mirrors
 * the backend's uniqueness check (kind + from + to + config discriminator);
 * mapping changes are updates to the same connection, identity changes are a
 * different connection.
 */
export function connectionIdentityKey(c: StackConnection): string {
  const cfg = c.config as { database?: string; superuser?: boolean } | undefined;
  const cfgKey = cfg ? (cfg.superuser ? "superuser" : `db:${cfg.database ?? ""}`) : "";
  return [c.kind ?? "", nodeKey(c.from), nodeKey(c.to), cfgKey].join("|");
}

/** Strip server-computed fields so a server resource compares against form-derived ones. */
export function cleanServerResource(r: StackResource): StackResourceUpdateRequest {
  const { id, stack_id, revision, status, outputs, ...rest } = r as StackResource & { outputs?: unknown };
  void id; void stack_id; void revision; void status; void outputs;
  const volume_mounts = rest.volume_mounts?.map((m) => {
    const { stack_resource_id, source_volume_type, ...mount } = m as Record<string, unknown>;
    void stack_resource_id; void source_volume_type;
    return mount;
  });
  return { ...rest, volume_mounts } as StackResourceUpdateRequest;
}

function cleanServerVolume(v: Volume): VolumeUpdateRequest {
  const { id, status, ...rest } = v as Volume & { status?: unknown };
  void id; void status;
  return rest as VolumeUpdateRequest;
}

export function serverStateFromStack(stack: Stack): ServerStackState {
  const resourcesByName = new Map<string, StackResourceUpdateRequest>();
  for (const r of stack.spec?.stack_resources ?? []) {
    if (r.name) resourcesByName.set(r.name, cleanServerResource(r));
  }
  const volumeIdByName = new Map<string, string>();
  const volumesByName = new Map<string, VolumeUpdateRequest>();
  for (const v of stack.spec?.volumes ?? []) {
    if (!v.name) continue;
    if (v.id) volumeIdByName.set(v.name, v.id);
    volumesByName.set(v.name, cleanServerVolume(v));
  }
  const connections = new Map<string, ServerConnectionEntry>();
  for (const c of stack.spec?.connections ?? []) {
    connections.set(connectionIdentityKey(c), { id: c.id, conn: c });
  }
  return { resourcesByName, volumeIdByName, volumesByName, connections };
}
```

Run the test file again. Expected: PASS. Adjust field access if the generated `StackConnection` type differs (check `frontend/src/api/types/openapi.d.ts`).

- [ ] **Step 5: Write failing tests for desired-state.ts**

```ts
import { describe, it, expect } from "vitest";
import { buildDesiredState } from "../desired-state";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";

const validResource = {
  name: "web",
  sourceType: "image" as const,
  image_spec: { image: "nginx:1" },
  execution_config: {
    environment_variables: [
      { from: "stack" as const, name: "MODE", value: "prod" },
      { from: "secret" as const, name: "TOKEN", secretId: "s-1", secretKey: "token" },
    ],
  },
};

describe("buildDesiredState", () => {
  it("includes valid resources keyed by name, with connections split out", () => {
    const d = buildDesiredState({ resources: [validResource], volumes: [] } as unknown as EditSessionDraft);
    expect([...d.resources.keys()]).toEqual(["web"]);
    const web = d.resources.get("web")!;
    // secret row does not ride as an env var
    expect(web.execution_config?.environment_variables).toEqual([{ name: "MODE", value: "prod" }]);
    expect(d.connections.size).toBe(1);
    const conn = [...d.connections.values()][0];
    expect(conn.from).toEqual({ type: "secret", id: "s-1" });
    expect(d.held.size).toBe(0);
  });

  it("holds an invalid named resource instead of dropping it", () => {
    const invalid = { name: "api", sourceType: "image" as const, image_spec: { image: "" } };
    const d = buildDesiredState({ resources: [invalid], volumes: [] } as unknown as EditSessionDraft);
    expect(d.resources.has("api")).toBe(false);
    expect(d.held.has("api")).toBe(true);
    expect(d.resourceIssues.get(0)?.length).toBeGreaterThan(0);
  });

  it("skips unnamed invalid resources without holding anything", () => {
    const d = buildDesiredState({ resources: [{ name: "", sourceType: "image" as const }], volumes: [] } as unknown as EditSessionDraft);
    expect(d.resources.size).toBe(0);
    expect(d.held.size).toBe(0);
    expect(d.resourceIssues.has(0)).toBe(true);
  });

  it("excludes connections whose rows are in-progress", () => {
    const r = { ...validResource, execution_config: { environment_variables: [{ from: "secret" as const, name: "X", secretId: "", secretKey: "" }] } };
    const d = buildDesiredState({ resources: [r], volumes: [] } as unknown as EditSessionDraft);
    expect(d.connections.size).toBe(0);
  });

  it("includes named volumes converted to API shape and skips unnamed ones", () => {
    const vol = { name: "web-data", sourceType: "None" as const, spec: { size: "1Gi", access_mode: "ReadWriteOnce" }, labels: [] };
    const d = buildDesiredState({ resources: [validResource], volumes: [vol, { name: "" }] } as unknown as EditSessionDraft);
    expect([...d.volumes.keys()]).toEqual(["web-data"]);
    expect((d.volumes.get("web-data") as Record<string, unknown>).sourceType).toBeUndefined();
  });
});
```

Run → FAIL (module not found).

- [ ] **Step 6: Implement desired-state.ts**

```ts
import type { z } from "zod";
import type { StackResourceUpdateRequest, VolumeUpdateRequest } from "@/api/stacks";
import type { StackConnection } from "@/api/connections";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";
import {
  FormStackResourceSchema,
  prepareFormResourceForApi,
  convertFormVolumeToApiVolume,
  type FormStackResourceData,
  type FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema";
import { buildDesiredConnections, type FormEnvRow } from "@/pages/stacks/lib/connection-mapping";
import { connectionIdentityKey } from "./server-state";

/**
 * What the server SHOULD hold, derived from the edit-session draft.
 * Invalid-but-named resources are "held": they produce no ops and exempt
 * their server-side counterparts (and connections) from deletion — a
 * half-typed resource must never read as deleted.
 */
export interface DesiredStackState {
  resources: Map<string, StackResourceUpdateRequest>;
  held: Set<string>;
  volumes: Map<string, VolumeUpdateRequest>;
  connections: Map<string, StackConnection>;
  /** zod issues per draft-resource index, for live drawer errors. */
  resourceIssues: Map<number, z.ZodIssue[]>;
}

export function buildDesiredState(draft: EditSessionDraft): DesiredStackState {
  const resources = new Map<string, StackResourceUpdateRequest>();
  const held = new Set<string>();
  const resourceIssues = new Map<number, z.ZodIssue[]>();
  const validForConnections: { name: string; rows: FormEnvRow[] }[] = [];

  draft.resources.forEach((raw, idx) => {
    const parsed = FormStackResourceSchema.safeParse(raw);
    if (parsed.success) {
      const data = parsed.data as FormStackResourceData;
      if (!data.name?.trim()) return; // unnamed: nothing to sync
      resources.set(data.name, prepareFormResourceForApi(data));
      validForConnections.push({
        name: data.name,
        rows: (data.execution_config?.environment_variables ?? []) as FormEnvRow[],
      });
      return;
    }
    resourceIssues.set(idx, parsed.error.issues);
    const name = (raw as Partial<FormStackResourceData>).name?.trim();
    if (name) held.add(name);
  });

  const volumes = new Map<string, VolumeUpdateRequest>();
  for (const raw of draft.volumes) {
    const name = (raw as Partial<FormVolumeExtendedData>).name?.trim();
    if (!name) continue;
    volumes.set(name, convertFormVolumeToApiVolume(raw as FormVolumeExtendedData));
  }

  const connections = new Map<string, StackConnection>();
  for (const conn of buildDesiredConnections(validForConnections)) {
    connections.set(connectionIdentityKey(conn), conn);
  }

  return { resources, held, volumes, connections, resourceIssues };
}
```

Run → PASS.

- [ ] **Step 7: Write failing tests for ops.ts**

Table-driven; the ordering invariants are the load-bearing cases:

```ts
import { describe, it, expect } from "vitest";
import { computeSyncOps, type SyncOp } from "../ops";
import type { ServerStackState } from "../server-state";
import type { DesiredStackState } from "../desired-state";

const emptyServer = (): ServerStackState => ({
  resourcesByName: new Map(), volumeIdByName: new Map(), volumesByName: new Map(), connections: new Map(),
});
const emptyDesired = (): DesiredStackState => ({
  resources: new Map(), held: new Set(), volumes: new Map(), connections: new Map(), resourceIssues: new Map(),
});
const kinds = (ops: SyncOp[]) => ops.map((o) => o.kind);

const webResource = { name: "web", image_spec: { image: "nginx:1" } } as never;
const secretConn = (to: string) => ({
  kind: "env", from: { type: "secret", id: "s-1" }, to: { type: "stack_resource", name: to },
  mappings: [{ target: { type: "env", name: "TOKEN" }, value: { output: "token" } }],
}) as never;

describe("computeSyncOps", () => {
  it("returns no ops when server and desired match", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    expect(computeSyncOps(server, desired)).toEqual([]);
  });

  it("creates a new resource", () => {
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    expect(kinds(computeSyncOps(emptyServer(), desired))).toEqual(["createResource"]);
  });

  it("updates a changed resource by name", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    const desired = emptyDesired();
    desired.resources.set("web", { name: "web", image_spec: { image: "nginx:2" } } as never);
    const ops = computeSyncOps(server, desired);
    expect(ops).toEqual([{ kind: "updateResource", name: "web", resource: desired.resources.get("web") }]);
  });

  it("treats structurally-empty differences as equal (no spurious updates)", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", { name: "web", image_spec: { image: "nginx:1" }, depends_on: [] } as never);
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    expect(computeSyncOps(server, desired)).toEqual([]);
  });

  it("deletes a resource's connections before the resource (no backend cascade)", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set("k1", { id: "c-1", conn: secretConn("web") });
    const ops = computeSyncOps(server, emptyDesired());
    const ks = kinds(ops);
    expect(ks.indexOf("deleteConnection")).toBeLessThan(ks.indexOf("deleteResource"));
  });

  it("orders a rename as create-new before delete-old", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    const desired = emptyDesired();
    desired.resources.set("web2", { name: "web2", image_spec: { image: "nginx:1" } } as never);
    const ks = kinds(computeSyncOps(server, desired));
    expect(ks.indexOf("createResource")).toBeLessThan(ks.indexOf("deleteResource"));
  });

  it("emits createVolume before resource ops", () => {
    const desired = emptyDesired();
    desired.volumes.set("web-data", { name: "web-data" } as never);
    desired.resources.set("web", webResource);
    const ks = kinds(computeSyncOps(emptyServer(), desired));
    expect(ks.indexOf("createVolume")).toBeLessThan(ks.indexOf("createResource"));
  });

  it("never deletes or updates volumes (no thin endpoints; revert handles removal)", () => {
    const server = emptyServer();
    server.volumesByName.set("old-data", { name: "old-data" } as never);
    server.volumeIdByName.set("old-data", "v-9");
    const ops = computeSyncOps(server, emptyDesired());
    expect(ops).toEqual([]);
  });

  it("updates a connection whose mappings changed, keyed by server id", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set("env|secret:s-1|stack_resource:web|", { id: "c-1", conn: secretConn("web") });
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    const changed = secretConn("web") as { mappings: unknown[] };
    changed.mappings = [{ target: { type: "env", name: "API_TOKEN" }, value: { output: "token" } }];
    desired.connections.set("env|secret:s-1|stack_resource:web|", changed as never);
    expect(computeSyncOps(server, desired)).toEqual([
      { kind: "updateConnection", id: "c-1", identityKey: "env|secret:s-1|stack_resource:web|", conn: changed },
    ]);
  });

  it("creates a connection with a new identity and deletes the replaced one", () => {
    const server = emptyServer();
    server.resourcesByName.set("web", webResource);
    server.connections.set("env|secret:s-1|stack_resource:web|", { id: "c-1", conn: secretConn("web") });
    const desired = emptyDesired();
    desired.resources.set("web", webResource);
    desired.connections.set("env|secret:s-2|stack_resource:web|", secretConn("web"));
    const ks = kinds(computeSyncOps(server, desired));
    expect(ks).toEqual(["deleteConnection", "createConnection"]);
  });

  it("exempts held resources and their connections from deletion", () => {
    const server = emptyServer();
    server.resourcesByName.set("api", { name: "api", image_spec: { image: "node:20" } } as never);
    server.connections.set("k", { id: "c-2", conn: secretConn("api") });
    const desired = emptyDesired();
    desired.held.add("api");
    expect(computeSyncOps(server, desired)).toEqual([]);
  });

  it("skips a server connection without an id for update/delete (heals on next refetch)", () => {
    const server = emptyServer();
    server.connections.set("k", { id: undefined, conn: secretConn("web") });
    expect(computeSyncOps(server, emptyDesired())).toEqual([]);
  });
});
```

Run → FAIL.

- [ ] **Step 8: Implement ops.ts**

```ts
import type { StackResourceUpdateRequest, VolumeUpdateRequest } from "@/api/stacks";
import type { StackConnection } from "@/api/connections";
import { deepEqual } from "@/pages/stacks/lib/stack-diff";
import type { ServerStackState } from "./server-state";
import type { DesiredStackState } from "./desired-state";

/**
 * One thin-endpoint mutation. Op order is a correctness invariant:
 * volumes exist before resources mount them; a resource exists before its
 * connections; connections die before their resource (the backend does not
 * cascade). Volume update/delete are intentionally absent — no thin endpoint
 * and no canvas affordance; revert handles removal wholesale.
 */
export type SyncOp =
  | { kind: "createVolume"; volume: VolumeUpdateRequest }
  | { kind: "createResource"; resource: StackResourceUpdateRequest }
  | { kind: "updateResource"; name: string; resource: StackResourceUpdateRequest }
  | { kind: "deleteConnection"; id: string; identityKey: string }
  | { kind: "updateConnection"; id: string; identityKey: string; conn: StackConnection }
  | { kind: "createConnection"; identityKey: string; conn: StackConnection }
  | { kind: "deleteResource"; name: string };

export function computeSyncOps(server: ServerStackState, desired: DesiredStackState): SyncOp[] {
  const createVolumes: SyncOp[] = [];
  for (const [name, volume] of desired.volumes) {
    if (!server.volumesByName.has(name)) createVolumes.push({ kind: "createVolume", volume });
  }

  const createResources: SyncOp[] = [];
  const updateResources: SyncOp[] = [];
  for (const [name, resource] of desired.resources) {
    const existing = server.resourcesByName.get(name);
    if (!existing) createResources.push({ kind: "createResource", resource });
    else if (!deepEqual(existing, resource)) updateResources.push({ kind: "updateResource", name, resource });
  }

  const deleteResources: SyncOp[] = [];
  for (const name of server.resourcesByName.keys()) {
    if (!desired.resources.has(name) && !desired.held.has(name)) {
      deleteResources.push({ kind: "deleteResource", name });
    }
  }
  const deletedResourceNames = new Set(deleteResources.map((op) => (op as { name: string }).name));

  const deleteConnections: SyncOp[] = [];
  const updateConnections: SyncOp[] = [];
  const createConnections: SyncOp[] = [];
  const connTouchesHeld = (conn: StackConnection): boolean => {
    const toName = conn.to?.name;
    const fromName = conn.from?.type === "stack_resource" ? conn.from?.name : undefined;
    return (!!toName && desired.held.has(toName)) || (!!fromName && desired.held.has(fromName));
  };

  for (const [key, entry] of server.connections) {
    const want = desired.connections.get(key);
    if (want) {
      if (entry.id && !deepEqual(entry.conn.mappings ?? [], want.mappings ?? [])) {
        updateConnections.push({ kind: "updateConnection", id: entry.id, identityKey: key, conn: want });
      }
      continue;
    }
    // Connections tied to a resource being deleted MUST go; otherwise spare
    // held resources' connections. Id-less entries are skipped and heal after
    // the next refetch rebuilds the mirror.
    const toName = entry.conn.to?.name;
    const forcedByDelete = !!toName && deletedResourceNames.has(toName);
    if (!forcedByDelete && connTouchesHeld(entry.conn)) continue;
    if (entry.id) deleteConnections.push({ kind: "deleteConnection", id: entry.id, identityKey: key });
  }
  for (const [key, conn] of desired.connections) {
    if (!server.connections.has(key) && !connTouchesHeld(conn)) {
      createConnections.push({ kind: "createConnection", identityKey: key, conn });
    }
  }

  return [
    ...createVolumes,
    ...createResources,
    ...updateResources,
    ...deleteConnections,
    ...updateConnections,
    ...createConnections,
    ...deleteResources,
  ];
}
```

Note the subtlety encoded above: a connection *to* a held resource is exempt from create as well — the resource may not exist server-side yet, and the backend validates connections against existing resources.

Run → PASS. Fix any test/impl mismatches (tests are the contract).

- [ ] **Step 9: Gates + commit**

```bash
pnpm --prefix frontend test -- --run && pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint
git add frontend/src/pages/stacks/lib/draft-sync frontend/src/pages/stacks/schemas/form-schema.ts
git commit -m "feat(stacks): pure draft-sync core — server mirror, desired state, op translator"
```

---

### Task 2: Thin API client functions

**Files:**
- Create: `frontend/src/api/stack-resources.ts`
- Create: `frontend/src/api/volumes.ts`
- Modify: `frontend/src/api/connections.ts` (add fns; today types-only)
- Modify: `frontend/src/api/stacks.ts` (add `deleteStack`)

**Interfaces:**
- Consumes: `api` axios client (`@/api/client`), generated `components` types.
- Produces (exact names Task 3+ imports):
  - `createStackResource(orgId, teamName, stackId, body: StackResourceUpdateRequest): Promise<StackResource>`
  - `updateStackResource(orgId, teamName, stackId, resourceName: string, body: StackResourceUpdateRequest): Promise<StackResource>`
  - `deleteStackResource(orgId, teamName, stackId, resourceName: string): Promise<void>`
  - `createStackConnection(orgId, teamName, stackId, body: StackConnection): Promise<StackConnection>`
  - `updateStackConnection(orgId, teamName, stackId, connectionId: string, body: StackConnection): Promise<StackConnection>`
  - `deleteStackConnection(orgId, teamName, stackId, connectionId: string): Promise<void>`
  - `createStackVolume(orgId, teamName, stackId, body: VolumeUpdateRequest): Promise<Volume>`
  - `deleteVolume(orgId, teamName, volumeId: string): Promise<void>`
  - `deleteStack(orgId, teamName, stackId: string): Promise<void>`

- [ ] **Step 1: stack-resources.ts**

```ts
import api from "./client";
import type { StackResource, StackResourceUpdateRequest } from "./stacks";

function resourcesPath(orgId: string, teamName: string, stackId: string): string {
  return `/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/resources`;
}

export async function createStackResource(orgId: string, teamName: string, stackId: string, body: StackResourceUpdateRequest): Promise<StackResource> {
  const response = await api.post<StackResource>(resourcesPath(orgId, teamName, stackId), body);
  return response.data;
}

export async function updateStackResource(orgId: string, teamName: string, stackId: string, resourceName: string, body: StackResourceUpdateRequest): Promise<StackResource> {
  const response = await api.put<StackResource>(`${resourcesPath(orgId, teamName, stackId)}/${encodeURIComponent(resourceName)}`, body);
  return response.data;
}

export async function deleteStackResource(orgId: string, teamName: string, stackId: string, resourceName: string): Promise<void> {
  await api.delete(`${resourcesPath(orgId, teamName, stackId)}/${encodeURIComponent(resourceName)}`);
}
```

- [ ] **Step 2: connections.ts additions** (keep existing type exports; import `api` like `releases.ts` does)

```ts
function connectionsPath(orgId: string, teamName: string, stackId: string): string {
  return `/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/connections`;
}

export async function createStackConnection(orgId: string, teamName: string, stackId: string, body: StackConnection): Promise<StackConnection> {
  const response = await api.post<StackConnection>(connectionsPath(orgId, teamName, stackId), body);
  return response.data;
}

export async function updateStackConnection(orgId: string, teamName: string, stackId: string, connectionId: string, body: StackConnection): Promise<StackConnection> {
  const response = await api.put<StackConnection>(`${connectionsPath(orgId, teamName, stackId)}/${connectionId}`, body);
  return response.data;
}

export async function deleteStackConnection(orgId: string, teamName: string, stackId: string, connectionId: string): Promise<void> {
  await api.delete(`${connectionsPath(orgId, teamName, stackId)}/${connectionId}`);
}
```

- [ ] **Step 3: volumes.ts**

```ts
import api from "./client";
import type { Volume, VolumeUpdateRequest } from "./stacks";

/** Thin stack-scoped create: the backend creates AND associates in one tx. */
export async function createStackVolume(orgId: string, teamName: string, stackId: string, body: VolumeUpdateRequest): Promise<Volume> {
  const response = await api.post<Volume>(`/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/volumes`, body);
  return response.data;
}

/** Destroys the cluster volume synchronously — confirm-gated callers only (revert). */
export async function deleteVolume(orgId: string, teamName: string, volumeId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/teams/${teamName}/volumes/${volumeId}`);
}
```

- [ ] **Step 4: stacks.ts — deleteStack**

```ts
export async function deleteStack(orgId: string, teamName: string, stackId: string): Promise<void> {
  await api.delete(`/organizations/${orgId}/teams/${teamName}/stacks/${stackId}`);
}
```

- [ ] **Step 5: Gates + commit**

```bash
pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint && pnpm --prefix frontend test -- --run
git add frontend/src/api
git commit -m "feat(stacks): thin resource/connection/volume API client functions"
```

---

### Task 3: Session `rebase` + the `useDraftSync` engine

**Files:**
- Modify: `frontend/src/pages/stacks/hooks/use-stack-edit-session.ts` (add `rebase`)
- Create: `frontend/src/pages/stacks/hooks/use-draft-sync.ts`
- Test: `frontend/src/pages/stacks/hooks/tests/use-draft-sync.test.tsx` (jsdom + fake timers; match the repo's existing hook-test layout — if hook tests live elsewhere, follow that convention)

**Interfaces:**
- Consumes: Task 1 calculations, Task 2 clients, `getStackById` (`api/stacks.ts:51`), `cloneJson`/`diffStack` (`lib/stack-diff.ts`), `UseStackEditSession`.
- Produces:
  - `UseStackEditSession.rebase(baseline: EditSessionDraft): void`
  - `useDraftSync(args: UseDraftSyncArgs): UseDraftSync` where

```ts
export interface UseDraftSyncArgs {
  enabled: boolean;
  stack: Stack | undefined;
  session: UseStackEditSession;
  ids: { orgId: string; teamName: string; stackId: string } | null;
  onStackRefreshed: (stack: Stack) => void;
}
export interface UseDraftSync {
  status: SyncStatus;
  failureCount: number;
  flush: () => Promise<boolean>;
  /** External writers (revert) hand the fresh stack here so the mirror stays truthful. */
  notifyExternalUpdate: (stack: Stack) => void;
}
```

- [ ] **Step 1: Add `rebase` to the session hook (with test)**

Failing test first — extend the existing session hook test file if one exists (`grep -rl "useStackEditSession" frontend/src --include="*.test.*"`), else add `frontend/src/pages/stacks/hooks/tests/use-stack-edit-session.test.tsx`:

```tsx
// @vitest-environment jsdom
import { renderHook, act } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { useStackEditSession } from "../use-stack-edit-session";

describe("rebase", () => {
  it("advances the baseline without touching the draft", () => {
    const { result } = renderHook(() => useStackEditSession());
    act(() => result.current.start({ resources: [{ name: "web" }], volumes: [] }));
    act(() => result.current.updateResources((prev) => [{ ...prev[0], name: "web2" }]));
    expect(result.current.dirty.dirtyResourceIdx.size).toBe(1);

    const snapshot = { resources: result.current.draft.resources, volumes: result.current.draft.volumes };
    act(() => result.current.rebase(snapshot));
    expect(result.current.dirty.dirtyResourceIdx.size).toBe(0);
    expect(result.current.draft.resources[0].name).toBe("web2");
  });

  it("is a no-op when the session is inactive", () => {
    const { result } = renderHook(() => useStackEditSession());
    act(() => result.current.rebase({ resources: [{ name: "x" }], volumes: [] }));
    expect(result.current.isActive).toBe(false);
    expect(result.current.baseline.resources).toEqual([]);
  });
});
```

Implementation — add to the interface and hook:

```ts
/** Advance the baseline to a synced snapshot; the draft is untouched, so
 *  edits made after the snapshot remain dirty. */
rebase: (baseline: EditSessionDraft) => void;
```

```ts
const rebase = useCallback((baseline: EditSessionDraft) => {
  setState((prev) => {
    if (!prev.isActive) return prev;
    return {
      ...prev,
      baseline: { resources: cloneJson(baseline.resources), volumes: cloneJson(baseline.volumes) },
    };
  });
}, []);
```

Return it from the hook. Run the test file → PASS.

- [ ] **Step 2: Write the failing engine tests**

`use-draft-sync.test.tsx`. Mock the api modules; drive a real session hook alongside the engine:

```tsx
// @vitest-environment jsdom
import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { useStackEditSession } from "../use-stack-edit-session";
import { useDraftSync } from "../use-draft-sync";
import { DEBOUNCE_IDLE_MS, DEBOUNCE_MAX_WAIT_MS, SYNC_STATUS } from "@/pages/stacks/lib/draft-sync/constants";
import type { Stack } from "@/api/stacks";

vi.mock("@/api/stacks", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/stacks")>()),
  getStackById: vi.fn(),
}));
vi.mock("@/api/stack-resources", () => ({
  createStackResource: vi.fn(), updateStackResource: vi.fn(), deleteStackResource: vi.fn(),
}));
vi.mock("@/api/connections", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/connections")>()),
  createStackConnection: vi.fn(), updateStackConnection: vi.fn(), deleteStackConnection: vi.fn(),
}));
vi.mock("@/api/volumes", () => ({ createStackVolume: vi.fn(), deleteVolume: vi.fn() }));

import { getStackById } from "@/api/stacks";
import { updateStackResource, createStackResource } from "@/api/stack-resources";

const serverStack = (image: string): Stack => ({
  id: "st-1", name: "demo",
  spec: { stack_resources: [{ id: "r-1", name: "web", image_spec: { image } }], volumes: [], connections: [] },
} as unknown as Stack);

const webForm = (image: string) => ({
  name: "web", sourceType: "image" as const, image_spec: { image },
});

function setup(stack: Stack) {
  const onStackRefreshed = vi.fn();
  const hook = renderHook(() => {
    const session = useStackEditSession();
    const sync = useDraftSync({
      enabled: true, stack, session,
      ids: { orgId: "o", teamName: "t", stackId: "st-1" },
      onStackRefreshed,
    });
    return { session, sync };
  });
  act(() => hook.result.current.session.start({ resources: [webForm("nginx:1")], volumes: [] }));
  return { hook, onStackRefreshed };
}

beforeEach(() => {
  vi.useFakeTimers();
  vi.mocked(getStackById).mockResolvedValue(serverStack("nginx:2"));
  vi.mocked(updateStackResource).mockResolvedValue({} as never);
  vi.mocked(createStackResource).mockResolvedValue({} as never);
});
afterEach(() => { vi.useRealTimers(); vi.clearAllMocks(); });

describe("useDraftSync", () => {
  it("debounces an edit and syncs after the idle window", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_IDLE_MS - 100));
    expect(updateStackResource).not.toHaveBeenCalled();
    await act(() => vi.advanceTimersByTimeAsync(200));
    expect(updateStackResource).toHaveBeenCalledOnce();
    expect(updateStackResource).toHaveBeenCalledWith("o", "t", "st-1", "web", expect.objectContaining({ name: "web" }));
    await waitFor(() => expect(hook.result.current.sync.status).toBe(SYNC_STATUS.saved));
    // baseline advanced: no more dirt
    expect(hook.result.current.session.dirty.dirtyResourceIdx.size).toBe(0);
  });

  it("coalesces rapid edits into one cycle", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    await act(() => vi.advanceTimersByTimeAsync(800));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:3")]));
    await act(() => vi.advanceTimersByTimeAsync(800));
    expect(updateStackResource).not.toHaveBeenCalled();
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_IDLE_MS));
    expect(updateStackResource).toHaveBeenCalledOnce();
  });

  it("fires at max-wait under continuous edits", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    for (let i = 0; i < 8; i++) {
      act(() => hook.result.current.session.updateResources(() => [webForm(`nginx:${i + 2}`)]));
      await act(() => vi.advanceTimersByTimeAsync(700)); // always inside the idle window
      if (700 * (i + 1) > DEBOUNCE_MAX_WAIT_MS + 100) break;
    }
    expect(updateStackResource).toHaveBeenCalled();
  });

  it("skips API calls when the diff is structurally empty, but still rebases", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    act(() => hook.result.current.session.updateResources((prev) => [{ ...prev[0], depends_on: [] }]));
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_IDLE_MS + 100));
    expect(updateStackResource).not.toHaveBeenCalled();
    expect(createStackResource).not.toHaveBeenCalled();
  });

  it("on failure keeps the draft, reports error status, retries with backoff, recovers", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    vi.mocked(updateStackResource).mockRejectedValueOnce(new Error("500"));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_IDLE_MS + 100));
    await waitFor(() => expect(hook.result.current.sync.status).toBe(SYNC_STATUS.error));
    expect(hook.result.current.session.draft.resources[0].image_spec?.image).toBe("nginx:2");
    // first backoff = RETRY_BASE_MS
    await act(() => vi.advanceTimersByTimeAsync(1100));
    await waitFor(() => expect(hook.result.current.sync.status).toBe(SYNC_STATUS.saved));
    expect(hook.result.current.sync.failureCount).toBe(0);
  });

  it("flush drains pending work and resolves true", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    let ok: boolean | undefined;
    await act(async () => { ok = await hook.result.current.sync.flush(); });
    expect(ok).toBe(true);
    expect(updateStackResource).toHaveBeenCalledOnce();
  });

  it("flush resolves false when the cycle fails", async () => {
    const { hook } = setup(serverStack("nginx:1"));
    vi.mocked(updateStackResource).mockRejectedValue(new Error("500"));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    let ok: boolean | undefined;
    await act(async () => { ok = await hook.result.current.sync.flush(); });
    expect(ok).toBe(false);
  });

  it("does nothing when disabled", async () => {
    const onStackRefreshed = vi.fn();
    const hook = renderHook(() => {
      const session = useStackEditSession();
      const sync = useDraftSync({ enabled: false, stack: serverStack("nginx:1"), session, ids: null, onStackRefreshed });
      return { session, sync };
    });
    act(() => hook.result.current.session.start({ resources: [webForm("nginx:1")], volumes: [] }));
    act(() => hook.result.current.session.updateResources(() => [webForm("nginx:2")]));
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_MAX_WAIT_MS * 2));
    expect(updateStackResource).not.toHaveBeenCalled();
  });
});
```

Run → FAIL (module not found).

- [ ] **Step 3: Implement use-draft-sync.ts**

```ts
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Stack } from "@/api/stacks";
import { getStackById } from "@/api/stacks";
import { createStackResource, updateStackResource, deleteStackResource } from "@/api/stack-resources";
import { createStackConnection, updateStackConnection, deleteStackConnection } from "@/api/connections";
import { createStackVolume } from "@/api/volumes";
import { cloneJson } from "@/pages/stacks/lib/stack-diff";
import {
  DEBOUNCE_IDLE_MS, DEBOUNCE_MAX_WAIT_MS, RETRY_BASE_MS, RETRY_MAX_MS, SYNC_STATUS, type SyncStatus,
} from "@/pages/stacks/lib/draft-sync/constants";
import { serverStateFromStack, type ServerStackState } from "@/pages/stacks/lib/draft-sync/server-state";
import { buildDesiredState } from "@/pages/stacks/lib/draft-sync/desired-state";
import { computeSyncOps, type SyncOp } from "@/pages/stacks/lib/draft-sync/ops";
import type { EditSessionDraft, UseStackEditSession } from "./use-stack-edit-session";

export interface UseDraftSyncArgs {
  enabled: boolean;
  stack: Stack | undefined;
  session: UseStackEditSession;
  ids: { orgId: string; teamName: string; stackId: string } | null;
  onStackRefreshed: (stack: Stack) => void;
}

export interface UseDraftSync {
  status: SyncStatus;
  failureCount: number;
  flush: () => Promise<boolean>;
  notifyExternalUpdate: (stack: Stack) => void;
}

type Ids = NonNullable<UseDraftSyncArgs["ids"]>;

async function executeOp(op: SyncOp, ids: Ids): Promise<void> {
  switch (op.kind) {
    case "createVolume": await createStackVolume(ids.orgId, ids.teamName, ids.stackId, op.volume); return;
    case "createResource": await createStackResource(ids.orgId, ids.teamName, ids.stackId, op.resource); return;
    case "updateResource": await updateStackResource(ids.orgId, ids.teamName, ids.stackId, op.name, op.resource); return;
    case "deleteResource": await deleteStackResource(ids.orgId, ids.teamName, ids.stackId, op.name); return;
    case "createConnection": await createStackConnection(ids.orgId, ids.teamName, ids.stackId, op.conn); return;
    case "updateConnection": await updateStackConnection(ids.orgId, ids.teamName, ids.stackId, op.id, op.conn); return;
    case "deleteConnection": await deleteStackConnection(ids.orgId, ids.teamName, ids.stackId, op.id); return;
  }
}

/**
 * Debounced, single-flight autosave of the edit-session draft through the thin
 * stack endpoints. Owns the server mirror (the only place server ids live).
 * On success the session baseline advances to the synced snapshot, so edits
 * made mid-flight stay dirty and trigger the next cycle.
 */
export function useDraftSync({ enabled, stack, session, ids, onStackRefreshed }: UseDraftSyncArgs): UseDraftSync {
  const [status, setStatus] = useState<SyncStatus>(SYNC_STATUS.idle);
  const [failureCount, setFailureCount] = useState(0);

  const sessionRef = useRef(session);
  sessionRef.current = session;
  const idsRef = useRef(ids);
  idsRef.current = ids;
  const onRefreshedRef = useRef(onStackRefreshed);
  onRefreshedRef.current = onStackRefreshed;

  const mirrorRef = useRef<ServerStackState | null>(null);
  const runningRef = useRef<Promise<boolean> | null>(null);
  const queuedRef = useRef(false);
  const failuresRef = useRef(0);
  const idleTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const maxWaitTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const retryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Seed the mirror once from the fetched stack; afterwards the engine's own
  // refetches (and notifyExternalUpdate) keep it truthful.
  useEffect(() => {
    if (enabled && stack && !mirrorRef.current) {
      mirrorRef.current = serverStateFromStack(stack);
    }
  }, [enabled, stack]);

  const clearDebounceTimers = useCallback(() => {
    if (idleTimerRef.current) { clearTimeout(idleTimerRef.current); idleTimerRef.current = null; }
    if (maxWaitTimerRef.current) { clearTimeout(maxWaitTimerRef.current); maxWaitTimerRef.current = null; }
  }, []);

  const startCycle = useCallback((): Promise<boolean> => {
    if (runningRef.current) {
      queuedRef.current = true;
      return runningRef.current;
    }
    const run = (async (): Promise<boolean> => {
      const s = sessionRef.current;
      const currentIds = idsRef.current;
      const mirror = mirrorRef.current;
      if (!currentIds || !mirror || !s.isActive) return true;

      const snapshot: EditSessionDraft = {
        resources: cloneJson(s.draft.resources),
        volumes: cloneJson(s.draft.volumes),
      };
      const desired = buildDesiredState(snapshot);
      const ops = computeSyncOps(mirror, desired);
      if (ops.length === 0) {
        s.rebase(snapshot);
        setStatus((prev) => (prev === SYNC_STATUS.idle ? SYNC_STATUS.idle : SYNC_STATUS.saved));
        return true;
      }

      setStatus(SYNC_STATUS.saving);
      try {
        for (const op of ops) await executeOp(op, currentIds);
        const fresh = await getStackById(currentIds.orgId, currentIds.teamName, currentIds.stackId);
        mirrorRef.current = serverStateFromStack(fresh);
        onRefreshedRef.current(fresh);
        sessionRef.current.rebase(snapshot);
        failuresRef.current = 0;
        setFailureCount(0);
        setStatus(SYNC_STATUS.saved);
        return true;
      } catch {
        failuresRef.current += 1;
        setFailureCount(failuresRef.current);
        setStatus(SYNC_STATUS.error);
        // Heal the mirror from server truth; the draft stays authoritative locally.
        try {
          const fresh = await getStackById(currentIds.orgId, currentIds.teamName, currentIds.stackId);
          mirrorRef.current = serverStateFromStack(fresh);
          onRefreshedRef.current(fresh);
        } catch { /* keep the stale mirror; the next attempt refetches again */ }
        const backoff = Math.min(RETRY_BASE_MS * 2 ** (failuresRef.current - 1), RETRY_MAX_MS);
        if (retryTimerRef.current) clearTimeout(retryTimerRef.current);
        retryTimerRef.current = setTimeout(() => { void startCycle(); }, backoff);
        return false;
      }
    })().finally(() => {
      runningRef.current = null;
      if (queuedRef.current) {
        queuedRef.current = false;
        idleTimerRef.current = setTimeout(() => { void startCycle(); }, 0);
      }
    });
    runningRef.current = run;
    return run;
  }, []);

  // Debounce: every draft change (while active+enabled) restarts the idle
  // window; a max-wait timer guarantees persistence under continuous typing.
  const draft = session.isActive ? session.draft : null;
  useEffect(() => {
    if (!enabled || !draft || !idsRef.current) return;
    if (idleTimerRef.current) clearTimeout(idleTimerRef.current);
    idleTimerRef.current = setTimeout(() => {
      clearDebounceTimers();
      void startCycle();
    }, DEBOUNCE_IDLE_MS);
    if (!maxWaitTimerRef.current) {
      maxWaitTimerRef.current = setTimeout(() => {
        clearDebounceTimers();
        void startCycle();
      }, DEBOUNCE_MAX_WAIT_MS);
    }
    return () => { /* timers cleared by the next run or unmount */ };
  }, [enabled, draft, clearDebounceTimers, startCycle]);

  const flush = useCallback(async (): Promise<boolean> => {
    clearDebounceTimers();
    if (retryTimerRef.current) { clearTimeout(retryTimerRef.current); retryTimerRef.current = null; }
    if (runningRef.current) {
      const ok = await runningRef.current;
      if (!ok) return false;
    }
    return startCycle();
  }, [clearDebounceTimers, startCycle]);
  const flushRef = useRef(flush);
  flushRef.current = flush;

  const notifyExternalUpdate = useCallback((fresh: Stack) => {
    mirrorRef.current = serverStateFromStack(fresh);
  }, []);

  // Best-effort persistence when the tab hides or the page unmounts.
  useEffect(() => {
    if (!enabled) return;
    const onVisibility = () => {
      if (document.visibilityState === "hidden") void flushRef.current();
    };
    document.addEventListener("visibilitychange", onVisibility);
    return () => {
      document.removeEventListener("visibilitychange", onVisibility);
      void flushRef.current();
    };
  }, [enabled]);

  useEffect(() => () => {
    clearDebounceTimers();
    if (retryTimerRef.current) clearTimeout(retryTimerRef.current);
  }, [clearDebounceTimers]);

  return useMemo(() => ({ status, failureCount, flush, notifyExternalUpdate }), [status, failureCount, flush, notifyExternalUpdate]);
}
```

Implementation notes for the engineer:
- The debounce effect keys on `session.draft` object identity — the session hook replaces `draft` on every mutation, and `rebase` does NOT touch `draft`, so a successful cycle never reschedules itself.
- The no-op cycle branch (`ops.length === 0`) MUST still `rebase` — structural-noise edits (clearing a field to empty) would otherwise stay dirty forever and re-trigger the debounce.
- A flush after a failed cycle returns `false` (deploy uses this).
- The max-wait timer only arms when absent, so it measures from the FIRST unsynced edit.

- [ ] **Step 4: Run engine tests until green**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/hooks/tests/use-draft-sync.test.tsx`
Expected: PASS (8 tests). Timing tests are the likely flake point — always advance with `advanceTimersByTimeAsync` inside `act`.

- [ ] **Step 5: Full gates + commit**

```bash
pnpm --prefix frontend test -- --run && pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint
git add frontend/src/pages/stacks/hooks
git commit -m "feat(stacks): draft autosave engine — debounced single-flight thin-endpoint sync"
```

---

### Task 4: Wire autosave into the detail page + shell rework

**Files:**
- Create: `frontend/src/pages/stacks/components/canvas/AutosaveStatus.tsx`
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx`
- Modify: `frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx`
- Test: update `frontend/src/pages/stacks/components/canvas/tests/canvas-editor-shell.test.tsx` (or wherever the shell tests live — `grep -rl CanvasEditorShell frontend/src --include='*.test.*'`)
- Test: `frontend/src/pages/stacks/components/canvas/tests/autosave-status.test.tsx`

**Interfaces:**
- Consumes: `useDraftSync` (Task 3), `buildDesiredState().resourceIssues` (Task 1), `SYNC_STATUS`/`SyncStatus`.
- Produces (Task 5-7 rely on): shell props `syncStatus: SyncStatus`, `onCreate?: () => void`, `isCreating?: boolean`, `onDiscardDraft?: () => void`, `canDiscardDraft: boolean`, `canDeleteStack: boolean`; page-level `draftSync` handle.

- [ ] **Step 1: AutosaveStatus component (test first)**

Test:

```tsx
// @vitest-environment jsdom
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import "@testing-library/jest-dom/vitest";
import { AutosaveStatus } from "../AutosaveStatus";
import { SYNC_STATUS } from "@/pages/stacks/lib/draft-sync/constants";

describe("AutosaveStatus", () => {
  it("renders nothing when idle", () => {
    const { container } = render(<AutosaveStatus status={SYNC_STATUS.idle} />);
    expect(container).toBeEmptyDOMElement();
  });
  it("shows saving", () => {
    render(<AutosaveStatus status={SYNC_STATUS.saving} />);
    expect(screen.getByText("Saving…")).toBeInTheDocument();
  });
  it("shows saved", () => {
    render(<AutosaveStatus status={SYNC_STATUS.saved} />);
    expect(screen.getByText("All changes saved")).toBeInTheDocument();
  });
  it("shows error", () => {
    render(<AutosaveStatus status={SYNC_STATUS.error} />);
    expect(screen.getByText("Save failed — retrying")).toBeInTheDocument();
  });
});
```

Component:

```tsx
import { Check, Loader2, TriangleAlert } from "lucide-react";
import { SYNC_STATUS, type SyncStatus } from "@/pages/stacks/lib/draft-sync/constants";

/** Compact autosave state readout for the canvas header. Purely presentational. */
export function AutosaveStatus({ status }: { status: SyncStatus }) {
  if (status === SYNC_STATUS.idle) return null;
  if (status === SYNC_STATUS.saving) {
    return (
      <span className="flex flex-none items-center gap-1.5 font-mono text-[11.5px] text-muted-foreground">
        <Loader2 className="size-3 animate-spin" aria-hidden /> Saving…
      </span>
    );
  }
  if (status === SYNC_STATUS.error) {
    return (
      <span className="flex flex-none items-center gap-1.5 font-mono text-[11.5px] text-danger">
        <TriangleAlert className="size-3" aria-hidden /> Save failed — retrying
      </span>
    );
  }
  return (
    <span className="flex flex-none items-center gap-1.5 font-mono text-[11.5px] text-muted-foreground">
      <Check className="size-3" aria-hidden /> All changes saved
    </span>
  );
}
```

Run both → PASS.

- [ ] **Step 2: Shell rework**

In `CanvasEditorShell.tsx`:
1. Props: add `syncStatus: SyncStatus`, `onCreate?: () => void`, `isCreating?: boolean`, `onDiscardDraft?: () => void`, `canDiscardDraft: boolean`, `canDeleteStack: boolean`. Remove `onSave` and `isSaving` (draft create uses `onCreate`/`isCreating`).
2. Replace the primary-button block (lines 121-145):

```tsx
const hasUnsaved = isActive && dirtyTotal > 0;
// Draft mode keeps ONE explicit action (nothing exists server-side until it
// runs); existing stacks autosave, so the primary is always Deploy.
const primaryButton = isDraft ? (
  <Button type="button" variant="default" size="sm" onClick={onCreate} disabled={isCreating}>
    {isCreating ? <Loader2 className="size-3.5 animate-spin" /> : <Save className="size-3.5" />}
    {isCreating ? "Creating" : "Create stack"}
  </Button>
) : (
  <Button
    type="button"
    variant="default"
    size="sm"
    onClick={onDeploy}
    disabled={deployBusy || !canWrite || !(isStaged || hasUnsaved)}
  >
    {deployBusy ? <Loader2 className="size-3.5 animate-spin" /> : <Rocket className="size-3.5" />}
    {deployBusy ? "Deploying" : "Deploy"}
  </Button>
);
```

3. Header row: replace the `isStaged && !hasUnsaved` mono text AND the `hasUnsaved` dirty label with:

```tsx
{!isDraft && isStaged && (
  <StatusPill variant="info" className="flex-none">DRAFT</StatusPill>
)}
<div className="flex-1" />
{!isDraft && <AutosaveStatus status={syncStatus} />}
{primaryButton}
```

(Check `StatusPill`'s variant names in `frontend/src/components/branded/status-pill.tsx:45-53` and pick the neutral/info one that isn't already used for `Pending`.)

4. Dropdown menu: remove the `Edit` item (autosave makes the session always-on — Step 3). Keep "Discard all changes" (session-scope, when `hasUnsaved`). Add above the Delete item:

```tsx
{canDiscardDraft && onDiscardDraft && (
  <DropdownMenuItem onClick={onDiscardDraft}>
    <Undo2 className="size-4" />
    Discard draft changes
  </DropdownMenuItem>
)}
```

Delete item: label to "Delete stack", gate `disabled={!canDeleteStack}` (wired in Task 7 — until then the page passes the existing "Not implemented" handler and `canDeleteStack={canWrite}`).

5. Update shell tests: primary-button matrix — draft→"Create stack"; existing+`isStaged`→Deploy enabled; existing+clean→Deploy disabled; existing+`hasUnsaved`→Deploy enabled; `syncStatus="saving"`→"Saving…" visible; DRAFT pill renders iff `isStaged && !isDraft`; Edit item gone.

- [ ] **Step 3: Detail page wiring**

In `detail/index.tsx`:
1. **Auto-start the session** for existing stacks: locate the current Edit-click handler (`onEdit` passed to the shell) that calls `session.start(baseline…)`. Convert to an effect (keep the same baseline construction — `mapStackResourceToFormData`/`mapVolumeToFormData` + `connectionAddonIds`):

```tsx
// Autosave model: the canvas is always editable for writers. The session
// starts as soon as the stack is loaded and restarts after discard/revert.
useEffect(() => {
  if (isDraft || !stackToShow || !canWriteStack || session.isActive) return;
  session.start(buildBaselineFromStack(stackToShow), { linkedAddonIds: connectionAddonIds });
}, [isDraft, stackToShow, canWriteStack, session, connectionAddonIds]);
```

(`buildBaselineFromStack` = whatever the Edit handler builds today; extract it to a local function if inline. Remove the `onEdit` prop/plumbing.)

2. **Mount the engine**:

```tsx
const draftSync = useDraftSync({
  enabled: !isDraft && canWriteStack,
  stack: stackToShow,
  session,
  ids: deployIds, // the existing {orgId, teamName, stackId} memo used by onDeploy
  onStackRefreshed: (fresh) => {
    setFetchedStack(fresh);
    setStacks((prev) => prev.map((s) => (s.id === fresh.id ? fresh : s))); // context write-through — stale currentStack must not win
  },
});
```

(Adapt to the actual context setter from `useStacks()` — see `stack-context.tsx:8`.)

3. **performSave → performCreate**: delete the existing-stack branch (the `updateStack` call and its state plumbing). Keep the draft branch exactly as-is; rename the function and the shell prop to `onCreate`/`isCreating`.
4. **Delete the save-time detach gate + dialog**: remove `handleSave`'s pendingDetach checks (lines ~433-451), the detach confirm dialog (~546-578), and `applyPendingDetach` (~247-253). (Session-field excision happens in Task 8.)
5. **Live drawer errors**: compute

```tsx
const desiredState = useMemo(() => buildDesiredState(session.draft), [session.draft]);
```

and feed `desiredState.resourceIssues` into the exact plumbing that previously carried save-time `validationErrors` to the resource drawer (adapt the issue→field-error mapping already present in this file; issue paths are now relative to a single resource, so drop the `["spec","stack_resources",idx]` prefix handling).
6. Pass new shell props: `syncStatus={draftSync.status}`, `canDiscardDraft={false}` (Task 6 flips), `canDeleteStack={canWrite}` + existing not-implemented `onDelete` (Task 7 wires).

- [ ] **Step 4: Full gates**

```bash
pnpm --prefix frontend test -- --run && pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint
```

Fix fallout: tests that clicked "Edit"/"Save" on the shell must adapt to the new matrix.

- [ ] **Step 5: Live smoke (Playwright MCP, http://localhost:5174, admin@stackdome.io / welcome@123)**

1. Open an existing stack → no Save/Edit affordance; canvas immediately editable.
2. Change an env var in the drawer → within ~2s "Saving…" then "All changes saved"; network shows `PUT …/resources/web` (thin), no stack PUT.
3. Reload → edit persisted.
4. Half-type a new resource (name only) → no failed requests fire; drawer shows field errors; existing resources untouched.

- [ ] **Step 6: Commit**

```bash
git add frontend/src
git commit -m "feat(stacks): autosave the canvas draft — no Save button, always-on edit session"
```

---

### Task 5: Deploy integration + flush semantics

**Files:**
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx` (onDeploy)
- Test: extend the shell test matrix if any enablement case is missing (page-level flush behavior is covered by engine tests; deploy plumbing is thin)

**Interfaces:**
- Consumes: `draftSync.flush()` (Task 3), `createRelease` (`api/releases.ts:25`), existing `runDeploy` helper (`index.tsx:275-286`).
- Produces: deploy always ships the just-synced draft.

- [ ] **Step 1: Flush-then-deploy**

Replace `onDeploy` (`index.tsx:288-291`):

```tsx
const onDeploy = useCallback(async () => {
  if (!deployIds) return;
  const flushed = await draftSync.flush();
  if (!flushed) {
    toast({
      title: "Deploy blocked",
      description: "Draft changes failed to save. Fix the save error and try again.",
      variant: "destructive",
    });
    return;
  }
  runDeploy(() => createRelease(deployIds.orgId, deployIds.teamName, deployIds.stackId), "Deploy started");
}, [deployIds, draftSync, runDeploy, toast]);
```

(Match the actual `runDeploy`/`toast` signatures in the file.)

- [ ] **Step 2: Gates + live smoke**

```bash
pnpm --prefix frontend test -- --run && pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint
```

Live: edit a field and click Deploy immediately (inside the debounce window) → network shows the thin PUT complete BEFORE `POST …/releases`; release goes `Pending → Released`; DRAFT pill clears once `last_converged` catches up (lifecycle `clean`).

- [ ] **Step 3: Commit**

```bash
git add frontend/src
git commit -m "feat(stacks): deploy flushes pending autosave before creating a release"
```

---

### Task 6: Revert — "Discard draft changes"

**Files:**
- Create: `frontend/src/pages/stacks/lib/draft-sync/snapshot-to-update.ts`
- Create: `frontend/src/pages/stacks/hooks/use-stack-revert.ts`
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx` (wire hook + confirm dialog)
- Modify: `frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx` (only if the Task 4 menu item needs adjustment)
- Test: `frontend/src/pages/stacks/lib/draft-sync/tests/snapshot-to-update.test.ts`

**Interfaces:**
- Consumes: `StackReleaseSnapshot` (`api/releases.ts`), `updateStack`/`StackUpdateRequest` (`api/stacks.ts`), `deleteVolume` (Task 2), `serverStateFromStack`/`cleanServerResource` (Task 1), the live-release snapshot already lazily loaded by `use-deploy-lifecycle.ts` via `detail.ensure/peek` (`use-release-detail.ts`), `draftSync.notifyExternalUpdate` (Task 3).
- Produces: `snapshotToUpdateRequest(snap, current: { name: string; labels?: Stack["labels"] }): StackUpdateRequest`; `volumesToDelete(stack: Stack, snap: StackReleaseSnapshot): { id: string; name: string }[]`; `useStackRevert(...): { reverting: boolean; revert: () => Promise<boolean> }`.

- [ ] **Step 1: snapshot-to-update tests (failing first)**

```ts
import { describe, it, expect } from "vitest";
import { snapshotToUpdateRequest, volumesToDelete } from "../snapshot-to-update";
import type { StackReleaseSnapshot } from "@/api/releases";
import type { Stack } from "@/api/stacks";

const snap = {
  resources: [{ id: "r-1", stack_id: "st-1", revision: 2, status: {}, name: "web", image_spec: { image: "nginx:1" }, volume_mounts: [{ source_volume_name: "web-data", target_path: "/data", stack_resource_id: "r-1", source_volume_type: "pvc" }] }],
  volumes: [{ id: "v-1", status: {}, name: "web-data", spec: { size: "1Gi" } }],
  connections: [{ id: "c-1", kind: "env", from: { type: "secret", id: "s-1" }, to: { type: "stack_resource", name: "web" }, mappings: [] }],
} as unknown as StackReleaseSnapshot;

describe("snapshotToUpdateRequest", () => {
  it("strips read-only fields from resources and volumes, keeps connection ids", () => {
    const req = snapshotToUpdateRequest(snap, { name: "demo", labels: [] });
    expect(req.name).toBe("demo");
    const res = req.spec.stack_resources[0] as Record<string, unknown>;
    expect(res.id).toBeUndefined();
    expect(res.revision).toBeUndefined();
    expect((req.spec.volumes?.[0] as Record<string, unknown>).id).toBeUndefined();
    // ids retained so the PUT's replace-all upserts instead of delete+create
    expect(req.spec.connections?.[0].id).toBe("c-1");
  });

  it("omits connections when the snapshot has none", () => {
    const req = snapshotToUpdateRequest({ ...snap, connections: [] } as StackReleaseSnapshot, { name: "demo" });
    expect(req.spec.connections).toBeUndefined();
  });
});

describe("volumesToDelete", () => {
  it("lists stack volumes absent from the snapshot (PUT never deletes volumes)", () => {
    const stack = {
      spec: { volumes: [{ id: "v-1", name: "web-data" }, { id: "v-2", name: "scratch" }] },
    } as unknown as Stack;
    expect(volumesToDelete(stack, snap)).toEqual([{ id: "v-2", name: "scratch" }]);
  });
});
```

Run → FAIL.

- [ ] **Step 2: Implement snapshot-to-update.ts**

```ts
import type { Stack, StackUpdateRequest, StackResource, Volume } from "@/api/stacks";
import type { StackReleaseSnapshot } from "@/api/releases";
import { cleanServerResource } from "./server-state";

function cleanVolume(v: Volume) {
  const { id, status, ...rest } = v as Volume & { status?: unknown };
  void id; void status;
  return rest;
}

/**
 * Turn a release snapshot back into a whole-stack PUT body. Replace-all
 * semantics are exactly right for a revert; connection ids are kept so the
 * backend upserts instead of churning rows.
 */
export function snapshotToUpdateRequest(
  snap: StackReleaseSnapshot,
  current: { name: string; labels?: Stack["labels"] },
): StackUpdateRequest {
  const connections = snap.connections ?? [];
  return {
    name: current.name,
    labels: current.labels,
    spec: {
      stack_resources: (snap.resources ?? []).map((r) => cleanServerResource(r as StackResource)),
      volumes: (snap.volumes ?? []).length > 0 ? (snap.volumes ?? []).map((v) => cleanVolume(v as Volume)) : undefined,
      ...(connections.length > 0 ? { connections } : {}),
    },
  } as StackUpdateRequest;
}

/** Volumes on the stack that the deployed snapshot doesn't know — draft
 *  artifacts to remove after the PUT (which never deletes volumes). */
export function volumesToDelete(stack: Stack, snap: StackReleaseSnapshot): { id: string; name: string }[] {
  const snapNames = new Set((snap.volumes ?? []).map((v) => v.name).filter(Boolean));
  return (stack.spec?.volumes ?? [])
    .filter((v) => v.id && v.name && !snapNames.has(v.name))
    .map((v) => ({ id: v.id!, name: v.name! }));
}
```

Run → PASS.

- [ ] **Step 3: use-stack-revert.ts**

```ts
import { useCallback, useState } from "react";
import type { Stack } from "@/api/stacks";
import { getStackById, updateStack } from "@/api/stacks";
import { deleteVolume } from "@/api/volumes";
import type { StackReleaseSnapshot } from "@/api/releases";
import { snapshotToUpdateRequest, volumesToDelete } from "@/pages/stacks/lib/draft-sync/snapshot-to-update";

export interface UseStackRevertArgs {
  ids: { orgId: string; teamName: string; stackId: string } | null;
  stack: Stack | undefined;
  liveSnapshot: StackReleaseSnapshot | undefined;
  /** session.discard — the page's session auto-start effect re-seeds from the refreshed stack. */
  onReverted: (fresh: Stack) => void;
}

/** Restore the authored stack to the last deployed snapshot. */
export function useStackRevert({ ids, stack, liveSnapshot, onReverted }: UseStackRevertArgs) {
  const [reverting, setReverting] = useState(false);

  const revert = useCallback(async (): Promise<boolean> => {
    if (!ids || !stack || !liveSnapshot) return false;
    setReverting(true);
    try {
      const req = snapshotToUpdateRequest(liveSnapshot, { name: stack.name, labels: stack.labels });
      await updateStack(ids.orgId, ids.teamName, ids.stackId, req);
      // Draft-only volumes are unmounted after the PUT; remove them (destroys
      // the cluster volume — the confirm dialog carries that warning).
      for (const v of volumesToDelete(stack, liveSnapshot)) {
        await deleteVolume(ids.orgId, ids.teamName, v.id);
      }
      const fresh = await getStackById(ids.orgId, ids.teamName, ids.stackId);
      onReverted(fresh);
      return true;
    } catch {
      return false;
    } finally {
      setReverting(false);
    }
  }, [ids, stack, liveSnapshot, onReverted]);

  return { reverting, revert };
}
```

- [ ] **Step 4: Wire into the detail page**

In `index.tsx`:
1. The live snapshot is already resolvable where `useDeployLifecycle` is wired: `detail.peek(liveReleaseId).data?.snapshot` (see `use-deploy-lifecycle.ts:119-129` — the page owns `detail`/`releases`; mirror that resolution).
2. Mount the hook:

```tsx
const stackRevert = useStackRevert({
  ids: deployIds,
  stack: stackToShow,
  liveSnapshot,
  onReverted: (fresh) => {
    setFetchedStack(fresh);
    setStacks((prev) => prev.map((s) => (s.id === fresh.id ? fresh : s)));
    draftSync.notifyExternalUpdate(fresh);
    session.discard(); // auto-start effect restarts the session on the reverted baseline
    toast({ title: "Draft discarded", description: "Stack restored to the last deployment." });
  },
});
```

3. Confirm dialog (reuse the existing discard `AlertDialog` idiom in the shell/page — destructive variant):

- Title: `Discard draft changes?`
- Body: `This restores the stack to its last deployment. Volumes added since then are deleted — their data is destroyed. This cannot be undone.`
- Confirm button: `Discard draft` (destructive), calls `void stackRevert.revert()`, disabled while `stackRevert.reverting`.

4. Shell props: `canDiscardDraft={isStaged && !!liveSnapshot && canWriteStack}`, `onDiscardDraft={() => setRevertConfirmOpen(true)}`.

- [ ] **Step 5: Gates + live smoke + commit**

```bash
pnpm --prefix frontend test -- --run && pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint
```

Live: deploy a stack → edit a field (autosaves; DRAFT pill appears) → "Discard draft changes" → confirm → canvas shows deployed values, pill clears, network shows one stack PUT.

```bash
git add frontend/src
git commit -m "feat(stacks): discard draft changes — revert authored state to the live release snapshot"
```

---

### Task 7: Delete stack

**Files:**
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx` (wire `onDelete`)
- Modify: `frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx` (only if menu gating needs it)

**Interfaces:**
- Consumes: `deleteStack` (Task 2), `useNavigate`, stacks context remover (check `stack-context.tsx` for the setter; filter the deleted id out).
- Produces: working delete for any stack (the primary UX driver is never-deployed drafts, where "Discard draft changes" has no snapshot to restore).

- [ ] **Step 1: Wire the delete flow**

Replace the "Not implemented" `onDelete` handler in `index.tsx`:

```tsx
const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
const [deleting, setDeleting] = useState(false);

const performDelete = useCallback(async () => {
  if (!deployIds) return;
  setDeleting(true);
  try {
    await deleteStack(deployIds.orgId, deployIds.teamName, deployIds.stackId);
    setStacks((prev) => prev.filter((s) => s.id !== deployIds.stackId));
    toast({ title: "Stack deleted", description: `"${stackToShow?.name}" was deleted.` });
    navigate("/stacks");
  } catch {
    toast({ title: "Delete failed", description: "The stack could not be deleted.", variant: "destructive" });
  } finally {
    setDeleting(false);
    setDeleteConfirmOpen(false);
  }
}, [deployIds, setStacks, stackToShow?.name, toast, navigate]);
```

Confirm dialog (destructive `AlertDialog`, same idiom as Task 6):
- Title: `Delete stack?`
- Body: `This permanently deletes "<name>", its resources, volumes and deployments. This cannot be undone.`
- Confirm: `Delete stack`, disabled while `deleting`.

Shell: `onDelete={() => setDeleteConfirmOpen(true)}`, `canDeleteStack={canWriteStack}`.

- [ ] **Step 2: Gates + live smoke + commit**

```bash
pnpm --prefix frontend test -- --run && pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint
```

Live: delete the two lingering test stacks (`canvas-wizard-check`, `smoke-canvas-test`) through the new UI — cleanup and verification in one move. Confirm `/stacks` list drops them.

```bash
git add frontend/src
git commit -m "feat(stacks): wire stack deletion from the canvas header"
```

---

### Task 8: Cleanup — excise `pendingDetach` and stale save plumbing

**Files:**
- Modify: `frontend/src/pages/stacks/hooks/use-stack-edit-session.ts` (remove `pendingDetach`/`setPendingDetach`)
- Modify: `frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx:53` (drop the read)
- Modify: `frontend/src/pages/stacks/lib/canvas/derive-graph.ts:87` (drop the param; update its tests)
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx` (drop remaining reads ~330, ~435, ~551 if Task 4 left any)
- Modify: any test fixtures referencing `pendingDetach`

**Interfaces:** none new — pure deletion. `setPendingDetach` has zero call sites (verified); every `pendingDetach` read receives a permanently-empty set today.

- [ ] **Step 1: Delete state + prop threading**

Remove from the session hook: the interface fields (lines 59, 77), the `useState` (83), the resets (104, 112), `setPendingDetach` callback (180-187), and both entries in the return (203, 213). Then chase compile errors: `StackCanvasTab`, `derive-graph` (remove the parameter and the dimmed/detached rendering branch it fed — it can never trigger), `detail/index.tsx`.

- [ ] **Step 2: Gates**

```bash
pnpm --prefix frontend test -- --run && pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint
```

Fix tests that constructed `pendingDetach` sets.

- [ ] **Step 3: Commit**

```bash
git add frontend/src
git commit -m "refactor(stacks): remove dead pendingDetach state and save-time detach plumbing"
```

---

## Final verification (after all tasks)

1. `mage test:unit && golangci-lint run ./...` — backend green.
2. `pnpm --prefix frontend test -- --run && pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint` — frontend green.
3. Playwright end-to-end sweep (http://localhost:5174):
   - Existing stack: edit env var → thin `PUT`, "All changes saved", reload persists, no Save button anywhere.
   - Add resource + inline volume + link addon → autosave uses `POST …/stacks/{id}/volumes`, `POST …/resources`, `POST …/connections`; no whole-stack PUT in the network log.
   - Rename a resource → `POST` new + `DELETE` old (+ connection churn), canvas stable.
   - Deploy mid-edit → flush precedes `POST …/releases`; DRAFT pill lifecycle staged → deploying → clean.
   - Discard draft changes → one stack PUT, canvas matches deployment.
   - Delete a never-deployed stack → lands on `/stacks`, list updated.
   - `/stacks/new` → compose → **Create stack** → `/stacks/:id` with autosave live.
4. Console: no errors beyond known environmental 403s on expired dev JWT.
