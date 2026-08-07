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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
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
func WaitForAddonReady(apiClient *openapi.APIClient, orgID, projectName, addonID string, timeout time.Duration) *openapi.PostgresAddon {
	var addon *openapi.PostgresAddon
	Eventually(func(g Gomega) {
		ctx := context.Background()
		resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet(ctx, orgID, projectName, addonID).Execute()
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
func WaitForAddonState(apiClient *openapi.APIClient, orgID, projectName, addonID, expectedState string, timeout time.Duration) *openapi.PostgresAddon {
	var addon *openapi.PostgresAddon
	Eventually(func(g Gomega) {
		ctx := context.Background()
		resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet(ctx, orgID, projectName, addonID).Execute()
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
func WaitForConditionTrue(apiClient *openapi.APIClient, orgID, projectName, addonID, conditionType string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		ctx := context.Background()
		resp, _, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet(ctx, orgID, projectName, addonID).Execute()
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
func WaitForAddonDeleted(apiClient *openapi.APIClient, orgID, projectName, addonID string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		ctx := context.Background()
		_, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet(ctx, orgID, projectName, addonID).Execute()
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
func GetCredentials(apiClient *openapi.APIClient, orgID, projectName, addonID, database string) *openapi.PostgresCredentials {
	ctx := context.Background()
	resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdCredentialsDatabaseGet(ctx, orgID, projectName, addonID, database).Execute()
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

// GetSuperuserCredentials fetches JIT superuser credentials for an addon via the API.
// The database path param is required by the route but ignored by the service in superuser mode.
func GetSuperuserCredentials(apiClient *openapi.APIClient, orgID, projectName, addonID string) *openapi.PostgresCredentials {
	ctx := context.Background()
	resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdCredentialsDatabaseGet(ctx, orgID, projectName, addonID, "postgres").
		Superuser(true).
		Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to get superuser credentials")
	Expect(httpResp.StatusCode).To(Equal(200))
	Expect(resp).NotTo(BeNil())
	return resp
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
func TriggerBackup(apiClient *openapi.APIClient, orgID, projectName, addonID string) {
	ctx := context.Background()
	req := openapi.NewApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest()
	req.SetDescription("e2e test backup")
	_, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost(ctx, orgID, projectName, addonID).
		ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest(*req).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to trigger backup")
	Expect(httpResp.StatusCode).To(Equal(202))
}

// ListBackups returns all backups for a postgres addon.
func ListBackups(apiClient *openapi.APIClient, orgID, projectName, addonID string) []openapi.PostgresBackup {
	ctx := context.Background()
	resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdBackupsGet(ctx, orgID, projectName, addonID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list backups")
	Expect(httpResp.StatusCode).To(Equal(200))
	Expect(resp).NotTo(BeNil())
	return resp.GetItems()
}

// WaitForBackupPhase polls until at least one backup reaches the expected phase.
func WaitForBackupPhase(apiClient *openapi.APIClient, orgID, projectName, addonID, expectedPhase string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		backups := ListBackups(apiClient, orgID, projectName, addonID)
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

// WaitForStackActive polls the stack API until it responds 200 with an
// active lifecycle. The wait for stacks that never deploy (e.g. skip-cluster-
// provisioning fixtures): they have no releases, so WaitForStackReady's
// converged_release gate would never pass.
func WaitForStackActive(apiClient *openapi.APIClient, orgID, projectName, stackID string, timeout time.Duration) *openapi.Stack {
	var stack *openapi.Stack
	Eventually(func(g Gomega) {
		resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet(context.Background(), orgID, projectName, stackID).Execute()
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(httpResp.StatusCode).To(Equal(200))
		g.Expect(resp.GetLifecycle()).To(Equal(openapi.STACK_LIFECYCLE_ACTIVE))
		stack = resp
	}, timeout, 2*time.Second).Should(Succeed())
	return stack
}

// WaitForStackReady polls the stack API until the current release has
// converged (Released) and its live status is healthy (Ok) — the
// release-centric equivalent of the old "stack.status.state == Ready" wait.
func WaitForStackReady(apiClient *openapi.APIClient, orgID, projectName, stackID string, timeout time.Duration) *openapi.Stack {
	var stack *openapi.Stack
	Eventually(func(g Gomega) {
		stack = assertStackConverged(g, apiClient, orgID, projectName, stackID)
	}, timeout, 5*time.Second).Should(Succeed())
	return stack
}

// WaitForStackReadyWithDump is WaitForStackReady plus a cluster-state dump
// (ImageBuild CRs, Jobs, Pods, Events in the stack namespace) appended to the
// failure message when the wait times out.
func WaitForStackReadyWithDump(apiClient *openapi.APIClient, orgID, projectName, stackID string, timeout time.Duration, clusterClient client.Client, namespace string) *openapi.Stack {
	var stack *openapi.Stack
	Eventually(func(g Gomega) {
		stack = assertStackConverged(g, apiClient, orgID, projectName, stackID)
	}, timeout, 5*time.Second).Should(Succeed(), func() string {
		return DumpNamespaceDiagnostics(context.Background(), clusterClient, namespace)
	})
	return stack
}

func assertStackConverged(g Gomega, apiClient *openapi.APIClient, orgID, projectName, stackID string) *openapi.Stack {
	ctx := context.Background()
	resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet(ctx, orgID, projectName, stackID).Execute()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(httpResp.StatusCode).To(Equal(200))

	current, ok := resp.GetConvergedReleaseOk()
	g.Expect(ok).To(BeTrue(), "stack should have a converged_release")

	state, stateOk := current.GetStateOk()
	g.Expect(stateOk).To(BeTrue(), "converged_release should have a state")
	if *state == openapi.RELEASE_STATE_FAILED {
		StopTrying(fmt.Sprintf("stack %s release %s went terminal Failed while waiting for Released", stackID, current.GetId())).Now()
	}
	g.Expect(*state).To(Equal(openapi.RELEASE_STATE_RELEASED), "current release should be Released, got: %s", *state)

	health, healthOk := current.GetHealthOk()
	g.Expect(healthOk).To(BeTrue(), "converged_release should have health")
	g.Expect(*health).To(Equal(openapi.RELEASE_HEALTH_OK), "current release should be healthy, got: %s", *health)

	return resp
}

// DumpNamespaceDiagnostics renders the build-relevant state of a namespace —
// ImageBuild CRs (phase + conditions), Jobs, Pods (phase + container states),
// and Events — for inclusion in a timeout failure message.
func DumpNamespaceDiagnostics(ctx context.Context, clusterClient client.Client, namespace string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n===== diagnostics for namespace %s =====\n", namespace)

	builds := &buildsv1alpha1.ImageBuildList{}
	if err := clusterClient.List(ctx, builds, client.InNamespace(namespace)); err != nil {
		fmt.Fprintf(&b, "imagebuilds: list error: %v\n", err)
	} else {
		for _, ib := range builds.Items {
			fmt.Fprintf(&b, "imagebuild %s phase=%s imageUrl=%q\n", ib.Name, ib.Status.Phase, ib.Status.ImageUrl)
			for _, c := range ib.Status.Conditions {
				fmt.Fprintf(&b, "  condition %s=%s reason=%s msg=%s at=%s\n", c.Type, c.Status, c.Reason, c.Message, c.LastTransitionTime.Format(time.RFC3339))
			}
			if d := ib.Status.LastBuildFailureDetail; d != nil {
				fmt.Fprintf(&b, "  lastBuildFailureDetail: %+v\n", *d)
			}
		}
		if len(builds.Items) == 0 {
			fmt.Fprintf(&b, "imagebuilds: none\n")
		}
	}

	jobs := &batchv1.JobList{}
	if err := clusterClient.List(ctx, jobs, client.InNamespace(namespace)); err != nil {
		fmt.Fprintf(&b, "jobs: list error: %v\n", err)
	} else {
		for _, j := range jobs.Items {
			fmt.Fprintf(&b, "job %s active=%d succeeded=%d failed=%d suspend=%v\n", j.Name, j.Status.Active, j.Status.Succeeded, j.Status.Failed, j.Spec.Suspend != nil && *j.Spec.Suspend)
			for _, c := range j.Status.Conditions {
				fmt.Fprintf(&b, "  condition %s=%s reason=%s msg=%s at=%s\n", c.Type, c.Status, c.Reason, c.Message, c.LastTransitionTime.Format(time.RFC3339))
			}
		}
		if len(jobs.Items) == 0 {
			fmt.Fprintf(&b, "jobs: none\n")
		}
	}

	pods := &corev1.PodList{}
	if err := clusterClient.List(ctx, pods, client.InNamespace(namespace)); err != nil {
		fmt.Fprintf(&b, "pods: list error: %v\n", err)
	} else {
		for _, p := range pods.Items {
			fmt.Fprintf(&b, "pod %s phase=%s\n", p.Name, p.Status.Phase)
			for _, cs := range p.Status.ContainerStatuses {
				fmt.Fprintf(&b, "  container %s ready=%v state=%+v\n", cs.Name, cs.Ready, cs.State)
			}
			for _, c := range p.Status.Conditions {
				if c.Status != corev1.ConditionTrue {
					fmt.Fprintf(&b, "  condition %s=%s reason=%s msg=%s\n", c.Type, c.Status, c.Reason, c.Message)
				}
			}
		}
		if len(pods.Items) == 0 {
			fmt.Fprintf(&b, "pods: none\n")
		}
	}

	events := &corev1.EventList{}
	if err := clusterClient.List(ctx, events, client.InNamespace(namespace)); err != nil {
		fmt.Fprintf(&b, "events: list error: %v\n", err)
	} else {
		for _, e := range events.Items {
			fmt.Fprintf(&b, "event %s %s %s %s/%s: %s\n", e.LastTimestamp.Format(time.RFC3339), e.Type, e.Reason, e.InvolvedObject.Kind, e.InvolvedObject.Name, e.Message)
		}
	}
	fmt.Fprintf(&b, "===== end diagnostics =====\n")
	return b.String()
}

func WaitForStackDeleted(apiClient *openapi.APIClient, orgID, projectName, stackID string, timeout time.Duration) {
	ctx := context.Background()
	_, httpResp, _ := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet(ctx, orgID, projectName, stackID).Execute()
	if httpResp != nil && httpResp.StatusCode == 404 {
		return
	}
	Eventually(func(g Gomega) {
		_, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet(context.Background(), orgID, projectName, stackID).Execute()
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

func GetStackResourceCR(ctx context.Context, clusterClient client.Client, name, namespace string) (*corev1alpha1.StackResource, error) {
	var cr corev1alpha1.StackResource
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

// DumpCrashDetectionDebugInfo prints the raw StackResource CR status, Deployment conditions,
// and pod container statuses to help debug crash detection failures.
func DumpCrashDetectionDebugInfo(ctx context.Context, clusterClient client.Client, namespace, resourceName string) {
	fmt.Printf("\n========== CRASH DETECTION DEBUG INFO (namespace=%s resource=%s) ==========\n", namespace, resourceName)

	var sr corev1alpha1.StackResource
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: resourceName, Namespace: namespace}, &sr); err != nil {
		fmt.Printf("[StackResource CR] error fetching: %v\n", err)
	} else {
		fmt.Printf("[StackResource CR] phase=%s statusHash=%s lastFailureDeploymentRevision=%q\n",
			sr.Status.Phase, sr.Status.StatusHash, sr.Status.LastFailureDeploymentRevision)
		fmt.Printf("[StackResource CR] lastFailureDetails count=%d\n", len(sr.Status.LastFailureDetails))
		for i, d := range sr.Status.LastFailureDetails {
			fmt.Printf("  [%d] container=%s restarts=%d reason=%q exitCode=%v message=%q\n",
				i, d.ContainerName, d.RestartCount, d.LastTerminationReason, d.LastTerminationExitCode, d.LastTerminationMessage)
		}
		fmt.Printf("[StackResource CR] conditions:\n")
		for _, c := range sr.Status.Conditions {
			fmt.Printf("  type=%s status=%s reason=%s message=%q\n", c.Type, c.Status, c.Reason, c.Message)
		}
	}

	var deploy appsv1.Deployment
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: resourceName, Namespace: namespace}, &deploy); err != nil {
		fmt.Printf("[Deployment] error fetching: %v\n", err)
	} else {
		fmt.Printf("[Deployment] revision=%s availableReplicas=%d readyReplicas=%d\n",
			deploy.Annotations["deployment.kubernetes.io/revision"],
			deploy.Status.AvailableReplicas, deploy.Status.ReadyReplicas)
		for _, c := range deploy.Status.Conditions {
			fmt.Printf("  condition type=%s status=%s reason=%s message=%q\n", c.Type, c.Status, c.Reason, c.Message)
		}
	}

	var podList corev1.PodList
	if err := clusterClient.List(ctx, &podList, client.InNamespace(namespace), client.MatchingLabels(map[string]string{"resource": resourceName})); err != nil {
		fmt.Printf("[Pods] error listing: %v\n", err)
	} else {
		fmt.Printf("[Pods] found %d pods\n", len(podList.Items))
		for _, pod := range podList.Items {
			fmt.Printf("  pod=%s phase=%s\n", pod.Name, pod.Status.Phase)
			allStatuses := append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...)
			for _, cs := range allStatuses {
				state := "unknown"
				switch {
				case cs.State.Running != nil:
					state = "running"
				case cs.State.Waiting != nil:
					state = fmt.Sprintf("waiting reason=%s", cs.State.Waiting.Reason)
				case cs.State.Terminated != nil:
					state = fmt.Sprintf("terminated exitCode=%d reason=%s", cs.State.Terminated.ExitCode, cs.State.Terminated.Reason)
				}
				fmt.Printf("    container=%s restarts=%d state=%s\n", cs.Name, cs.RestartCount, state)
				if cs.LastTerminationState.Terminated != nil {
					t := cs.LastTerminationState.Terminated
					fmt.Printf("      lastTermination exitCode=%d reason=%s\n", t.ExitCode, t.Reason)
				}
			}
		}
	}

	var rsList appsv1.ReplicaSetList
	if err := clusterClient.List(ctx, &rsList, client.InNamespace(namespace), client.MatchingLabels(map[string]string{"resource": resourceName})); err != nil {
		fmt.Printf("[ReplicaSets] error listing: %v\n", err)
	} else {
		fmt.Printf("[ReplicaSets] found %d replicasets\n", len(rsList.Items))
		for _, rs := range rsList.Items {
			fmt.Printf("  rs=%s revision=%s replicas=%d\n",
				rs.Name, rs.Annotations["deployment.kubernetes.io/revision"], rs.Status.Replicas)
		}
	}

	fmt.Println("========== END CRASH DETECTION DEBUG INFO ==========")
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

func GetContainerEnvVarSecretRef(deploy *appsv1.Deployment, envName string) (*corev1.SecretKeySelector, bool) {
	for _, container := range deploy.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == envName && env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				return env.ValueFrom.SecretKeyRef, true
			}
		}
	}
	return nil, false
}

func ResolveContainerEnvVar(ctx context.Context, clusterClient client.Client, deploy *appsv1.Deployment, namespace, envName string) (string, bool) {
	if val, found := GetContainerEnvVar(deploy, envName); found && val != "" {
		return val, true
	}
	ref, found := GetContainerEnvVarSecretRef(deploy, envName)
	if !found {
		return "", false
	}
	secret := &corev1.Secret{}
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: ref.Name, Namespace: namespace}, secret); err != nil {
		return "", false
	}
	val, ok := secret.Data[ref.Key]
	if !ok {
		return "", false
	}
	return string(val), true
}

func GetContainerVolumeMount(deploy *appsv1.Deployment, mountPath string) (*corev1.VolumeMount, bool) {
	for _, container := range deploy.Spec.Template.Spec.Containers {
		for i, vm := range container.VolumeMounts {
			if vm.MountPath == mountPath {
				return &container.VolumeMounts[i], true
			}
		}
	}
	return nil, false
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
		if strings.HasPrefix(secret.Name, "git-integration-") {
			found = true
			break
		}
	}
	Expect(found).To(BeTrue(), "expected a git-integration credential secret with stack ID label in namespace %s", namespace)
}

func GetIngressesForStackResource(ctx context.Context, clusterClient client.Client, resource *corev1alpha1.StackResource) ([]networkingv1.Ingress, error) {
	var ingressList networkingv1.IngressList
	if err := clusterClient.List(ctx, &ingressList, client.InNamespace(resource.Namespace)); err != nil {
		return nil, err
	}

	ingresses := make([]networkingv1.Ingress, 0, len(ingressList.Items))
	for _, ingress := range ingressList.Items {
		if metav1.IsControlledBy(&ingress, resource) {
			ingresses = append(ingresses, ingress)
		}
	}
	return ingresses, nil
}

// getStackResourceLiveStatus resolves the named resource's live status by fetching
// the stack's release (current if already converged, otherwise latest, since a
// resource can be crash-looping without its release ever converging) and reading
// live_status.resources[name] off the release detail.
func getStackResourceLiveStatus(apiClient *openapi.APIClient, orgID, projectName, stackID, resourceName string) (*openapi.StackResourceStatus, bool) {
	ctx := context.Background()
	stack, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet(ctx, orgID, projectName, stackID).Execute()
	Expect(err).NotTo(HaveOccurred())
	Expect(httpResp.StatusCode).To(Equal(200))

	releaseID := ""
	if current, ok := stack.GetConvergedReleaseOk(); ok {
		releaseID = current.GetId()
	} else if latest, ok := stack.GetLatestReleaseOk(); ok {
		releaseID = latest.GetId()
	} else {
		return nil, false
	}

	release, httpResp, err := apiClient.ReleasesApi.GetRelease(ctx, orgID, projectName, stackID, releaseID).Execute()
	Expect(err).NotTo(HaveOccurred())
	Expect(httpResp.StatusCode).To(Equal(200))

	liveStatus, ok := release.GetLiveStatusOk()
	if !ok {
		return nil, false
	}
	resources, ok := liveStatus.GetResourcesOk()
	if !ok {
		return nil, false
	}
	status, found := (*resources)[resourceName]
	if !found {
		return nil, false
	}
	return &status, true
}

// WaitForStackResourceFailed polls the API until the named resource reaches Failed state.
func WaitForStackResourceFailed(apiClient *openapi.APIClient, orgID, projectName, stackID, resourceName string, timeout time.Duration) *openapi.StackResourceStatus {
	var resource *openapi.StackResourceStatus
	Eventually(func(g Gomega) {
		status, ok := getStackResourceLiveStatus(apiClient, orgID, projectName, stackID, resourceName)
		g.Expect(ok).To(BeTrue(), "resource should have live status")

		state, stateOk := status.GetStateOk()
		g.Expect(stateOk).To(BeTrue(), "status should have state")
		g.Expect(*state).To(Equal(string(models.StackResourcePhaseFailed)), "resource should be Failed, got: %s", *state)
		resource = status
	}, timeout, 5*time.Second).Should(Succeed())
	return resource
}

// WaitForStackResourceLastFailure polls the API until last_failure is populated on the named resource.
func WaitForStackResourceLastFailure(apiClient *openapi.APIClient, orgID, projectName, stackID, resourceName string, timeout time.Duration) *openapi.StackResourceFailure {
	var lastFailure *openapi.StackResourceFailure
	Eventually(func(g Gomega) {
		status, ok := getStackResourceLiveStatus(apiClient, orgID, projectName, stackID, resourceName)
		g.Expect(ok).To(BeTrue(), "resource should have live status")

		failure, failureOk := status.GetLastFailureOk()
		g.Expect(failureOk).To(BeTrue(), "last_failure should be set on resource status")
		g.Expect(failure).NotTo(BeNil(), "last_failure should not be nil")
		lastFailure = failure
	}, timeout, 5*time.Second).Should(Succeed())
	return lastFailure
}

// WaitForBuildLastFailureDetail polls the API until last_build_failure_detail is populated on
// the named resource's most recent image build.
func WaitForBuildLastFailureDetail(apiClient *openapi.APIClient, orgID, projectName, stackID, resourceName string, timeout time.Duration) *openapi.BuildFailureDetail {
	var detail *openapi.BuildFailureDetail
	Eventually(func(g Gomega) {
		ctx := context.Background()
		resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameBuildsGet(ctx, orgID, projectName, stackID, resourceName).Execute()
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(httpResp.StatusCode).To(Equal(200))

		items := resp.GetItems()
		g.Expect(items).NotTo(BeEmpty(), "should have at least one image build")

		build := items[len(items)-1]
		status, ok := build.GetStatusOk()
		g.Expect(ok).To(BeTrue())

		d, dOk := status.GetLastBuildFailureDetailOk()
		g.Expect(dOk).To(BeTrue(), "last_build_failure_detail should be set")
		g.Expect(d).NotTo(BeNil())
		detail = d
	}, timeout, 5*time.Second).Should(Succeed())
	return detail
}

func ListStackBuilds(apiClient *openapi.APIClient, orgID, projectName, stackID string) []openapi.ImageBuild {
	ctx := context.Background()
	resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsGet(ctx, orgID, projectName, stackID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list stack builds")
	Expect(httpResp.StatusCode).To(Equal(200))
	Expect(resp).NotTo(BeNil())
	return resp.GetItems()
}

// DumpBuildSourceDebugInfo prints cluster and API state to help diagnose stuck builds.
func DumpBuildSourceDebugInfo(ctx context.Context, apiClient *openapi.APIClient, clusterClient client.Client, clientset *kubernetes.Clientset, orgID, projectName, stackID, namespace string) {
	fmt.Println("\n========== BUILD SOURCE DEBUG INFO ==========")

	// Stack status via API
	resp, _, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet(ctx, orgID, projectName, stackID).Execute()
	if err != nil {
		fmt.Printf("[Stack API] error fetching stack: %v\n", err)
	} else {
		fmt.Printf("[Stack API] lifecycle=%s\n", resp.GetLifecycle())
		if latest, ok := resp.GetLatestReleaseOk(); ok {
			fmt.Printf("[Stack API] latest_release id=%s state=%s health=%s message=%q\n",
				latest.GetId(), latest.GetState(), latest.GetHealth(), latest.GetMessage())
			release, _, releaseErr := apiClient.ReleasesApi.GetRelease(ctx, orgID, projectName, stackID, latest.GetId()).Execute()
			if releaseErr != nil {
				fmt.Printf("[Release API] error fetching release %s: %v\n", latest.GetId(), releaseErr)
			} else if liveStatus, liveOk := release.GetLiveStatusOk(); liveOk {
				for _, c := range liveStatus.GetConditions() {
					fmt.Printf("[Release API]   condition: type=%s status=%s reason=%s\n", c.GetType(), c.GetStatus(), c.GetReason())
				}
			}
		}
	}

	// Image builds via API
	buildsResp, _, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsGet(ctx, orgID, projectName, stackID).Execute()
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
