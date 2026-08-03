// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { ConfirmProvider, useConfirm, type ConfirmOptions } from "../confirm";

afterEach(() => {
  cleanup();
  // A leaked lock would silently fail the next test rather than this one.
  document.body.style.pointerEvents = "";
});

function Harness({ opts, onResult }: { opts: ConfirmOptions; onResult: (ok: boolean) => void }) {
  const confirm = useConfirm();
  return (
    <button type="button" onClick={() => void confirm(opts).then(onResult)}>
      ask
    </button>
  );
}

function renderHarness(opts: ConfirmOptions) {
  const results: boolean[] = [];
  render(
    <ConfirmProvider>
      <Harness opts={opts} onResult={(ok) => results.push(ok)} />
    </ConfirmProvider>,
  );
  return results;
}

describe("ConfirmProvider", () => {
  it("resolves true when the action is confirmed", async () => {
    const results = renderHarness({
      title: "Delete thing?",
      description: "Gone forever.",
      confirmLabel: "Delete",
      variant: "destructive",
    });
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "ask" }));
    // Opening is deferred a tick — the dialog appears asynchronously.
    await user.click(await screen.findByRole("button", { name: "Delete" }));

    await waitFor(() => expect(results).toEqual([true]));
  });

  it("resolves false on cancel and leaves body pointer-events unlocked", async () => {
    const results = renderHarness({ title: "Sure?" });
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "ask" }));
    await user.click(await screen.findByRole("button", { name: "Cancel" }));

    await waitFor(() => expect(results).toEqual([false]));
    // Regression guard for the radix-ui/primitives#1836 wedge: once the
    // confirm settles, no stale pointer-events lock may remain on <body>.
    await waitFor(() => expect(document.body.style.pointerEvents).toBe(""));
  });

  it("keeps navigation and teardown off the close tick: resolution arrives after the dialog is gone", async () => {
    const results = renderHarness({ title: "Go?" });
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "ask" }));
    await user.click(await screen.findByRole("button", { name: "Confirm" }));

    await waitFor(() => expect(results).toEqual([true]));
    expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument();
    expect(document.body.style.pointerEvents).toBe("");
  });
});
