import { describe, it, expect } from "vitest";
import {
  splitEnvRows,
  connectionsToEnvRows,
  buildDesiredConnections,
  mountsToConnections,
  connectionsToMounts,
} from "../connection-mapping";
import type { FormEnvRow, FormMountRow } from "../connection-mapping";
import type { StackConnection } from "@/api/connections";
import { connectionIdentityKey } from "@/pages/stacks/lib/draft-sync/server-state";

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
      { target: { type: "env", name: "LOCKBOX_MASTER_KEY" }, value: { output: "LOCKBOX_MASTER_KEY" } },
      { target: { type: "env", name: "SECRET_KEY_BASE" }, value: { output: "SECRET_KEY_BASE" } },
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

describe("connectionsToEnvRows", () => {
  it("expands a secret connection into secret rows", () => {
    const conns: StackConnection[] = [{
      id: "c1", kind: "env",
      from: { type: "secret", id: "s1" },
      to: { type: "stack_resource", name: "web" },
      mappings: [{ target: { type: "env", name: "LOCKBOX_MASTER_KEY" }, value: { output: "LOCKBOX_MASTER_KEY" } }],
    }];
    expect(connectionsToEnvRows("web", conns)).toEqual([
      { from: "secret", name: "LOCKBOX_MASTER_KEY", secretId: "s1", secretKey: "LOCKBOX_MASTER_KEY" },
    ]);
  });

  it("expands an addon connection into addon rows using config database", () => {
    const conns: StackConnection[] = [{
      id: "c2", kind: "env",
      from: { type: "addon/postgres", id: "a1" },
      to: { type: "stack_resource", name: "web" },
      config: { database: "tooljet", superuser: false },
      mappings: [{ target: { type: "env", name: "PG_HOST" }, value: { output: "host" } }],
    }];
    expect(connectionsToEnvRows("web", conns)).toEqual([
      { from: "addon", name: "PG_HOST", addonId: "a1", database: "tooljet", superuser: false, credField: "host" },
    ]);
  });

  it("expands a resource connection into resource rows", () => {
    const conns: StackConnection[] = [{
      id: "c3", kind: "env",
      from: { type: "stack_resource", name: "mailhog" },
      to: { type: "stack_resource", name: "web" },
      mappings: [{ target: { type: "env", name: "SMTP_DOMAIN" }, value: { output: "host" } }],
    }];
    expect(connectionsToEnvRows("web", conns)).toEqual([
      { from: "resource", name: "SMTP_DOMAIN", resourceName: "mailhog", output: "host" },
    ]);
  });

  it("ignores connections whose `to` is a different resource", () => {
    const conns: StackConnection[] = [{
      id: "c4", kind: "env",
      from: { type: "secret", id: "s1" },
      to: { type: "stack_resource", name: "other" },
      mappings: [{ target: { type: "env", name: "X" }, value: { output: "X" } }],
    }];
    expect(connectionsToEnvRows("web", conns)).toEqual([]);
  });

  it("ignores non-env connections", () => {
    const conns: StackConnection[] = [{
      id: "v1", kind: "volume_mount",
      from: { type: "volume", name: "vol" },
      to: { type: "stack_resource", name: "web" },
      mappings: [{ target: { type: "file", path: "/x" }, value: { output: "host" } }],
    }];
    expect(connectionsToEnvRows("web", conns)).toEqual([]);
  });

  it("expands a superuser addon connection with no database", () => {
    const conns: StackConnection[] = [{
      id: "c5", kind: "env",
      from: { type: "addon/postgres", id: "a1" },
      to: { type: "stack_resource", name: "web" },
      config: { superuser: true },
      mappings: [{ target: { type: "env", name: "PG_URL" }, value: { output: "url" } }],
    }];
    expect(connectionsToEnvRows("web", conns)).toEqual([
      { from: "addon", name: "PG_URL", addonId: "a1", database: undefined, superuser: true, credField: "url" },
    ]);
  });
});

describe("buildDesiredConnections", () => {
  it("collects connections across every resource", () => {
    const resources = [
      { name: "web", rows: [{ from: "secret", name: "X", secretId: "s1", secretKey: "X" }] as FormEnvRow[] },
      { name: "api", rows: [{ from: "resource", name: "H", resourceName: "web", output: "host" }] as FormEnvRow[] },
    ];
    expect(buildDesiredConnections(resources)).toHaveLength(2);
  });
});

describe("mountsToConnections", () => {
  it("converts a complete mount row to a volume_mount connection", () => {
    const mounts: FormMountRow[] = [{ source_volume_name: "web-data", target_path: "/data" }];
    const conns = mountsToConnections("web", mounts);
    expect(conns).toHaveLength(1);
    expect(conns[0]).toMatchObject({
      kind: "volume_mount",
      from: { type: "volume", name: "web-data" },
      to: { type: "stack_resource", name: "web" },
      config: { mount_path: "/data" },
    });
    expect((conns[0].config as { sub_path?: string }).sub_path).toBeUndefined();
  });

  it("includes sub_path in config when provided", () => {
    const mounts: FormMountRow[] = [{ source_volume_name: "v", source_sub_path: "logs", target_path: "/logs" }];
    const [conn] = mountsToConnections("api", mounts);
    expect(conn.config).toEqual({ mount_path: "/logs", sub_path: "logs" });
  });

  it("includes read_only in config when set on row", () => {
    const mounts: FormMountRow[] = [{ source_volume_name: "v", target_path: "/data", read_only: true }];
    const [conn] = mountsToConnections("web", mounts);
    expect(conn.config).toEqual({ mount_path: "/data", read_only: true });
  });

  it("omits read_only from config when undefined on row", () => {
    const mounts: FormMountRow[] = [{ source_volume_name: "v", target_path: "/data" }];
    const [conn] = mountsToConnections("web", mounts);
    expect((conn.config as Record<string, unknown>).read_only).toBeUndefined();
  });

  it("skips in-progress rows missing source_volume_name", () => {
    const mounts: FormMountRow[] = [{ source_volume_name: "", target_path: "/data" }];
    expect(mountsToConnections("web", mounts)).toHaveLength(0);
  });

  it("skips in-progress rows missing target_path", () => {
    const mounts: FormMountRow[] = [{ source_volume_name: "vol", target_path: "" }];
    expect(mountsToConnections("web", mounts)).toHaveLength(0);
  });

  it("skips rows where both fields are absent", () => {
    const mounts: FormMountRow[] = [{}];
    expect(mountsToConnections("web", mounts)).toHaveLength(0);
  });

  it("produces one connection per complete row", () => {
    const mounts: FormMountRow[] = [
      { source_volume_name: "v1", target_path: "/a" },
      { source_volume_name: "v2", target_path: "/b" },
      { source_volume_name: "", target_path: "/c" }, // in-progress — skipped
    ];
    expect(mountsToConnections("web", mounts)).toHaveLength(2);
  });
});

describe("connectionsToMounts", () => {
  const mountConn = (
    volName: string,
    resourceName: string,
    mountPath: string,
    subPath?: string,
    readOnly?: boolean,
  ): StackConnection => ({
    id: "vm-1",
    kind: "volume_mount",
    from: { type: "volume", name: volName },
    to: { type: "stack_resource", name: resourceName },
    config: {
      mount_path: mountPath,
      ...(subPath ? { sub_path: subPath } : {}),
      ...(readOnly !== undefined ? { read_only: readOnly } : {}),
    },
  });

  it("expands a volume_mount connection into a mount row", () => {
    const rows = connectionsToMounts("web", [mountConn("web-data", "web", "/data")]);
    expect(rows).toEqual([{ source_volume_name: "web-data", source_sub_path: "", target_path: "/data" }]);
  });

  it("populates source_sub_path from config.sub_path", () => {
    const rows = connectionsToMounts("api", [mountConn("v", "api", "/logs", "app/logs")]);
    expect(rows[0].source_sub_path).toBe("app/logs");
  });

  it("ignores connections whose `to` is a different resource", () => {
    expect(connectionsToMounts("web", [mountConn("v", "other", "/data")])).toHaveLength(0);
  });

  it("ignores non-volume_mount connections", () => {
    const envConn: StackConnection = {
      kind: "env",
      from: { type: "secret", id: "s1" },
      to: { type: "stack_resource", name: "web" },
      mappings: [{ target: { type: "env", name: "X" }, value: { output: "X" } }],
    };
    expect(connectionsToMounts("web", [envConn])).toHaveLength(0);
  });

  it("carries read_only: true from config into row", () => {
    const rows = connectionsToMounts("web", [mountConn("v", "web", "/data", undefined, true)]);
    expect(rows[0].read_only).toBe(true);
  });

  it("omits read_only from row when absent in config", () => {
    const rows = connectionsToMounts("web", [mountConn("v", "web", "/data")]);
    expect(rows[0].read_only).toBeUndefined();
  });

  it("round-trips through mountsToConnections → connectionsToMounts", () => {
    const original: FormMountRow[] = [
      { source_volume_name: "vol-a", source_sub_path: "sub", target_path: "/mnt/a" },
      { source_volume_name: "vol-b", source_sub_path: "", target_path: "/mnt/b" },
    ];
    const conns = mountsToConnections("web", original);
    const recovered = connectionsToMounts("web", conns);
    expect(recovered).toEqual(original);
  });

  it("round-trips read_only: true through mountsToConnections → connectionsToMounts", () => {
    const original: FormMountRow[] = [
      { source_volume_name: "vol-a", source_sub_path: "logs", target_path: "/mnt/a", read_only: true },
    ];
    const conns = mountsToConnections("web", original);
    const recovered = connectionsToMounts("web", conns);
    expect(recovered).toEqual(original);
    expect(recovered[0].read_only).toBe(true);
  });

  it("two mounts of the same volume to the same resource produce the same identity key (collide like the backend)", () => {
    // The backend enforces uniqueness by (kind, from, to), so two rows pointing
    // to the same volume produce the same frontend identity key regardless of config.
    const a = mountsToConnections("web", [{ source_volume_name: "vol", target_path: "/a" }])[0];
    const b = mountsToConnections("web", [{ source_volume_name: "vol", target_path: "/b" }])[0];
    expect(connectionIdentityKey(a)).toBe(connectionIdentityKey(b));
  });
});

