import { describe, it, expect } from "vitest";
import { parseOutputKey, groupOutputs } from "./parse-output-key";

describe("parseOutputKey", () => {
  it("maps clean single-port keys to internal labels", () => {
    expect(parseOutputKey("host")).toEqual({ group: "internal", label: "Host", key: "host" });
    expect(parseOutputKey("port")).toEqual({ group: "internal", label: "Port", key: "port" });
    expect(parseOutputKey("url")).toEqual({ group: "internal", label: "URL", key: "url" });
  });

  it("maps public_ keys to the public group", () => {
    expect(parseOutputKey("public_host")).toEqual({ group: "public", label: "Public Host", key: "public_host" });
    expect(parseOutputKey("public_url")).toEqual({ group: "public", label: "Public URL", key: "public_url" });
  });

  it("extracts the port suffix without confusing url vs public_url", () => {
    expect(parseOutputKey("url.3306")).toEqual({ group: "internal", label: "URL", port: "3306", key: "url.3306" });
    expect(parseOutputKey("public_url.web")).toEqual({ group: "public", label: "Public URL", port: "web", key: "public_url.web" });
  });

  it("falls back to the raw key for anything unrecognized", () => {
    expect(parseOutputKey("weird")).toEqual({ group: "internal", label: "weird", key: "weird" });
  });
});

describe("groupOutputs", () => {
  it("splits into internal and public buckets preserving order", () => {
    const g = groupOutputs(["host", "port", "url", "public_host", "public_url"]);
    expect(g.internal.map((o) => o.key)).toEqual(["host", "port", "url"]);
    expect(g.public.map((o) => o.key)).toEqual(["public_host", "public_url"]);
  });
});
