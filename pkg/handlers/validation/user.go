package validation

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
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
		func() *errors.ServiceError {
			if in.HasInviteToken() {
				return nil
			}
			if in.Organisation == nil || in.Organisation.GetName() == "" {
				return errors.Validation("organisation name is required")
			}
			return nil
		},
	})
}

func ValidateUserLogin(in *openapi.LoginRequest) Validate {
	return ValidateAll([]Validate{
		validateNotEmpty(in, "Email", "email"),
		validateNotEmpty(in, "Password", "password"),
	})
}
