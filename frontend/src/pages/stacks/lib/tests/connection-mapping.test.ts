import { describe, it, expect } from "vitest";
import { secretOutputAccessor, parseSecretOutput, splitEnvRows } from "../connection-mapping";
import type { FormEnvRow } from "../connection-mapping";

describe("secretOutputAccessor", () => {
  it("uses dot form for simple keys", () => {
    expect(secretOutputAccessor("LOCKBOX_MASTER_KEY")).toBe("key.LOCKBOX_MASTER_KEY");
    expect(secretOutputAccessor("a1_B2")).toBe("key.a1_B2");
    // Backend `^[A-Za-z0-9_]+$` treats digit-leading keys as simple.
    expect(secretOutputAccessor("1leading")).toBe("key.1leading");
  });

  it("uses bracket form for keys with special characters", () => {
    expect(secretOutputAccessor("my-key")).toBe("key['my-key']");
    expect(secretOutputAccessor("has space")).toBe("key['has space']");
  });

  it("escapes single quotes and backslashes in bracket form", () => {
    expect(secretOutputAccessor("a'b")).toBe("key['a\\'b']");
    expect(secretOutputAccessor("a\\b")).toBe("key['a\\\\b']");
  });
});

describe("parseSecretOutput", () => {
  it("reverses the dot form", () => {
    expect(parseSecretOutput("key.LOCKBOX_MASTER_KEY")).toBe("LOCKBOX_MASTER_KEY");
    expect(parseSecretOutput("key.1leading")).toBe("1leading");
  });
  it("reverses the bracket form, unescaping", () => {
    expect(parseSecretOutput("key['my-key']")).toBe("my-key");
    expect(parseSecretOutput("key['a\\'b']")).toBe("a'b");
    expect(parseSecretOutput("key['a\\\\b']")).toBe("a\\b");
  });
  it("returns null for unrecognized accessors", () => {
    expect(parseSecretOutput("host")).toBeNull();
    expect(parseSecretOutput("")).toBeNull();
    expect(parseSecretOutput("key")).toBeNull();
    expect(parseSecretOutput("key.")).toBeNull();
    expect(parseSecretOutput("key['unclosed")).toBeNull();
  });

  it("round-trips every key form", () => {
    const keys = [
      "LOCKBOX_MASTER_KEY",
      "tls.crt",
      "my-key",
      "has space",
      "a'b",
      "a\\b",
      "1leading",
    ];
    for (const k of keys) {
      expect(parseSecretOutput(secretOutputAccessor(k))).toBe(k);
    }
  });
});

describe("splitEnvRows", () => {
  it("emits literal rows as env vars with value", () => {
    const rows: FormEnvRow[] = [{ from: "stack", name: "NODE_ENV", value: "production" }];
    const { envVars, connections } = splitEnvRows("web", rows);
    expect(envVars).toEqual([{ name: "NODE_ENV", value: "production" }]);
    expect(connections).toEqual([]);
  });

  it("emits self rows as env vars with self_output", () => {
    const rows: FormEnvRow[] = [{ from: "self", name: "TOOLJET_HOST", selfOutput: "public.http.url" }];
    const { envVars, connections } = splitEnvRows("web", rows);
    expect(envVars).toEqual([{ name: "TOOLJET_HOST", self_output: "public.http.url" }]);
    expect(connections).toEqual([]);
  });

  it("groups all secret rows for one secret into a single connection", () => {
    const rows: FormEnvRow[] = [
      { from: "secret", name: "LOCKBOX_MASTER_KEY", secretId: "s1", secretKey: "LOCKBOX_MASTER_KEY" },
      { from: "secret", name: "SECRET_KEY_BASE", secretId: "s1", secretKey: "SECRET_KEY_BASE" },
    ];
    const { envVars, connections } = splitEnvRows("web", rows);
    expect(envVars).toEqual([]);
    expect(connections).toHaveLength(1);
    expect(connections[0]).toMatchObject({
      kind: "env",
      from: { type: "secret", id: "s1" },
      to: { type: "stack_resource", name: "web" },
    });
    expect(connections[0].config).toBeUndefined();
    expect(connections[0].mappings).toEqual([
      { target: { type: "env", name: "LOCKBOX_MASTER_KEY" }, value: { output: "key.LOCKBOX_MASTER_KEY" } },
      { target: { type: "env", name: "SECRET_KEY_BASE" }, value: { output: "key.SECRET_KEY_BASE" } },
    ]);
  });

  it("groups 5 addon fields for one (addon, database) into a single connection with config", () => {
    const rows: FormEnvRow[] = [
      { from: "addon", name: "PG_HOST", addonId: "a1", database: "tooljet", superuser: false, credField: "host" },
      { from: "addon", name: "PG_PORT", addonId: "a1", database: "tooljet", superuser: false, credField: "port" },
      { from: "addon", name: "PG_USER", addonId: "a1", database: "tooljet", superuser: false, credField: "username" },
      { from: "addon", name: "PG_PASS", addonId: "a1", database: "tooljet", superuser: false, credField: "password" },
      { from: "addon", name: "PG_DB", addonId: "a1", database: "tooljet", superuser: false, credField: "database" },
    ];
    const { connections } = splitEnvRows("web", rows);
    expect(connections).toHaveLength(1);
    expect(connections[0]).toMatchObject({
      kind: "env",
      from: { type: "addon/postgres", id: "a1" },
      to: { type: "stack_resource", name: "web" },
      config: { database: "tooljet", superuser: false },
    });
    expect(connections[0].mappings).toHaveLength(5);
    expect(connections[0].mappings![0]).toEqual({
      target: { type: "env", name: "PG_HOST" },
      value: { output: "host" },
    });
  });

  it("splits addon rows that differ by database into separate connections", () => {
    const rows: FormEnvRow[] = [
      { from: "addon", name: "A_HOST", addonId: "a1", database: "db1", superuser: false, credField: "host" },
      { from: "addon", name: "B_HOST", addonId: "a1", database: "db2", superuser: false, credField: "host" },
    ];
    const { connections } = splitEnvRows("web", rows);
    expect(connections).toHaveLength(2);
  });

  it("groups resource rows for one source resource into a single connection", () => {
    const rows: FormEnvRow[] = [
      { from: "resource", name: "SMTP_DOMAIN", resourceName: "mailhog", output: "host" },
    ];
    const { connections } = splitEnvRows("web", rows);
    expect(connections).toHaveLength(1);
    expect(connections[0]).toMatchObject({
      kind: "env",
      from: { type: "stack_resource", name: "mailhog" },
      to: { type: "stack_resource", name: "web" },
    });
    expect(connections[0].config).toBeUndefined();
    expect(connections[0].mappings).toEqual([
      { target: { type: "env", name: "SMTP_DOMAIN" }, value: { output: "host" } },
    ]);
  });

  it("uses superuser config and omits database when superuser is true", () => {
    const rows: FormEnvRow[] = [
      { from: "addon", name: "PG_URL", addonId: "a1", superuser: true, credField: "url" },
    ];
    const { connections } = splitEnvRows("web", rows);
    expect(connections[0].config).toEqual({ superuser: true });
  });

  it("returns both env vars and connections for mixed rows", () => {
    const rows: FormEnvRow[] = [
      { from: "stack", name: "NODE_ENV", value: "production" },
      { from: "secret", name: "K", secretId: "s1", secretKey: "K" },
    ];
    const { envVars, connections } = splitEnvRows("web", rows);
    expect(envVars).toEqual([{ name: "NODE_ENV", value: "production" }]);
    expect(connections).toHaveLength(1);
    expect(connections[0].from).toEqual({ type: "secret", id: "s1" });
  });

  it("skips in-progress rows missing required identifiers", () => {
    const rows: FormEnvRow[] = [
      { from: "secret", name: "X", secretId: "", secretKey: "K" },
      { from: "addon", name: "Y", addonId: "a1", database: "d", superuser: false }, // no credField
      { from: "resource", name: "Z", resourceName: "", output: "host" },
    ];
    const { envVars, connections } = splitEnvRows("web", rows);
    expect(envVars).toEqual([]);
    expect(connections).toEqual([]);
  });
});
