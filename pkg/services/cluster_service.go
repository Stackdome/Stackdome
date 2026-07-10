package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	cmacme "github.com/cert-manager/cert-manager/pkg/apis/acme/v1"
	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const httpsScheme = "https"

//go:generate mockgen -destination=../mocks/mock_cluster_service.go -package=mocks github.com/Stackdome/stackdome/pkg/services ClusterService

type ClusterService interface {
	GetClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError)
	GetDefaultCluster(ctx context.Context) (*models.Cluster, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError)
	InternalGet(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	AddCluster(ctx context.Context, cluster *models.Cluster) (*models.Cluster, *errors.ServiceError)
	InternalListAllClusters(ctx context.Context) ([]*models.Cluster, *errors.ServiceError)
	InjectClusterManager(clusterManager clustermanager.ClusterManager)
}

type clusterService struct {
	clusterStore         stores.ClusterStore
	logger               logger.Logger
	clusterManager       clustermanager.ClusterManager
	imageRegistryService ImageRegistryService
	permissions          auth.PermissionService
	encryptionService    EncryptionService
}

func NewClusterService(spec ClusterServiceSpec) ClusterService {
	return &clusterService{
		clusterStore: pgstore.NewClusterStore(pgstore.ClusterStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		clusterManager:       spec.ClusterManager,
		logger:               spec.Logger,
		imageRegistryService: spec.ImageRegistryService,
		permissions:          spec.Permissions,
		encryptionService:    spec.EncryptionService,
	}
}

type ClusterServiceSpec struct {
	SessionFactory       db.SessionFactory
	ClusterManager       clustermanager.ClusterManager
	ImageRegistryService ImageRegistryService
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
		s.logger.Errorf("failed to list all clusters: %v", err)
		return nil, err
	}
	for _, cluster := range clusters {
		if decErr := s.decryptClusterCredentials(cluster); decErr != nil {
			s.logger.Errorf("failed to decrypt cluster credentials for cluster %s: %v", cluster.ID, decErr)
			return nil, decErr
		}
	}
	return clusters, nil
}

func (s *clusterService) AddCluster(ctx context.Context, cluster *models.Cluster) (*models.Cluster, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, cluster.OrganisationID, auth.ResourceClusters, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}
	// Check if the cluster already exists for the org
	existingCluster, err := s.clusterStore.GetClusterForOrg(ctx, cluster.OrganisationID)
	if err != nil && err.Code != errors.ErrorNotFound {
		s.logger.Errorf("failed to get existing cluster for org: %v", err)
		return nil, err
	}
	if existingCluster != nil {
		return nil, errors.Conflict("cluster already exists for org")
	}
	var (
		createdCluster *models.Cluster
		createdErr     *errors.ServiceError
	)

	// Validate the cluster
	err = s.validateCluster(cluster)
	if err != nil {
		s.logger.Errorf("failed to validate cluster: %v", err)
		return nil, err
	}

	if encErr := s.encryptClusterCredentials(cluster); encErr != nil {
		s.logger.Errorf("failed to encrypt cluster credentials: %v", encErr)
		return nil, encErr
	}

	cerr := s.clusterStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		createdCluster, createdErr = s.clusterStore.CreateWithTx(ctx, cluster)
		if createdErr != nil {
			return createdErr
		}

		merr := s.clusterManager.RegisterCluster(createdCluster)
		if merr != nil {
			return errors.GeneralError("failed to register cluster with manager")
		}
		if len(cluster.ImageRegistries) != 0 {
			// Create the image registry in the cluster
			registry := cluster.ImageRegistries[0]
			registry.ClusterID = createdCluster.ID
			registry.OrganisationID = createdCluster.OrganisationID
			createdRegistry, err := s.imageRegistryService.CreateWithTx(ctx, registry)
			if err != nil {
				s.logger.Errorf("failed to create image registry: %v", err)
				return err
			}
			createdCluster.ImageRegistries = []*models.ClusterImageRegistry{createdRegistry}
		}
		return nil
	})
	if cerr != nil {
		s.logger.Errorf("failed to create cluster: %v", cerr)
		return nil, cerr
	}

	if err := s.ensureClusterIssuer(ctx, createdCluster); err != nil {
		s.logger.Errorf("failed to create ClusterIssuer on cluster %s: %v", createdCluster.ID, err)
	}

	return createdCluster, nil
}

func (s *clusterService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	cluster, err := s.clusterStore.Get(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get cluster: %v", err)
		return err
	}
	if permErr := s.permissions.Check(ctx, cluster.OrganisationID, auth.ResourceClusters, ID, auth.ActionDelete); permErr != nil {
		return permErr
	}
	// // Unregister the cluster from the cluster manager
	// cerr := s.clusterManager.UnregisterCluster(cluster.ID)
	// if cerr != nil {
	// 	s.logger.Errorf("failed to unregister cluster from manager: %v", cerr)
	// 	return errors.GeneralError("failed to unregister cluster from manager: %v", cerr)
	// }

	if len(cluster.ImageRegistries) != 0 {
		for _, registry := range cluster.ImageRegistries {
			// Delete the image registry from the cluster
			if registry == nil {
				s.logger.Errorf("image registry is nil")
				return errors.GeneralError("image registry is nil")
			}
			s.logger.Infof("Deleting image registry %s", registry.ID)
			err := s.imageRegistryService.Delete(ctx, cluster.OrganisationID, registry.ID)
			if err != nil {
				s.logger.Errorf("failed to delete image registry: %v", err)
				return err
			}
		}
	}

	// Delete the cluster from the database
	err = s.clusterStore.Delete(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to delete cluster: %v", err)
		return err
	}
	return nil
}

func (s *clusterService) validateCluster(cluster *models.Cluster) *errors.ServiceError {
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

	existingCluster, serr := s.clusterStore.GetByClusterUrl(context.Background(), cluster.ClusterURL)
	if serr != nil && serr.Code != errors.ErrorNotFound {
		s.logger.Errorf("failed to get cluster by URL: %v", serr)
		return serr
	}

	if existingCluster != nil {
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

func (s *clusterService) GetClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceClusters, "", auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	cluster, err := s.clusterStore.GetClusterForOrg(ctx, orgID)
	if err != nil {
		s.logger.Errorf("failed to get cluster for org: %v", err)
		return nil, err
	}
	if decErr := s.decryptClusterCredentials(cluster); decErr != nil {
		s.logger.Errorf("failed to decrypt cluster credentials: %v", decErr)
		return nil, decErr
	}
	return cluster, nil
}

func (s *clusterService) GetDefaultCluster(ctx context.Context) (*models.Cluster, *errors.ServiceError) {
	cluster, err := s.clusterStore.GetDefaultCluster(ctx)
	if err != nil {
		s.logger.Errorf("failed to get default cluster: %v", err)
		return nil, err
	}
	if decErr := s.decryptClusterCredentials(cluster); decErr != nil {
		s.logger.Errorf("failed to decrypt cluster credentials: %v", decErr)
		return nil, decErr
	}
	return cluster, nil
}

func (s *clusterService) Get(ctx context.Context, ID string) (*models.Cluster, *errors.ServiceError) {
	cluster, err := s.clusterStore.Get(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get cluster: %v", err)
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, cluster.OrganisationID, auth.ResourceClusters, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	if decErr := s.decryptClusterCredentials(cluster); decErr != nil {
		s.logger.Errorf("failed to decrypt cluster credentials: %v", decErr)
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

	user, uerr := auth.GetCurrentUserFromCtx(ctx)
	if uerr != nil {
		return fmt.Errorf("getting current user: %w", uerr)
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
					Email:  user.Email,
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
		s.logger.Infof("created ClusterIssuer %s on cluster %s", models.DefaultClusterIssuerName, cluster.ID)
		return nil
	}

	s.logger.Infof("ClusterIssuer %s already exists on cluster %s, skipping", models.DefaultClusterIssuerName, cluster.ID)
	return nil
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
