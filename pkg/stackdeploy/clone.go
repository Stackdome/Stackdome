package stackdeploy

import (
	"encoding/json"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/models"
)

// CloneStack deep-copies a stack via JSON so the resolver never mutates the
// caller's draft.
func CloneStack(stack *models.Stack) (*models.Stack, error) {
	if stack == nil {
		return nil, fmt.Errorf("stack is nil")
	}
	data, err := json.Marshal(stack)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stack for clone: %w", err)
	}
	var cloned models.Stack
	if err := json.Unmarshal(data, &cloned); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stack for clone: %w", err)
	}
	return &cloned, nil
}
