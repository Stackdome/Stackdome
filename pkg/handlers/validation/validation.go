package validation

import (
	"reflect"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
)

type Validate func() *errors.ServiceError

func validateNotEmpty(i interface{}, fieldName string, field string) Validate {
	return func() *errors.ServiceError {
		value := reflect.ValueOf(i).Elem().FieldByName(fieldName)
		if value.Kind() == reflect.Ptr {
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

func validateNotRequiredIsNotEmpty(i interface{}, fieldName string, field string) Validate {
	return func() *errors.ServiceError {
		value := reflect.ValueOf(i).Elem().FieldByName(fieldName)
		if value.Kind() == reflect.Ptr {
			if value.IsNil() {
				return nil
			}
			value = value.Elem()
		}
		if len(value.String()) == 0 {
			return errors.Validation("%s can not be empty", field)
		}
		return nil
	}
}

func validateEmpty(i interface{}, fieldName string, field string) Validate {
	return func() *errors.ServiceError {
		value := reflect.ValueOf(i).Elem().FieldByName(fieldName)
		if value.Kind() == reflect.Ptr {
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

func validateNotEmptyStringField(field *string, name string) Validate {
	return func() *errors.ServiceError {
		if field == nil || len(*field) == 0 {
			return errors.Validation("%s is required", name)
		}
		return nil
	}
}

func validateNotNilField(field interface{}, name string) Validate {
	return func() *errors.ServiceError {
		if reflect.ValueOf(field).IsNil() {
			return errors.Validation("%s is required", name)
		}
		return nil
	}
}

func validateNotEmptyMap(field map[string]interface{}, name string) Validate {
	return func() *errors.ServiceError {
		if len(field) == 0 {
			return errors.Validation("%s is empty, it must contain at least one element", name)
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

func isEmpty[T any](field *T) bool {
	return field == nil || reflect.ValueOf(*field).IsZero()
}
