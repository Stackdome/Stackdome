#!/usr/bin/env bash
#
# Run Stackdome locally for development.
#
# Bootstraps a complete local Stackdome environment:
#   1. PostgreSQL database for the API server
#   2. Kind cluster with stackdome-agent Helm chart (operators + CRDs + cluster agent)
#   3. API server (built and run locally)
#   4. Service account, cluster registration, org domain
#   5. (Optional) Postgres addon deployment
#   6. (Optional) Stack deployment
#
# Prerequisites:
#   - Go 1.22+
#   - Docker
#   - kind (https://kind.sigs.k8s.io)
#   - kubectl
#   - jq
#   - mage (https://magefile.org)
#
# Usage:
#   # Start environment only (no stack)
#   ./hack/run_local.sh
#
#   # Deploy a stack
#   ./hack/run_local.sh samples/tooljet.json
#
#   # Deploy a stack with a postgres addon
#   ADDON_FILE=samples/tooljet_addon_postgres.json ./hack/run_local.sh samples/tooljet_with_addon.json
#
# Environment variables (all optional, defaults provided):
#   DB_HOST              PostgreSQL host (default: localhost)
#   DB_PORT              PostgreSQL port (default: 5432)
#   DB_USERNAME          PostgreSQL user (default: postgres)
#   DB_PASSWORD          PostgreSQL password (default: foobar-bizz-buzz)
#   DB_NAME              Database name (default: stackdome_local_dev)
#   API_PORT             API server port (default: 8000)
#   ORG_DOMAIN           Organisation domain (default: local.stackdome.io)
#   ADDON_FILE             Postgres addon JSON file (creates addon before stack)
#   SKIP_CLUSTER           Set to "true" to skip Kind cluster setup (reuse existing)
#   SKIP_DB                Set to "true" to skip database creation
#   SKIP_API_SERVER        Set to "true" to skip building/starting the API server

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_SERVER_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# --- Configuration ---
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USERNAME="${DB_USERNAME:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-foobar-bizz-buzz}"
DB_NAME="${DB_NAME:-stackdome_local_dev}"
API_PORT="${API_PORT:-8000}"
API_BASE="http://localhost:${API_PORT}"
ADMIN_EMAIL="admin@stackdome.io"
ADMIN_PASS="welcome@123"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-stackdome-dev-cluster}"
ORG_DOMAIN="${ORG_DOMAIN:-local.stackdome.io}"
SKIP_CLUSTER="${SKIP_CLUSTER:-false}"
SKIP_DB="${SKIP_DB:-false}"
SKIP_API_SERVER="${SKIP_API_SERVER:-false}"

STACK_FILE="${1:-}"
ADDON_FILE="${ADDON_FILE:-}"

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

PG_CONTAINER_NAME="psql-stackdome-dev"

# ============================================================
# Cleanup
# ============================================================
cleanup() {
    log ""
    log "Cleaning up..."

    if [[ -n "${STACK_ID:-}" && -n "${AUTH_TOKEN:-}" && -n "${ORG_ID:-}" ]]; then
        log "Deleting stack (${STACK_ID})..."
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

    if kind get clusters 2>/dev/null | grep -q "^${KIND_CLUSTER_NAME}$"; then
        log "Deleting Kind cluster (${KIND_CLUSTER_NAME})..."
        kind delete cluster --name "$KIND_CLUSTER_NAME" 2>/dev/null || true
    fi

    if [[ -f "$API_SERVER_DIR/.env.bak" ]]; then
        mv "$API_SERVER_DIR/.env.bak" "$API_SERVER_DIR/.env"
        log "Restored original .env"
    fi

    # Delete the local_dev_env.yaml config file
    if [[ -f "$API_SERVER_DIR/local_dev_env.yaml" ]]; then
        rm "$API_SERVER_DIR/local_dev_env.yaml"
        log "Deleted local_dev_env.yaml"
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
    for cmd in go docker kind kubectl jq mage; do
        if ! command -v "$cmd" &>/dev/null; then
            missing+=("$cmd")
        fi
    done
    if [[ ${#missing[@]} -gt 0 ]]; then
        err "Missing required tools: ${missing[*]}"
    fi

    if [[ -n "$ADDON_FILE" && -z "$STACK_FILE" ]]; then
        err "ADDON_FILE requires a stack file argument (the stack must reference the addon)"
    fi

    if [[ -n "$STACK_FILE" && ! -f "$STACK_FILE" ]]; then
        err "Stack file not found: $STACK_FILE"
    fi

    if [[ -n "$ADDON_FILE" && ! -f "$ADDON_FILE" ]]; then
        err "Addon file not found: $ADDON_FILE"
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
# Step 2: Kind cluster + cluster agent
# ============================================================
setup_cluster() {
    if [[ "$SKIP_CLUSTER" == "true" ]]; then
        warn "Skipping cluster setup (SKIP_CLUSTER=true)"
        return
    fi

    if kind get clusters 2>/dev/null | grep -q "^${KIND_CLUSTER_NAME}$"; then
        warn "Kind cluster '${KIND_CLUSTER_NAME}' already exists. Reusing."
    else
        log "Setting up Kind cluster with stackdome-agent chart (includes all operators and CRDs)..."
        log "This may take 5-10 minutes..."
        cd "$API_SERVER_DIR"
        KIND_CLUSTER_NAME="$KIND_CLUSTER_NAME" mage cluster:setup
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
# Step 4: Authenticate
# ============================================================
authenticate() {
    log "Authenticating as ${ADMIN_EMAIL}..."
    local response
    response=$(curl -sf -X POST "${API_BASE}/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"${ADMIN_EMAIL}\", \"password\": \"${ADMIN_PASS}\"}")

    AUTH_TOKEN=$(echo "$response" | jq -r '.token')
    if [[ -z "$AUTH_TOKEN" || "$AUTH_TOKEN" == "null" ]]; then
        err "Failed to authenticate. Response: $response"
    fi
    log "Authenticated. Token obtained."
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
# Step 5: Get default organization
# ============================================================
get_default_org() {
    log "Fetching default organization..."
    local response
    response=$(api GET "/api/v1/organizations/default")
    ORG_ID=$(echo "$response" | jq -r '.id')
    if [[ -z "$ORG_ID" || "$ORG_ID" == "null" ]]; then
        err "Failed to get default organization. Response: $response"
    fi
    log "Organization ID: ${ORG_ID}"
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
    log "Setting up service account in Kind cluster..."
    local kubeconfig
    kubeconfig=$(kind get kubeconfig --name "$KIND_CLUSTER_NAME")

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
        --arg name "local-kind-cluster" \
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
                    backend_storage_class: "standard",
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
# Step 9: Create addon (optional)
# ============================================================
create_addon() {
    if [[ -z "$ADDON_FILE" ]]; then
        return
    fi

    local addon_name
    addon_name=$(jq -r '.name' "$ADDON_FILE")
    log "Creating Postgres addon '${addon_name}'..."

    local response
    response=$(api POST "/api/v1/organizations/${ORG_ID}/addons/postgres" -d @"$ADDON_FILE")
    ADDON_ID=$(echo "$response" | jq -r '.id')
    if [[ -z "$ADDON_ID" || "$ADDON_ID" == "null" ]]; then
        err "Failed to create postgres addon. Response: $response"
    fi
    log "Postgres addon created. ID: ${ADDON_ID}"

    log "Waiting for Postgres addon to become Ready..."
    for i in $(seq 1 30); do
        local state
        state=$(api GET "/api/v1/organizations/${ORG_ID}/addons/postgres/${ADDON_ID}" 2>/dev/null \
            | jq -r '.status.state // empty' 2>/dev/null || echo "")
        if [[ "$state" == "Ready" ]]; then
            log "Postgres addon is Ready."
            return
        fi
        if (( i % 15 == 0 )); then
            warn "  Addon state: ${state:-unknown} (${i}s elapsed, waiting up to 30s)..."
        fi
        sleep 1
    done
    warn "Postgres addon did not reach Ready state within 30s. Proceeding anyway."
}

# ============================================================
# Step 10: Deploy stack (optional)
# ============================================================
deploy_stack() {
    if [[ -z "$STACK_FILE" ]]; then
        return
    fi

    local stack_name
    stack_name=$(jq -r '.name' "$STACK_FILE")
    log "Deploying stack '${stack_name}'..."

    local response
    if [[ -n "${ADDON_ID:-}" ]]; then
        local stack_payload
        stack_payload=$(sed "s/<POSTGRES_ADDON_ID>/${ADDON_ID}/g" "$STACK_FILE")
        response=$(echo "$stack_payload" | api POST "/api/v1/organizations/${ORG_ID}/stacks" -d @-)
    else
        response=$(api POST "/api/v1/organizations/${ORG_ID}/stacks" -d @"$STACK_FILE")
    fi

    STACK_ID=$(echo "$response" | jq -r '.id')
    if [[ -z "$STACK_ID" || "$STACK_ID" == "null" ]]; then
        err "Failed to create stack. Response: $response"
    fi
    log "Stack '${stack_name}' created. ID: ${STACK_ID}"
}

# ============================================================
# Print environment info
# ============================================================
print_info() {
    log ""
    log "============================================"
    log " Stackdome local environment is running"
    log "============================================"
    log ""
    info "API Server:   ${API_BASE}"
    info "Org ID:       ${ORG_ID}"
    info "Cluster ID:   ${CLUSTER_ID}"
    info "Auth Token:   ${AUTH_TOKEN}"
    info "Org Domain:   ${ORG_DOMAIN}"
    info "Kubectl ctx:  kind-${KIND_CLUSTER_NAME}"

    local config_file="${API_SERVER_DIR}/local_dev_env.yaml"
    cat > "$config_file" <<EOF
api_server: ${API_BASE}
org_id: ${ORG_ID}
cluster_id: ${CLUSTER_ID}
auth_token: ${AUTH_TOKEN}
org_domain: ${ORG_DOMAIN}
kubectl_context: kind-${KIND_CLUSTER_NAME}
admin_email: ${ADMIN_EMAIL}
admin_password: ${ADMIN_PASS}
EOF
    log "Config saved to ${config_file}"

    log ""
    log "Useful commands:"
    log ""
    log "  # Deploy a stack"
    log "  curl -s -X POST -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
    log "    -H 'Content-Type: application/json' \\"
    log "    ${API_BASE}/api/v1/organizations/${ORG_ID}/stacks -d @samples/tooljet.json | jq"
    log ""
    log "  # Create a postgres addon"
    log "  curl -s -X POST -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
    log "    -H 'Content-Type: application/json' \\"
    log "    ${API_BASE}/api/v1/organizations/${ORG_ID}/addons/postgres -d @samples/postgres_addon_basic.json | jq"
    log ""
    log "  # Export kubeconfig"
    log "  kind get kubeconfig --name ${KIND_CLUSTER_NAME} > /tmp/kind-kubeconfig.yaml"
    log ""
    log "  # Watch pods in the cluster"
    log "  kubectl --context kind-${KIND_CLUSTER_NAME} get pods -A -w"
    log ""
}

# ============================================================
# Print stack/addon info and wait
# ============================================================
print_stack_info() {
    if [[ -z "${ADDON_ID:-}" && -z "${STACK_ID:-}" ]]; then
        return
    fi


    log ""
    log "============================================"
    log " Stack deployment info"
    log "============================================"
    log ""

    if [[ -n "${ADDON_ID:-}" ]]; then
        info "Addon ID:     ${ADDON_ID}"
    fi
    if [[ -n "${STACK_ID:-}" ]]; then
        info "Stack ID:     ${STACK_ID}"
    fi

    log ""
    log "Useful commands:"
    log ""
    log "  # List stacks"
    log "  curl -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
    log "    ${API_BASE}/api/v1/organizations/${ORG_ID}/stacks | jq '.items[].name'"
    log ""
    log "  # List postgres addons"
    log "  curl -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
    log "    ${API_BASE}/api/v1/organizations/${ORG_ID}/addons/postgres | jq '.items[] | {name, id, state: .status.state}'"

    if [[ -n "${STACK_ID:-}" ]]; then
        log ""
        log "  # Check stack status"
        log "  curl -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
        log "    ${API_BASE}/api/v1/organizations/${ORG_ID}/stacks/${STACK_ID} | jq '.status'"
        log ""
        log "  # List stack resources"
        log "  curl -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
        log "    ${API_BASE}/api/v1/organizations/${ORG_ID}/stacks/${STACK_ID}/resources | jq '.[].name'"
    fi
    log ""
}

# ============================================================
# Main
# ============================================================
main() {
    log "Starting local Stackdome environment"
    if [[ -n "$STACK_FILE" ]]; then
        info "Stack file: ${STACK_FILE}"
    fi
    if [[ -n "$ADDON_FILE" ]]; then
        info "Addon file: ${ADDON_FILE}"
    fi
    log ""

    check_prerequisites
    setup_database
    setup_cluster
    start_api_server
    authenticate
    get_default_org
    create_org_domain
    setup_service_account
    register_cluster
    print_info
    wait_for_registry
    create_addon
    deploy_stack
    print_stack_info

    log "Press Ctrl+C to tear down and exit."
    wait "$API_SERVER_PID"
}

main
