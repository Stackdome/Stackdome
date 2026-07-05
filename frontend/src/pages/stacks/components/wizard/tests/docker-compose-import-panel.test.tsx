// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DockerComposeImportPanel } from "../docker-compose-import-panel";

afterEach(cleanup);

describe("DockerComposeImportPanel", () => {
  it("imports pasted YAML", async () => {
    const user = userEvent.setup();
    const onImport = vi.fn().mockResolvedValue(undefined);
    render(
      <DockerComposeImportPanel
        onImport={onImport}
        isLoading={false}
        error={null}
        onClearError={vi.fn()}
        onBack={vi.fn()}
      />,
    );
    await user.type(
      screen.getByRole("textbox"),
      "services:\n  web:\n    image: nginx",
    );
    await user.click(screen.getByRole("button", { name: /Continue/i }));
    expect(onImport).toHaveBeenCalledWith(
      "services:\n  web:\n    image: nginx",
    );
  });

  it("Continue button is disabled when textarea is empty", () => {
    render(
      <DockerComposeImportPanel
        onImport={vi.fn()}
        isLoading={false}
        error={null}
        onClearError={vi.fn()}
        onBack={vi.fn()}
      />,
    );
    expect(screen.getByRole("button", { name: /Continue/i })).toBeDisabled();
  });

  it("calls onBack when Back is clicked", async () => {
    const user = userEvent.setup();
    const onBack = vi.fn();
    render(
      <DockerComposeImportPanel
        onImport={vi.fn()}
        isLoading={false}
        error={null}
        onClearError={vi.fn()}
        onBack={onBack}
      />,
    );
    await user.click(screen.getByRole("button", { name: /Back/i }));
    expect(onBack).toHaveBeenCalled();
  });

  it("shows error when error prop is set", () => {
    render(
      <DockerComposeImportPanel
        onImport={vi.fn()}
        isLoading={false}
        error="Invalid YAML"
        onClearError={vi.fn()}
        onBack={vi.fn()}
      />,
    );
    expect(screen.getByText("Invalid YAML")).toBeInTheDocument();
  });
});
