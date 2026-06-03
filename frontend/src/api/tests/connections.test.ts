import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("../client", () => ({
  default: {
    get: vi.fn(() => Promise.resolve({ data: { items: [] } })),
    post: vi.fn(() => Promise.resolve({ data: { id: "c1" } })),
    put: vi.fn(() => Promise.resolve({ data: { id: "c1" } })),
    delete: vi.fn(() => Promise.resolve({ data: undefined })),
  },
}));

import api from "../client";
import {
  listStackConnections,
  createStackConnection,
  updateStackConnection,
  deleteStackConnection,
} from "../connections";

const ORG = "org1";
const TEAM = "default";
const STACK = "stack1";
const base = `/organizations/${ORG}/teams/${TEAM}/stacks/${STACK}/connections`;

describe("connections api client", () => {
  beforeEach(() => vi.clearAllMocks());

  it("lists connections at the team-scoped stack path", async () => {
    await listStackConnections(ORG, TEAM, STACK);
    expect(api.get).toHaveBeenCalledWith(base);
  });

  it("creates a connection with POST to the collection path", async () => {
    const conn = { kind: "env" as const, from: { type: "secret" as const, id: "s1" }, to: { type: "stack_resource" as const, name: "web" } };
    await createStackConnection(ORG, TEAM, STACK, conn);
    expect(api.post).toHaveBeenCalledWith(base, conn);
  });

  it("updates a connection with PUT to the item path", async () => {
    const conn = { id: "c1", kind: "env" as const, from: { type: "secret" as const, id: "s1" }, to: { type: "stack_resource" as const, name: "web" } };
    await updateStackConnection(ORG, TEAM, STACK, "c1", conn);
    expect(api.put).toHaveBeenCalledWith(`${base}/c1`, conn);
  });

  it("deletes a connection with DELETE to the item path", async () => {
    await deleteStackConnection(ORG, TEAM, STACK, "c1");
    expect(api.delete).toHaveBeenCalledWith(`${base}/c1`);
  });
});
