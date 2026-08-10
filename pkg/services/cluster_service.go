package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

const httpsScheme = "https"

//go:generate mockgen -destination=../mocks/mock_cluster_service.go -package=mocks github.com/Stackdome/stackdome/pkg/services ClusterService

type ClusterService interface {
	BackgroundJobEnqueuerInjectable
	GetClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError)
	GetOwnedClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError)
	InternalGet(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	AddCluster(ctx context.Context, cluster *models.Cluster) (*models.Cluster, *errors.ServiceError)
	InternalUpsertSharedComputeCluster(ctx context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError)
	InternalListAllClusters(ctx context.Context) ([]*models.Cluster, *errors.ServiceError)
	InjectClusterManager(clusterManager clustermanager.ClusterManager)
	InternalUpdateClusterInfo(ctx context.Context, clusterID string, info *models.ClusterInfo) *errors.ServiceError
	InternalEnsurePlatformWildcardTLS(ctx context.Context, cluster *models.Cluster, cfg *config.BootstrapConfig) *errors.ServiceError
	DefaultStorageClass(ctx context.Context, clusterID string) (string, *errors.ServiceError)
}

type clusterService struct {
	clusterStore         stores.ClusterStore
	organisationStore    stores.OrganisationStore
	computeMode          config.ComputeMode
	platformTLSEnabled   bool
	logger               logger.Logger
	clusterManager       clustermanager.ClusterManager
	imageRegistryService ImageRegistryService
	permissions          auth.PermissionService
	encryptionService    EncryptionService
	BackgroundJobEnqueuerDep
}

func NewClusterService(spec ClusterServiceSpec) ClusterService {
	clusterStore := spec.ClusterStore
	if clusterStore == nil {
		clusterStore = pgstore.NewClusterStore(pgstore.ClusterStoreSpec{
			SessionFactory: spec.SessionFactory,
		})
	}
	return &clusterService{
		clusterStore: clusterStore,
		organisationStore: pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		clusterManager:       spec.ClusterManager,
		logger:               spec.Logger,
		computeMode:          spec.ComputeMode,
		platformTLSEnabled:   spec.PlatformTLSEnabled,
		imageRegistryService: spec.ImageRegistryService,
		permissions:          spec.Permissions,
		encryptionService:    spec.EncryptionService,
	}
}

type ClusterServiceSpec struct {
	SessionFactory       db.SessionFactory
	ClusterStore         stores.ClusterStore
	ClusterManager       clustermanager.ClusterManager
	ImageRegistryService ImageRegistryService
	ComputeMode          config.ComputeMode
	PlatformTLSEnabled   bool
	Permissions          auth.PermissionService
	EncryptionService    EncryptionService
	Logger               logger.Logger
}

// inject cluster manager
func (s *clusterService) InjectClusterManager(clusterManager clustermanager.ClusterManager) {
	s.clusterManager = clusterManager
}

// InternalListAllClusters lists all clusters in the database
func (s *clusterService) InternalListAllClusters(ctx context.Context) ([]*models.Cluster, *errors.ServiceError) {
	clusters, err := s.clusterStore.ListAll(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to list all clusters: %v", err)
		return nil, err
	}
	for _, cluster := range clusters {
		if decErr := s.decryptClusterCredentials(cluster); decErr != nil {
			s.logger.Error(ctx, "failed to decrypt cluster credentials for cluster %s: %v", cluster.ID, decErr)
			return nil, decErr
		}
	}
	return clusters, nil
}

func (s *clusterService) AddCluster(ctx context.Context, cluster *models.Cluster) (*models.Cluster, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, cluster.OrganisationID, auth.ResourceClusters, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}
	if s.computeMode == config.ComputeModeShared {
		return nil, errors.BadRequest("tenant clusters cannot be added when shared compute is enabled")
	}
	if s.computeMode != config.ComputeModeBYOC {
		return nil, errors.GeneralError("unsupported compute mode %q", s.computeMode)
	}

	var createdCluster *models.Cluster
	cerr := s.clusterStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		if err := s.organisationStore.LockByID(txCtx, cluster.OrganisationID); err != nil {
			return err
		}

		existingClusters, err := s.clusterStore.ListBYOCClustersForOrg(txCtx, cluster.OrganisationID)
		if err != nil {
			s.logger.Error(txCtx, "failed to list existing BYOC clusters for org: %v", err)
			return err
		}
		if len(existingClusters) != 0 {
			return errors.Conflict("cluster already exists for org")
		}

		createdCluster, err = s.createClusterWithTx(txCtx, cluster)
		return err
	})
	if cerr != nil {
		s.logger.Error(ctx, "failed to create cluster: %v", cerr)
		return nil, cerr
	}

	s.ensureIssuerForCreatedCluster(ctx, createdCluster)
	return createdCluster, nil
}

func (s *clusterService) createClusterWithTx(ctx context.Context, cluster *models.Cluster) (*models.Cluster, *errors.ServiceError) {
	// The caller must enforce its access and compute-topology policy before
	// entering this shared persistence path, and must hold the owning
	// organisation's row lock in the current transaction.

	// Validate the cluster
	err := s.validateCluster(ctx, cluster)
	if err != nil {
		s.logger.Error(ctx, "failed to validate cluster: %v", err)
		return nil, err
	}

	if encErr := s.encryptClusterCredentials(cluster); encErr != nil {
		s.logger.Error(ctx, "failed to encrypt cluster credentials: %v", encErr)
		return nil, encErr
	}

	// Builds resolve the registry on the stack's cluster, so every owned
	// cluster ships with one: fall back to a default when none was supplied.
	// The platform org stays infrastructure-only — no registry on its cluster.
	var registry *models.ClusterImageRegistry
	org, oerr := s.organisationStore.Get(ctx, cluster.OrganisationID)
	if oerr != nil {
		s.logger.Error(ctx, "failed to get organisation for default registry: %v", oerr)
		return nil, oerr
	}
	if len(cluster.ImageRegistries) != 0 {
		registry = cluster.ImageRegistries[0]
	} else if !org.Platform {
		registry = &models.ClusterImageRegistry{}
	}

	createdCluster, createdErr := s.clusterStore.CreateWithTx(ctx, cluster)
	if createdErr != nil {
		return nil, createdErr
	}
	if hookErr := db.OnRollback(ctx, func(context.Context) error {
		return s.clusterManager.UnregisterCluster(createdCluster.ID)
	}); hookErr != nil {
		return nil, errors.GeneralError("failed to register cluster manager rollback compensation: %s", hookErr.Error())
	}
	if registerErr := s.clusterManager.RegisterCluster(createdCluster); registerErr != nil {
		return nil, errors.GeneralError("failed to register cluster with manager: %s", registerErr.Error())
	}

	if registry != nil {
		if registry.Name == "" {
			registry.Name = orgRegistryName(org.Name, org.ID, createdCluster.ID)
		}
		registry.ClusterID = createdCluster.ID
		registry.OrganisationID = createdCluster.OrganisationID
		createdRegistry, err := s.imageRegistryService.CreateWithTx(ctx, registry)
		if err != nil {
			s.logger.Error(ctx, "failed to create image registry: %v", err)
			return nil, err
		}
		createdCluster.ImageRegistries = []*models.ClusterImageRegistry{createdRegistry}
	}

	return createdCluster, nil
}

func (s *clusterService) ensureIssuerForCreatedCluster(ctx context.Context, createdCluster *models.Cluster) {
	shouldEnsureIssuer := !createdCluster.SharedCompute || s.platformTLSEnabled
	if shouldEnsureIssuer {
		if err := s.ensureClusterIssuer(ctx, createdCluster); err != nil {
			s.logger.Error(ctx, "failed to create ClusterIssuer on cluster %s: %v", createdCluster.ID, err)
		}
	}
}

func (s *clusterService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	cluster, err := s.clusterStore.Get(ctx, ID)
	if err != nil {
		s.logger.Error(ctx, "failed to get cluster: %v", err)
		return err
	}
	if permErr := s.permissions.Check(ctx, cluster.OrganisationID, auth.ResourceClusters, ID, auth.ActionDelete); permErr != nil {
		return permErr
	}
	deleteErr := s.clusterStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		if lockErr := s.organisationStore.LockByID(txCtx, cluster.OrganisationID); lockErr != nil {
			return lockErr
		}
		current, getErr := s.clusterStore.Get(txCtx, ID)
		if getErr != nil {
			return getErr
		}
		if current.DeletionTimestamp != nil {
			return nil
		}
		if storeErr := s.clusterStore.MarkDeletingWithTx(txCtx, ID, time.Now().UTC()); storeErr != nil {
			return storeErr
		}
		if registryErr := s.imageRegistryService.InternalMarkAllDeletingByClusterIDWithTx(txCtx, ID); registryErr != nil {
			return registryErr
		}
		if s.BackgroundJobEnqueuer == nil {
			return errors.GeneralError("background job enqueuer is not configured")
		}
		if enqueueErr := s.BackgroundJobEnqueuer.EnqueueAfterCommit(txCtx, models.ClusterImageRegistryOperand{ClusterID: ID}); enqueueErr != nil {
			return errors.GeneralError("failed to enqueue cluster deletion: %s", enqueueErr.Error())
		}
		return nil
	})
	if deleteErr != nil {
		s.logger.Error(ctx, "failed to delete cluster: %v", deleteErr)
	}
	return deleteErr
}

func (s *clusterService) validateCluster(ctx context.Context, cluster *models.Cluster) *errors.ServiceError {
	if cluster == nil {
		return errors.BadRequest("cluster cannot be nil")
	}

	if cluster.Name == "" {
		return errors.BadRequest("cluster name cannot be empty")
	}

	if cluster.OrganisationID == "" {
		return errors.BadRequest("organisation ID cannot be empty")
	}

	if cluster.ClusterCAData == "" {
		return errors.BadRequest("cluster CA data cannot be empty")
	}

	if cluster.ClusterURL == "" {
		return errors.BadRequest("cluster URL cannot be empty")
	}
	if cluster.Token == "" {
		return errors.BadRequest("cluster token cannot be empty")
	}

	// validation for cluster url
	url, err := url.Parse(cluster.ClusterURL)
	if err != nil {
		return errors.BadRequest("cluster URL is not valid: %s", err.Error())
	}
	if url.Scheme != httpsScheme {
		return errors.BadRequest("cluster URL must use https scheme")
	}

	existingCluster, serr := s.clusterStore.GetByClusterUrl(ctx, cluster.ClusterURL)
	if serr != nil && serr.Code != errors.ErrorNotFound {
		s.logger.Errorf("failed to get cluster by URL: %v", serr)
		return serr
	}

	if existingCluster != nil && existingCluster.ID != cluster.ID {
		return errors.Conflict("cluster with this api URL already exists")
	}

	var (
		clusterCADataDecoded []byte
		derr                 error
	)
	if IsBase64(cluster.ClusterCAData) {
		clusterCADataDecoded, derr = base64.StdEncoding.DecodeString(cluster.ClusterCAData)
		if derr != nil {
			return errors.BadRequest("cluster CA data is not valid base64: %s", derr.Error())
		}
		if len(clusterCADataDecoded) == 0 {
			return errors.BadRequest("cluster CA data is empty after decoding")
		}
	} else {
		cluster.ClusterCAData = base64.StdEncoding.EncodeToString([]byte(cluster.ClusterCAData))
	}

	if IsBase64(cluster.Token) {
		tokenDecoded, derr := base64.StdEncoding.DecodeString(cluster.Token)
		if derr != nil {
			return errors.BadRequest("cluster token is not valid base64: %s", derr.Error())
		}
		if len(tokenDecoded) == 0 {
			return errors.BadRequest("cluster token is empty after decoding")
		}
	} else {
		cluster.Token = base64.StdEncoding.EncodeToString([]byte(cluster.Token))
	}
	if _, err := certutil.NewPoolFromBytes(clusterCADataDecoded); err != nil {
		return errors.BadRequest("cluster CA data is not valid: %s", err.Error())
	}
	return nil
}

func (s *clusterService) PersistManagerState(ctx context.Context, clusterID string, running bool) error {
	err := s.clusterStore.PersistManagerState(ctx, clusterID, running)
	if err != nil {
		return fmt.Errorf("failed to persist cluster manager state: %w", err)
	}
	return nil
}

// GetOwnedClusterForOrg returns the BYOC cluster owned by the org.
func (s *clusterService) GetOwnedClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceClusters, "", auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.resolveBYOCClusterForOrg(ctx, orgID)
}

func (s *clusterService) GetClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError) {
	if err := s.permissions.Check(ctx, orgID, auth.ResourceClusters, "", auth.ActionRead); err != nil {
		return nil, err
	}
	switch s.computeMode {
	case config.ComputeModeBYOC:
		return s.resolveBYOCClusterForOrg(ctx, orgID)
	case config.ComputeModeShared:
		return s.resolveSharedComputeClusterForOrg(ctx, orgID)
	default:
		return nil, errors.GeneralError("unsupported compute mode %q", s.computeMode)
	}
}

func (s *clusterService) resolveBYOCClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError) {
	clusters, err := s.clusterStore.ListBYOCClustersForOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if len(clusters) == 0 {
		return nil, errors.NotFound("cluster for organisation '%s' not found", orgID)
	}
	if len(clusters) > 1 {
		return nil, errors.GeneralError("multiple BYOC clusters found for organisation '%s'", orgID)
	}
	cluster := clusters[0]
	if decErr := s.decryptClusterCredentials(cluster); decErr != nil {
		s.logger.Error(ctx, "failed to decrypt cluster credentials: %v", decErr)
		return nil, decErr
	}
	return cluster, nil
}

func (s *clusterService) resolveSharedComputeClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError) {
	clusters, err := s.clusterStore.ListSharedComputeClusters(ctx)
	if err != nil {
		return nil, err
	}
	if len(clusters) == 0 {
		return nil, errors.NotFound("shared-compute cluster for organisation '%s' not found", orgID)
	}
	if len(clusters) > 1 {
		return nil, errors.GeneralError("multiple shared-compute clusters configured")
	}
	cluster := clusters[0]
	if decErr := s.decryptClusterCredentials(cluster); decErr != nil {
		s.logger.Error(ctx, "failed to decrypt cluster credentials: %v", decErr)
		return nil, decErr
	}
	return cluster, nil
}

func (s *clusterService) Get(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError) {
	cluster, err := s.clusterStore.Get(ctx, ID)
	if err != nil {
		s.logger.Error(ctx, "failed to get cluster: %v", err)
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, cluster.OrganisationID, auth.ResourceClusters, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	if decErr := s.decryptClusterCredentials(cluster); decErr != nil {
		s.logger.Error(ctx, "failed to decrypt cluster credentials: %v", decErr)
		return nil, decErr
	}
	return cluster, nil
}

func (s *clusterService) InternalGet(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError) {
	cluster, err := s.clusterStore.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	if decErr := s.decryptClusterCredentials(cluster); decErr != nil {
		return nil, decErr
	}
	return cluster, nil
}

func (s *clusterService) encryptClusterCredentials(cluster *models.Cluster) *errors.ServiceError {
	encryptedToken, err := s.encryptionService.EncryptData([]byte(cluster.Token))
	if err != nil {
		return errors.GeneralError("failed to encrypt cluster token: %v", err)
	}
	encryptedCAData, err := s.encryptionService.EncryptData([]byte(cluster.ClusterCAData))
	if err != nil {
		return errors.GeneralError("failed to encrypt cluster CA data: %v", err)
	}
	cluster.EncryptedToken = encryptedToken
	cluster.EncryptedClusterCAData = encryptedCAData
	return nil
}

func (s *clusterService) decryptClusterCredentials(cluster *models.Cluster) *errors.ServiceError {
	token, err := s.encryptionService.DecryptData(cluster.EncryptedToken)
	if err != nil {
		return errors.GeneralError("failed to decrypt cluster token: %v", err)
	}
	caData, err := s.encryptionService.DecryptData(cluster.EncryptedClusterCAData)
	if err != nil {
		return errors.GeneralError("failed to decrypt cluster CA data: %v", err)
	}
	cluster.Token = string(token)
	cluster.ClusterCAData = string(caData)
	return nil
}

func (s *clusterService) ensureClusterIssuer(ctx context.Context, cluster *models.Cluster) error {
	k8sClient, err := s.clusterManager.GetClient(cluster.ID)
	if err != nil {
		return fmt.Errorf("getting cluster client: %w", err)
	}

	email := auth.ContactEmailFromCtx(ctx)
	if email == "" {
		return fmt.Errorf("no ACME contact email available for ClusterIssuer")
	}

	issuer := &cmv1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: models.DefaultClusterIssuerName,
			Labels: map[string]string{
				"app.kubernetes.io/part-of":    "stackdome",
				"app.kubernetes.io/managed-by": "stackdome-api-server",
			},
		},
		Spec: cmv1.IssuerSpec{
			IssuerConfig: cmv1.IssuerConfig{
				ACME: &cmacme.ACMEIssuer{
					Server: "https://acme-v02.api.letsencrypt.org/directory",
					Email:  email,
					PrivateKey: cmmeta.SecretKeySelector{
						LocalObjectReference: cmmeta.LocalObjectReference{
							Name: "letsencrypt-prod-key",
						},
					},
					Solvers: []cmacme.ACMEChallengeSolver{{
						HTTP01: &cmacme.ACMEChallengeSolverHTTP01{
							Ingress: &cmacme.ACMEChallengeSolverHTTP01Ingress{
								Class: ptr.To("traefik"),
							},
						},
					}},
				},
			},
		},
	}

	existing := &cmv1.ClusterIssuer{}
	if gerr := k8sClient.Get(ctx, client.ObjectKeyFromObject(issuer), existing); gerr != nil {
		if !k8serrors.IsNotFound(gerr) {
			return fmt.Errorf("checking existing ClusterIssuer: %w", gerr)
		}
		if cerr := k8sClient.Create(ctx, issuer); cerr != nil {
			return fmt.Errorf("creating ClusterIssuer: %w", cerr)
		}
		s.logger.Info(ctx, "created ClusterIssuer %s on cluster %s", models.DefaultClusterIssuerName, cluster.ID)
		return nil
	}

	s.logger.Info(ctx, "ClusterIssuer %s already exists on cluster %s, skipping", models.DefaultClusterIssuerName, cluster.ID)
	return nil
}

// InternalEnsurePlatformWildcardTLS creates or reconciles the namespace-scoped
// DNS-01 resources used to issue a wildcard certificate for the platform apps
// domain.
func (s *clusterService) InternalEnsurePlatformWildcardTLS(ctx context.Context, cluster *models.Cluster, cfg *config.BootstrapConfig) *errors.ServiceError {
	if cfg.BaseDomain == "" {
		return nil
	}

	k8sClient, err := s.clusterManager.GetClient(cluster.ID)
	if err != nil {
		return errors.GeneralError("failed to get client for cluster %s: %v", cluster.ID, err)
	}

	email := auth.ContactEmailFromCtx(ctx)
	if email == "" {
		return errors.GeneralError("no ACME contact email available for DNS Issuer %s", models.DNSIssuerName)
	}

	if err := ensureTLSNamespace(ctx, k8sClient, cfg.TLSNamespace); err != nil {
		return errors.GeneralError("failed to ensure Namespace %s: %v", cfg.TLSNamespace, err)
	}
	if err := ensureCloudflareToken(ctx, k8sClient, cfg.TLSNamespace, cfg.DNSCloudflareAPIToken); err != nil {
		return errors.GeneralError("failed to ensure Secret %s: %v", models.CloudflareAPITokenSecretName, err)
	}
	if err := ensureDNSIssuer(ctx, k8sClient, cfg.TLSNamespace, email, cfg.ACMEDirectoryURL()); err != nil {
		return errors.GeneralError("failed to ensure Issuer %s: %v", models.DNSIssuerName, err)
	}
	if err := ensureWildcardCertificate(ctx, k8sClient, cfg.TLSNamespace, cfg.BaseDomain); err != nil {
		return errors.GeneralError("failed to ensure Certificate %s: %v", models.PlatformWildcardTLSName, err)
	}

	s.logger.Info(ctx, "platform wildcard TLS ensured for *.%s on cluster %s", cfg.BaseDomain, cluster.ID)
	return nil
}

func ensureTLSNamespace(ctx context.Context, k8sClient client.Client, namespace string) error {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, ns, func() error {
		return nil
	})
	return err
}

func ensureCloudflareToken(ctx context.Context, k8sClient client.Client, namespace, token string) error {
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:      models.CloudflareAPITokenSecretName,
		Namespace: namespace,
	}}
	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, secret, func() error {
		secret.Data = map[string][]byte{models.CloudflareAPITokenSecretKey: []byte(token)}
		return nil
	})
	return err
}

func ensureDNSIssuer(ctx context.Context, k8sClient client.Client, namespace, email, acmeDirectoryURL string) error {
	issuer := &cmv1.Issuer{ObjectMeta: metav1.ObjectMeta{
		Name:      models.DNSIssuerName,
		Namespace: namespace,
	}}
	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, issuer, func() error {
		issuer.Spec = cmv1.IssuerSpec{IssuerConfig: cmv1.IssuerConfig{ACME: &cmacme.ACMEIssuer{
			Server: acmeDirectoryURL,
			Email:  email,
			PrivateKey: cmmeta.SecretKeySelector{
				LocalObjectReference: cmmeta.LocalObjectReference{Name: models.DNSIssuerPrivateKeySecretName},
			},
			Solvers: []cmacme.ACMEChallengeSolver{{DNS01: &cmacme.ACMEChallengeSolverDNS01{
				Cloudflare: &cmacme.ACMEIssuerDNS01ProviderCloudflare{APIToken: &cmmeta.SecretKeySelector{
					LocalObjectReference: cmmeta.LocalObjectReference{Name: models.CloudflareAPITokenSecretName},
					Key:                  models.CloudflareAPITokenSecretKey,
				}},
			}}},
		}}}
		return nil
	})
	return err
}

func ensureWildcardCertificate(ctx context.Context, k8sClient client.Client, namespace, baseDomain string) error {
	certificate := &cmv1.Certificate{ObjectMeta: metav1.ObjectMeta{
		Name:      models.PlatformWildcardTLSName,
		Namespace: namespace,
	}}
	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, certificate, func() error {
		certificate.Spec = cmv1.CertificateSpec{
			SecretName: models.PlatformWildcardTLSName,
			DNSNames:   []string{"*." + baseDomain},
			SecretTemplate: &cmv1.CertificateSecretTemplate{Labels: map[string]string{
				corev1alpha1.LabelPlatformWildcardTLSSecret: "true",
			}},
			IssuerRef: cmmeta.IssuerReference{
				Name:  models.DNSIssuerName,
				Kind:  cmv1.IssuerKind,
				Group: cmv1.SchemeGroupVersion.Group,
			},
		}
		return nil
	})
	return err
}

func (s *clusterService) InternalUpsertSharedComputeCluster(ctx context.Context, spec *models.Cluster) (*models.Cluster, *errors.ServiceError) {
	// Shared-compute bootstrap is a trusted internal caller, but the topology is
	// valid only when the installation explicitly uses shared compute.
	if s.computeMode != config.ComputeModeShared {
		return nil, errors.BadRequest("shared-compute clusters require shared compute mode")
	}
	spec.SharedCompute = true

	// Canonicalize before create AND compare, so the stored encoding never
	// depends on which path first persisted the credentials.
	spec.Token = normalizeBase64(spec.Token)
	spec.ClusterCAData = normalizeBase64(spec.ClusterCAData)

	var (
		result  *models.Cluster
		created bool
	)
	txErr := s.clusterStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		if err := s.organisationStore.LockByID(txCtx, spec.OrganisationID); err != nil {
			return err
		}

		owned, err := s.clusterStore.ListSharedComputeClustersForOrg(txCtx, spec.OrganisationID)
		if err != nil {
			return err
		}
		if len(owned) > 1 {
			return errors.GeneralError("multiple shared-compute clusters found for organisation '%s'", spec.OrganisationID)
		}

		var existing *models.Cluster
		if len(owned) == 1 {
			existing = owned[0]
		} else {
			existing, err = s.clusterStore.GetByClusterUrl(txCtx, spec.ClusterURL)
			if err != nil && err.Code != errors.ErrorNotFound {
				return err
			}
			if existing != nil && existing.OrganisationID != spec.OrganisationID {
				return errors.Conflict("cluster with this api URL already exists")
			}
			if existing == nil {
				result, err = s.createClusterWithTx(txCtx, spec)
				created = err == nil
				return err
			}
		}

		spec.ID = existing.ID
		if validateErr := s.validateCluster(txCtx, spec); validateErr != nil {
			return validateErr
		}
		if decErr := s.decryptClusterCredentials(existing); decErr != nil {
			return decErr
		}
		previousManagerCluster := *existing
		metadataChanged := existing.Name != spec.Name || !existing.SharedCompute || existing.ClusterURL != spec.ClusterURL
		credentialsChanged := existing.Token != spec.Token || existing.ClusterCAData != spec.ClusterCAData
		if !metadataChanged && !credentialsChanged {
			result = existing
			return nil
		}
		updated := &models.Cluster{
			ID:             existing.ID,
			OrganisationID: existing.OrganisationID,
			Name:           spec.Name,
			SharedCompute:  true,
			ClusterURL:     spec.ClusterURL,
			Token:          spec.Token,
			ClusterCAData:  spec.ClusterCAData,
		}
		if encErr := s.encryptClusterCredentials(updated); encErr != nil {
			return encErr
		}
		if uErr := s.clusterStore.UpdateSharedComputeCluster(txCtx, updated); uErr != nil {
			return uErr
		}
		result, err = s.clusterStore.Get(txCtx, existing.ID)
		if err != nil {
			return err
		}
		reregister := existing.ClusterURL != spec.ClusterURL || credentialsChanged
		if reregister {
			if hookErr := db.OnRollback(txCtx, func(context.Context) error {
				return s.clusterManager.ReRegisterCluster(&previousManagerCluster)
			}); hookErr != nil {
				return errors.GeneralError("failed to register cluster manager rollback compensation: %s", hookErr.Error())
			}
			if rErr := s.clusterManager.ReRegisterCluster(result); rErr != nil {
				return errors.GeneralError("failed to re-register cluster with manager: %s", rErr.Error())
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	if created {
		s.ensureIssuerForCreatedCluster(ctx, result)
		return result, nil
	}
	if decErr := s.decryptClusterCredentials(result); decErr != nil {
		return nil, decErr
	}
	return result, nil
}

func (s *clusterService) InternalUpdateClusterInfo(ctx context.Context, clusterID string, info *models.ClusterInfo) *errors.ServiceError {
	return s.clusterStore.UpdateClusterInfo(ctx, clusterID, info)
}

func (s *clusterService) DefaultStorageClass(ctx context.Context, clusterID string) (string, *errors.ServiceError) {
	cluster, err := s.clusterStore.Get(ctx, clusterID)
	if err != nil {
		return "", err
	}
	return cluster.ClusterInfo.DefaultStorageClass(), nil
}

func IsBase64(s string) bool {
	// Base64 string must be a multiple of 4
	if len(s)%4 != 0 {
		return false
	}

	// Basic character check: must only contain valid base64 characters
	if strings.ContainsAny(s, " \t\r\n") {
		return false
	}

	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func normalizeBase64(s string) string {
	if IsBase64(s) {
		return s
	}
	return base64.StdEncoding.EncodeToString([]byte(s))
}
