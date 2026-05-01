import { describe, it, expect } from "vitest";
import {
  CRED_FIELDS,
  CLUSTER_WIDE_FIELDS,
  DEFAULT_ENV_NAMES,
  applyPreset,
} from "../src/pages/stacks/lib/addon-presets";

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

describe("DEFAULT_ENV_NAMES", () => {
  it("provides defaults for every credential field", () => {
    expect(DEFAULT_ENV_NAMES.host).toBe("PG_HOST");
    expect(DEFAULT_ENV_NAMES.port).toBe("PG_PORT");
    expect(DEFAULT_ENV_NAMES.username).toBe("PG_USER");
    expect(DEFAULT_ENV_NAMES.password).toBe("PG_PASS");
    expect(DEFAULT_ENV_NAMES.database).toBe("PG_DB");
    expect(DEFAULT_ENV_NAMES.sslmode).toBe("PG_SSLMODE");
    expect(DEFAULT_ENV_NAMES.connectionString).toBe("DATABASE_URL");
    expect(DEFAULT_ENV_NAMES.caCertificate).toBe("PG_CA_CERT");
  });
});

describe("applyPreset", () => {
  it("postgres-conventions selects host/port/username/password/database with default names", () => {
    const result = applyPreset("postgres-conventions");
    expect([...result.selected].sort()).toEqual(
      ["database", "host", "password", "port", "username"].sort(),
    );
    expect(result.envNames.host).toBe("PG_HOST");
    expect(result.envNames.password).toBe("PG_PASS");
  });

  it("connection-string selects only connectionString as DATABASE_URL", () => {
    const result = applyPreset("connection-string");
    expect([...result.selected]).toEqual(["connectionString"]);
    expect(result.envNames.connectionString).toBe("DATABASE_URL");
  });

  it("clear returns empty selection", () => {
    const result = applyPreset("clear");
    expect(result.selected.size).toBe(0);
    expect(Object.keys(result.envNames)).toEqual([]);
  });
});
