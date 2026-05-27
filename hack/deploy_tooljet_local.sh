#!/usr/bin/env bash
#
# Deploy ToolJet to a local Stackdome environment.
#
# This script bootstraps everything from scratch:
#   1. PostgreSQL database for the API server
#   2. k3d cluster with stackdome-agent Helm chart (operators + CRDs + cluster agent)
#   3. API server (built and run locally)
#   4. Service account, cluster registration, org domain
#   5. (Optional) Postgres addon deployment
#   6. ToolJet stack deployment
#
# Prerequisites:
#   - Go 1.22+
#   - Docker
#   - k3d (https://k3d.io)
#   - kubectl
#   - jq
#   - mage (https://magefile.org)
#
# Usage:
#   # Deploy ToolJet (without postgres addon)
#   ./hack/deploy_tooljet_local.sh
#
#   # Deploy ToolJet with a postgres addon
#   USE_ADDON=true ./hack/deploy_tooljet_local.sh
#
# Environment variables (all optional, defaults provided):
#   DB_HOST              PostgreSQL host (default: localhost)
#   DB_PORT              PostgreSQL port (default: 5432)
#   DB_USERNAME          PostgreSQL user (default: postgres)
#   DB_PASSWORD          PostgreSQL password (default: foobar-bizz-buzz)
#   DB_NAME              Database name (default: stackdome_tooljet_demo)
#   API_PORT             API server port (default: 8000)
#   ORG_DOMAIN           Organisation domain (default: local.stackdome.io)
#   USE_ADDON            Set to "true" to use Postgres addon instead of raw postgres image
#   SKIP_CLUSTER         Set to "true" to skip k3d cluster setup (reuse existing)
#   SKIP_DB              Set to "true" to skip database creation
#   SKIP_API_SERVER      Set to "true" to skip building/starting the API server

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_SERVER_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# --- Configuration ---
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USERNAME="${DB_USERNAME:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-foobar-bizz-buzz}"
DB_NAME="${DB_NAME:-stackdome_tooljet_demo}"
API_PORT="${API_PORT:-8000}"
API_BASE="http://localhost:${API_PORT}"
ADMIN_EMAIL="admin@stackdome.io"
ADMIN_PASS="welcome@123"
K3D_CLUSTER_NAME="${K3D_CLUSTER_NAME:-stackdome-dev}"
ORG_DOMAIN="${ORG_DOMAIN:-127.0.0.1.nip.io}"
SKIP_CLUSTER="${SKIP_CLUSTER:-false}"
SKIP_DB="${SKIP_DB:-false}"
SKIP_API_SERVER="${SKIP_API_SERVER:-false}"
USE_ADDON="${USE_ADDON:-false}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log()  { echo -e "${GREEN}[+]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }
err()  { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }
info() { echo -e "${BLUE}[i]${NC} $*"; }

PG_CONTAINER_NAME="psql-stackdome-demo"

# ============================================================
# Cleanup
# ============================================================
cleanup() {
    log ""
    log "Cleaning up..."

    if [[ -n "${STACK_ID:-}" && -n "${AUTH_TOKEN:-}" && -n "${ORG_ID:-}" ]]; then
        log "Deleting ToolJet stack (${STACK_ID})..."
        curl -sf -X DELETE "${API_BASE}/api/v1/organizations/${ORG_ID}/stacks/${STACK_ID}" \
            -H "Authorization: Bearer ${AUTH_TOKEN}" >/dev/null 2>&1 || true
    fi

    if [[ -n "${ADDON_ID:-}" && -n "${AUTH_TOKEN:-}" && -n "${ORG_ID:-}" ]]; then
        log "Deleting Postgres addon (${ADDON_ID})..."
        curl -sf -X DELETE "${API_BASE}/api/v1/organizations/${ORG_ID}/addons/postgres/${ADDON_ID}" \
            -H "Authorization: Bearer ${AUTH_TOKEN}" >/dev/null 2>&1 || true
    fi

    if [[ -n "${CLUSTER_ID:-}" && -n "${AUTH_TOKEN:-}" && -n "${ORG_ID:-}" ]]; then
        log "Deleting cluster registration (${CLUSTER_ID})..."
        curl -sf -X DELETE "${API_BASE}/api/v1/organizations/${ORG_ID}/clusters/${CLUSTER_ID}" \
            -H "Authorization: Bearer ${AUTH_TOKEN}" >/dev/null 2>&1 || true
    fi

    if [[ -n "${API_SERVER_PID:-}" ]]; then
        log "Stopping API server (PID: $API_SERVER_PID)"
        kill "$API_SERVER_PID" 2>/dev/null || true
        wait "$API_SERVER_PID" 2>/dev/null || true
    fi

    if docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${PG_CONTAINER_NAME}$"; then
        log "Removing PostgreSQL container (${PG_CONTAINER_NAME})..."
        docker rm -f "$PG_CONTAINER_NAME" >/dev/null 2>&1 || true
    fi

    if k3d cluster list -o json 2>/dev/null | jq -e ".[] | select(.name == \"${K3D_CLUSTER_NAME}\")" >/dev/null 2>&1; then
        log "Deleting k3d cluster (${K3D_CLUSTER_NAME})..."
        k3d cluster delete "$K3D_CLUSTER_NAME" 2>/dev/null || true
    fi

    if [[ -f "$API_SERVER_DIR/.env.bak" ]]; then
        mv "$API_SERVER_DIR/.env.bak" "$API_SERVER_DIR/.env"
        log "Restored original .env"
    fi

    log "Cleanup complete."
}
trap cleanup EXIT

# ============================================================
# Prerequisites
# ============================================================
check_prerequisites() {
    log "Checking prerequisites..."
    local missing=()
    for cmd in go docker k3d kubectl jq mage helm; do
        if ! command -v "$cmd" &>/dev/null; then
            missing+=("$cmd")
        fi
    done
    if [[ ${#missing[@]} -gt 0 ]]; then
        err "Missing required tools: ${missing[*]}"
    fi
    log "All prerequisites found."
}

# ============================================================
# Step 1: Database
# ============================================================
setup_database() {
    if [[ "$SKIP_DB" == "true" ]]; then
        warn "Skipping database setup (SKIP_DB=true)"
        return
    fi

    if docker ps --format '{{.Names}}' | grep -q "^${PG_CONTAINER_NAME}$"; then
        log "PostgreSQL container '${PG_CONTAINER_NAME}' already running."
    elif docker ps -a --format '{{.Names}}' | grep -q "^${PG_CONTAINER_NAME}$"; then
        log "Starting existing PostgreSQL container '${PG_CONTAINER_NAME}'..."
        docker start "$PG_CONTAINER_NAME"
    else
        log "Starting PostgreSQL container '${PG_CONTAINER_NAME}'..."
        docker run --name "$PG_CONTAINER_NAME" \
            -e POSTGRES_USER="$DB_USERNAME" \
            -e POSTGRES_PASSWORD="$DB_PASSWORD" \
            -e POSTGRES_DB="$DB_NAME" \
            -p "${DB_PORT}:5432" \
            -d postgres
    fi

    log "Waiting for PostgreSQL to be ready..."
    for i in $(seq 1 30); do
        if docker exec "$PG_CONTAINER_NAME" pg_isready -U "$DB_USERNAME" >/dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    docker exec "$PG_CONTAINER_NAME" pg_isready -U "$DB_USERNAME" >/dev/null 2>&1 \
        || err "PostgreSQL failed to start within 30 seconds."
    log "PostgreSQL ready with database '${DB_NAME}'."
}

# ============================================================
# Step 2: k3d cluster + cluster agent
# ============================================================
setup_cluster() {
    if [[ "$SKIP_CLUSTER" == "true" ]]; then
        warn "Skipping cluster setup (SKIP_CLUSTER=true)"
        return
    fi

    if k3d cluster list -o json 2>/dev/null | jq -e ".[] | select(.name == \"${K3D_CLUSTER_NAME}\")" >/dev/null 2>&1; then
        warn "k3d cluster '${K3D_CLUSTER_NAME}' already exists. Reusing."
    else
        log "Creating k3d cluster '${K3D_CLUSTER_NAME}' with port mappings..."
        k3d cluster create "$K3D_CLUSTER_NAME" \
            --port "80:80@loadbalancer" \
            --port "443:443@loadbalancer" \
            --k3s-arg "--disable=traefik@server:0" \
            --agents 2 \
            --wait

        log "Installing stackdome-agent Helm chart (operators + CRDs)..."
        log "This may take 5-10 minutes..."

        local stackdome_chart_version="${STACKDOME_CHART_VERSION:-0.5.6-alpha}"
        helm upgrade --install stackdome-agent \
            "oci://quay.io/stackdome/charts/stackdome-agent" \
            --version "$stackdome_chart_version" \
            --namespace stackdome-system \
            --create-namespace \
            --wait \
            --timeout 10m
    fi

}

# ============================================================
# Step 3: Build and start API server
# ============================================================
start_api_server() {
    if [[ "$SKIP_API_SERVER" == "true" ]]; then
        warn "Skipping API server start (SKIP_API_SERVER=true)"
        return
    fi

    log "Building API server..."
    cd "$API_SERVER_DIR"
    make binary

    if [[ -f "$API_SERVER_DIR/.env" ]]; then
        cp "$API_SERVER_DIR/.env" "$API_SERVER_DIR/.env.bak"
        warn "Existing .env backed up to .env.bak"
    fi
    cat > "$API_SERVER_DIR/.env" <<EOF
JWT_SECRET="ScmCX4vNcS5nj9HFSQbq7PYnRaxM29Lz9E5Z5r1A5RAWZz9li6CMqi2YSxJK5uEU"
LOG_LEVEL="info"
ENCRYPTION_KEY="6193d7a7dec2e569548f0eaa46a87fb6a2d9288649dd35c827208d5e2b751d3c"
DB_HOST="${DB_HOST}"
DB_PORT="${DB_PORT}"
DB_NAME="${DB_NAME}"
DB_USERNAME="${DB_USERNAME}"
DB_PASSWORD="${DB_PASSWORD}"
DB_DEBUG_MODE="false"
DEFAULT_USER_EMAIL="${ADMIN_EMAIL}"
DEFAULT_USER_NAME="admin"
DEFAULT_USER_PASS="${ADMIN_PASS}"
EOF

    log "Running database migrations..."
    ./bin/stackdome-server migrate

    log "Starting API server on port ${API_PORT}..."
    ./bin/stackdome-server serve &
    API_SERVER_PID=$!

    log "Waiting for API server to be ready..."
    for i in $(seq 1 30); do
        if curl -sf "${API_BASE}/health" >/dev/null 2>&1; then
            log "API server is ready."
            return
        fi
        sleep 1
    done
    err "API server failed to start within 30 seconds."
}

# ============================================================
# Step 4: Sign up admin user and get organization
# ============================================================
signup_and_authenticate() {
    log "Signing up admin user ${ADMIN_EMAIL}..."
    local response
    response=$(curl -sf -X POST "${API_BASE}/api/v1/user-signup" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"Admin\",
            \"email\": \"${ADMIN_EMAIL}\",
            \"password\": \"${ADMIN_PASS}\",
            \"organisation\": { \"name\": \"Default\" }
        }")

    AUTH_TOKEN=$(echo "$response" | jq -r '.jwt_token')
    if [[ -z "$AUTH_TOKEN" || "$AUTH_TOKEN" == "null" ]]; then
        # User may already exist from a previous run — try login instead
        warn "Signup failed (user may already exist). Trying login..."
        response=$(curl -sf -X POST "${API_BASE}/api/v1/auth/login" \
            -H "Content-Type: application/json" \
            -d "{\"email\": \"${ADMIN_EMAIL}\", \"password\": \"${ADMIN_PASS}\"}")

        AUTH_TOKEN=$(echo "$response" | jq -r '.token')
        if [[ -z "$AUTH_TOKEN" || "$AUTH_TOKEN" == "null" ]]; then
            err "Failed to authenticate. Response: $response"
        fi
        ORG_ID=$(echo "$response" | jq -r '.user.organisation_id')
    else
        ORG_ID=$(echo "$response" | jq -r '.user.organisation_id')
    fi

    if [[ -z "$ORG_ID" || "$ORG_ID" == "null" ]]; then
        err "Failed to get organization ID."
    fi

    log "Authenticated. Token obtained."
    log "Organization ID: ${ORG_ID}"
}

api() {
    local method="$1" path="$2"
    shift 2
    curl -sf -X "$method" "${API_BASE}${path}" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${AUTH_TOKEN}" \
        "$@"
}

# ============================================================
# Step 6: Add domain to organisation
# ============================================================
create_org_domain() {
    log "Adding domain '${ORG_DOMAIN}' to organisation..."
    local response
    response=$(api PUT "/api/v1/organizations/${ORG_ID}" \
        -d "{\"domains\": [{\"fqdn\": \"${ORG_DOMAIN}\"}]}")
    local domain_count
    domain_count=$(echo "$response" | jq '.domains | length')
    if [[ "$domain_count" -lt 1 ]]; then
        err "Failed to add domain. Response: $response"
    fi
    log "Organisation domain '${ORG_DOMAIN}' added."
}

# ============================================================
# Step 7: Setup service account and register cluster
# ============================================================
setup_service_account() {
    log "Setting up service account in k3d cluster..."
    local kubeconfig
    kubeconfig=$(k3d kubeconfig get "$K3D_CLUSTER_NAME")

    export KUBECONFIG_DATA="$kubeconfig"

    kubectl --kubeconfig <(echo "$kubeconfig") create namespace stackdome-control-plane 2>/dev/null || true

    kubectl --kubeconfig <(echo "$kubeconfig") apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: stackdome-api-server-account
  namespace: stackdome-control-plane
EOF

    kubectl --kubeconfig <(echo "$kubeconfig") apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: stackdome-api-server-role
rules:
  - apiGroups: ["*"]
    resources: ["*"]
    verbs: ["*"]
EOF

    kubectl --kubeconfig <(echo "$kubeconfig") apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: stackdome-api-server-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: stackdome-api-server-role
subjects:
  - kind: ServiceAccount
    name: stackdome-api-server-account
    namespace: stackdome-control-plane
EOF

    kubectl --kubeconfig <(echo "$kubeconfig") apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: stackdome-api-server-account-secret
  namespace: stackdome-control-plane
  annotations:
    kubernetes.io/service-account.name: stackdome-api-server-account
type: kubernetes.io/service-account-token
EOF

    log "Waiting for service account token..."
    for i in $(seq 1 30); do
        local token
        token=$(kubectl --kubeconfig <(echo "$kubeconfig") get secret \
            stackdome-api-server-account-secret \
            -n stackdome-control-plane \
            -o jsonpath='{.data.token}' 2>/dev/null || echo "")
        if [[ -n "$token" ]]; then
            SA_TOKEN=$(echo "$token" | base64 -d)
            break
        fi
        sleep 1
    done
    [[ -n "${SA_TOKEN:-}" ]] || err "Timeout waiting for service account token."

    CLUSTER_URL=$(kubectl --kubeconfig <(echo "$kubeconfig") config view --raw --minify \
        --flatten -o jsonpath='{.clusters[0].cluster.server}')
    CLUSTER_CA=$(kubectl --kubeconfig <(echo "$kubeconfig") get secret \
        stackdome-api-server-account-secret \
        -n stackdome-control-plane \
        -o jsonpath='{.data.ca\.crt}')

    log "Service account ready."
    log "  Cluster URL: ${CLUSTER_URL}"
}

register_cluster() {
    log "Registering cluster with API server..."
    local payload
    payload=$(jq -n \
        --arg name "local-k3d-cluster" \
        --arg url "$CLUSTER_URL" \
        --arg ca "$CLUSTER_CA" \
        --arg token "$SA_TOKEN" \
        '{
            name: $name,
            cluster_url: $url,
            cluster_ca_data: $ca,
            cluster_sa_token: $token,
            cluster_image_registry: {
                name: "local-registry",
                spec: {
                    backend_storage_size: "10Gi",
                    backend_storage_class: "local-path",
                    max_repositories: 100,
                    tags_per_repository: 50,
                    delete_untagged: true
                }
            }
        }')

    local response
    response=$(api POST "/api/v1/organizations/${ORG_ID}/clusters" -d "$payload")
    CLUSTER_ID=$(echo "$response" | jq -r '.id')
    if [[ -z "$CLUSTER_ID" || "$CLUSTER_ID" == "null" ]]; then
        err "Failed to register cluster. Response: $response"
    fi
    log "Cluster registered. ID: ${CLUSTER_ID}"
}

# ============================================================
# Step 8: Wait for image registry
# ============================================================
wait_for_registry() {
    log "Waiting for in-cluster image registry to reach Running state..."
    for i in $(seq 1 120); do
        local response
        response=$(api GET "/api/v1/organizations/${ORG_ID}/clusters/${CLUSTER_ID}/image_registries" 2>/dev/null || echo "[]")
        local state
        state=$(echo "$response" | jq -r '.items[0].status.state // empty' 2>/dev/null || echo "")
        if [[ "$state" == "ImageRegistryRunning" ]]; then
            log "Image registry is Running."
            return
        fi
        if (( i % 10 == 0 )); then
            warn "  Registry state: ${state:-unknown} (${i}s elapsed, waiting up to 120s)..."
        fi
        sleep 1
    done
    warn "Registry did not reach Running state within 120s. Proceeding anyway."
}

# ============================================================
# Step 9: Create Postgres addon (if USE_ADDON=true)
# ============================================================
create_postgres_addon() {
    if [[ "$USE_ADDON" != "true" ]]; then
        return
    fi

    local addon_file="${API_SERVER_DIR}/samples/tooljet_addon_postgres.json"
    if [[ ! -f "$addon_file" ]]; then
        err "Addon file not found: $addon_file"
    fi

    local addon_name
    addon_name=$(jq -r '.name' "$addon_file")
    log "Creating Postgres addon '${addon_name}'..."

    local response
    response=$(api POST "/api/v1/organizations/${ORG_ID}/addons/postgres" -d @"$addon_file")
    ADDON_ID=$(echo "$response" | jq -r '.id')
    if [[ -z "$ADDON_ID" || "$ADDON_ID" == "null" ]]; then
        err "Failed to create postgres addon. Response: $response"
    fi
    log "Postgres addon created. ID: ${ADDON_ID}"

    log "Waiting for Postgres addon to become Ready..."
    for i in $(seq 1 300); do
        local state
        state=$(api GET "/api/v1/organizations/${ORG_ID}/addons/postgres/${ADDON_ID}" 2>/dev/null \
            | jq -r '.status.state // empty' 2>/dev/null || echo "")
        if [[ "$state" == "Ready" ]]; then
            log "Postgres addon is Ready."
            return
        fi
        if (( i % 15 == 0 )); then
            warn "  Addon state: ${state:-unknown} (${i}s elapsed, waiting up to 300s)..."
        fi
        sleep 1
    done
    warn "Postgres addon did not reach Ready state within 300s. Proceeding anyway."
}

# ============================================================
# Step 10: Deploy ToolJet stack
# ============================================================
deploy_tooljet() {
    local stack_file response
    if [[ "$USE_ADDON" == "true" ]]; then
        log "Deploying ToolJet stack (with Postgres addon)..."
        stack_file="${API_SERVER_DIR}/samples/tooljet_with_addon.json"
        [[ -f "$stack_file" ]] || err "Stack file not found: $stack_file"

        local stack_payload
        stack_payload=$(sed "s/<POSTGRES_ADDON_ID>/${ADDON_ID}/g" "$stack_file")
        response=$(echo "$stack_payload" | api POST "/api/v1/organizations/${ORG_ID}/stacks" -d @-)
    else
        log "Deploying ToolJet stack..."
        stack_file="${API_SERVER_DIR}/samples/tooljet.json"
        [[ -f "$stack_file" ]] || err "Stack file not found: $stack_file"

        response=$(api POST "/api/v1/organizations/${ORG_ID}/stacks" -d @"$stack_file")
    fi

    STACK_ID=$(echo "$response" | jq -r '.id')
    if [[ -z "$STACK_ID" || "$STACK_ID" == "null" ]]; then
        err "Failed to create stack. Response: $response"
    fi
    log "ToolJet stack created. ID: ${STACK_ID}"
}

# ============================================================
# Print environment info and wait
# ============================================================
print_info() {
    log ""
    log "============================================"
    log " ToolJet deployed to local Stackdome"
    log "============================================"
    log ""
    info "API Server:   ${API_BASE}"
    info "Org ID:       ${ORG_ID}"
    info "Cluster ID:   ${CLUSTER_ID}"
    info "Auth Token:   ${AUTH_TOKEN}"
    info "Org Domain:   ${ORG_DOMAIN}"
    info "Kubectl ctx:  k3d-${K3D_CLUSTER_NAME}"
    info "Stack ID:     ${STACK_ID}"

    if [[ -n "${ADDON_ID:-}" ]]; then
        info "Addon ID:     ${ADDON_ID}"
    fi

    local config_file="${API_SERVER_DIR}/dev_env.yaml"
    cat > "$config_file" <<EOF
api_server: ${API_BASE}
org_id: ${ORG_ID}
cluster_id: ${CLUSTER_ID}
auth_token: ${AUTH_TOKEN}
org_domain: ${ORG_DOMAIN}
kubectl_context: k3d-${K3D_CLUSTER_NAME}
admin_email: ${ADMIN_EMAIL}
admin_password: ${ADMIN_PASS}
stack_id: ${STACK_ID}
EOF
    if [[ -n "${ADDON_ID:-}" ]]; then
        echo "addon_id: ${ADDON_ID}" >> "$config_file"
    fi
    log "Config saved to ${config_file}"

    log ""
    log "Useful commands:"
    log ""
    log "  # Check stack status"
    log "  curl -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
    log "    ${API_BASE}/api/v1/organizations/${ORG_ID}/stacks/${STACK_ID} | jq '.status'"
    log ""
    log "  # List stack resources"
    log "  curl -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
    log "    ${API_BASE}/api/v1/organizations/${ORG_ID}/stacks/${STACK_ID}/resources | jq '.[].name'"
    log ""
    log "  # Watch pods in the cluster"
    log "  kubectl --context k3d-${K3D_CLUSTER_NAME} get pods -A -w"
    log ""
    log "Press Ctrl+C to tear down and exit."
    wait "$API_SERVER_PID"
}

# ============================================================
# Main
# ============================================================
main() {
    log "Starting ToolJet deployment to local Stackdome environment"
    if [[ "$USE_ADDON" == "true" ]]; then
        info "Using Postgres addon"
    fi
    log ""

    check_prerequisites
    setup_database
    setup_cluster
    start_api_server
    signup_and_authenticate
    create_org_domain
    setup_service_account
    register_cluster
    wait_for_registry
    create_postgres_addon
    deploy_tooljet
    print_info
}

main "$@"
