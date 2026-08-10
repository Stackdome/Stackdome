package validation

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

func ValidateVolume(in *openapi.Volume) Validate {
	// NOTE: validate only fields that exist on the generated openapi.Volume —
	// reflect.FieldByName on a missing field yields an invalid Value whose
	// String() is "<invalid Value>", making validateEmpty fail unconditionally.
	return ValidateAll([]Validate{
		validateEmpty(in, "Id", "id"),
		validateEmpty(in, "ProjectId", "project_id"),
		validateEmpty(in, "Status", "status"),
		validateLabels(&in.Labels),
		validateAnnotations(&in.Annotations),
		validateNotEmpty(in, "Name", "name"),
		func() *errors.ServiceError {
			if !ValidateName(in.Name) {
				return errors.Validation("name is not a valid name")
			}
			return nil
		},
		func() *errors.ServiceError {
			if in.Spec.AccessMode == "" {
				return errors.Validation("spec.access_mode is required")
			}

			if in.Spec.Size == "" {
				return errors.Validation("spec.size is required")
			}
			if _, err := k8sresource.ParseQuantity(in.Spec.Size); err != nil {
				return errors.Validation("spec.size is not a valid quantity")
			}
			if in.Spec.Source != nil {
				if err := validateGitRepoSource(&in.Spec.Source.GitRepoSource); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

func validateGitRepoSource(gitRepoSource *openapi.GitRepoSource) *errors.ServiceError {
	if gitRepoSource == nil {
		return errors.Validation("git repo source is required")
	}
	if gitRepoSource.RepoUrl == "" {
		return errors.Validation("git repo source repo url cannot be empty")
	}

	return validateGitRepoRevision(&gitRepoSource.Revision)
}

func validateGitRepoRevision(gitRepoRevision *openapi.GitRepoRevision) *errors.ServiceError {
	if gitRepoRevision == nil {
		return nil
	}

	if gitRepoRevision.Branch == nil && gitRepoRevision.Tag == nil {
		return errors.Validation("git repo revision requires a branch or tag (the commit SHA is resolved at release time)")
	}

	if gitRepoRevision.Branch != nil && gitRepoRevision.Tag != nil {
		return errors.Validation("git repo revision branch and tag cannot be set at the same time")
	}

	return nil
}
