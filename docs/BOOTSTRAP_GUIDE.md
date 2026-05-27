# Stackdome Platform Bootstrap Guide

Complete guide to bootstrapping the entire Stackdome platform from scratch, including database, API server, cluster agent, and your first deployment.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Phase 1: Database Setup](#phase-1-database-setup)
- [Phase 2: Cluster Agent Setup](#phase-2-cluster-agent-setup)
- [Phase 3: API Server Setup](#phase-3-api-server-setup)
- [Phase 4: Cluster Registration](#phase-4-cluster-registration)
- [Phase 5: First Deployment](#phase-5-first-deployment)
- [Troubleshooting](#troubleshooting)
- [Architecture Details](#architecture-details)

## Overview

Stackdome is a self-hosted PaaS (Platform-as-a-Service) that manages workloads across multiple Kubernetes clusters with a Heroku-like developer experience.

**What you'll have after this guide:**
- PostgreSQL database for API server state
- API server running with REST API and web UI
- Kubernetes cluster with cluster-agent operator
- A sample application deployed and running

**Time estimate:** 30-45 minutes

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Stackdome Platform                 │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌──────────────┐         ┌──────────────────┐    │
│  │  PostgreSQL  │◄────────│   API Server     │    │
│  │   Database   │         │  (Hub)           │    │
│  └──────────────┘         │  - REST API      │    │
│                           │  - Web UI        │    │
│                           │  - Controllers   │    │
│                           │  - Workers       │    │
│                           └────────┬─────────┘    │
│                                    │               │
│                           ┌────────▼─────────┐    │
│                           │  Kubernetes API  │    │
│                           └────────┬─────────┘    │
│                                    │               │
│                           ┌────────▼─────────┐    │
│                           │  Cluster Agent   │    │
│                           │  (Spoke)         │    │
│                           │  - Operator      │    │
│                           │  - Reconcilers   │    │
│                           └──────────────────┘    │
│                                                     │
└─────────────────────────────────────────────────────┘

Flow: User → API Server → K8s CR → Cluster Agent → K8s Resources
```

**Components:**

1. **PostgreSQL**: Stores all resource definitions (stacks, users, organizations, clusters)
2. **API Server**: Central control plane exposing REST API and web UI
3. **Cluster Agent**: Kubernetes operator that reconciles Custom Resources (CRs) to actual workloads
4. **Kubernetes Cluster**: Where your applications actually run

## Prerequisites

### Required Software

- **Go**: 1.24 or later
  ```bash
  go version  # Should show 1.24+
  ```

- **Docker**: For PostgreSQL and container images
  ```bash
  docker --version
  docker ps  # Verify Docker daemon is running
  ```

- **kubectl**: Kubernetes CLI
  ```bash
  kubectl version --client
  ```

- **k3d**: Lightweight Kubernetes in Docker (https://k3d.io)
  ```bash
  k3d version
  # If not installed: see https://k3d.io/#installation
  ```

- **Mage**: Build tool (for cluster-agent)
  ```bash
  mage -version
  # If not installed: go install github.com/magefile/mage@latest
  ```

- **jq**: JSON processor
  ```bash
  jq --version
  # macOS: brew install jq
  # Linux: apt-get install jq
  ```

### System Requirements

- **CPU**: 4+ cores recommended
- **RAM**: 8GB minimum, 16GB recommended
- **Disk**: 20GB free space
- **Ports available**: 5432 (PostgreSQL), 8000 (API server)

### Repository Setup

Clone both repositories:

```bash
# Create working directory
mkdir -p ~/stackdome && cd ~/stackdome

# Clone API server
git clone https://github.com/yourusername/stackdome-api-server.git
cd stackdome-api-server

# Clone cluster agent (in parent directory)
cd ..
git clone https://github.com/yourusername/stackdome-cluster-agent.git
cd stackdome-cluster-agent
```

**Directory structure:**
```
~/stackdome/
├── stackdome-api-server/      # API server repository
└── stackdome-cluster-agent/   # Cluster agent repository
```

---

## Phase 1: Database Setup

### Option A: PostgreSQL via Docker (Recommended for Development)

**Advantages:** Quick setup, isolated, easy cleanup

```bash
# Start PostgreSQL container
docker run -d \
  --name stackdome-postgres \
  -e POSTGRES_PASSWORD=your-secure-password \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=stackdome \
  -p 5432:5432 \
  --restart unless-stopped \
  postgres:15-alpine

# Verify it's running
docker ps | grep stackdome-postgres

# Test connection
docker exec stackdome-postgres psql -U postgres -c "SELECT version();"
```

**Database URL:** `postgresql://postgres:your-secure-password@localhost:5432/stackdome`

### Option B: PostgreSQL via System Package

**For production or if you prefer native installation:**

<details>
<summary>macOS (Homebrew)</summary>

```bash
# Install PostgreSQL
brew install postgresql@15
brew services start postgresql@15

# Create database
createdb stackdome

# Create user (optional)
psql postgres -c "CREATE USER stackdome WITH PASSWORD 'your-password';"
psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE stackdome TO stackdome;"
```
</details>

<details>
<summary>Ubuntu/Debian</summary>

```bash
# Install PostgreSQL
sudo apt update
sudo apt install postgresql-15 postgresql-contrib

# Start service
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres psql << EOF
CREATE DATABASE stackdome;
CREATE USER stackdome WITH PASSWORD 'your-password';
GRANT ALL PRIVILEGES ON DATABASE stackdome TO stackdome;
\q
EOF
```
</details>

### Verify Database

```bash
# Test connection (adjust credentials as needed)
psql -h localhost -U postgres -d stackdome -c "\l"

# You should see 'stackdome' in the database list
```

---

## Phase 2: Cluster Agent Setup

The cluster agent runs inside your Kubernetes cluster and reconciles Stack Custom Resources.

### 2.1: Create Development Cluster

Create a k3d cluster with port mappings for ingress and two agent nodes:

```bash
k3d cluster create stackdome-dev \
  --port "80:80@loadbalancer" \
  --port "443:443@loadbalancer" \
  --k3s-arg "--disable=traefik@server:0" \
  --k3s-arg "--disable=servicelb@server:0" \
  --agents 2
```

This creates a 3-node k3d cluster (1 server, 2 agents) with host ports 80/443 forwarded to the load balancer. The built-in Traefik and ServiceLB are disabled so we can install the versions the stackdome-agent requires.

**Time:** 1-2 minutes

### 2.2: Install Required Operators and Cluster Agent

Install the dependencies and the stackdome-agent helm chart into the k3d cluster:

```bash
# Switch to the k3d cluster context
kubectl config use-context k3d-stackdome-dev

# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
kubectl wait --for=condition=Available deployment --all -n cert-manager --timeout=120s

# Install CloudNativePG operator (for PostgreSQL addons)
helm repo add cnpg https://cloudnative-pg.github.io/charts
helm repo update
helm install cnpg cnpg/cloudnative-pg -n cnpg-system --create-namespace --wait

# Install Traefik ingress controller
helm repo add traefik https://traefik.github.io/charts
helm install traefik traefik/traefik -n traefik-v2 --create-namespace --wait

# Install stackdome-agent (cluster-agent operator + CRDs)
# Replace <version> with the desired release version
helm install stackdome-agent oci://ghcr.io/stackdome/charts/stackdome-agent \
  -n stackdome-control-plane --create-namespace --wait
```

> **Note:** The cluster-agent repo has its own `mage dev:setup` that creates a separate Kind cluster for operator development. That workflow is independent and unaffected by using k3d here.

### 2.3: Verify Cluster Agent

```bash
# Check cluster nodes
kubectl get nodes
# Should show 3 nodes: 1 server, 2 agents

# Check cluster-agent operator
kubectl get deployment -n stackdome-control-plane
# Should show: stackdome-operator-manager

# Check installed operators
kubectl get pods -n cnpg-system
# Should show: cnpg-cloudnative-pg

kubectl get pods -n cert-manager
# Should show: cert-manager, cert-manager-cainjector, cert-manager-webhook

kubectl get pods -n traefik-v2
# Should show: traefik pods

# Verify CRDs are installed
kubectl get crds | grep stackdome.io
# Should show: stacks.core.stackdome.io, volumes.storage.stackdome.io, etc.
```

### 2.4: Save Kubeconfig Path

```bash
# Export k3d kubeconfig for the stackdome-dev cluster
export KUBECONFIG=$(k3d kubeconfig get stackdome-dev)

# Add to your shell profile for persistence
echo 'export STACKDOME_KUBECONFIG=$(k3d kubeconfig get stackdome-dev)' >> ~/.zshrc
# or ~/.bashrc depending on your shell
```

---

## Phase 3: API Server Setup

### 3.1: Configure Environment

```bash
cd ~/stackdome/stackdome-api-server

# Copy environment template
cp .env_template .env

# Edit .env file
nano .env  # or use your preferred editor
```

**Required .env configuration:**

```bash
# Application Config
JWT_SECRET="your-random-secret-key-min-32-chars"
ENCRYPTION_KEY="your-random-encryption-key-64-hex-chars"
LOG_LEVEL="info"

# Database Config
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="stackdome"
DB_USERNAME="postgres"
DB_PASSWORD="your-secure-password"
DB_DEBUG_MODE="false"

# Default User (created on first start)
DEFAULT_USER_EMAIL="admin@stackdome.local"
DEFAULT_USER_NAME="admin"
DEFAULT_USER_PASS="changeme123"

# Cluster Config (will be configured later)
DEFAULT_CLUSTER_NAME=""
DEFAULT_CLUSTER_API_URL=""
DEFAULT_CLUSTER_CA_DATA=""
DEFAULT_CLUSTER_TOKEN=""
```

**Generate secure secrets:**

```bash
# Generate JWT secret (32+ characters)
openssl rand -base64 32

# Generate encryption key (64 hex characters)
openssl rand -hex 32
```

### 3.2: Build API Server

```bash
# Build binary
go build -o bin/stackdome-server ./cmd/main.go

# Verify build
./bin/stackdome-server --help
# Should show: Available Commands: migrate, serve
```

### 3.3: Run Database Migrations

```bash
./bin/stackdome-server migrate

# Expected output:
# I0307 23:04:18 Environment variable "STACKDOME_ENV" not specified, using default "DEVELOPMENT"
# Migrations completed successfully
```

**What migrations do:**
- Create users, organizations, clusters tables
- Create stacks, resources, volumes tables
- Create secrets, image registries tables
- Create postgres addons, object stores tables
- Set up indexes and foreign keys

### 3.4: Start API Server

**Option A: Foreground (for debugging)**

```bash
./bin/stackdome-server serve

# Expected output:
# time=2026-03-07 23:04:44 level=info msg=Creating user with email: admin@stackdome.local
# time=2026-03-07 23:04:44 level=info msg=Created default platform admin user
# time=2026-03-07 23:04:44 level=info msg=stack-worker worker started
# I0307 23:04:44 Serving without TLS at 0.0.0.0:8000
```

**Option B: Background (for long-running)**

```bash
nohup ./bin/stackdome-server serve > api-server.log 2>&1 &
echo $! > api-server.pid

# Check logs
tail -f api-server.log

# Stop server later
kill $(cat api-server.pid)
```

### 3.5: Verify API Server

```bash
# Health check
curl http://localhost:8000/health
# Should return: OK

# Login to get JWT token
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@stackdome.local",
    "password": "changeme123"
  }' | jq .

# Should return:
# {
#   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#   "user": {
#     "email": "admin@stackdome.local",
#     "id": "...",
#     "name": "admin",
#     "organisation": "DefaultOrganisation",
#     "organisation_id": "...",
#     "role": "PlatformAdmin"
#   }
# }
```

**Save token and org ID for next steps:**

```bash
# Get and save token
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@stackdome.local", "password": "changeme123"}' \
  | jq -r '.token')

echo $TOKEN > /tmp/stackdome-token.txt

# Get and save organization ID
ORG_ID=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@stackdome.local", "password": "changeme123"}' \
  | jq -r '.user.organisation_id')

echo $ORG_ID > /tmp/stackdome-org-id.txt

echo "Token saved to: /tmp/stackdome-token.txt"
echo "Org ID saved to: /tmp/stackdome-org-id.txt"
```

---

## Phase 4: Cluster Registration

Connect your Kubernetes cluster to the API server so it can deploy workloads.

### 4.1: Create Service Account in Cluster

The API server needs credentials to communicate with the Kubernetes cluster.

```bash
cd ~/stackdome/stackdome-api-server

# Set kubeconfig to the k3d cluster
export KUBECONFIG=$(k3d kubeconfig get stackdome-dev)

# Deploy service account
kubectl apply -f deploy/sa.yaml

# Expected output:
# namespace/stackdome-control-plane configured
# serviceaccount/stackdome-api-server-account created
# clusterrole.rbac.authorization.k8s.io/stackdome-api-server-role created
# clusterrolebinding.rbac.authorization.k8s.io/stackdome-api-server-role-binding created
# secret/stackdome-api-server-account-secret created
```

**What this creates:**
- Namespace: `stackdome-control-plane` (if not exists)
- ServiceAccount: `stackdome-api-server-account` with cluster-admin permissions
- Secret: `stackdome-api-server-account-secret` containing token

### 4.2: Extract Cluster Credentials

```bash
# Wait for secret to be populated
sleep 3

# Extract service account token
SA_TOKEN=$(kubectl get secret -n stackdome-control-plane \
  stackdome-api-server-account-secret \
  -o jsonpath='{.data.token}' | base64 -d | base64)

# Extract CA certificate
SA_CA=$(kubectl get secret -n stackdome-control-plane \
  stackdome-api-server-account-secret \
  -o jsonpath='{.data.ca\.crt}')

# Get cluster API URL
CLUSTER_URL=$(kubectl cluster-info | \
  grep "Kubernetes control plane" | \
  awk '{print $NF}' | \
  sed 's/\x1b\[[0-9;]*m//g')

echo "Cluster URL: $CLUSTER_URL"
echo "Token length: ${#SA_TOKEN}"
echo "CA data length: ${#SA_CA}"
```

### 4.3: Register Cluster with API Server

```bash
# Load saved credentials
TOKEN=$(cat /tmp/stackdome-token.txt)
ORG_ID=$(cat /tmp/stackdome-org-id.txt)

# Create cluster registration JSON
cat > /tmp/cluster-register.json << EOF
{
  "name": "dev-cluster",
  "cluster_url": "${CLUSTER_URL}",
  "cluster_ca_data": "${SA_CA}",
  "cluster_sa_token": "${SA_TOKEN}",
  "default": true
}
EOF

# Register cluster
curl -X POST "http://localhost:8000/api/v1/organizations/${ORG_ID}/clusters" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @/tmp/cluster-register.json | jq .

# Expected output:
# {
#   "id": "...",
#   "name": "dev-cluster",
#   "cluster_url": "https://127.0.0.1:xxxxx",
#   "organisation_id": "...",
#   "default": true
# }
```

### 4.4: Verify Cluster Registration

```bash
# List registered clusters
curl -s "http://localhost:8000/api/v1/organizations/${ORG_ID}/clusters" \
  -H "Authorization: Bearer $TOKEN" | jq .

# Should show your registered cluster
```

---

## Phase 5: First Deployment

Deploy a sample nginx application to verify the complete workflow.

### 5.1: Configure Organization Domain

Stacks with public ports require a domain to be configured.

For local development, use `stackdome.127.0.0.1.nip.io` as the organization domain. The nip.io service resolves any `<anything>.stackdome.127.0.0.1.nip.io` address to `127.0.0.1` via public DNS, so no `/etc/hosts` editing is needed.

```bash
TOKEN=$(cat /tmp/stackdome-token.txt)
ORG_ID=$(cat /tmp/stackdome-org-id.txt)

# Create organization update JSON
cat > /tmp/org-update.json << EOF
{
  "name": "DefaultOrganisation",
  "domains": [
    {
      "fqdn": "stackdome.127.0.0.1.nip.io"
    }
  ]
}
EOF

# Update organization
curl -X PUT "http://localhost:8000/api/v1/organizations/${ORG_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @/tmp/org-update.json | jq .

# Should return organization with domains array populated
```

### 5.2: Create Stack

```bash
# Create stack definition
cat > /tmp/nginx-stack.json << 'EOF'
{
  "name": "nginx-demo",
  "labels": [
    {
      "key": "environment",
      "value": "development"
    },
    {
      "key": "app",
      "value": "nginx"
    }
  ],
  "spec": {
    "stack_resources": [
      {
        "name": "web",
        "image_spec": {
          "image": "nginx:latest"
        },
        "ports": [
          {
            "number": 80,
            "exposed_to_public": true,
            "protocol": "HTTP"
          }
        ]
      }
    ]
  }
}
EOF

# Deploy stack
curl -X POST "http://localhost:8000/api/v1/organizations/${ORG_ID}/stacks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @/tmp/nginx-stack.json | jq .

# Save stack ID
STACK_ID=$(curl -s -X POST "http://localhost:8000/api/v1/organizations/${ORG_ID}/stacks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @/tmp/nginx-stack.json | jq -r '.id')

echo $STACK_ID > /tmp/stackdome-stack-id.txt
echo "Stack ID: $STACK_ID"
```

### 5.3: Monitor Deployment

```bash
TOKEN=$(cat /tmp/stackdome-token.txt)
ORG_ID=$(cat /tmp/stackdome-org-id.txt)
STACK_ID=$(cat /tmp/stackdome-stack-id.txt)

# Check stack status (poll every 5 seconds)
watch -n 5 "curl -s http://localhost:8000/api/v1/organizations/${ORG_ID}/stacks/${STACK_ID} \
  -H 'Authorization: Bearer $TOKEN' | jq '.status'"

# Or check once
curl -s "http://localhost:8000/api/v1/organizations/${ORG_ID}/stacks/${STACK_ID}" \
  -H "Authorization: Bearer $TOKEN" | jq '.status'

# Expected progression:
# State: "Pending" → "Ready"
# Conditions: [] → [{"type": "Available", "status": "True", ...}]
```

### 5.4: Verify in Kubernetes

```bash
# Set kubeconfig
export KUBECONFIG=$(k3d kubeconfig get stackdome-dev)

# List Stack CRs
kubectl get stacks -A

# Expected output:
# NAMESPACE                                  NAME         PHASE
# nginx-demo-xxxxx-xxxxx-xxxxx-xxxxx-xxxxx   nginx-demo   Ready

# Get stack namespace
STACK_NS=$(kubectl get stacks -A -o jsonpath='{.items[0].metadata.namespace}')

# View all resources
kubectl get all -n $STACK_NS

# Expected output:
# NAME                       READY   STATUS    RESTARTS   AGE
# pod/web-xxxxxxxxxx-xxxxx   1/1     Running   0          30s
#
# NAME          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
# service/web   ClusterIP   10.96.xxx.xxx   <none>        80/TCP    30s
#
# NAME                  READY   UP-TO-DATE   AVAILABLE   AGE
# deployment.apps/web   1/1     1            1           30s

# Check ingress
kubectl get ingress -n $STACK_NS

# Check pod logs
kubectl logs -n $STACK_NS deployment/web --tail=20
```

### 5.5: Access the Application

```bash
# Get ingress host
INGRESS_HOST=$(kubectl get ingress -n $STACK_NS -o jsonpath='{.items[0].spec.rules[0].host}')
echo "Ingress Host: $INGRESS_HOST"

# Port-forward to access locally
kubectl port-forward -n $STACK_NS svc/web 8080:80 &

# Test nginx
curl http://localhost:8080

# Should return nginx default page HTML

# Stop port-forward when done
pkill -f "port-forward.*svc/web"
```

---

## Troubleshooting

### Database Issues

**Problem: Connection refused to PostgreSQL**

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check logs
docker logs stackdome-postgres

# Restart container
docker restart stackdome-postgres

# Verify connection
docker exec stackdome-postgres psql -U postgres -c "SELECT 1"
```

**Problem: Migration fails**

```bash
# Check database connectivity
psql -h localhost -U postgres -d stackdome -c "\dt"

# Re-run migrations
./bin/stackdome-server migrate

# Check API server logs for detailed error
```

### Cluster Issues

**Problem: k3d cluster create fails**

```bash
# Delete existing cluster and retry
k3d cluster delete stackdome-dev

# Verify Docker is running
docker ps

# Check available resources
docker stats --no-stream

# Retry cluster creation
k3d cluster create stackdome-dev \
  --port "80:80@loadbalancer" \
  --port "443:443@loadbalancer" \
  --k3s-arg "--disable=traefik@server:0" \
  --k3s-arg "--disable=servicelb@server:0" \
  --agents 2
```

**Problem: Port 80/443 already in use**

```bash
# Check what is using the ports
lsof -i :80
lsof -i :443

# Stop the conflicting process, or use different host ports:
k3d cluster create stackdome-dev \
  --port "8080:80@loadbalancer" \
  --port "8443:443@loadbalancer" \
  --k3s-arg "--disable=traefik@server:0" \
  --k3s-arg "--disable=servicelb@server:0" \
  --agents 2
```

**Problem: CRDs not installed**

```bash
export KUBECONFIG=$(k3d kubeconfig get stackdome-dev)

# Check CRDs
kubectl get crds | grep stackdome

# If missing, re-install the stackdome-agent helm chart
helm install stackdome-agent oci://ghcr.io/stackdome/charts/stackdome-agent \
  -n stackdome-control-plane --create-namespace --wait
```

**Problem: Operator pod not running**

```bash
# Check operator pod
kubectl get pods -n stackdome-control-plane

# Check logs
kubectl logs -n stackdome-control-plane deployment/stackdome-operator-manager

# Check events
kubectl get events -n stackdome-control-plane --sort-by='.lastTimestamp'
```

### API Server Issues

**Problem: Binary won't build**

```bash
# If cd command is broken (gvm issue)
unset -f cd

# Build from absolute path
go build -o bin/stackdome-server ./cmd/main.go

# Check Go version
go version  # Must be 1.24+
```

**Problem: Authentication fails**

```bash
# Verify JWT_SECRET in .env is set and matches
grep JWT_SECRET .env

# Get fresh token
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@stackdome.local", "password": "changeme123"}'

# Check API server logs
tail -f api-server.log | grep -i auth
```

**Problem: Cluster registration fails**

```bash
# Verify service account exists
kubectl get sa -n stackdome-control-plane

# Verify secret exists
kubectl get secret -n stackdome-control-plane stackdome-api-server-account-secret

# Check if token is valid
kubectl get secret -n stackdome-control-plane \
  stackdome-api-server-account-secret \
  -o jsonpath='{.data.token}' | base64 -d | wc -c
# Should be > 800 characters

# Check cluster URL is reachable from API server
curl -k $(kubectl cluster-info | grep "Kubernetes control plane" | awk '{print $NF}')
```

### Stack Deployment Issues

**Problem: Stack stuck in Pending**

```bash
# Check stack status from API
curl -s "http://localhost:8000/api/v1/organizations/${ORG_ID}/stacks/${STACK_ID}" \
  -H "Authorization: Bearer $TOKEN" | jq '.status'

# Check if Stack CR was created in Kubernetes
kubectl get stacks -A

# Check operator logs
kubectl logs -n stackdome-control-plane deployment/stackdome-operator-manager --tail=50

# Check stack events
STACK_NS=$(kubectl get stacks -A -o jsonpath='{.items[0].metadata.namespace}')
kubectl get events -n $STACK_NS --sort-by='.lastTimestamp'
```

**Problem: Pod won't start**

```bash
# Check pod status
kubectl get pods -n $STACK_NS

# Describe pod
kubectl describe pod -n $STACK_NS <pod-name>

# Check logs
kubectl logs -n $STACK_NS <pod-name>

# Check if image can be pulled
kubectl run test-nginx --image=nginx:latest --rm -it --restart=Never
```

**Problem: Ingress not working**

```bash
# Check if Traefik is running
kubectl get pods -n traefik-v2

# Check ingress resource
kubectl get ingress -n $STACK_NS -o yaml

# Check Traefik logs
kubectl logs -n traefik-v2 deployment/traefik
```

---

## Architecture Details

### Request Flow

**1. User creates a Stack via API:**
```
POST /api/v1/organizations/{org_id}/stacks
```

**2. API Server processes request:**
- Validates stack definition (ports, images, resources)
- Stores stack in PostgreSQL database
- Creates Stack Custom Resource in Kubernetes cluster

**3. Kubernetes notifies Cluster Agent:**
- Watch event triggers Stack reconciler
- Reconciler reads Stack CR specification

**4. Cluster Agent creates resources:**
- Namespace (if doesn't exist)
- Deployment (with pod spec from stack resource)
- Service (ClusterIP for internal communication)
- Ingress (if ports are exposed_to_public)
- Secrets (if environment variables reference secrets)
- ConfigMaps (if needed)

**5. Kubernetes schedules pods:**
- Pulls container images
- Starts containers
- Assigns cluster IPs

**6. Status propagation:**
- Cluster Agent updates Stack CR status
- API Server controller watches Stack CR
- Status written back to PostgreSQL
- User sees updated status in UI/API

### Custom Resource Definitions (CRDs)

**Core CRDs:**

1. **Stack** (`stacks.core.stackdome.io`):
   - Main deployment unit containing multiple resources
   - Spec: stack_resources[], volumes[], domains[]
   - Status: state, conditions, observed_revision

2. **StackResource** (embedded in Stack):
   - Individual service/component
   - Types: image-based, build-from-git
   - Ports, environment variables, volume mounts

3. **Volume** (`volumes.storage.stackdome.io`):
   - Persistent storage definitions
   - NFS, hostPath, or PVC-backed

4. **PostgresCluster** (`postgresclusters.addons.stackdome.io`):
   - Managed PostgreSQL databases
   - Uses CloudNativePG operator underneath
   - Supports backups, restore, HA

5. **ObjectStore** (`objectstores.barmancloud.cnpg.io`):
   - Backup storage configuration (S3, Azure, GCS)
   - Used by PostgresCluster for backups

### Database Schema

**Key tables:**

- `users`: Platform users with authentication
- `organisations`: Tenants, each with domains and clusters
- `clusters`: Registered Kubernetes clusters
- `stacks`: Stack definitions and status
- `stack_resources`: Individual resources within stacks
- `secrets`: Encrypted key-value pairs
- `volumes`: Persistent storage definitions
- `postgres_addons`: PostgreSQL cluster definitions
- `object_stores`: Backup storage configurations

**Relationships:**
```
organisations (1) ──< (many) users
organisations (1) ──< (many) clusters
organisations (1) ──< (many) stacks
stacks (1) ──< (many) stack_resources
organisations (1) ──< (many) postgres_addons
postgres_addons (many) >── (1) object_stores
```

### Controllers & Workers

**API Server Controllers:**

1. **Stack Controller**: Watches Stack CRs, syncs status to database
2. **PostgresAddon Controller**: Watches PostgresCluster CRs, syncs status
3. **Volume Controller**: Watches Volume CRs, syncs status

**Background Workers:**

1. **Stack Worker**: Reconciles Stack database records → Kubernetes CRs
2. **Queue Manager**: Coordinates async job processing

**Cluster Agent Reconcilers:**

1. **Stack Reconciler**: Stack CR → Deployment, Service, Ingress
2. **PostgresCluster Reconciler**: PostgresCluster CR → CNPG Cluster
3. **Volume Reconciler**: Volume CR → PVC or NFS resources

---

## Next Steps

**Production Deployment:**
- Read [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) for production setup
- Configure TLS certificates for API server
- Set up external PostgreSQL with replication
- Configure ingress with real domain names
- Set up monitoring (Prometheus, Grafana)

**Multi-Cluster Setup:**
- Read [MULTI_CLUSTER_SETUP.md](./MULTI_CLUSTER_SETUP.md)
- Register additional clusters (prod, staging, different regions)
- Configure network policies between clusters

**Development:**
- Read [CONTRIBUTING.md](./CONTRIBUTING.md)
- Set up integration test environment
- Review [ARCHITECTURE.md](./ARCHITECTURE.md) for deep dive

**Features:**
- Deploy applications from Git repositories
- Set up PostgreSQL databases with automated backups
- Configure secrets and environment variables
- Set up custom domains and SSL certificates
- Monitor logs and metrics via API

---

## Quick Reference

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `JWT_SECRET` | Secret for signing JWT tokens | Random 32+ chars |
| `ENCRYPTION_KEY` | Key for encrypting secrets | Random 64 hex chars |
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_NAME` | Database name | stackdome |
| `DB_USERNAME` | Database user | postgres |
| `DB_PASSWORD` | Database password | your-password |

### Default Ports

| Service | Port | Purpose |
|---------|------|---------|
| PostgreSQL | 5432 | Database |
| API Server | 8000 | REST API & Web UI |
| Traefik | 80, 443 | Ingress (inside cluster) |

### Important Paths

| Path | Description |
|------|-------------|
| `~/.stackdome/` | Data directory (if used) |
| `/tmp/stackdome-token.txt` | Saved API token |
| `/tmp/stackdome-org-id.txt` | Organization ID |
| `k3d kubeconfig get stackdome-dev` | k3d cluster kubeconfig (generated dynamically) |

### Useful Commands

```bash
# Check API server health
curl http://localhost:8000/health

# Get current user
curl -H "Authorization: Bearer $(cat /tmp/stackdome-token.txt)" \
  http://localhost:8000/api/v1/users/current

# List all stacks
curl -H "Authorization: Bearer $(cat /tmp/stackdome-token.txt)" \
  http://localhost:8000/api/v1/organizations/$(cat /tmp/stackdome-org-id.txt)/stacks

# Check cluster agent operator
kubectl get deployment -n stackdome-control-plane

# List all Stack CRs
kubectl get stacks -A

# Clean up everything
docker stop stackdome-postgres && docker rm stackdome-postgres
k3d cluster delete stackdome-dev
pkill -f stackdome-server
```

---

## Support

- **GitHub Issues**: Report bugs and request features
- **Documentation**: Check [docs/](./docs/) directory
- **Community**: Join discussions on GitHub Discussions

**Bootstrap complete!** You now have a fully functional Stackdome platform ready to deploy applications across Kubernetes clusters.
