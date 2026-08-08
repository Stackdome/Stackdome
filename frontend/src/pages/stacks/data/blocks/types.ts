export type BlockCategory = "services" | "databases" | "analytics";

/** Categories whose blocks are internal dependencies (tcp ports, generated passwords). */
export const DATA_BLOCK_CATEGORIES: ReadonlySet<BlockCategory> = new Set(["databases", "analytics"]);

export interface BlockCategoryMeta {
  id: BlockCategory;
  label: string; //  marker label
}

export const BlockId = {
  Web: "web",
  Custom: "custom",
  Postgres: "postgres",
  Redis: "redis",
  Mysql: "mysql",
  Mongo: "mongo",
  Mariadb: "mariadb",
  Mssql: "mssql",
  Elasticsearch: "elasticsearch",
  Couchdb: "couchdb",
  Influxdb: "influxdb",
  Clickhouse: "clickhouse",
} as const;
export type BlockId = (typeof BlockId)[keyof typeof BlockId];

export interface BlockPreset {
  id: string;
  name: string;
  category: BlockCategory;
  icon: string;      // BlockGlyph key: a lucide name ("globe", "box") or a brand key ("postgres", "redis", "mysql", "mongo")
  summary: string;   // mono one-liner shown on the card, e.g. "postgres:16 · :5432 · pgdata"
  compose?: string;  // 1-service docker-compose YAML; omitted for generic blocks
}
