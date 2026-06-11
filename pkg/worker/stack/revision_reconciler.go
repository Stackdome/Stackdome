package stack

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

const revisionReconcilerName = "revision-reconciler"

type revisionReconciler struct {
	stackService   stackService
	stackCRBuilder builders.ClusterResourceBuilder
}

type RevisionReconcilerSpec struct {
	StackService   stackService
	StackCRBuilder builders.ClusterResourceBuilder
}

func NewRevisionReconciler(spec RevisionReconcilerSpec) *revisionReconciler {
	return &revisionReconciler{
		stackService:   spec.StackService,
		stackCRBuilder: spec.StackCRBuilder,
	}
}

func (r *revisionReconciler) Name() string {
	return revisionReconcilerName
}

func (r *revisionReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	if stack.DeletionTimestamp != nil {
		return resultNil, nil
	}

	crHash, err := r.stackCRBuilder.GetStackCRHash(stack)
	if err != nil {
		return resultNil, fmt.Errorf("failed to compute stack CR hash for stack '%s': %w", stack.ID, err)
	}
	if stack.CrRevision == crHash {
		return resultNil, nil
	}

	stack.CrRevision = crHash
	if serr := r.stackService.UpdateStackCrRevision(ctx, stack.ID, crHash); serr != nil {
		return resultNil, fmt.Errorf("failed to persist stack CR revision for stack '%s': %w", stack.ID, serr)
	}
	return resultNil, nil
}
