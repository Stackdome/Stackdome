// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { GitSourcePanel } from "../git-source-panel";
import type { PickedRepo } from "@/components/git-source-picker/types";

afterEach(cleanup);

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useNavigate: () => mockNavigate,
}));

// The picker's network concerns are covered by its own tests; here it is a
// controlled stub that lets the test choose a repo.
const picked: PickedRepo = {
  fullName: "acme/webapp",
  cloneUrl: "https://github.com/acme/webapp.git",
  defaultBranch: "develop",
  integrationId: "int-app",
};
vi.mock("@/components/git-source-picker/git-source-picker", () => ({
  GitSourcePicker: ({ onChange }: { onChange: (r: PickedRepo | null) => void }) => (
    <button type="button" onClick={() => onChange(picked)}>
      stub-pick-repo
    </button>
  ),
}));

function renderPanel() {
  const onClose = vi.fn();
  render(
    <MemoryRouter>
      <GitSourcePanel onBack={vi.fn()} onClose={onClose} />
    </MemoryRouter>,
  );
  return { onClose };
}

describe("GitSourcePanel", () => {
  it("gates Continue until a repo is picked, then prefills the service form", async () => {
    const user = userEvent.setup();
    renderPanel();
    expect(screen.getByRole("button", { name: /continue/i })).toBeDisabled();
    await user.click(screen.getByText("stub-pick-repo"));
    await user.click(screen.getByRole("button", { name: /continue/i }));
    expect(screen.getByLabelText(/service name/i)).toHaveValue("webapp");
    expect(screen.getByLabelText(/branch/i)).toHaveValue("develop");
    expect(screen.getByLabelText(/dockerfile path/i)).toHaveValue("Dockerfile");
  });

  it("requires a port, then seeds /stacks/new and closes", async () => {
    const user = userEvent.setup();
    const { onClose } = renderPanel();
    await user.click(screen.getByText("stub-pick-repo"));
    await user.click(screen.getByRole("button", { name: /continue/i }));

    const openInEditor = screen.getByRole("button", { name: /open in editor/i });
    expect(openInEditor).toBeDisabled();
    await user.type(screen.getByLabelText(/port/i), "3000");
    await user.click(openInEditor);

    expect(mockNavigate).toHaveBeenCalledWith(
      "/stacks/new",
      expect.objectContaining({
        state: expect.objectContaining({
          seed: expect.objectContaining({
            name: "webapp",
            resources: [
              expect.objectContaining({
                sourceType: "git",
                gitRevisionValue: "develop",
                ports: [expect.objectContaining({ number: 3000 })],
              }),
            ],
          }),
        }),
      }),
    );
    expect(onClose).toHaveBeenCalled();
  });
});
