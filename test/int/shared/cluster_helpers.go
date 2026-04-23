package shared

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	addonsv1alpha1 "stackdome.io/cluster-agent/api/addons/v1alpha1"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"

	// postgres driver for connectivity check
	_ "github.com/lib/pq"
)

// WaitForCRExists polls the cluster until a PostgresCluster CR exists with the given name and namespace.
func WaitForCRExists(ctx context.Context, clusterClient client.Client, name, namespace string, timeout time.Duration) *addonsv1alpha1.PostgresCluster {
	var cr addonsv1alpha1.PostgresCluster
	Eventually(func(g Gomega) {
		err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cr)
		g.Expect(err).NotTo(HaveOccurred(), "PostgresCluster CR should exist")
	}, timeout, 2*time.Second).Should(Succeed())
	return &cr
}

// WaitForAddonReady polls the API until the addon status is "Ready".
func WaitForAddonReady(apiClient *openapi.APIClient, orgID, addonID string, timeout time.Duration) *openapi.PostgresAddon {
	var addon *openapi.PostgresAddon
	Eventually(func(g Gomega) {
		ctx := context.Background()
		resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdGet(ctx, orgID, addonID).Execute()
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(httpResp.StatusCode).To(Equal(200))

		status, ok := resp.GetStatusOk()
		g.Expect(ok).To(BeTrue(), "addon should have status")

		state, stateOk := status.GetStateOk()
		g.Expect(stateOk).To(BeTrue(), "status should have state")
		g.Expect(*state).To(Equal("Ready"), "addon should be Ready, got: %s", *state)
		addon = resp
	}, timeout, 5*time.Second).Should(Succeed())
	return addon
}

// WaitForAddonState polls the API until the addon status matches the expected state.
func WaitForAddonState(apiClient *openapi.APIClient, orgID, addonID, expectedState string, timeout time.Duration) *openapi.PostgresAddon {
	var addon *openapi.PostgresAddon
	Eventually(func(g Gomega) {
		ctx := context.Background()
		resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdGet(ctx, orgID, addonID).Execute()
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(httpResp.StatusCode).To(Equal(200))

		status, ok := resp.GetStatusOk()
		g.Expect(ok).To(BeTrue())

		state, stateOk := status.GetStateOk()
		g.Expect(stateOk).To(BeTrue())
		g.Expect(*state).To(Equal(expectedState))
		addon = resp
	}, timeout, 5*time.Second).Should(Succeed())
	return addon
}

// WaitForConditionTrue polls the API until a specific condition on the addon is True.
func WaitForConditionTrue(apiClient *openapi.APIClient, orgID, addonID, conditionType string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		ctx := context.Background()
		resp, _, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdGet(ctx, orgID, addonID).Execute()
		g.Expect(err).NotTo(HaveOccurred())

		status, ok := resp.GetStatusOk()
		g.Expect(ok).To(BeTrue())

		conditions, condOk := status.GetConditionsOk()
		g.Expect(condOk).To(BeTrue())

		found := false
		for _, c := range conditions {
			if c.GetType() == conditionType && c.GetStatus() == "True" {
				found = true
				break
			}
		}
		g.Expect(found).To(BeTrue(), "condition %s should be True", conditionType)
	}, timeout, 5*time.Second).Should(Succeed())
}

// WaitForAddonDeleted polls the API until the addon returns 404.
func WaitForAddonDeleted(apiClient *openapi.APIClient, orgID, addonID string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		ctx := context.Background()
		_, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdGet(ctx, orgID, addonID).Execute()
		g.Expect(err).To(HaveOccurred())
		g.Expect(httpResp.StatusCode).To(Equal(404))
	}, timeout, 5*time.Second).Should(Succeed())
}

// WaitForCRDeleted polls the cluster until the PostgresCluster CR no longer exists.
func WaitForCRDeleted(ctx context.Context, clusterClient client.Client, name, namespace string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		var cr addonsv1alpha1.PostgresCluster
		err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cr)
		g.Expect(err).To(HaveOccurred(), "PostgresCluster CR should be deleted")
	}, timeout, 2*time.Second).Should(Succeed())
}

// GetPostgresClusterCR retrieves the PostgresCluster CR from the cluster.
func GetPostgresClusterCR(ctx context.Context, clusterClient client.Client, name, namespace string) (*addonsv1alpha1.PostgresCluster, error) {
	var cr addonsv1alpha1.PostgresCluster
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

// VerifyCRSpec asserts that the PostgresCluster CR spec matches expected values.
func VerifyCRSpec(cr *addonsv1alpha1.PostgresCluster, expectedInstances int, expectedMajorVersion int, expectedStorage string) {
	Expect(cr.Spec.Instances).To(Equal(expectedInstances), "CR instances mismatch")
	Expect(cr.Spec.PostgreSQLSpec).NotTo(BeNil(), "PostgreSQLSpec should be set")
	Expect(cr.Spec.PostgreSQLSpec.PostgreSQLMajorVersion).To(Equal(expectedMajorVersion), "CR PostgreSQL major version mismatch")
	Expect(cr.Spec.StorageSpec).NotTo(BeNil(), "StorageSpec should be set")
	Expect(cr.Spec.StorageSpec.Size).To(Equal(expectedStorage), "CR storage size mismatch")
}

// VerifyCRLabel checks that the CR has the expected addon ID label.
func VerifyCRLabel(cr *addonsv1alpha1.PostgresCluster, addonID string) {
	labels := cr.GetLabels()
	Expect(labels).To(HaveKeyWithValue(models.PostgresAddonIDLabel, addonID), "CR should have addon ID label")
}

// GetCredentials fetches JIT credentials for an addon database via the API.
func GetCredentials(apiClient *openapi.APIClient, orgID, addonID, database string) *openapi.PostgresCredentials {
	ctx := context.Background()
	resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdCredentialsDatabaseGet(ctx, orgID, addonID, database).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to get credentials")
	Expect(httpResp.StatusCode).To(Equal(200))
	Expect(resp).NotTo(BeNil())
	return resp
}

// ConnectToPostgres opens a real PostgreSQL connection and runs SELECT 1 to verify connectivity.
func ConnectToPostgres(host string, port int32, username, password, dbName, sslMode string) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, username, password, dbName, sslMode)

	db, err := sql.Open("postgres", connStr)
	Expect(err).NotTo(HaveOccurred(), "failed to open postgres connection")

	err = db.Ping()
	Expect(err).NotTo(HaveOccurred(), "failed to ping postgres")

	return db
}

// CnpgClusterName returns the CNPG Cluster CR name derived from addon name and PG major version.
func CnpgClusterName(addonName string, majorVersion int) string {
	return fmt.Sprintf("%s-%d", addonName, majorVersion)
}

// PortForwardPod sets up port-forwarding to a pod.
// Returns the local port and a stop channel. Close stopChan to tear down the forward.
func PortForwardPod(restConfig *rest.Config, clientset *kubernetes.Clientset, namespace, podName string, remotePort int) (int32, chan struct{}) {
	reqURL := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward").
		URL()

	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	Expect(err).NotTo(HaveOccurred(), "failed to create SPDY round tripper")

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, reqURL)

	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	portMapping := fmt.Sprintf("0:%d", remotePort)
	fw, err := portforward.New(dialer, []string{portMapping}, stopChan, readyChan, nil, nil)
	Expect(err).NotTo(HaveOccurred(), "failed to create port forwarder")

	go func() {
		if err := fw.ForwardPorts(); err != nil {
			select {
			case <-stopChan:
			default:
				fmt.Printf("port-forward error: %v\n", err)
			}
		}
	}()

	select {
	case <-readyChan:
	case <-time.After(30 * time.Second):
		close(stopChan)
		Expect(fmt.Errorf("port-forward ready timeout")).NotTo(HaveOccurred())
	}

	ports, err := fw.GetPorts()
	Expect(err).NotTo(HaveOccurred(), "failed to get forwarded ports")
	Expect(ports).NotTo(BeEmpty())

	return int32(ports[0].Local), stopChan
}

// PortForwardPostgres sets up port-forwarding to the CNPG primary pod for a given addon.
func PortForwardPostgres(ctx context.Context, restConfig *rest.Config, clientset *kubernetes.Clientset, namespace, cnpgClusterName string) (int32, chan struct{}) {
	podName := findCNPGPrimaryPod(ctx, clientset, namespace, cnpgClusterName)
	return PortForwardPod(restConfig, clientset, namespace, podName, 5432)
}

// PortForwardStackResource sets up port-forwarding to a pod backing a StackResource Deployment.
func PortForwardStackResource(ctx context.Context, restConfig *rest.Config, clientset *kubernetes.Clientset, namespace, resourceName string, remotePort int) (int32, chan struct{}) {
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("resource=%s", resourceName),
	})
	Expect(err).NotTo(HaveOccurred(), "failed to list pods for stack resource %s", resourceName)
	Expect(pods.Items).NotTo(BeEmpty(), "no running pod found for stack resource %s", resourceName)

	return PortForwardPod(restConfig, clientset, namespace, pods.Items[0].Name, remotePort)
}

func findCNPGPrimaryPod(ctx context.Context, clientset *kubernetes.Clientset, namespace, clusterName string) string {
	// CNPG labels the primary pod with cnpg.io/cluster=<name> and role=primary
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("cnpg.io/cluster=%s,role=primary", clusterName),
	})
	Expect(err).NotTo(HaveOccurred(), "failed to list CNPG pods")
	Expect(pods.Items).NotTo(BeEmpty(), "no primary pod found for cluster %s", clusterName)

	return pods.Items[0].Name
}

// CRNameForAddon returns the expected PostgresCluster CR name for a given addon.
// The CR name matches the addon name directly (set in postgres_cluster_builder.go).
func CRNameForAddon(addonName string) string {
	return addonName
}

// TriggerBackup triggers an immediate backup for a postgres addon.
func TriggerBackup(apiClient *openapi.APIClient, orgID, addonID string) {
	ctx := context.Background()
	req := openapi.NewApiV1OrganizationsOrgIdAddonsPostgresIdActionsBackupPostRequest()
	req.SetDescription("e2e test backup")
	_, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdActionsBackupPost(ctx, orgID, addonID).
		ApiV1OrganizationsOrgIdAddonsPostgresIdActionsBackupPostRequest(*req).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to trigger backup")
	Expect(httpResp.StatusCode).To(Equal(202))
}

// ListBackups returns all backups for a postgres addon.
func ListBackups(apiClient *openapi.APIClient, orgID, addonID string) []openapi.PostgresBackup {
	ctx := context.Background()
	resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdBackupsGet(ctx, orgID, addonID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list backups")
	Expect(httpResp.StatusCode).To(Equal(200))
	Expect(resp).NotTo(BeNil())
	return resp.GetItems()
}

// WaitForBackupPhase polls until at least one backup reaches the expected phase.
func WaitForBackupPhase(apiClient *openapi.APIClient, orgID, addonID, expectedPhase string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		backups := ListBackups(apiClient, orgID, addonID)
		g.Expect(len(backups)).To(BeNumerically(">=", 1), "should have at least one backup")

		found := false
		for _, b := range backups {
			if b.GetPhase() == expectedPhase {
				found = true
				break
			}
		}
		g.Expect(found).To(BeTrue(), "no backup in phase %s", expectedPhase)
	}, timeout, 10*time.Second).Should(Succeed())
}

// Stack cluster helpers

func WaitForStackCRExists(ctx context.Context, clusterClient client.Client, name, namespace string, timeout time.Duration) *corev1alpha1.Stack {
	var cr corev1alpha1.Stack
	Eventually(func(g Gomega) {
		err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cr)
		g.Expect(err).NotTo(HaveOccurred(), "Stack CR should exist")
	}, timeout, 2*time.Second).Should(Succeed())
	return &cr
}

func WaitForStackReady(apiClient *openapi.APIClient, orgID, stackID string, timeout time.Duration) *openapi.Stack {
	var stack *openapi.Stack
	Eventually(func(g Gomega) {
		ctx := context.Background()
		resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksIdGet(ctx, orgID, stackID).Execute()
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(httpResp.StatusCode).To(Equal(200))

		status, ok := resp.GetStatusOk()
		g.Expect(ok).To(BeTrue(), "stack should have status")

		state, stateOk := status.GetStateOk()
		g.Expect(stateOk).To(BeTrue(), "status should have state")
		g.Expect(*state).To(Equal("Ready"), "stack should be Ready, got: %s", *state)
		stack = resp
	}, timeout, 5*time.Second).Should(Succeed())
	return stack
}

func WaitForStackDeleted(apiClient *openapi.APIClient, orgID, stackID string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		ctx := context.Background()
		_, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksIdGet(ctx, orgID, stackID).Execute()
		g.Expect(err).To(HaveOccurred())
		g.Expect(httpResp.StatusCode).To(Equal(404))
	}, timeout, 2*time.Second).Should(Succeed())
}

func GetStackCR(ctx context.Context, clusterClient client.Client, name, namespace string) (*corev1alpha1.Stack, error) {
	var cr corev1alpha1.Stack
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

func VerifyStackCRLabel(cr *corev1alpha1.Stack, stackID string) {
	labels := cr.GetLabels()
	Expect(labels).To(HaveKeyWithValue(models.StackIDLabel, stackID), "Stack CR should have stack ID label")
}

func WaitForStackResourceCRAvailable(ctx context.Context, clusterClient client.Client, name, namespace string, timeout time.Duration) *corev1alpha1.StackResource {
	var cr corev1alpha1.StackResource
	Eventually(func(g Gomega) {
		err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cr)
		g.Expect(err).NotTo(HaveOccurred(), "StackResource CR should exist")

		available := false
		for _, c := range cr.Status.Conditions {
			if c.Type == string(corev1alpha1.StackResourceStatusAvailable) && c.Status == "True" {
				available = true
				break
			}
		}
		g.Expect(available).To(BeTrue(), "StackResource should be Available")
	}, timeout, 5*time.Second).Should(Succeed())
	return &cr
}

func WaitForStackCRDeleted(ctx context.Context, clusterClient client.Client, name, namespace string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		var cr corev1alpha1.Stack
		err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cr)
		g.Expect(err).To(HaveOccurred(), "Stack CR should be deleted")
	}, timeout, 2*time.Second).Should(Succeed())
}

func GetDeploymentForStackResource(ctx context.Context, clusterClient client.Client, namespace, name string) (*appsv1.Deployment, error) {
	var deploy appsv1.Deployment
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &deploy); err != nil {
		return nil, err
	}
	return &deploy, nil
}

func GetServiceForStackResource(ctx context.Context, clusterClient client.Client, namespace, name string) (*corev1.Service, error) {
	var svc corev1.Service
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &svc); err != nil {
		return nil, err
	}
	return &svc, nil
}

func GetContainerEnvVar(deploy *appsv1.Deployment, envName string) (string, bool) {
	for _, container := range deploy.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == envName {
				return env.Value, true
			}
		}
	}
	return "", false
}

func VerifyGitCredentialsSecretExists(ctx context.Context, clusterClient client.Client, namespace, stackID string) {
	secretList := &corev1.SecretList{}
	err := clusterClient.List(ctx, secretList,
		client.InNamespace(namespace),
		client.MatchingLabels{models.StackIDLabel: stackID},
	)
	Expect(err).NotTo(HaveOccurred(), "failed to list secrets in namespace")

	found := false
	for _, secret := range secretList.Items {
		if strings.HasPrefix(secret.Name, "git-credentials-") {
			found = true
			break
		}
	}
	Expect(found).To(BeTrue(), "expected a git-credentials secret with stack ID label in namespace %s", namespace)
}

func GetIngressForStackResource(ctx context.Context, clusterClient client.Client, namespace, resourceName string) (*networkingv1.Ingress, error) {
	ingressName := fmt.Sprintf("%s-http-proxy", resourceName)
	var ingress networkingv1.Ingress
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: ingressName, Namespace: namespace}, &ingress); err != nil {
		return nil, err
	}
	return &ingress, nil
}

func ListStackBuilds(apiClient *openapi.APIClient, orgID, stackID string) []openapi.ImageBuild {
	ctx := context.Background()
	resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdBuildsGet(ctx, orgID, stackID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list stack builds")
	Expect(httpResp.StatusCode).To(Equal(200))
	Expect(resp).NotTo(BeNil())
	return resp.GetItems()
}

// DumpBuildSourceDebugInfo prints cluster and API state to help diagnose stuck builds.
func DumpBuildSourceDebugInfo(ctx context.Context, apiClient *openapi.APIClient, clusterClient client.Client, clientset *kubernetes.Clientset, orgID, stackID, namespace string) {
	fmt.Println("\n========== BUILD SOURCE DEBUG INFO ==========")

	// Stack status via API
	resp, _, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksIdGet(ctx, orgID, stackID).Execute()
	if err != nil {
		fmt.Printf("[Stack API] error fetching stack: %v\n", err)
	} else if status, ok := resp.GetStatusOk(); ok {
		fmt.Printf("[Stack API] state=%s message=%q\n", status.GetState(), status.GetMessage())
		for _, c := range status.GetConditions() {
			fmt.Printf("[Stack API]   condition: type=%s status=%s reason=%s\n", c.GetType(), c.GetStatus(), c.GetReason())
		}
	}

	// Image builds via API
	buildsResp, _, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdBuildsGet(ctx, orgID, stackID).Execute()
	if err != nil {
		fmt.Printf("[Builds API] error: %v\n", err)
	} else {
		builds := buildsResp.GetItems()
		fmt.Printf("[Builds API] count=%d\n", len(builds))
		for i, b := range builds {
			st := b.GetStatus()
			fmt.Printf("[Builds API]   [%d] resource=%s state=%s image=%s\n", i, b.GetStackResourceName(), st.GetState(), st.GetImageUrl())
			for _, c := range st.GetConditions() {
				fmt.Printf("[Builds API]     condition: type=%s status=%s reason=%s\n", c.GetType(), c.GetStatus(), c.GetReason())
			}
		}
	}

	// Stack CR from cluster
	var stackList corev1alpha1.StackList
	if err := clusterClient.List(ctx, &stackList, client.InNamespace(namespace)); err != nil {
		fmt.Printf("[Cluster] error listing Stack CRs: %v\n", err)
	} else {
		for _, s := range stackList.Items {
			fmt.Printf("[Cluster] Stack CR name=%s phase=%s\n", s.Name, s.Status.Phase)
			for _, c := range s.Status.Conditions {
				fmt.Printf("[Cluster]   condition: type=%s status=%s reason=%s msg=%q\n", c.Type, c.Status, c.Reason, c.Message)
			}
		}
	}

	// StackResource CRs
	var srList corev1alpha1.StackResourceList
	if err := clusterClient.List(ctx, &srList, client.InNamespace(namespace)); err != nil {
		fmt.Printf("[Cluster] error listing StackResource CRs: %v\n", err)
	} else {
		for _, sr := range srList.Items {
			fmt.Printf("[Cluster] StackResource CR name=%s\n", sr.Name)
			for _, c := range sr.Status.Conditions {
				fmt.Printf("[Cluster]   condition: type=%s status=%s reason=%s msg=%q\n", c.Type, c.Status, c.Reason, c.Message)
			}
		}
	}

	// ImageBuild CRs from cluster
	var ibList buildsv1alpha1.ImageBuildList
	if err := clusterClient.List(ctx, &ibList, client.InNamespace(namespace)); err != nil {
		fmt.Printf("[Cluster] error listing ImageBuild CRs: %v\n", err)
	} else {
		fmt.Printf("[Cluster] ImageBuild CRs: %d\n", len(ibList.Items))
		for _, ib := range ibList.Items {
			fmt.Printf("[Cluster]   name=%s phase=%s image=%s\n", ib.Name, ib.Status.Phase, ib.Status.ImageUrl)
			for _, c := range ib.Status.Conditions {
				fmt.Printf("[Cluster]     condition: type=%s status=%s reason=%s msg=%q\n", c.Type, c.Status, c.Reason, c.Message)
			}
		}
	}

	// Pods in namespace
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("[Cluster] error listing pods: %v\n", err)
	} else {
		fmt.Printf("[Cluster] pods in namespace %s: %d\n", namespace, len(pods.Items))
		for _, p := range pods.Items {
			fmt.Printf("[Cluster]   pod=%s phase=%s", p.Name, p.Status.Phase)
			for _, cs := range p.Status.ContainerStatuses {
				if cs.State.Waiting != nil {
					fmt.Printf(" container=%s waiting=%s msg=%q", cs.Name, cs.State.Waiting.Reason, cs.State.Waiting.Message)
				}
				if cs.State.Terminated != nil {
					fmt.Printf(" container=%s terminated=%s exit=%d msg=%q", cs.Name, cs.State.Terminated.Reason, cs.State.Terminated.ExitCode, cs.State.Terminated.Message)
				}
			}
			fmt.Println()
		}
	}

	// Jobs in namespace (Kaniko build jobs)
	jobs, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("[Cluster] error listing jobs: %v\n", err)
	} else {
		fmt.Printf("[Cluster] jobs in namespace %s: %d\n", namespace, len(jobs.Items))
		for _, j := range jobs.Items {
			fmt.Printf("[Cluster]   job=%s active=%d succeeded=%d failed=%d\n", j.Name, j.Status.Active, j.Status.Succeeded, j.Status.Failed)
			for _, c := range j.Status.Conditions {
				fmt.Printf("[Cluster]     condition: type=%s status=%s reason=%s msg=%q\n", c.Type, c.Status, c.Reason, c.Message)
			}
		}
	}

	// Recent events in namespace
	events, err := clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Printf("[Cluster] error listing events: %v\n", err)
	} else {
		fmt.Printf("[Cluster] events in namespace %s (last 20):\n", namespace)
		start := 0
		if len(events.Items) > 20 {
			start = len(events.Items) - 20
		}
		for _, e := range events.Items[start:] {
			fmt.Printf("[Cluster]   %s %s/%s: %s - %s\n", e.LastTimestamp.Format(time.RFC3339), e.InvolvedObject.Kind, e.InvolvedObject.Name, e.Reason, e.Message)
		}
	}

	fmt.Println("========== END DEBUG INFO ==========")
}
