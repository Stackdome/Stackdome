# PostgreSQL Addon Integration Tests - Design Document

**Date:** 2026-03-08
**Status:** Approved
**Scope:** Fix compilation errors and implement hybrid cluster management for integration tests

---

## Overview

This design addresses the broken integration test bootstrap by implementing hybrid cluster management. Tests will support both self-contained cluster creation (for CI/CD) and external Mage-managed clusters (for local development).

**Goal:** Enable PostgreSQL addon integration tests to run in two modes:
1. **Self-Contained Mode** - Tests create their own Kind cluster (no external dependencies)
2. **External Cluster Mode** - Tests use pre-created Mage-managed cluster (faster iteration)

**Test Scope:** API-only testing validating CRUD operations via REST API and database state verification. No Kubernetes CR creation or controller reconciliation testing in this iteration.

---

## Architecture

### Hybrid Cluster Management

The integration test bootstrap supports two modes:

#### Mode A: External Cluster (Mage-managed)

```
User runs: mage Dev.Deploy
  ↓
Kind cluster created + operators deployed
  ↓
TEST_KUBECONFIG=/path/to/kubeconfig go test ./test/int/...
  ↓
Tests detect existing cluster, verify connectivity, skip creation
```

#### Mode B: Self-Contained (CI/CD)

```
go test ./test/int/...
  ↓
No TEST_KUBECONFIG set
  ↓
Tests create new Kind cluster using testutil.TestCluster.Setup()
  ↓
Deploy operators and cluster agent
  ↓
Run tests, cleanup cluster on exit
```

### Bootstrap Flow (Updated)

```
Phase 1: Database Bootstrap (unchanged)
  ↓
Phase 2: Cluster Bootstrap (NEW HYBRID LOGIC)
  ↓
  ├─ Check TEST_KUBECONFIG env var
  │  ├─ If set: Use existing cluster
  │  │   ├─ Verify kubeconfig file exists
  │  │   ├─ Connect to cluster
  │  │   ├─ Verify cluster agent is running
  │  │   └─ Ready to use
  │  │
  │  └─ If not set: Create new cluster
  │      ├─ Use testutil.TestCluster with DefaultClusterConfig
  │      ├─ Call Setup() to create Kind cluster
  │      ├─ Deploy operators (Traefik, CNPG, Cert-Manager)
  │      ├─ Deploy cluster agent
  │      └─ Wait for readiness
  ↓
Phase 3: Server Bootstrap (unchanged)
  ↓
Phase 4: Client Bootstrap (unchanged)
```

### Key Design Decisions

1. **Detection happens in bootstrap layer** - `test/int/bootstrap/cluster.go` checks env var
2. **testutil remains unchanged** - No modifications to `pkg/testutil/` package
3. **Cleanup behavior differs** - External clusters preserved, created clusters destroyed
4. **Both paths return same interface** - `*testutil.TestCluster` regardless of mode

---

## Components & Implementation

### Modified Files

**`test/int/bootstrap/cluster.go`** - Fix compilation errors and add hybrid logic

**Current issues:**
- Line 53: `Kubeconfig` field doesn't exist in `ClusterConfig`
- Line 58: Passing value instead of pointer to `NewTestCluster()`
- Line 72: `DeployCRDs()` method doesn't exist

**New structure:**

```go
func (cm *ClusterManager) Bootstrap(ctx context.Context) error {
    // Check for external cluster first
    if kubeconfig := os.Getenv("TEST_KUBECONFIG"); kubeconfig != "" {
        return cm.useExistingCluster(ctx, kubeconfig)
    }
    // Fall back to creating new cluster
    return cm.createNewCluster(ctx)
}

func (cm *ClusterManager) useExistingCluster(ctx, kubeconfig) error {
    // 1. Verify kubeconfig file exists
    // 2. Load REST config from kubeconfig
    // 3. Create TestCluster wrapper (construct without Setup())
    // 4. Verify connectivity via GetKubeClient()
    // 5. Check if cluster agent deployment exists
    // 6. Deploy cluster agent if missing
    // 7. Wait for readiness
}

func (cm *ClusterManager) createNewCluster(ctx) error {
    // 1. Create DefaultClusterConfig with logger
    // 2. Create TestCluster
    // 3. Call Setup() to create Kind cluster
    // 4. Deploy cluster agent via DeployClusterAgent()
    // 5. Wait for readiness
}
```

### Environment Variables

**Existing:**
- `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD` - Database configuration
- `TEST_LOG_LEVEL` - Logging verbosity (debug, info, error)
- `KEEP_CLUSTER` - Preserve infrastructure for debugging
- `CLUSTER_AGENT_IMAGE_TAG` - Override cluster agent version

**New:**
- `TEST_KUBECONFIG` - Path to existing cluster kubeconfig (optional)

### Cluster Lifecycle Matrix

| Mode | Create Cluster | Deploy Agent | Cleanup Cluster |
|------|----------------|--------------|-----------------|
| External (`TEST_KUBECONFIG` set) | ❌ Skip | ✅ Verify/Deploy if needed | ❌ Preserve |
| Self-Contained (not set) | ✅ Create via Setup() | ✅ Deploy | ✅ Destroy (unless `KEEP_CLUSTER=true`) |

---

## Data Flow & Error Handling

### Bootstrap Data Flow

```
START: BeforeSuite()
  ↓
Load .env file
  ↓
Phase 1: Database Bootstrap
  ├─ Create temp database with unique name
  ├─ Run all migrations
  └─ Return SessionFactory
  ↓
Phase 2: Cluster Bootstrap (HYBRID)
  ├─ Read TEST_KUBECONFIG env var
  │
  ├─ Path A: External Cluster
  │   ├─ Stat kubeconfig file → Error if not exists
  │   ├─ Load REST config from file → Error if invalid
  │   ├─ Create TestCluster wrapper (no Setup() call)
  │   ├─ Verify connectivity via GetKubeClient() → Error if unreachable
  │   ├─ Check cluster agent deployment exists
  │   ├─ Deploy cluster agent if missing → Error if fails
  │   └─ Return TestCluster
  │
  └─ Path B: New Cluster
      ├─ Create DefaultClusterConfig(name, logger)
      ├─ Create TestCluster(&config)
      ├─ Call Setup(ctx) → Error if fails (10min timeout)
      ├─ Deploy cluster agent via DeployClusterAgent(ctx, imageTag)
      ├─ Wait for readiness (30s sleep)
      └─ Return TestCluster
  ↓
Phase 3: Server Bootstrap
  ├─ Initialize environment with all 18 services
  ├─ Start API server on random port (8987+)
  ├─ Wait for /health endpoint to respond
  └─ Return server manager
  ↓
Phase 4: Client Bootstrap
  ├─ Create test user and organization
  ├─ Obtain JWT authentication token
  ├─ Configure OpenAPI client with token
  ├─ Deploy service account to cluster
  ├─ Register cluster with API server
  └─ Return authenticated client
  ↓
READY: Tests run with shared environment
  ├─ OpenAPI client ready
  ├─ Test organization and cluster IDs available
  ├─ Database state clean (cleared between tests)
  └─ Kubernetes cluster accessible
```

### Error Handling Strategy

**Cluster Bootstrap Errors:**

1. **TEST_KUBECONFIG file not found**
   - Error: `"Kubeconfig file does not exist at %s. Either create cluster with 'mage Dev.Deploy' or unset TEST_KUBECONFIG to create new cluster"`
   - Action: Exit immediately, no cleanup needed

2. **External cluster unreachable**
   - Error: `"Failed to connect to external cluster: %v. Check if cluster is running: kubectl cluster-info"`
   - Action: Exit immediately, preserve cluster (user-managed)

3. **Cluster agent not found and deploy fails**
   - Error: `"Failed to deploy cluster agent to external cluster: %v. Check cluster access and CRD deployment"`
   - Action: Exit immediately, preserve cluster, log deployment details

4. **New cluster creation timeout**
   - Error: `"Failed to create test cluster after 10 minutes: %v. Check Docker resources: docker system df"`
   - Action: Attempt cluster cleanup, exit with error

5. **Cluster agent deployment fails on new cluster**
   - Error: `"Failed to deploy cluster agent: %v"`
   - Action: Cleanup cluster (unless KEEP_CLUSTER=true), exit

**Graceful Degradation:**
- If cluster bootstrap fails → cleanup database, log clear instructions, exit
- If server bootstrap fails → cleanup cluster and database, exit
- If client bootstrap fails → cleanup all resources, exit
- All errors logged with context and recovery suggestions

**Cleanup Guarantees:**

| Scenario | Database | External Cluster | Created Cluster |
|----------|----------|------------------|-----------------|
| Bootstrap success | Drop on exit | Preserve | Destroy (unless KEEP_CLUSTER) |
| Bootstrap failure | Drop immediately | Preserve | Destroy (unless KEEP_CLUSTER) |
| Test failure | Drop on exit | Preserve | Destroy (unless KEEP_CLUSTER) |

---

## Test Coverage

### Existing Test Coverage (Unchanged)

**PostgreSQL Addon CRUD Operations:**
- ✅ Create minimal addon (default values)
- ✅ Create addon with resources (CPU, memory, databases)
- ✅ Retrieve addon by ID
- ✅ List multiple addons
- ✅ Update addon (instances, resources, databases)
- ✅ Delete addon (currently skipped with `XIt`)

**Validation Tests:**
- ✅ Reject invalid storage size format
- ✅ Reject zero instance count
- ✅ Reject invalid PostgreSQL version

**Advanced Features:**
- ✅ Backup configuration
- ✅ High availability with placement policies

**Test Infrastructure:**
- ✅ Shared environment with BeforeEach cleanup
- ✅ OpenAPI client helpers (`test/int/shared/`)
- ✅ Fixture creation functions
- ✅ Assertion helpers (ExpectPostgresAddonEqual)

### What's Being Fixed

**Compilation Errors (3 issues):**
1. Remove `Kubeconfig: kubeconfig` from ClusterConfig struct literal
2. Change `NewTestCluster(config)` to `NewTestCluster(&config)`
3. Remove `DeployCRDs(ctx)` call (handled internally by Setup())

**New Functionality:**
- Hybrid cluster detection in `cluster.go`
- External cluster connection handling
- Conditional cleanup based on cluster mode
- Enhanced error messages with recovery instructions

### Success Criteria

After implementation, the integration tests must:

1. ✅ **Compile without errors** - All 3 compilation issues resolved
2. ✅ **Run in self-contained mode** - Create cluster when TEST_KUBECONFIG unset
3. ✅ **Run in external mode** - Use existing cluster when TEST_KUBECONFIG set
4. ✅ **Pass all test cases** - 14 tests pass, 1 pending (delete test)
5. ✅ **Clean up appropriately** - Respect cluster ownership and KEEP_CLUSTER flag
6. ✅ **Provide clear errors** - Actionable messages when bootstrap fails

---

## Validation Plan

### Phase 1: Verify Compilation

```bash
cd /Users/asnaraya/projects/skysync/api-server/api-server
go build ./test/int/...
```

**Expected:** No compilation errors

### Phase 2: Test Self-Contained Mode

```bash
# Clean environment
unset TEST_KUBECONFIG

# Run integration tests
go test ./test/int/... -v -count=1

# Verify behavior:
# - Creates new Kind cluster
# - Deploys operators and cluster agent
# - Runs all tests
# - Destroys cluster on exit
```

**Expected:** Tests create cluster, pass, cleanup

### Phase 3: Test External Cluster Mode

```bash
# Terminal 1: Create Mage-managed cluster
cd /Users/asnaraya/projects/skysync/cluster-agent
./mage Dev.Deploy
# Wait for cluster to be ready

# Terminal 2: Run tests with existing cluster
export TEST_KUBECONFIG=~/.kube/config
cd /Users/asnaraya/projects/skysync/api-server/api-server
go test ./test/int/... -v -count=1

# Verify behavior:
# - Uses existing cluster
# - Skips cluster creation
# - Verifies cluster agent is running
# - Runs all tests
# - Preserves cluster on exit
```

**Expected:** Tests use external cluster, pass, no cleanup

### Phase 4: Verify KEEP_CLUSTER Behavior

```bash
# Clean environment
unset TEST_KUBECONFIG

# Run with keep cluster flag
KEEP_CLUSTER=true go test ./test/int/... -v

# After tests complete:
kind get clusters

# Verify cluster still exists:
kubectl cluster-info --context kind-stackdome-int-test
```

**Expected:** Cluster remains after test completion

### Phase 5: Error Handling Validation

**Test missing kubeconfig:**
```bash
export TEST_KUBECONFIG=/nonexistent/path
go test ./test/int/... -v
# Expected: Clear error about missing file
```

**Test unreachable cluster:**
```bash
export TEST_KUBECONFIG=/path/to/stopped/cluster/config
go test ./test/int/... -v
# Expected: Clear error about unreachable cluster
```

**Test cluster creation failure:**
```bash
# Stop Docker
docker stop $(docker ps -q)
unset TEST_KUBECONFIG
go test ./test/int/... -v
# Expected: Clear error about Docker/resources
```

---

## Implementation Notes

### Files to Modify

1. **`test/int/bootstrap/cluster.go`**
   - Add `useExistingCluster()` method
   - Add `createNewCluster()` method
   - Update `Bootstrap()` to implement hybrid logic
   - Fix compilation errors (3 places)
   - Add helper to load kubeconfig and create REST config

2. **`test/int/README.md`** (optional documentation update)
   - Document TEST_KUBECONFIG environment variable
   - Add examples for both modes
   - Update troubleshooting section

### Files NOT Modified

- ✅ `pkg/testutil/cluster.go` - Remains unchanged
- ✅ `test/int/integration_test.go` - No changes needed
- ✅ `test/int/postgres_addon_test.go` - Tests remain as-is
- ✅ `test/int/shared/*` - Helper functions unchanged
- ✅ `test/int/bootstrap/server.go` - Server bootstrap unchanged
- ✅ `test/int/bootstrap/client.go` - Client bootstrap unchanged
- ✅ `test/int/bootstrap/database.go` - Database bootstrap unchanged

### Dependencies

**Required packages (already in go.mod):**
- `github.com/onsi/ginkgo/v2` - BDD test framework
- `github.com/onsi/gomega` - Assertion library
- `sigs.k8s.io/kind` - Kind cluster management (via testutil)
- `k8s.io/client-go` - Kubernetes client
- `github.com/mt-sre/devkube/dev` - Development cluster utilities

**No new dependencies required.**

---

## Risks & Mitigation

### Risk 1: External cluster in unknown state

**Scenario:** User provides TEST_KUBECONFIG to cluster with wrong operators or old CRDs

**Mitigation:**
- Verify cluster agent deployment exists and is running
- Check CRD versions if possible
- Provide clear error if cluster agent missing: "Cluster agent not found. Ensure cluster was created with 'mage Dev.Deploy'"

### Risk 2: Cluster cleanup fails

**Scenario:** Kind cluster deletion fails, leaving orphaned resources

**Mitigation:**
- Log cluster name before deletion attempt
- Provide manual cleanup instructions in error message
- Document `kind delete cluster --name <name>` in README

### Risk 3: Port conflicts in parallel test runs

**Scenario:** Multiple test runs try to use same ports

**Mitigation:**
- Server already uses random ports (8987+)
- Database uses unique names with timestamps
- Kind clusters use unique names
- Parallel runs should not conflict

### Risk 4: Long bootstrap times

**Scenario:** Creating cluster takes 5+ minutes, slowing development iteration

**Mitigation:**
- This is exactly why we support external cluster mode!
- Developers use `mage Dev.Deploy` once, then TEST_KUBECONFIG for fast iterations
- Document this workflow clearly in README

---

## Future Enhancements (Out of Scope)

**Not included in this implementation:**

1. **Kubernetes CR verification** - Tests don't verify PostgresCluster CRs are created
2. **Controller reconciliation testing** - No validation of operator behavior
3. **Bidirectional status sync** - No testing of cluster → API server status updates
4. **Advanced backup/restore** - No actual backup/restore operation testing
5. **Multi-cluster testing** - Single cluster only for now

These features can be added incrementally in future iterations.

---

## Timeline & Effort Estimate

**Implementation phases:**

1. **Fix compilation errors** - 15 minutes
   - Remove Kubeconfig field reference
   - Fix pointer passing
   - Remove DeployCRDs call

2. **Implement hybrid cluster logic** - 45 minutes
   - Add useExistingCluster() method
   - Add createNewCluster() method
   - Add kubeconfig loading helper
   - Update Bootstrap() method

3. **Testing & validation** - 30 minutes
   - Test self-contained mode
   - Test external cluster mode
   - Test error cases
   - Verify cleanup behavior

4. **Documentation updates** - 15 minutes (optional)
   - Update README with TEST_KUBECONFIG usage
   - Add examples for both modes

**Total estimated effort: 1.5-2 hours**

---

## Approval & Sign-Off

- ✅ Architecture approved
- ✅ Component design approved
- ✅ Error handling strategy approved
- ✅ Test coverage plan approved

**Next step:** Create implementation plan using writing-plans skill.
