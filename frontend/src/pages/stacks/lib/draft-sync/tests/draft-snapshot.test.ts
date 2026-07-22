import { describe, it, expect } from "vitest";
import { draftToSnapshot } from "../draft-snapshot";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";

const validResource = {
  name: "web",
  sourceType: "image" as const,
  source: { image: { ref: "nginx:1" } },
  execution_config: {
    command: "sh -c 'echo hi'",
    environment_variables: [
      { from: "stack" as const, name: "MODE", value: "prod" },
      { from: "secret" as const, name: "TOKEN", secretId: "s-1", secretKey: "token" },
    ],
  },
};

describe("draftToSnapshot", () => {
  it("emits server-shaped resources, volumes, and connections as arrays", () => {
    const vol = { name: "web-data", sourceType: "None" as const, spec: { size: "1Gi", access_mode: "ReadWriteOnce" }, labels: [] };
    const snap = draftToSnapshot({ resources: [validResource], volumes: [vol] } as unknown as EditSessionDraft);
    expect(snap.resources.map((r) => r.name)).toEqual(["web"]);
    expect(snap.resources[0].execution_config?.command).toEqual(["sh", "-c", "echo hi"]);
    // secret row rides as a connection, not an env var
    expect(snap.resources[0].execution_config?.environment_variables).toEqual([
      { name: "MODE", value: "prod" },
    ]);
    expect(snap.connections).toHaveLength(1);
    expect(snap.volumes.map((v) => v.name)).toEqual(["web-data"]);
  });

  it("omits invalid (held) resources — matching what autosave would persist", () => {
    const broken = { name: "api", sourceType: "image" as const, source: { image: { ref: "" } } };
    const snap = draftToSnapshot({ resources: [validResource, broken], volumes: [] } as unknown as EditSessionDraft);
    expect(snap.resources.map((r) => r.name)).toEqual(["web"]);
  });

  it("returns empty collections for an empty draft", () => {
    const snap = draftToSnapshot({ resources: [], volumes: [] } as unknown as EditSessionDraft);
    expect(snap).toEqual({ resources: [], volumes: [], connections: [] });
  });
});
