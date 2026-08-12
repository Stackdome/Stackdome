import { BlockId, type BlockCategoryMeta, type BlockPreset } from "./types";

// Placeholder root/admin password baked into the datastore preset compose
// snippets. MS SQL Server rejects weak passwords, so it keeps its own stronger
// placeholder below. block-to-form swaps every placeholder for a generated
// secret at add time so the container boots on first deploy.
const PLACEHOLDER_PASSWORD = "changeme";
const MSSQL_PLACEHOLDER_PASSWORD = "Change_me_123";

/** Every placeholder password used by the preset compose snippets. */
export const PLACEHOLDER_PASSWORDS: ReadonlySet<string> = new Set([
  PLACEHOLDER_PASSWORD,
  MSSQL_PLACEHOLDER_PASSWORD,
]);

export const BLOCK_CATEGORY_META: BlockCategoryMeta[] = [
  { id: "services", label: "SERVICES" },
  { id: "databases", label: "DATABASES" },
  { id: "analytics", label: "CACHE, SEARCH & ANALYTICS" },
];

export const blockCatalog: BlockPreset[] = [
  { id: BlockId.Web, name: "Web service", category: "services", icon: "globe", summary: "your image · :80 · public" },
  { id: BlockId.Custom, name: "Custom", category: "services", icon: "box", summary: "empty container shape" },
  {
    id: BlockId.Postgres, name: "Postgres", category: "databases", icon: "postgres", summary: "postgres:16 · :5432",
    compose: [
      "services:",
      "  postgres:",
      "    image: postgres:16",
      '    ports: ["5432:5432"]',
      '    volumes: ["pgdata:/var/lib/postgresql/data"]',
      "    environment:",
      `      POSTGRES_PASSWORD: "${PLACEHOLDER_PASSWORD}"`,
      '      PGDATA: "/var/lib/postgresql/data/pgdata"',
      "volumes:",
      "  pgdata: {}",
      "",
    ].join("\n"),
  },
  {
    id: BlockId.Redis, name: "Redis", category: "analytics", icon: "redis", summary: "redis:7 · :6379",
    compose: [
      "services:", "  redis:", "    image: redis:7", '    ports: ["6379:6379"]',
      '    volumes: ["redis-data:/data"]', "volumes:", "  redis-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Mysql, name: "MySQL", category: "databases", icon: "mysql", summary: "mysql:8 · :3306",
    args: "--datadir=/var/lib/mysql/data",
    compose: [
      "services:", "  mysql:", "    image: mysql:8", '    ports: ["3306:3306"]',
      '    volumes: ["mysql-data:/var/lib/mysql"]', "    environment:", `      MYSQL_ROOT_PASSWORD: "${PLACEHOLDER_PASSWORD}"`,
      "volumes:", "  mysql-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Mongo, name: "MongoDB", category: "databases", icon: "mongo", summary: "mongo:7 · :27017",
    compose: [
      "services:", "  mongo:", "    image: mongo:7", '    ports: ["27017:27017"]',
      '    volumes: ["mongo-data:/data/db"]', "volumes:", "  mongo-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Mariadb, name: "MariaDB", category: "databases", icon: "mariadb", summary: "mariadb:11.4 · :3306",
    args: "--datadir=/var/lib/mysql/data",
    compose: [
      "services:", "  mariadb:", "    image: mariadb:11.4", '    ports: ["3306:3306"]',
      '    volumes: ["mariadb-data:/var/lib/mysql"]', "    environment:", `      MARIADB_ROOT_PASSWORD: "${PLACEHOLDER_PASSWORD}"`,
      "volumes:", "  mariadb-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Mssql, name: "MS SQL Server", category: "databases", icon: "mssql", summary: "mssql:2022 · :1433",
    compose: [
      "services:", "  mssql:", "    image: mcr.microsoft.com/mssql/server:2022-latest", '    ports: ["1433:1433"]',
      '    volumes: ["mssql-data:/var/opt/mssql"]', "    environment:", '      ACCEPT_EULA: "Y"', `      MSSQL_SA_PASSWORD: "${MSSQL_PLACEHOLDER_PASSWORD}"`,
      "volumes:", "  mssql-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Elasticsearch, name: "Elasticsearch", category: "analytics", icon: "elasticsearch", summary: "elasticsearch:8 · :9200",
    compose: [
      "services:", "  elasticsearch:", "    image: docker.elastic.co/elasticsearch/elasticsearch:8.15.0", '    ports: ["9200:9200"]',
      '    volumes: ["es-data:/usr/share/elasticsearch/data"]', "    environment:",
      '      discovery.type: "single-node"', '      xpack.security.enabled: "false"', '      ES_JAVA_OPTS: "-Xms512m -Xmx512m"',
      "volumes:", "  es-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Couchdb, name: "CouchDB", category: "databases", icon: "couchdb", summary: "couchdb:3.3 · :5984",
    compose: [
      "services:", "  couchdb:", "    image: couchdb:3.3", '    ports: ["5984:5984"]',
      '    volumes: ["couchdb-data:/opt/couchdb/data"]', "    environment:", '      COUCHDB_USER: "admin"', `      COUCHDB_PASSWORD: "${PLACEHOLDER_PASSWORD}"`,
      "volumes:", "  couchdb-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Influxdb, name: "InfluxDB", category: "analytics", icon: "influxdb", summary: "influxdb:2.7 · :8086",
    compose: [
      "services:", "  influxdb:", "    image: influxdb:2.7", '    ports: ["8086:8086"]',
      '    volumes: ["influxdb-data:/var/lib/influxdb2"]', "volumes:", "  influxdb-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Clickhouse, name: "ClickHouse", category: "analytics", icon: "clickhouse", summary: "clickhouse:24 · :8123",
    compose: [
      "services:", "  clickhouse:", "    image: clickhouse/clickhouse-server:24.8", '    ports: ["8123:8123", "9000:9000"]',
      '    volumes: ["clickhouse-data:/var/lib/clickhouse"]', "    environment:", `      CLICKHOUSE_PASSWORD: "${PLACEHOLDER_PASSWORD}"`,
      "volumes:", "  clickhouse-data: {}", "",
    ].join("\n"),
  },
];

export function getBlockById(id: string): BlockPreset | undefined {
  return blockCatalog.find((b) => b.id === id);
}

export { BlockId } from "./types";
