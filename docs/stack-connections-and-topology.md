# Stack Connections and Topology API

This document covers the Connections and Topology APIs for frontend engineers and API clients. It explains how to wire resources together in a Stack and how to render a topology graph.

## Overview

A Stack is a graph of resources. **Connections** are the edges — they describe how resources relate to each other (environment variables, volume mounts, build artifact sources). **Topology** is the read-only graph view that combines connections with derived relationships like `depends_on`.

Connections are user-authored and persisted. Topology is server-computed and returned on demand.

## Concepts

### Topology Nodes

Anything that can appear on a Stack canvas is a topology node. Each node has a type and an identifier:

| Type | Identifier | Description |
|------|-----------|-------------|
| `stack_resource` | `name` | A service, worker, or job within the stack |
| `addon/postgres` | `id` | A managed PostgreSQL addon |
| `secret` | `id` | A secret (env vars, TLS certs, credentials) |
| `volume` | `name` | A persistent volume |
| `object_store` | `id` | A backup storage target |

Stack-local resources (StackResources, Volumes) are referenced by `name`. External resources (PostgresAddons, Secrets) are referenced by `id`.

### Connection Kinds

| Kind | From | To | Description |
|------|------|----|-------------|
| `env` | Any producer | `stack_resource` | Injects values into environment variables |
| `volume_mount` | `volume` | `stack_resource` | Mounts a volume into a container |
| `build_artifact_source` | `stack_resource` | `volume` | Seeds a volume from build output |

### Outputs

Every topology node declares **outputs** — named values that other resources can consume via connections. Outputs are read-only metadata returned by the API.

**StackResource outputs:**
- `host` — internal service hostname (e.g., `api.my-stack-ns.svc.cluster.local`)
- `port.<port_name>` — port number (e.g., `port.http` = `8080`)
- `url.<port_name>` — internal URL (e.g., `url.http` = `http://api.my-stack-ns.svc:8080`)
- `public.<port_name>.host` — public FQDN (if port is exposed)
- `public.<port_name>.url` — public URL (if port is exposed)

**PostgresAddon outputs:**
- `host`, `port`, `database`, `username`, `password`, `sslmode`, `ca_certificate`, `url`

**Secret outputs:**
- `key.<key_name>` — for simple key names (e.g., `key.api_key`)
- `key['<key_name>']` — for keys with dots or special characters (e.g., `key['tls.crt']`)

## API Endpoints

Base path: `/api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/topology` | Get the full topology graph |
| `GET` | `/connections` | List all connections |
| `POST` | `/connections` | Create a connection |
| `PUT` | `/connections/{connection_id}` | Update a connection |
| `DELETE` | `/connections/{connection_id}` | Delete a connection |

Connections can also be included inline in the `StackSpec.connections` array when creating or updating a stack.

## Creating Connections

### Environment Variable Injection

Connect a PostgresAddon to a StackResource to inject database credentials:

```json
POST /connections
{
  "kind": "env",
  "from": { "type": "addon/postgres", "id": "pg-abc123" },
  "to": { "type": "stack_resource", "name": "api" },
  "config": {
    "database": "myapp"
  },
  "mappings": [
    {
      "target": { "type": "env", "name": "DATABASE_URL" },
      "value": { "output": "url" }
    },
    {
      "target": { "type": "env", "name": "DB_HOST" },
      "value": { "output": "host" }
    }
  ]
}
```

Connect a Secret to a StackResource:

```json
{
  "kind": "env",
  "from": { "type": "secret", "id": "sec-xyz789" },
  "to": { "type": "stack_resource", "name": "api" },
  "mappings": [
    {
      "target": { "type": "env", "name": "API_KEY" },
      "value": { "output": "key.api_key" }
    },
    {
      "target": { "type": "env", "name": "TLS_CERT" },
      "value": { "output": "key['tls.crt']" }
    }
  ]
}
```

Connect one StackResource to another:

```json
{
  "kind": "env",
  "from": { "type": "stack_resource", "name": "api" },
  "to": { "type": "stack_resource", "name": "worker" },
  "mappings": [
    {
      "target": { "type": "env", "name": "API_HOST" },
      "value": { "output": "host" }
    },
    {
      "target": { "type": "env", "name": "API_URL" },
      "value": { "output": "url.http" }
    }
  ]
}
```

### Template Values

When a single env var needs values from multiple outputs, use a template:

```json
{
  "kind": "env",
  "from": { "type": "addon/postgres", "id": "pg-abc123" },
  "to": { "type": "stack_resource", "name": "api" },
  "config": { "database": "myapp" },
  "mappings": [
    {
      "target": { "type": "env", "name": "DATABASE_URL" },
      "value": {
        "template": "postgres://{{ user }}:{{ pass }}@{{ host }}:{{ port }}/{{ db }}",
        "values": {
          "user": { "output": "username" },
          "pass": { "output": "password" },
          "host": { "output": "host" },
          "port": { "output": "port" },
          "db":   { "output": "database" }
        }
      }
    }
  ]
}
```

A mapping's `value` must have either `output` (direct reference) or `template` + `values` (composed reference), never both.

### PostgresAddon Connection Config

Env connections from a PostgresAddon require a `config.database` field specifying which database to get credentials for. Optionally set `config.credential_scope` to `superuser` to get superuser credentials (requires `enableSuperuserAccess: true` on the addon).

```json
{
  "config": {
    "database": "myapp",
    "credential_scope": "superuser"
  }
}
```

### Volume Mount

Mount a volume into a StackResource:

```json
{
  "kind": "volume_mount",
  "from": { "type": "volume", "name": "uploads" },
  "to": { "type": "stack_resource", "name": "api" },
  "config": {
    "mount_path": "/app/uploads",
    "sub_path": "media",
    "read_only": false
  }
}
```

Config fields:
- `mount_path` (required) — container path to mount at
- `sub_path` (optional) — subdirectory within the volume
- `read_only` (optional, default `false`) — mount as read-only

### Build Artifact Source

Seed a volume from a StackResource's build output:

```json
{
  "kind": "build_artifact_source",
  "from": { "type": "stack_resource", "name": "builder" },
  "to": { "type": "volume", "name": "static-assets" },
  "config": {
    "source_path": "/app/dist"
  }
}
```

Config fields:
- `source_path` (required) — path inside the build container to copy from

## Self-Output Environment Variables

A StackResource can reference its own outputs in env vars using `self_output` instead of `value`:

```json
{
  "name": "api",
  "execution_config": {
    "env": [
      { "name": "SELF_URL", "self_output": "url.http" },
      { "name": "APP_PORT", "self_output": "port.http" }
    ]
  }
}
```

This is useful for injecting the resource's own hostname or URL into its configuration without creating a connection.

## Reading Topology

`GET /topology` returns a graph with nodes and edges:

```json
{
  "nodes": [
    {
      "ref": { "type": "stack_resource", "id": "res-1", "name": "api" },
      "label": "api",
      "outputs": [
        { "name": "host", "type": "string", "sensitive": false },
        { "name": "port.http", "type": "integer", "sensitive": false },
        { "name": "url.http", "type": "string", "sensitive": false }
      ],
      "state": "running"
    },
    {
      "ref": { "type": "addon/postgres", "id": "pg-abc123", "name": "main-db" },
      "label": "main-db",
      "outputs": [
        { "name": "host", "type": "string", "sensitive": false },
        { "name": "url", "type": "string", "sensitive": true }
      ],
      "state": "ready"
    }
  ],
  "edges": [
    {
      "id": "conn-1",
      "kind": "env",
      "source": { "type": "addon/postgres", "id": "pg-abc123", "name": "main-db" },
      "target": { "type": "stack_resource", "id": "res-1", "name": "api" },
      "mappings": [ ... ],
      "config": { "database": "myapp" },
      "source_of_truth": "connection"
    },
    {
      "kind": "depends_on",
      "source": { "type": "stack_resource", "id": "res-2", "name": "db-migrator" },
      "target": { "type": "stack_resource", "id": "res-1", "name": "api" },
      "source_of_truth": "derived"
    }
  ]
}
```

### Edge Source of Truth

Each edge has a `source_of_truth` field:
- `connection` — user-authored via the connections API
- `derived` — inferred by the system (e.g., `depends_on` relationships from StackResource config)

Connection edges have an `id` that matches the `StackConnection.id`. Derived edges have no `id`.

### Node State

Nodes with runtime status include a `state` field. Possible values depend on the node type. If a referenced resource has been deleted, the node will have `state: "missing"`.

### Naming: `from`/`to` vs `source`/`target`

`StackConnection` uses `from`/`to` (user-facing, reads naturally: "from postgres to api"). `TopologyEdge` uses `source`/`target` (graph terminology). They mean the same thing: `source` = `from` (the producer), `target` = `to` (the consumer).

## Listing Connections

`GET /connections` returns a wrapped list:

```json
{
  "items": [
    {
      "id": "conn-1",
      "kind": "env",
      "from": { "type": "addon/postgres", "id": "pg-abc123" },
      "to": { "type": "stack_resource", "name": "api" },
      "mappings": [ ... ],
      "config": { "database": "myapp" }
    }
  ],
  "total": 1
}
```

## Inline Connections in Stack Spec

Connections can be included in the `StackSpec.connections` array when creating or updating a full stack:

```json
POST /stacks
{
  "name": "my-app",
  "stack_resources": [
    { "name": "api", "image_spec": { "image": "myapp:latest" }, "ports": [{ "name": "http", "number": 8080, "protocol": "http" }] },
    { "name": "worker", "image_spec": { "image": "myapp-worker:latest" } }
  ],
  "connections": [
    {
      "kind": "env",
      "from": { "type": "stack_resource", "name": "api" },
      "to": { "type": "stack_resource", "name": "worker" },
      "mappings": [
        { "target": { "type": "env", "name": "API_URL" }, "value": { "output": "url.http" } }
      ]
    }
  ]
}
```

## Delete Protection

Resources referenced by connections cannot be deleted. Attempting to delete a Secret, PostgresAddon, or Volume that is used by a connection will return a `400 Bad Request` error. Remove the connection first, then delete the resource.

## Validation Rules

- Each connection must have exactly one of each `(stack_id, kind, from, to)` combination (duplicates are rejected)
- `env` connections require at least one mapping with `target.type = "env"` and a non-empty `target.name`
- Each mapping value must have either `output` or `template` + `values`, not both, not neither
- Templates must have a non-empty `values` map
- Referenced outputs must exist in the source node's declared outputs
- `volume_mount` connections require `config.mount_path`
- `build_artifact_source` connections require `config.source_path` and must target a `volume`
- PostgresAddon connections require `config.database`
- Port names are required and must be unique within a StackResource
