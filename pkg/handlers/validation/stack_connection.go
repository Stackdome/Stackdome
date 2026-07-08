package validation

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
)

// ValidateStackConnection validates the structurally-required fields of a stack
// connection on the thin create/update paths. It mirrors the fields the
// model-level validator relies on (kind, from.type, to.type) without re-checking
// the type-specific or graph-relative constraints, which need the surrounding
// stack to resolve. Id is never required here: it is generated on create and
// supplied via the URL path on update.
func ValidateStackConnection(in *openapi.StackConnection) *errors.ServiceError {
	if in.Kind == "" {
		return errors.BadRequest("kind: %s", "kind is required")
	}
	if in.From.Type == "" {
		return errors.BadRequest("from.type: %s", "from.type is required")
	}
	if in.To.Type == "" {
		return errors.BadRequest("to.type: %s", "to.type is required")
	}
	return nil
}
