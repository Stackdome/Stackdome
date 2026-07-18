import { describe, it, expect } from "vitest";
import { deriveResourceOutputNames } from "./derive-resource-outputs";

// Mirror of pkg/models/output_descriptor.go::StackResourceOutputDescriptors.
describe("deriveResourceOutputNames", () => {
  it("returns only host when the resource has no ports", () => {
    expect(deriveResourceOutputNames({ ports: [] })).toEqual(["host"]);
    expect(deriveResourceOutputNames({})).toEqual(["host"]);
    expect(deriveResourceOutputNames({ ports: null })).toEqual(["host"]);
  });

  it("drops the port suffix for a single private port", () => {
    expect(
      deriveResourceOutputNames({ ports: [{ name: "3306", exposed_to_public: false }] }),
    ).toEqual(["host", "port", "url"]);
  });

  it("adds unsuffixed public_host/public_url for a single public port", () => {
    expect(
      deriveResourceOutputNames({ ports: [{ name: "80", exposed_to_public: true }] }),
    ).toEqual(["host", "port", "url", "public_host", "public_url"]);
  });

  it("suffixes every per-port key with the port name when multi-port", () => {
    expect(
      deriveResourceOutputNames({
        ports: [
          { name: "80", exposed_to_public: true },
          { name: "grpc", exposed_to_public: false },
        ],
      }),
    ).toEqual([
      "host",
      "port.80", "url.80", "public_host.80", "public_url.80",
      "port.grpc", "url.grpc",
    ]);
  });

  it("skips ports with no name and does not count them toward multi-port", () => {
    expect(
      deriveResourceOutputNames({ ports: [{ exposed_to_public: true }] }),
    ).toEqual(["host"]);
    // one named + one unnamed = single-port scheme (unnamed dropped)
    expect(
      deriveResourceOutputNames({ ports: [{ name: "3306" }, { exposed_to_public: true }] }),
    ).toEqual(["host", "port", "url"]);
  });
});
