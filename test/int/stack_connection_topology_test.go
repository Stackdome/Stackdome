package int

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/test/int/shared"
)

var _ = Describe("Stack Connections & Topology", func() {
	var client *openapi.APIClient
	var orgID string
	teamName := models.DefaultTeamName

	BeforeEach(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil(), "Test environment should be initialized")
		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	// -----------------------------------------------------------------
	// Connection CRUD — Happy Path
	// -----------------------------------------------------------------
	Context("Connection CRUD Happy Path", func() {
		var stackID string

		BeforeEach(func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			stack := shared.CreateSkipProvisioningStack("test-conn-crud", []openapi.StackResource{api, web})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID = created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)
		})

		It("should create an env connection (resource-to-resource)", func() {
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			created := shared.CreateStackConnection(client, orgID, teamName, stackID, &conn)

			Expect(created.GetId()).NotTo(BeEmpty())
			Expect(created.GetKind()).To(Equal("env"))
			Expect(created.From.GetName()).To(Equal("api"))
			Expect(created.To.GetName()).To(Equal("web"))
			Expect(created.GetMappings()).To(HaveLen(1))
			Expect(created.GetMappings()[0].Target.GetName()).To(Equal("API_HOST"))
		})

		It("should create a secret env connection", func() {
			secret := shared.CreateGenericSecret("test-secret-conn", map[string]string{
				"api_key": "test-key-value",
			})
			createdSecret := shared.CreateSecret(client, orgID, teamName, secret)
			secretID := createdSecret.GetId()

			target := openapi.NewConnectionTarget("env")
			target.SetName("API_KEY")
			value := openapi.NewValueRef()
			value.SetOutput("key.api_key")
			mapping := *openapi.NewConnectionMapping(*target, *value)

			conn := shared.EnvConnectionWithID("secret", secretID, "stack_resource", "api", []openapi.ConnectionMapping{mapping})
			created := shared.CreateStackConnection(client, orgID, teamName, stackID, &conn)

			Expect(created.GetId()).NotTo(BeEmpty())
			Expect(created.From.GetType()).To(Equal("secret"))
			Expect(created.From.GetId()).To(Equal(secretID))
			Expect(created.GetMappings()[0].Value.GetOutput()).To(Equal("key.api_key"))
		})

		It("should create a volume mount connection", func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			vol := openapi.NewVolume("data", *openapi.NewVolumeSpec("1Gi", false, openapi.READ_WRITE_ONCE))
			stack := shared.CreateSkipProvisioningStackWithVolumes("test-vol-conn", []openapi.StackResource{api, web}, []openapi.Volume{*vol})
			created := shared.CreateStack(client, orgID, teamName, stack)
			volStackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, volStackID, 1*time.Minute)

			conn := shared.VolumeMountConn("data", "api", "/mnt/data", "subdir")
			createdConn := shared.CreateStackConnection(client, orgID, teamName, volStackID, &conn)

			Expect(createdConn.GetId()).NotTo(BeEmpty())
			Expect(createdConn.GetKind()).To(Equal("volume_mount"))
			config := createdConn.GetConfig()
			Expect(config.VolumeMountConfig).NotTo(BeNil())
			Expect(config.VolumeMountConfig.MountPath).To(Equal("/mnt/data"))
			Expect(config.VolumeMountConfig.GetSubPath()).To(Equal("subdir"))
		})

		It("should return empty list for stack with no connections", func() {
			connections := shared.ListStackConnections(client, orgID, teamName, stackID)
			Expect(connections).To(BeEmpty())
		})

		It("should list multiple connections", func() {
			conn1 := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			shared.CreateStackConnection(client, orgID, teamName, stackID, &conn1)

			target := openapi.NewConnectionTarget("env")
			target.SetName("API_URL")
			value := openapi.NewValueRef()
			value.SetOutput("url.http")
			conn2 := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{
				*openapi.NewConnectionMapping(*target, *value),
			})
			// conn2 has same from/to but different kind won't work, so let's create a different pairing
			// Actually duplicate from/to/kind will 409. Let's use a secret instead.
			secret := shared.CreateGenericSecret("test-multi-conn-secret", map[string]string{"k": "v"})
			createdSecret := shared.CreateSecret(client, orgID, teamName, secret)

			secretTarget := openapi.NewConnectionTarget("env")
			secretTarget.SetName("SECRET_KEY")
			secretValue := openapi.NewValueRef()
			secretValue.SetOutput("key.k")
			conn2 = shared.EnvConnectionWithID("secret", createdSecret.GetId(), "stack_resource", "web", []openapi.ConnectionMapping{
				*openapi.NewConnectionMapping(*secretTarget, *secretValue),
			})
			shared.CreateStackConnection(client, orgID, teamName, stackID, &conn2)

			connections := shared.ListStackConnections(client, orgID, teamName, stackID)
			Expect(connections).To(HaveLen(2))
		})

		It("should update connection mappings", func() {
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			created := shared.CreateStackConnection(client, orgID, teamName, stackID, &conn)
			connectionID := created.GetId()

			updatedTarget := openapi.NewConnectionTarget("env")
			updatedTarget.SetName("UPSTREAM_HOST")
			updatedValue := openapi.NewValueRef()
			updatedValue.SetOutput("host")
			conn.SetMappings([]openapi.ConnectionMapping{*openapi.NewConnectionMapping(*updatedTarget, *updatedValue)})

			updated := shared.UpdateStackConnection(client, orgID, teamName, stackID, connectionID, &conn)
			Expect(updated.GetMappings()).To(HaveLen(1))
			Expect(updated.GetMappings()[0].Target.GetName()).To(Equal("UPSTREAM_HOST"))
		})

		It("should update connection config", func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			vol := openapi.NewVolume("cfg-data", *openapi.NewVolumeSpec("1Gi", false, openapi.READ_WRITE_ONCE))
			stack := shared.CreateSkipProvisioningStackWithVolumes("test-upd-cfg", []openapi.StackResource{api, web}, []openapi.Volume{*vol})
			created := shared.CreateStack(client, orgID, teamName, stack)
			cfgStackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, cfgStackID, 1*time.Minute)

			conn := shared.VolumeMountConn("cfg-data", "api", "/mnt/data", "")
			createdConn := shared.CreateStackConnection(client, orgID, teamName, cfgStackID, &conn)

			readOnly := true
			subPath := "new-sub"
			conn.SetConfig(openapi.StackConnectionConfig{
				VolumeMountConfig: &openapi.VolumeMountConfig{
					MountPath: "/mnt/updated",
					SubPath:   &subPath,
					ReadOnly:  &readOnly,
				},
			})
			updated := shared.UpdateStackConnection(client, orgID, teamName, cfgStackID, createdConn.GetId(), &conn)
			updatedCfg := updated.GetConfig()
			Expect(updatedCfg.VolumeMountConfig).NotTo(BeNil())
			Expect(updatedCfg.VolumeMountConfig.MountPath).To(Equal("/mnt/updated"))
			Expect(updatedCfg.VolumeMountConfig.GetSubPath()).To(Equal("new-sub"))
			Expect(updatedCfg.VolumeMountConfig.GetReadOnly()).To(BeTrue())
		})

		It("should delete a connection", func() {
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			created := shared.CreateStackConnection(client, orgID, teamName, stackID, &conn)
			connectionID := created.GetId()

			shared.DeleteStackConnection(client, orgID, teamName, stackID, connectionID)

			connections := shared.ListStackConnections(client, orgID, teamName, stackID)
			Expect(connections).To(BeEmpty())
		})

		It("should complete a full CRUD lifecycle on a single connection", func() {
			By("Creating")
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			created := shared.CreateStackConnection(client, orgID, teamName, stackID, &conn)
			connectionID := created.GetId()
			Expect(connectionID).NotTo(BeEmpty())

			By("Listing after create")
			connections := shared.ListStackConnections(client, orgID, teamName, stackID)
			Expect(connections).To(HaveLen(1))
			Expect(connections[0].GetId()).To(Equal(connectionID))

			By("Updating")
			updatedTarget := openapi.NewConnectionTarget("env")
			updatedTarget.SetName("CHANGED_HOST")
			updatedValue := openapi.NewValueRef()
			updatedValue.SetOutput("host")
			conn.SetMappings([]openapi.ConnectionMapping{*openapi.NewConnectionMapping(*updatedTarget, *updatedValue)})
			updated := shared.UpdateStackConnection(client, orgID, teamName, stackID, connectionID, &conn)
			Expect(updated.GetMappings()[0].Target.GetName()).To(Equal("CHANGED_HOST"))

			By("Listing after update")
			connections = shared.ListStackConnections(client, orgID, teamName, stackID)
			Expect(connections).To(HaveLen(1))
			Expect(connections[0].GetMappings()[0].Target.GetName()).To(Equal("CHANGED_HOST"))

			By("Deleting")
			shared.DeleteStackConnection(client, orgID, teamName, stackID, connectionID)

			By("Listing after delete")
			connections = shared.ListStackConnections(client, orgID, teamName, stackID)
			Expect(connections).To(BeEmpty())
		})
	})

	// -----------------------------------------------------------------
	// Topology — Happy Path
	// -----------------------------------------------------------------
	Context("Topology Happy Path", func() {
		It("should return nodes for resources with no edges when there are no connections", func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			stack := shared.CreateSkipProvisioningStack("test-topo-empty", []openapi.StackResource{api, web})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)

			topology := shared.GetStackTopology(client, orgID, teamName, stackID)
			Expect(topology.GetNodes()).To(HaveLen(2))
			Expect(topology.GetEdges()).To(BeEmpty())

			nodeNames := []string{}
			for _, n := range topology.GetNodes() {
				nodeNames = append(nodeNames, n.GetLabel())
			}
			Expect(nodeNames).To(ContainElements("api", "web"))
		})

		It("should show an edge with source_of_truth connection for resource-to-resource env", func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			stack := shared.CreateSkipProvisioningStackWithConnections("test-topo-r2r", []openapi.StackResource{api, web}, []openapi.StackConnection{conn})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)

			topology := shared.GetStackTopology(client, orgID, teamName, stackID)
			Expect(topology.GetEdges()).To(HaveLen(1))
			edge := topology.GetEdges()[0]
			Expect(edge.GetSourceOfTruth()).To(Equal("connection"))
			Expect(edge.GetKind()).To(Equal("env"))
			Expect(edge.Source.GetName()).To(Equal("api"))
			Expect(edge.Target.GetName()).To(Equal("web"))
		})

		It("should show depends_on edges with source_of_truth depends_on", func() {
			db := shared.ResourceWithPort("database", 5432)
			app := shared.ResourceWithDependsOn("app", []string{"database"})
			stack := shared.CreateSkipProvisioningStack("test-topo-deps", []openapi.StackResource{db, app})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)

			topology := shared.GetStackTopology(client, orgID, teamName, stackID)
			Expect(topology.GetEdges()).To(HaveLen(1))
			edge := topology.GetEdges()[0]
			Expect(edge.GetSourceOfTruth()).To(Equal("derived"))
			Expect(edge.GetKind()).To(Equal("depends_on"))
		})

		It("should show both connection and depends_on edges together", func() {
			api := shared.ResourceWithPort("api", 8080)
			worker := shared.ResourceWithDependsOn("worker", []string{"api"})
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "worker", []openapi.ConnectionMapping{shared.HostMapping()})
			stack := shared.CreateSkipProvisioningStackWithConnections("test-topo-mixed", []openapi.StackResource{api, worker}, []openapi.StackConnection{conn})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)

			topology := shared.GetStackTopology(client, orgID, teamName, stackID)
			Expect(len(topology.GetEdges())).To(BeNumerically(">=", 2))

			hasDepsEdge := false
			hasConnEdge := false
			for _, e := range topology.GetEdges() {
				if e.GetSourceOfTruth() == "derived" {
					hasDepsEdge = true
				}
				if e.GetSourceOfTruth() == "connection" {
					hasConnEdge = true
				}
			}
			Expect(hasDepsEdge).To(BeTrue(), "should have a depends_on edge")
			Expect(hasConnEdge).To(BeTrue(), "should have a connection edge")
		})

		It("should have topology node labels matching resource names", func() {
			svc := shared.ResourceWithPort("my-service", 3000)
			bg := shared.SimpleResource("background-job")
			stack := shared.CreateSkipProvisioningStack("test-topo-labels", []openapi.StackResource{svc, bg})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)

			topology := shared.GetStackTopology(client, orgID, teamName, stackID)
			labels := map[string]bool{}
			for _, n := range topology.GetNodes() {
				labels[n.GetLabel()] = true
			}
			Expect(labels).To(HaveKey("my-service"))
			Expect(labels).To(HaveKey("background-job"))
		})

		It("should reflect connection mutations in topology", func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			stack := shared.CreateSkipProvisioningStack("test-topo-mutate", []openapi.StackResource{api, web})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)

			By("Topology should start with no edges")
			topology := shared.GetStackTopology(client, orgID, teamName, stackID)
			Expect(topology.GetEdges()).To(BeEmpty())

			By("Adding a connection should add an edge")
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			createdConn := shared.CreateStackConnection(client, orgID, teamName, stackID, &conn)

			topology = shared.GetStackTopology(client, orgID, teamName, stackID)
			Expect(topology.GetEdges()).To(HaveLen(1))
			Expect(topology.GetEdges()[0].GetId()).To(Equal(createdConn.GetId()))

			By("Deleting the connection should remove the edge")
			shared.DeleteStackConnection(client, orgID, teamName, stackID, createdConn.GetId())

			topology = shared.GetStackTopology(client, orgID, teamName, stackID)
			Expect(topology.GetEdges()).To(BeEmpty())
		})
	})

	// -----------------------------------------------------------------
	// Multi-database discriminator
	// -----------------------------------------------------------------
	Context("Multi-database addon connections", func() {
		It("should allow two env connections from the same postgres addon to the same resource with different databases", func() {
			By("Creating a postgres addon with two databases")
			addon := shared.CreateMinimalPostgresAddon("test-multi-db")
			db1 := openapi.NewPostgresDatabase("appdb")
			db1.SetExtensions([]string{})
			db2 := openapi.NewPostgresDatabase("tooljetdb")
			db2.SetExtensions([]string{})
			addon.Spec.SetDatabases([]openapi.PostgresDatabase{*db1, *db2})
			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)
			addonID := createdAddon.GetId()

			By("Creating a stack with a resource")
			web := shared.ResourceWithPort("web", 8080)
			stack := shared.CreateSkipProvisioningStack("test-multi-db-conn", []openapi.StackResource{web})
			createdStack := shared.CreateStack(client, orgID, teamName, stack)
			stackID := createdStack.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)

			By("Creating first connection targeting appdb")
			conn1 := shared.EnvConnectionWithID("addon/postgres", addonID, "stack_resource", "web", []openapi.ConnectionMapping{
				shared.PostgresMapping("host", "PG_HOST"),
				shared.PostgresMapping("database", "PG_DB"),
			})
			appdb := "appdb"
			conn1.SetConfig(openapi.StackConnectionConfig{
				PostgresEnvConfig: &openapi.PostgresEnvConfig{Database: &appdb},
			})
			created1 := shared.CreateStackConnection(client, orgID, teamName, stackID, &conn1)
			Expect(created1.GetId()).NotTo(BeEmpty())

			By("Creating second connection targeting tooljetdb (same addon, same resource, different database)")
			conn2 := shared.EnvConnectionWithID("addon/postgres", addonID, "stack_resource", "web", []openapi.ConnectionMapping{
				shared.PostgresMapping("host", "TOOLJET_DB_HOST"),
				shared.PostgresMapping("database", "TOOLJET_DB"),
			})
			tooljetdb := "tooljetdb"
			conn2.SetConfig(openapi.StackConnectionConfig{
				PostgresEnvConfig: &openapi.PostgresEnvConfig{Database: &tooljetdb},
			})
			created2 := shared.CreateStackConnection(client, orgID, teamName, stackID, &conn2)
			Expect(created2.GetId()).NotTo(BeEmpty())

			By("Verifying both connections exist")
			connections := shared.ListStackConnections(client, orgID, teamName, stackID)
			Expect(connections).To(HaveLen(2))

			databases := map[string]bool{}
			for _, c := range connections {
				Expect(c.GetConfig().PostgresEnvConfig).NotTo(BeNil())
				databases[c.GetConfig().PostgresEnvConfig.GetDatabase()] = true
			}
			Expect(databases).To(HaveKey("appdb"))
			Expect(databases).To(HaveKey("tooljetdb"))
		})

		It("should reject a second env connection targeting the same database with 409", func() {
			By("Creating a postgres addon")
			addon := shared.CreateMinimalPostgresAddon("test-dup-db")
			db1 := openapi.NewPostgresDatabase("appdb")
			db1.SetExtensions([]string{})
			addon.Spec.SetDatabases([]openapi.PostgresDatabase{*db1})
			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)
			addonID := createdAddon.GetId()

			By("Creating a stack with a resource")
			web := shared.ResourceWithPort("web", 8080)
			stack := shared.CreateSkipProvisioningStack("test-dup-db-conn", []openapi.StackResource{web})
			createdStack := shared.CreateStack(client, orgID, teamName, stack)
			stackID := createdStack.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)

			By("Creating first connection targeting appdb")
			conn1 := shared.EnvConnectionWithID("addon/postgres", addonID, "stack_resource", "web", []openapi.ConnectionMapping{
				shared.PostgresMapping("host", "PG_HOST"),
			})
			appdb := "appdb"
			conn1.SetConfig(openapi.StackConnectionConfig{
				PostgresEnvConfig: &openapi.PostgresEnvConfig{Database: &appdb},
			})
			created := shared.CreateStackConnection(client, orgID, teamName, stackID, &conn1)
			Expect(created.GetId()).NotTo(BeEmpty())

			By("Attempting duplicate connection targeting same database")
			conn2 := shared.EnvConnectionWithID("addon/postgres", addonID, "stack_resource", "web", []openapi.ConnectionMapping{
				shared.PostgresMapping("host", "PG_HOST_2"),
			})
			conn2.SetConfig(openapi.StackConnectionConfig{
				PostgresEnvConfig: &openapi.PostgresEnvConfig{Database: &appdb},
			})
			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, &conn2, http.StatusConflict)
		})
	})

	// -----------------------------------------------------------------
	// Connection Create — Negative Paths
	// -----------------------------------------------------------------
	Context("Connection Create Negative Paths", func() {
		var stackID string

		BeforeEach(func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			stack := shared.CreateSkipProvisioningStack("test-conn-neg", []openapi.StackResource{api, web})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID = created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)
		})

		It("should reject duplicate connection (same kind, from, to) with 409", func() {
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			shared.CreateStackConnection(client, orgID, teamName, stackID, &conn)

			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, &conn, http.StatusConflict)
		})

		It("should reject connection with invalid kind", func() {
			from := openapi.NewTopologyNodeRef("stack_resource")
			from.SetName("api")
			to := openapi.NewTopologyNodeRef("stack_resource")
			to.SetName("web")
			conn := openapi.NewStackConnection("invalid_kind", *from, *to)

			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, conn, http.StatusBadRequest)
		})

		It("should reject env connection with missing from.name for stack_resource", func() {
			from := openapi.NewTopologyNodeRef("stack_resource")
			// from.name not set
			to := openapi.NewTopologyNodeRef("stack_resource")
			to.SetName("web")
			conn := openapi.NewStackConnection("env", *from, *to)
			conn.SetMappings([]openapi.ConnectionMapping{shared.HostMapping()})

			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, conn, http.StatusBadRequest)
		})

		It("should reject env connection with missing to.name for stack_resource", func() {
			from := openapi.NewTopologyNodeRef("stack_resource")
			from.SetName("api")
			to := openapi.NewTopologyNodeRef("stack_resource")
			// to.name not set
			conn := openapi.NewStackConnection("env", *from, *to)
			conn.SetMappings([]openapi.ConnectionMapping{shared.HostMapping()})

			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, conn, http.StatusBadRequest)
		})

		It("should reject connection referencing non-existent from resource", func() {
			conn := shared.EnvConnection("stack_resource", "nonexistent", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, &conn, http.StatusBadRequest)
		})

		It("should reject connection referencing non-existent to resource", func() {
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "nonexistent", []openapi.ConnectionMapping{shared.HostMapping()})
			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, &conn, http.StatusBadRequest)
		})

		It("should reject secret connection with missing from.id", func() {
			from := openapi.NewTopologyNodeRef("secret")
			// from.id not set
			to := openapi.NewTopologyNodeRef("stack_resource")
			to.SetName("api")
			conn := openapi.NewStackConnection("env", *from, *to)

			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, conn, http.StatusBadRequest)
		})

		It("should reject secret connection referencing non-existent secret", func() {
			conn := shared.EnvConnectionWithID("secret", "00000000-0000-0000-0000-000000000000", "stack_resource", "api", nil)
			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, &conn, http.StatusBadRequest)
		})

		It("should reject volume mount connection with missing from.name", func() {
			from := openapi.NewTopologyNodeRef("volume")
			// from.name not set
			to := openapi.NewTopologyNodeRef("stack_resource")
			to.SetName("api")
			conn := openapi.NewStackConnection("volume_mount", *from, *to)
			conn.SetConfig(openapi.StackConnectionConfig{
				VolumeMountConfig: openapi.NewVolumeMountConfig("/mnt"),
			})

			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, conn, http.StatusBadRequest)
		})

		It("should reject volume mount referencing non-existent volume", func() {
			conn := shared.VolumeMountConn("nonexistent-vol", "api", "/mnt", "")
			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, &conn, http.StatusBadRequest)
		})

		It("should reject volume mount missing config.mount_path", func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			vol := openapi.NewVolume("test-vol", *openapi.NewVolumeSpec("1Gi", false, openapi.READ_WRITE_ONCE))
			stack := shared.CreateSkipProvisioningStackWithVolumes("test-no-mountpath", []openapi.StackResource{api, web}, []openapi.Volume{*vol})
			created := shared.CreateStack(client, orgID, teamName, stack)
			volStackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, volStackID, 1*time.Minute)

			from := openapi.NewTopologyNodeRef("volume")
			from.SetName("test-vol")
			to := openapi.NewTopologyNodeRef("stack_resource")
			to.SetName("api")
			conn := openapi.NewStackConnection("volume_mount", *from, *to)
			// No config set — mount_path missing

			shared.CreateStackConnectionExpectError(client, orgID, teamName, volStackID, conn, http.StatusBadRequest)
		})

		It("should reject env mapping referencing unsupported output", func() {
			target := openapi.NewConnectionTarget("env")
			target.SetName("BAD_OUTPUT")
			value := openapi.NewValueRef()
			value.SetOutput("nonexistent_output")
			mapping := *openapi.NewConnectionMapping(*target, *value)

			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{mapping})
			shared.CreateStackConnectionExpectError(client, orgID, teamName, stackID, &conn, http.StatusBadRequest)
		})
	})

	// -----------------------------------------------------------------
	// Connection Update/Delete — Negative Paths
	// -----------------------------------------------------------------
	Context("Connection Update/Delete Negative Paths", func() {
		var stackID string

		BeforeEach(func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			stack := shared.CreateSkipProvisioningStack("test-conn-upd-neg", []openapi.StackResource{api, web})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID = created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)
		})

		It("should return 404 when updating non-existent connection", func() {
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			shared.UpdateStackConnectionExpectError(client, orgID, teamName, stackID, "00000000-0000-0000-0000-000000000000", &conn, http.StatusNotFound)
		})

		It("should return 404 when deleting non-existent connection", func() {
			shared.DeleteStackConnectionExpectError(client, orgID, teamName, stackID, "00000000-0000-0000-0000-000000000000", http.StatusNotFound)
		})

		It("should reject update that creates invalid state (bad target resource)", func() {
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			created := shared.CreateStackConnection(client, orgID, teamName, stackID, &conn)

			conn.To.SetName("nonexistent")
			shared.UpdateStackConnectionExpectError(client, orgID, teamName, stackID, created.GetId(), &conn, http.StatusBadRequest)
		})

		It("should reject update that creates duplicate connection", func() {
			secret := shared.CreateGenericSecret("test-dup-upd-secret", map[string]string{"k": "v"})
			createdSecret := shared.CreateSecret(client, orgID, teamName, secret)
			secretID := createdSecret.GetId()

			secretTarget := openapi.NewConnectionTarget("env")
			secretTarget.SetName("KEY")
			secretValue := openapi.NewValueRef()
			secretValue.SetOutput("key.k")
			secretMapping := *openapi.NewConnectionMapping(*secretTarget, *secretValue)

			conn1 := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			shared.CreateStackConnection(client, orgID, teamName, stackID, &conn1)

			conn2 := shared.EnvConnectionWithID("secret", secretID, "stack_resource", "web", []openapi.ConnectionMapping{secretMapping})
			created2 := shared.CreateStackConnection(client, orgID, teamName, stackID, &conn2)

			conn2Update := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			shared.UpdateStackConnectionExpectError(client, orgID, teamName, stackID, created2.GetId(), &conn2Update, http.StatusConflict)
		})
	})

	// -----------------------------------------------------------------
	// Cross-cutting / Data Integrity
	// -----------------------------------------------------------------
	Context("Cross-cutting Data Integrity", func() {
		It("should preserve connections when included in stack PUT body", func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			stack := shared.CreateSkipProvisioningStackWithConnections("test-conn-persist", []openapi.StackResource{api, web}, []openapi.StackConnection{conn})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)

			By("Fetching persisted connection with its server-assigned ID")
			connections := shared.ListStackConnections(client, orgID, teamName, stackID)
			Expect(connections).To(HaveLen(1))
			existingConn := connections[0]
			connectionID := existingConn.GetId()

			By("Updating the stack spec with the existing connection (including ID) in body")
			updatedApi := shared.ResourceWithPort("api", 8080)
			exec := openapi.NewExecutionConfig()
			exec.SetEnvironmentVariables([]openapi.EnvVar{
				func() openapi.EnvVar { v := openapi.NewEnvVar("NEW_VAR"); v.SetValue("new-value"); return *v }(),
			})
			updatedApi.SetExecutionConfig(*exec)
			updateStack := shared.CreateSkipProvisioningStackWithConnections("test-conn-persist", []openapi.StackResource{updatedApi, web}, []openapi.StackConnection{existingConn})
			shared.UpdateStack(client, orgID, teamName, stackID, updateStack)

			By("Verifying connection still exists with same ID after update")
			connections = shared.ListStackConnections(client, orgID, teamName, stackID)
			Expect(connections).To(HaveLen(1))
			Expect(connections[0].GetId()).To(Equal(connectionID))
			Expect(connections[0].GetMappings()[0].Target.GetName()).To(Equal("API_HOST"))
		})

		It("should remove connections when PUT body omits them", func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			conn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			stack := shared.CreateSkipProvisioningStackWithConnections("test-conn-wipe", []openapi.StackResource{api, web}, []openapi.StackConnection{conn})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)

			By("Verifying connection exists")
			Expect(shared.ListStackConnections(client, orgID, teamName, stackID)).To(HaveLen(1))

			By("Updating the stack without connections in body")
			updateStack := shared.CreateSkipProvisioningStack("test-conn-wipe", []openapi.StackResource{api, web})
			shared.UpdateStack(client, orgID, teamName, stackID, updateStack)

			By("Verifying connections are removed")
			Expect(shared.ListStackConnections(client, orgID, teamName, stackID)).To(BeEmpty())
		})

		It("should support multiple connection types on the same stack", func() {
			api := shared.ResourceWithPort("api", 8080)
			web := shared.SimpleResource("web")
			vol := openapi.NewVolume("shared-data", *openapi.NewVolumeSpec("1Gi", false, openapi.READ_WRITE_ONCE))
			stack := shared.CreateSkipProvisioningStackWithVolumes("test-multi-kinds", []openapi.StackResource{api, web}, []openapi.Volume{*vol})
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			shared.WaitForStackReady(client, orgID, teamName, stackID, 1*time.Minute)

			secret := shared.CreateGenericSecret("test-multi-kind-secret", map[string]string{"token": "abc"})
			createdSecret := shared.CreateSecret(client, orgID, teamName, secret)

			By("Creating an env connection (resource-to-resource)")
			envConn := shared.EnvConnection("stack_resource", "api", "stack_resource", "web", []openapi.ConnectionMapping{shared.HostMapping()})
			shared.CreateStackConnection(client, orgID, teamName, stackID, &envConn)

			By("Creating a volume mount connection")
			volConn := shared.VolumeMountConn("shared-data", "api", "/mnt/shared", "")
			shared.CreateStackConnection(client, orgID, teamName, stackID, &volConn)

			By("Creating a secret env connection")
			secretTarget := openapi.NewConnectionTarget("env")
			secretTarget.SetName("AUTH_TOKEN")
			secretValue := openapi.NewValueRef()
			secretValue.SetOutput("key.token")
			secretConn := shared.EnvConnectionWithID("secret", createdSecret.GetId(), "stack_resource", "web", []openapi.ConnectionMapping{
				*openapi.NewConnectionMapping(*secretTarget, *secretValue),
			})
			shared.CreateStackConnection(client, orgID, teamName, stackID, &secretConn)

			By("Verifying all 3 connections exist")
			connections := shared.ListStackConnections(client, orgID, teamName, stackID)
			Expect(connections).To(HaveLen(3))

			kinds := map[string]bool{}
			for _, c := range connections {
				kinds[c.GetKind()] = true
			}
			Expect(kinds).To(HaveKey("env"))
			Expect(kinds).To(HaveKey("volume_mount"))

			By("Verifying topology reflects all connections")
			topology := shared.GetStackTopology(client, orgID, teamName, stackID)
			Expect(len(topology.GetEdges())).To(BeNumerically(">=", 3))
		})

		It("should return 404 for topology on non-existent stack", func() {
			shared.GetStackTopologyExpectError(client, orgID, teamName, "00000000-0000-0000-0000-000000000000", http.StatusNotFound)
		})

		It("should return 404 for list connections on non-existent stack", func() {
			shared.ListStackConnectionsExpectError(client, orgID, teamName, "00000000-0000-0000-0000-000000000000", http.StatusNotFound)
		})
	})
})
