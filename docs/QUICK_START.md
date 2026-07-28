# Stackdome Quick Start

TL;DR guide for experienced users. For detailed instructions, see [BOOTSTRAP_GUIDE.md](./BOOTSTRAP_GUIDE.md).

## Prerequisites

✅ Go 1.24+, Docker, kubectl, k3d (https://k3d.io), jq, mage

## 5-Minute Setup

### 1. Start PostgreSQL

```bash
docker run -d --name stackdome-postgres \
  -e POSTGRES_PASSWORD=mypassword \
  -e POSTGRES_DB=stackdome \
  -p 5432:5432 postgres:15-alpine
```

### 2. Setup Cluster Agent

```bash
cd ~/stackdome/stackdome-cluster-agent
unset -f cd 2>/dev/null || true
./mage dev:setup
export KUBECONFIG=$(pwd)/.cache/dev-env/kubeconfig.yaml
```

### 3. Setup API Server

```bash
cd ~/stackdome/stackdome-api-server

# Configure .env
cat > .env << 'EOF'
JWT_SECRET=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -hex 32)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=stackdome
DB_USERNAME=postgres
DB_PASSWORD=mypassword
EOF

# Build and run
go build -o bin/stackdome-server ./cmd/main.go
./bin/stackdome-server migrate
./bin/stackdome-server serve &
```

### 4. Register Cluster

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@stackdome.127.0.0.1.nip.io","password":"admin123"}' | jq -r '.token')

ORG_ID=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@stackdome.127.0.0.1.nip.io","password":"admin123"}' | jq -r '.user.organisation_id')

# Deploy service account
export KUBECONFIG=~/stackdome/stackdome-cluster-agent/.cache/dev-env/kubeconfig.yaml
kubectl apply -f deploy/sa.yaml
sleep 3

# Extract credentials
SA_TOKEN=$(kubectl get secret -n stackdome-control-plane \
  stackdome-api-server-account-secret \
  -o jsonpath='{.data.token}' | base64 -d | base64)

SA_CA=$(kubectl get secret -n stackdome-control-plane \
  stackdome-api-server-account-secret \
  -o jsonpath='{.data.ca\.crt}')

CLUSTER_URL=$(kubectl cluster-info | grep "Kubernetes control plane" | \
  awk '{print $NF}' | sed 's/\x1b\[[0-9;]*m//g')

# Register
curl -X POST "http://localhost:8000/api/v1/organizations/${ORG_ID}/clusters" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"dev-cluster\",\"cluster_url\":\"${CLUSTER_URL}\",\"cluster_ca_data\":\"${SA_CA}\",\"cluster_sa_token\":\"${SA_TOKEN}\",\"default\":true}"
```

### 5. Deploy First App

```bash
# Add domain to organization
curl -X PUT "http://localhost:8000/api/v1/organizations/${ORG_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"DefaultOrganisation","domains":[{"fqdn":"stackdome.127.0.0.1.nip.io"}]}'

# Create stack
curl -X POST "http://localhost:8000/api/v1/organizations/${ORG_ID}/stacks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx-demo",
    "spec": {
      "stack_resources": [{
        "name": "web",
        "image_spec": {"image": "nginx:latest"},
        "ports": [{"number": 80, "exposed_to_public": true, "protocol": "HTTP"}]
      }]
    }
  }'

# Verify
kubectl get stacks -A
kubectl get pods -A | grep nginx-demo
```

## Common Commands

```bash
# API Server
curl http://localhost:8000/health
./bin/stackdome-server migrate
./bin/stackdome-server serve

# Get Token
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@stackdome.127.0.0.1.nip.io","password":"admin123"}' | jq -r '.token')

# List Stacks
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/v1/organizations/${ORG_ID}/stacks" | jq .

# Cluster Agent
cd ~/stackdome/stackdome-cluster-agent
./mage dev:setup      # Create cluster
./mage dev:teardown   # Delete cluster
./mage dev:deploy     # Redeploy operator

# Kubernetes
export KUBECONFIG=~/stackdome/stackdome-cluster-agent/.cache/dev-env/kubeconfig.yaml
kubectl get stacks -A
kubectl get pods -A
kubectl logs -n <namespace> deployment/<name>
```

## Cleanup

```bash
# Stop API server
pkill -f stackdome-server

# Delete cluster
cd ~/stackdome/stackdome-cluster-agent && ./mage dev:teardown

# Remove PostgreSQL
docker stop stackdome-postgres && docker rm stackdome-postgres
```

## Next Steps

- Full guide: [BOOTSTRAP_GUIDE.md](./BOOTSTRAP_GUIDE.md)
- Architecture: [ARCHITECTURE.md](./ARCHITECTURE.md)
- Production: [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md)
