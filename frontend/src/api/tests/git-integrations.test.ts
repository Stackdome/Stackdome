import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/api/client", () => ({
  default: { get: vi.fn(), post: vi.fn(), delete: vi.fn() },
}));

import api from "@/api/client";
import {
  listGitIntegrations,
  deleteGitIntegration,
  verifyGitIntegration,
  createGitHubAppManifest,
  listInstallations,
  searchRepositories,
  getRepository,
  listRepositoryBranches,
} from "../git-integrations";

const ORG = "org1";
const BASE = `/organizations/${ORG}/git-integrations`;

beforeEach(() => vi.clearAllMocks());

describe("git-integrations api", () => {
  it("lists integrations", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { items: [{ id: "gi1" }], total: 1 } });
    const out = await listGitIntegrations(ORG);
    expect(api.get).toHaveBeenCalledWith(BASE);
    expect(out.items?.[0].id).toBe("gi1");
  });

  it("deletes an integration", async () => {
    (api.delete as ReturnType<typeof vi.fn>).mockResolvedValue({});
    await deleteGitIntegration(ORG, "gi1");
    expect(api.delete).toHaveBeenCalledWith(`${BASE}/gi1`);
  });

  it("verifies against a repo url", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({});
    await verifyGitIntegration(ORG, "gi1", "https://github.com/acme/webapp");
    expect(api.post).toHaveBeenCalledWith(`${BASE}/gi1/verify`, { repo_url: "https://github.com/acme/webapp" });
  });

  it("creates the GitHub App manifest flow", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({
      data: { manifest: { name: "x" }, github_url: "https://github.com/settings/apps/new?state=s", state: "s" },
    });
    const out = await createGitHubAppManifest(ORG);
    expect(api.post).toHaveBeenCalledWith(`${BASE}/github/manifest`);
    expect(out.github_url).toContain("github.com");
  });

  it("lists installations with refresh", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { items: [], total: 0 } });
    await listInstallations(ORG, "gi1", true);
    expect(api.get).toHaveBeenCalledWith(`${BASE}/gi1/installations`, { params: { refresh: true } });
  });

  it("searches repositories with query and page", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { items: [], page: 1, total_count: 0, has_next: false } });
    await searchRepositories(ORG, "gi1", { query: "web", page: 2 });
    expect(api.get).toHaveBeenCalledWith(`${BASE}/gi1/repositories`, { params: { query: "web", page: 2 } });
  });

  it("gets a repository", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { full_name: "acme/webapp", default_branch: "main" } });
    const out = await getRepository(ORG, "gi1", "acme", "webapp");
    expect(api.get).toHaveBeenCalledWith(`${BASE}/gi1/repositories/acme/webapp`);
    expect(out.default_branch).toBe("main");
  });

  it("lists branches", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { items: ["main", "dev"], total: 2 } });
    const out = await listRepositoryBranches(ORG, "gi1", "acme", "webapp");
    expect(api.get).toHaveBeenCalledWith(`${BASE}/gi1/repositories/acme/webapp/branches`);
    expect(out.items).toContain("main");
  });
});
