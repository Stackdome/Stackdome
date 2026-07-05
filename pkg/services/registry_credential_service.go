package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/clients"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
)

// AffectedStackRef identifies a stack that depends on an org-level registry
// credential — either implicitly (live-spec host match, no explicit secret ref)
// or explicitly (a pinned credential ID recorded in the reference graph, which
// also spans retained releases).
type AffectedStackRef struct {
	ID   string
	Name string
}

// RegistryCredentialDeleteResult reports which live stacks were resolving
// against a deleted credential. Deletion is not blocked; this is rotation
// visibility.
type RegistryCredentialDeleteResult struct {
	AffectedStacks []AffectedStackRef
}

//go:generate mockgen -source=registry_credential_service.go -destination=registry_credential_service_mock_test.go -package=services -exclude_interfaces RegistryCredentialService
type verifyRegistryClientProvider interface {
	ClientFor(username, password string) (clients.RegistryClient, error)
}

type defaultVerifyRegistryClientProvider struct{}

func (defaultVerifyRegistryClientProvider) ClientFor(username, password string) (clients.RegistryClient, error) {
	return clients.NewRegistryClientWithAuth(username, password)
}

type RegistryCredentialService interface {
	Create(ctx context.Context, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.RegistryCredential, *errors.ServiceError)
	ListByOrgID(ctx context.Context, organisationID string) ([]*models.RegistryCredential, *errors.ServiceError)
	Update(ctx context.Context, ID string, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError)
	Delete(ctx context.Context, ID string) (*RegistryCredentialDeleteResult, *errors.ServiceError)
	Verify(ctx context.Context, ID, repository string, purpose models.RegistryCredentialPurpose) *errors.ServiceError

	// InternalGetForHost returns the decrypted credential covering the given
	// purpose for (org, normalized host). No permission checks; used by the
	// credential resolver.
	InternalGetForHost(ctx context.Context, organisationID, host string, purpose models.RegistryCredentialPurpose) (*models.RegistryCredential, *errors.ServiceError)
	// InternalGetByID returns the decrypted credential by ID. No permission
	// checks; used by the credential resolver for explicit overrides.
	InternalGetByID(ctx context.Context, ID string) (*models.RegistryCredential, *errors.ServiceError)
}

type RegistryCredentialServiceSpec struct {
	Store             stores.RegistryCredentialStore
	StackStore        stores.StackStore
	ReferenceService  ReferenceService
	EncryptionService EncryptionService
	Permissions       auth.PermissionService
	Logger            logger.Logger
	// RegistryClients is optional; it defaults to real registry clients.
	RegistryClients verifyRegistryClientProvider
}

type registryCredentialService struct {
	store             stores.RegistryCredentialStore
	stackStore        stores.StackStore
	referenceService  ReferenceService
	encryptionService EncryptionService
	permissions       auth.PermissionService
	logger            logger.Logger
	registryClients   verifyRegistryClientProvider
}

func NewRegistryCredentialService(spec RegistryCredentialServiceSpec) RegistryCredentialService {
	registryClients := spec.RegistryClients
	if registryClients == nil {
		registryClients = defaultVerifyRegistryClientProvider{}
	}
	return &registryCredentialService{
		store:             spec.Store,
		stackStore:        spec.StackStore,
		referenceService:  spec.ReferenceService,
		encryptionService: spec.EncryptionService,
		permissions:       spec.Permissions,
		logger:            spec.Logger,
		registryClients:   registryClients,
	}
}

func (s *registryCredentialService) Create(ctx context.Context, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, credential.OrganisationID, auth.ResourceRegistryCredentials, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}

	if credential.Host == "" {
		return nil, errors.BadRequest("host is required")
	}
	if credential.Username == "" {
		return nil, errors.BadRequest("username is required")
	}
	if credential.Password == "" {
		return nil, errors.BadRequest("password is required")
	}
	if credential.Purpose == "" {
		credential.Purpose = models.RegistryCredentialPurposeBoth
	}
	if !credential.Purpose.Valid() {
		return nil, errors.BadRequest("unsupported purpose '%s'", credential.Purpose)
	}

	host, err := clients.NormalizeRegistryHostInput(credential.Host)
	if err != nil {
		return nil, errors.BadRequest("invalid registry host '%s'", credential.Host)
	}
	credential.Host = host

	existing, serr := s.store.GetByOrgAndHost(ctx, credential.OrganisationID, credential.Host)
	if serr != nil {
		return nil, serr
	}
	for _, cred := range existing {
		if cred.Purpose.ConflictsWith(credential.Purpose) {
			return nil, errors.Conflict("a registry credential for host '%s' with purpose '%s' already exists", credential.Host, cred.Purpose)
		}
	}

	if serr := s.sealCredential(credential, credential.Password); serr != nil {
		return nil, serr
	}

	return s.store.Create(ctx, credential)
}

func (s *registryCredentialService) GetByID(ctx context.Context, ID string) (*models.RegistryCredential, *errors.ServiceError) {
	credential, serr := s.store.GetByID(ctx, ID)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, credential.OrganisationID, auth.ResourceRegistryCredentials, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return credential, nil
}

func (s *registryCredentialService) ListByOrgID(ctx context.Context, organisationID string) ([]*models.RegistryCredential, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, organisationID, auth.ResourceRegistryCredentials, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.store.ListByOrgID(ctx, organisationID)
}

func (s *registryCredentialService) Update(ctx context.Context, ID string, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError) {
	existing, serr := s.store.GetByID(ctx, ID)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, existing.OrganisationID, auth.ResourceRegistryCredentials, ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	if credential.Host != "" {
		host, err := clients.NormalizeRegistryHostInput(credential.Host)
		if err != nil {
			return nil, errors.BadRequest("invalid registry host '%s'", credential.Host)
		}
		if host != existing.Host {
			return nil, errors.BadRequest("host cannot be updated")
		}
	}
	if credential.Purpose != "" && credential.Purpose != existing.Purpose {
		return nil, errors.BadRequest("purpose cannot be updated")
	}

	if credential.Username != "" {
		existing.Username = credential.Username
	}

	password := credential.Password
	if password == "" {
		// Username-only update: recompute the hash over the stored password.
		decrypted, err := s.encryptionService.DecryptData(existing.EncryptedPassword)
		if err != nil {
			return nil, errors.GeneralError("failed to decrypt registry credential password: %s", err.Error())
		}
		password = string(decrypted)
		existing.DataHash = registryCredentialDataHash(existing.Username, password)
	} else if serr := s.sealCredential(existing, password); serr != nil {
		return nil, serr
	}

	return s.store.Update(ctx, existing)
}

func (s *registryCredentialService) Delete(ctx context.Context, ID string) (*RegistryCredentialDeleteResult, *errors.ServiceError) {
	credential, serr := s.store.GetByID(ctx, ID)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, credential.OrganisationID, auth.ResourceRegistryCredentials, ID, auth.ActionDelete); permErr != nil {
		return nil, permErr
	}

	affected, serr := s.affectedStacks(ctx, credential)
	if serr != nil {
		return nil, serr
	}

	if serr := s.store.Delete(ctx, ID); serr != nil {
		return nil, serr
	}

	return &RegistryCredentialDeleteResult{AffectedStacks: affected}, nil
}

func (s *registryCredentialService) Verify(ctx context.Context, ID, repository string, purpose models.RegistryCredentialPurpose) *errors.ServiceError {
	credential, serr := s.store.GetByID(ctx, ID)
	if serr != nil {
		return serr
	}
	if permErr := s.permissions.Check(ctx, credential.OrganisationID, auth.ResourceRegistryCredentials, ID, auth.ActionRead); permErr != nil {
		return permErr
	}

	if repository == "" {
		return errors.BadRequest("repository is required")
	}
	host, err := clients.NormalizeRegistryHost(repository)
	if err != nil {
		return errors.BadRequest("invalid repository '%s'", repository)
	}
	if host != credential.Host {
		return errors.BadRequest("repository host '%s' does not match credential host '%s'", host, credential.Host)
	}

	if purpose == "" {
		purpose = credential.Purpose
	}
	if !purpose.Valid() {
		return errors.BadRequest("unsupported purpose '%s'", purpose)
	}
	if purpose == models.RegistryCredentialPurposeBoth {
		if credential.Purpose != models.RegistryCredentialPurposeBoth {
			return errors.BadRequest("credential purpose '%s' does not cover '%s'", credential.Purpose, purpose)
		}
	} else if !credential.Purpose.Covers(purpose) {
		return errors.BadRequest("credential purpose '%s' does not cover '%s'", credential.Purpose, purpose)
	}

	decrypted, err := s.encryptionService.DecryptData(credential.EncryptedPassword)
	if err != nil {
		return errors.GeneralError("failed to decrypt registry credential password: %s", err.Error())
	}

	client, err := s.registryClients.ClientFor(credential.Username, string(decrypted))
	if err != nil {
		return errors.GeneralError("failed to create registry client: %s", err.Error())
	}

	if purpose.Covers(models.RegistryCredentialPurposePull) {
		exists, err := client.CheckImage(ctx, repository)
		if err != nil {
			return errors.BadRequest("pull verification failed for '%s': %s", repository, err.Error())
		}
		if !exists {
			return errors.BadRequest("pull verification failed for '%s': image not found or not pullable", repository)
		}
	}
	if purpose.Covers(models.RegistryCredentialPurposePush) {
		if err := client.CheckPushAccess(ctx, repository); err != nil {
			return errors.BadRequest("push verification failed for '%s': %s", repository, err.Error())
		}
	}

	return nil
}

func (s *registryCredentialService) InternalGetForHost(ctx context.Context, organisationID, host string, purpose models.RegistryCredentialPurpose) (*models.RegistryCredential, *errors.ServiceError) {
	credentials, serr := s.store.GetByOrgAndHost(ctx, organisationID, host)
	if serr != nil {
		return nil, serr
	}

	// The purpose-conflict rule on create guarantees at most one credential
	// covers a given purpose.
	for _, credential := range credentials {
		if !credential.Purpose.Covers(purpose) {
			continue
		}
		if serr := s.unsealCredential(credential); serr != nil {
			return nil, serr
		}
		return credential, nil
	}

	return nil, errors.NotFound("no registry credential for host '%s' with purpose '%s'", host, purpose)
}

func (s *registryCredentialService) InternalGetByID(ctx context.Context, ID string) (*models.RegistryCredential, *errors.ServiceError) {
	credential, serr := s.store.GetByID(ctx, ID)
	if serr != nil {
		return nil, serr
	}
	if serr := s.unsealCredential(credential); serr != nil {
		return nil, serr
	}
	return credential, nil
}

// unsealCredential decrypts the stored password onto the transient Password
// field, verifying integrity against the data hash.
func (s *registryCredentialService) unsealCredential(credential *models.RegistryCredential) *errors.ServiceError {
	decrypted, err := s.encryptionService.DecryptData(credential.EncryptedPassword)
	if err != nil {
		return errors.GeneralError("failed to decrypt registry credential password: %s", err.Error())
	}
	if registryCredentialDataHash(credential.Username, string(decrypted)) != credential.DataHash {
		return errors.GeneralError("data integrity check failed for registry credential %s", credential.ID)
	}
	credential.Password = string(decrypted)
	return nil
}

// sealCredential encrypts the given plaintext password onto the credential and
// stamps the data hash. The transient Password field is cleared.
func (s *registryCredentialService) sealCredential(credential *models.RegistryCredential, password string) *errors.ServiceError {
	encrypted, err := s.encryptionService.EncryptData([]byte(password))
	if err != nil {
		return errors.GeneralError("failed to encrypt registry credential password: %s", err.Error())
	}
	credential.EncryptedPassword = encrypted
	credential.DataHash = registryCredentialDataHash(credential.Username, password)
	credential.Password = ""
	return nil
}

func (s *registryCredentialService) affectedStacks(ctx context.Context, credential *models.RegistryCredential) ([]AffectedStackRef, *errors.ServiceError) {
	stacks, serr := s.stackStore.ListByOrganisationID(ctx, credential.OrganisationID)
	if serr != nil {
		return nil, serr
	}

	namesByID := make(map[string]string, len(stacks))
	for _, stack := range stacks {
		namesByID[stack.ID] = stack.Name
	}

	var order []string
	seen := make(map[string]struct{})
	addStack := func(id string) {
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		order = append(order, id)
	}

	// Implicit host-match scan over the current live specs.
	//
	// KNOWN GAP: this only scans live specs, and implicit host-matches are not
	// recorded in the reference graph (only explicit credential IDs are). So a
	// retained (Released) release whose frozen snapshot implicitly resolves
	// against this credential is invisible here — deleting the credential is
	// reported as "0 affected", but a rollback to that release re-resolves the
	// snapshot against the now-empty inventory, syncs no secret, and fails
	// (ImagePullBackOff for a private image; hard failure for short-lived git
	// tokens). Deletion is warn-not-block, so nothing breaks at delete time — the
	// risk is a silent-until-rollback failure. Tracked for a proper fix (extend
	// this scan to Released release snapshots). See the stackdome issue board.
	for _, stack := range stacks {
		if stackResolvesAgainstCredential(stack, credential) {
			addStack(stack.ID)
		}
	}

	// Explicit references from the reference graph, spanning both live specs
	// and retained releases. The boolean is intentionally ignored — deletion is
	// never blocked; this is rotation visibility only.
	_, refs, serr := s.referenceService.IsReferentInUse(ctx, models.ReferentRegistryCredential, credential.ID)
	if serr != nil {
		return nil, serr
	}
	for _, ref := range refs {
		addStack(ref.StackID)
	}

	affected := make([]AffectedStackRef, 0, len(order))
	for _, id := range order {
		affected = append(affected, AffectedStackRef{ID: id, Name: namesByID[id]})
	}
	return affected, nil
}

// stackResolvesAgainstCredential reports whether any resource in the stack
// implicitly resolves against the credential: a matching registry host with no
// explicit secret ref for the covered purpose.
func stackResolvesAgainstCredential(stack *models.Stack, credential *models.RegistryCredential) bool {
	for _, resource := range stack.StackResources {
		if credential.Purpose.Covers(models.RegistryCredentialPurposePull) &&
			resource.ImageConfig != nil && resource.ImageConfig.RegistryCredentialID == "" {
			if host, err := clients.NormalizeRegistryHost(resource.ImageConfig.Image); err == nil && host == credential.Host {
				return true
			}
		}
		if credential.Purpose.Covers(models.RegistryCredentialPurposePush) &&
			resource.BuildConfig != nil && resource.BuildConfig.PushRegistryCredentialID == "" &&
			!resource.BuildConfig.BuildImageRepository.UseInClusterRegistry &&
			resource.BuildConfig.BuildImageRepository.ExternalImageRef != "" {
			if host, err := clients.NormalizeRegistryHost(resource.BuildConfig.BuildImageRepository.ExternalImageRef); err == nil && host == credential.Host {
				return true
			}
		}
	}
	return false
}

func registryCredentialDataHash(username, password string) string {
	jsonData, _ := json.Marshal(map[string]string{
		models.UsernameSecretKey: username,
		models.PasswordSecretKey: password,
	})
	hash := sha256.Sum256(jsonData)
	return base64.StdEncoding.EncodeToString(hash[:])
}
