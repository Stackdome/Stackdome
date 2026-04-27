# Environment Variable Interpolation

Stackdome supports Go template-based interpolation in environment variable values, allowing stack resources to dynamically reference each other's internal addresses and public URLs without hardcoding hostnames or FQDNs.

---

## Template Syntax

Templates use Go's `text/template` syntax with double curly braces:

```
{{ STACKDOME_<RESOURCE_NAME>_<TYPE> }}
```

Resource names are uppercased and hyphens are converted to underscores. For example, a resource named `api-service` becomes `API_SERVICE` in the template function name.

---

## Available Template Functions

### Internal Service Address

```
{{ STACKDOME_<RESOURCE>_INTERNAL }}
```

Resolves to the Kubernetes internal DNS name of the resource's Service (e.g., `webapp.default.svc.cluster.local`). Available for any resource that has at least one port defined.

**Example:**
```json
{
    "name": "DATABASE_HOST",
    "value": "{{ STACKDOME_POSTGRES_INTERNAL }}"
}
```

### Public URL (Single Port)

```
{{ STACKDOME_<RESOURCE>_PUBLIC }}
```

Resolves to the public ingress URL of the resource. Only available when the resource has **exactly one** port with `exposed_to_public: true`. If the resource has multiple public ports, this function is not registered and will produce a validation error.

**Example:**
```json
{
    "name": "APP_BASE_URL",
    "value": "https://{{ STACKDOME_WEBAPP_PUBLIC }}"
}
```

### Public URL (Port-Specific)

```
{{ STACKDOME_<RESOURCE>_PUBLIC_<PORT> }}
```

Resolves to the public ingress URL for a specific port. Available for every port that has `exposed_to_public: true`. This is the only way to reference public URLs when a resource exposes multiple ports.

**Example:**
```json
{
    "name": "API_URL",
    "value": "https://{{ STACKDOME_API_SERVICE_PUBLIC_8000 }}"
}
```

### Self-Referencing

A resource can reference its own template functions. For example, a resource named `webapp` can use `{{ STACKDOME_WEBAPP_PUBLIC }}` in its own environment variables to discover its own public URL at runtime.

---

## How It Works

Interpolation is processed at two stages: validation at the API server and resolution at the cluster agent.

### Stage 1: API Server Validation (Create/Update Time)

When a stack is created or updated, the API server validates all interpolation templates before persisting the stack.

**Flow:**

1. The API receives the stack spec with `environment_variables` in the OpenAPI schema.
2. The presenter layer converts these to the internal model (`ExecutionConfig.Env`) via `convertExecutionConfig()` in `pkg/presenters/stack.go:446`.
3. The stack validator calls `ValidateStackInterpolations()` in `pkg/validator/stack/stack_validator.go:323`.
4. The interpolation validator builds a validation context from all stack resources (`pkg/validator/stack/interpolation_validator.go:37-64`):
   - Each resource's name is registered as the `InternalService` address.
   - Each public port is registered with a placeholder URL (`"validation_url"`).
   - This means all template functions that will exist at runtime are also available during validation.
5. Every env var value in every resource is passed through `interpolator.InterpolateString()`. If any template references a non-existent resource or uses invalid syntax, the stack creation is rejected with a `400 Bad Request`.

**What validation catches:**
- References to resources that don't exist in the stack (e.g., `{{ STACKDOME_NONEXISTENT_INTERNAL }}`)
- Syntax errors (e.g., unclosed `{{`, bad characters like hyphens inside templates)
- References to `_PUBLIC` on resources with multiple public ports (must use `_PUBLIC_<PORT>`)
- References to `_PUBLIC` on resources with no public ports

**What validation does NOT check:**
- Whether the actual FQDN or internal address will be resolvable (those are only known at runtime)
- Whether the referenced resource will actually be ready when the env var is needed

**Key implementation detail:** The validation context includes ALL resources in the stack, including the resource being validated. This enables self-referencing patterns like a resource discovering its own public URL.

### Stage 2: Cluster Agent Resolution (Runtime)

When the cluster agent reconciles a StackResource CR into a Deployment or StatefulSet, it resolves templates to actual addresses.

**Flow:**

1. The workload reconciler in the cluster agent receives a StackResource CR to reconcile (`cluster-agent/internal/controller/stackresource/workload_reconciler.go:125`).
2. It fetches all sibling StackResources in the same namespace via `GetSiblings()` (line 408-426).
3. For the current resource, it explicitly sets `InternalAddress` to the resource's Service name (`ResourceSVCName(resource)`) — this ensures self-references resolve correctly even before the Service is fully reconciled.
4. An `InterpolationContext` is built from all siblings using `interpolation.NewInterpolationContext()` (`cluster-agent/pkg/interpolation/interpolation_context.go:28`):
   - `InternalService` is set to each resource's `Status.InternalAddress` (the K8s Service DNS name).
   - `PublicIngresses` are built from each resource's port spec, using the `FQDN` field.
5. Each env var value is interpolated via `interpolator.InterpolateString()`.
6. The resolved env vars are set on the container spec of the resulting Deployment/StatefulSet.

**Where runtime values come from:**
- **Internal address**: Set by the service reconciler after creating the K8s Service (`cluster-agent/internal/controller/stackresource/svc_reconciler.go:41`). The value is the Service's `metadata.name`.
- **Public FQDN**: Set by the API server when creating the CR. The stack domain service generates FQDNs using the pattern `<subdomain>.<resource_name>.<org_domain>` (`pkg/services/stack_domain_service.go:137-142`) and the builder sets `Port.FQDN` on the CR spec (`pkg/builders/cluster_resource_builder.go:385`).

---

## FQDN Generation

When a port has `exposed_to_public: true`, the API server generates an FQDN before creating the CR in the cluster.

**Format:**
```
<subdomain_prefix>.<resource_name>.<org_domain>
```

- If `subdomain_prefix` is set on the port spec, it is used directly.
- If not set, a generated prefix is created from the encoded stack resource ID and port number.
- `<org_domain>` comes from the organization's configured domain.

**Example:** A resource named `webapp` with `subdomain_prefix: "app"` in an org with domain `example.com` gets:
```
app.webapp.example.com
```

The FQDN is:
1. Stored in the database as `ExposedFqdn` on the port.
2. Set as `Port.FQDN` on the StackResource CR spec.
3. Used by the cluster agent's service reconciler to create an Ingress rule with that hostname.
4. Available for interpolation as `{{ STACKDOME_WEBAPP_PUBLIC }}`.

---

## Naming Conventions

| Resource Name | Template Prefix |
|---|---|
| `postgres` | `STACKDOME_POSTGRES_` |
| `api-service` | `STACKDOME_API_SERVICE_` |
| `web-app` | `STACKDOME_WEB_APP_` |
| `my-db-v2` | `STACKDOME_MY_DB_V2_` |

The conversion rule: uppercase the resource name and replace all hyphens with underscores.

---

## Complete Reference

| Template | Resolves To | Availability |
|---|---|---|
| `{{ STACKDOME_<R>_INTERNAL }}` | K8s Service DNS name | Any resource with ports |
| `{{ STACKDOME_<R>_PUBLIC }}` | Public ingress URL | Resource with exactly 1 public port |
| `{{ STACKDOME_<R>_PUBLIC_<PORT> }}` | Public ingress URL for port | Each public port on a resource |

---

## Error Handling

### Validation Errors (API Server)

| Error | Cause |
|---|---|
| `Resource reference 'STACKDOME_X_INTERNAL' is not available` | Referenced resource doesn't exist in the stack |
| `Resource reference 'STACKDOME_X_PUBLIC' is not available` | Resource has no public ports, or has multiple public ports |
| `Resource reference 'STACKDOME_X_PUBLIC_9090' is not available` | Port 9090 is not exposed on the resource |
| `Invalid template: Make sure all '{{' have matching '}}' pairs` | Syntax error — unclosed braces |
| `Invalid template: Contains characters that aren't allowed` | Bad characters in template (e.g., hyphens) |

### Runtime Errors (Cluster Agent)

At runtime, if interpolation fails (e.g., a sibling resource's Service doesn't exist yet), the workload reconciler returns an error and the resource will be requeued for retry. The `InterpolateEnvVars` method preserves original values on error, so partial failures don't corrupt other env vars.

---

## Examples

### Internal cross-service reference

Resource `backend` connecting to resource `postgres`:

```json
{
    "name": "DATABASE_URL",
    "value": "postgres://user:pass@{{ STACKDOME_POSTGRES_INTERNAL }}:5432/mydb"
}
```

At runtime resolves to: `postgres://user:pass@postgres-svc.namespace.svc.cluster.local:5432/mydb`

### Self-referencing public URL

Resource `webapp` discovering its own public URL:

```json
{
    "name": "APP_BASE_URL",
    "value": "https://{{ STACKDOME_WEBAPP_PUBLIC }}"
}
```

At runtime resolves to: `https://app.webapp.example.com`

### Port-specific public URL (multiple public ports)

Resource `api-service` with ports 8000 and 9000 both public:

```json
{
    "name": "API_URL",
    "value": "{{ STACKDOME_API_SERVICE_PUBLIC_8000 }}"
},
{
    "name": "ADMIN_URL",
    "value": "{{ STACKDOME_API_SERVICE_PUBLIC_9000 }}"
}
```

### Combining multiple references

```json
{
    "name": "CONFIG",
    "value": "db={{ STACKDOME_POSTGRES_INTERNAL }}:5432,redis={{ STACKDOME_REDIS_INTERNAL }}:6379,public={{ STACKDOME_APP_PUBLIC }}"
}
```

---

## Source Files

| Component | File |
|---|---|
| Interpolation engine | `cluster-agent/pkg/interpolation/resource_interpolation.go` |
| Interpolation context | `cluster-agent/pkg/interpolation/interpolation_context.go` |
| Interpolation tests | `cluster-agent/pkg/interpolation/interpolation_test.go` |
| API server validation | `api-server/pkg/validator/stack/interpolation_validator.go` |
| Runtime resolution | `cluster-agent/internal/controller/stackresource/workload_reconciler.go:304-325` |
| FQDN generation | `api-server/pkg/services/stack_domain_service.go:114-171` |
| FQDN → CR propagation | `api-server/pkg/builders/cluster_resource_builder.go:375-389` |
| Presenter mapping | `api-server/pkg/presenters/stack.go:446` |
