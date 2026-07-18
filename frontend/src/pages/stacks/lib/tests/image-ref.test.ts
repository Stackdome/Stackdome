import { describe, it, expect } from "vitest";
import { splitImageRef, joinImageRef, dockerHostsEqual } from "../image-ref";

describe("splitImageRef", () => {
  it("returns null host for bare Docker Hub refs", () => {
    expect(splitImageRef("nginx:latest")).toEqual({ host: null, remainder: "nginx:latest" });
    expect(splitImageRef("acme/api:1.2")).toEqual({ host: null, remainder: "acme/api:1.2" });
  });

  it("detects a registry host by dot, port, or localhost", () => {
    expect(splitImageRef("ghcr.io/acme/api:1.4.2")).toEqual({ host: "ghcr.io", remainder: "acme/api:1.4.2" });
    expect(splitImageRef("localhost:5000/app")).toEqual({ host: "localhost:5000", remainder: "app" });
    expect(splitImageRef("registry:5000/app")).toEqual({ host: "registry:5000", remainder: "app" });
  });

  it("handles empty and host-only refs without crashing", () => {
    expect(splitImageRef("")).toEqual({ host: null, remainder: "" });
    expect(splitImageRef("ghcr.io/")).toEqual({ host: "ghcr.io", remainder: "" });
  });
});

describe("joinImageRef", () => {
  it("prefixes host when present", () => {
    expect(joinImageRef("ghcr.io", "acme/api:1")).toBe("ghcr.io/acme/api:1");
  });
  it("returns remainder alone for null host", () => {
    expect(joinImageRef(null, "nginx:latest")).toBe("nginx:latest");
  });
  it("does not double-slash an empty remainder", () => {
    expect(joinImageRef("ghcr.io", "")).toBe("ghcr.io/");
  });
});

describe("dockerHostsEqual", () => {
  it("treats Docker Hub aliases as equal", () => {
    expect(dockerHostsEqual("docker.io", "index.docker.io")).toBe(true);
    expect(dockerHostsEqual("registry-1.docker.io", "docker.io")).toBe(true);
  });
  it("compares other hosts case-insensitively", () => {
    expect(dockerHostsEqual("GHCR.IO", "ghcr.io")).toBe(true);
    expect(dockerHostsEqual("ghcr.io", "quay.io")).toBe(false);
  });
});
