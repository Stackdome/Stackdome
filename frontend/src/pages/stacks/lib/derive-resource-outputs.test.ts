import { describe, it, expect } from "vitest";
import { deriveResourceOutputNames } from "./derive-resource-outputs";

// Mirror of pkg/models/output_descriptor.go::StackResourceOutputDescriptors.
// Keep these cases in sync with that function.
describe("deriveResourceOutputNames", () => {
  it("returns only host when the resource has no ports", () => {
    expect(deriveResourceOutputNames({ ports: [] })).toEqual(["host"]);
    expect(deriveResourceOutputNames({})).toEqual(["host"]);
    expect(deriveResourceOutputNames({ ports: null })).toEqual(["host"]);
  });

  it("adds port.<name> and url.<name> for a private port", () => {
    expect(
      deriveResourceOutputNames({ ports: [{ name: "http-80", exposed_to_public: false }] }),
    ).toEqual(["host", "port.http-80", "url.http-80"]);
  });

  it("adds public.<name>.host and public.<name>.url for a public port", () => {
    expect(
      deriveResourceOutputNames({ ports: [{ name: "http-80", exposed_to_public: true }] }),
    ).toEqual([
      "host",
      "port.http-80",
      "url.http-80",
      "public.http-80.host",
      "public.http-80.url",
    ]);
  });

  it("handles multiple ports (mixed public/private) in order", () => {
    expect(
      deriveResourceOutputNames({
        ports: [
          { name: "api", exposed_to_public: true },
          { name: "grpc", exposed_to_public: false },
        ],
      }),
    ).toEqual([
      "host",
      "port.api",
      "url.api",
      "public.api.host",
      "public.api.url",
      "port.grpc",
      "url.grpc",
    ]);
  });

  it("skips ports with no name (cannot form a stable output key)", () => {
    expect(
      deriveResourceOutputNames({ ports: [{ exposed_to_public: true }] }),
    ).toEqual(["host"]);
  });
});
