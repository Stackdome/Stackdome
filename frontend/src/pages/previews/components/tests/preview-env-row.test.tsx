// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { PreviewEnvRow } from "../preview-env-row";
import type { PreviewStack } from "@/api/preview-envs";

afterEach(() => cleanup());

const base: PreviewStack = {
  id: "p1",
  pr_number: "42",
  branch: "feat/login",
  commit: "a3f9c21deadbeef",
  source: "manual",
  status: { phase: "Ready", outputs: { urls: [{ resource: "web", url: "https://pr-42.example.com" }] } },
};

describe("PreviewEnvRow", () => {
  it("renders PR, branch, short commit, phase and URL", () => {
    render(<PreviewEnvRow env={base} onSync={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText("PR #42")).toBeTruthy();
    expect(screen.getByText("feat/login")).toBeTruthy();
    expect(screen.getByText("a3f9c21")).toBeTruthy();
    expect(screen.getByText("Ready")).toBeTruthy();
    const link = screen.getByRole("link", { name: /web/i });
    expect(link.getAttribute("href")).toBe("https://pr-42.example.com");
    expect(link.getAttribute("target")).toBe("_blank");
  });

  it("expands failure reason with stackfile hint", () => {
    const failed: PreviewStack = {
      ...base,
      status: { phase: "Failed", reason: "StackfileNotFound", message: "no stackfile at path" },
    };
    render(<PreviewEnvRow env={failed} onSync={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText("Failed")).toBeTruthy();
    expect(screen.getByText(/no stackfile at path/i)).toBeTruthy();
    expect(screen.getByText(/check the stackfile path/i)).toBeTruthy();
  });
});
