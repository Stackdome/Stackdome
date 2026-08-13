# Mage Build System

This project uses Mage as its build system. All Mage functionality is consolidated in a single `magefile.go` with organized namespaces for different functionality areas. Mage is a build tool similar to Make but written in Go.

## Installation

First, install Mage:
```bash
go install github.com/magefile/mage@latest
```

## Namespace Organization

The `magefile.go` is organized into logical namespaces:

- **Global Namespace**: Core build commands (build, run, generate, fmt, lint, clean, migrate)
- **Dev**: Local development environment lifecycle (setup, teardown)
- **Deps**: Dependency management (install/clean build tools like kind, helm, kubectl)
- **Cluster**: Test cluster lifecycle (create, delete, status, kubeconfig)
- **Test**: Test execution (unit, integration, coverage, clean)

## Available Commands

### Core Build Commands

- `mage build` - Build the API server binary
- `mage run` - Build and run the API server locally
- `mage clean` - Remove build artifacts
- `mage fmt` - Format the code
- `mage generate` - Regenerate OpenAPI client code

### Dev Environment

- `mage dev:setup` - Bootstrap a complete local dev environment (PostgreSQL + Kind cluster + RBAC)
- `mage dev:teardown` - Tear down the dev environment (removes postgres container, Kind cluster, and config file)

Both commands are idempotent — `dev:setup` reuses an existing PostgreSQL container and Kind cluster, and `dev:teardown` handles already-removed components gracefully.

`dev:setup` reads database configuration from `.env` if present, otherwise uses defaults (`localhost:5432`, user `postgres`, database `stackdome_dev`). Any missing DB variables are automatically appended to `.env` (created from `.env_template` if absent) so that `mage migrate` and `mage run` work immediately. Cluster credentials and DB config are written to `dev_env.yaml`.

```bash
# One-command setup
mage dev:setup

# Then run the API server
mage migrate
mage run

# Clean up when done
mage dev:teardown
```

### Test Commands

- `mage test:unit` - Run unit tests
- `mage test:integration` - Run integration tests
- `mage test:integrationVerbose` - Run integration tests with verbose output
- `mage test:integrationFocus <pattern>` - Run specific integration tests
- `mage test:all` - Run all tests (unit and integration)
- `mage test:coverage` - Run tests with coverage reporting
- `mage test:clean` - Clean test artifacts

### Cluster Management

- `mage cluster:setup` - Create a test cluster with all operators
- `mage cluster:delete` - Delete the test cluster
- `mage cluster:status` - Show test cluster status
- `mage cluster:kubeconfig` - Print the kubeconfig path

### Dependency Management

- `mage deps:install` - Install all required dependencies
- `mage deps:clean` - Remove installed dependencies

## Integration Test Workflow

The new Mage-based integration test workflow improves developer experience by pre-creating and reusing test clusters:

### Quick Start

```bash
# Run integration tests (automatically handles dependencies)
mage test:integration

# Keep cluster running for multiple test runs
KEEP_CLUSTER=true mage test:integration

# Run specific tests
mage test:integrationFocus PostgresAddon

# Clean up when done
mage cluster:delete
```

**Automatic Dependency Management:**
- `mage test:integration` automatically calls `mage cluster:setup`
- `mage cluster:setup` automatically calls `mage deps:install`
- No need to manually run dependency steps

### Environment Variables

- `KEEP_CLUSTER=true` - Keep the cluster after tests complete
- `TEST_LOG_LEVEL=debug` - Enable debug logging
- `CLUSTER_AGENT_IMAGE_TAG=latest` - Override cluster agent version

### Performance Benefits

- **Initial cluster creation**: 3-5 minutes (one-time)
- **Subsequent test runs**: Start immediately (no cluster bootstrap)
- **Parallel test execution**: Supported within shared infrastructure

### How It Works

1. **Pre-created Cluster**: `mage cluster:setup` creates a Kind cluster with all operators installed
2. **State Management**: Cluster state is cached in `~/.cache/stackdome-api-server/clusters/`
3. **Test Integration**: Tests require `TEST_KUBECONFIG` environment variable pointing to the Mage cluster
4. **Lifecycle Control**: Cluster persists across test runs until explicitly deleted

### Comparison: Old vs New

**Old Workflow (per test run)**:
1. Create Kind cluster (2-3 min)
2. Install operators (2-3 min)
3. Deploy CRDs and agent (1 min)
4. Run tests
5. Tear down everything

**New Workflow**:
1. `mage test:integration` (handles all dependencies automatically)
2. Reuse cluster for subsequent runs (immediate start)
3. `mage cluster:delete` (when done)

### CI/CD Compatibility

The system is designed for CI/CD pipelines with automatic dependency management:
```bash
# CI mode - dependencies handled automatically
mage test:integration
mage cluster:delete  # cleanup after tests
```

The dependency chain ensures:
1. `test:integration` → `cluster:setup` → `deps:install`
2. All dependencies are resolved automatically
3. No manual setup steps required

## Advanced Usage

### Running Tests Against Custom Cluster

```bash
# Use your own cluster by setting the kubeconfig path
export TEST_KUBECONFIG=/path/to/kubeconfig
go test ./test/int/...
```

### Debugging Failed Tests

```bash
# Keep cluster for inspection
KEEP_CLUSTER=true mage test:integration

# Check cluster status
mage cluster:status

# Get kubeconfig for manual inspection
export KUBECONFIG=$(mage cluster:kubeconfig)
kubectl get pods -A
```

## Troubleshooting

### Cluster Creation Fails

```bash
# Clean up and retry
mage cluster:delete
mage deps:clean
mage deps:install
mage cluster:setup
```

### Tests Can't Find Cluster

```bash
# Check cluster status
mage cluster:status

# Verify environment
mage test:integrationVerbose
```

### Permission Errors

Ensure Docker/Podman is running and you have necessary permissions.