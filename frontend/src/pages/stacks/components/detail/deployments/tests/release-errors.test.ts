import { describe, expect, it } from "vitest";
import { releaseValidationBannerItems } from "../release-errors";
import type { StackRelease } from "@/api/releases";
import type { StackResource } from "@/api/stacks";

function release(validationErrors: StackRelease["validation_errors"]): StackRelease {
  return { state: "Failed", validation_errors: validationErrors } as StackRelease;
}

const resources = [{ name: "web" }, { name: "worker" }] as StackResource[];

describe("releaseValidationBannerItems", () => {
  it("maps errors to jumpable banner items", () => {
    const items = releaseValidationBannerItems(
      release([
        { resource_name: "web", field: "source.image.ref", code: "image_not_found", message: 'image "ghcr.io/a/web:v42" not found in registry' },
        { resource_name: "worker", field: "source.git.push.repository", code: "git_auth_failed", message: "git auth failed" },
      ]),
      resources,
    );
    expect(items).toHaveLength(2);
    expect(items[0]).toMatchObject({ label: "web", resourceIndex: 0, tab: "configuration" });
    expect(items[0].message).toContain("image_not_found");
    expect(items[0].message).toContain('image "ghcr.io/a/web:v42" not found in registry');
    expect(items[1]).toMatchObject({ label: "worker", resourceIndex: 1 });
  });

  it("leaves resourceIndex undefined when resource_name isn't found", () => {
    const items = releaseValidationBannerItems(
      release([{ resource_name: "gone", field: "source.image.ref", code: "image_not_found", message: "not found" }]),
      resources,
    );
    expect(items[0].resourceIndex).toBeUndefined();
  });

  it("labels stack-level errors (empty resource_name) as Stack", () => {
    const items = releaseValidationBannerItems(
      release([{ resource_name: "", field: "", code: "stack_settings_invalid", message: "bad settings" }]),
      resources,
    );
    expect(items[0]).toMatchObject({ label: "Stack", resourceIndex: undefined });
  });

  it("returns an empty array when the release has no validation_errors", () => {
    expect(releaseValidationBannerItems(release(undefined), resources)).toEqual([]);
  });

  it("derives the environment tab from an env field prefix", () => {
    const items = releaseValidationBannerItems(
      release([{ resource_name: "web", field: "execution_config.env[0].name", code: "env_name_required", message: "name required" }]),
      resources,
    );
    expect(items[0].tab).toBe("environment");
  });
});
