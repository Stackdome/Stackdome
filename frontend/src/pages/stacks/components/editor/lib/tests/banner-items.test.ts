import { describe, it, expect } from "vitest";
import { buildBannerItems } from "@/pages/stacks/components/editor/lib/banner-items";
import type { ParsedFieldError } from "@/api/errors";

function fe(field: string, message = "msg", code = "x"): ParsedFieldError {
  return { field, code, message };
}

describe("buildBannerItems", () => {
  it("returns [] when there are no deploy field errors", () => {
    expect(buildBannerItems([], [])).toEqual([]);
  });

  it("labels resource errors with the resource name and routes env fields to the environment tab", () => {
    const errors = [fe("spec.stack_resources[0].execution_config.env[2].name", "name required")];
    const items = buildBannerItems(errors, [{ name: "web" }]);
    expect(items[0]).toMatchObject({ label: "web", resourceIndex: 0, tab: "environment" });
  });

  it("falls back to 'Resource N' when the resource has no name and routes non-env fields to configuration", () => {
    const errors = [fe("spec.stack_resources[1].ports[1].protocol", "bad protocol")];
    const items = buildBannerItems(errors, [{ name: "web" }, {}]);
    expect(items[0]).toMatchObject({ label: "Resource 2", resourceIndex: 1, tab: "configuration" });
  });

  it("emits stack-level rows (name, settings, connections, unmapped) without a jump index", () => {
    const errors = [
      fe("name", "stack name invalid"),
      fe("spec.settings", "retention too high"),
      fe("spec.connections[0]", "bad connection"),
      fe("spec.something_new", "huh"),
    ];
    const items = buildBannerItems(errors, []);
    expect(items).toEqual([
      { label: "Stack name", message: "stack name invalid" },
      { label: "Stack settings", message: "retention too high" },
      { label: "Connection", message: "bad connection" },
      { label: "spec.something_new", message: "huh" },
    ]);
    for (const item of items) {
      expect(item.resourceIndex).toBeUndefined();
    }
  });
});
