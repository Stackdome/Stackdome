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

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/test/int/shared"
)

var _ = Describe("Stack E2E", Ordered, func() {
	var client *openapi.APIClient
	var orgID string
	teamName := models.DefaultTeamName

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

			By("Creating a stack via API")
			stack := shared.CreateSimpleStack("test-lifecycle")
			created := shared.CreateStack(client, orgID, teamName, stack)

			stackID := created.GetId()
			stackName := created.GetName()
			namespace := created.GetNamespace()

			Expect(stackID).NotTo(BeEmpty())
			Expect(namespace).NotTo(BeEmpty())

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
			})

			By("Waiting for the Stack CR to appear in the cluster")
			cr := shared.WaitForStackCRExists(ctx, clusterClient, stackName, namespace, 2*time.Minute)

			By("Verifying CR has the stack ID label")
			shared.VerifyStackCRLabel(cr, stackID)

			By("Waiting for stack to become Ready via API")
			readyStack := shared.WaitForStackReady(client, orgID, teamName, stackID, 5*time.Minute)

			By("Verifying status has conditions")
			status, ok := readyStack.GetStatusOk()
			Expect(ok).To(BeTrue())
			Expect(status.GetConditions()).NotTo(BeEmpty())

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

			By("Creating a stack")
			stack := shared.CreateSimpleStack("test-delete-e2e")
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			stackName := created.GetName()
			namespace := created.GetNamespace()

			By("Waiting for stack to become Ready")
			shared.WaitForStackCRExists(ctx, clusterClient, stackName, namespace, 2*time.Minute)
			shared.WaitForStackReady(client, orgID, teamName, stackID, 5*time.Minute)

			By("Deleting the stack via API")
			shared.DeleteStack(client, orgID, teamName, stackID)

			By("Verifying the CR is deleted from the cluster")
			shared.WaitForStackCRDeleted(ctx, clusterClient, stackName, namespace, 2*time.Minute)

			By("Verifying the stack is gone from the API")
			shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
		})
	})

	Context("Multi-Resource Stack", func() {
		It("should deploy multiple resources and interpolate env vars", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a multi-resource stack")
			stack := shared.CreateMultiResourceStack("test-multi")
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, teamName, stackID, 5*time.Minute)

			By("Verifying both StackResource CRs are Available")
			shared.WaitForStackResourceCRAvailable(ctx, clusterClient, shared.MultiResourceBackendName, namespace, 5*time.Minute)
			shared.WaitForStackResourceCRAvailable(ctx, clusterClient, shared.MultiResourceFrontendName, namespace, 5*time.Minute)

			By("Verifying backend Deployment exists")
			backendDeploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, shared.MultiResourceBackendName)
			Expect(err).NotTo(HaveOccurred())
			Expect(backendDeploy).NotTo(BeNil())

			By("Verifying frontend Deployment has interpolated BACKEND_URL env var")
			frontendDeploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, shared.MultiResourceFrontendName)
			Expect(err).NotTo(HaveOccurred())

			backendURL, found := shared.GetContainerEnvVar(frontendDeploy, "BACKEND_URL")
			Expect(found).To(BeTrue(), "BACKEND_URL env var should exist")
			Expect(backendURL).To(Equal(shared.MultiResourceBackendName), "BACKEND_URL should be the backend K8s service name")
		})
	})

	Context("Dependencies", func() {
		It("should respect depends_on ordering", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a stack with dependencies")
			stack := shared.CreateStackWithDependencies("test-deps")
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, teamName, stackID, 5*time.Minute)

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

			By("Creating a stack with env vars and ports")
			stack := shared.CreateStackWithEnvAndPorts("test-env-ports")
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, teamName, stackID, 5*time.Minute)

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

			By("Creating a stack with init container")
			stack := shared.CreateStackWithInitContainer("test-init")
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, teamName, stackID, 5*time.Minute)

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

			By("Creating a stack")
			stack := shared.CreateSimpleStack("test-update-e2e")
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			stackName := created.GetName()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackCRExists(ctx, clusterClient, stackName, namespace, 2*time.Minute)
			shared.WaitForStackReady(client, orgID, teamName, stackID, 5*time.Minute)

			By("Updating the stack with a new env var")
			updateStack := shared.CreateSimpleStack("test-update-e2e")
			exec := openapi.NewExecutionConfig()
			exec.SetEnvironmentVariables([]openapi.EnvVar{
				*openapi.NewEnvVar("UPDATED", "true"),
			})
			updateStack.Spec.StackResources[0].SetExecutionConfig(*exec)
			shared.UpdateStack(client, orgID, teamName, stackID, updateStack)

			By("Waiting for the CR to reflect the update")
			Eventually(func(g Gomega) {
				cr, err := shared.GetStackCR(ctx, clusterClient, stackName, namespace)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(cr.Spec.StackResources).To(HaveLen(1))

				foundEnv := false
				for _, env := range cr.Spec.StackResources[0].Spec.EnvironmentVariables {
					if env.Name == "UPDATED" && env.Value == "true" {
						foundEnv = true
						break
					}
				}
				g.Expect(foundEnv).To(BeTrue(), "Stack CR should have UPDATED=true env var")
			}, 2*time.Minute, 5*time.Second).Should(Succeed())

			By("Waiting for stack to return to Ready state")
			shared.WaitForStackReady(client, orgID, teamName, stackID, 5*time.Minute)
		})
	})

	Context("Stack with PostgresAddon", func() {
		It("should inject postgres addon env vars into stack resource", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a postgres addon with a database")
			addon := shared.CreatePostgresAddonWithResources("test-stack-pg")
			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)
			addonID := createdAddon.GetId()

			DeferCleanup(func() {
				shared.DeletePostgresAddon(client, orgID, teamName, addonID)
			})

			By("Waiting for the postgres addon to become Ready")
			shared.WaitForAddonReady(client, orgID, teamName, addonID, 10*time.Minute)

			By("Creating a stack that references the postgres addon")
			stack := shared.CreateStackWithPostgresAddon("test-pg-stack", addonID, "testdb")
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, teamName, stackID, 5*time.Minute)

			By("Verifying Deployment has all postgres env vars from the mapping")
			deploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, "app")
			Expect(err).NotTo(HaveOccurred())

			for credField, envName := range shared.PostgresEnvMapping {
				val, found := shared.GetContainerEnvVar(deploy, envName)
				Expect(found).To(BeTrue(), "env var %s (mapped from %s) should be injected", envName, credField)
				Expect(val).NotTo(BeEmpty(), "env var %s should have a non-empty value", envName)
			}

			By("Verifying postgres credential values are structurally valid")
			pgHost, _ := shared.GetContainerEnvVar(deploy, shared.PostgresEnvMapping["host"])
			Expect(pgHost).To(ContainSubstring(".svc.cluster.local"), "PG_HOST should be a fully-qualified K8s service DNS name")

			pgPort, _ := shared.GetContainerEnvVar(deploy, shared.PostgresEnvMapping["port"])
			Expect(pgPort).To(Equal("5432"), "PG_PORT should be the default postgres port")

			pgUser, _ := shared.GetContainerEnvVar(deploy, shared.PostgresEnvMapping["username"])
			Expect(pgUser).NotTo(BeEmpty(), "PG_USER should be a valid database username")

			dbURL, _ := shared.GetContainerEnvVar(deploy, shared.PostgresEnvMapping["connectionString"])
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
			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)
			addonID := createdAddon.GetId()
			addonName := createdAddon.GetName()
			addonNamespace := createdAddon.GetNamespace()

			DeferCleanup(func() {
				shared.DeletePostgresAddon(client, orgID, teamName, addonID)
			})

			By("Waiting for the postgres addon to become Ready")
			shared.WaitForAddonReady(client, orgID, teamName, addonID, 10*time.Minute)

			By("Waiting for databases to be applied")
			shared.WaitForConditionTrue(client, orgID, teamName, addonID, string(models.PostgresAddonConditionDatabasesApplied), 2*time.Minute)

			By("Creating a stack that references the addon with superuser mode")
			stack := shared.CreateStackWithPostgresAddonSuperuser("test-su-stack", addonID)
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
			})

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, teamName, stackID, 5*time.Minute)

			By("Verifying Deployment has all postgres env vars from the mapping")
			deploy, err := shared.GetDeploymentForStackResource(ctx, clusterClient, namespace, "app")
			Expect(err).NotTo(HaveOccurred())

			for credField, envName := range shared.PostgresEnvMapping {
				val, found := shared.GetContainerEnvVar(deploy, envName)
				Expect(found).To(BeTrue(), "env var %s (mapped from %s) should be injected", envName, credField)
				Expect(val).NotTo(BeEmpty(), "env var %s should have a non-empty value", envName)
			}

			By("Verifying superuser credential values on deployment")
			pgUser, _ := shared.GetContainerEnvVar(deploy, shared.PostgresEnvMapping["username"])
			Expect(pgUser).To(Equal("postgres"), "superuser username should be 'postgres'")

			dbURL, _ := shared.GetContainerEnvVar(deploy, shared.PostgresEnvMapping["connectionString"])
			Expect(dbURL).To(HavePrefix("postgresql://"), "DATABASE_URL should be a postgresql:// connection string")

			By("Fetching superuser credentials via API")
			creds := shared.GetSuperuserCredentials(client, orgID, teamName, addonID)
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
			defer testDB.Close()

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
			defer appDB.Close()

			_, err = appDB.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS e2e_su_test (id serial PRIMARY KEY, val text)")
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to create tables in app database")

			_, err = appDB.ExecContext(ctx, "INSERT INTO e2e_su_test (val) VALUES ('superuser_write')")
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to insert into app database")

			err = appDB.QueryRowContext(ctx, "SELECT val FROM e2e_su_test WHERE val = 'superuser_write'").Scan(&val)
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to read from app database")
			Expect(val).To(Equal("superuser_write"))

			By("Verifying superuser can create a new database")
			postgresDB := shared.ConnectToPostgres("127.0.0.1", localPort, creds.GetUsername(), creds.GetPassword(), "postgres", "disable")
			defer postgresDB.Close()

			_, err = postgresDB.ExecContext(ctx, "CREATE DATABASE e2e_superuser_created")
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to create new databases")

			newDB := shared.ConnectToPostgres("127.0.0.1", localPort, creds.GetUsername(), creds.GetPassword(), "e2e_superuser_created", "disable")
			defer newDB.Close()

			var result int
			err = newDB.QueryRowContext(ctx, "SELECT 1").Scan(&result)
			Expect(err).NotTo(HaveOccurred(), "superuser should be able to connect to the newly created database")
			Expect(result).To(Equal(1))
		})
	})

	Context("Crash Detection", func() {
		It("should populate last_failure when a container crashes", func() {
			By("Creating a stack whose container immediately exits with error")
			stack := shared.CreateCrashingStack("test-crash-detection")
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
			})

			By("Waiting for the resource to enter Failed state")
			shared.WaitForStackResourceFailed(client, orgID, teamName, stackID, shared.CrashResourceName, 3*time.Minute)

			By("Waiting for last_failure to be populated on the resource")
			lastFailure := shared.WaitForStackResourceLastFailure(client, orgID, teamName, stackID, shared.CrashResourceName, 2*time.Minute)

			By("Verifying last_failure has the correct type and container details")
			Expect(lastFailure.Type).NotTo(BeNil())
			Expect(*lastFailure.Type).To(Equal("runtime_crash"))

			Expect(lastFailure.Container).NotTo(BeNil(), "container failure detail should be set")
			Expect(lastFailure.Container.FailureType).NotTo(BeNil())
			Expect(*lastFailure.Container.FailureType).To(BeElementOf("exit_error", "crash_loop"),
				"failure_type should be exit_error or crash_loop depending on restart count")

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
		})

		It("should build from a private git repo and expose to public", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			clientset, err := testEnv.Cluster.GetKubeClient()
			Expect(err).NotTo(HaveOccurred())
			ctx := context.Background()

			By("Creating a GitCredentials secret with the GitHub token")
			gitSecret := shared.CreateGitCredentialsSecret(shared.BuildSourceSecretName, githubToken)
			createdSecret := shared.CreateSecret(client, orgID, teamName, gitSecret)
			secretID := createdSecret.GetId()

			DeferCleanup(func() {
				shared.DeleteSecret(client, orgID, teamName, secretID)
			})

			By("Creating a stack with build source and exposed port")
			stack := shared.CreateStackWithBuildSource("test-build-source", shared.BuildSourceRepoURL, secretID)
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			stackName := created.GetName()
			namespace := created.GetNamespace()

			Expect(stackID).NotTo(BeEmpty())
			Expect(namespace).NotTo(BeEmpty())

			DeferCleanup(func() {
				if CurrentSpecReport().Failed() {
					shared.DumpBuildSourceDebugInfo(ctx, client, clusterClient, clientset, orgID, teamName, stackID, namespace)
				}
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
			})

			By("Waiting for the Stack CR to appear in the cluster")
			cr := shared.WaitForStackCRExists(ctx, clusterClient, stackName, namespace, 2*time.Minute)

			By("Verifying Stack CR has build spec with git repo URL")
			Expect(cr.Spec.StackResources).To(HaveLen(1))
			buildSpec := cr.Spec.StackResources[0].Spec.BuildSpec
			Expect(buildSpec).NotTo(BeNil(), "Stack CR resource should have a BuildSpec")
			Expect(buildSpec.SourceContext.Git).NotTo(BeNil(), "BuildSpec should have a Git source context")
			Expect(buildSpec.SourceContext.Git.RepoUrl).To(Equal(shared.BuildSourceRepoURL))

			By("Verifying git credentials secret exists in the cluster")
			shared.VerifyGitCredentialsSecretExists(ctx, clusterClient, namespace, stackID)

			By("Waiting for stack to become Ready")
			shared.WaitForStackReady(client, orgID, teamName, stackID, 10*time.Minute)

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
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK), "expected 200, got %d; body: %s", resp.StatusCode, string(body))

			By("Verifying image build record via API")
			builds := shared.ListStackBuilds(client, orgID, teamName, stackID)
			Expect(builds).NotTo(BeEmpty(), "should have at least one image build record")
		})

		It("should populate last_build_failure_detail when a build fails", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a GitCredentials secret with the GitHub token")
			gitSecret := shared.CreateGitCredentialsSecret("test-build-fail-creds", githubToken)
			createdSecret := shared.CreateSecret(client, orgID, teamName, gitSecret)
			secretID := createdSecret.GetId()

			DeferCleanup(func() {
				shared.DeleteSecret(client, orgID, teamName, secretID)
			})

			By("Creating a stack with a build source pointing to a branch with a broken Dockerfile")
			stack := shared.CreateStackWithBrokenBuildSource("test-build-fail", shared.BuildSourceRepoURL, secretID)
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			stackName := created.GetName()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				if CurrentSpecReport().Failed() {
					kubeClient, _ := testEnv.Cluster.GetKubeClient()
					shared.DumpBuildSourceDebugInfo(ctx, client, clusterClient, kubeClient, orgID, teamName, stackID, namespace)
				}
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
			})

			By("Waiting for the Stack CR to appear in the cluster")
			shared.WaitForStackCRExists(ctx, clusterClient, stackName, namespace, 2*time.Minute)

			By("Waiting for last_build_failure_detail to be populated on the image build")
			detail := shared.WaitForBuildLastFailureDetail(client, orgID, teamName, stackID, shared.BrokenBuildResourceName, 5*time.Minute)

			By("Verifying last_build_failure_detail has failure info")
			Expect(detail.FailureType).NotTo(BeNil())
			Expect(*detail.FailureType).To(BeElementOf("exit_error", "image_pull_failed"))

			By("Verifying last_failure on the stack resource reflects the build failure")
			lastFailure := shared.WaitForStackResourceLastFailure(client, orgID, teamName, stackID, shared.BrokenBuildResourceName, 2*time.Minute)
			Expect(lastFailure.Type).NotTo(BeNil())
			Expect(*lastFailure.Type).To(Equal("build_failure"))
			Expect(lastFailure.Build).NotTo(BeNil())
		})
	})
})
