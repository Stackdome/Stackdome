import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/api/client", () => ({
  default: { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn() },
}));

import api from "@/api/client";
import {
  listPreviewConfigs,
  getPreviewConfig,
  createPreviewConfig,
  updatePreviewConfig,
  deletePreviewConfig,
} from "../preview-configs";

const ORG = "org1";
const PROJECT = "default";
const BASE = `/organizations/${ORG}/projects/${PROJECT}/stack-preview-configs`;

beforeEach(() => vi.clearAllMocks());

describe("preview-configs api", () => {
  it("lists configs with pagination", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { items: [{ id: "c1" }], total: 1 } });
    const out = await listPreviewConfigs(ORG, PROJECT, 2, 50);
    expect(api.get).toHaveBeenCalledWith(BASE, { params: { page: 2, page_size: 50 } });
    expect(out.items?.[0].id).toBe("c1");
  });

  it("gets one config", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "c1", name: "webapp" } });
    const out = await getPreviewConfig(ORG, PROJECT, "c1");
    expect(api.get).toHaveBeenCalledWith(`${BASE}/c1`);
    expect(out.name).toBe("webapp");
  });

  it("creates a config", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "c1" } });
    const input = { name: "webapp", git_repository: { repo_url: "https://github.com/acme/webapp", base_branch: "main" } };
    await createPreviewConfig(ORG, PROJECT, input);
    expect(api.post).toHaveBeenCalledWith(BASE, input);
  });

  it("updates a config", async () => {
    (api.put as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "c1" } });
    await updatePreviewConfig(ORG, PROJECT, "c1", { stackfile_path: "deploy/stackfile.yaml" });
    expect(api.put).toHaveBeenCalledWith(`${BASE}/c1`, { stackfile_path: "deploy/stackfile.yaml" });
  });

  it("deletes a config", async () => {
    (api.delete as ReturnType<typeof vi.fn>).mockResolvedValue({});
    await deletePreviewConfig(ORG, PROJECT, "c1");
    expect(api.delete).toHaveBeenCalledWith(`${BASE}/c1`);
  });
});
