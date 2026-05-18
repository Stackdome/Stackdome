# Failure Details API Design

## Background

The cluster-agent operator (PR #12) added two new CR status fields:

- `StackResource.Status.LastFailureDetails []LastFailureDetail` — captures crashing container info when a deployment is not ready. Populated per-revision (gated by `LastFailureRevision`, an internal field not surfaced in the API).
- `ImageBuild.Status.LastBuildFailureDetail *LastFailureDetail` — captures kaniko container failure info when a build job fails.

The `LastFailureDetail` type (shared, defined in `api/core/v1alpha1`):

```go
type LastFailureDetail struct {
    ContainerName           string
    RestartCount            int32
    LastTerminationReason   string
    LastTerminationMessage  string
    LastTerminationExitCode *int32
}
```

Container naming conventions in the cluster-agent:
- Main container: `resource.Name`
- Init container: `resource.Name + "-init"` (via `InitContainerName()`)

## Goals

1. Surface runtime crash details on `StackResourceStatus` — unified `last_failure` block covering both runtime crashes and build failures.
2. Surface build failure details on `ImageBuildStatus` — for direct build queries.
3. Classify raw k8s reasons into a product-level `failure_type` enum so UI/CLI don't need k8s knowledge.

## API Shape

### StackResourceStatus — `last_failure`

A unified block with a `type` discriminator. Only one of `container`/`init_container` will be set for `runtime_crash`; only `build` will be set for `build_failure`.

```json
{
  "last_failure": {
    "type": "runtime_crash",
    "container": {
      "failure_type": "crash_loop",
      "reason": "CrashLoopBackOff",
      "message": "back-off 5m0s restarting failed container",
      "restart_count": 5,
      "exit_code": 1
    },
    "init_container": null,
    "build": null
  }
}
```

For a build failure:

```json
{
  "last_failure": {
    "type": "build_failure",
    "container": null,
    "init_container": null,
    "build": {
      "failure_type": "exit_error",
      "reason": "Error",
      "message": "COPY failed: file not found in Dockerfile",
      "restart_count": 0,
      "exit_code": 2
    }
  }
}
```

`last_failure` is null when the resource is healthy.

### ImageBuildStatus — `last_build_failure_detail`

Surfaces the kaniko container failure directly on the build object. Null when build is successful or pending.

```json
{
  "last_build_failure_detail": {
    "failure_type": "exit_error",
    "reason": "Error",
    "message": "COPY failed: file not found in Dockerfile",
    "restart_count": 0,
    "exit_code": 2
  }
}
```

## Failure Type Enum

| k8s reason | `failure_type` |
|---|---|
| `CrashLoopBackOff` | `crash_loop` |
| `OOMKilled` | `out_of_memory` |
| `ImagePullBackOff` | `image_pull_failed` |
| `ErrImagePull` | `image_pull_failed` |
| `CreateContainerError` | `create_container_error` |
| `Error` / non-zero exit / anything else | `exit_error` |

## Implementation Layers

### 1. Models (`pkg/models/`)

**`stack_resource.go`** — add to `StackResourceStatus`:

```go
type StackResourceFailure struct {
    Type          StackResourceFailureType  `json:"type"`
    Container     *ContainerFailureDetail   `json:"container,omitempty"`
    InitContainer *ContainerFailureDetail   `json:"init_container,omitempty"`
    Build         *BuildFailureDetail       `json:"build,omitempty"`
}

type StackResourceFailureType string

const (
    FailureTypeRuntimeCrash StackResourceFailureType = "runtime_crash"
    FailureTypeBuildFailure StackResourceFailureType = "build_failure"
)

type ContainerFailureDetail struct {
    FailureType  string  `json:"failure_type"`
    Reason       string  `json:"reason,omitempty"`
    Message      string  `json:"message,omitempty"`
    RestartCount int32   `json:"restart_count"`
    ExitCode     *int32  `json:"exit_code,omitempty"`
}

// StackResourceStatus gets a new field:
// LastFailure *StackResourceFailure `json:"last_failure,omitempty"`
```

**`image_build.go`** — add to `ImageBuildStatus`:

```go
type BuildFailureDetail struct {
    FailureType  string `json:"failure_type"`
    Reason       string `json:"reason,omitempty"`
    Message      string `json:"message,omitempty"`
    RestartCount int32  `json:"restart_count"`
    ExitCode     *int32 `json:"exit_code,omitempty"`
}

// ImageBuildStatus gets a new field:
// LastBuildFailureDetail *BuildFailureDetail `json:"last_build_failure_detail,omitempty"`
```

### 2. Controllers

**`pkg/controllers/stackresource/stack_resource_controller.go`**

`mapClusterStatusToServerStatus` maps `LastFailureDetails` from the CR:
- Iterates `clusterInstance.Status.LastFailureDetails`
- `ContainerName == resource.Name` → `Container` field
- `ContainerName == resource.Name + "-init"` → `InitContainer` field
- Sets `Type = "runtime_crash"`
- Maps k8s reason → `failure_type` enum
- Clears `LastFailure` when `LastFailureDetails` is nil/empty

**`pkg/controllers/imagebuild/image_build_controller.go`**

`mapClusterStatusToServerStatus` maps `LastBuildFailureDetail` from the CR to `ImageBuildStatus.LastBuildFailureDetail`.

Additionally, when a build failure is detected (`LastBuildFailureDetail != nil`), the controller also writes `LastFailure` onto the parent StackResource DB record with `Type = "build_failure"`. This uses the existing `DBResourceService` already available on the reconciler.

When the build succeeds (`LastBuildFailureDetail == nil` and phase is `Success`), the controller clears `LastFailure` on the StackResource DB record.

The StackResource controller owns clearing `LastFailure` for `runtime_crash`; the ImageBuild controller owns clearing it for `build_failure`. Both write to the same `last_failure` JSONB column — last writer wins, which is safe because the two failure modes are mutually exclusive (a resource is either waiting for a build or running, not both).

### 3. OpenAPI Spec (`config/openapi/stackdome_api.yaml`)

New schemas:
- `ContainerFailureDetail` — used by `last_failure.container` and `last_failure.init_container`
- `BuildFailureDetail` — used by `last_failure.build` and `ImageBuildStatus.last_build_failure_detail`
- `StackResourceFailure` — the unified `last_failure` block

`StackResourceStatus` gets `last_failure` (`$ref: StackResourceFailure`, nullable).
`ImageBuildStatus` gets `last_build_failure_detail` (`$ref: BuildFailureDetail`, nullable).

### 4. Presenters (`pkg/presenters/`)

**`stack_resource.go`** and **`stack.go`** — `presentStackResourceStatus` includes `last_failure`, mapping from `models.StackResourceFailure`.

**`image_build.go`** — `presentImageBuildStatus` includes `last_build_failure_detail`.

### 5. Client Regeneration

After updating the OpenAPI spec, regenerate the frontend client:

```bash
make generate
```

## Failure Type Mapping Function

A single `mapFailureType(reason string, exitCode *int32) string` helper in the controllers package handles the k8s reason → enum translation, shared by both controllers.

## Clearing Semantics

| Event | Who clears | What is cleared |
|---|---|---|
| Deployment becomes available | stack-resource-controller | `StackResource.LastFailure` |
| Build succeeds | image-build-controller | `ImageBuild.LastBuildFailureDetail` + `StackResource.LastFailure` |
| New deployment revision (healthy) | stack-resource-controller | `StackResource.LastFailure` |

## Out of Scope

- Logs (not captured by `LastFailureDetail` in the final CRD)
- `LastFailureRevision` (internal cluster-agent field, never surfaced)
- Historical failure tracking (only the most recent failure is stored)
