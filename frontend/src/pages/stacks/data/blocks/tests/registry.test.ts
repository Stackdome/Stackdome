import { describe, it, expect } from "vitest";
import { blockCatalog, getBlockById, BlockId, BLOCK_CATEGORY_META } from "../registry";

describe("block registry", () => {
  it("ships the expected catalog (services + v1 data sources)", () => {
    expect(blockCatalog.map((b) => b.id).sort()).toEqual(
      [
        BlockId.Web,
        BlockId.Custom,
        BlockId.Postgres,
        BlockId.Redis,
        BlockId.Mysql,
        BlockId.Mongo,
        BlockId.Mariadb,
        BlockId.Mssql,
        BlockId.Elasticsearch,
        BlockId.Couchdb,
        BlockId.Influxdb,
        BlockId.Clickhouse,
      ].sort(),
    );
  });

  it("gives known-software blocks a compose snippet and generic blocks none", () => {
    expect(getBlockById(BlockId.Postgres)?.compose).toContain("postgres:");
    expect(getBlockById(BlockId.Web)?.compose).toBeUndefined();
    expect(getBlockById(BlockId.Custom)?.compose).toBeUndefined();
  });

  it("only uses the two v1 categories", () => {
    const cats = new Set(blockCatalog.map((b) => b.category));
    expect([...cats].sort()).toEqual(["data", "services"]);
    expect(BLOCK_CATEGORY_META.map((c) => c.id)).toEqual(["services", "data"]);
  });
});
