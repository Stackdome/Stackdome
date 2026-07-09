import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/api/client", () => ({
  default: { get: vi.fn(), post: vi.fn(), delete: vi.fn() },
}));

import api from "@/api/client";
import {
  listPreviewEnvs,
  getPreviewEnv,
  createPreviewEnv,
  deletePreviewEnv,
  syncPreviewEnv,
  TERMINAL_PHASES,
} from "../preview-envs";

const ORG = "org1";
const TEAM = "default";
const BASE = `/organizations/${ORG}/teams/${TEAM}/preview-stacks`;

beforeEach(() => vi.clearAllMocks());

describe("preview-envs api", () => {
  it("lists envs filtered by config", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { items: [], total: 0 } });
    await listPreviewEnvs(ORG, TEAM, { configId: "c1", page: 1, pageSize: 20 });
    expect(api.get).toHaveBeenCalledWith(BASE, { params: { config_id: "c1", page: 1, page_size: 20 } });
  });

  it("lists envs without filter", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { items: [], total: 0 } });
    await listPreviewEnvs(ORG, TEAM);
    expect(api.get).toHaveBeenCalledWith(BASE, { params: {} });
  });

  it("gets one env", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "p1", status: { phase: "Ready" } } });
    const out = await getPreviewEnv(ORG, TEAM, "p1");
    expect(api.get).toHaveBeenCalledWith(`${BASE}/p1`);
    expect(out.status?.phase).toBe("Ready");
  });

  it("creates an env", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "p1" } });
    const input = { config_id: "c1", pr_number: 42, branch: "feat/login" };
    await createPreviewEnv(ORG, TEAM, input);
    expect(api.post).toHaveBeenCalledWith(BASE, input);
  });

  it("deletes an env (202 returns the env)", async () => {
    (api.delete as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "p1", status: { phase: "Deleting" } } });
    const out = await deletePreviewEnv(ORG, TEAM, "p1");
    expect(api.delete).toHaveBeenCalledWith(`${BASE}/p1`);
    expect(out.status?.phase).toBe("Deleting");
  });

  it("syncs an env with force", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "p1" } });
    await syncPreviewEnv(ORG, TEAM, "p1", { force_sync: true });
    expect(api.post).toHaveBeenCalledWith(`${BASE}/p1/sync`, { force_sync: true });
  });

  it("syncs with empty body by default", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "p1" } });
    await syncPreviewEnv(ORG, TEAM, "p1");
    expect(api.post).toHaveBeenCalledWith(`${BASE}/p1/sync`, {});
  });

  it("exports terminal phases", () => {
    expect(TERMINAL_PHASES).toEqual(["Ready", "Failed"]);
  });
});
