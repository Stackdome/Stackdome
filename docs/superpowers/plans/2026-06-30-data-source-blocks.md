# Plan: expand the wizard's Data Store composing blocks

**Status:** v1 set shipped (MariaDB, MS SQL Server, Elasticsearch, CouchDB, InfluxDB, ClickHouse) — commit `72f56d53`. v2 specialized stores still pending.
**Owner area:** `frontend/src/pages/stacks/data/blocks/*`, `frontend/src/pages/stacks/components/wizard/*`, `frontend/src/assets/brand/*`.

## Goal

The Build-from-blocks composer today ships 4 data stores (Postgres, Redis, MySQL, MongoDB). Extend the **DATA STORES** category with the rest of the self-hostable data sources, each with:
- the right brand icon (from selfh.st, confirmed dark/light safe),
- a **stable, deployable** image + tag,
- a **sane minimal default** config (ports, named volume, required env) that round-trips through the existing block → docker-compose → form pipeline.

## Scope decision: deployable vs managed

A "data source" list mixes two kinds:
- **Deployable** (open-source DB you run in your cluster) → these become **blocks** (a container the stack runs). In scope.
- **Managed / SaaS** (Snowflake, BigQuery, DynamoDB, Firestore, Athena, Databricks, Cosmos DB, SAP HANA, Oracle Cloud, etc.) → you can't "deploy" these; they're external **connections**, not blocks. **Out of scope** for blocks (would belong to a future "external connection" data-source flow, not the composer).

> Action: confirm the in-scope list below against Image #26 — add/remove rows; each row is one registry entry.

## Candidate matrix (deployable; all icons verified present on selfh.st with base+light+dark)

| Data source | BlockId | selfh.st slug | Stable image (pin + verify latest patch) | Ports | Named volume → path | Key env defaults | Notes |
|---|---|---|---|---|---|---|---|
| PostgreSQL ✅ shipped | `postgres` | `postgresql` (reuses addon art) | `postgres:16` | 5432 | `pgdata:/var/lib/postgresql/data` | `POSTGRES_PASSWORD: ""` | — |
| MySQL ✅ shipped | `mysql` | `mysql` | `mysql:8` | 3306 | `mysql-data:/var/lib/mysql` | `MYSQL_ROOT_PASSWORD: ""` | — |
| MongoDB ✅ shipped | `mongo` | `mongodb` | `mongo:7` | 27017 | `mongo-data:/data/db` | — | — |
| Redis ✅ shipped | `redis` | `redis` (reuses addon art) | `redis:7` | 6379 | — | — | — |
| MariaDB | `mariadb` | `mariadb` | `mariadb:11.4` (LTS) | 3306 | `mariadb-data:/var/lib/mysql` | `MARIADB_ROOT_PASSWORD: ""` | drop-in MySQL alt |
| MS SQL Server | `mssql` | `microsoft-sql-server` | `mcr.microsoft.com/mssql/server:2022-latest` | 1433 | `mssql-data:/var/opt/mssql` | `ACCEPT_EULA: "Y"`, `MSSQL_SA_PASSWORD: ""` | heavy (~2GB RAM); SA pw must be strong; EULA |
| Elasticsearch | `elasticsearch` | `elasticsearch` | `docker.elastic.co/elasticsearch/elasticsearch:8.15.0` | 9200 | `es-data:/usr/share/elasticsearch/data` | `discovery.type: single-node`, `xpack.security.enabled: "false"`, `ES_JAVA_OPTS: "-Xms512m -Xmx512m"` | heavy; security off for dev |
| OpenSearch | `opensearch` | `opensearch` | `opensearchproject/opensearch:2.16.0` | 9200 | `os-data:/usr/share/opensearch/data` | `discovery.type: single-node`, `plugins.security.disabled: "true"`, `OPENSEARCH_INITIAL_ADMIN_PASSWORD: ""` | ES alternative; heavy |
| ClickHouse | `clickhouse` | `clickhouse` | `clickhouse/clickhouse-server:24.8` | 8123, 9000 | `clickhouse-data:/var/lib/clickhouse` | `CLICKHOUSE_PASSWORD: ""` | columnar analytics |
| CouchDB | `couchdb` | `couchdb` | `couchdb:3.3` | 5984 | `couchdb-data:/opt/couchdb/data` | `COUCHDB_USER: admin`, `COUCHDB_PASSWORD: ""` | document DB |
| InfluxDB | `influxdb` | `influxdb` | `influxdb:2.7` | 8086 | `influxdb-data:/var/lib/influxdb2` | — (first-run setup via UI) | time-series |
| Cassandra | `cassandra` | `apache-cassandra` | `cassandra:5.0` | 9042 | `cassandra-data:/var/lib/cassandra` | — | heavy; wide-column |
| Neo4j | `neo4j` | `neo4j` | `neo4j:5.23-community` | 7474, 7687 | `neo4j-data:/data` | `NEO4J_AUTH: neo4j/password` | graph DB |
| Qdrant | `qdrant` | `qdrant` | `qdrant/qdrant:v1.11.0` | 6333, 6334 | `qdrant-data:/qdrant/storage` | — | vector DB |
| Weaviate | `weaviate` | `weaviate` | `cr.weaviate.io/semitechnologies/weaviate:1.26.1` | 8080 | `weaviate-data:/var/lib/weaviate` | `AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: "true"`, `PERSISTENCE_DATA_PATH: /var/lib/weaviate` | vector DB |
| RabbitMQ | `rabbitmq` | `rabbitmq` | `rabbitmq:3.13-management` | 5672, 15672 | `rabbitmq-data:/var/lib/rabbitmq` | — | message broker (optional in "data sources") |
| Oracle XE | `oracle` | `oracle` | `gvenzl/oracle-free:23-slim` | 1521 | `oracle-data:/opt/oracle/oradata` | `ORACLE_PASSWORD: ""` | community image; confirm licensing before shipping |

Icons NOT on selfh.st (would need custom art or skip): Couchbase, ScyllaDB, TimescaleDB, KeyDB, Dragonfly, Milvus.

## Recommended phasing

- **v1 (relational + mainstream):** MariaDB, MS SQL Server, Elasticsearch, CouchDB, InfluxDB, ClickHouse. Covers the common "data sources" set with low surprise.
- **v2 (specialized):** Neo4j, Cassandra, Qdrant, Weaviate, OpenSearch, RabbitMQ, Oracle.

## Implementation steps (per data source it's the same shape)

1. **Icons** — download base color SVG into `frontend/src/assets/brand/<slug>.svg` from `https://raw.githubusercontent.com/selfhst/icons/main/svg/<slug>.svg` (sandbox can write to the repo FS). Verify each on the dark chip (`#161c26`) and light chip (`#e8e4d6`) in Playwright; for any whose base is too dark/low-contrast, fall back to the selfh.st `-light` variant via a `dark:`-class swap (theme is a `dark` class on `<html>`).
2. **BlockId** — add enum members in `frontend/src/pages/stacks/data/blocks/types.ts` (no magic strings; registry references `BlockId.*`).
3. **BlockGlyph** — extend the `BRAND` map in `frontend/src/pages/stacks/components/wizard/block-glyph.tsx` (`<icon-key> → imported svg url`). Lucide stays the fallback for non-brand blocks.
4. **Registry** — add `BlockPreset` rows in `frontend/src/pages/stacks/data/blocks/registry.ts` with `icon` = brand key and a single-service `compose` snippet (mirror the existing Postgres entry exactly: `services:` → image/ports/volumes/environment, plus top-level `volumes:` for named volumes).
5. **Category model** — decide: keep one `DATA STORES` group (simple, the picker already has search) or split into sub-categories (Relational / Document & Search / Analytics & Time-series / Vector / Messaging). Recommendation: keep one group for v1; revisit if the list exceeds ~8.
6. **Pipeline validation** — each `compose` must round-trip through `parseAndValidateDockerCompose` → `convertDockerComposeToStackData` (`block-to-form.ts`). Add/extend a unit test that asserts every `blockCatalog` entry converts without error and yields the expected resource name/ports/volumes (guards multi-port services like ClickHouse/Neo4j/RabbitMQ).
7. **Verify** — Playwright: open composer, confirm each new block adds to "Your stack so far" with the correct brand glyph and lands in the prefilled form; check both themes.

## Risks / decisions to confirm

- **Resource-heavy images** (MS SQL, Elasticsearch, OpenSearch, Cassandra) — consider a small "heavy" hint on the card so users know these need more cluster headroom.
- **Empty-password defaults** — consistent with current blocks (`POSTGRES_PASSWORD: ""`), the user must set a secret before deploy. Confirm this is the intended UX vs. generating a placeholder/secret.
- **MS SQL EULA / Oracle licensing** — both need explicit acceptance / license review before shipping as defaults.
- **Managed data sources** (Snowflake/BigQuery/DynamoDB/etc.) are intentionally excluded from blocks; if the screenshot expects them, they belong to a separate "external connection" data-source feature, not the composer.
- **Image tags** above are pinned to known-stable majors; verify the latest patch tag at implementation time.

## Open question for you

Please confirm the exact data-source list from Image #26 so I can trim the matrix to just those rows, then I'll implement the chosen phase (icons + registry + tests + Playwright verification) in one pass.
