import { describe, it, expect } from "vitest";
import { emptyDraftSeed, buildDraftFormData } from "../draft-seed";

describe("emptyDraftSeed", () => {
  it("returns an empty, named-blank seed", () => {
    expect(emptyDraftSeed()).toEqual({
      name: "",
      labels: [],
      resources: [],
      volumes: [],
      linkedAddonIds: [],
    });
  });
});

describe("buildDraftFormData", () => {
  it("assembles FormStackData from draft name/labels/resources/volumes", () => {
    const resources = [{ name: "api", sourceType: "image", image_spec: { image: "nginx:1" } }] as never;
    const out = buildDraftFormData("my-app", [{ key: "k", value: "v" }], resources, []);
    expect(out).toEqual({
      name: "my-app",
      labels: [{ key: "k", value: "v" }],
      spec: { stack_resources: resources, volumes: [] },
    });
  });
});
