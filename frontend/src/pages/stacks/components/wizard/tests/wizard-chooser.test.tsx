// @vitest-environment jsdom
// frontend/src/pages/stacks/components/wizard/tests/wizard-chooser.test.tsx
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { WizardChooser } from "../wizard-chooser";

afterEach(cleanup);

describe("WizardChooser", () => {
  it("routes each start option to its handler", async () => {
    const user = userEvent.setup();
    const onPickBlocks = vi.fn(),
      onPickTemplate = vi.fn(),
      onPickCompose = vi.fn(),
      onPickBlank = vi.fn();
    render(
      <WizardChooser
        onPickBlocks={onPickBlocks}
        onPickTemplate={onPickTemplate}
        onPickCompose={onPickCompose}
        onPickBlank={onPickBlank}
      />,
    );

    await user.click(screen.getByRole("button", { name: /Compose blocks/i }));
    await user.click(screen.getByRole("button", { name: /From template/i }));
    await user.click(screen.getByRole("button", { name: /Docker compose/i }));
    await user.click(screen.getByRole("button", { name: /blank slate/i }));

    expect(onPickBlocks).toHaveBeenCalledOnce();
    expect(onPickTemplate).toHaveBeenCalledOnce();
    expect(onPickCompose).toHaveBeenCalledOnce();
    expect(onPickBlank).toHaveBeenCalledOnce();
  });

  it("renders the GitHub tile disabled with a 'soon' marker", () => {
    render(
      <WizardChooser
        onPickBlocks={vi.fn()}
        onPickTemplate={vi.fn()}
        onPickCompose={vi.fn()}
        onPickBlank={vi.fn()}
      />,
    );
    const github = screen.getByRole("button", { name: /GitHub repo/i });
    expect(github).toBeDisabled();
    expect(screen.getByText(/soon/i)).toBeInTheDocument();
  });
});
