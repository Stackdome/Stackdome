import { BlockId, type BlockCategoryMeta, type BlockPreset } from "./types";

export const BLOCK_CATEGORY_META: BlockCategoryMeta[] = [
  { id: "services", label: "SERVICES", note: "your code, running" },
  { id: "data", label: "DATA STORES", note: "run in your cluster" },
];

export const blockCatalog: BlockPreset[] = [
  { id: BlockId.Web, name: "Web service", category: "services", icon: "globe", summary: "your image · :8080" },
  { id: BlockId.Custom, name: "Custom", category: "services", icon: "box", summary: "empty container shape" },
  {
    id: BlockId.Postgres, name: "Postgres", category: "data", icon: "postgres", summary: "postgres:16 · :5432 · pgdata",
    compose: [
      "services:",
      "  postgres:",
      "    image: postgres:16",
      '    ports: ["5432:5432"]',
      '    volumes: ["pgdata:/var/lib/postgresql/data"]',
      "    environment:",
      '      POSTGRES_PASSWORD: ""',
      "volumes:",
      "  pgdata: {}",
      "",
    ].join("\n"),
  },
  {
    id: BlockId.Redis, name: "Redis", category: "data", icon: "redis", summary: "redis:7 · :6379",
    compose: ["services:", "  redis:", "    image: redis:7", '    ports: ["6379:6379"]', ""].join("\n"),
  },
  {
    id: BlockId.Mysql, name: "MySQL", category: "data", icon: "mysql", summary: "mysql:8 · :3306 · mysql-data",
    compose: [
      "services:", "  mysql:", "    image: mysql:8", '    ports: ["3306:3306"]',
      '    volumes: ["mysql-data:/var/lib/mysql"]', "    environment:", '      MYSQL_ROOT_PASSWORD: ""',
      "volumes:", "  mysql-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Mongo, name: "MongoDB", category: "data", icon: "mongo", summary: "mongo:7 · :27017 · mongo-data",
    compose: [
      "services:", "  mongo:", "    image: mongo:7", '    ports: ["27017:27017"]',
      '    volumes: ["mongo-data:/data/db"]', "volumes:", "  mongo-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Mariadb, name: "MariaDB", category: "data", icon: "mariadb", summary: "mariadb:11.4 · :3306 · mariadb-data",
    compose: [
      "services:", "  mariadb:", "    image: mariadb:11.4", '    ports: ["3306:3306"]',
      '    volumes: ["mariadb-data:/var/lib/mysql"]', "    environment:", '      MARIADB_ROOT_PASSWORD: ""',
      "volumes:", "  mariadb-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Mssql, name: "MS SQL Server", category: "data", icon: "mssql", summary: "mssql:2022 · :1433 · mssql-data",
    compose: [
      "services:", "  mssql:", "    image: mcr.microsoft.com/mssql/server:2022-latest", '    ports: ["1433:1433"]',
      '    volumes: ["mssql-data:/var/opt/mssql"]', "    environment:", '      ACCEPT_EULA: "Y"', '      MSSQL_SA_PASSWORD: ""',
      "volumes:", "  mssql-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Elasticsearch, name: "Elasticsearch", category: "data", icon: "elasticsearch", summary: "elasticsearch:8 · :9200 · es-data",
    compose: [
      "services:", "  elasticsearch:", "    image: docker.elastic.co/elasticsearch/elasticsearch:8.15.0", '    ports: ["9200:9200"]',
      '    volumes: ["es-data:/usr/share/elasticsearch/data"]', "    environment:",
      '      discovery.type: "single-node"', '      xpack.security.enabled: "false"', '      ES_JAVA_OPTS: "-Xms512m -Xmx512m"',
      "volumes:", "  es-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Couchdb, name: "CouchDB", category: "data", icon: "couchdb", summary: "couchdb:3.3 · :5984 · couchdb-data",
    compose: [
      "services:", "  couchdb:", "    image: couchdb:3.3", '    ports: ["5984:5984"]',
      '    volumes: ["couchdb-data:/opt/couchdb/data"]', "    environment:", '      COUCHDB_USER: "admin"', '      COUCHDB_PASSWORD: ""',
      "volumes:", "  couchdb-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Influxdb, name: "InfluxDB", category: "data", icon: "influxdb", summary: "influxdb:2.7 · :8086 · influxdb-data",
    compose: [
      "services:", "  influxdb:", "    image: influxdb:2.7", '    ports: ["8086:8086"]',
      '    volumes: ["influxdb-data:/var/lib/influxdb2"]', "volumes:", "  influxdb-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Clickhouse, name: "ClickHouse", category: "data", icon: "clickhouse", summary: "clickhouse:24 · :8123 · clickhouse-data",
    compose: [
      "services:", "  clickhouse:", "    image: clickhouse/clickhouse-server:24.8", '    ports: ["8123:8123", "9000:9000"]',
      '    volumes: ["clickhouse-data:/var/lib/clickhouse"]', "    environment:", '      CLICKHOUSE_PASSWORD: ""',
      "volumes:", "  clickhouse-data: {}", "",
    ].join("\n"),
  },
];

export function getBlockById(id: string): BlockPreset | undefined {
  return blockCatalog.find((b) => b.id === id);
}

export { BlockId } from "./types";
