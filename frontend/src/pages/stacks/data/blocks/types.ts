export type BlockCategory = "services" | "data";

export interface BlockCategoryMeta {
  id: BlockCategory;
  label: string; // uppercase marker label
  note: string;  // muted sub-note
}

export const BlockId = {
  Web: "web",
  Custom: "custom",
  Postgres: "postgres",
  Redis: "redis",
  Mysql: "mysql",
  Mongo: "mongo",
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
