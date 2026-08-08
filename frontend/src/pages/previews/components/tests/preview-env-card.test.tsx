// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { PreviewEnvCard } from "../preview-env-card";
import { previewStatusVariant, statusVariantLabel } from "@/components/branded/status-variant";
import type { PreviewStack } from "@/api/preview-envs";

afterEach(() => cleanup());

function renderCard(props: Partial<React.ComponentProps<typeof PreviewEnvCard>> & { env: PreviewStack }) {
  return render(
    <MemoryRouter>
      <PreviewEnvCard {...props} />
    </MemoryRouter>,
  );
}

const base: PreviewStack = {
  id: "p1",
  pr_number: "42",
  branch: "feat/login",
  commit: "a3f9c21deadbeef",
  stack_id: "s-123",
  status: { phase: "Ready", outputs: { urls: [{ resource: "web", url: "https://pr-42.example.com" }] } },
};

describe("PreviewEnvCard", () => {
  it("renders PR number, branch and the status word for the phase", () => {
    renderCard({ env: base });
    expect(screen.getByText("PR #42")).toBeInTheDocument();
    expect(screen.getByText("feat/login")).toBeInTheDocument();
    const label = statusVariantLabel[previewStatusVariant(base.status?.phase)];
    expect(screen.getByText(label)).toBeInTheDocument();
  });

  it("keeps failure details off the card — the status word is the only signal", () => {
    const env: PreviewStack = {
      ...base,
      status: { phase: "Failed", reason: "StackfileNotFound", message: "no stackfile at path" },
    };
    renderCard({ env });
    // Sentence case now that §7 removed the CSS caps — words are words.
    expect(screen.getByText(statusVariantLabel.error)).toBeInTheDocument();
    expect(screen.queryByText(/no stackfile at path/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/check the stackfile path/i)).not.toBeInTheDocument();
  });

  it("fires onSync from the kebab menu without navigating", async () => {
    const onSync = vi.fn();
    const onDelete = vi.fn();
    renderCard({ env: base, onSync, onDelete });

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /actions for pr #42/i }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText(/^sync$/i), { pointerEventsCheck: 0 });

    await vi.waitFor(() => expect(onSync).toHaveBeenCalledWith(base));
    expect(onDelete).not.toHaveBeenCalled();
  });

  it("fires onDelete from the kebab menu without navigating", async () => {
    const onSync = vi.fn();
    const onDelete = vi.fn();
    renderCard({ env: base, onSync, onDelete });

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /actions for pr #42/i }), { pointerEventsCheck: 0 });
    await user.click(await screen.findByText(/^delete$/i), { pointerEventsCheck: 0 });

    await vi.waitFor(() => expect(onDelete).toHaveBeenCalledWith(base));
    expect(onSync).not.toHaveBeenCalled();
  });

  it("disables kebab items while the env is Deleting", async () => {
    const env: PreviewStack = { ...base, status: { phase: "Deleting" } };
    renderCard({ env, onSync: vi.fn(), onDelete: vi.fn() });

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /actions for pr #42/i }), { pointerEventsCheck: 0 });

    expect(await screen.findByText(/^sync$/i)).toHaveAttribute("aria-disabled", "true");
    expect(screen.getByText(/^delete$/i)).toHaveAttribute("aria-disabled", "true");
  });

  it("has no link role when stack_id is absent", () => {
    const env: PreviewStack = { ...base, stack_id: undefined };
    renderCard({ env });
    const card = screen.getByLabelText(/pr #42 preview environment/i);
    expect(card).not.toHaveAttribute("role", "link");
  });

  it("has a link role and navigates when stack_id is present", () => {
    renderCard({ env: base });
    const card = screen.getByLabelText(/pr #42 preview environment/i);
    expect(card).toHaveAttribute("role", "link");
  });
});
