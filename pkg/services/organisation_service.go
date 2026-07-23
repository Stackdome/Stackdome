package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	"github.com/Stackdome/stackdome/pkg/slug"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
)

//go:generate mockgen -destination=../mocks/mock_organisation_service.go -package=mocks github.com/Stackdome/stackdome/pkg/services OrganisationService

const (
	shortOrgIDLength      = 8
	maxDomainSeedAttempts = 5
)

type OrganisationService interface {
	InternalCreate(ctx context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError)
	InternalGetPlatformOrg(ctx context.Context) (*models.Organisation, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Organisation, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	Update(ctx context.Context, ID string, spec *models.Organisation) (*models.Organisation, *errors.ServiceError)

	PromoteToOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError
	DemoteOrgAdmin(ctx context.Context, orgID, userID, projectID string, role models.ProjectRole) *errors.ServiceError
	ListOrgAdmins(ctx context.Context, orgID string) ([]*models.User, *errors.ServiceError)
}

type organisationService struct {
	organisationStore         stores.OrganisationStore
	organisationDomainService OrganisationDomainsService
	stackQueryService         StackQueryService
	userStore                 stores.UserStore
	clusterStore              stores.ClusterStore
	imageRegistryService      ImageRegistryService
	orgRegistryDefaults       models.OrgRegistryDefaults
	projectService            ProjectService
	atomicExecutor            stores.AtomicExecutor
	policyMgr                 resourceaccess.ResourceAccessPolicyManager
	permissions               auth.PermissionService
	logger                    logger.Logger
}

func NewOrganisationService(spec OrganisationServiceSpec) OrganisationService {
	return &organisationService{
		organisationStore: pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		userStore: pgstore.NewUserStore(pgstore.UserStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		clusterStore: pgstore.NewClusterStore(pgstore.ClusterStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		stackQueryService:         spec.StackQueryService,
		organisationDomainService: spec.OrganisationDomainService,
		imageRegistryService:      spec.ImageRegistryService,
		orgRegistryDefaults:       spec.OrgRegistryDefaults,
		projectService:            spec.ProjectService,
		atomicExecutor:            pgstore.NewAtomicExecutor(spec.SessionFactory),
		policyMgr:                 spec.PolicyManager,
		permissions:               spec.Permissions,
		logger:                    spec.Logger,
	}
}

type OrganisationServiceSpec struct {
	SessionFactory            db.SessionFactory
	OrganisationDomainService OrganisationDomainsService
	ImageRegistryService      ImageRegistryService
	OrgRegistryDefaults       models.OrgRegistryDefaults
	StackQueryService         StackQueryService
	ProjectService            ProjectService
	PolicyManager             resourceaccess.ResourceAccessPolicyManager
	Permissions               auth.PermissionService
	Logger                    logger.Logger
}

func (s *organisationService) InternalCreate(ctx context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
	if len(spec.Name) == 0 {
		return nil, errors.BadRequest("organisation name is required")
	}

	org, err := s.organisationStore.Create(ctx, spec)
	if err != nil {
		s.logger.Error(ctx, "failed to create organisation: %v", err)
		return nil, err
	}

	for _, domain := range spec.Domains {
		domain.OrganisationID = org.ID
		if _, err := s.organisationDomainService.Create(ctx, domain); err != nil {
			return nil, err
		}
	}

	if !org.Platform {
		if seedErr := s.seedPlatformInfra(ctx, org.ID, org.Name); seedErr != nil {
			return nil, seedErr
		}
	}

	return s.organisationStore.Get(ctx, org.ID)
}

// seedPlatformInfra gives a new tenant org a subdomain of the platform base
// domain and a seed registry on the shared platform cluster. No platform
// cluster configured (self-hosted install) → no-op.
func (s *organisationService) seedPlatformInfra(ctx context.Context, orgID, orgName string) *errors.ServiceError {
	platformCluster, err := s.clusterStore.GetPlatformCluster(ctx)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			return nil
		}
		return err
	}
	ctx = auth.SetIdentityInContext(ctx, &auth.Identity{IsSystem: true, OrgID: orgID})

	platformOrg, pErr := s.InternalGetPlatformOrg(ctx)
	if pErr != nil {
		return pErr
	}
	baseDomain, dErr := s.organisationDomainService.GetDefaultDomainForOrganisation(ctx, platformOrg.ID)
	if dErr != nil {
		return dErr
	}

	orgSlug := slug.FromOrgName(orgName)
	if sErr := s.seedOrgDomain(ctx, orgID, orgSlug, baseDomain.Domain); sErr != nil {
		return sErr
	}
	return s.seedOrgRegistry(ctx, orgID, orgSlug, platformCluster.ID)
}

func (s *organisationService) seedOrgDomain(ctx context.Context, orgID, orgSlug, baseDomain string) *errors.ServiceError {
	candidate := fmt.Sprintf("%s.%s", orgSlug, baseDomain)
	for attempt := 0; attempt < maxDomainSeedAttempts; attempt++ {
		_, err := s.organisationDomainService.Create(ctx, &models.OrganisationDomain{
			OrganisationID: orgID,
			Domain:         candidate,
		})
		if err == nil {
			return nil
		}
		if err.Code != errors.ErrorConflict {
			return err
		}
		candidate = fmt.Sprintf("%s-%s.%s", orgSlug, slug.RandomSuffix(), baseDomain)
	}
	return errors.Conflict("could not allocate a unique domain for organisation")
}

func (s *organisationService) seedOrgRegistry(ctx context.Context, orgID, orgSlug, clusterID string) *errors.ServiceError {
	shortID := strings.ReplaceAll(orgID, "-", "")
	if len(shortID) > shortOrgIDLength {
		shortID = shortID[:shortOrgIDLength]
	}
	_, err := s.imageRegistryService.InternalCreateSeedRegistry(ctx, &models.ClusterImageRegistry{
		ClusterID:           clusterID,
		OrganisationID:      orgID,
		Name:                fmt.Sprintf("%s-%s", orgSlug, shortID),
		BackendStorageSize:  s.orgRegistryDefaults.StorageSize,
		BackendStorageClass: s.orgRegistryDefaults.StorageClass,
	})
	return err
}

func (s *organisationService) InternalGetPlatformOrg(ctx context.Context) (*models.Organisation, *errors.ServiceError) {
	return s.organisationStore.GetPlatformOrg(ctx)
}

func (s *organisationService) Get(ctx context.Context, ID string) (*models.Organisation, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, ID, auth.ResourceOrgs, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	org, err := s.organisationStore.Get(ctx, ID)
	if err != nil {
		s.logger.Error(ctx, "failed to get organisation: %v", err)
		return nil, err
	}
	return org, nil
}

// TODO: Org is the root of almost everything, so we need to be careful when deleting it.
func (s *organisationService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	if permErr := s.permissions.Check(ctx, ID, auth.ResourceOrgs, ID, auth.ActionDelete); permErr != nil {
		return permErr
	}
	stacks, err := s.stackQueryService.GetStacksByOrganisationID(ctx, ID)
	if err != nil {
		return err
	}
	if len(stacks) > 0 {
		return errors.BadRequest("cannot delete organisation with stacks")
	}
	err = s.organisationStore.Delete(ctx, ID)
	if err != nil {
		s.logger.Error(ctx, "failed to delete organisation: %v", err)
		return err
	}
	return nil
}

// Update updates an organisation
func (s *organisationService) Update(ctx context.Context, ID string, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, ID, auth.ResourceOrgs, ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	existing, err := s.Get(ctx, ID)
	if err != nil {
		return nil, err
	}

	if spec.Name != "" && existing.Name != spec.Name {
		nameExists, err := s.organisationStore.OrganisationNameExists(ctx, spec.Name)
		if err != nil {
			s.logger.Error(ctx, "failed to check if organisation name exists: %v", err)
			return nil, errors.GeneralError("failed to update organisation")
		}
		if nameExists {
			return nil, errors.Conflict("organisation with the same name already exists")
		}
	}

	if len(spec.Name) == 0 {
		spec.Name = existing.Name
	}
	org, err := s.organisationStore.Update(ctx, ID, spec)
	if err != nil {
		s.logger.Error(ctx, "failed to update organisation: %v", err)
		return nil, err
	}

	existingDomains, err := s.organisationDomainService.ListByOrganisationID(ctx, ID)
	if err != nil {
		return nil, err
	}

	existingDomainsMap := make(map[string]struct{})
	for _, domain := range existingDomains {
		existingDomainsMap[domain.Domain] = struct{}{}
	}

	for _, domain := range spec.Domains {
		if _, exists := existingDomainsMap[domain.Domain]; !exists {
			domain.OrganisationID = org.ID
			if _, err := s.organisationDomainService.Create(ctx, domain); err != nil {
				return nil, err
			}
		}
	}

	currentDomainsMap := make(map[string]struct{})
	for _, domain := range spec.Domains {
		currentDomainsMap[domain.Domain] = struct{}{}
	}
	for _, domain := range existingDomains {
		if _, exists := currentDomainsMap[domain.Domain]; !exists {
			if err := s.organisationDomainService.Delete(ctx, domain.ID); err != nil {
				return nil, err
			}
		}
	}
	return s.Get(ctx, org.ID)
}

func (s *organisationService) PromoteToOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	user, serr := s.userStore.GetByID(ctx, userID)
	if serr != nil {
		return serr
	}

	if user.OrganisationID != orgID {
		return errors.BadRequest("user does not belong to this organisation")
	}

	if user.IsOrgAdmin() {
		return errors.BadRequest("user is already an OrgAdmin")
	}

	user.Role = models.OrgAdminRole
	if _, serr := s.userStore.Update(ctx, userID, user); serr != nil {
		return serr
	}

	if err := s.policyMgr.AddGroupingPolicy(userID, string(models.OrgAdminRole), orgID); err != nil {
		s.logger.Error(ctx, "failed to add OrgAdmin grouping: %s", err.Error())
		return errors.InternalServerError("failed to add OrgAdmin grouping")
	}

	return nil
}

func (s *organisationService) DemoteOrgAdmin(ctx context.Context, orgID, userID, projectID string, role models.ProjectRole) *errors.ServiceError {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	admins, serr := s.userStore.ListByOrgAndRole(ctx, orgID, models.OrgAdminRole)
	if serr != nil {
		return serr
	}
	if len(admins) <= 1 {
		return errors.BadRequest("cannot demote the last OrgAdmin")
	}

	user, serr := s.userStore.GetByID(ctx, userID)
	if serr != nil {
		return serr
	}

	if user.OrganisationID != orgID {
		return errors.BadRequest("user does not belong to this organisation")
	}

	if !user.IsOrgAdmin() {
		return errors.BadRequest("user is not an OrgAdmin")
	}

	return s.atomicExecutor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		user.Role = models.NoRole
		if _, serr := s.userStore.Update(txCtx, userID, user); serr != nil {
			return serr
		}

		if _, serr := s.projectService.InternalAddMember(txCtx, projectID, userID, role); serr != nil {
			s.logger.Error(ctx, "failed to add demoted user to project: %s", serr.Error())
			return serr
		}

		if err := s.policyMgr.RemoveGroupingPolicy(userID, string(models.OrgAdminRole), orgID); err != nil {
			s.logger.Error(ctx, "failed to remove OrgAdmin grouping: %s", err.Error())
			return errors.InternalServerError("failed to remove OrgAdmin grouping")
		}

		return nil
	})
}

func (s *organisationService) ListOrgAdmins(ctx context.Context, orgID string) ([]*models.User, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.userStore.ListByOrgAndRole(ctx, orgID, models.OrgAdminRole)
}
