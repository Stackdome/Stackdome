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
		validateEmpty(in, "TeamId", "team_id"),
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
			// If source is not nil, validate it
			if in.Spec.Source != nil {
				if !in.Spec.Source.SourceType.IsValid() {
					return errors.Validation("spec.source is not a valid source type")
				}
				switch in.Spec.Source.SourceType {
				case openapi.REMOTE_DIR:
					if err := validateRemoteSource(in.Spec.Source.RemoteSource); err != nil {
						return err
					}
				case openapi.BUILD_ARTIFACT:
					if err := validateBuildSource(in.Spec.Source.BuildSource); err != nil {
						return err
					}
				case openapi.GIT_REPO:
					if err := validateGitRepoSource(in.Spec.Source.GitRepoSource); err != nil {
						return err
					}
				default:
					return errors.Validation("spec.source is not a valid source type")
				}
			}
			return nil
		},
	})
}

func validateBuildSource(buildSource []openapi.BuildArtifact) *errors.ServiceError {
	if len(buildSource) == 0 {
		return nil
	}
	for _, source := range buildSource {
		if len(source.ResourceRef) == 0 {
			return errors.Validation("build source resource ref cannot be empty")
		}
		if len(source.DestinationPath) == 0 {
			return errors.Validation("build source destination path cannot be empty")
		}
		if len(source.SourcePath) == 0 {
			return errors.Validation("build source source path cannot be empty")
		}
	}
	return nil
}

func validateRemoteSource(remoteSource *openapi.RemoteSource) *errors.ServiceError {
	if remoteSource == nil {
		return nil
	}
	if remoteSource.Path == "" {
		return errors.Validation("remote source path cannot be empty")
	}
	if remoteSource.CurrentDirectoryHash == "" {
		return errors.Validation("remote source current directory hash cannot be empty")
	}
	return nil
}

func validateGitRepoSource(gitRepoSource *openapi.GitRepoSource) *errors.ServiceError {
	if gitRepoSource == nil {
		return nil
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
