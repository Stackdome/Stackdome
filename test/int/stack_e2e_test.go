package int

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/test/int/shared"
)

var _ = Describe("Stack E2E", Ordered, func() {
	var client *openapi.APIClient
	var orgID string
	projectName := models.DefaultProjectName

	BeforeAll(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil(), "Test environment should be initialized")

		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	Context("Basic Lifecycle", func() {
		It("should create a Stack CR in the cluster and reach Ready", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a stack via API and deploying")
			stack := shared.CreateSimpleStack("test-lifecycle")
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)

			stackID := created.GetId()
			stackName := created.GetName()
			namespace := created.GetNamespace()

			Expect(stackID).NotTo(BeEmpty())
			Expect(namespace).NotTo(BeEmpty())

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
			})

			By("Waiting for the Stack CR to appear in the cluster")
			cr := shared.WaitForStackCRExists(ctx, clusterClient, stackName, namespace, 2*time.Minute)

			By("Verifying CR has the stack ID label")
			shared.VerifyStackCRLabel(cr, stackID)

			By("Waiting for stack to become Ready via API")
			readyStack := shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)

			By("Verifying current release is healthy with conditions")
			convergedRelease, ok := readyStack.GetConvergedReleaseOk()
			Expect(ok).To(BeTrue())
			releaseDetail := shared.GetRelease(client, orgID, projectName, stackID, convergedRelease.GetId())
			liveStatus, ok := releaseDetail.GetLiveStatusOk()
			Expect(ok).To(BeTrue())
			Expect(liveStatus.GetConditions()).NotTo(BeEmpty())

			By("Verifying StackResource CR is Available in the cluster")
			shared.WaitForStackResourceCRAvailable(ctx, clusterClient, "web", namespace, 5*time.Minute)

			By("Verifying Deployment was created for the resource")
			deploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, "web")
			Expect(err).NotTo(HaveOccurred())
			Expect(deploy.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(deploy.Spec.Template.Spec.Containers[0].Image).To(Equal(shared.TestImage))

			By("Verifying Service was created for the resource")
			svc, err := shared.GetServiceForStackResource(ctx, clusterClient, namespace, "web")
			Expect(err).NotTo(HaveOccurred())
			Expect(svc.Spec.Ports).To(HaveLen(1))
		})

		It("should delete the Stack CR when stack is deleted via API", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a stack and deploying")
			stack := shared.CreateSimpleStack("test-delete-e2e")
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()
			stackName := created.GetName()
			namespace := created.GetNamespace()

			By("Waiting for stack to become Ready")
			shared.WaitForStackCRExists(ctx, clusterClient, stackName, namespace, 2*time.Minute)
			shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)

			By("Deleting the stack via API")
			shared.DeleteStack(client, orgID, projectName, stackID)

			By("Verifying the CR is deleted from the cluster")
			shared.WaitForStackCRDeleted(ctx, clusterClient, stackName, namespace, 2*time.Minute)

			By("Verifying the stack is gone from the API")
			shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
		})
	})

	Context("Multi-Resource Stack", func() {
		It("should deploy multiple resources and resolve connection env vars", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a multi-resource stack and deploying")
			stack := shared.CreateMultiResourceStack("test-multi")
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)

			By("Verifying both StackResource CRs are Available")
			shared.WaitForStackResourceCRAvailable(ctx, clusterClient, shared.MultiResourceBackendName, namespace, 5*time.Minute)
			shared.WaitForStackResourceCRAvailable(ctx, clusterClient, shared.MultiResourceFrontendName, namespace, 5*time.Minute)

			By("Verifying backend Deployment exists")
			backendDeploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, shared.MultiResourceBackendName)
			Expect(err).NotTo(HaveOccurred())
			Expect(backendDeploy).NotTo(BeNil())

			By("Verifying frontend Deployment has connection-resolved BACKEND_URL env var")
			frontendDeploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, shared.MultiResourceFrontendName)
			Expect(err).NotTo(HaveOccurred())

			backendURL, found := shared.GetContainerEnvVar(frontendDeploy, "BACKEND_URL")
			Expect(found).To(BeTrue(), "BACKEND_URL env var should exist")
			Expect(backendURL).To(HavePrefix(shared.MultiResourceBackendName+"."), "BACKEND_URL should be the backend K8s internal service host")
			Expect(backendURL).To(HaveSuffix(".svc"), "BACKEND_URL should be a cluster-local service FQDN")
		})
	})

	Context("Dependencies", func() {
		It("should respect depends_on ordering", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a stack with dependencies and deploying")
			stack := shared.CreateStackWithDependencies("test-deps")
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)

			By("Verifying both resources are Available")
			shared.WaitForStackResourceCRAvailable(ctx, clusterClient, "database", namespace, 5*time.Minute)
			shared.WaitForStackResourceCRAvailable(ctx, clusterClient, "app", namespace, 5*time.Minute)

			By("Verifying both Deployments exist")
			_, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, "database")
			Expect(err).NotTo(HaveOccurred())
			_, err = shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, "app")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Env Vars and Ports", func() {
		It("should create resources with multiple ports and env vars", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a stack with env vars and ports and deploying")
			stack := shared.CreateStackWithEnvAndPorts("test-env-ports")
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)

			By("Verifying Deployment has the expected env vars")
			deploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, shared.EnvPortsResourceName)
			Expect(err).NotTo(HaveOccurred())

			appEnv, found := shared.GetContainerEnvVar(deploy, shared.EnvPortsAppEnvKey)
			Expect(found).To(BeTrue())
			Expect(appEnv).To(Equal(shared.EnvPortsAppEnvVal))

			appPort, found := shared.GetContainerEnvVar(deploy, shared.EnvPortsAppPortKey)
			Expect(found).To(BeTrue())
			Expect(appPort).To(Equal(shared.EnvPortsAppPortVal))

			logLevel, found := shared.GetContainerEnvVar(deploy, shared.EnvPortsLogLevelKey)
			Expect(found).To(BeTrue())
			Expect(logLevel).To(Equal(shared.EnvPortsLogLevelVal))

			By("Verifying Service has both ports")
			svc, err := shared.GetServiceForStackResource(ctx, clusterClient, namespace, shared.EnvPortsResourceName)
			Expect(err).NotTo(HaveOccurred())
			Expect(svc.Spec.Ports).To(HaveLen(2))

			portNumbers := []int32{}
			for _, p := range svc.Spec.Ports {
				portNumbers = append(portNumbers, p.Port)
			}
			Expect(portNumbers).To(ContainElements(int32(shared.EnvPortsPort1), int32(shared.EnvPortsPort2)))
		})
	})

	Context("Init Container", func() {
		// Skipped: cluster-agent bug — init container uses main image instead of InitSpec.ImageSpec.
		// See docs/plans/cluster-agent-fixes.md
		PIt("should deploy a resource with an init container", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a stack with init container and deploying")
			stack := shared.CreateStackWithInitContainer("test-init")
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)

			By("Verifying Deployment has init container")
			deploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, "app")
			Expect(err).NotTo(HaveOccurred())
			Expect(deploy.Spec.Template.Spec.InitContainers).To(HaveLen(1))
			Expect(deploy.Spec.Template.Spec.InitContainers[0].Image).To(Equal(shared.InitImage))
			Expect(deploy.Spec.Template.Spec.InitContainers[0].Command).To(Equal([]string{"sh", "-c", shared.InitCommand}))
		})
	})

	Context("Update Propagation", func() {
		It("should propagate spec updates to the cluster CR", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a stack and deploying")
			stack := shared.CreateSimpleStack("test-update-e2e")
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()
			stackName := created.GetName()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackCRExists(ctx, clusterClient, stackName, namespace, 2*time.Minute)
			shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)

			By("Updating the stack with a new env var")
			updateStack := shared.CreateSimpleStack("test-update-e2e")
			exec := openapi.NewExecutionConfig()
			exec.SetEnvironmentVariables([]openapi.EnvVar{
				func() openapi.EnvVar { v := openapi.NewEnvVar("UPDATED"); v.SetValue("true"); return *v }(),
			})
			updateStack.Spec.StackResources[0].SetExecutionConfig(*exec)
			shared.UpdateStack(client, orgID, projectName, stackID, updateStack)

			By("Creating a release for the update")
			shared.CreateRelease(client, orgID, projectName, stackID)

			By("Waiting for the StackResource CR to reflect the update")
			Eventually(func(g Gomega) {
				srCR, err := shared.GetStackResourceCR(ctx, clusterClient, "web", namespace)
				g.Expect(err).NotTo(HaveOccurred())

				foundEnv := false
				for _, env := range srCR.Spec.EnvironmentVariables {
					if env.Name == "UPDATED" && env.Value == "true" {
						foundEnv = true
						break
					}
				}
				g.Expect(foundEnv).To(BeTrue(), "StackResource CR should have UPDATED=true env var")
			}, 2*time.Minute, 5*time.Second).Should(Succeed())

			By("Waiting for stack to return to Ready state")
			shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)
		})
	})

	Context("Stack with PostgresAddon", func() {
		It("should inject postgres addon env vars into stack resource", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a postgres addon with a database")
			addon := shared.CreatePostgresAddonWithResources("test-stack-pg")
			createdAddon := shared.CreatePostgresAddon(client, orgID, projectName, addon)
			addonID := createdAddon.GetId()

			DeferCleanup(func() {
				shared.DeletePostgresAddon(client, orgID, projectName, addonID)
			})

			By("Waiting for the postgres addon to become Ready")
			shared.WaitForAddonReady(client, orgID, projectName, addonID, 10*time.Minute)

			By("Creating a stack that references the postgres addon and deploying")
			stack := shared.CreateStackWithPostgresAddon("test-pg-stack", addonID, "testdb")
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)

			By("Verifying Deployment has all postgres env vars from the mapping")
			deploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, "app")
			Expect(err).NotTo(HaveOccurred())

			for credField, envName := range shared.PostgresEnvMapping {
				val, found := shared.ResolveContainerEnvVar(ctx, clusterClient, deploy, namespace, envName)
				Expect(found).To(BeTrue(), "env var %s (mapped from %s) should be injected", envName, credField)
				Expect(val).NotTo(BeEmpty(), "env var %s should have a non-empty value", envName)
			}

			By("Verifying postgres credential values are structurally valid")
			pgHost, _ := shared.ResolveContainerEnvVar(ctx, clusterClient, deploy, namespace, shared.PostgresEnvMapping["host"])
			Expect(pgHost).To(ContainSubstring(".svc.cluster.local"), "PG_HOST should be a fully-qualified K8s service DNS name")

			pgPort, _ := shared.ResolveContainerEnvVar(ctx, clusterClient, deploy, namespace, shared.PostgresEnvMapping["port"])
			Expect(pgPort).To(Equal("5432"), "PG_PORT should be the default postgres port")

			pgUser, _ := shared.ResolveContainerEnvVar(ctx, clusterClient, deploy, namespace, shared.PostgresEnvMapping["username"])
			Expect(pgUser).NotTo(BeEmpty(), "PG_USER should be a valid database username")

			dbURL, _ := shared.ResolveContainerEnvVar(ctx, clusterClient, deploy, namespace, shared.PostgresEnvMapping["url"])
			Expect(dbURL).To(HavePrefix("postgresql://"), "DATABASE_URL should be a postgresql:// connection string")
			Expect(dbURL).To(ContainSubstring("testdb"), "DATABASE_URL should reference the requested database")
		})
	})

	Context("Stack with PostgresAddon Superuser", func() {
		It("should inject superuser credentials and allow privileged database operations", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a postgres addon with superuser access enabled")
			addon := shared.CreatePostgresAddonWithSuperuser("test-stack-pg-su")
			createdAddon := shared.CreatePostgresAddon(client, orgID, projectName, addon)
			addonID := createdAddon.GetId()
			addonName := createdAddon.GetName()
			addonNamespace := createdAddon.GetNamespace()

			DeferCleanup(func() {
				shared.DeletePostgresAddon(client, orgID, projectName, addonID)
			})

			By("Waiting for the postgres addon to become Ready")
			shared.WaitForAddonReady(client, orgID, projectName, addonID, 10*time.Minute)

			By("Waiting for databases to be applied")
			shared.WaitForConditionTrue(client, orgID, projectName, addonID, string(models.PostgresAddonConditionDatabasesApplied), 2*time.Minute)

			By("Creating a stack that references the addon with superuser mode and deploying")
			stack := shared.CreateStackWithPostgresAddonSuperuser("test-su-stack", addonID)
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)

			By("Verifying Deployment has all postgres env vars from the mapping")
			deploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, "app")
			Expect(err).NotTo(HaveOccurred())

			for credField, envName := range shared.PostgresEnvMapping {
				val, found := shared.ResolveContainerEnvVar(ctx, clusterClient, deploy, namespace, envName)
				Expect(found).To(BeTrue(), "env var %s (mapped from %s) should be injected", envName, credField)
				Expect(val).NotTo(BeEmpty(), "env var %s should have a non-empty value", envName)
			}

			By("Verifying superuser credential values on deployment")
			pgUser, _ := shared.ResolveContainerEnvVar(ctx, clusterClient, deploy, namespace, shared.PostgresEnvMapping["username"])
			Expect(pgUser).To(Equal("postgres"), "superuser username should be 'postgres'")

			dbURL, _ := shared.ResolveContainerEnvVar(ctx, clusterClient, deploy, namespace, shared.PostgresEnvMapping["url"])
			Expect(dbURL).To(HavePrefix("postgresql://"), "DATABASE_URL should be a postgresql:// connection string")

			By("Fetching superuser credentials via API")
			creds := shared.GetSuperuserCredentials(client, orgID, projectName, addonID)
			Expect(creds.GetUsername()).To(Equal("postgres"), "superuser username should be 'postgres'")
			Expect(creds.GetPassword()).NotTo(BeEmpty())

			By("Port-forwarding to the primary postgres pod")
			clientset, err := testEnv.Cluster.GetKubeClient()
			Expect(err).NotTo(HaveOccurred())

			cnpgName := shared.CnpgClusterName(addonName, int(addon.Spec.Version.Major))
			localPort, stopChan := shared.PortForwardPostgres(ctx, testEnv.Cluster.GetRESTConfig(), clientset, addonNamespace, cnpgName)
			defer close(stopChan)

			By("Verifying superuser can read and write to testdb")
			testDB := shared.ConnectToPostgres("127.0.0.1", localPort, creds.GetUsername(), creds.GetPassword(), "testdb", "disable")
			defer func() { _ = testDB.Close() }()

			_, err = testDB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS e2e_su_test (id serial PRIMARY KEY, val text)")
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to create tables in testdb")

			_, err = testDB.ExecContext(ctx, "INSERT INTO e2e_su_test (val) VALUES ('superuser_write')")
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to insert into testdb")

			var val string
			err = testDB.QueryRowContext(ctx, "SELECT val FROM e2e_su_test WHERE val = 'superuser_write'").Scan(&val)
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to read from testdb")
			Expect(val).To(Equal("superuser_write"))

			By("Verifying superuser can read and write to the default app database")
			appDB := shared.ConnectToPostgres("127.0.0.1", localPort, creds.GetUsername(), creds.GetPassword(), "app", "disable")
			defer func() { _ = appDB.Close() }()

			_, err = appDB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS e2e_su_test (id serial PRIMARY KEY, val text)")
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to create tables in app database")

			_, err = appDB.ExecContext(ctx, "INSERT INTO e2e_su_test (val) VALUES ('superuser_write')")
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to insert into app database")

			err = appDB.QueryRowContext(ctx, "SELECT val FROM e2e_su_test WHERE val = 'superuser_write'").Scan(&val)
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to read from app database")
			Expect(val).To(Equal("superuser_write"))

			By("Verifying superuser can create a new database")
			postgresDB := shared.ConnectToPostgres("127.0.0.1", localPort, creds.GetUsername(), creds.GetPassword(), "postgres", "disable")
			defer func() { _ = postgresDB.Close() }()

			_, err = postgresDB.ExecContext(ctx, "CREATE DATABASE e2e_superuser_created")
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to create new databases")

			newDB := shared.ConnectToPostgres("127.0.0.1", localPort, creds.GetUsername(), creds.GetPassword(), "e2e_superuser_created", "disable")
			defer func() { _ = newDB.Close() }()

			var result int
			err = newDB.QueryRowContext(ctx, "SELECT 1").Scan(&result)
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to connect to the newly created database")
			Expect(result).To(Equal(1))
		})
	})

	Context("Full Stack with all connection types", func() {
		It("should deploy a stack with postgres, secret, volume mount, resource-to-resource, and depends_on", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a postgres addon with a database")
			addon := shared.CreatePostgresAddonWithResources("test-full-stack-pg")
			createdAddon := shared.CreatePostgresAddon(client, orgID, projectName, addon)
			addonID := createdAddon.GetId()

			DeferCleanup(func() {
				shared.DeletePostgresAddon(client, orgID, projectName, addonID)
			})

			By("Waiting for the postgres addon to become Ready")
			shared.WaitForAddonReady(client, orgID, projectName, addonID, 10*time.Minute)

			By("Creating a secret with TLS credentials")
			secret := shared.CreateGenericSecret(shared.FullStackSecretName, map[string]string{
				"tls_cert": "-----BEGIN CERTIFICATE-----\ntest-cert-data\n-----END CERTIFICATE-----",
				"tls_key":  "-----BEGIN PRIVATE KEY-----\ntest-key-data\n-----END PRIVATE KEY-----",
			})
			createdSecret := shared.CreateSecret(client, orgID, projectName, secret)
			secretID := createdSecret.GetId()

			DeferCleanup(func() {
				shared.DeleteSecret(client, orgID, projectName, secretID)
			})

			By("Creating the full stack and deploying")
			stack := shared.CreateFullStack("test-full", addonID, "testdb", secretID)
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)

			By("Verifying both StackResource CRs are Available")
			shared.WaitForStackResourceCRAvailable(ctx, clusterClient, shared.FullStackAPIName, namespace, 5*time.Minute)
			shared.WaitForStackResourceCRAvailable(ctx, clusterClient, shared.FullStackWorkerName, namespace, 5*time.Minute)

			By("Verifying API deployment has postgres env vars")
			apiDeploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, shared.FullStackAPIName)
			Expect(err).NotTo(HaveOccurred())

			for credField, envName := range shared.PostgresEnvMapping {
				val, found := shared.ResolveContainerEnvVar(ctx, clusterClient, apiDeploy, namespace, envName)
				Expect(found).To(BeTrue(), "api: env var %s (mapped from %s) should be injected", envName, credField)
				Expect(val).NotTo(BeEmpty(), "api: env var %s should have a non-empty value", envName)
			}

			By("Verifying API deployment has secret env vars")
			for _, envName := range shared.FullStackSecretEnvMapping {
				val, found := shared.ResolveContainerEnvVar(ctx, clusterClient, apiDeploy, namespace, envName)
				Expect(found).To(BeTrue(), "api: env var %s should be injected from secret", envName)
				Expect(val).NotTo(BeEmpty(), "api: env var %s should have a non-empty value", envName)
			}

			By("Verifying API deployment has volume mount")
			vm, found := shared.GetContainerVolumeMount(apiDeploy, shared.FullStackMountPath)
			Expect(found).To(BeTrue(), "api: volume mount at %s should exist", shared.FullStackMountPath)
			Expect(vm.SubPath).To(Equal(shared.FullStackSubPath), "api: volume mount sub_path should match")

			By("Verifying worker deployment has postgres env vars")
			workerDeploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, shared.FullStackWorkerName)
			Expect(err).NotTo(HaveOccurred())

			dbURL, found := shared.ResolveContainerEnvVar(ctx, clusterClient, workerDeploy, namespace, shared.PostgresEnvMapping["url"])
			Expect(found).To(BeTrue(), "worker: DATABASE_URL should be injected")
			Expect(dbURL).To(HavePrefix("postgresql://"), "worker: DATABASE_URL should be a postgresql:// connection string")

			By("Verifying worker deployment has resource-to-resource env vars")
			apiHost, found := shared.GetContainerEnvVar(workerDeploy, "API_HOST")
			Expect(found).To(BeTrue(), "worker: API_HOST should be injected from api resource connection")
			Expect(apiHost).To(HaveSuffix(".svc"), "worker: API_HOST should be a cluster-local service FQDN")

			apiURL, found := shared.GetContainerEnvVar(workerDeploy, "API_URL")
			Expect(found).To(BeTrue(), "worker: API_URL should be injected from api resource connection")
			Expect(apiURL).To(ContainSubstring(fmt.Sprintf(":%d", shared.FullStackAPIPort)), "worker: API_URL should include api port")

			By("Verifying topology API returns correct graph")
			topology := shared.GetStackTopology(client, orgID, projectName, stackID)
			Expect(topology.Nodes).To(HaveLen(5), "topology should have 5 nodes: api, worker, volume, addon, secret")
			Expect(len(topology.Edges)).To(BeNumerically(">=", 6), "topology should have at least 6 edges: 5 connections + 1 depends_on")

			By("Verifying connections API returns all 5 connections")
			connections := shared.ListStackConnections(client, orgID, projectName, stackID)
			Expect(connections).To(HaveLen(5), "should have 5 persisted connections")
		})
	})

	Context("Crash Detection", func() {
		It("should populate last_failure when a container crashes", func() {
			By("Creating a stack whose container immediately exits with error and deploying")
			stack := shared.CreateCrashingStack("test-crash-detection")
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
			})

			By("Waiting for last_failure to be populated on the resource")
			// Crashing containers stay in Pending phase (not Failed) — we poll for last_failure directly.
			// Allow enough time for: image validation + deployment creation + crash + cluster-agent capture + DB sync.
			lastFailure := shared.WaitForStackResourceLastFailure(client, orgID, projectName, stackID, shared.CrashResourceName, 5*time.Minute)

			By("Verifying last_failure has the correct type and container details")
			Expect(lastFailure.Type).NotTo(BeNil())
			Expect(*lastFailure.Type).To(Equal("runtime_crash"))

			Expect(lastFailure.Container).NotTo(BeNil(), "container failure detail should be set")
			Expect(lastFailure.Container.FailureType).NotTo(BeNil())
			Expect(*lastFailure.Container.FailureType).To(BeElementOf("exit_error", "crash_loop"),
				"failure_type should be exit_error or crash_loop depending on restart count")

			Expect(lastFailure.Container.Message).NotTo(BeNil(), "message should contain container logs before exit")
			Expect(*lastFailure.Container.Message).To(ContainSubstring(shared.CrashMessage),
				"message should contain the output printed before the container exited")

			Expect(lastFailure.Build).To(BeNil(), "build failure should not be set for a runtime crash")
		})
	})

	Context("Build from Source with Private Repo", func() {
		var githubToken string

		BeforeAll(func() {
			githubToken = os.Getenv("GITHUB_TOKEN")
			if githubToken == "" {
				Skip("GITHUB_TOKEN not set — skipping build-from-source tests")
			}
			// For now print first 20 characters to verify it's being picked up without exposing the full token in logs
			testenv := GetEnvironment()
			testenv.Logger().Info("GITHUB_TOKEN set", "token", githubToken[:20]+"...")

			// Register an org git integration for the repo host so private clones
			// authenticate via host auto-match (inline source credentials are gone).
			integration := shared.CreateGitCredentialsIntegration(client, orgID, shared.BuildSourceGitHost, shared.BuildSourceGitUsername, githubToken)
			DeferCleanup(func() {
				shared.DeleteGitIntegration(client, orgID, integration.GetId())
			})
		})

		It("should build from a private git repo and expose to public", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			clientset, err := testEnv.Cluster.GetKubeClient()
			Expect(err).NotTo(HaveOccurred())
			ctx := context.Background()

			By("Creating a stack with build source and exposed port, then deploying")
			stack := shared.CreateStackWithBuildSource("test-build-source", shared.BuildSourceRepoURL)
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()
			stackName := created.GetName()
			namespace := created.GetNamespace()

			Expect(stackID).NotTo(BeEmpty())
			Expect(namespace).NotTo(BeEmpty())

			DeferCleanup(func() {
				if CurrentSpecReport().Failed() {
					shared.DumpBuildSourceDebugInfo(ctx, client, clusterClient, clientset, orgID, projectName, stackID, namespace)
				}
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 2*time.Minute)
			})

			By("Waiting for the Stack CR to appear in the cluster")
			cr := shared.WaitForStackCRExists(ctx, clusterClient, stackName, namespace, 2*time.Minute)

			By("Verifying Stack CR has the resource name listed")
			Expect(cr.Spec.ResourceNames).To(ContainElement(shared.BuildSourceResourceName))

			By("Verifying StackResource CR has build spec with git repo URL")
			srCR, err := shared.GetStackResourceCR(ctx, clusterClient, shared.BuildSourceResourceName, namespace)
			Expect(err).NotTo(HaveOccurred())
			buildSpec := srCR.Spec.BuildSpec
			Expect(buildSpec).NotTo(BeNil(), "StackResource CR should have a BuildSpec")
			Expect(buildSpec.SourceContext.Git).NotTo(BeNil(), "BuildSpec should have a Git source context")
			Expect(buildSpec.SourceContext.Git.RepoUrl).To(Equal(shared.BuildSourceRepoURL))

			By("Verifying git credentials secret exists in the cluster")
			shared.VerifyGitCredentialsSecretExists(ctx, clusterClient, namespace, stackID)

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, projectName, stackID, 10*time.Minute)

			By("Verifying StackResource CR is Available")
			shared.WaitForStackResourceCRAvailable(ctx, clusterClient, shared.BuildSourceResourceName, namespace, 5*time.Minute)

			By("Verifying Deployment uses a built image from in-cluster registry")
			deploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, shared.BuildSourceResourceName)
			Expect(err).NotTo(HaveOccurred())
			Expect(deploy.Spec.Template.Spec.Containers).To(HaveLen(1))
			containerImage := deploy.Spec.Template.Spec.Containers[0].Image
			Expect(containerImage).NotTo(BeEmpty(), "container image should be set")
			Expect(containerImage).To(ContainSubstring(shared.TestRegistryName), "image should be from the in-cluster registry")

			By("Verifying Service has port 3000")
			svc, err := shared.GetServiceForStackResource(ctx, clusterClient, namespace, shared.BuildSourceResourceName)
			Expect(err).NotTo(HaveOccurred())
			portFound := false
			for _, p := range svc.Spec.Ports {
				if p.Port == int32(shared.BuildSourcePort) {
					portFound = true
					break
				}
			}
			Expect(portFound).To(BeTrue(), "Service should have port %d", shared.BuildSourcePort)

			By("Verifying Ingress was created for exposed port")
			ingress, err := shared.GetIngressForStackResource(ctx, clusterClient, namespace, shared.BuildSourceResourceName)
			Expect(err).NotTo(HaveOccurred(), "Ingress should exist for exposed port")
			Expect(ingress.Spec.Rules).NotTo(BeEmpty(), "Ingress should have at least one rule")

			By("Port-forwarding to the app and verifying HTTP response")
			localPort, stopChan := shared.PortForwardStackResource(ctx, testEnv.Cluster.GetRESTConfig(), clientset, namespace, shared.BuildSourceResourceName, shared.BuildSourcePort)
			defer close(stopChan)

			httpClient := &http.Client{Timeout: 10 * time.Second}
			resp, err := httpClient.Get(fmt.Sprintf("http://127.0.0.1:%d/", localPort))
			Expect(err).NotTo(HaveOccurred(), "HTTP GET to app should succeed")
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK), "expected 200, got %d; body: %s", resp.StatusCode, string(body))

			By("Verifying image build record via API")
			builds := shared.ListStackBuilds(client, orgID, projectName, stackID)
			Expect(builds).NotTo(BeEmpty(), "should have at least one image build record")
		})

		It("should populate last_build_failure_detail when a build fails", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a stack with a build source pointing to a branch with a broken Dockerfile and deploying")
			stack := shared.CreateStackWithBrokenBuildSource("test-build-fail", shared.BuildSourceRepoURL)
			created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, stack)
			stackID := created.GetId()
			stackName := created.GetName()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				if CurrentSpecReport().Failed() {
					kubeClient, _ := testEnv.Cluster.GetKubeClient()
					shared.DumpBuildSourceDebugInfo(ctx, client, clusterClient, kubeClient, orgID, projectName, stackID, namespace)
				}
				shared.DeleteStack(client, orgID, projectName, stackID)
				shared.WaitForStackDeleted(client, orgID, projectName, stackID, 2*time.Minute)
			})

			By("Waiting for the Stack CR to appear in the cluster")
			shared.WaitForStackCRExists(ctx, clusterClient, stackName, namespace, 2*time.Minute)

			By("Waiting for last_build_failure_detail to be populated on the image build")
			detail := shared.WaitForBuildLastFailureDetail(client, orgID, projectName, stackID, shared.BrokenBuildResourceName, 5*time.Minute)

			By("Verifying last_build_failure_detail has failure info")
			Expect(detail.FailureType).NotTo(BeNil())
			Expect(*detail.FailureType).To(BeElementOf("exit_error", "image_pull_failed"))

			By("Verifying last_failure on the stack resource reflects the build failure")
			lastFailure := shared.WaitForStackResourceLastFailure(client, orgID, projectName, stackID, shared.BrokenBuildResourceName, 2*time.Minute)
			Expect(lastFailure.Type).NotTo(BeNil())
			Expect(*lastFailure.Type).To(Equal("build_failure"))
			Expect(lastFailure.Build).NotTo(BeNil())
		})
	})
})
