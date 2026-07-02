import { describe, it, expect } from "vitest";
import { buildDesiredState } from "../desired-state";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";

const validResource = {
  name: "web",
  sourceType: "image" as const,
  image_spec: { image: "nginx:1" },
  execution_config: {
    environment_variables: [
      { from: "stack" as const, name: "MODE", value: "prod" },
      { from: "secret" as const, name: "TOKEN", secretId: "s-1", secretKey: "token" },
    ],
  },
};

describe("buildDesiredState", () => {
  it("includes valid resources keyed by name, with connections split out", () => {
    const d = buildDesiredState({ resources: [validResource], volumes: [] } as unknown as EditSessionDraft);
    expect([...d.resources.keys()]).toEqual(["web"]);
    const web = d.resources.get("web")!;
    // secret row does not ride as an env var
    expect(web.execution_config?.environment_variables).toEqual([{ name: "MODE", value: "prod" }]);
    expect(d.connections.size).toBe(1);
    const conn = [...d.connections.values()][0];
    expect(conn.from).toEqual({ type: "secret", id: "s-1" });
    expect(d.held.size).toBe(0);
  });

  it("holds an invalid named resource instead of dropping it", () => {
    const invalid = { name: "api", sourceType: "image" as const, image_spec: { image: "" } };
    const d = buildDesiredState({ resources: [invalid], volumes: [] } as unknown as EditSessionDraft);
    expect(d.resources.has("api")).toBe(false);
    expect(d.held.has("api")).toBe(true);
    expect(d.resourceIssues.get(0)?.length).toBeGreaterThan(0);
  });

  it("skips unnamed invalid resources without holding anything", () => {
    const d = buildDesiredState({ resources: [{ name: "", sourceType: "image" as const }], volumes: [] } as unknown as EditSessionDraft);
    expect(d.resources.size).toBe(0);
    expect(d.held.size).toBe(0);
    expect(d.resourceIssues.has(0)).toBe(true);
  });

  it("excludes connections whose rows are in-progress", () => {
    const r = { ...validResource, execution_config: { environment_variables: [{ from: "secret" as const, name: "X", secretId: "", secretKey: "" }] } };
    const d = buildDesiredState({ resources: [r], volumes: [] } as unknown as EditSessionDraft);
    expect(d.connections.size).toBe(0);
  });

  it("includes named volumes converted to API shape and skips unnamed ones", () => {
    const vol = { name: "web-data", sourceType: "None" as const, spec: { size: "1Gi", access_mode: "ReadWriteOnce" }, labels: [] };
    const d = buildDesiredState({ resources: [validResource], volumes: [vol, { name: "" }] } as unknown as EditSessionDraft);
    expect([...d.volumes.keys()]).toEqual(["web-data"]);
    expect((d.volumes.get("web-data") as Record<string, unknown>).sourceType).toBeUndefined();
  });
});
