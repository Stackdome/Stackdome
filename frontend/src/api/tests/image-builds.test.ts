import { describe, it, expect } from "vitest";
import {
  buildImageBuildLogStreamUrl,
  isBuildJobCreated,
  BUILD_JOB_CREATED_CONDITION,
} from "@/api/image-builds";

describe("buildImageBuildLogStreamUrl", () => {
  it("builds the stream url with params", () => {
    const url = buildImageBuildLogStreamUrl("o1", "p1", "s1", "b1", { follow: true, tail: 200 });
    expect(url).toContain("/organizations/o1/projects/p1/stacks/s1/builds/b1/logs");
    expect(url).toContain("follow=true");
    expect(url).toContain("tail=200");
  });

  it("omits the query string without params", () => {
    expect(buildImageBuildLogStreamUrl("o1", "p1", "s1", "b1")).not.toContain("?");
  });
});

describe("isBuildJobCreated", () => {
  it("is true only for a True BuildJobCreated condition", () => {
    expect(
      isBuildJobCreated({
        status: { conditions: [{ type: BUILD_JOB_CREATED_CONDITION, status: "True" }] },
      } as never),
    ).toBe(true);
    expect(
      isBuildJobCreated({
        status: { conditions: [{ type: BUILD_JOB_CREATED_CONDITION, status: "False" }] },
      } as never),
    ).toBe(false);
    expect(isBuildJobCreated({ status: {} } as never)).toBe(false);
    expect(isBuildJobCreated({} as never)).toBe(false);
  });
});
