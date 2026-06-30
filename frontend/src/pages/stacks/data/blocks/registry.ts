import { BlockId, type BlockCategoryMeta, type BlockPreset } from "./types";

export const BLOCK_CATEGORY_META: BlockCategoryMeta[] = [
  { id: "services", label: "SERVICES", note: "your code, running" },
  { id: "data", label: "DATA STORES", note: "run in your cluster" },
];

export const blockCatalog: BlockPreset[] = [
  { id: BlockId.Web, name: "Web service", category: "services", icon: "globe", summary: "your image · :8080" },
  { id: BlockId.Custom, name: "Custom", category: "services", icon: "box", summary: "empty container shape" },
  {
    id: BlockId.Postgres, name: "Postgres", category: "data", icon: "database", summary: "postgres:16 · :5432 · pgdata",
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
    id: BlockId.Redis, name: "Redis", category: "data", icon: "zap", summary: "redis:7 · :6379",
    compose: ["services:", "  redis:", "    image: redis:7", '    ports: ["6379:6379"]', ""].join("\n"),
  },
  {
    id: BlockId.Mysql, name: "MySQL", category: "data", icon: "database", summary: "mysql:8 · :3306 · mysql-data",
    compose: [
      "services:", "  mysql:", "    image: mysql:8", '    ports: ["3306:3306"]',
      '    volumes: ["mysql-data:/var/lib/mysql"]', "    environment:", '      MYSQL_ROOT_PASSWORD: ""',
      "volumes:", "  mysql-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Mongo, name: "MongoDB", category: "data", icon: "database", summary: "mongo:7 · :27017 · mongo-data",
    compose: [
      "services:", "  mongo:", "    image: mongo:7", '    ports: ["27017:27017"]',
      '    volumes: ["mongo-data:/data/db"]', "volumes:", "  mongo-data: {}", "",
    ].join("\n"),
  },
];

export function getBlockById(id: string): BlockPreset | undefined {
  return blockCatalog.find((b) => b.id === id);
}

export { BlockId } from "./types";
