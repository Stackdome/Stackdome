package stack

import (
	"github.com/Stackdome/stackdome/pkg/models"
)

type interpolationValidation struct{}

func NewInterpolationValidation() *interpolationValidation {
	return &interpolationValidation{}
}

// ValidateStackInterpolations validates env var interpolation expressions.
// The cluster-agent interpolation package was removed during the CRD redesign;
// interpolation validation will be re-implemented in a future phase.
func (i *interpolationValidation) ValidateStackInterpolations(in *models.Stack) error {
	return nil
}
