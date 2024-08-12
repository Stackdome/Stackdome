package validation

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
)

func ValidateUserCreate(in *openapi.UserCreateRequest) Validate {
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
	})
}

func ValidateUserLogin(in *openapi.LoginRequest) Validate {
	return ValidateAll([]Validate{
		validateNotEmpty(in, "Email", "email"),
		validateNotEmpty(in, "Password", "password"),
	})
}
