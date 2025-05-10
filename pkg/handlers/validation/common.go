package validation

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
)

func validateLabels(labelsPtr *[]openapi.Label) Validate {
	return func() *errors.ServiceError {
		if labelsPtr == nil {
			return nil
		}
		labels := *labelsPtr
		if len(labels) == 0 {
			return nil
		}
		for _, label := range labels {
			if label.Key == "" {
				return errors.Validation("label key cannot be empty")
			}
			if !ValidateLabelKey(label.Key) {
				return errors.Validation("label key '%s' is not a valid label key", label.Key)
			}
		}
		return nil
	}
}

func validateAnnotations(annotationsPtr *[]openapi.Annotation) Validate {
	return func() *errors.ServiceError {
		if annotationsPtr == nil {
			return nil
		}
		annotations := *annotationsPtr
		if len(annotations) == 0 {
			return nil
		}
		for _, annotation := range annotations {
			if annotation.Key == "" {
				return errors.Validation("annotation key cannot be empty")
			}
			if !ValidateAnnotationKey(annotation.Key) {
				return errors.Validation("annotation key '%s' is not a valid annotation key", annotation.Key)
			}
		}
		return nil
	}
}

// K8s string validations
func ValidateLabelKey(key string) bool {
	errors := validation.IsQualifiedName(key)
	return len(errors) == 0
}

func ValidateAnnotationKey(key string) bool {
	errors := validation.IsQualifiedName(key)
	return len(errors) == 0
}

func ValidateName(name string) bool {
	errors := validation.IsDNS1123Subdomain(name)
	return len(errors) == 0
}

func ValidateNamespace(namespace string) bool {
	errors := validation.IsDNS1123Label(namespace)
	return len(errors) == 0
}
