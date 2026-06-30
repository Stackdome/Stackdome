import { describe, it, expect } from "vitest";
import { blockToResources, addBlockToStack, emptyStack } from "../block-to-form";
import { getBlockById, BlockId } from "@/pages/stacks/data/blocks/registry";

describe("block-to-form", () => {
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
});
