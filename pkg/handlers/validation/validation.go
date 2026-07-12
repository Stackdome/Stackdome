package validation

import (
	"reflect"

	"github.com/Stackdome/stackdome/pkg/errors"
)

type Validate func() *errors.ServiceError

func validateNotEmpty(i interface{}, fieldName string, field string) Validate {
	return func() *errors.ServiceError {
		value := reflect.ValueOf(i).Elem().FieldByName(fieldName)
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return errors.Validation("%s is required", field)
			}
			value = value.Elem()
		}
		if len(value.String()) == 0 {
			return errors.Validation("%s is required", field)
		}
		return nil
	}
}

func validateEmpty(i interface{}, fieldName string, field string) Validate {
	return func() *errors.ServiceError {
		value := reflect.ValueOf(i).Elem().FieldByName(fieldName)
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return nil
			}
			value = value.Elem()
		}
		if len(value.String()) != 0 {
			return errors.Validation("%s must be empty", field)
		}
		return nil
	}
}

func validateRules(rules []Validate) *errors.ServiceError {
	for _, rule := range rules {
		if err := rule(); err != nil {
			return err
		}
	}
	return nil
}

func ValidateAll(rules []Validate) Validate {
	return func() *errors.ServiceError {
		return validateRules(rules)
	}
}
