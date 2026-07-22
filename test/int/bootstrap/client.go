package bootstrap

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/testutil"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	controlPlaneNamespace   = "stackdome-control-plane"
	apiServerServiceAccount = "stackdome-api-server-account"
)

type ClientManager struct {
	client       *openapi.APIClient
	userToken    string
	orgID        string
	clusterID    string
	registryName string
	logger       logr.Logger
}

func NewClientManager(baseURL string, logger logr.Logger) *ClientManager {
	config := openapi.NewConfiguration()
	config.Host = strings.TrimPrefix(baseURL, "http://")
	config.Scheme = "http"
	config.UserAgent = "stackdome-integration-tests/1.0"

	client := openapi.NewAPIClient(config)

	return &ClientManager{
		client: client,
		logger: logger,
	}
}

func (cm *ClientManager) Bootstrap(ctx context.Context, platformClusterID string) error {
	cm.logger.Info("Starting client bootstrap")

	// Create context with 10-minute timeout for client bootstrap
	bootstrapCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// The server, on boot, provisioned the infrastructure-only platform org,
	// cluster and base domain from the PLATFORM_* env. Sign up a tenant org for
	// the suite; signup seeds its <slug>.<base> domain and registry on the
	// platform cluster, which stacks resolve via the read-time fallback.
	if err := cm.signupSuiteUser(bootstrapCtx); err != nil {
		return fmt.Errorf("failed to sign up suite user: %w", err)
	}

	cm.configureAuthentication()
	cm.clusterID = platformClusterID

	cm.logger.Info("Waiting for the seeded org image registry to become Running")
	if err := cm.waitForRegistryRunning(bootstrapCtx); err != nil {
		return fmt.Errorf("failed waiting for registry: %w", err)
	}

	cm.logger.Info("Client bootstrap completed successfully",
		"orgID", cm.orgID,
		"clusterID", cm.clusterID,
		"registryName", cm.registryName)
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

func (cm *ClientManager) GetRegistryName() string {
	return cm.registryName
}

func (cm *ClientManager) signupSuiteUser(ctx context.Context) error {
	org := openapi.NewOrganisation()
	org.SetName(suiteOrgName)
	req := openapi.NewUserSignupRequest(suiteUserName, suiteUserEmail, suiteUserPassword)
	req.SetOrganisation(*org)

	resp, httpResp, err := cm.client.DefaultApi.ApiV1UserSignupPost(ctx).UserSignupRequest(*req).Execute()
	if err != nil {
		return fmt.Errorf("suite user signup failed: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	if resp.GetJwtToken() == "" {
		return fmt.Errorf("no token in signup response")
	}
	if resp.User.GetOrganisationId() == "" {
		return fmt.Errorf("no organisation ID in signup response")
	}

	cm.userToken = resp.GetJwtToken()
	cm.orgID = resp.User.GetOrganisationId()
	return nil
}

func (cm *ClientManager) configureAuthentication() {
	// Set authentication context for all future API calls
	cm.client.GetConfig().DefaultHeader["Authorization"] = "Bearer " + cm.userToken
}

func deployAPIServerServiceAccount(ctx context.Context, cluster *testutil.TestCluster) error {
	kubeClient, err := cluster.GetKubeClient()
	if err != nil {
		return fmt.Errorf("failed to get kube client: %w", err)
	}

	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: controlPlaneNamespace,
		},
	}
	_, err = kubeClient.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	// Create service account
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiServerServiceAccount,
			Namespace: controlPlaneNamespace,
		},
	}
	_, err = kubeClient.CoreV1().ServiceAccounts(controlPlaneNamespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create service account: %w", err)
	}

	// Create cluster role — rules mirror install/manifests/rbac.yaml (keep in sync)
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "stackdome-api-server-role",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"core.stackdome.io"},
				Resources: []string{"stacks", "stackresources"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			},
			{
				APIGroups: []string{"storage.stackdome.io"},
				Resources: []string{"volumes"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			},
			{
				APIGroups: []string{"users.stackdome.io"},
				Resources: []string{"users"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			},
			{
				APIGroups: []string{"registry.stackdome.io"},
				Resources: []string{"clusterregistries"},
				Verbs:     []string{"get", "list", "watch", "create", "delete"},
			},
			{
				APIGroups: []string{"addons.stackdome.io"},
				Resources: []string{"postgresclusters"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			},
			{
				APIGroups: []string{"builds.stackdome.io"},
				Resources: []string{"imagebuilds"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"clusterissuers"},
				Verbs:     []string{"get", "create"},
			},
			{
				APIGroups: []string{"postgresql.cnpg.io"},
				Resources: []string{"imagecatalogs"},
				Verbs:     []string{"get", "create"},
			},
			{
				APIGroups: []string{"postgresql.cnpg.io"},
				Resources: []string{"backups"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"barmancloud.cnpg.io"},
				Resources: []string{"objectstores"},
				Verbs:     []string{"get", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs:     []string{"get", "create", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/log"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"list"},
			},
			{
				APIGroups: []string{"metrics.k8s.io"},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list"},
			},
		},
	}
	_, err = kubeClient.RbacV1().ClusterRoles().Create(ctx, clusterRole, metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create cluster role: %w", err)
		}
		if _, err := kubeClient.RbacV1().ClusterRoles().Update(ctx, clusterRole, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to update cluster role: %w", err)
		}
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
				Name:      apiServerServiceAccount,
				Namespace: controlPlaneNamespace,
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
			Namespace: controlPlaneNamespace,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": apiServerServiceAccount,
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}
	_, err = kubeClient.CoreV1().Secrets(controlPlaneNamespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	// Wait for secret to be populated
	for i := 0; i < 30; i++ {
		s, err := kubeClient.CoreV1().Secrets(controlPlaneNamespace).Get(ctx, "stackdome-api-server-account-secret", metav1.GetOptions{})
		if err == nil && len(s.Data["token"]) > 0 && len(s.Data["ca.crt"]) > 0 {
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for service account secret to be populated")
}

func ExtractAPIServerClusterCredentials(ctx context.Context, cluster *testutil.TestCluster) (string, string, string, error) {
	clientset, err := cluster.GetKubeClient()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get kube client: %w", err)
	}

	// Get cluster URL from rest config
	config, err := cluster.GetKubeConfig()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get kube config: %w", err)
	}

	clusterURL := config.Host

	// Get CA certificate and token from secret
	secret, err := clientset.CoreV1().Secrets(controlPlaneNamespace).Get(ctx, "stackdome-api-server-account-secret", metav1.GetOptions{})
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get secret: %w", err)
	}

	// Base64 encode the CA certificate data
	caData := base64.StdEncoding.EncodeToString(secret.Data["ca.crt"])
	saToken := string(secret.Data["token"])

	return clusterURL, caData, saToken, nil
}

func (cm *ClientManager) waitForRegistryRunning(ctx context.Context) error {
	timeout := 5 * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, httpResp, err := cm.client.DefaultApi.ApiV1OrganizationsOrgIdImageRegistriesGet(ctx, cm.orgID).Execute()
		if err != nil {
			cm.logger.Info("Registry list request failed, retrying", "error", err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		if httpResp.StatusCode != http.StatusOK {
			time.Sleep(5 * time.Second)
			continue
		}

		for _, reg := range resp.GetItems() {
			if reg.Status != nil && reg.Status.State != nil && *reg.Status.State == openapi.IMAGE_REGISTRY_RUNNING {
				cm.logger.Info("Cluster image registry is Running", "name", reg.Name)
				cm.registryName = reg.GetName()
				return nil
			}
			if reg.Status != nil {
				cm.logger.Info("Registry not yet Running", "name", reg.Name, "state", string(reg.Status.GetState()))
			}
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timed out after %v waiting for cluster image registry to become Running", timeout)
}
