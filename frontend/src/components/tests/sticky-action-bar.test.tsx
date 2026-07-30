// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import StickyActionBar from "../sticky-action-bar";
import { ConfirmProvider } from "@/components/branded/confirm";

const renderBar: typeof render = ((ui: Parameters<typeof render>[0]) =>
  render(ui, { wrapper: ConfirmProvider })) as typeof render;

afterEach(cleanup);

// The component portals into #page-sticky-bar (provided by the app layout).
// Tests must mount that slot before render — without it, nothing paints.
beforeEach(() => {
  const slot = document.createElement("div");
  slot.id = "page-sticky-bar";
  document.body.appendChild(slot);
});

const noopPrimary = (over: Partial<Parameters<typeof StickyActionBar>[0]["primary"]> = {}) => ({
  label: "Deploy",
  onClick: vi.fn(),
  ...over,
});

describe("StickyActionBar — segment rendering", () => {
  it("renders the lead label", () => {
    renderBar(<StickyActionBar leadLabel="Draft" segments={[]} primary={noopPrimary()} />);
    expect(screen.getByText("Draft")).toBeInTheDocument();
  });

  it("renders each segment's number and label", () => {
    renderBar(
      <StickyActionBar
        leadLabel="Draft"
        segments={[
          { num: 2, label: "RESOURCES" },
          { num: 1, label: "ADDON" },
        ]}
        primary={noopPrimary()}
      />,
    );
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("RESOURCES")).toBeInTheDocument();
    expect(screen.getByText("1")).toBeInTheDocument();
    expect(screen.getByText("ADDON")).toBeInTheDocument();
  });
});

describe("StickyActionBar — primary button", () => {
  it("fires onClick when clicked", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    renderBar(
      <StickyActionBar leadLabel="Draft" segments={[]} primary={noopPrimary({ onClick })} />,
    );
    await user.click(screen.getByRole("button", { name: /deploy/i }));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it("disables both buttons and shows loadingLabel while loading", () => {
    renderBar(
      <StickyActionBar
        leadLabel="Draft"
        segments={[]}
        primary={noopPrimary({ isLoading: true, loadingLabel: "Deploying" })}
        secondary={{ label: "Cancel", onClick: vi.fn() }}
      />,
    );
    expect(screen.getByRole("button", { name: /deploying/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /cancel/i })).toBeDisabled();
  });
});

describe("StickyActionBar — secondary confirm gating", () => {
  // Regression-worthy: "Discard all" must NOT silently drop edits when the
  // dirty count is at or above the threshold. The dialog is the safety net.

  it("calls secondary.onClick directly when no `confirm` is configured", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    renderBar(
      <StickyActionBar
        leadLabel="Draft"
        segments={[]}
        primary={noopPrimary()}
        secondary={{ label: "Cancel", onClick }}
      />,
    );
    await user.click(screen.getByRole("button", { name: /cancel/i }));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it("calls secondary.onClick directly when dirtyCount is below threshold", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    renderBar(
      <StickyActionBar
        leadLabel="Draft"
        segments={[]}
        primary={noopPrimary()}
        secondary={{
          label: "Discard all",
          onClick,
          dirtyCount: 1,
          confirm: {
            threshold: 2,
            title: "Discard all changes?",
            description: "Body",
          },
        }}
      />,
    );
    await user.click(screen.getByRole("button", { name: /discard all/i }));
    expect(onClick).toHaveBeenCalledOnce();
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
  });

  it("opens the confirm dialog instead of firing onClick when dirtyCount meets threshold", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    renderBar(
      <StickyActionBar
        leadLabel="Draft"
        segments={[]}
        primary={noopPrimary()}
        secondary={{
          label: "Discard all",
          onClick,
          dirtyCount: 3,
          confirm: {
            threshold: 2,
            title: "Discard all changes?",
            description: "Body",
          },
        }}
      />,
    );
    await user.click(screen.getByRole("button", { name: /discard all/i }));
    expect(onClick).not.toHaveBeenCalled();
    expect(await screen.findByRole("alertdialog")).toBeInTheDocument();
    expect(screen.getByText("Discard all changes?")).toBeInTheDocument();
  });

  it("fires onClick after the dialog's confirm action", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    renderBar(
      <StickyActionBar
        leadLabel="Draft"
        segments={[]}
        primary={noopPrimary()}
        secondary={{
          label: "Discard all",
          onClick,
          dirtyCount: 5,
          confirm: {
            threshold: 2,
            title: "Discard all changes?",
            description: "Body",
            confirmLabel: "Yes, discard",
          },
        }}
      />,
    );
    await user.click(screen.getByRole("button", { name: /^discard all$/i }));
    await user.click(await screen.findByRole("button", { name: /yes, discard/i }));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it("does NOT fire onClick when the dialog is cancelled", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    renderBar(
      <StickyActionBar
        leadLabel="Draft"
        segments={[]}
        primary={noopPrimary()}
        secondary={{
          label: "Discard all",
          onClick,
          dirtyCount: 5,
          confirm: {
            threshold: 2,
            title: "Discard all changes?",
            description: "Body",
            cancelLabel: "Keep editing",
          },
        }}
      />,
    );
    await user.click(screen.getByRole("button", { name: /^discard all$/i }));
    await user.click(await screen.findByRole("button", { name: /keep editing/i }));
    expect(onClick).not.toHaveBeenCalled();
  });
});

describe("StickyActionBar — tone / optional primary", () => {
  it("clean tone renders the lead label with no action buttons", () => {
    renderBar(<StickyActionBar leadLabel="All deployed" segments={[]} tone="clean" />);
    expect(screen.getByText("All deployed")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("deploying tone shows only the secondary (Cancel) — no primary", () => {
    renderBar(
      <StickyActionBar
        leadLabel="Deploying"
        segments={[]}
        tone="deploying"
        secondary={{ label: "Cancel", onClick: vi.fn() }}
      />,
    );
    expect(screen.getByRole("button", { name: /cancel/i })).toBeInTheDocument();
    expect(screen.getAllByRole("button")).toHaveLength(1);
  });
});
