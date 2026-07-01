// @vitest-environment jsdom
// frontend/src/pages/stacks/components/wizard/tests/block-composer.test.tsx
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeAll, afterEach } from "vitest";
import { render, screen, cleanup, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { BlockComposer } from "../block-composer";

const navigate = vi.fn();
vi.mock("react-router-dom", () => ({ useNavigate: () => navigate }));
beforeAll(() => { Element.prototype.scrollIntoView = vi.fn(); });
afterEach(() => { cleanup(); navigate.mockReset(); });

describe("BlockComposer", () => {
  it("adds blocks and navigates to /stacks/new with a seed", async () => {
    const user = userEvent.setup();
    render(<BlockComposer onBack={vi.fn()} onClose={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /Web service/i }));
    await user.click(screen.getByRole("button", { name: /Postgres/i }));

    // "your stack so far" shows both
    const panel = screen.getByTestId("stack-so-far");
    expect(within(panel).getByText("web")).toBeInTheDocument();
    expect(within(panel).getByText("postgres")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Continue/i }));

    expect(navigate).toHaveBeenCalledWith(
      "/stacks/new",
      expect.objectContaining({
        state: expect.objectContaining({
          seed: expect.objectContaining({
            resources: expect.arrayContaining([expect.objectContaining({ name: expect.any(String) })]),
            linkedAddonIds: expect.any(Array),
          }),
        }),
      }),
    );
  });

  it("disables Continue until at least one block is added", () => {
    render(<BlockComposer onBack={vi.fn()} onClose={vi.fn()} />);
    expect(screen.getByRole("button", { name: /Continue/i })).toBeDisabled();
  });

  it("offers a Create add-on link to the addons page when the workspace has none", () => {
    render(<BlockComposer onBack={vi.fn()} onClose={vi.fn()} />);
    const link = screen.getByRole("link", { name: /Create an add-on/i });
    expect(link).toHaveAttribute("href", "/addons");
    expect(link).toHaveAttribute("target", "_blank");
  });
});
