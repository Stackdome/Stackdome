# Stackdome Stackfile Reference

A stackfile is a YAML document that describes a complete stack: its workloads, how they are built or pulled, how they wire to each other, and how they consume secrets, addons, and volumes. It is the declarative authoring surface for both humans and agents.

A stackfile **describes and connects**. It never creates secrets or addons — those must already exist in the project, and the stackfile references them by name.

The machine-readable schema lives at `pkg/stackfile/schema.json` (JSON Schema draft-07). Unknown fields are rejected at parse time — a typo'd key is an error, never silently ignored.

## Top-level structure

```yaml
name: my-stack          # required — stack name, unique within the project

resources:              # required — at least 1, at most 50 workloads
  api: { ... }
  worker: { ... }

volumes:                # optional — persistent volumes, at most 50
  pg-data:
    size: 5Gi
```

## Resources

Each entry under `resources:` is one workload. Exactly one of `image` or `build` is required.

```yaml
resources:
  api:
    image: myorg/api:v1.2.0        # OR build: (never both)
    command: ["/app/server"]        # optional container entrypoint override
    args: ["--port", "3000"]        # optional args
    workload_type: Service          # Service | StatefulService | Worker | Job | CronJob
    replicas: 2                     # optional, >= 0
    depends_on:                     # start ordering; must name resources in this file
      - redis
    ports: [ ... ]
    env: { ... }
    secrets: { ... }
    addons: { ... }
    volumes: [ ... ]
```

### Workload types

| Type | Meaning |
|---|---|
| `Service` | Long-running, receives traffic. Default. |
| `StatefulService` | Long-running with stable identity (databases, queues). |
| `Worker` | Long-running, no ports. |
| `Job` | Run-to-completion (migrations, one-off tasks). |
| `CronJob` | Scheduled run-to-completion. Requires `schedule` (cron expression). |

```yaml
  nightly-report:
    image: myorg/reporter:latest
    workload_type: CronJob
    schedule: "0 3 * * *"
```

### Building from source

```yaml
  api:
    build:
      repo: https://github.com/myorg/api.git   # required
      branch: main                             # branch OR tag, not both
      commit: 4f2a91c                          # optional SHA pin; requires branch or tag
      dockerfile: Dockerfile.prod              # default: Dockerfile
      context: ./services/api                  # default: .
```

Rules:

- `branch` and `tag` are mutually exclusive.
- `commit` pins a SHA and requires `branch` or `tag` alongside it.
- Neither branch nor tag: the server resolves the repository's default branch.
- Clone credentials come from the preview config or an org-level git integration — never from the stackfile.

### Ports

```yaml
    ports:
      - name: http          # required
        port: 3000          # required, 1–65535
        protocol: TCP       # optional
        public: true        # expose via ingress
        subdomain: api      # subdomain prefix for the public URL
```

At most 20 ports per resource. Ports do double duty: they configure networking **and** they define the resource's outputs (next section).

## Outputs and how to consume them

Three kinds of things produce outputs: resources, addons, and secrets. Each has its own vocabulary and consumption syntax.

### Resource outputs

A resource's outputs are derived entirely from its `ports`:

| Output | Available |
|---|---|
| `host` | always — in-cluster service DNS name |
| `port` | per port — the port number |
| `url` | per port — in-cluster `host:port` URL |
| `public_host` | per port, only if `public: true` |
| `public_url` | per port, only if `public: true` |

**Naming rule:** a resource with **one** port uses bare names (`port`, `url`, `public_url`). A resource with **multiple** ports suffixes with the port name (`port.http`, `url.grpc`, `public_url.web`). `host` is always bare. Referencing bare `url` on a multi-port resource is rejected as ambiguous.

#### Your own outputs — `{{ self.X }}`

The ref must be the **entire value**. No templates, no mixing with other refs.

```yaml
  api:
    ports:
      - name: http
        port: 3000
        public: true
        subdomain: api
    env:
      SITE_URL: "{{ self.public_url }}"   # OK
      PORT: "{{ self.port }}"             # OK
      # INVALID: "https://{{ self.public_host }}/cb" — self refs cannot be templated
```

#### Another resource's outputs — `{{ <resource>.X }}`

Exact ref or template. **All refs in one value must come from the same source resource.**

```yaml
  api:
    env:
      REDIS_HOST: "{{ redis.host }}"                          # exact
      REDIS_URL: "redis://{{ redis.host }}:{{ redis.port }}"  # template, one source
      # INVALID: "{{ redis.host }}:{{ db.port }}" — two sources in one value

  redis:
    image: redis:7
    ports:
      - name: redis
        port: 6379
```

Multi-port source:

```yaml
  api:
    env:
      GRPC_ADDR: "{{ backend.url.grpc }}"
      HTTP_ADDR: "{{ backend.url.http }}"

  backend:
    image: myorg/backend:latest
    ports:
      - name: http
        port: 8080
      - name: grpc
        port: 9090
```

Everything is validated at parse time: an unknown output name errors with the list of valid outputs for that resource.

### Addon outputs

Declared under a resource's `addons:` block. The addon must already exist in the project; the block key is its name. Currently supported type: `postgres`.

Postgres outputs: `host`, `port`, `database`, `username`, `password`, `sslmode`, `ca_certificate`, `url`.

Inside the addon's `env:`, outputs are referenced **bare** — no prefix, since the block already names the addon. Templates are fine and may mix any outputs.

```yaml
  api:
    addons:
      main-db:
        type: postgres
        database: api_production     # scope credentials to this database
        env:
          DATABASE_URL: "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}"
          PGSSLMODE: "{{ sslmode }}"
```

`{{ postgres.host }}` (dotted) is invalid — bare names only, rejected at parse time.

#### Superuser credentials

```yaml
  migration-runner:
    image: myorg/api:latest
    workload_type: Job
    addons:
      main-db:
        type: postgres
        database: api_production
        superuser: true
        env:
          DATABASE_URL: "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}"
```

With `superuser: true`, `{{ username }}` / `{{ password }}` resolve to superuser credentials instead of the app owner role. The addon itself must have been provisioned with `enable_superuser_access` — the stackfile only requests the connection, it cannot enable it.

Typical pattern: the app connects plain (owner credentials); a dedicated `Job` resource connects with `superuser: true` for DDL, extensions, or migrations.

### Secret outputs

Each key of a project secret is an output. Consumption is a plain map — no `{{ }}`, no templating:

```yaml
  api:
    secrets:
      jwt-secret:                    # secret name in the project
        JWT_SECRET: jwt_signing_key  # ENV_VAR: secret_key
      smtp-creds:
        SMTP_USER: username
        SMTP_PASS: password
```

### Cheat sheet

| Source | Syntax | Where | Templates |
|---|---|---|---|
| self | `{{ self.port }}` | `env:` | no — whole value only |
| other resource | `{{ redis.host }}` | `env:` | yes — one source per value |
| addon | `{{ host }}` (bare) | addon's `env:` | yes — free mix |
| secret | `ENV_VAR: key` map | `secrets:` | no |

## Volumes

Declare volumes at the top level, mount them per resource:

```yaml
resources:
  db:
    image: postgres:16
    workload_type: StatefulService
    volumes:
      - name: pg-data                       # must be declared below
        path: /var/lib/postgresql/data      # absolute mount path

volumes:
  pg-data:
    size: 5Gi                # required — e.g. 512Mi, 5Gi, 1Ti
    access_mode: ReadWriteOnce   # default; also ReadOnlyMany, ReadWriteMany
```

A mount referencing an undeclared volume is a parse error.

## Complete example

```yaml
name: my-app

resources:
  api:
    build:
      repo: https://github.com/myorg/api.git
      branch: main
    ports:
      - name: http
        port: 3000
        public: true
        subdomain: api
    env:
      SITE_URL: "{{ self.public_url }}"
      CACHE_URL: "redis://{{ redis.host }}:{{ redis.port }}"
    secrets:
      jwt-secret:
        JWT_SECRET: jwt_signing_key
    addons:
      main-db:
        type: postgres
        database: app_production
        env:
          DATABASE_URL: "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}"
    depends_on:
      - redis

  migrate:
    image: myorg/api:latest
    workload_type: Job
    command: ["/app/migrate"]
    addons:
      main-db:
        type: postgres
        database: app_production
        superuser: true
        env:
          DATABASE_URL: "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}"

  redis:
    image: redis:7-alpine
    workload_type: StatefulService
    ports:
      - name: redis
        port: 6379
    volumes:
      - name: redis-data
        path: /data

volumes:
  redis-data:
    size: 1Gi
```

## Validation summary

Checked at parse time (before anything reaches the server):

- `name` and at least one resource required; unknown fields anywhere are errors.
- Exactly one of `image` / `build` per resource.
- `branch` xor `tag`; `commit` requires one of them.
- Port numbers 1–65535; every port named; max 20 ports, 50 resources, 50 volumes; file size max 1 MiB.
- Volume mounts must reference declared volumes; `depends_on` must reference resources in the file; no self-dependency.
- Every `{{ ref }}` checked against the source's real outputs; self refs whole-value only; one source resource per env value; addon refs bare-name only.
- Secrets and addons are validated for existence when the stack is resolved server-side.

## Example stackfiles

These live in `pkg/stackfile/testdata/` and double as test fixtures — every one is validated against the schema and round-tripped through the transpiler on each test run.

### Shared database, owner + superuser (`superuser_migration.yaml`)

One postgres addon, two consumers: the app on owner credentials, a migration `Job` on superuser credentials.

```yaml
name: superuser-migration

resources:
  app:
    image: myorg/api:latest
    ports:
      - name: http
        port: 3000
        public: true
        subdomain: api
    addons:
      main-db:
        type: postgres
        database: app_production
        env:
          DATABASE_URL: "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}"

  migrate:
    image: myorg/api:latest
    workload_type: Job
    command: ["/app/migrate"]
    addons:
      main-db:
        type: postgres
        database: app_production
        superuser: true
        env:
          DATABASE_URL: "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}"
    depends_on:
      - app
```

### Everything at once (`kitchen_sink.yaml`)

Secrets, addon templating, cross-resource refs, self refs, a stateful cache with a volume:

```yaml
name: kitchen-sink

resources:
  api:
    image: api:latest
    ports:
      - name: http
        port: 3000
        public: true
        subdomain: api
    env:
      SITE_URL: "{{ self.public_url }}"
      CACHE_HOST: "{{ redis.host }}"
      CACHE_URL: "redis://{{ redis.host }}:6379"
      LOG_LEVEL: info
    secrets:
      jwt-secret:
        JWT_SECRET: jwt_signing_key
      smtp-creds:
        SMTP_USER: username
        SMTP_PASS: password
    addons:
      main-db:
        type: postgres
        database: api_production
        env:
          DATABASE_URL: "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}"
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - name: redis
        port: 6379
        protocol: TCP
    volumes:
      - name: redis-data
        path: /data
    workload_type: StatefulService

  worker:
    image: api:latest
    env:
      ROLE: worker
      CACHE_URL: "redis://{{ redis.host }}:6379"
    secrets:
      jwt-secret:
        JWT_SECRET: jwt_signing_key
    addons:
      main-db:
        type: postgres
        database: api_production
        env:
          DATABASE_URL: "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}"
    depends_on:
      - redis

volumes:
  redis-data:
    size: 2Gi
```

Other fixtures: `basic_image.yaml` (minimal), `build_from_source.yaml` (git build), `with_secrets.yaml`, `with_addon.yaml`, `with_addon_superuser.yaml`, `infisical.yaml` (real-world multi-resource app), `simple_nginx.yaml`.

## JSON Schema

Source of truth: `pkg/stackfile/schema.json`, embedded in the server binary. Point your editor at it (`# yaml-language-server: $schema=...`) for autocomplete and inline validation.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://stackdome.com/schemas/stackfile.schema.json",
  "title": "Stackdome Stackfile",
  "type": "object",
  "required": ["name", "resources"],
  "additionalProperties": false,
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "description": "Stack name, unique within the project."
    },
    "resources": {
      "type": "object",
      "minProperties": 1,
      "maxProperties": 50,
      "description": "Workloads keyed by resource name.",
      "additionalProperties": { "$ref": "#/definitions/resource" }
    },
    "volumes": {
      "type": "object",
      "maxProperties": 50,
      "description": "Persistent volumes keyed by volume name.",
      "additionalProperties": { "$ref": "#/definitions/volume" }
    }
  },
  "definitions": {
    "resource": {
      "type": "object",
      "additionalProperties": false,
      "oneOf": [
        { "required": ["image"], "not": { "required": ["build"] } },
        { "required": ["build"], "not": { "required": ["image"] } }
      ],
      "properties": {
        "image": {
          "type": "string",
          "minLength": 1,
          "description": "Container image reference. Mutually exclusive with build."
        },
        "build": { "$ref": "#/definitions/build" },
        "command": { "type": "array", "items": { "type": "string" } },
        "args": { "type": "array", "items": { "type": "string" } },
        "ports": {
          "type": "array",
          "maxItems": 20,
          "items": { "$ref": "#/definitions/port" }
        },
        "env": {
          "type": "object",
          "description": "Env vars. Values may reference outputs: {{ self.port }}, {{ other-resource.host }}, or templates like redis://{{ redis.host }}:{{ redis.port }}. All refs in one value must use the same source resource.",
          "additionalProperties": { "type": "string" }
        },
        "secrets": {
          "type": "object",
          "description": "Secret name -> { ENV_VAR: secret_key }. Secrets must already exist in the project.",
          "additionalProperties": {
            "type": "object",
            "additionalProperties": { "type": "string" }
          }
        },
        "addons": {
          "type": "object",
          "description": "Addon name -> connection config. Addons must already exist in the project.",
          "additionalProperties": { "$ref": "#/definitions/addon" }
        },
        "volumes": {
          "type": "array",
          "items": { "$ref": "#/definitions/volumeMount" }
        },
        "depends_on": { "type": "array", "items": { "type": "string" } },
        "workload_type": {
          "type": "string",
          "enum": ["Service", "StatefulService", "Worker", "Job", "CronJob"],
          "description": "Defaults to Service."
        },
        "schedule": {
          "type": "string",
          "description": "Cron expression. Only for workload_type CronJob."
        },
        "replicas": { "type": "integer", "minimum": 0 }
      }
    },
    "build": {
      "type": "object",
      "additionalProperties": false,
      "required": ["repo"],
      "not": { "required": ["branch", "tag"] },
      "dependencies": {
        "commit": { "anyOf": [{ "required": ["branch"] }, { "required": ["tag"] }] }
      },
      "properties": {
        "repo": { "type": "string", "minLength": 1, "description": "Git repository URL." },
        "branch": { "type": "string" },
        "tag": { "type": "string" },
        "commit": { "type": "string", "description": "Commit SHA pin; requires branch or tag." },
        "dockerfile": { "type": "string", "description": "Path to the Dockerfile." },
        "context": { "type": "string", "description": "Build context directory." }
      }
    },
    "port": {
      "type": "object",
      "additionalProperties": false,
      "required": ["name", "port"],
      "properties": {
        "name": { "type": "string", "minLength": 1 },
        "port": { "type": "integer", "minimum": 1, "maximum": 65535 },
        "protocol": { "type": "string" },
        "public": { "type": "boolean", "description": "Expose publicly via ingress." },
        "subdomain": { "type": "string", "description": "Subdomain prefix for the public URL." }
      }
    },
    "volume": {
      "type": "object",
      "additionalProperties": false,
      "required": ["size"],
      "properties": {
        "size": {
          "type": "string",
          "pattern": "^[0-9]+[KMGTP]i?$",
          "description": "Volume size, e.g. 5Gi."
        },
        "access_mode": {
          "type": "string",
          "enum": ["ReadWriteOnce", "ReadOnlyMany", "ReadWriteMany"],
          "description": "Defaults to ReadWriteOnce."
        }
      }
    },
    "volumeMount": {
      "type": "object",
      "additionalProperties": false,
      "required": ["name", "path"],
      "properties": {
        "name": { "type": "string", "description": "Name of a volume declared in top-level volumes." },
        "path": { "type": "string", "description": "Absolute mount path in the container." }
      }
    },
    "addon": {
      "type": "object",
      "additionalProperties": false,
      "required": ["type"],
      "properties": {
        "type": { "type": "string", "enum": ["postgres"] },
        "database": { "type": "string", "description": "Database to connect to (postgres)." },
        "superuser": { "type": "boolean", "description": "Use superuser credentials (postgres)." },
        "env": {
          "type": "object",
          "description": "Env vars templated from addon outputs, e.g. postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}. Bare output names, no prefix.",
          "additionalProperties": { "type": "string" }
        }
      }
    }
  }
}
```

## Not expressible in a stackfile

These exist in the Stack API but have no stackfile syntax (the exporter fails loudly on them rather than dropping them):

- Init containers, restart triggers, labels/annotations.
- Private image pull credentials, git integration selection, push targets.
- Building an image from a volume; volume sources (git repo / remote dir / build artifact), storage class, sub-path or read-only mounts.
- Mounting secrets or outputs as files (everything is env).
- Addon provisioning (version, storage, backups, instances) — stackfiles connect to addons, they never create them.
