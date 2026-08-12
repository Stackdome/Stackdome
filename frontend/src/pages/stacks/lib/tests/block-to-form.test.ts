import { describe, it, expect } from "vitest";
import { blockToResources, addBlockToStack, emptyStack } from "../block-to-form";
import { blockCatalog, getBlockById, BlockId } from "@/pages/stacks/data/blocks/registry";
import { DATA_BLOCK_CATEGORIES } from "@/pages/stacks/data/blocks/types";

describe("block-to-form", () => {
  it("every catalog block converts to a resource named after the block, without throwing", () => {
    for (const block of blockCatalog) {
      expect(() => blockToResources(block), `block ${block.id} should convert`).not.toThrow();
      const { resources } = blockToResources(block);
      expect(resources.length, `block ${block.id} resource count`).toBeGreaterThan(0);
      expect(resources[0].name, `block ${block.id} resource name`).toBe(block.id);
    }
  });

  it("converts the postgres block into one resource + one volume", () => {
    const { resources, volumes } = blockToResources(getBlockById(BlockId.Postgres)!);
    expect(resources).toHaveLength(1);
    expect(resources[0].name).toBe("postgres");
    expect(resources[0].source?.image?.ref).toBe("postgres:16");
    expect(volumes).toHaveLength(1);
    expect(volumes[0].name).toBe("pgdata");
  });

  it("converts a generic web block into an empty-image resource and no volumes", () => {
    const { resources, volumes } = blockToResources(getBlockById(BlockId.Web)!);
    expect(resources).toHaveLength(1);
    expect(resources[0].name).toBe("web");
    expect(resources[0].sourceType).toBe("image");
    expect(resources[0].source?.image?.ref).toBe("");
    expect(volumes).toHaveLength(0);
  });

  it("web block exposes port 80 as public http in the API Port shape", () => {
    const { resources } = blockToResources(getBlockById(BlockId.Web)!);
    expect(resources[0].ports).toEqual([
      { name: "http-80", number: 80, protocol: "http", exposed_to_public: true },
    ]);
  });

  it("data-store ports are internal tcp, never public http", () => {
    for (const block of blockCatalog.filter((b) => DATA_BLOCK_CATEGORIES.has(b.category))) {
      const { resources } = blockToResources(block);
      for (const r of resources) {
        for (const p of r.ports ?? []) {
          expect(p.exposed_to_public, `${block.id} port ${p.number} must be internal`).toBe(false);
          expect(p.protocol, `${block.id} port ${p.number} protocol`).toBe("tcp");
          expect(p.name).toBe(`tcp-${p.number}`);
        }
      }
    }
  });

  it("fills empty password env values with a generated secret so containers boot", () => {
    for (const id of [BlockId.Postgres, BlockId.Mysql, BlockId.Mariadb, BlockId.Mssql, BlockId.Couchdb]) {
      const { resources } = blockToResources(getBlockById(id)!);
      const env = (resources[0].execution_config?.environment_variables ?? []) as { name?: string; value?: string }[];
      for (const row of env) {
        expect(row.value, `${id} env ${row.name} must not be empty`).not.toBe("");
      }
      // MSSQL requires upper+lower+digit complexity — the generator guarantees it.
      const values = env.map((r) => r.value ?? "");
      for (const v of values.filter((val) => val.startsWith("Sd1"))) {
        expect(v).toMatch(/[A-Z]/);
        expect(v).toMatch(/[a-z]/);
        expect(v).toMatch(/[0-9]/);
        expect(v.length).toBeGreaterThanOrEqual(16);
      }
    }
  });

  it("two adds of the same data block generate distinct passwords", () => {
    const pg = getBlockById(BlockId.Postgres)!;
    const first = blockToResources(pg).resources[0].execution_config?.environment_variables as { value?: string }[];
    const second = blockToResources(pg).resources[0].execution_config?.environment_variables as { value?: string }[];
    expect(first[0].value).not.toBe(second[0].value);
  });

  it("de-duplicates resource names when the same block is added twice", () => {
    let stack = emptyStack();
    stack = addBlockToStack(stack, getBlockById(BlockId.Postgres)!);
    stack = addBlockToStack(stack, getBlockById(BlockId.Postgres)!);
    expect(stack.spec.stack_resources.map((r) => r.name)).toEqual(["postgres", "postgres-2"]);
  });

  it("rewires volume_mounts.source_volume_name after de-duplicating volumes (no orphan mounts)", () => {
    let stack = emptyStack();
    stack = addBlockToStack(stack, getBlockById(BlockId.Postgres)!);
    stack = addBlockToStack(stack, getBlockById(BlockId.Postgres)!);

    // Both volumes must be present and uniquely named
    expect(stack.spec.volumes?.map((v) => v.name)).toEqual(["pgdata", "pgdata-2"]);

    // The second resource's volume mount must point at the renamed volume, not the original
    const postgres2 = stack.spec.stack_resources.find((r) => r.name === "postgres-2");
    const mount = postgres2?.volume_mounts?.[0];
    expect(mount?.source_volume_name).toBe("pgdata-2");
  });

  /**
   * A block is self-contained: when de-duplication renames its resources, its
   * own references must follow the copy. Otherwise the added block binds to the
   * same-named resource already in the stack, which validates cleanly and
   * deploys wrong.
   *
   * Catalog blocks are single-service today (`BlockPreset.compose`), so nothing
   * shipped can trigger this. The multi-service preset here is what makes the
   * wiring testable, and it fails the day the carry is removed.
   */
  it("repoints a de-duplicated block's own references at its copy", () => {
    const pair = {
      id: "pair",
      name: "Pair",
      category: getBlockById(BlockId.Postgres)!.category,
      icon: "box",
      summary: "two services",
      compose: [
        "services:",
        "  postgres:",
        "    image: postgres:16",
        "  app:",
        "    image: app:1",
        "    depends_on:",
        "      - postgres",
        "    environment:",
        '      DB_HOST: "${postgres.host}"',
      ].join("\n"),
    };

    let stack = emptyStack();
    stack = addBlockToStack(stack, getBlockById(BlockId.Postgres)!);
    stack = addBlockToStack(stack, pair);

    expect(stack.spec.stack_resources.map((r) => r.name)).toContain("postgres-2");
    const app = stack.spec.stack_resources.find((r) => r.name === "app");
    expect(app?.depends_on).toEqual(["postgres-2"]);
    const rows = (app?.execution_config?.environment_variables ?? []) as Array<{ resourceName?: string }>;
    expect(rows.find((row) => row.resourceName != null)?.resourceName).toBe("postgres-2");
  });
});
