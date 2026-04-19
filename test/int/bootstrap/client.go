package bootstrap

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/testutil"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

type ClientManager struct {
	client    *openapi.APIClient
	cluster   *testutil.TestCluster
	baseURL   string
	userToken string
	orgID     string
	clusterID string
	logger    logr.Logger
}

type UserInfo struct {
	ID    string
	Token string
	OrgID string
}

func NewClientManager(baseURL string, cluster *testutil.TestCluster, logger logr.Logger) *ClientManager {
	config := openapi.NewConfiguration()
	config.Host = strings.TrimPrefix(baseURL, "http://")
	config.Scheme = "http"
	config.UserAgent = "stackdome-integration-tests/1.0"

	client := openapi.NewAPIClient(config)

	return &ClientManager{
		client:  client,
		cluster: cluster,
		baseURL: baseURL,
		logger:  logger,
	}
}

func (cm *ClientManager) Bootstrap(ctx context.Context) error {
	cm.logger.Info("Starting client bootstrap")

	// Create context with 10-minute timeout for client bootstrap
	bootstrapCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Create test user and get authentication token
	userInfo, err := cm.createTestUser(bootstrapCtx)
	if err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}

	cm.userToken = userInfo.Token
	cm.orgID = userInfo.OrgID

	// Configure client with authentication
	cm.configureAuthentication()

	// Register cluster with API server
	clusterID, err := cm.registerCluster(bootstrapCtx)
	if err != nil {
		return fmt.Errorf("failed to register cluster: %w", err)
	}

	cm.clusterID = clusterID

	cm.logger.Info("Client bootstrap completed successfully",
		"orgID", cm.orgID,
		"clusterID", cm.clusterID)
	return nil
}

func (cm *ClientManager) GetClient() *openapi.APIClient {
	return cm.client
}

func (cm *ClientManager) GetUserToken() string {
	return cm.userToken
}

func (cm *ClientManager) GetOrgID() string {
	return cm.orgID
}

func (cm *ClientManager) GetClusterID() string {
	return cm.clusterID
}

func (cm *ClientManager) createTestUser(ctx context.Context) (*UserInfo, error) {
	// Create user signup request
	email := fmt.Sprintf("test-%d@example.com", time.Now().Unix())
	req := openapi.NewUserSignupRequest("Test User", email, "testpassword123")
	organisation := openapi.NewOrganisation()
	organisation.SetName("stackdome")
	req.SetOrganisation(*organisation)

	cm.logger.Info("Creating test user: %+v", *req)

	// Make signup request
	resp, httpResp, err := cm.client.DefaultApi.ApiV1UserSignupPost(ctx).UserSignupRequest(*req).Execute()
	if err != nil {
		return nil, fmt.Errorf("user signup failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("user signup failed with status %d: %s", httpResp.StatusCode, string(body))
	}

	if resp.GetJwtToken() == "" {
		return nil, fmt.Errorf("no JWT token in signup response")
	}

	if resp.User.GetId() == "" {
		return nil, fmt.Errorf("no user ID in signup response")
	}

	if resp.User.GetOrganisationId() == "" {
		return nil, fmt.Errorf("no organisation ID in signup response")
	}

	return &UserInfo{
		ID:    resp.User.GetId(),
		Token: resp.GetJwtToken(),
		OrgID: resp.User.GetOrganisationId(),
	}, nil
}

func (cm *ClientManager) configureAuthentication() {
	// Set authentication context for all future API calls
	cm.client.GetConfig().DefaultHeader["Authorization"] = "Bearer " + cm.userToken
}

func (cm *ClientManager) registerCluster(ctx context.Context) (string, error) {
	// Deploy service account and get credentials
	if err := cm.deployServiceAccount(ctx); err != nil {
		return "", fmt.Errorf("failed to deploy service account: %w", err)
	}

	// Extract cluster connection details
	clusterURL, caData, saToken, err := cm.extractClusterCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to extract cluster credentials: %w", err)
	}

	// Register cluster via API
	clusterID, err := cm.registerClusterViaAPI(ctx, clusterURL, caData, saToken)
	if err != nil {
		return "", fmt.Errorf("failed to register cluster via API: %w", err)
	}

	return clusterID, nil
}

func (cm *ClientManager) deployServiceAccount(ctx context.Context) error {
	kubeClient, err := cm.cluster.GetKubeClient()
	if err != nil {
		return fmt.Errorf("failed to get kube client: %w", err)
	}

	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "stackdome-control-plane",
		},
	}
	_, err = kubeClient.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// Create service account
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "stackdome-api-server-account",
			Namespace: "stackdome-control-plane",
		},
	}
	_, err = kubeClient.CoreV1().ServiceAccounts("stackdome-control-plane").Create(ctx, sa, metav1.CreateOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create service account: %w", err)
	}

	// Create cluster role
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "stackdome-api-server-role",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}
	_, err = kubeClient.RbacV1().ClusterRoles().Create(ctx, clusterRole, metav1.CreateOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create cluster role: %w", err)
	}

	// Create cluster role binding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "stackdome-api-server-role-binding",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "stackdome-api-server-role",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "stackdome-api-server-account",
				Namespace: "stackdome-control-plane",
			},
		},
	}
	_, err = kubeClient.RbacV1().ClusterRoleBindings().Create(ctx, clusterRoleBinding, metav1.CreateOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create cluster role binding: %w", err)
	}

	// Create secret for service account token
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "stackdome-api-server-account-secret",
			Namespace: "stackdome-control-plane",
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": "stackdome-api-server-account",
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}
	_, err = kubeClient.CoreV1().Secrets("stackdome-control-plane").Create(ctx, secret, metav1.CreateOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	// Wait for secret to be populated
	for i := 0; i < 30; i++ {
		s, err := kubeClient.CoreV1().Secrets("stackdome-control-plane").Get(ctx, "stackdome-api-server-account-secret", metav1.GetOptions{})
		if err == nil && len(s.Data["token"]) > 0 && len(s.Data["ca.crt"]) > 0 {
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for service account secret to be populated")
}

func (cm *ClientManager) extractClusterCredentials(ctx context.Context) (string, string, string, error) {
	clientset, err := cm.cluster.GetKubeClient()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get kube client: %w", err)
	}

	// Get cluster URL from rest config
	config, err := cm.cluster.GetKubeConfig()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get kube config: %w", err)
	}

	clusterURL := config.Host

	// Get CA certificate and token from secret
	secret, err := clientset.CoreV1().Secrets("stackdome-control-plane").Get(ctx, "stackdome-api-server-account-secret", metav1.GetOptions{})
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get secret: %w", err)
	}

	// Base64 encode the CA certificate data
	caData := base64.StdEncoding.EncodeToString(secret.Data["ca.crt"])
	saToken := string(secret.Data["token"])

	return clusterURL, caData, saToken, nil
}

func (cm *ClientManager) registerClusterViaAPI(ctx context.Context, clusterURL, caData, saToken string) (string, error) {
	// Create cluster payload with image registry
	cluster := openapi.Cluster{
		Name:           "test-cluster",
		ClusterUrl:     clusterURL,
		ClusterCaData:  caData,
		ClusterSaToken: saToken,
		ClusterImageRegistry: &openapi.ClusterImageRegistry{
			Name: "test-registry",
			Spec: &openapi.ClusterImageRegistrySpec{
				BackendStorageSize:  ptr.To("10Gi"),
				BackendStorageClass: ptr.To("standard"),
				MaxRepositories:     ptr.To(int32(100)),
				TagsPerRepository:   ptr.To(int32(50)),
				DeleteUntagged:      ptr.To(bool(true)),
			},
		},
	}

	// Marshall to JSON
	payload, err := json.Marshal(cluster)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cluster payload: %w", err)
	}

	// Make API request to register cluster
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/organizations/%s/clusters", cm.baseURL, cm.orgID), bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cm.userToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to register cluster: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response to get cluster ID
	var responseCluster openapi.Cluster
	if err := json.NewDecoder(resp.Body).Decode(&responseCluster); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if responseCluster.Id == nil {
		return "", fmt.Errorf("cluster ID not found in response")
	}

	return *responseCluster.Id, nil
}
