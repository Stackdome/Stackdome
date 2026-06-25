# Stack Topology and Connections API Design

## Summary

Stackdome should make Stack topology explicit. A Stack is not only a list of StackResources; it is a graph of resources, addons, secrets, volumes, object stores, and the relationships between them.

This spec introduces a user-facing API for typed connections between topology nodes. Connections are the canonical expression of user intent. Explicit wiring is persisted in `stack_connections`; `ResourceUsage` is reserved for non-connection config references such as build/image secrets and addon backup settings.

The first implementation should focus on Stack-scoped topology:

- StackResources with named ports and typed outputs.
- Connections with explicit `kind`, `from`, `to`, and mapping/config.
- A structured value reference DSL for reading outputs from connected resources.
- Topology read APIs for a node-graph canvas UI.
- Usage views derived from `stack_connections` on read, plus persisted `ResourceUsage` records for direct config references.

## Goals

- Let users create a Stack on a canvas and connect resources together.
- Make resource wiring visible, inspectable, and editable.
- Support more than env var injection: env, secret mounts, and volume mounts.
- Support StackResources exposing multiple ports, each with its own public FQDN.
- Provide stable, typed accessors for reading values from other resources.
- Keep the cluster agent mostly unchanged by resolving connection-driven env vars in the API server.
- Replace implicit environment interpolation with explicit connections.

## Non-Goals

- Cross-stack connections in the first phase.
- Cross-cluster connections in the first phase.
- Making `ResourceUsage` user-editable.
- Building the full canvas UI in this spec.

## Core Model

### Direct Configuration vs Topology Relationships

The redesign does not turn every StackResource field into a connection. A StackResource still owns its intrinsic workload configuration:

- `labels`
- `annotations`
- `image_spec`
- `build_spec`
- `init_spec`
- `execution_config.command`
- `execution_config.args`
- literal environment variables
- `lifecycle_config`
- `stateful` / workload type
- `ports`

Connections model relationships between topology nodes. Intrinsic workload config remains on the StackResource.

Do not shoehorn every reference into a canvas connection. A field should become a `Connection` only when the relationship is meaningful as user-facing topology that users should edit as an edge: runtime wiring, mounts, startup ordering, or another relationship that naturally belongs on the canvas. Configuration details that belong to one resource, such as build credentials, image pull credentials, or Postgres backup object-store settings, should stay inside that resource's config while still producing derived `ResourceUsage` records and topology metadata for delete protection and impact analysis.

Some current fields are relationships hidden inside StackResource config. Those should either become explicit connections or produce derived topology edges:

| Current capability | New canonical expression |
|---|---|
| Literal env var | StackResource `execution_config.env[]` literal value |
| Env var self interpolation | StackResource `execution_config.env[]` with `self_output` |
| Env var referencing another StackResource | `kind=env` connection from producer StackResource to consumer StackResource |
| `env_from_addons` | `kind=env` connection from addon to StackResource |
| `environment_variables_from_secret` | `kind=env` connection from Secret to StackResource |
| `volume_mounts` | `kind=volume_mount` connection from Volume to StackResource |
| `depends_on` | Keep on StackResource; expose as derived topology edge |
| Image pull secret | Keep on `image_spec`; derive `ResourceUsage` and topology metadata |
| Build git secret / registry push secret | Keep on `build_spec`; derive `ResourceUsage` and topology metadata |
| Build from Git | Keep in `build_spec.source_context.git_repo` |
| Build from Volume | Keep in `build_spec.source_context.volume`; derive topology edge |
| Volume seeded from build artifact | `kind=build_artifact_source` connection from StackResource to Volume |
| Volume sourced from Git/remote dir | Keep in Volume source config |

The first implementation should avoid making the API too abstract. Connections are for resource relationships users see on the canvas; image/build/runtime knobs remain ordinary resource fields.

### Stack

`StackSpec` gains a top-level `connections` array.

```yaml
StackSpec:
  type: object
  required:
    - stack_resources
  properties:
    stack_resources:
      type: array
      items:
        $ref: "#/components/schemas/StackResource"
    volumes:
      type: array
      items:
        $ref: "#/components/schemas/Volume"
    connections:
      type: array
      items:
        $ref: "#/components/schemas/StackConnection"
```

Connections live on the Stack spec so import/export, create/update, template expansion, and review workflows can see the full desired topology in one document.

Independent connection CRUD endpoints may patch this field under the hood.

### StackResource Ports

Ports should be named. The current API identifies ports mostly by number; the redesigned API should use stable names for outputs, UI handles, and connection mappings.

A StackResource can expose a public URL without being connected to another resource. This is common for frontend services, API gateways, dashboards, webhooks, and any resource that serves traffic directly to users. Public exposure is therefore modeled on the port itself, not as a connection.

```json
{
  "name": "api",
  "image_spec": { "image": "example/api:latest" },
  "ports": [
    {
      "name": "http",
      "number": 8080,
      "protocol": "http",
      "public": {
        "enabled": true,
        "subdomain_prefix": "api"
      }
    },
    {
      "name": "metrics",
      "number": 9090,
      "protocol": "http",
      "public": {
        "enabled": true,
        "subdomain_prefix": "api-metrics"
      }
    },
    {
      "name": "grpc",
      "number": 50051,
      "protocol": "grpc",
      "public": {
        "enabled": false
      }
    }
  ]
}
```

The API server resolves public FQDNs for exposed ports and returns them as read-only fields.

```json
{
  "name": "http",
  "number": 8080,
  "protocol": "http",
  "public": {
    "enabled": true,
    "subdomain_prefix": "api",
    "fqdn": "api.api.example.com",
    "url": "http://api.api.example.com"
  }
}
```

Validation rules:

- `ports[].name` is required in the new API shape.
- Port names are unique within a StackResource.
- Port names use a DNS-label-like format: lowercase letters, numbers, and hyphens.
- `ports[].number` is unique within a StackResource.
- `public.enabled=true` requires a protocol that Stackdome can expose.
- Each public port gets a distinct FQDN.
- If `subdomain_prefix` is omitted, the API server generates one.

Public URL rules:

- A StackResource may have zero, one, or many public ports.
- Each public port gets its own public FQDN and URL.
- A public URL is an output of the StackResource: `public.<port_name>.host` and `public.<port_name>.url`.
- Public URLs use `http://` until Stackdome implements TLS certificate provisioning for public endpoints.
- Public exposure does not require a `Connection`.
- The topology API should mark public ports/endpoints on the node so the canvas can show which resources receive external traffic.
- Future topology may add synthetic external nodes or edges for inbound traffic, but the canonical user-authored public exposure remains the port's `public` config.

Migration/presentation rule:

- Existing unnamed ports may be presented with generated names like `port-8080`.
- Generated port names should be stable for existing resources unless the user explicitly renames the port.

## Topology Nodes

A topology node is anything that can appear on the Stack canvas.

```yaml
TopologyNodeRef:
  type: object
  required:
    - type
  properties:
    type:
      type: string
      enum:
        - stack_resource
        - addon/postgres
        - secret
        - volume
        - object_store
    name:
      type: string
      description: Name-scoped reference, used for Stack-local resources like StackResource and Volume.
    id:
      type: string
      description: Stable ID reference, used for persisted external resources like addons, secrets, and object stores.
```

Reference rules:

- Same-stack `stack_resource` references use `name`.
- Same-stack `volume` references may use `name`.
- Addons, secrets, and object stores use stable `id`.
- API responses may include both `id` and `name`/`label` for display.
- The first phase only accepts nodes in the same Stack, Team, and Organisation boundary.

Example refs:

```json
{ "type": "stack_resource", "name": "web" }
{ "type": "stack_resource", "name": "redis" }
{ "type": "addon/postgres", "id": "pg_123" }
{ "type": "secret", "id": "sec_123" }
{ "type": "volume", "name": "uploads" }
{ "type": "object_store", "id": "os_123" }
```

## Outputs and Accessors

Outputs are typed values that a node exposes to consumers. Connections read outputs from the `from` node.

### StackResource Outputs

Every StackResource exposes these automatic outputs:

```text
host
port.<port_name>
url.<port_name>
public.<port_name>.host
public.<port_name>.url
```

Meaning:

| Accessor | Type | Description |
|---|---|---|
| `host` | string | Internal service DNS/name for the StackResource. |
| `port.<port_name>` | integer | Numeric port for the named port. |
| `url.<port_name>` | string | Internal URL using protocol, host, and port. |
| `public.<port_name>.host` | string | Public FQDN for an exposed port. |
| `public.<port_name>.url` | string | Public URL for an exposed port. |

Example for resource `api`:

```json
{
  "host": "api",
  "port.http": 8080,
  "url.http": "http://api:8080",
  "public.http.host": "api.api.example.com",
  "public.http.url": "http://api.api.example.com",
  "port.metrics": 9090,
  "url.metrics": "http://api:9090",
  "public.metrics.host": "api-metrics.api.example.com",
  "public.metrics.url": "http://api-metrics.api.example.com"
}
```

Availability rules:

- `host` exists for every StackResource.
- `port.<port_name>` exists for each declared port.
- `url.<port_name>` exists for each declared port with a supported protocol.
- `public.<port_name>.host` and `public.<port_name>.url` exist only when that port has `public.enabled=true`.

### Postgres Addon Outputs

PostgresAddon exposes platform-defined outputs:

```text
host
port
database
username
password
sslmode
ca_certificate
url
```

Some outputs are secret values. The API must track output sensitivity and avoid leaking secret values in normal topology responses.

### Secret Outputs

Secrets expose keys as outputs:

```text
key.<key_name>
```

Example:

```text
key.SECRET_KEY_BASE
key.JWT_PRIVATE_KEY
```

Secret accessor rules:

- Preserve the original secret key name.
- Simple keys may use dot syntax: `key.JWT_PRIVATE_KEY`.
- Keys containing dots, slashes, spaces, quotes, or other special characters use bracket syntax with single quotes: `key['tls.crt']`.
- Bracket syntax still preserves the original key exactly.
- The API must reject ambiguous or invalid accessors rather than normalizing keys.

Secret output values are sensitive. Topology APIs should expose key names and sensitivity metadata, not raw values.

### ObjectStore Outputs

ObjectStore outputs depend on provider, but the common output surface should start with:

```text
bucket
endpoint
region
access_key_id
secret_access_key
url
```

Sensitive fields must be marked sensitive.

### Volume Outputs

Volumes should initially expose minimal outputs:

```text
name
```

Mount paths are target-specific and belong in the `volume_mount` connection config, not in the Volume output surface.

## Connections

A `StackConnection` is a user-authored topology edge.

```yaml
StackConnection:
  type: object
  required:
    - kind
    - from
    - to
  properties:
    id:
      type: string
      description: Stable connection ID. Generated when omitted.
    kind:
      type: string
      enum:
        - env
        - volume_mount
        - build_artifact_source
    from:
      $ref: "#/components/schemas/TopologyNodeRef"
    to:
      $ref: "#/components/schemas/TopologyNodeRef"
    mappings:
      type: array
      items:
        $ref: "#/components/schemas/ConnectionMapping"
    config:
      type: object
      additionalProperties: true
```

Design rules:

- A connection has exactly one `from` node and one `to` node.
- A connection may have multiple mappings.
- One target per connection maps cleanly to one canvas edge.
- Fan-out is represented as multiple connections.
- `kind` defines the mechanism of use.
- Connections are canonical user intent.
- Connection-backed usage is read from `stack_connections`; non-connection config usage is indexed in `ResourceUsage`.

### Env Connection

Injects one or more values as environment variables into a StackResource.

```json
{
  "id": "conn_postgres_web",
  "kind": "env",
  "from": { "type": "addon/postgres", "id": "pg_123" },
  "to": { "type": "stack_resource", "name": "web" },
  "mappings": [
    {
      "target": { "type": "env", "name": "DATABASE_URL" },
      "value": { "output": "url" }
    },
    {
      "target": { "type": "env", "name": "PGHOST" },
      "value": { "output": "host" }
    }
  ]
}
```

Postgres addon env connections must preserve the current `env_from_addons` capability to select a database and request superuser credentials.

```json
{
  "id": "conn_postgres_tooljet",
  "kind": "env",
  "from": { "type": "addon/postgres", "id": "pg_123" },
  "to": { "type": "stack_resource", "name": "tooljet" },
  "config": {
    "database": "tooljet",
    "superuser": false
  },
  "mappings": [
    {
      "target": { "type": "env", "name": "PG_HOST" },
      "value": { "output": "host" }
    },
    {
      "target": { "type": "env", "name": "PG_PORT" },
      "value": { "output": "port" }
    },
    {
      "target": { "type": "env", "name": "PG_USER" },
      "value": { "output": "username" }
    },
    {
      "target": { "type": "env", "name": "PG_PASS" },
      "value": { "output": "password" }
    },
    {
      "target": { "type": "env", "name": "PG_DB" },
      "value": { "output": "database" }
    }
  ]
}
```

Resolution:

- The API server resolves env mappings during Stack create/update and connection CRUD.
- Resolved env vars are injected into the StackResource model/CR payload.
- The cluster agent receives plain env vars.
- Sensitive values should be represented as secret refs where possible, not copied as plaintext.

Collision behavior:

- If a literal env var and a connection-injected env var use the same name, validation should reject the Stack unless the request explicitly chooses an override policy.
- The first phase should prefer rejection over warning to avoid surprising runtime behavior.

### Secret File Mounts

Secret file mounts are deferred. In the current API, secrets are consumed through
`kind=env` connections only.

### Volume Mount Connection

Mounts a Volume into a StackResource.

```json
{
  "kind": "volume_mount",
  "from": { "type": "volume", "name": "uploads" },
  "to": { "type": "stack_resource", "name": "web" },
  "config": {
    "mount_path": "/uploads",
    "sub_path": "",
    "read_only": false
  }
}
```

Volume mount path is connection config because it is a property of the relationship, not the Volume itself.

### Build Artifact Source Connection

The current Volume API can seed a Volume from build artifacts produced by a StackResource. That relationship should be visible in topology.

```json
{
  "kind": "build_artifact_source",
  "from": { "type": "stack_resource", "name": "web" },
  "to": { "type": "volume", "name": "assets" },
  "config": {
    "source_path": "/app/public",
    "destination_path": "/"
  }
}
```

This replaces `volume.spec.source.build_source[]` as the public API shape. The implementation may still lower the connection into whatever internal CRD shape the cluster agent needs. Since this is a relationship between two canvas nodes, the connection shape is the canonical API.

### Derived Depends-On Edges

`depends_on` remains a field on StackResource in this design. It is startup/readiness ordering, not value wiring, and it already belongs naturally to the resource that waits.

```json
{
  "name": "web",
  "depends_on": ["redis"]
}
```

The topology API should expose each dependency as a derived edge:

```json
{
  "kind": "depends_on",
  "from": { "type": "stack_resource", "name": "redis" },
  "to": { "type": "stack_resource", "name": "web" },
  "source_of_truth": "stack_resource.depends_on"
}
```

Rules:

- `to` waits for `from`.
- `depends_on` must be acyclic.
- `depends_on` does not expose outputs.
- `depends_on` does not inject env vars.
- `depends_on` does not imply network traffic.

## Value Reference DSL

Connection mappings use a structured `ValueRef`.

Unless this section explicitly says otherwise, the examples below are entries inside a `StackConnection.mappings[]` array.

Full context:

```json
{
  "kind": "env",
  "from": { "type": "addon/postgres", "id": "pg_123" },
  "to": { "type": "stack_resource", "name": "web" },
  "mappings": [
    {
      "target": { "type": "env", "name": "DATABASE_URL" },
      "value": { "output": "url" }
    }
  ]
}
```

In this example:

- `from` says which node produces values.
- `to` says which node receives the value.
- `target` says where to put the value on the receiving node.
- `value` says which output to read from the producing node.

```yaml
ConnectionMapping:
  type: object
  required:
    - target
    - value
  properties:
    target:
      $ref: "#/components/schemas/ConnectionTarget"
    value:
      $ref: "#/components/schemas/ValueRef"
```

### Direct Output Reference

Inside a connection, `output` is read from the connection's `from` node.

```json
{
  "kind": "env",
  "from": { "type": "stack_resource", "name": "redis" },
  "to": { "type": "stack_resource", "name": "web" },
  "target": { "type": "env", "name": "REDIS_URL" },
  "value": { "output": "url.redis" }
}
```

This means: set `REDIS_URL` on `web` using `redis.outputs["url.redis"]`. The mapping does not repeat `redis` because the connection already names the producer in `from`.

For StackResource ports:

```json
{ "output": "host" }
{ "output": "port.http" }
{ "output": "url.http" }
{ "output": "public.http.host" }
{ "output": "public.http.url" }
```

For Postgres:

```json
{ "output": "url" }
{ "output": "host" }
{ "output": "password" }
```

### Self Output Reference

A StackResource may need one of its own generated values as an environment variable. Common examples:

- A frontend needs `PUBLIC_URL` set to its own public URL.
- An API needs `BASE_URL` set to its own public URL for callbacks.
- A service needs its own internal URL for self-registration.

These should not require a self-connection on the topology canvas. A self-reference is local StackResource config.

```json
{
  "name": "frontend",
  "ports": [
    {
      "name": "http",
      "number": 3000,
      "protocol": "http",
      "public": {
        "enabled": true,
        "subdomain_prefix": "app"
      }
    }
  ],
  "execution_config": {
    "env": [
      {
        "name": "PUBLIC_URL",
        "value": {
          "self_output": "public.http.url"
        }
      }
    ]
  }
}
```

Rules:

- `self_output` is only valid inside a StackResource's own config.
- `self_output` uses the same accessor names as normal StackResource outputs.
- `self_output` is resolved by the API server after port/FQDN resolution.
- Invalid self outputs fail validation. For example, `public.http.url` is invalid if port `http` is not public.
- Self references should appear in topology as node annotations or output usage metadata, not as canvas edges.

### Literal Environment Values

Literal environment variables remain supported. The new value DSL augments literals; it does not remove them.

```json
{
  "execution_config": {
    "env": [
      {
        "name": "NODE_ENV",
        "value": "production"
      },
      {
        "name": "PUBLIC_URL",
        "value": { "self_output": "public.http.url" }
      }
    ]
  }
}
```

The API should model env var values as either a literal string or a structured `ValueRef`.

### Template Value

Templates are used when the target value must be composed from multiple outputs rather than copied from one output directly.

The common case should be a direct output reference:

```json
{
  "kind": "env",
  "from": { "type": "addon/postgres", "id": "pg_123" },
  "to": { "type": "stack_resource", "name": "web" },
  "config": {
    "database": "app"
  },
  "mappings": [
    {
      "target": { "type": "env", "name": "DATABASE_URL" },
      "value": { "output": "url" }
    }
  ]
}
```

Use a template only when the producer does not expose the exact value the consumer needs. For example, if a Postgres addon exposes `username`, `password`, `host`, `port`, and `database`, but does not expose a ready-made URL, the connection can compose one:

```json
{
  "kind": "env",
  "from": { "type": "addon/postgres", "id": "pg_123" },
  "to": { "type": "stack_resource", "name": "web" },
  "config": {
    "database": "app"
  },
  "mappings": [
    {
      "target": { "type": "env", "name": "DATABASE_URL" },
      "value": {
        "template": "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}",
        "values": {
          "username": { "output": "username" },
          "password": { "output": "password" },
          "host": { "output": "host" },
          "port": { "output": "port" },
          "database": { "output": "database" }
        }
      }
    }
  ]
}
```

Templates are evaluated by the API server while resolving a mapping. They are not used to discover topology, and they are not free-form interpolation inside arbitrary strings. Topology still comes from the connection's `from` and `to`; the template only describes how to construct one target value.

Template rules:

- Template variables must be declared in `values`.
- Template variables may only use simple identifiers like `host`, `port`, `password`.
- Templates cannot directly reference arbitrary resources by name.
- The backend can statically validate all referenced outputs.
- If any referenced value is sensitive, the resulting value is sensitive.

### Future Explicit Source References

The first phase keeps value refs relative to the connection's `from` node. If value refs are later allowed outside `connections`, add explicit source references:

```json
{
  "from": { "type": "stack_resource", "name": "api" },
  "output": "public.http.url"
}
```

Do not introduce this broader form until there is a real field outside connections that needs it.

## Topology API

Topology is a read model built from:

- User-authored `connections`.
- StackResources and their ports/outputs.
- Volumes attached to the Stack.
- Addons, secrets, and object stores referenced by the Stack.
- Derived `ResourceUsage` for reverse lookup and dependency indexing.

### Endpoints

```text
GET    /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{stack_id}/topology
GET    /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{stack_id}/connections
POST   /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{stack_id}/connections
PUT    /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{stack_id}/connections/{connection_id}
DELETE /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{stack_id}/connections/{connection_id}
GET    /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{stack_id}/connectable-nodes
GET    /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{stack_id}/outputs?node_type=stack_resource&name=redis
GET    /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{stack_id}/outputs?node_type=addon/postgres&id=pg_123
```

### Topology Response

```json
{
  "nodes": [
    {
      "ref": { "type": "stack_resource", "name": "web" },
      "id": "sr_123",
      "label": "web",
      "status": "Ready",
      "ports": [
        {
          "name": "http",
          "number": 3000,
          "protocol": "http",
          "public": {
            "enabled": true,
            "fqdn": "web.example.com",
            "url": "http://web.example.com"
          }
        }
      ],
      "outputs": [
        { "name": "host", "type": "string", "sensitive": false },
        { "name": "port.http", "type": "integer", "sensitive": false },
        { "name": "url.http", "type": "string", "sensitive": false },
        { "name": "public.http.host", "type": "string", "sensitive": false },
        { "name": "public.http.url", "type": "string", "sensitive": false }
      ]
    },
    {
      "ref": { "type": "addon/postgres", "id": "pg_123" },
      "label": "postgres",
      "status": "Ready",
      "outputs": [
        { "name": "host", "type": "string", "sensitive": false },
        { "name": "port", "type": "integer", "sensitive": false },
        { "name": "database", "type": "string", "sensitive": false },
        { "name": "username", "type": "string", "sensitive": true },
        { "name": "password", "type": "string", "sensitive": true },
        { "name": "url", "type": "string", "sensitive": true }
      ]
    }
  ],
  "edges": [
    {
      "id": "conn_postgres_web",
      "kind": "env",
      "from": { "type": "addon/postgres", "id": "pg_123" },
      "to": { "type": "stack_resource", "name": "web" },
      "mappings": [
        {
          "target": { "type": "env", "name": "DATABASE_URL" },
          "value": { "output": "url" }
        }
      ],
      "source_of_truth": "connection"
    }
  ]
}
```

`source_of_truth` values:

```text
connection
resource_usage
observed_traffic
```

This lets the UI distinguish user-authored canvas edges from derived dependency-index edges and future observed-traffic edges.

## Connectable Nodes API

The canvas needs to know which nodes can be added or connected.

```json
{
  "items": [
    {
      "ref": { "type": "stack_resource", "name": "web" },
      "label": "web",
      "connectable_as_from": true,
      "connectable_as_to": true,
      "allowed_connection_kinds": ["env"]
    },
    {
      "ref": { "type": "addon/postgres", "id": "pg_123" },
      "label": "postgres",
      "connectable_as_from": true,
      "connectable_as_to": true,
      "allowed_connection_kinds": ["env"]
    },
    {
      "ref": { "type": "secret", "id": "sec_123" },
      "label": "app-secret",
      "connectable_as_from": true,
      "connectable_as_to": false,
      "allowed_connection_kinds": ["env"]
    }
  ]
}
```

## Outputs API

```text
GET /stacks/{stack_id}/outputs?node_type=stack_resource&name=api
```

```json
{
  "node": { "type": "stack_resource", "name": "api" },
  "outputs": [
    {
      "name": "host",
      "type": "string",
      "sensitive": false,
      "available": true
    },
    {
      "name": "port.http",
      "type": "integer",
      "sensitive": false,
      "available": true
    },
    {
      "name": "public.http.url",
      "type": "string",
      "sensitive": false,
      "available": true
    }
  ]
}
```

The outputs API should not return secret values. It returns metadata so the UI can build mapping controls.

## ResourceUsage Relationship

`ResourceUsage` is not user-facing source of truth. It is a persisted dependency index for references that are not represented as explicit connections.

Canonical sources:

- `stack_connections` for explicit user-authored wiring.
- `resource_usages` for direct StackResource image/build credential references.
- `resource_usages` for PostgresAddon backup/restore object-store config.

Usage views can union connection-backed usage with direct config usage:

```json
{
  "resource_type": "addon/postgres",
  "resource_id": "pg_123",
  "consumer_type": "stack_resource",
  "consumer_id": "sr_web"
}
```

```json
{
  "resource_type": "secret",
  "resource_id": "sec_tls",
  "consumer_type": "stack_resource",
  "consumer_id": "sr_web"
}
```

Invariant:

> Do not duplicate connection-backed usage into `ResourceUsage`. If a relationship is represented by `stack_connections`, query `stack_connections` directly for delete protection, impact analysis, and topology metadata.

## Validation

Connection validation:

- `from` node exists.
- `to` node exists.
- `kind` is allowed for the source and target node types.
- Each referenced output exists on the `from` node.
- Sensitive outputs can only flow to supported targets.
- Env var names are valid and do not collide with literal env vars unless override behavior is explicit.
- File mount paths are absolute and do not collide.
- Volume mount paths are absolute and do not collide.
- `depends_on` fields form a DAG.
- Public output references require the referenced port to be public.
- Port output references require the named port to exist.

Validation must run against the fully materialized desired stack graph after
applying the requested mutation onto persisted stack state. The request body
alone is not the source of truth for validation once StackResources and
Connections become independently mutable subresources.

Examples:

- Stack create: desired graph = request.
- Stack update: desired graph = persisted Stack merged with the request.
- StackResource create/update/delete: desired graph = persisted Stack merged
  with the StackResource mutation.
- Connection create/update/delete: desired graph = persisted Stack merged with
  the Connection mutation.

Invariant:

> A Stack must exist first, but after that StackResources and Connections should
> be treated as independently mutable children of the Stack aggregate.
> Validation should resolve references against the post-merge desired aggregate,
> not just the incoming payload.

Port validation:

- Named port is required.
- Named port is unique within the StackResource.
- Port number is valid.
- Port number is unique within the StackResource.
- Public FQDN is unique within the Organisation domain.

Template validation:

- Template syntax is valid.
- All variables used in `template` are declared in `values`.
- All `values` entries resolve to existing outputs.
- Sensitivity propagates from input values to the composed result.

## Current API Capability Coverage

The redesigned API must be able to express every capability currently present in the Stack API, even when the public shape changes.

| Current API capability | Covered by new approach? | Notes |
|---|---:|---|
| Multiple StackResources in one Stack | Yes | `StackSpec.stack_resources[]` remains. |
| Stack-level Volumes | Yes | `StackSpec.volumes[]` remains. |
| StackResource labels/annotations | Yes | Intrinsic StackResource config remains. |
| Image-based resource | Yes | `image_spec` remains. |
| Image pull secret | Yes | Keep on `image_spec` initially; derive usage/topology metadata. |
| Build-from-git resource | Yes | `build_spec.source_context.git_repo` remains. |
| Build-from-volume resource | Yes | `build_spec.source_context.volume` remains; derive topology metadata. |
| Git credential secret | Yes | Keep on `build_spec` initially; derive usage/topology metadata. |
| Registry push secret | Yes | Keep on `build_spec` initially; derive usage/topology metadata. |
| Init command/args/image | Yes | `init_spec` remains intrinsic StackResource config. |
| Runtime command/args | Yes | `execution_config.command` and `execution_config.args` remain. |
| Literal env vars | Yes | Env values may be literal strings. |
| Env var from same resource's generated public URL | Yes | `self_output`, for example `public.http.url`. |
| Env var from another StackResource internal address | Yes | `kind=env` connection using `host`, `port.<name>`, or `url.<name>`. |
| Env var from another StackResource public URL | Yes | `kind=env` connection using `public.<name>.url`. |
| Env var from Secret key | Yes | `kind=env` connection from Secret using `key.<key_name>`. |
| Secret mounted as file | Deferred | Not part of the current connection API. |
| Postgres addon env injection | Yes | `kind=env` connection from PostgresAddon. |
| Postgres addon database selection | Yes | Postgres env connection `config.database`. |
| Postgres addon superuser credential selection | Yes | Postgres env connection `config.superuser`. |
| Volume mount | Yes | `kind=volume_mount` connection. |
| Volume mount source sub-path | Yes | `volume_mount.config.sub_path`. |
| Volume mount read-only flag | Yes | `volume_mount.config.read_only`. |
| Stateful workload | Yes | `stateful` remains, or maps to future `workload_type`. |
| Lifecycle restart request | Yes | `lifecycle_config` remains. |
| Multiple ports on one resource | Yes | Named `ports[]`. |
| Public exposure per port | Yes | `ports[].public`. |
| Distinct FQDN per public port | Yes | Each named public port resolves its own `public.<name>.*` outputs. |
| Startup dependency | Yes | Keep on StackResource `depends_on`; expose as derived topology edge. |
| Volume from empty PVC | Yes | Volume source remains absent/empty. |
| Volume from remote dir | Yes | Volume source remains direct config. |
| Volume from git repo | Yes | Volume source remains direct config. |
| Volume from build artifact | Yes | `kind=build_artifact_source` connection. |
| ObjectStore used by Postgres backups | Yes | Keep in PostgresAddon backup config; internally track as `ResourceUsage` and expose as topology metadata. |

Known intentional changes:

- `{{ STACKDOME_* }}` interpolation is not preserved as public API. Use `self_output`, `kind=env` connections, and structured templates instead.
- Unnamed ports are replaced by named ports. Existing dev/test specs should be updated or one-time migrated.
- `env_from_addons`, `environment_variables_from_secret`, and `volume_mounts` should no longer be the preferred public shape once their connection equivalents are implemented.

## Resolution Flow

On Stack create/update:

1. Validate StackResources, ports, outputs, and connections.
2. Resolve public FQDNs for public ports.
3. Resolve connection mappings.
4. Inject resolved env vars or secret refs into effective StackResource runtime config.
5. Persist canonical `connections`.
6. Recompute `ResourceUsage`.
7. Build/update StackResource CRs.

On connection CRUD:

1. Load Stack.
2. Apply the connection change to `StackSpec.connections`.
3. Re-run connection validation.
4. Re-resolve affected target resources.
5. Persist Stack and affected StackResources transactionally.
6. Recompute affected `ResourceUsage`.
7. Trigger reconciliation for affected resources.

On output-affecting changes:

- Port added/removed/renamed.
- Public exposure changed.
- StackResource renamed.
- Addon output changed.
- Secret key changed.

The API server should find connections where the changed node is `from`, revalidate mappings, and update affected consumers.

Future subresource flows should use the same model:

1. Load the current Stack aggregate from the database.
2. Apply the requested child mutation in memory.
3. Build the fully materialized desired Stack graph.
4. Validate that desired graph.
5. Persist only the changed child resources transactionally.
6. Recompute derived `ResourceUsage`.
7. Trigger reconciliation for affected resources.

This avoids baking full-stack-replacement assumptions into validation logic and
keeps the design compatible with future independent CRUD for StackResources and
Connections.

## Breaking Changes and Migration

Stackdome has not released this product surface to customers yet, so the redesign should favor a clean API over long-term compatibility with early internal shapes.

Breaking changes:

- Remove public support for `{{ STACKDOME_* }}` environment interpolation as a wiring mechanism.
- Replace `env_from_addons` with `kind=env` connections from addons to StackResources.
- Replace `env_vars_from_secrets` with `kind=env` connections from Secrets to StackResources.
- Replace StackResource `volume_mounts` with `kind=volume_mount` connections from Volumes to StackResources.
- Keep StackResource `depends_on` as the public startup-ordering field and expose it as a derived topology edge.

Migration approach:

- Internal sample stacks, seed data, and tests should be updated to use explicit connections.
- Any existing dev/test data can be recreated or migrated with a one-off script.
- Generated clients should expose the new connection model as the only preferred API shape.
- The cluster agent can still receive plain env vars and volume mounts after API-server resolution; the breaking change is in the hub API contract, not necessarily in the spoke CRD.

## OpenAPI Schema Sketch

```yaml
StackConnection:
  type: object
  required: [kind, from, to]
  properties:
    id:
      type: string
    kind:
      type: string
      enum: [env, volume_mount, build_artifact_source]
    from:
      $ref: "#/components/schemas/TopologyNodeRef"
    to:
      $ref: "#/components/schemas/TopologyNodeRef"
    mappings:
      type: array
      items:
        $ref: "#/components/schemas/ConnectionMapping"
    config:
      type: object
      additionalProperties: true

TopologyNodeRef:
  type: object
  required: [type]
  properties:
    type:
      type: string
      enum: [stack_resource, addon/postgres, secret, volume, object_store]
    id:
      type: string
    name:
      type: string

ConnectionMapping:
  type: object
  required: [target, value]
  properties:
    target:
      $ref: "#/components/schemas/ConnectionTarget"
    value:
      $ref: "#/components/schemas/ValueRef"

ConnectionTarget:
  type: object
  required: [type]
  properties:
    type:
      type: string
      enum: [env, file]
    name:
      type: string
    path:
      type: string

ValueRef:
  oneOf:
    - $ref: "#/components/schemas/LiteralValue"
    - $ref: "#/components/schemas/OutputValueRef"
    - $ref: "#/components/schemas/SelfOutputValueRef"
    - $ref: "#/components/schemas/TemplateValueRef"

LiteralValue:
  type: string

OutputValueRef:
  type: object
  required: [output]
  properties:
    output:
      type: string

SelfOutputValueRef:
  type: object
  required: [self_output]
  properties:
    self_output:
      type: string

TemplateValueRef:
  type: object
  required: [template, values]
  properties:
    template:
      type: string
    values:
      type: object
      additionalProperties:
        $ref: "#/components/schemas/OutputValueRef"
```

## Implementation Phases

### Phase 1: API Contract and Read Model

- Add OpenAPI schemas for named ports, connection refs, value refs, and topology.
- Add `connections` to `StackSpec`.
- Add topology and outputs endpoints.
- Update sample stacks, frontend forms, and tests to use the new API shape.

### Phase 2: Env Connections

- Implement `kind=env`.
- Resolve StackResource and PostgresAddon outputs.
- Inject resolved env vars or secret refs into effective runtime config.
- Read env connection usage directly from `stack_connections`.

### Phase 3: Canvas Support APIs

- Implement connection CRUD endpoints.
- Implement connectable nodes endpoint.
- Add validation responses shaped for UI field errors.

### Phase 4: Additional Connection Kinds

- Add `volume_mount` connection support.
- Keep secret file mounts deferred; only secret-to-env connections are supported.
- Ensure StackResource `depends_on` appears in topology as derived edges.

### Phase 5: Legacy Conversion

- Remove old public fields and validation paths once internal samples/tests are migrated.
- Keep only temporary internal conversion helpers if needed to reduce implementation risk.

### Phase 6: Aggregate-Merge Validation for Subresource CRUD

- Refactor validation and service flows so they do not assume that a request
  contains the full desired Stack state.
- Treat `Stack` as the aggregate root that must exist first.
- Treat `StackResource` and `Connection` as independently mutable children of
  the Stack.
- Load persisted Stack state before validating subresource mutations.
- Apply the requested mutation in memory and validate the post-merge desired
  graph.
- Reuse the same validation model for future `StackResource` and `Connection`
  create/update/delete endpoints.

This phase is required before the API can safely add independent CRUD for
StackResources and Connections without inconsistent reference validation.

## Design Decisions

- Connections are user-facing source of truth.
- `ResourceUsage` is derived internal state.
- Ports have names.
- Accessors use named ports: `port.http`, `url.http`, `public.http.url`.
- Env is one connection kind, not the whole relationship model.
- Value references are structured first; templates are a controlled escape hatch.
- Topology includes canonical edges, derived dependency edges, and future observed-traffic edges.
- The first phase is Stack-scoped. Cross-stack and cross-cluster references are future work.
- Validation should be designed around a post-merge desired aggregate, not a
  full-stack replacement assumption.

## Review Decisions

- `ports[].name` is mandatory for new resources.
- Env var collisions reject by default.
- `depends_on` remains on StackResource and appears in topology as a derived edge.
- User-authored network connections are out of scope for this phase. Revisit them when adding network policies.
- Secret output accessors preserve original key names. Use dot syntax for simple keys and bracket syntax with single quotes for special keys, for example `key['tls.crt']`.
