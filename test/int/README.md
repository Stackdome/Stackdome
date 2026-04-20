# Integration Tests

This directory contains comprehensive integration tests for the Stackdome API server using Ginkgo BDD framework. The tests bootstrap a complete end-to-end environment including real Kubernetes clusters, full API server initialization, and test client setup.

## Architecture Overview

The integration tests follow a **shared infrastructure** pattern where:
- **One complete environment** is bootstrapped per test run
- **All test specs share** the same infrastructure for efficiency  
- **Real Kubernetes clusters** are always created for authentic testing
- **Full API server setup** with all controllers, workers, and services
- **Production-like configuration** with environment variable support

## Running Tests

### Using Make (Recommended)
```bash
# Run all integration tests
make test-integration

# Keep cluster for debugging
KEEP_CLUSTER=true make test-integration

# With debug logging
TEST_LOG_LEVEL=debug make test-integration
```

Output is logged to `test/int/last-run.log`.

### Using go test directly
```bash
# Run all integration tests
go test ./test/int/... -v -ginkgo.v -timeout 30m -count=1

# Run specific test suite
go test ./test/int/... -v -ginkgo.v -timeout 30m -count=1 -ginkgo.focus="PostgresAddon"
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TEST_LOG_LEVEL` | Test logging level (debug, info, error) | `info` |
| `TEST_JWT_SECRET` | Custom JWT secret for tests | Auto-generated |
| `TEST_ENCRYPTION_KEY` | Custom encryption key for tests | Auto-generated |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USERNAME` | Database username | `postgres` |
| `DB_PASSWORD` | Database password | `foobar-bizz-buzz` |
| `KEEP_CLUSTER` | Keep infrastructure after tests for debugging | `false` |
| `CLUSTER_AGENT_IMAGE_TAG` | Cluster agent container image tag | `v0.4.6-alpha` |

## Test Structure

```
test/int/
├── README.md                    # This file
├── integration_test.go          # Ginkgo test suite entry point
├── postgres_addon_test.go       # PostgreSQL addon test specifications
└── bootstrap/                  # Infrastructure bootstrap components
    ├── bootstrap.go             # Main bootstrap orchestrator
    ├── cluster.go               # Kubernetes cluster management  
    ├── server.go                # API server lifecycle management
    ├── client.go                # OpenAPI client setup and cluster registration
    └── database.go              # Database setup and migrations
```

## Bootstrap Architecture

The integration tests use a **4-phase bootstrap process**:

### Phase 1: Database Bootstrap (5-10 seconds)
- Creates temporary PostgreSQL database with unique name
- Loads database configuration from environment variables
- Runs all database migrations
- Prepares clean state for testing

### Phase 2: Cluster Bootstrap (2-5 minutes)  
- Creates Kind cluster using testutil package
- Deploys cluster agent operator with all CRDs
- Waits for operator pods to be ready
- Sets up cluster networking and dependencies

### Phase 3: Server Bootstrap (30-60 seconds)
- Initializes **complete test environment** following development pattern:
  - Environment variable loading (.env file support)
  - Structured logging with component prefixes
  - Resource access policy manager (Casbin)
  - **All 18 services** with proper dependency injection
  - **Cluster manager** with all 7 controllers:
    - Volume Controller, Stack Controller, PostgresAddon Controller
    - StackResource Controller, ImageBuild Controller
    - WorkspaceUser Controller, ClusterImageRegistry Controller
  - **Worker manager** with stack worker for async processing
  - Cluster resource services injection
  - RBAC policy initialization
  - Default platform admin user creation
- Starts API server on random port
- Waits for `/health` endpoint to respond

### Phase 4: Client Bootstrap (10-30 seconds)
- Creates test user and organization
- Obtains JWT authentication token
- Configures OpenAPI client with authentication
- Deploys service account to cluster
- Extracts cluster credentials (CA cert + token)
- Registers cluster with API server
- Verifies cluster connectivity

### Shared Test Environment
After bootstrap, all test specs share:
- **Authenticated OpenAPI client** ready for API calls
- **Registered Kubernetes cluster** with running operator
- **Complete API server** with all controllers and services
- **Clean database state** (cleared between test specs)
- **Test organization and cluster IDs** for test isolation

## Key Features

### Complete End-to-End Testing
- **Real Kubernetes Clusters**: Always uses Kind clusters for authentic testing environment
- **Full Operator Deployment**: Installs actual cluster agent with all CRDs and controllers
- **Production Parity**: Test environment mirrors development environment exactly
- **Service Account Authentication**: Uses real Kubernetes RBAC and service account tokens
- **Bidirectional Communication**: Verifies API server ↔ cluster communication and status propagation

### Advanced Infrastructure
- **Ginkgo BDD Framework**: Structured test specifications with shared setup/teardown
- **Shared Infrastructure**: Single bootstrap per test run for efficiency (3-8 minute setup)
- **Environment Configuration**: Full .env file support with environment variable overrides
- **Structured Logging**: Component-specific loggers with configurable levels
- **Resource Cleanup**: Automatic infrastructure teardown with debug preservation option

### Development Experience  
- **Hot Reloading**: Keep infrastructure running with `KEEP_CLUSTER=true` for rapid iteration
- **Debugging Support**: Detailed logging and infrastructure preservation for troubleshooting
- **Flexible Configuration**: Environment-based configuration for different development setups
- **CI/CD Ready**: Reliable, isolated test execution suitable for continuous integration

## Dependencies

### Required
- **Go 1.21+**: For running the test suite
- **PostgreSQL**: Database server for test database creation
- **Docker**: Container runtime for Kind cluster creation
- **Git**: For fetching cluster agent manifests from GitHub

### Auto-Installed by Tests
- **Kind**: Lightweight Kubernetes cluster (installed via testutil if needed)
- **Cluster Agent**: Operator and CRDs (deployed automatically from GitHub)
- **Required Operators**: CloudNativePG, Traefik, Cert-Manager (deployed automatically)

### System Requirements
- **Memory**: 4+ GB RAM recommended for cluster operations
- **Disk**: 2+ GB free space for container images and temporary files
- **Network**: Internet access for fetching container images and manifests

## Troubleshooting

### Common Issues

1. **Bootstrap Timeout Errors**:
   ```bash
   # Check Docker resources
   docker system df
   docker system prune  # If low on space
   
   # Verify PostgreSQL connection
   psql -h localhost -U postgres -c "SELECT 1;"
   
   # Run with debug logging
   TEST_LOG_LEVEL=debug go test ./test/int/... -v
   ```

2. **Cluster Agent Image Issues**:
   ```bash
   # Check image availability
   docker pull stackdome/cluster-agent:v0.4.6-alpha
   
   # Use specific image tag
   export CLUSTER_AGENT_IMAGE_TAG=latest
   go test ./test/int/... -v
   ```

3. **Database Connection Failures**:
   ```bash
   # Set custom database configuration
   export DB_HOST=127.0.0.1
   export DB_USERNAME=myuser
   export DB_PASSWORD=mypass
   go test ./test/int/... -v
   ```

4. **Port Conflicts**:
   - Tests use random ports (8987+ for API server)
   - Check for firewall interference
   - Ensure no other services on port ranges

### Debug Mode

For detailed troubleshooting:
```bash
# Maximum debug output
TEST_LOG_LEVEL=debug KEEP_CLUSTER=true make test-integration

# Keep infrastructure for manual inspection
KEEP_CLUSTER=true make test-integration
# Infrastructure remains running after tests for kubectl debugging
```

## Performance Characteristics

| Phase | Duration | Description |
|-------|----------|-------------|
| Database Bootstrap | 5-10s | Database creation and migrations |
| Cluster Bootstrap | 2-5m | Kind cluster + operator deployment |
| Server Bootstrap | 30-60s | Full API server initialization |
| Client Bootstrap | 10-30s | Authentication and cluster registration |
| **Total Bootstrap** | **3-8m** | **One-time setup per test run** |
| Individual Tests | 1-30s | Per test specification execution |

## Contributing

### Adding New Tests

1. **Test Specifications**: Add new `Describe/It` blocks to existing test files or create new ones
2. **Shared Environment**: Use the shared `env` variable for API client, cluster, and organization IDs
3. **Data Cleanup**: Database is automatically cleared between test specs
4. **Resource Isolation**: Use unique names and namespace the resources to test organization

### Bootstrap Extensions

1. **New Bootstrap Components**: Add to `test/int/bootstrap/` following existing patterns
2. **Environment Variables**: Update `.env_template` and this README
3. **Infrastructure Changes**: Modify bootstrap phases in `bootstrap.go`
4. **Configuration**: Environment-based configuration preferred over hardcoded values

### Example Test Addition
```go
var _ = Describe("New Feature", func() {
    It("should create resource successfully", func() {
        // Use shared environment
        client := env.Client
        orgID := env.OrgID
        clusterID := env.ClusterID
        
        // Test implementation using OpenAPI client
        // Database cleanup automatic between tests
    })
})
```