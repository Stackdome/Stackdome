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

    await user.click(screen.getByRole("button", { name: /Build from blocks/i }));
    await user.click(screen.getByRole("button", { name: /From template/i }));
    await user.click(screen.getByRole("button", { name: /Docker compose/i }));
    await user.click(screen.getByRole("button", { name: /blank slate/i }));
    await user.click(screen.getByRole("button", { name: /From git provider/i }));

    expect(onPickBlocks).toHaveBeenCalledOnce();
    expect(onPickTemplate).toHaveBeenCalledOnce();
    expect(onPickCompose).toHaveBeenCalledOnce();
    expect(onPickBlank).toHaveBeenCalledOnce();
    expect(onPickGit).toHaveBeenCalledOnce();
  });
});
