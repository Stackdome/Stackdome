// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("@/api/preview-configs", () => ({
  createPreviewConfig: vi.fn(),
}));
vi.mock("@/api/git-integrations", () => ({
  listRepositoryBranches: vi.fn().mockResolvedValue({ items: ["main", "develop"], total: 2 }),
}));
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));
vi.mock("@/hooks/use-resource-projects", () => ({
  useResourceProjects: () => ({ projects: [], projectNameById: () => undefined, defaultProjectName: "default" }),
}));

import { createPreviewConfig } from "@/api/preview-configs";
import { AxiosError } from "axios";
import { ConfigurePhase } from "../configure-phase";

const repo = {
  fullName: "acme/webapp",
  cloneUrl: "https://github.com/acme/webapp.git",
  defaultBranch: "main",
  integrationId: "gi1",
};

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("ConfigurePhase", () => {
  it("prefills name from repo and creates the config", async () => {
    (createPreviewConfig as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "c1" });
    const onCreated = vi.fn();
    render(<ConfigurePhase repo={repo} onCreated={onCreated} onBack={() => {}} />);

    expect((screen.getByLabelText(/name/i) as HTMLInputElement).value).toBe("webapp");
    expect((screen.getByLabelText(/stackfile path/i) as HTMLInputElement).value).toBe("stackfile.yaml");

    await userEvent.click(screen.getByRole("button", { name: /enable previews/i }));

    await waitFor(() => {
      expect(createPreviewConfig).toHaveBeenCalledWith("org1", "default", {
        name: "webapp",
        git_repository: { repo_url: "https://github.com/acme/webapp.git", base_branch: "main" },
        stackfile_path: "stackfile.yaml",
        max_active_previews: 10,
      });
      expect(onCreated).toHaveBeenCalledWith("c1");
    });
  });

  it("falls back to the default stackfile path when the field is cleared", async () => {
    (createPreviewConfig as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "c1" });
    render(<ConfigurePhase repo={repo} onCreated={vi.fn()} onBack={() => {}} />);

    await userEvent.clear(screen.getByLabelText(/stackfile path/i));
    await userEvent.click(screen.getByRole("button", { name: /enable previews/i }));

    await waitFor(() => {
      expect(createPreviewConfig).toHaveBeenCalledWith(
        "org1",
        "default",
        expect.objectContaining({ stackfile_path: "stackfile.yaml" }),
      );
    });
  });

  it("marks name and base branch as required and blocks submit with an empty name", async () => {
    (createPreviewConfig as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "c1" });
    render(<ConfigurePhase repo={repo} onCreated={vi.fn()} onBack={() => {}} />);

    const nameInput = screen.getByLabelText(/^name/i);
    await userEvent.clear(nameInput);
    await userEvent.click(screen.getByRole("button", { name: /enable previews/i }));

    expect(await screen.findByText(/name is required/i)).toBeInTheDocument();
    expect(createPreviewConfig).not.toHaveBeenCalled();

    const nameLabel = screen.getByText(/^name$/i).closest("label");
    expect(nameLabel?.querySelector('[aria-hidden]')).toHaveTextContent("*");
  });

  it("clears the name error once the field is edited", async () => {
    render(<ConfigurePhase repo={repo} onCreated={vi.fn()} onBack={() => {}} />);
    const nameInput = screen.getByLabelText(/^name/i);
    await userEvent.clear(nameInput);
    await userEvent.click(screen.getByRole("button", { name: /enable previews/i }));
    expect(await screen.findByText(/name is required/i)).toBeInTheDocument();

    await userEvent.type(nameInput, "webapp");
    expect(screen.queryByText(/name is required/i)).not.toBeInTheDocument();
  });

  it("shows an inline error on 409", async () => {
    const err = new AxiosError("conflict");
    Object.defineProperty(err, "response", { value: { status: 409, data: { reason: "exists" } } });
    (createPreviewConfig as ReturnType<typeof vi.fn>).mockRejectedValue(err);
    render(<ConfigurePhase repo={repo} onCreated={() => {}} onBack={() => {}} />);
    await userEvent.click(screen.getByRole("button", { name: /enable previews/i }));
    await waitFor(() => {
      expect(screen.getByText(/already exists/i)).toBeTruthy();
    });
  });
});
