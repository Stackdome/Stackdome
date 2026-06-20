// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/api/client", () => ({
  default: { get: vi.fn(), post: vi.fn() },
}));

import api from "@/api/client";
import { listReleases, getRelease, createRelease, rollbackRelease, cancelRelease } from "../releases";

const ORG = "org1", TEAM = "team1", STACK = "s1";
const BASE = `/organizations/${ORG}/teams/${TEAM}/stacks/${STACK}/releases`;

beforeEach(() => vi.clearAllMocks());

describe("releases api", () => {
  it("lists releases", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { items: [{ id: "r1" }], total: 1 } });
    const out = await listReleases(ORG, TEAM, STACK);
    expect(api.get).toHaveBeenCalledWith(BASE);
    expect(out.items?.[0].id).toBe("r1");
  });

  it("gets one release", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "r1" } });
    await getRelease(ORG, TEAM, STACK, "r1");
    expect(api.get).toHaveBeenCalledWith(`${BASE}/r1`);
  });

  it("creates a deploy release with empty body", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "r2" } });
    await createRelease(ORG, TEAM, STACK);
    expect(api.post).toHaveBeenCalledWith(BASE, {});
  });

  it("rolls back via from_release_id", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "r3" } });
    await rollbackRelease(ORG, TEAM, STACK, "r1");
    expect(api.post).toHaveBeenCalledWith(BASE, { from_release_id: "r1" });
  });

  it("cancels a release", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({ data: undefined });
    await cancelRelease(ORG, TEAM, STACK, "r1");
    expect(api.post).toHaveBeenCalledWith(`${BASE}/r1/cancel`);
  });
});
