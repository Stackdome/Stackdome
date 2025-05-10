package validation

// func ValidateStackStorage(in *openapi.StackStorage) Validate {
// 	return ValidateAll([]Validate{
// 		validateEmpty(in, "Id", "id"),
// 		validateEmpty(in, "OrganisationId", "organisation_id"),
// 		validateEmpty(in, "Status", "status"),
// 		validateLabels(&in.Labels),
// 		validateAnnotations(&in.Annotations),
// 		validateNotEmpty(in, "Name", "name"),
// 		validateEmpty(in, "Namespace", "namespace"),
// 		validateStorageSpec(in),
// 		validateVolumes(in),
// 		func() *errors.ServiceError {
// 			if !ValidateName(in.Name) {
// 				return errors.Validation("name is not a valid name")
// 			}
// 			return nil
// 		},
// 	})
// }
// func ValidateStackStorageUpdate(in *openapi.StackStorage) Validate {
// 	return ValidateAll([]Validate{
// 		validateEmpty(in, "Id", "id"),
// 		validateEmpty(in, "OrganisationId", "organisation_id"),
// 		validateEmpty(in, "Status", "status"),
// 		validateLabels(&in.Labels),
// 		validateAnnotations(&in.Annotations),
// 		validateNotEmpty(in, "Name", "name"),
// 		validateEmpty(in, "Namespace", "namespace"),
// 		validateStorageSpec(in),
// 		validateVolumes(in),
// 		func() *errors.ServiceError {
// 			if !ValidateName(in.Name) {
// 				return errors.Validation("name is not a valid name")
// 			}
// 			return nil
// 		},
// 	})
// }

// func validateStorageSpec(in *openapi.StackStorage) Validate {
// 	return func() *errors.ServiceError {
// 		if in.Spec.WorkspaceName == "" {
// 			return errors.Validation("WorkspaceName for cannot be empty")
// 		}
// 		if err := validateVolumes(in)(); err != nil {
// 			return errors.Validation("validation error in volumes: %s", err.Error())
// 		}
// 		return nil
// 	}
// }

// func validateVolumes(in *openapi.StackStorage) Validate {
// 	return func() *errors.ServiceError {
// 		for _, volume := range in.Spec.Volumes {
// 			if volume.Name == "" {
// 				return errors.Validation("volume name cannot be empty")
// 			}
// 			if !ValidateName(volume.Name) {
// 				return errors.Validation("volume name is not a valid name")
// 			}
// 			if err := validateLabels(&volume.Labels)(); err != nil {
// 				return errors.Validation("validation error in volume labels: %s", err.Error())
// 			}
// 			if err := validateAnnotations(&volume.Annotations)(); err != nil {
// 				return errors.Validation("validation error in volume annotations: %s", err.Error())
// 			}
// 			if err := validateVolumeSpec(volume.Spec); err != nil {
// 				return errors.Validation("validation error in volume spec: %s", err.Error())
// 			}
// 		}
// 		return nil
// 	}
// }
