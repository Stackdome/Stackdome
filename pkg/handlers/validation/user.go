package validation

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
)

func ValidateUserCreate(in *openapi.UserSignupRequest) Validate {
	return ValidateAll([]Validate{
		validateNotEmpty(in, "Email", "email"),
		validateNotEmpty(in, "Password", "password"),
		validateNotEmpty(in, "Name", "name"),
		func() *errors.ServiceError {
			if len(in.GetPassword()) < 8 {
				return errors.Validation("min password length should be 8")
			}
			return nil
		},
		validateOrganisation(in),
	})
}

func validateOrganisation(in *openapi.UserSignupRequest) Validate {
	return func() *errors.ServiceError {
		if in.HasOrganisationId() && in.HasOrganisation() {
			return errors.Validation("only one of organisation_id or organisation should be set")
		}
		if in.HasOrganisationId() && in.GetOrganisationId() == "" {
			return errors.Validation("organisation_id should not be empty")
		}
		if in.HasOrganisation() && in.Organisation.GetName() == "" {
			return errors.Validation("organisation name should not be empty")
		}
		return nil
	}
}

func ValidateUserLogin(in *openapi.LoginRequest) Validate {
	return ValidateAll([]Validate{
		validateNotEmpty(in, "Email", "email"),
		validateNotEmpty(in, "Password", "password"),
	})
}
