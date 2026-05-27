# Stackdome API Server

Stackdome is a self-hosted PaaS (Platform-as-a-Service) that deploys and manages workloads across multiple Kubernetes clusters. It provides a Heroku-like developer experience with a REST API, web UI, managed PostgreSQL databases, and git-based builds.

## Architecture

```
                    ┌──────────────┐
                    │   Web UI /   │
                    │   REST API   │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐         ┌──────────────┐
                    │  API Server  │────────►│  PostgreSQL  │
                    │    (Hub)     │         │   Database   │
                    └──────┬───────┘         └──────────────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
       ┌──────▼──┐  ┌──────▼──┐  ┌──────▼──┐
       │ Cluster │  │ Cluster │  │ Cluster │
       │ Agent 1 │  │ Agent 2 │  │ Agent N │
       │ (Spoke) │  │ (Spoke) │  │ (Spoke) │
       └─────────┘  └─────────┘  └─────────┘
```

The system follows a **hub-and-spoke model**:

- **API Server (Hub)** -- Central control plane that manages clusters, stores state in PostgreSQL, and exposes a REST API with web UI.
- **Cluster Agent (Spoke)** -- A Kubernetes operator (built with Kubebuilder) that runs on each managed cluster, reconciling Custom Resources into actual workloads.

**Request flow:** User -> API Server -> Kubernetes CR -> Cluster Agent -> Kubernetes Resources (Deployments, Services, Ingress)

**Status flow:** Cluster Agent updates CR status -> API Server controllers watch changes -> PostgreSQL updated -> UI/API reflects state

## Features

- **Multi-cluster management** -- Register and manage workloads across multiple Kubernetes clusters
- **Stack deployments** -- Deploy multi-service applications from container images or git repositories
- **Managed PostgreSQL** -- Provision PostgreSQL clusters with automated backups, restore, and HA (via CloudNativePG)
- **Environment interpolation** -- Template-based cross-service references (`{{ STACKDOME_POSTGRES_INTERNAL }}`)
- **RBAC** -- Role-based access control with Casbin (viewer, editor, admin roles with org-scoped isolation)
- **Secrets management** -- Encrypted key-value storage for sensitive configuration
- **Container builds** -- Build container images from git repos using Kaniko
- **In-cluster registry** -- Managed container registry (Zot) per cluster
- **Volume management** -- Persistent storage with git, remote dir, and build artifact sources
- **WebSocket log streaming** -- Real-time log access for deployed services

## Prerequisites

- Go 1.25+
- Docker
- kubectl
- [k3d](https://k3d.io) (for local development)
- [mage](https://magefile.org)
- jq

## Quick Start

### One-Command Setup (Recommended)

`mage dev:setup` bootstraps everything you need to develop locally — a PostgreSQL container, a k3d cluster with the stackdome-agent operator, and RBAC credentials for cluster registration:

```bash
mage dev:setup
```

This will:
1. Start a PostgreSQL container (reads config from `.env` if present, otherwise uses defaults)
2. Create a k3d cluster and install the stackdome-agent Helm chart
3. Deploy RBAC resources (ServiceAccount, ClusterRole, ClusterRoleBinding) for API server access
4. Extract cluster credentials (API URL, CA data, SA token) and write them to `dev_env.yaml`

Then run the API server:

```bash
mage migrate    # Run database migrations
mage run        # Build and start the API server
```

To tear everything down:

```bash
mage dev:teardown
```

The command is fully idempotent — safe to run multiple times. It will reuse an existing PostgreSQL container and k3d cluster instead of recreating them.

**Database configuration:** `mage dev:setup` loads the `.env` file if present, then falls back to these defaults. Any missing DB variables are automatically appended to `.env` so that `mage migrate` and `mage run` work without manual editing. If `.env` doesn't exist at all, it is created from `.env_template`.

| Variable | Default |
|----------|---------|
| `DB_HOST` | `localhost` |
| `DB_PORT` | `5432` |
| `DB_NAME` | `stackdome_dev` |
| `DB_USERNAME` | `postgres` |
| `DB_PASSWORD` | `foobar-bizz-buzz` |

### Alternative: Full Automated Setup

The `hack/run_local.sh` script bootstraps a complete local environment including the API server and cluster registration (useful for end-to-end demos):

```bash
# Start environment only
./hack/run_local.sh

# Deploy a stack
./hack/run_local.sh samples/tooljet.json

# Deploy a stack with a postgres addon
ADDON_FILE=samples/tooljet_addon_postgres.json ./hack/run_local.sh samples/tooljet_with_addon.json
```

Press `Ctrl+C` to tear down all resources.

### Manual Setup

See [docs/QUICK_START.md](docs/QUICK_START.md) for a 5-minute setup, or [docs/BOOTSTRAP_GUIDE.md](docs/BOOTSTRAP_GUIDE.md) for a comprehensive walkthrough.

**TL;DR:**

```bash
# 1. Start PostgreSQL
docker run -d --name stackdome-postgres \
  -e POSTGRES_PASSWORD=mypassword -e POSTGRES_DB=stackdome \
  -p 5432:5432 postgres:15-alpine

# 2. Configure environment
cat > .env << 'EOF'
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -hex 32)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=stackdome
DB_USERNAME=postgres
DB_PASSWORD=mypassword
DEFAULT_USER_EMAIL=admin@stackdome.local
DEFAULT_USER_NAME=admin
DEFAULT_USER_PASS=admin123
EOF

# 3. Build, migrate, and run
make binary
./bin/stackdome-server migrate
./bin/stackdome-server serve
```

## Project Structure

```
├── cmd/                          # Application entrypoints
│   ├── main.go                   # Binary entrypoint
│   ├── servecmd/                 # serve command
│   ├── migratecmd/               # migrate command
│   ├── server/                   # HTTP server, routes, middleware
│   └── environment/              # Environment initialization
├── pkg/
│   ├── handlers/                 # HTTP request handlers
│   ├── services/                 # Business logic layer
│   ├── models/                   # Domain models
│   ├── stores/                   # Database interface definitions
│   │   └── pgstore/              # PostgreSQL implementations
│   ├── presenters/               # API <-> model conversion
│   ├── controllers/              # Controller-runtime watchers (CR status sync)
│   ├── worker/                   # Async background processing
│   │   ├── workermanager/        # Worker coordination
│   │   └── stack/                # Stack reconciliation worker
│   ├── resourceaccess/           # Casbin-based RBAC authorization
│   ├── auth/                     # JWT and cookie authentication
│   ├── validator/                # Input validation per resource type
│   ├── builders/                 # Kubernetes resource construction
│   ├── clients/                  # External service clients (git, k8s, registry)
│   ├── clustermanager/           # Multi-cluster connection management
│   ├── db/                       # Database migrations and sessions
│   │   └── migrations/           # Ordered migration files
│   ├── api/openapi/              # Generated OpenAPI client code
│   └── testutil/                 # Test cluster utilities
├── config/
│   └── openapi/                  # OpenAPI 3.0 specification
│       └── stackdome_api.yaml
├── frontend/                     # React + TypeScript web UI
├── test/
│   └── int/                      # Ginkgo integration test suite
├── samples/                      # Example JSON payloads
├── hack/                         # Development scripts
│   ├── run_local.sh              # Full local environment bootstrap
│   └── create_migration.sh       # Database migration scaffolding
├── magefile.go                   # Mage build targets
└── Makefile                      # Common build tasks
```

## Development

### Build

```bash
make binary                  # Build API server binary
make generate                # Regenerate OpenAPI client code
```

### Mage Targets

```bash
mage build                   # Build the API server
mage run                     # Build and run
mage fmt                     # Format code
mage lint                    # Run linter
mage migrate                 # Run database migrations

# Dev environment (recommended for onboarding)
mage dev:setup               # Bootstrap full dev environment (postgres + cluster + RBAC)
mage dev:teardown            # Tear down dev environment

# Cluster management
mage cluster:setup           # Create k3d cluster with stackdome-agent chart
mage cluster:delete          # Delete the cluster
mage cluster:status          # Show cluster status

# Testing
mage test:unit               # Run unit tests
mage test:integration        # Run integration tests (creates cluster)
mage test:all                # Run all tests
mage test:coverage           # Tests with coverage report
```

### Database Migrations

```bash
# Create a new migration
./hack/create_migration.sh <migration_name>

# Migrations are in pkg/db/migrations/ and run automatically on startup
```

### Running Tests

```bash
# Unit tests
mage test:unit

# Integration tests (requires Docker, creates Kind cluster)
make test-integration

# Integration tests with debug logging
TEST_LOG_LEVEL=debug make test-integration

# Keep cluster after tests for debugging
KEEP_CLUSTER=true make test-integration

# Run specific test suite
go test ./test/int/... -v -ginkgo.v -timeout 30m -count=1 -ginkgo.focus="PostgresAddon"
```

The integration test suite uses real Kind clusters with the full operator stack -- no mocks. Bootstrap takes 3-8 minutes (one-time), individual tests run in 1-30 seconds.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_SECRET` | Secret for signing JWT tokens | (required) |
| `ENCRYPTION_KEY` | Key for encrypting secrets (64 hex chars) | (required) |
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_NAME` | Database name | stackdome |
| `DB_USERNAME` | Database user | postgres |
| `DB_PASSWORD` | Database password | (required) |
| `DB_DEBUG_MODE` | Enable SQL query logging | false |
| `LOG_LEVEL` | Log level (debug, info, warn, error) | info |
| `DEFAULT_USER_EMAIL` | Initial admin user email | (required) |
| `DEFAULT_USER_NAME` | Initial admin user name | (required) |
| `DEFAULT_USER_PASS` | Initial admin user password | (required) |

## API

The API server exposes a REST API documented via OpenAPI 3.0 at `config/openapi/stackdome_api.yaml`.

```bash
# Health check
curl http://localhost:8000/health

# Authenticate
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@stackdome.local","password":"admin123"}' | jq -r '.token')

# List stacks
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8000/api/v1/organizations/{org_id}/stacks

# Create a stack
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8000/api/v1/organizations/{org_id}/stacks \
  -d @samples/tooljet.json

# Create a postgres addon
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8000/api/v1/organizations/{org_id}/addons/postgres \
  -d @samples/postgres_addon_basic.json
```

## Custom Resources

The cluster agent reconciles these CRDs (defined in the [cluster-agent](https://github.com/Stackdome/cluster-agent) repo):

| CRD | API Group | Description |
|-----|-----------|-------------|
| Stack | core.stackdome.io | Multi-service deployment unit |
| StackResource | core.stackdome.io | Individual service within a stack |
| PostgresCluster | addons.stackdome.io | Managed PostgreSQL (wraps CloudNativePG) |
| Volume | storage.stackdome.io | Persistent storage with source sync |
| Storage | storage.stackdome.io | Stack-level storage aggregation |
| NFSServer | storage.stackdome.io | In-cluster NFS for shared volumes |
| ImageBuild | builds.stackdome.io | Kaniko-based container image builds |
| ClusterRegistry | registry.stackdome.io | In-cluster container registry (Zot) |
| WorkspaceUser | users.stackdome.io | User SSH access for workspaces |

## Documentation

- [Quick Start](docs/QUICK_START.md) -- 5-minute setup guide
- [Bootstrap Guide](docs/BOOTSTRAP_GUIDE.md) -- Complete walkthrough with troubleshooting
- [Casbin RBAC](docs/casbin-explained.md) -- How authorization policies work
- [Env Interpolation](docs/env-interpolation.md) -- Cross-service environment variable templates

## Sample Configurations

The `samples/` directory contains example payloads:

- `tooljet.json` -- Basic stack deployment
- `tooljet_with_build.json` -- Stack with git-based build
- `tooljet_with_addon.json` -- Stack referencing a postgres addon
- `postgres_addon_basic.json` -- Single-instance PostgreSQL
- `postgres_addon_with_backup.json` -- PostgreSQL with automated backups
- `postgres_addon_restore_from_backup.json` -- Restore from a backup
- `postgres_addon_with_placement.json` -- HA with anti-affinity constraints
- `postgres_addon_dev_environment.json` -- Minimal dev setup
- `postgres_addon_multi_az.json` -- Multi-AZ deployment

## Technology Stack

**Backend:** Go, PostgreSQL, Gorilla Mux, GORM, Casbin, Controller-Runtime, JWT

**Frontend:** React, TypeScript, Vite, Tailwind CSS

**Infrastructure:** Kubernetes, CloudNativePG, Traefik, Cert-Manager, Kaniko, Zot
