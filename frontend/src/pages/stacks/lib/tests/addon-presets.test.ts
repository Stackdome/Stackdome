import { describe, it, expect } from "vitest";
import {
  CRED_FIELDS,
  CLUSTER_WIDE_FIELDS,
} from "../addon-presets";

describe("CRED_FIELDS", () => {
  it("lists all 8 fields the backend supports", () => {
    expect(CRED_FIELDS).toEqual([
      "host",
      "port",
      "username",
      "password",
      "database",
      "sslmode",
      "connectionString",
      "caCertificate",
    ]);
  });
});

describe("CLUSTER_WIDE_FIELDS", () => {
  it("contains the four cluster-scoped credentials", () => {
    expect(CLUSTER_WIDE_FIELDS.has("host")).toBe(true);
    expect(CLUSTER_WIDE_FIELDS.has("port")).toBe(true);
    expect(CLUSTER_WIDE_FIELDS.has("sslmode")).toBe(true);
    expect(CLUSTER_WIDE_FIELDS.has("caCertificate")).toBe(true);
  });

  it("does not contain database-scoped credentials", () => {
    expect(CLUSTER_WIDE_FIELDS.has("username")).toBe(false);
    expect(CLUSTER_WIDE_FIELDS.has("password")).toBe(false);
    expect(CLUSTER_WIDE_FIELDS.has("database")).toBe(false);
    expect(CLUSTER_WIDE_FIELDS.has("connectionString")).toBe(false);
  });
});
