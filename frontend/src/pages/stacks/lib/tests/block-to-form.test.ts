import { describe, it, expect } from "vitest";
import { blockToResources, addBlockToStack, emptyStack } from "../block-to-form";
import { blockCatalog, getBlockById, BlockId } from "@/pages/stacks/data/blocks/registry";

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
    expect(resources[0].image_spec?.image).toBe("postgres:16");
    expect(volumes).toHaveLength(1);
    expect(volumes[0].name).toBe("pgdata");
  });

  it("converts a generic web block into an empty-image resource and no volumes", () => {
    const { resources, volumes } = blockToResources(getBlockById(BlockId.Web)!);
    expect(resources).toHaveLength(1);
    expect(resources[0].name).toBe("web");
    expect(resources[0].sourceType).toBe("image");
    expect(resources[0].image_spec?.image).toBe("");
    expect(volumes).toHaveLength(0);
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
});
