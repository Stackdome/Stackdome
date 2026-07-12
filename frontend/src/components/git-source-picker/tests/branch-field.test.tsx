// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeAll, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("@/api/git-integrations", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  listRepositoryBranches: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));

import { listRepositoryBranches } from "@/api/git-integrations";
import { BranchField } from "../branch-field";

// Radix Select uses pointer-capture APIs that jsdom doesn't implement.
// Stub them so userEvent.click on a SelectTrigger opens the popover.
beforeAll(() => {
  const stubs: Record<string, () => unknown> = {
    hasPointerCapture: () => false,
    releasePointerCapture: () => undefined,
    setPointerCapture: () => undefined,
    scrollIntoView: () => undefined,
  };
  for (const [name, impl] of Object.entries(stubs)) {
    const proto = Element.prototype as unknown as Record<string, unknown>;
    if (!proto[name]) proto[name] = vi.fn(impl);
  }
});

beforeEach(() => vi.clearAllMocks());
afterEach(() => cleanup());

describe("BranchField", () => {
  it("renders a Select of listed branches and fires onChange on select", async () => {
    const user = userEvent.setup();
    vi.mocked(listRepositoryBranches).mockResolvedValue({ items: ["main", "develop"] });
    const onChange = vi.fn();
    render(
      <BranchField
        id="branch"
        value="main"
        onChange={onChange}
        integrationId="int-1"
        repoFullName="acme/webapp"
      />,
    );

    const trigger = await screen.findByRole("combobox");
    await user.click(trigger);
    await user.click(await screen.findByRole("option", { name: "develop" }));
    expect(onChange).toHaveBeenCalledWith("develop");
  });

  it("renders a free-text Input when integrationId is null", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <BranchField
        id="branch"
        value=""
        onChange={onChange}
        integrationId={null}
        repoFullName="acme/webapp"
      />,
    );

    expect(listRepositoryBranches).not.toHaveBeenCalled();
    const input = screen.getByPlaceholderText("main");
    await user.type(input, "x");
    expect(onChange).toHaveBeenCalledWith("x");
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
  });

  it("falls back to free text when listing branches rejects", async () => {
    vi.mocked(listRepositoryBranches).mockRejectedValue(new Error("nope"));
    render(
      <BranchField
        id="branch"
        value="main"
        onChange={vi.fn()}
        integrationId="int-1"
        repoFullName="acme/webapp"
      />,
    );

    await waitFor(() => expect(listRepositoryBranches).toHaveBeenCalled());
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
    expect(screen.getByDisplayValue("main")).toBeInTheDocument();
  });
});
