import { describe, it, expect } from "vitest";
import { ImportSource, isPrefillSource } from "../import-source";

describe("import-source", () => {
  it("exposes Blocks as a source", () => {
    expect(ImportSource.Blocks).toBe("blocks");
  });

  it("treats blocks, template, and docker-compose as prefill sources", () => {
    expect(isPrefillSource(ImportSource.Blocks)).toBe(true);
    expect(isPrefillSource(ImportSource.Template)).toBe(true);
    expect(isPrefillSource(ImportSource.DockerCompose)).toBe(true);
  });

  it("rejects unknown sources", () => {
    expect(isPrefillSource("github")).toBe(false);
    expect(isPrefillSource(undefined)).toBe(false);
  });
});
