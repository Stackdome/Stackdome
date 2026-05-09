package validation

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
)

func ValidateAPITokenCreate(in *openapi.APITokenCreateRequest) Validate {
	return ValidateAll([]Validate{
		func() *errors.ServiceError {
			if in.Name == "" {
				return errors.Validation("name is required")
			}
			return nil
		},
		func() *errors.ServiceError {
			if len(in.Scopes) == 0 {
				return errors.Validation("at least one scope is required")
			}
			return nil
		},
	})
}
