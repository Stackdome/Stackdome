import { describe, it, expect } from "vitest";
import {
  ImportSource,
  PREFILL_IMPORT_SOURCES,
  isPrefillSource,
} from "../import-source";

describe("import-source", () => {
  it("treats docker-compose and template as prefill sources", () => {
    expect(isPrefillSource(ImportSource.DockerCompose)).toBe(true);
    expect(isPrefillSource(ImportSource.Template)).toBe(true);
  });

  it("rejects unknown or missing sources", () => {
    expect(isPrefillSource("something-else")).toBe(false);
    expect(isPrefillSource(undefined)).toBe(false);
    expect(isPrefillSource(null)).toBe(false);
  });

  it("lists every prefill source", () => {
    expect(PREFILL_IMPORT_SOURCES).toEqual([
      ImportSource.DockerCompose,
      ImportSource.Template,
    ]);
  });
});
