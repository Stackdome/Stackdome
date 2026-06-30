// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import { renderHook, act, cleanup } from "@testing-library/react";
import { useDockerComposeImport } from "../use-docker-compose-import";

const navigate = vi.fn();
vi.mock("react-router-dom", () => ({ useNavigate: () => navigate }));
vi.mock("@/components/ui/use-toast", () => ({
  useToast: () => ({ toast: vi.fn(), toasts: [], dismiss: vi.fn() }),
}));

afterEach(() => {
  cleanup();
  navigate.mockReset();
});

const VALID_YAML = [
  "services:",
  "  web:",
  "    image: nginx:latest",
  '    ports: ["80:80"]',
  "",
].join("\n");

describe("useDockerComposeImport", () => {
  it("handleImport returns true and navigates on valid docker-compose YAML", async () => {
    const { result } = renderHook(() => useDockerComposeImport());
    let ok: boolean;
    await act(async () => {
      ok = await result.current.handleImport(VALID_YAML);
    });
    expect(ok!).toBe(true);
    expect(navigate).toHaveBeenCalledWith("/stacks/create", expect.objectContaining({ state: expect.anything() }));
  });

  it("handleImport returns false and does NOT navigate on invalid YAML", async () => {
    const { result } = renderHook(() => useDockerComposeImport());
    let ok: boolean;
    await act(async () => {
      ok = await result.current.handleImport("not: valid: yaml: ::");
    });
    expect(ok!).toBe(false);
    expect(navigate).not.toHaveBeenCalled();
    expect(result.current.error).not.toBeNull();
  });
});
