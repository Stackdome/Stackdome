# Failure Details API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Surface container crash details and build failure details from the cluster-agent CRs through the API, giving UI and CLI a single unified `last_failure` field on `StackResourceStatus` and a `last_build_failure_detail` field on `ImageBuildStatus`.

**Architecture:** The cluster-agent already populates `LastFailureDetails` on `StackResource` CRs and `LastBuildFailureDetail` on `ImageBuild` CRs. The api-server's existing controllers watch those CRs and write status to the DB — we extend those mappers to include the new fields. For build failures, the image-build-controller also propagates to the parent StackResource's `last_failure` so the UI can read everything from a single object. A shared helper in `pkg/controllers/failure.go` owns the k8s-reason → product enum translation.

**Tech Stack:** Go, GORM/PostgreSQL (JSONB status columns), controller-runtime, OpenAPI 3.0 + openapi-generator v6.0.1

---

## File Map

| File | Action | Purpose |
|---|---|---|
| `pkg/models/stack_resource.go` | Modify | Add `StackResourceFailure`, `ContainerFailureDetail` types; add `LastFailure` to `StackResourceStatus` |
| `pkg/models/image_build.go` | Modify | Add `BuildFailureDetail` type; add `LastBuildFailureDetail` to `ImageBuildStatus` |
| `pkg/controllers/failure.go` | Create | `MapFailureType`, `MapLastFailureDetails`, `MapBuildFailureDetail` — shared mapping helpers |
| `pkg/controllers/failure_test.go` | Create | Unit tests for all three helpers |
| `pkg/controllers/stackresource/stack_resource_controller.go` | Modify | Call `MapLastFailureDetails` in `mapClusterStatusToServerStatus` |
| `pkg/controllers/stackresource/stack_resource_controller_test.go` | Create | Unit test `mapClusterStatusToServerStatus` with failure details |
| `pkg/controllers/imagebuild/image_build_controller.go` | Modify | Call `MapBuildFailureDetail` in `mapClusterStatusToServerStatus`; add `propagateBuildFailureToStackResource` |
| `pkg/controllers/imagebuild/image_build_controller_test.go` | Create | Unit test `mapClusterStatusToServerStatus` with build failure detail |
| `config/openapi/stackdome_api.yaml` | Modify | Add `ContainerFailureDetail`, `BuildFailureDetail`, `StackResourceFailure` schemas; extend `StackResourceStatus` and `ImageBuildStatus` |
| `pkg/api/openapi/` | Regenerate | Run `make generate` |
| `pkg/presenters/stack_resource.go` | Modify | Add `presentStackResourceFailure`, `presentContainerFailureDetail`; call from `presentStackResourceStatus` |
| `pkg/presenters/stack.go` | Modify | Call `presentStackResourceFailure` in `presentResourceStatus` |
| `pkg/presenters/image_build.go` | Modify | Add `presentBuildFailureDetail`; call from `presentImageBuildStatus` |
| `pkg/presenters/stack_resource_test.go` | Create | Unit test presenter with failure data |
| `pkg/presenters/image_build_test.go` | Create | Unit test presenter with build failure data |

---

### Task 1: Add model types

**Files:**
- Modify: `pkg/models/stack_resource.go`
- Modify: `pkg/models/image_build.go`

- [ ] **Step 1: Add `BuildFailureDetail` to `pkg/models/image_build.go`**

Add this type and the new field on `ImageBuildStatus`. Insert after the closing brace of `ImageBuildStatus`:

```go
type BuildFailureDetail struct {
	FailureType  string `json:"failure_type"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
	RestartCount int32  `json:"restart_count"`
	ExitCode     *int32 `json:"exit_code,omitempty"`
}
```

Add `LastBuildFailureDetail *BuildFailureDetail` to `ImageBuildStatus`:

```go
type ImageBuildStatus struct {
	Conditions             []Condition         `json:"conditions"`
	State                  string              `json:"state"`
	BuildSourceHash        string              `json:"build_source_hash"`
	ImageURL               string              `json:"image_url"`
	BuildSourceRevision    string              `json:"build_source_revision"`
	LastObservedStatusHash string              `json:"last_observed_status_hash,omitempty"`
	LastBuildFailureDetail *BuildFailureDetail `json:"last_build_failure_detail,omitempty"`
}
```

- [ ] **Step 2: Add failure types to `pkg/models/stack_resource.go`**

Add after the `Ingress` type (around line 89):

```go
type StackResourceFailureType string

const (
	FailureTypeRuntimeCrash StackResourceFailureType = "runtime_crash"
	FailureTypeBuildFailure  StackResourceFailureType = "build_failure"
)

type ContainerFailureDetail struct {
	FailureType  string `json:"failure_type"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
	RestartCount int32  `json:"restart_count"`
	ExitCode     *int32 `json:"exit_code,omitempty"`
}

type StackResourceFailure struct {
	Type          StackResourceFailureType `json:"type"`
	Container     *ContainerFailureDetail  `json:"container,omitempty"`
	InitContainer *ContainerFailureDetail  `json:"init_container,omitempty"`
	Build         *BuildFailureDetail      `json:"build,omitempty"`
}
```

Add `LastFailure *StackResourceFailure` to `StackResourceStatus`:

```go
type StackResourceStatus struct {
	State                         StackResourceState    `json:"state"`
	Message                       string                `json:"message"`
	ObservedCrRevision            string                `json:"observed_cr_revision"`
	Conditions                    []Condition           `json:"conditions"`
	PublicIngresses               []Ingress             `json:"public_ingresses"`
	InternalServiceName           *string               `json:"internal_service_name,omitempty"`
	LastObservedStatusHash        string                `json:"last_observed_status_hash,omitempty"`
	LastRestartRequestProcessedAt *time.Time            `json:"last_restart_request_processed_at,omitempty"`
	LastFailure                   *StackResourceFailure `json:"last_failure,omitempty"`
}
```

- [ ] **Step 3: Verify compilation**

```bash
make binary
```

Expected: compiles without errors.

---

### Task 2: Add shared failure mapping helpers

**Files:**
- Create: `pkg/controllers/failure.go`
- Create: `pkg/controllers/failure_test.go`

- [ ] **Step 1: Write the failing tests**

Create `pkg/controllers/failure_test.go`:

```go
package controllers_test

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/controllers"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestMapFailureType(t *testing.T) {
	cases := []struct {
		reason   string
		expected string
	}{
		{"CrashLoopBackOff", "crash_loop"},
		{"OOMKilled", "out_of_memory"},
		{"ImagePullBackOff", "image_pull_failed"},
		{"ErrImagePull", "image_pull_failed"},
		{"CreateContainerError", "create_container_error"},
		{"Error", "exit_error"},
		{"", "exit_error"},
		{"SomethingUnknown", "exit_error"},
	}
	for _, tc := range cases {
		t.Run(tc.reason, func(t *testing.T) {
			got := controllers.MapFailureType(tc.reason)
			if got != tc.expected {
				t.Errorf("MapFailureType(%q) = %q, want %q", tc.reason, got, tc.expected)
			}
		})
	}
}

func TestMapLastFailureDetails_nil(t *testing.T) {
	got := controllers.MapLastFailureDetails("web", nil)
	if got != nil {
		t.Errorf("expected nil for empty details, got %v", got)
	}
}

func TestMapLastFailureDetails_mainContainer(t *testing.T) {
	details := []corev1alpha1.LastFailureDetail{
		{
			ContainerName:           "web",
			RestartCount:            3,
			LastTerminationReason:   "CrashLoopBackOff",
			LastTerminationMessage:  "back-off restarting",
			LastTerminationExitCode: ptr.To(int32(1)),
		},
	}
	got := controllers.MapLastFailureDetails("web", details)
	if got == nil {
		t.Fatal("expected non-nil failure")
	}
	if got.Type != models.FailureTypeRuntimeCrash {
		t.Errorf("Type = %q, want %q", got.Type, models.FailureTypeRuntimeCrash)
	}
	if got.Container == nil {
		t.Fatal("expected Container to be set")
	}
	if got.Container.FailureType != "crash_loop" {
		t.Errorf("Container.FailureType = %q, want crash_loop", got.Container.FailureType)
	}
	if got.Container.RestartCount != 3 {
		t.Errorf("Container.RestartCount = %d, want 3", got.Container.RestartCount)
	}
	if got.InitContainer != nil {
		t.Errorf("expected InitContainer to be nil")
	}
}

func TestMapLastFailureDetails_initContainer(t *testing.T) {
	details := []corev1alpha1.LastFailureDetail{
		{
			ContainerName:           "web-init",
			RestartCount:            1,
			LastTerminationReason:   "Error",
			LastTerminationExitCode: ptr.To(int32(2)),
		},
	}
	got := controllers.MapLastFailureDetails("web", details)
	if got == nil {
		t.Fatal("expected non-nil failure")
	}
	if got.InitContainer == nil {
		t.Fatal("expected InitContainer to be set")
	}
	if got.InitContainer.FailureType != "exit_error" {
		t.Errorf("InitContainer.FailureType = %q, want exit_error", got.InitContainer.FailureType)
	}
	if got.Container != nil {
		t.Errorf("expected Container to be nil")
	}
}

func TestMapBuildFailureDetail_nil(t *testing.T) {
	got := controllers.MapBuildFailureDetail(nil)
	if got != nil {
		t.Errorf("expected nil for nil input")
	}
}

func TestMapBuildFailureDetail(t *testing.T) {
	detail := &corev1alpha1.LastFailureDetail{
		ContainerName:           "kaniko",
		RestartCount:            0,
		LastTerminationReason:   "Error",
		LastTerminationMessage:  "COPY failed: file not found",
		LastTerminationExitCode: ptr.To(int32(1)),
	}
	got := controllers.MapBuildFailureDetail(detail)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if got.FailureType != "exit_error" {
		t.Errorf("FailureType = %q, want exit_error", got.FailureType)
	}
	if got.Message != "COPY failed: file not found" {
		t.Errorf("Message = %q", got.Message)
	}
	if got.ExitCode == nil || *got.ExitCode != 1 {
		t.Errorf("ExitCode = %v, want 1", got.ExitCode)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./pkg/controllers/... -run "TestMapFailureType|TestMapLastFailureDetails|TestMapBuildFailureDetail" -v
```

Expected: compile error — `controllers.MapFailureType` etc. do not exist yet.

- [ ] **Step 3: Create `pkg/controllers/failure.go`**

```go
package controllers

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func MapFailureType(reason string) string {
	switch reason {
	case "CrashLoopBackOff":
		return "crash_loop"
	case "OOMKilled":
		return "out_of_memory"
	case "ImagePullBackOff", "ErrImagePull":
		return "image_pull_failed"
	case "CreateContainerError":
		return "create_container_error"
	default:
		return "exit_error"
	}
}

func MapLastFailureDetails(resourceName string, details []corev1alpha1.LastFailureDetail) *models.StackResourceFailure {
	if len(details) == 0 {
		return nil
	}
	failure := &models.StackResourceFailure{Type: models.FailureTypeRuntimeCrash}
	initName := resourceName + "-init"
	for _, d := range details {
		fd := mapContainerFailureDetail(d)
		switch d.ContainerName {
		case resourceName:
			failure.Container = fd
		case initName:
			failure.InitContainer = fd
		}
	}
	return failure
}

func MapBuildFailureDetail(d *corev1alpha1.LastFailureDetail) *models.BuildFailureDetail {
	if d == nil {
		return nil
	}
	return &models.BuildFailureDetail{
		FailureType:  MapFailureType(d.LastTerminationReason),
		Reason:       d.LastTerminationReason,
		Message:      d.LastTerminationMessage,
		RestartCount: d.RestartCount,
		ExitCode:     d.LastTerminationExitCode,
	}
}

func mapContainerFailureDetail(d corev1alpha1.LastFailureDetail) *models.ContainerFailureDetail {
	return &models.ContainerFailureDetail{
		FailureType:  MapFailureType(d.LastTerminationReason),
		Reason:       d.LastTerminationReason,
		Message:      d.LastTerminationMessage,
		RestartCount: d.RestartCount,
		ExitCode:     d.LastTerminationExitCode,
	}
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./pkg/controllers/... -run "TestMapFailureType|TestMapLastFailureDetails|TestMapBuildFailureDetail" -v
```

Expected: all PASS.

- [ ] **Step 5: Run go fmt**

```bash
go fmt ./pkg/controllers/...
```

---

### Task 3: Update StackResource controller

**Files:**
- Create: `pkg/controllers/stackresource/stack_resource_controller_test.go`
- Modify: `pkg/controllers/stackresource/stack_resource_controller.go`

- [ ] **Step 1: Write the failing test**

Create `pkg/controllers/stackresource/stack_resource_controller_test.go`:

```go
package workspaceresource

import (
	"testing"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestMapClusterStatusToServerStatus_withRuntimeCrash(t *testing.T) {
	cr := &corev1alpha1.StackResource{}
	cr.Name = "web"
	cr.Status = corev1alpha1.StackResourceStatus{
		Phase:      corev1alpha1.StackResourcePhaseFailed,
		StatusHash: "abc123",
		LastFailureDetails: []corev1alpha1.LastFailureDetail{
			{
				ContainerName:           "web",
				RestartCount:            5,
				LastTerminationReason:   "CrashLoopBackOff",
				LastTerminationMessage:  "back-off restarting",
				LastTerminationExitCode: ptr.To(int32(1)),
			},
		},
	}

	got := mapClusterStatusToServerStatus(cr)

	if got.LastFailure == nil {
		t.Fatal("expected LastFailure to be set")
	}
	if got.LastFailure.Type != models.FailureTypeRuntimeCrash {
		t.Errorf("Type = %q, want runtime_crash", got.LastFailure.Type)
	}
	if got.LastFailure.Container == nil {
		t.Fatal("expected Container to be set")
	}
	if got.LastFailure.Container.FailureType != "crash_loop" {
		t.Errorf("Container.FailureType = %q, want crash_loop", got.LastFailure.Container.FailureType)
	}
	if got.LastFailure.Container.RestartCount != 5 {
		t.Errorf("Container.RestartCount = %d, want 5", got.LastFailure.Container.RestartCount)
	}
	if got.LastFailure.InitContainer != nil {
		t.Error("expected InitContainer to be nil")
	}
}

func TestMapClusterStatusToServerStatus_withInitContainerCrash(t *testing.T) {
	cr := &corev1alpha1.StackResource{}
	cr.Name = "api"
	cr.Status = corev1alpha1.StackResourceStatus{
		Phase:      corev1alpha1.StackResourcePhaseFailed,
		StatusHash: "def456",
		LastFailureDetails: []corev1alpha1.LastFailureDetail{
			{
				ContainerName:           "api-init",
				RestartCount:            0,
				LastTerminationReason:   "Error",
				LastTerminationExitCode: ptr.To(int32(2)),
			},
		},
	}

	got := mapClusterStatusToServerStatus(cr)

	if got.LastFailure == nil {
		t.Fatal("expected LastFailure to be set")
	}
	if got.LastFailure.InitContainer == nil {
		t.Fatal("expected InitContainer to be set")
	}
	if got.LastFailure.InitContainer.FailureType != "exit_error" {
		t.Errorf("InitContainer.FailureType = %q, want exit_error", got.LastFailure.InitContainer.FailureType)
	}
	if got.LastFailure.Container != nil {
		t.Error("expected Container to be nil")
	}
}

func TestMapClusterStatusToServerStatus_noFailure(t *testing.T) {
	cr := &corev1alpha1.StackResource{}
	cr.Name = "web"
	cr.Status = corev1alpha1.StackResourceStatus{
		Phase:      corev1alpha1.StackResourcePhaseReady,
		StatusHash: "ghi789",
	}

	got := mapClusterStatusToServerStatus(cr)

	if got.LastFailure != nil {
		t.Errorf("expected LastFailure to be nil for ready resource, got %v", got.LastFailure)
	}
}

func TestMapClusterStatusToServerStatus_lastRestartTime(t *testing.T) {
	now := metav1.NewTime(time.Now())
	cr := &corev1alpha1.StackResource{}
	cr.Name = "web"
	cr.Status = corev1alpha1.StackResourceStatus{
		LastRestartRequestProcessedAt: &now,
	}

	got := mapClusterStatusToServerStatus(cr)

	if got.LastRestartRequestProcessedAt == nil {
		t.Error("expected LastRestartRequestProcessedAt to be set")
	}
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
go test ./pkg/controllers/stackresource/... -v
```

Expected: FAIL — `mapClusterStatusToServerStatus` does not populate `LastFailure`.

- [ ] **Step 3: Update `mapClusterStatusToServerStatus` in `pkg/controllers/stackresource/stack_resource_controller.go`**

Replace the existing `mapClusterStatusToServerStatus` function:

```go
func mapClusterStatusToServerStatus(clusterInstance *corev1alpha1.StackResource) *models.StackResourceStatus {
	res := &models.StackResourceStatus{
		LastObservedStatusHash: clusterInstance.Status.StatusHash,
		State:                  mapStackResourceState(clusterInstance.Status.Phase),
		Conditions:             models.ConvertConditions(clusterInstance.Status.Conditions),
		PublicIngresses:        mapToPublicIngresses(clusterInstance.Status.ExternalAddress),
		ObservedCrRevision:     clusterInstance.Status.ObservedStackdomeServerObjectRevision,
		InternalServiceName:    clusterInstance.Status.InternalAddress,
		LastFailure:            controllers.MapLastFailureDetails(clusterInstance.Name, clusterInstance.Status.LastFailureDetails),
	}
	if clusterInstance.Status.LastRestartRequestProcessedAt != nil {
		res.LastRestartRequestProcessedAt = ptr.To(clusterInstance.Status.LastRestartRequestProcessedAt.UTC())
	}
	return res
}
```

Add the import for `controllers` package at the top of the file:

```go
"github.com/ashishmax31/stackdome-api-server/pkg/controllers"
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./pkg/controllers/stackresource/... -v
```

Expected: all PASS.

- [ ] **Step 5: Run go fmt**

```bash
go fmt ./pkg/controllers/stackresource/...
```

---

### Task 4: Update ImageBuild controller

**Files:**
- Create: `pkg/controllers/imagebuild/image_build_controller_test.go`
- Modify: `pkg/controllers/imagebuild/image_build_controller.go`

- [ ] **Step 1: Write the failing tests**

Create `pkg/controllers/imagebuild/image_build_controller_test.go`:

```go
package imagebuild

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"k8s.io/utils/ptr"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestMapClusterStatusToServerStatus_noFailure(t *testing.T) {
	status := buildsv1alpha1.ImageBuildStatus{
		Phase:      buildsv1alpha1.BuildPhaseSuccess,
		StatusHash: "abc123",
		ImageUrl:   "registry.io/img:tag",
	}

	got := mapClusterStatusToServerStatus(status)

	if got.LastBuildFailureDetail != nil {
		t.Errorf("expected LastBuildFailureDetail to be nil, got %v", got.LastBuildFailureDetail)
	}
	if got.ImageURL != "registry.io/img:tag" {
		t.Errorf("ImageURL = %q", got.ImageURL)
	}
}

func TestMapClusterStatusToServerStatus_withFailure(t *testing.T) {
	status := buildsv1alpha1.ImageBuildStatus{
		Phase:      buildsv1alpha1.BuildPhaseFailed,
		StatusHash: "def456",
		LastBuildFailureDetail: &corev1alpha1.LastFailureDetail{
			ContainerName:           "kaniko",
			RestartCount:            0,
			LastTerminationReason:   "Error",
			LastTerminationMessage:  "COPY failed: file not found",
			LastTerminationExitCode: ptr.To(int32(1)),
		},
	}

	got := mapClusterStatusToServerStatus(status)

	if got.LastBuildFailureDetail == nil {
		t.Fatal("expected LastBuildFailureDetail to be set")
	}
	if got.LastBuildFailureDetail.FailureType != "exit_error" {
		t.Errorf("FailureType = %q, want exit_error", got.LastBuildFailureDetail.FailureType)
	}
	if got.LastBuildFailureDetail.Message != "COPY failed: file not found" {
		t.Errorf("Message = %q", got.LastBuildFailureDetail.Message)
	}
	if got.LastBuildFailureDetail.ExitCode == nil || *got.LastBuildFailureDetail.ExitCode != 1 {
		t.Errorf("ExitCode = %v, want 1", got.LastBuildFailureDetail.ExitCode)
	}
}

func TestBuildStackResourceFailureFromBuild_failure(t *testing.T) {
	detail := &corev1alpha1.LastFailureDetail{
		ContainerName:           "kaniko",
		LastTerminationReason:   "OOMKilled",
		LastTerminationExitCode: ptr.To(int32(137)),
	}

	got := buildStackResourceFailureFromBuild(detail)

	if got == nil {
		t.Fatal("expected non-nil failure")
	}
	if got.Type != models.FailureTypeBuildFailure {
		t.Errorf("Type = %q, want build_failure", got.Type)
	}
	if got.Build == nil {
		t.Fatal("expected Build to be set")
	}
	if got.Build.FailureType != "out_of_memory" {
		t.Errorf("Build.FailureType = %q, want out_of_memory", got.Build.FailureType)
	}
}

func TestBuildStackResourceFailureFromBuild_nil(t *testing.T) {
	got := buildStackResourceFailureFromBuild(nil)
	if got != nil {
		t.Errorf("expected nil for nil input")
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./pkg/controllers/imagebuild/... -v
```

Expected: FAIL — `mapClusterStatusToServerStatus` doesn't populate `LastBuildFailureDetail`; `buildStackResourceFailureFromBuild` doesn't exist.

- [ ] **Step 3: Update `pkg/controllers/imagebuild/image_build_controller.go`**

Replace `mapClusterStatusToServerStatus`:

```go
func mapClusterStatusToServerStatus(clusterStatus buildsv1alpha1.ImageBuildStatus) *models.ImageBuildStatus {
	return &models.ImageBuildStatus{
		Conditions:             models.ConvertConditions(clusterStatus.Conditions),
		State:                  string(clusterStatus.Phase),
		ImageURL:               clusterStatus.ImageUrl,
		BuildSourceRevision:    clusterStatus.BuildSourceRevision,
		LastObservedStatusHash: clusterStatus.StatusHash,
		LastBuildFailureDetail: controllers.MapBuildFailureDetail(clusterStatus.LastBuildFailureDetail),
	}
}
```

Add the new helper function (unexported, used only within this package):

```go
func buildStackResourceFailureFromBuild(d *corev1alpha1.LastFailureDetail) *models.StackResourceFailure {
	if d == nil {
		return nil
	}
	return &models.StackResourceFailure{
		Type:  models.FailureTypeBuildFailure,
		Build: controllers.MapBuildFailureDetail(d),
	}
}
```

Add the propagation method on `ImageBuildReconciler`:

```go
func (r *ImageBuildReconciler) propagateBuildFailureToStackResource(
	ctx context.Context,
	dbStackResource *models.StackResource,
	clusterBuildStatus buildsv1alpha1.ImageBuildStatus,
) error {
	if dbStackResource.Status == nil {
		return nil
	}

	status := *dbStackResource.Status

	switch {
	case clusterBuildStatus.LastBuildFailureDetail != nil:
		status.LastFailure = buildStackResourceFailureFromBuild(clusterBuildStatus.LastBuildFailureDetail)
	case clusterBuildStatus.Phase == buildsv1alpha1.BuildPhaseSuccess:
		if status.LastFailure != nil && status.LastFailure.Type == models.FailureTypeBuildFailure {
			status.LastFailure = nil
		}
	default:
		return nil
	}

	if serr := r.DBResourceService.UpdateStatus(ctx, dbStackResource.ID, &status); serr != nil {
		return serr.AsError()
	}
	return nil
}
```

Wire the propagation into `Reconcile`. In the block that updates the build status (around line 117), add the propagation call after the status update:

```go
if dbResourceBuild.Status == nil || dbResourceBuild.Status.LastObservedStatusHash != imageBuild.Status.StatusHash {
    dbResourceBuild.Status = mapClusterStatusToServerStatus(imageBuild.Status)
    if serr := r.DBImageBuildService.InternalUpdateStatus(ctx, dbResourceBuild.ID, dbResourceBuild.Status); serr != nil {
        return ctrl.Result{}, fmt.Errorf("failed to update image build status: %v", serr)
    }
    if err := r.propagateBuildFailureToStackResource(ctx, dbStackResouce, imageBuild.Status); err != nil {
        return ctrl.Result{}, fmt.Errorf("failed to propagate build failure to stack resource: %v", err)
    }
    return ctrl.Result{}, nil
}
```

Add the missing imports to `image_build_controller.go`:

```go
"github.com/ashishmax31/stackdome-api-server/pkg/controllers"
corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./pkg/controllers/imagebuild/... -v
```

Expected: all PASS.

- [ ] **Step 5: Verify compilation**

```bash
make binary
```

- [ ] **Step 6: Run go fmt**

```bash
go fmt ./pkg/controllers/imagebuild/...
```

---

### Task 5: Update OpenAPI spec

**Files:**
- Modify: `config/openapi/stackdome_api.yaml`

- [ ] **Step 1: Add three new schemas to `components/schemas`**

Insert before the `Ingress:` schema (around line 3530):

```yaml
    ContainerFailureDetail:
      type: object
      properties:
        failure_type:
          type: string
          enum:
            - crash_loop
            - out_of_memory
            - image_pull_failed
            - create_container_error
            - exit_error
        reason:
          type: string
        message:
          type: string
        restart_count:
          type: integer
          format: int32
        exit_code:
          type: integer
          format: int32
    BuildFailureDetail:
      type: object
      properties:
        failure_type:
          type: string
          enum:
            - crash_loop
            - out_of_memory
            - image_pull_failed
            - create_container_error
            - exit_error
        reason:
          type: string
        message:
          type: string
        restart_count:
          type: integer
          format: int32
        exit_code:
          type: integer
          format: int32
    StackResourceFailure:
      type: object
      properties:
        type:
          type: string
          enum:
            - runtime_crash
            - build_failure
        container:
          $ref: '#/components/schemas/ContainerFailureDetail'
        init_container:
          $ref: '#/components/schemas/ContainerFailureDetail'
        build:
          $ref: '#/components/schemas/BuildFailureDetail'
```

- [ ] **Step 2: Add `last_failure` to `StackResourceStatus`**

In the `StackResourceStatus` schema (around line 3510), add after `conditions`:

```yaml
        last_failure:
          $ref: '#/components/schemas/StackResourceFailure'
```

- [ ] **Step 3: Add `last_build_failure_detail` to `ImageBuildStatus`**

In the `ImageBuildStatus` schema (around line 3370), add after `build_source_revision`:

```yaml
        last_build_failure_detail:
          $ref: '#/components/schemas/BuildFailureDetail'
```

---

### Task 6: Regenerate OpenAPI client

**Files:**
- Regenerate: `pkg/api/openapi/`

- [ ] **Step 1: Run the generator**

```bash
make generate
```

Expected: `pkg/api/openapi/` is updated with new types `ContainerFailureDetail`, `BuildFailureDetail`, `StackResourceFailure` and new fields on `StackResourceStatus` and `ImageBuildStatus`.

- [ ] **Step 2: Verify compilation after generation**

```bash
make binary
```

Expected: may have compile errors in the presenter files since they reference the old `openapi.StackResourceStatus` and `openapi.ImageBuildStatus` structs — these will be fixed in Task 7.

---

### Task 7: Update presenters

**Files:**
- Modify: `pkg/presenters/stack_resource.go`
- Modify: `pkg/presenters/stack.go`
- Modify: `pkg/presenters/image_build.go`
- Create: `pkg/presenters/stack_resource_test.go`
- Create: `pkg/presenters/image_build_test.go`

- [ ] **Step 1: Write failing tests for stack resource presenter**

Create `pkg/presenters/stack_resource_test.go`:

```go
package presenters_test

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"k8s.io/utils/ptr"
)

func TestPresentStackResource_lastFailure_runtimeCrash(t *testing.T) {
	resource := &models.StackResource{
		Name: "web",
		Status: &models.StackResourceStatus{
			State: models.StackResourcePhaseReady,
			LastFailure: &models.StackResourceFailure{
				Type: models.FailureTypeRuntimeCrash,
				Container: &models.ContainerFailureDetail{
					FailureType:  "crash_loop",
					Reason:       "CrashLoopBackOff",
					Message:      "back-off restarting",
					RestartCount: 5,
					ExitCode:     ptr.To(int32(1)),
				},
			},
		},
	}

	out := presenters.PresentStackResource(resource)

	if out.Status == nil {
		t.Fatal("expected Status to be set")
	}
	if out.Status.LastFailure == nil {
		t.Fatal("expected LastFailure to be set")
	}
	if out.Status.LastFailure.Type == nil || *out.Status.LastFailure.Type != "runtime_crash" {
		t.Errorf("LastFailure.Type = %v, want runtime_crash", out.Status.LastFailure.Type)
	}
	if out.Status.LastFailure.Container == nil {
		t.Fatal("expected Container to be set")
	}
	if out.Status.LastFailure.Container.FailureType == nil || *out.Status.LastFailure.Container.FailureType != "crash_loop" {
		t.Errorf("Container.FailureType = %v, want crash_loop", out.Status.LastFailure.Container.FailureType)
	}
	if out.Status.LastFailure.Container.RestartCount == nil || *out.Status.LastFailure.Container.RestartCount != 5 {
		t.Errorf("Container.RestartCount = %v, want 5", out.Status.LastFailure.Container.RestartCount)
	}
}

func TestPresentStackResource_lastFailure_buildFailure(t *testing.T) {
	resource := &models.StackResource{
		Name: "web",
		Status: &models.StackResourceStatus{
			State: models.StackResourcePhaseFailed,
			LastFailure: &models.StackResourceFailure{
				Type: models.FailureTypeBuildFailure,
				Build: &models.BuildFailureDetail{
					FailureType:  "exit_error",
					Reason:       "Error",
					Message:      "COPY failed",
					RestartCount: 0,
					ExitCode:     ptr.To(int32(2)),
				},
			},
		},
	}

	out := presenters.PresentStackResource(resource)

	if out.Status.LastFailure == nil {
		t.Fatal("expected LastFailure to be set")
	}
	if out.Status.LastFailure.Type == nil || *out.Status.LastFailure.Type != "build_failure" {
		t.Errorf("LastFailure.Type = %v, want build_failure", out.Status.LastFailure.Type)
	}
	if out.Status.LastFailure.Build == nil {
		t.Fatal("expected Build to be set")
	}
	if out.Status.LastFailure.Build.FailureType == nil || *out.Status.LastFailure.Build.FailureType != "exit_error" {
		t.Errorf("Build.FailureType = %v, want exit_error", out.Status.LastFailure.Build.FailureType)
	}
}

func TestPresentStackResource_noFailure(t *testing.T) {
	resource := &models.StackResource{
		Name: "web",
		Status: &models.StackResourceStatus{
			State: models.StackResourcePhaseReady,
		},
	}

	out := presenters.PresentStackResource(resource)

	if out.Status != nil && out.Status.LastFailure != nil {
		t.Errorf("expected LastFailure to be nil")
	}
}
```

- [ ] **Step 2: Write failing test for image build presenter**

Create `pkg/presenters/image_build_test.go`:

```go
package presenters_test

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"k8s.io/utils/ptr"
)

func TestPresentImageBuild_lastBuildFailureDetail(t *testing.T) {
	build := &models.ImageBuild{
		Status: &models.ImageBuildStatus{
			State: "Failed",
			LastBuildFailureDetail: &models.BuildFailureDetail{
				FailureType:  "exit_error",
				Reason:       "Error",
				Message:      "COPY failed: file not found",
				RestartCount: 0,
				ExitCode:     ptr.To(int32(1)),
			},
		},
	}

	out := presenters.PresentImageBuild(build)

	if out.Status == nil {
		t.Fatal("expected Status to be set")
	}
	if out.Status.LastBuildFailureDetail == nil {
		t.Fatal("expected LastBuildFailureDetail to be set")
	}
	if out.Status.LastBuildFailureDetail.FailureType == nil || *out.Status.LastBuildFailureDetail.FailureType != "exit_error" {
		t.Errorf("FailureType = %v, want exit_error", out.Status.LastBuildFailureDetail.FailureType)
	}
	if out.Status.LastBuildFailureDetail.Message == nil || *out.Status.LastBuildFailureDetail.Message != "COPY failed: file not found" {
		t.Errorf("Message = %v", out.Status.LastBuildFailureDetail.Message)
	}
}

func TestPresentImageBuild_noFailure(t *testing.T) {
	build := &models.ImageBuild{
		Status: &models.ImageBuildStatus{
			State:    "Success",
			ImageURL: "registry.io/img:tag",
		},
	}

	out := presenters.PresentImageBuild(build)

	if out.Status != nil && out.Status.LastBuildFailureDetail != nil {
		t.Errorf("expected LastBuildFailureDetail to be nil")
	}
}
```

- [ ] **Step 3: Run tests to confirm they fail**

```bash
go test ./pkg/presenters/... -v
```

Expected: FAIL — `LastFailure` and `LastBuildFailureDetail` fields not yet populated.

- [ ] **Step 4: Add `presentBuildFailureDetail` to `pkg/presenters/image_build.go`**

Add the helper and update `presentImageBuildStatus`:

```go
func presentImageBuildStatus(status *models.ImageBuildStatus) *openapi.ImageBuildStatus {
	if status == nil {
		return nil
	}
	return &openapi.ImageBuildStatus{
		State:                  &status.State,
		Conditions:             presentConditions(status.Conditions),
		ImageUrl:               &status.ImageURL,
		BuildSourceRevision:    &status.BuildSourceRevision,
		LastBuildFailureDetail: presentBuildFailureDetail(status.LastBuildFailureDetail),
	}
}

func presentBuildFailureDetail(d *models.BuildFailureDetail) *openapi.BuildFailureDetail {
	if d == nil {
		return nil
	}
	return &openapi.BuildFailureDetail{
		FailureType:  ptr.To(d.FailureType),
		Reason:       ptr.To(d.Reason),
		Message:      ptr.To(d.Message),
		RestartCount: ptr.To(d.RestartCount),
		ExitCode:     d.ExitCode,
	}
}
```

Add `"k8s.io/utils/ptr"` to the imports in `image_build.go`.

- [ ] **Step 5: Add failure helpers to `pkg/presenters/stack_resource.go`**

Update `presentStackResourceStatus` and add helpers:

```go
func presentStackResourceStatus(status *models.StackResourceStatus) *openapi.StackResourceStatus {
	if status == nil {
		return nil
	}
	return &openapi.StackResourceStatus{
		State:                         ptr.To(string(status.State)),
		ObservedRevision:              &status.ObservedCrRevision,
		Conditions:                    presentConditions(status.Conditions),
		PublicIngress:                 presentIngress(status.PublicIngresses),
		InternalServiceName:           status.InternalServiceName,
		LastRestartRequestProcessedAt: status.LastRestartRequestProcessedAt,
		LastFailure:                   presentStackResourceFailure(status.LastFailure),
	}
}

func presentStackResourceFailure(f *models.StackResourceFailure) *openapi.StackResourceFailure {
	if f == nil {
		return nil
	}
	return &openapi.StackResourceFailure{
		Type:          ptr.To(string(f.Type)),
		Container:     presentContainerFailureDetail(f.Container),
		InitContainer: presentContainerFailureDetail(f.InitContainer),
		Build:         presentBuildFailureDetail(f.Build),
	}
}

func presentContainerFailureDetail(d *models.ContainerFailureDetail) *openapi.ContainerFailureDetail {
	if d == nil {
		return nil
	}
	return &openapi.ContainerFailureDetail{
		FailureType:  ptr.To(d.FailureType),
		Reason:       ptr.To(d.Reason),
		Message:      ptr.To(d.Message),
		RestartCount: ptr.To(d.RestartCount),
		ExitCode:     d.ExitCode,
	}
}
```

Note: `presentBuildFailureDetail` is defined in `image_build.go` in the same `presenters` package — no duplication needed.

- [ ] **Step 6: Update `presentResourceStatus` in `pkg/presenters/stack.go`**

Replace `presentResourceStatus`:

```go
func presentResourceStatus(status *models.StackResourceStatus) *openapi.StackResourceStatus {
	if status == nil {
		return nil
	}
	return &openapi.StackResourceStatus{
		State:            ptr.To(string(status.State)),
		ObservedRevision: &status.ObservedCrRevision,
		Conditions:       presentConditions(status.Conditions),
		LastFailure:      presentStackResourceFailure(status.LastFailure),
	}
}
```

- [ ] **Step 7: Run tests to confirm they pass**

```bash
go test ./pkg/presenters/... -v
```

Expected: all PASS.

- [ ] **Step 8: Run go fmt**

```bash
go fmt ./pkg/presenters/...
```

---

### Task 8: Final verification

- [ ] **Step 1: Run all tests**

```bash
go test ./pkg/... -v 2>&1 | tail -30
```

Expected: all PASS, no failures.

- [ ] **Step 2: Full build**

```bash
make binary
```

Expected: binary builds without errors.
