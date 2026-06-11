package stack

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/clients"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/samber/lo"
)

const (
	validationReconcilerName    = "validation-reconciler"
	imageExistsValidationName   = "image-exists-validation"
	gitRepoExistsValidationName = "git-repo-exists-validation"
)

type validator interface {
	Run(ctx context.Context, stack *models.Stack) error
	Name() string
}

// Runs the validation for stack with talks to external apis like git repos, image registries, etc.
type validationReconciler struct {
	logger       logger.Logger
	stackService stackService
	validators   []validator
}

type ValidationReconcilerSpec struct {
	Logger        logger.Logger
	SecretService secretService
	StackService  stackService
}

func NewValidationReconciler(spec ValidationReconcilerSpec) *validationReconciler {
	return &validationReconciler{
		logger:       spec.Logger,
		stackService: spec.StackService,
		validators: []validator{
			&gitRepoExistsValidation{
				secretService: spec.SecretService,
			},
			&imageExistsValidation{
				secretService: spec.SecretService,
			},
		},
	}
}

func (v *validationReconciler) Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error) {
	stackCRHash := stack.CrRevision
	if v.shouldRunValidations(stack, stackCRHash) {
		currentValidationTypes := lo.Map(v.validators, func(v validator, _ int) string {
			return v.Name()
		})

		errors := []error{}
		for _, validator := range v.validators {
			v.logger.Infof("Running validation '%s' for stack '%s'", validator.Name(), stack.ID)
			if err := validator.Run(ctx, stack); err != nil {
				errors = append(errors, fmt.Errorf("validation '%s' failed for stack '%s': %s", validator.Name(), stack.ID, err.Error()))
			}
		}
		if len(errors) > 0 {
			v.logger.Errorf("Validations failed for stack '%s': %v", stack.ID, errors)
			stack.Status.State = models.StackError
			stack.Status.LastValidationRun = &models.ValidationRun{
				ValidationTypes: currentValidationTypes,
				StackRevision:   stackCRHash,
				Passed:          false,
				// Combine all error messages into a single string
				Message: lo.Reduce(errors, func(acc string, err error, _ int) string {
					return acc + err.Error() + "; "
				}, ""),
			}
			return resultStop, v.stackService.UpdateStatus(ctx, stack.ID, stack.Status).AsError()
		}
		stack.Status.LastValidationRun = &models.ValidationRun{
			ValidationTypes: currentValidationTypes,
			StackRevision:   stackCRHash,
			Passed:          true,
			Message:         "All validations passed",
		}
		return resultNil, v.stackService.UpdateStatus(ctx, stack.ID, stack.Status).AsError()
	}
	v.logger.Infof("Skipping validations for stack '%s' as they are not required", stack.ID)
	return resultNil, nil
}

func (v *validationReconciler) Name() string {
	return validationReconcilerName
}

func (v *validationReconciler) shouldRunValidations(stack *models.Stack, stackCRHash string) bool {
	if stack.Status == nil {
		return true
	}
	if stack.Status.LastValidationRun == nil || stack.Status.LastValidationRun.StackRevision != stackCRHash {
		return true
	}

	lastValidationRun := stack.Status.LastValidationRun
	currentValidationTypes := lo.Map(v.validators, func(v validator, _ int) string {
		return v.Name()
	})

	// if there are new validation types that were not part of the last validation run, we need to run validations again
	if len(lo.Intersect(lastValidationRun.ValidationTypes, currentValidationTypes)) != len(currentValidationTypes) {
		return true
	}

	return false
}

type gitRepoExistsValidation struct {
	secretService secretService
}

func (v *gitRepoExistsValidation) Name() string {
	return gitRepoExistsValidationName
}

func (v *gitRepoExistsValidation) Run(ctx context.Context, stack *models.Stack) error {
	for _, stackResource := range stack.StackResources {
		if stackResource.BuildConfig != nil && stackResource.BuildConfig.SourceContext.Git != nil {
			if stackResource.BuildConfig.SourceContext.Git.GitSecretRef != nil {
				if err := v.validateGitRepoExistsWithSecret(ctx, stackResource); err != nil {
					return err
				}
			} else {
				if err := v.validateGitRepoExists(ctx, stackResource); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (w *gitRepoExistsValidation) validateGitRepoExistsWithSecret(ctx context.Context, resource *models.StackResource) error {
	secretRef := resource.BuildConfig.SourceContext.Git.GitSecretRef
	secret, err := w.secretService.InternalGetByID(context.Background(), secretRef.SecretID)
	if err != nil {
		return fmt.Errorf("failed to get git secret for stack resource '%s': %s", resource.Name, err.Error())
	}

	gitClient, cerr := w.createGitClient(secret)
	if cerr != nil {
		return fmt.Errorf("failed to create git client for stack resource '%s': %s", resource.Name, cerr.Error())
	}

	return w.checkGitAccess(ctx, gitClient, resource)
}

func (w *gitRepoExistsValidation) validateGitRepoExists(ctx context.Context, resource *models.StackResource) error {
	gitClient, err := clients.NewGitClientAnonymous()
	if err != nil {
		return fmt.Errorf("failed to create git client for stack resource '%s': %s", resource.Name, err.Error())
	}

	return w.checkGitAccess(ctx, gitClient, resource)
}

func (w *gitRepoExistsValidation) createGitClient(secret *models.Secret) (clients.GitClient, error) {
	if token, ok := secret.Data[models.TokenSecretKey]; ok {
		return clients.NewGitClientWithToken(token)
	}

	return clients.NewGitClient(
		secret.Data[models.UsernameSecretKey],
		secret.Data[models.PasswordSecretKey],
	)
}

func (w *gitRepoExistsValidation) checkGitAccess(ctx context.Context, gitClient clients.GitClient, resource *models.StackResource) error {
	repoURL := resource.BuildConfig.SourceContext.Git.RepoURL

	canAccess, err := gitClient.CheckAccess(ctx, repoURL)
	if err != nil {
		return fmt.Errorf("failed to check git access for stack resource '%s': %s", resource.Name, err.Error())
	}
	if !canAccess {
		return fmt.Errorf("stack resource '%s' git repository '%s' is not accessible with the provided credentials", resource.Name, repoURL)
	}

	return nil
}

type imageExistsValidation struct {
	secretService secretService
}

func (v *imageExistsValidation) Name() string {
	return imageExistsValidationName
}

func (v *imageExistsValidation) Run(ctx context.Context, stack *models.Stack) error {
	for _, stackResource := range stack.StackResources {
		if stackResource.ImageConfig != nil {
			if stackResource.ImageConfig.PullSecretRef != nil {
				if err := v.validateImageExistsWithSecret(ctx, stackResource); err != nil {
					return err
				}
			} else {
				if err := v.validateImageExists(ctx, stackResource); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (v *imageExistsValidation) validateImageExistsWithSecret(ctx context.Context, resource *models.StackResource) error {
	if resource.ImageConfig.PullSecretRef == nil {
		return nil
	}
	secretRef := resource.ImageConfig.PullSecretRef

	if secretRef.SecretID == "" {
		return fmt.Errorf("stack resource '%s' has empty pull secret ID", resource.Name)
	}

	secret, err := v.secretService.InternalGetByID(ctx, secretRef.SecretID)
	if err != nil {
		return fmt.Errorf("failed to get pull secret for stack resource '%s': %s", resource.Name, err.Error())
	}

	client, cerr := clients.NewRegistryClientWithAuth(
		secret.Data[models.UsernameSecretKey],
		secret.Data[models.PasswordSecretKey],
	)
	if cerr != nil {
		return fmt.Errorf("failed to create image checker for stack resource '%s': %s", resource.Name, cerr.Error())
	}

	return v.checkImageExists(ctx, client, resource, "does not exist or is not pullable")
}

func (v *imageExistsValidation) validateImageExists(ctx context.Context, resource *models.StackResource) error {
	client, err := clients.NewRegistryClientAnonymous()
	if err != nil {
		return fmt.Errorf("failed to create anonymous registry client for stack resource '%s': %s", resource.Name, err.Error())
	}

	return v.checkImageExists(ctx, client, resource, "does not exist or is not pullable")
}

func (v *imageExistsValidation) checkImageExists(ctx context.Context, client clients.RegistryClient, resource *models.StackResource, errorSuffix string) error {
	exists, err := client.CheckImage(ctx, resource.ImageConfig.Image)
	if err != nil {
		return fmt.Errorf("failed to check image for stack resource '%s': %s", resource.Name, err.Error())
	}
	if !exists {
		return fmt.Errorf("stack resource '%s' image '%s' %s", resource.Name, resource.ImageConfig.Image, errorSuffix)
	}
	return nil
}
