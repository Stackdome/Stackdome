// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("@/api/git-integrations", () => ({
  searchRepositories: vi.fn(),
  getRepository: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: () => "org1",
}));

import { searchRepositories, getRepository } from "@/api/git-integrations";
import { RepoPickerPhase } from "../repo-picker-phase";

beforeEach(() => {
  vi.clearAllMocks();
  (searchRepositories as ReturnType<typeof vi.fn>).mockResolvedValue({
    items: [
      { full_name: "acme/webapp", clone_url: "https://github.com/acme/webapp.git", private: true, default_branch: "main" },
    ],
    page: 1,
    total_count: 1,
    has_next: false,
  });
  (getRepository as ReturnType<typeof vi.fn>).mockResolvedValue({
    full_name: "acme/webapp",
    clone_url: "https://github.com/acme/webapp.git",
    default_branch: "main",
  });
});

afterEach(() => cleanup());

describe("RepoPickerPhase", () => {
  it("lists repositories and picks one with its default branch", async () => {
    const onPicked = vi.fn();
    render(<RepoPickerPhase integrationId="gi1" onPicked={onPicked} onBack={() => {}} />);
    await waitFor(() => screen.getByText("acme/webapp"));
    await userEvent.click(screen.getByText("acme/webapp"));
    await waitFor(() => {
      expect(onPicked).toHaveBeenCalledWith({
        fullName: "acme/webapp",
        cloneUrl: "https://github.com/acme/webapp.git",
        defaultBranch: "main",
        integrationId: "gi1",
      });
    });
  });

  it("falls back to manual URL entry", async () => {
    const onPicked = vi.fn();
    render(<RepoPickerPhase integrationId="gi1" onPicked={onPicked} onBack={() => {}} />);
    await userEvent.click(await screen.findByRole("button", { name: /enter repository url/i }));
    const input = screen.getByPlaceholderText(/https:\/\/github.com/i);
    await userEvent.type(input, "https://github.com/acme/api");
    await userEvent.click(screen.getByRole("button", { name: /continue/i }));
    expect(onPicked).toHaveBeenCalledWith({
      fullName: "acme/api",
      cloneUrl: "https://github.com/acme/api",
      defaultBranch: "",
      integrationId: null,
    });
  });

  it("shows only manual mode without an integration", () => {
    render(<RepoPickerPhase integrationId={null} onPicked={() => {}} onBack={() => {}} />);
    expect(screen.getByPlaceholderText(/https:\/\/github.com/i)).toBeTruthy();
    expect(searchRepositories).not.toHaveBeenCalled();
  });
});
