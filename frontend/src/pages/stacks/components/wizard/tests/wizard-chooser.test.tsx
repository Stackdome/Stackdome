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
      onPickBlank = vi.fn(),
      onPickGit = vi.fn();
    render(
      <WizardChooser
        onPickBlocks={onPickBlocks}
        onPickTemplate={onPickTemplate}
        onPickCompose={onPickCompose}
        onPickBlank={onPickBlank}
        onPickGit={onPickGit}
      />,
    );

    await user.click(screen.getByRole("button", { name: /Compose with blocks/i }));
    await user.click(screen.getByRole("button", { name: /Start from a template/i }));
    await user.click(screen.getByRole("button", { name: /Import docker-compose/i }));
    await user.click(screen.getByRole("button", { name: /start from scratch/i }));
    await user.click(screen.getByRole("button", { name: /Deploy from git/i }));

    expect(onPickBlocks).toHaveBeenCalledOnce();
    expect(onPickTemplate).toHaveBeenCalledOnce();
    expect(onPickCompose).toHaveBeenCalledOnce();
    expect(onPickBlank).toHaveBeenCalledOnce();
    expect(onPickGit).toHaveBeenCalledOnce();
  });
});
