package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/testutil"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Environment struct {
	Client    *openapi.APIClient
	ClusterID string
	OrgID     string
	UserToken string
	// RegistryName is the suite org's signup-seeded registry,
	// <org-slug>-<shortOrgID>: the suffix comes from the org UUID minted at
	// signup, so the name differs every run and is read back from the API
	// after signup instead of being a test constant.
	RegistryName string
	Database     *DatabaseManager
	Cluster      *testutil.TestCluster

	// Internal managers
	clusterManager *ClusterManager
	serverManager  *ServerManager
	clientManager  *ClientManager
	logger         logr.Logger
}

// logger
func (env *Environment) Logger() logr.Logger {
	return env.logger
}

func Setup(env *Environment, ctx context.Context) (retErr error) {
	// Configure logger to output to stdout for visibility during test execution
	logger := stdr.New(log.New(os.Stderr, "", log.LstdFlags))
	logger.Info("Starting integration test bootstrap")

	env.logger = logger

	// Clean up all resources acquired so far if setup fails for any reason.
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Errorf("panic during bootstrap: %v", r), "Bootstrap failed with panic, attempting cleanup")
			env.Cleanup()
			panic(r)
		}
		if retErr != nil {
			logger.Error(retErr, "Bootstrap failed, cleaning up")
			env.Cleanup()
		}
	}()

	// Initialize database
	logger.Info("Bootstrapping database")
	dbManager := NewDatabaseManager()
	env.Database = dbManager
	if err := dbManager.Bootstrap(ctx); err != nil {
		return fmt.Errorf("database bootstrap failed: %w", err)
	}

	// Initialize cluster
	clusterManager := NewClusterManager()
	env.clusterManager = clusterManager
	if err := clusterManager.Bootstrap(ctx); err != nil {
		return fmt.Errorf("cluster bootstrap failed: %w", err)
	}
	env.Cluster = clusterManager.GetCluster()

	// Deploy MinIO for backup tests
	logger.Info("Deploying MinIO to cluster")
	if err := deployMinIO(ctx, env.Cluster.GetClient()); err != nil {
		return fmt.Errorf("MinIO deployment failed: %w", err)
	}

	// Export shared-compute and platform-routing environment values so the
	// server creates the platform org, shared-compute cluster, and optional
	// wildcard TLS resources at boot.
	logger.Info("Exporting shared-compute provisioning environment")
	if err := exportSharedComputeProvisioningEnv(ctx, env.Cluster); err != nil {
		return fmt.Errorf("failed to export shared-compute provisioning env: %w", err)
	}

	// Initialize server
	logger.Info("Bootstrapping server")
	serverManager := NewServerManager(dbManager.GetSessionFactory(), dbManager.GetConfig(), logger)
	env.serverManager = serverManager
	if err := serverManager.Bootstrap(ctx); err != nil {
		return fmt.Errorf("server bootstrap failed: %w", err)
	}

	// The booted server created the infrastructure-only platform org and its
	// shared-compute cluster. Resolve the shared-compute cluster from the DB
	// (tenants never see it via API),
	// then sign up the suite's tenant org, which inherits it at read time.
	var sharedComputeCluster models.Cluster
	if err := dbManager.GetSessionFactory().New(ctx).
		Where("shared_compute = ?", true).First(&sharedComputeCluster).Error; err != nil {
		return fmt.Errorf("failed to resolve shared-compute cluster: %w", err)
	}

	logger.Info("Bootstrapping client against platform defaults")
	clientManager := NewClientManager(serverManager.GetBaseURL(), logger)
	env.clientManager = clientManager
	if err := clientManager.Bootstrap(ctx, sharedComputeCluster.ID, dbManager.GetSessionFactory()); err != nil {
		return fmt.Errorf("client bootstrap failed: %w", err)
	}

	// Set final client details
	env.Client = clientManager.GetClient()
	env.ClusterID = clientManager.GetClusterID()
	env.OrgID = clientManager.GetOrgID()
	env.UserToken = clientManager.GetUserToken()
	env.RegistryName = clientManager.GetRegistryName()

	logger.Info("Integration test bootstrap completed successfully",
		"orgID", env.OrgID,
		"clusterID", env.ClusterID,
		"serverURL", serverManager.GetBaseURL())

	return nil
}

func (env *Environment) Cleanup() {
	if env.logger.GetSink() == nil {
		// Logger not initialized yet, create a basic one with stdout output
		env.logger = stdr.New(log.New(os.Stdout, "", log.LstdFlags))
	}

	env.logger.Info("Starting integration test cleanup")

	// Create context with timeout for cleanup operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Only cleanup if KEEP_CLUSTER is not set
	if os.Getenv("KEEP_CLUSTER") != "true" {
		// Cleanup in reverse order of setup with error recovery
		if env.serverManager != nil {
			env.logger.Info("Cleaning up server")
			func() {
				defer func() {
					if r := recover(); r != nil {
						env.logger.Error(fmt.Errorf("panic during server cleanup: %v", r), "Server cleanup panicked")
					}
				}()
				if err := env.serverManager.Cleanup(ctx); err != nil {
					env.logger.Error(err, "Failed to cleanup server")
				}
			}()
		}

		if env.clusterManager != nil {
			env.logger.Info("Cleaning up cluster")
			func() {
				defer func() {
					if r := recover(); r != nil {
						env.logger.Error(fmt.Errorf("panic during cluster cleanup: %v", r), "Cluster cleanup panicked")
					}
				}()
				if err := env.clusterManager.Cleanup(ctx); err != nil {
					env.logger.Error(err, "Failed to cleanup cluster")
				}
			}()
		}

		if env.Database != nil {
			env.logger.Info("Cleaning up database")
			func() {
				defer func() {
					if r := recover(); r != nil {
						env.logger.Error(fmt.Errorf("panic during database cleanup: %v", r), "Database cleanup panicked")
					}
				}()
				if err := env.Database.Cleanup(ctx); err != nil {
					env.logger.Error(err, "Failed to cleanup database")
				}
			}()
		}
	} else {
		env.logger.Info("Skipping cleanup due to KEEP_CLUSTER=true")
		env.logger.Info("Manual cleanup required:",
			"cluster", env.clusterManager != nil && env.clusterManager.GetCluster() != nil,
			"database", env.Database != nil)
	}

	env.logger.Info("Integration test cleanup completed")
}

func deployMinIO(ctx context.Context, k8sClient client.Client) error {
	labels := map[string]string{"app": testutil.MinIOName}
	replicas := int32(1)

	// Ensure namespace exists
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: testutil.MinIONamespace},
	}
	if err := k8sClient.Create(ctx, ns); err != nil && !k8sapierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testutil.MinIOName,
			Namespace: testutil.MinIONamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  testutil.MinIOName,
							Image: testutil.MinIOImage,
							Args:  []string{"server", "/data"},
							Ports: []corev1.ContainerPort{
								{ContainerPort: int32(testutil.MinIOServicePort), Protocol: corev1.ProtocolTCP},
							},
							Env: []corev1.EnvVar{
								{Name: "MINIO_ROOT_USER", Value: testutil.MinIOAccessKey},
								{Name: "MINIO_ROOT_PASSWORD", Value: testutil.MinIOSecretKey},
							},
						},
					},
				},
			},
		},
	}

	existing := &appsv1.Deployment{}
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), existing); err != nil {
		if !k8sapierrors.IsNotFound(err) {
			return fmt.Errorf("failed to check existing deployment: %w", err)
		}
		if err := k8sClient.Create(ctx, deployment); err != nil {
			return fmt.Errorf("failed to create MinIO deployment: %w", err)
		}
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testutil.MinIOName,
			Namespace: testutil.MinIONamespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       int32(testutil.MinIOServicePort),
					TargetPort: intstr.FromInt(testutil.MinIOServicePort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	existingSvc := &corev1.Service{}
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(svc), existingSvc); err != nil {
		if !k8sapierrors.IsNotFound(err) {
			return fmt.Errorf("failed to check existing service: %w", err)
		}
		if err := k8sClient.Create(ctx, svc); err != nil {
			return fmt.Errorf("failed to create MinIO service: %w", err)
		}
	}

	// Wait for MinIO to be ready
	waitCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	for {
		var dep appsv1.Deployment
		if err := k8sClient.Get(waitCtx, client.ObjectKey{Name: testutil.MinIOName, Namespace: testutil.MinIONamespace}, &dep); err == nil {
			if dep.Status.ReadyReplicas >= 1 {
				break
			}
		}
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("timed out waiting for MinIO to become ready")
		case <-time.After(5 * time.Second):
		}
	}

	// Create the backup bucket using mc (MinIO Client) via a Job
	return createMinioBucket(ctx, k8sClient)
}

func createMinioBucket(ctx context.Context, k8sClient client.Client) error {
	minioEndpoint := testutil.MinIOEndpoint()
	backoffLimit := int32(3)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "minio-create-bucket",
			Namespace: testutil.MinIONamespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:  "mc",
							Image: testutil.MinIOClientImage,
							Command: []string{"/bin/sh", "-c",
								fmt.Sprintf("mc alias set local %s %s %s && mc mb --ignore-existing local/%s",
									minioEndpoint, testutil.MinIOAccessKey, testutil.MinIOSecretKey, testutil.MinIOBucket),
							},
						},
					},
				},
			},
		},
	}

	// Delete old job if exists
	existing := &batchv1.Job{}
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(job), existing); err == nil {
		propagation := metav1.DeletePropagationBackground
		_ = k8sClient.Delete(ctx, existing, &client.DeleteOptions{
			PropagationPolicy: &propagation,
		})
		time.Sleep(5 * time.Second)
	}

	if err := k8sClient.Create(ctx, job); err != nil {
		return fmt.Errorf("failed to create bucket creation job: %w", err)
	}

	// Wait for job to complete
	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	for {
		var j batchv1.Job
		if err := k8sClient.Get(waitCtx, client.ObjectKeyFromObject(job), &j); err == nil {
			if j.Status.Succeeded >= 1 {
				return nil
			}
			if j.Status.Failed >= backoffLimit {
				return fmt.Errorf("bucket creation job failed")
			}
		}
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("timed out waiting for bucket creation")
		case <-time.After(5 * time.Second):
		}
	}
}
