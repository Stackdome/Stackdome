package computequota

import "github.com/Stackdome/stackdome/pkg/errors"

type StackLimitOperation string

const (
	StackLimitCreateStack    StackLimitOperation = "create_stack"
	StackLimitReplaceStack   StackLimitOperation = "replace_stack"
	StackLimitAddResource    StackLimitOperation = "add_resource"
	StackLimitUpdateResource StackLimitOperation = "update_resource"
)

// StackLimitChange describes how a proposed draft edit changes counted stack
// usage. It does not decide compute access or runtime reconciliation.
type StackLimitChange struct {
	Operation      StackLimitOperation
	OrganisationID string
	// StackID is required only for StackLimitReplaceStack, where the desired
	// resources replace every currently persisted resource in this stack.
	StackID string
	// DesiredResourceCount is the complete desired count for a whole-stack
	// create or replacement. Single-resource operations do not use it.
	DesiredResourceCount int64
}

// stackResourceExclusionFor identifies the one operation that replaces all of
// an existing stack's persisted resources. Other operations count every row.
func stackResourceExclusionFor(change StackLimitChange) (string, *errors.ServiceError) {
	switch change.Operation {
	case StackLimitReplaceStack:
		if change.StackID == "" {
			return "", errors.GeneralError("stack ID is required for a whole-stack update")
		}
		return change.StackID, nil
	case StackLimitCreateStack, StackLimitAddResource, StackLimitUpdateResource:
		return "", nil
	default:
		return "", errors.GeneralError("unsupported stack limit operation %q", change.Operation)
	}
}

// stackUsageAfterChange turns persisted usage into the totals that would exist
// after the requested mutation commits.
func stackUsageAfterChange(current ComputeUsage, change StackLimitChange) (ComputeUsage, *errors.ServiceError) {
	if (change.Operation == StackLimitCreateStack || change.Operation == StackLimitReplaceStack) && change.DesiredResourceCount < 0 {
		return ComputeUsage{}, errors.GeneralError("desired resource count must be non-negative")
	}

	proposed := current
	switch change.Operation {
	case StackLimitCreateStack:
		// 1 existing stack with 2 resources + a new 3-resource stack =
		// 2 stacks and 5 resources.
		proposed.StackCount = current.StackCount + 1
		proposed.StackResourceCount = current.StackResourceCount + change.DesiredResourceCount
	case StackLimitReplaceStack:
		// The query excludes the replaced stack's old resources. If the other
		// stacks have 2 resources and the replacement has 4, the total is 6;
		// replacing a stack leaves the stack count unchanged.
		proposed.StackCount = current.StackCount
		proposed.StackResourceCount = current.StackResourceCount + change.DesiredResourceCount
	case StackLimitAddResource:
		// 5 persisted resources + 1 added resource = 6 resources.
		proposed.StackCount = current.StackCount
		proposed.StackResourceCount = current.StackResourceCount + 1
	case StackLimitUpdateResource:
		// Updating 1 of 6 persisted resources in place leaves 6 resources.
		proposed.StackCount = current.StackCount
		proposed.StackResourceCount = current.StackResourceCount
	default:
		return ComputeUsage{}, errors.GeneralError("unsupported stack limit operation %q", change.Operation)
	}
	return proposed, nil
}
