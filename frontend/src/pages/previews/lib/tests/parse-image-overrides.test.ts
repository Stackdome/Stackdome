import { describe, it, expect } from "vitest";
import { parseImageOverrides } from "../parse-image-overrides";

describe("parseImageOverrides", () => {
  it("parses one resource=image line per row", () => {
    expect(
      parseImageOverrides("web=registry.example.com/web:sha-1\napi=registry.example.com/api:sha-2"),
    ).toEqual({
      web: "registry.example.com/web:sha-1",
      api: "registry.example.com/api:sha-2",
    });
  });

  it("ignores blank lines", () => {
    expect(parseImageOverrides("web=image:tag\n\n\n")).toEqual({ web: "image:tag" });
  });

  it("drops lines with no '=' or an empty key", () => {
    expect(parseImageOverrides("not-a-pair\n=image:tag\nweb=image:tag")).toEqual({
      web: "image:tag",
    });
  });

  it("trims whitespace around key and value", () => {
    expect(parseImageOverrides("  web  =  image:tag  ")).toEqual({ web: "image:tag" });
  });

  it("returns undefined for empty input", () => {
    expect(parseImageOverrides("")).toBeUndefined();
    expect(parseImageOverrides("   \n  \n")).toBeUndefined();
  });
});
