#!/usr/bin/env bash
#
# Run Stackdome locally for development.
#
# Bootstraps a complete local Stackdome environment:
#   1. mage dev:setup (PostgreSQL + k3d cluster + operators + RBAC)
#   2. API server (built and run locally)
#   3. User signup, cluster registration, org domain
#   4. (Optional) Postgres addon deployment
#   5. (Optional) Stack deployment
#
# Idempotent: safe to run multiple times. Reuses existing infrastructure,
# user, org, and cluster registration. Only the API server process is
# restarted on each run.
#
# Prerequisites:
#   - Go 1.22+
#   - Docker
#   - k3d (https://k3d.io)
#   - kubectl
#   - jq
#   - mage (https://magefile.org)
#   - helm
#
# Usage:
#   # Start environment only (no stack)
#   ./hack/run_local.sh
#
#   # Start Stackdome Cloud with the local cluster as shared compute
#   CLOUD_MODE=true ./hack/run_local.sh
#
#   # Recreate PostgreSQL when switching between BYOC and shared compute
#   CLOUD_MODE=true FORCE_PG_RECREATE=true ./hack/run_local.sh
#
#   # Deploy a stack
#   ./hack/run_local.sh samples/tooljet.json
#
#   # Deploy a stack with a postgres addon
#   ADDON_FILE=samples/tooljet_addon_postgres.json ./hack/run_local.sh samples/tooljet_with_addon.json
#
# Environment variables (all optional, defaults provided):
#   API_PORT                   API server port (default: 8000)
#   CLOUD_MODE                 Set to "true" for stackdome_cloud + shared compute
#   ORG_DOMAIN                 Organisation domain (default: stackdome.127.0.0.1.nip.io)
#   ADDON_FILE                 Postgres addon JSON file (creates addon before stack)
#   SKIP_INFRA                 Set to "true" to skip mage dev:setup (reuse existing)
#   SKIP_API_SERVER            Set to "true" to skip building/starting the API server
#   FORCE_CLUSTER_RECREATE     Set to "true" to delete and recreate k3d cluster
#   FORCE_PG_RECREATE          Set to "true" to delete and recreate PostgreSQL container

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_SERVER_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# --- Configuration ---
API_PORT="${API_PORT:-8000}"
API_BASE="http://localhost:${API_PORT}"
ADMIN_EMAIL="admin@stackdome.io"
ADMIN_PASS="welcome@123"
K3D_CLUSTER_NAME="${K3D_CLUSTER_NAME:-stackdome-dev}"
ORG_DOMAIN="${ORG_DOMAIN:-stackdome.127.0.0.1.nip.io}"
CLOUD_MODE="${CLOUD_MODE:-false}"
SKIP_INFRA="${SKIP_INFRA:-false}"
SKIP_API_SERVER="${SKIP_API_SERVER:-false}"
FORCE_CLUSTER_RECREATE="${FORCE_CLUSTER_RECREATE:-false}"
FORCE_PG_RECREATE="${FORCE_PG_RECREATE:-false}"
DEV_CONFIG="${API_SERVER_DIR}/dev_env.yaml"

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
# Cleanup — only stops the API server process, leaves infra intact
# ============================================================
cleanup() {
    if [[ -n "${API_SERVER_PID:-}" ]] && ps -p "$API_SERVER_PID" -o command= 2>/dev/null | grep -q "stackdome-server"; then
        log "Stopping API server (PID: $API_SERVER_PID)"
        kill "$API_SERVER_PID" 2>/dev/null || true
        wait "$API_SERVER_PID" 2>/dev/null || true
    fi

    if [[ -f "$API_SERVER_DIR/.env.bak" ]]; then
        mv "$API_SERVER_DIR/.env.bak" "$API_SERVER_DIR/.env"
    fi
}
trap cleanup EXIT
trap 'exit 0' INT TERM

# ============================================================
# Prerequisites
# ============================================================
check_prerequisites() {
    log "Checking prerequisites..."

    if [[ "$CLOUD_MODE" != "true" && "$CLOUD_MODE" != "false" ]]; then
        err "CLOUD_MODE must be either 'true' or 'false'"
    fi

    local missing=()
    for cmd in go docker k3d kubectl jq mage helm; do
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
# Step 1: Infrastructure (mage dev:setup)
# ============================================================
setup_infra() {
    if [[ "$SKIP_INFRA" == "true" ]]; then
        warn "Skipping infrastructure setup (SKIP_INFRA=true)"
        if [[ ! -f "$DEV_CONFIG" ]]; then
            err "SKIP_INFRA=true but ${DEV_CONFIG} not found. Run mage dev:setup first."
        fi
        return
    fi

    # Force recreate cluster if requested
    if [[ "$FORCE_CLUSTER_RECREATE" == "true" ]]; then
        warn "FORCE_CLUSTER_RECREATE=true — deleting existing k3d cluster..."
        k3d cluster delete "$K3D_CLUSTER_NAME" 2>/dev/null || true
    fi

    # Force recreate postgres if requested
    if [[ "$FORCE_PG_RECREATE" == "true" ]]; then
        warn "FORCE_PG_RECREATE=true — removing existing PostgreSQL container..."
        docker rm -f "$PG_CONTAINER_NAME" 2>/dev/null || true
    fi

    log "Running mage dev:setup (PostgreSQL + k3d cluster + operators + RBAC)..."
    log "This may take 5-10 minutes on first run..."
    cd "$API_SERVER_DIR"
    mage dev:setup

    if [[ ! -f "$DEV_CONFIG" ]]; then
        err "mage dev:setup completed but ${DEV_CONFIG} not found."
    fi
    log "Infrastructure ready."
}

# ============================================================
# Read credentials from dev_env.yaml
# ============================================================
read_dev_config() {
    log "Reading cluster credentials from ${DEV_CONFIG}..."
    CLUSTER_URL=$(grep '^cluster_url:' "$DEV_CONFIG" | awk '{print $2}')
    CLUSTER_CA=$(grep '^cluster_ca_data:' "$DEV_CONFIG" | awk '{print $2}')
    SA_TOKEN=$(grep '^cluster_sa_token:' "$DEV_CONFIG" | awk '{print $2}')
    DB_HOST=$(grep '^db_host:' "$DEV_CONFIG" | awk '{print $2}')
    DB_PORT=$(grep '^db_port:' "$DEV_CONFIG" | awk '{print $2}')
    DB_NAME=$(grep '^db_name:' "$DEV_CONFIG" | awk '{print $2}')
    DB_USERNAME=$(grep '^db_username:' "$DEV_CONFIG" | awk '{print $2}')
    DB_PASSWORD=$(grep '^db_password:' "$DEV_CONFIG" | awk '{print $2}')

    if [[ -z "$CLUSTER_URL" || -z "$SA_TOKEN" ]]; then
        err "Failed to read cluster credentials from ${DEV_CONFIG}."
    fi
    log "Cluster credentials loaded."
}

# ============================================================
# Step 2: Build and start API server
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
EOF

    if [[ "$CLOUD_MODE" == "true" ]]; then
        cat >> "$API_SERVER_DIR/.env" <<EOF
RUNTIME_MODE="stackdome_cloud"
COMPUTE_MODE="shared"
SHARED_COMPUTE_CLUSTER_API_URL="${CLUSTER_URL}"
SHARED_COMPUTE_CLUSTER_CA_DATA="${CLUSTER_CA}"
SHARED_COMPUTE_CLUSTER_TOKEN="${SA_TOKEN}"
PLATFORM_BASE_DOMAIN="${ORG_DOMAIN}"
PLATFORM_TLS_ENABLED="true"
PLATFORM_EMAIL="local@stackdome.dev"
PLATFORM_DNS_CLOUDFLARE_API_TOKEN="local-development-token"
PLATFORM_ACME_ENVIRONMENT="staging"
PLATFORM_TLS_NAMESPACE="stackdome-control-plane"
STACKDOME_CLOUD_CONFIG="${API_SERVER_DIR}/config/stackdome_cloud.dev.yaml"
TURNSTILE_SECRET="local-development-secret"
EOF
    else
        cat >> "$API_SERVER_DIR/.env" <<EOF
RUNTIME_MODE="self_hosted"
COMPUTE_MODE="bring_your_own"
EOF
    fi

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
# Step 3: Sign up or login
# ============================================================
signup_and_authenticate() {
    # Try login first (existing user from previous run)
    log "Authenticating..."
    local response
    response=$(curl -sf -X POST "${API_BASE}/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"${ADMIN_EMAIL}\", \"password\": \"${ADMIN_PASS}\"}" 2>/dev/null || echo "")

    AUTH_TOKEN=$(echo "$response" | jq -r '.token // empty' 2>/dev/null)
    if [[ -n "$AUTH_TOKEN" ]]; then
        ORG_ID=$(echo "$response" | jq -r '.user.organisation_id')
        log "Logged in as existing user."
    else
        # First run — sign up
        log "No existing user. Signing up ${ADMIN_EMAIL}..."
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
            err "Failed to sign up. Response: $response"
        fi
        ORG_ID=$(echo "$response" | jq -r '.user.organisation_id')
    fi

    if [[ -z "$ORG_ID" || "$ORG_ID" == "null" ]]; then
        err "Failed to get organization ID."
    fi

    # Get the default project name
    local projects
    projects=$(curl -sf -H "Authorization: Bearer ${AUTH_TOKEN}" \
        "${API_BASE}/api/v1/organizations/${ORG_ID}/projects")
    PROJECT_NAME=$(echo "$projects" | jq -r '.items[0].name // empty')
    if [[ -z "$PROJECT_NAME" ]]; then
        PROJECT_NAME="default"
    fi

    log "Organization ID: ${ORG_ID}"
    log "Project: ${PROJECT_NAME}"
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
# Step 4: Ensure domain on organisation
# ============================================================
ensure_org_domain() {
    if [[ "$CLOUD_MODE" == "true" ]]; then
        log "Using Cloud platform domain '${ORG_DOMAIN}'."
        return
    fi

    log "Ensuring domain '${ORG_DOMAIN}' on organisation..."
    local org_response
    org_response=$(api GET "/api/v1/organizations/${ORG_ID}")
    local existing_domain
    existing_domain=$(echo "$org_response" | jq -r ".domains[]? | select(.fqdn == \"${ORG_DOMAIN}\") | .fqdn" 2>/dev/null)

    if [[ "$existing_domain" == "$ORG_DOMAIN" ]]; then
        log "Domain '${ORG_DOMAIN}' already configured."
        return
    fi

    local response
    response=$(api PUT "/api/v1/organizations/${ORG_ID}" \
        -d "{\"domains\": [{\"fqdn\": \"${ORG_DOMAIN}\"}]}")
    local domain_count
    domain_count=$(echo "$response" | jq '.domains | length')
    if [[ "$domain_count" -lt 1 ]]; then
        err "Failed to add domain. Response: $response"
    fi
    log "Domain '${ORG_DOMAIN}' added."
}

# ============================================================
# Step 5: Ensure cluster is registered
# ============================================================
ensure_cluster_registered() {
    if [[ "$CLOUD_MODE" == "true" ]]; then
        log "Using the shared-compute cluster bootstrapped by the API server..."
        CLUSTER_ID=$(docker exec -e PGPASSWORD="$DB_PASSWORD" "$PG_CONTAINER_NAME" \
            psql -U "$DB_USERNAME" -d "$DB_NAME" -tAc \
            "SELECT id FROM clusters WHERE shared_compute = TRUE ORDER BY id LIMIT 1" \
            | tr -d '[:space:]')
        if [[ -z "$CLUSTER_ID" ]]; then
            err "Failed to find the bootstrapped shared-compute cluster."
        fi
        log "Shared-compute cluster ready. ID: ${CLUSTER_ID}"
        log "The organisation registry remains pending until a build release needs it."
        return
    fi

    log "Ensuring cluster is registered..."

    # Check if a cluster is already registered
    local clusters_response
    clusters_response=$(api GET "/api/v1/organizations/${ORG_ID}/clusters" 2>/dev/null || echo '{"items":[]}')
    CLUSTER_ID=$(echo "$clusters_response" | jq -r '.items[0].id // empty' 2>/dev/null)

    if [[ -n "$CLUSTER_ID" ]]; then
        log "Cluster already registered. ID: ${CLUSTER_ID}"
        ensure_image_registry
        wait_for_image_registry
        return
    fi

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
    log "Cluster registered with image registry. ID: ${CLUSTER_ID}"
    wait_for_image_registry
}

ensure_image_registry() {
    local registries_response
    registries_response=$(api GET "/api/v1/organizations/${ORG_ID}/clusters/${CLUSTER_ID}/image_registries" 2>/dev/null || echo '{"items":[]}')
    local registry_id
    registry_id=$(echo "$registries_response" | jq -r '.items[0].id // empty' 2>/dev/null)

    if [[ -n "$registry_id" ]]; then
        log "Image registry already exists. ID: ${registry_id}"
        return
    fi

    log "Creating image registry on cluster..."
    local payload
    payload=$(jq -n '{
        name: "local-registry",
        spec: {
            backend_storage_size: "10Gi",
            max_repositories: 100,
            tags_per_repository: 50,
            delete_untagged: true
        }
    }')

    local response
    response=$(api POST "/api/v1/organizations/${ORG_ID}/clusters/${CLUSTER_ID}/image_registries" -d "$payload")
    registry_id=$(echo "$response" | jq -r '.id')
    if [[ -z "$registry_id" || "$registry_id" == "null" ]]; then
        err "Failed to create image registry. Response: $response"
    fi
    log "Image registry created. ID: ${registry_id}"
}

wait_for_image_registry() {
    log "Waiting for image registry to become Running..."
    for i in $(seq 1 30); do
        local registries_response
        registries_response=$(api GET "/api/v1/organizations/${ORG_ID}/clusters/${CLUSTER_ID}/image_registries" 2>/dev/null || echo '{"items":[]}')
        local state
        state=$(echo "$registries_response" | jq -r '.items[0].status.state // empty' 2>/dev/null)

        if [[ "$state" == "ImageRegistryRunning" ]]; then
            log "Image registry is Running."
            return
        fi
        if [[ "$state" == "ImageRegistryError" ]]; then
            warn "Image registry entered ImageRegistryError state. Proceeding anyway."
            return
        fi
        if (( i % 10 == 0 )); then
            info "  Registry state: ${state:-unknown} (${i}s elapsed)..."
        fi
        sleep 1
    done
    warn "Image registry did not reach ImageRegistryRunning within 30s. Proceeding anyway."
}

# ============================================================
# Step 6: Create addon (optional)
# ============================================================
create_addon() {
    if [[ -z "$ADDON_FILE" ]]; then
        return
    fi

    local addon_name
    addon_name=$(jq -r '.name' "$ADDON_FILE")
    log "Creating Postgres addon '${addon_name}'..."

    local response
    response=$(api POST "/api/v1/organizations/${ORG_ID}/projects/${PROJECT_NAME}/addons/postgres" -d @"$ADDON_FILE")
    ADDON_ID=$(echo "$response" | jq -r '.id')
    if [[ -z "$ADDON_ID" || "$ADDON_ID" == "null" ]]; then
        err "Failed to create postgres addon. Response: $response"
    fi
    log "Postgres addon created. ID: ${ADDON_ID}"

    log "Waiting for Postgres addon to become Ready..."
    for i in $(seq 1 30); do
        local state
        state=$(api GET "/api/v1/organizations/${ORG_ID}/projects/${PROJECT_NAME}/addons/postgres/${ADDON_ID}" 2>/dev/null \
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
# Step 7: Deploy stack (optional)
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
        response=$(echo "$stack_payload" | api POST "/api/v1/organizations/${ORG_ID}/projects/${PROJECT_NAME}/stacks" -d @-)
    else
        response=$(api POST "/api/v1/organizations/${ORG_ID}/projects/${PROJECT_NAME}/stacks" -d @"$STACK_FILE")
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
    info "Project:         ${PROJECT_NAME}"
    info "Cluster ID:   ${CLUSTER_ID}"
    info "Auth Token:   ${AUTH_TOKEN}"
    info "Org Domain:   ${ORG_DOMAIN}"
    info "Kubectl ctx:  k3d-${K3D_CLUSTER_NAME}"

    local project_base="${API_BASE}/api/v1/organizations/${ORG_ID}/projects/${PROJECT_NAME}"

    # Append API-level config to dev_env.yaml
    # Remove old API-level entries first to avoid duplicates on re-runs
    local tmp_config
    tmp_config=$(grep -v '^api_server:\|^org_id:\|^project_name:\|^cluster_id:\|^auth_token:\|^org_domain:\|^admin_email:\|^admin_password:' "$DEV_CONFIG" 2>/dev/null || true)
    echo "$tmp_config" > "$DEV_CONFIG"
    cat >> "$DEV_CONFIG" <<EOF
api_server: ${API_BASE}
org_id: ${ORG_ID}
project_name: ${PROJECT_NAME}
cluster_id: ${CLUSTER_ID}
auth_token: ${AUTH_TOKEN}
org_domain: ${ORG_DOMAIN}
admin_email: ${ADMIN_EMAIL}
admin_password: ${ADMIN_PASS}
EOF
    log "Config saved to ${DEV_CONFIG}"

    log ""
    log "Useful commands:"
    log ""
    log "  # Deploy a stack"
    log "  curl -s -X POST -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
    log "    -H 'Content-Type: application/json' \\"
    log "    ${project_base}/stacks -d @samples/tooljet.json | jq"
    log ""
    log "  # Create a postgres addon"
    log "  curl -s -X POST -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
    log "    -H 'Content-Type: application/json' \\"
    log "    ${project_base}/addons/postgres -d @samples/postgres_addon_basic.json | jq"
    log ""
    log "  # Export kubeconfig"
    log "  k3d kubeconfig get ${K3D_CLUSTER_NAME} > /tmp/k3d-kubeconfig.yaml"
    log ""
    log "  # Watch pods in the cluster"
    log "  kubectl --context k3d-${K3D_CLUSTER_NAME} get pods -A -w"
    log ""
    log "  # Tear down everything"
    log "  mage dev:teardown"
    log ""
}

# ============================================================
# Print stack/addon info
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

    local project_base="${API_BASE}/api/v1/organizations/${ORG_ID}/projects/${PROJECT_NAME}"

    log ""
    log "Useful commands:"
    log ""
    log "  # List stacks"
    log "  curl -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
    log "    ${project_base}/stacks | jq '.items[].name'"
    log ""
    log "  # List postgres addons"
    log "  curl -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
    log "    ${project_base}/addons/postgres | jq '.items[] | {name, id, state: .status.state}'"

    if [[ -n "${STACK_ID:-}" ]]; then
        log ""
        log "  # Check stack status"
        log "  curl -s -H 'Authorization: Bearer ${AUTH_TOKEN}' \\"
        log "    ${project_base}/stacks/${STACK_ID} | jq '.status'"
    fi
    log ""
}

# ============================================================
# Main
# ============================================================
main() {
    log "Starting local Stackdome environment"
    if [[ "$CLOUD_MODE" == "true" ]]; then
        info "Mode: stackdome_cloud + shared"
    else
        info "Mode: self_hosted + bring_your_own"
    fi
    if [[ -n "$STACK_FILE" ]]; then
        info "Stack file: ${STACK_FILE}"
    fi
    if [[ -n "$ADDON_FILE" ]]; then
        info "Addon file: ${ADDON_FILE}"
    fi
    log ""

    check_prerequisites
    setup_infra
    read_dev_config
    start_api_server
    signup_and_authenticate
    ensure_org_domain
    ensure_cluster_registered
    print_info

    if [[ -n "$ADDON_FILE" || -n "$STACK_FILE" ]]; then
        create_addon
        deploy_stack
        print_stack_info
    fi

    log "Press Ctrl+C to stop the API server."
    log "Infrastructure (k3d + PostgreSQL) will remain running."
    log "To tear down everything: mage dev:teardown"
    wait "$API_SERVER_PID"
}

main
