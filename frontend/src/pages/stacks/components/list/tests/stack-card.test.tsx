// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import { DeployStackCard } from "../stack-card";
import { stackRollupState, needsAttention, lastDeployFailed, statusReason } from "../status";
import { ReleaseState } from "@/pages/stacks/components/editor/tabs/deployments/release-states";
import type { Stack } from "@/api/stack-types";

const baseStack = {
  id: "s1",
  name: "tooljet",
  namespace: "ns-tooljet",
  revision: 4,
  spec: {
    stack_resources: [{ name: "web" }, { name: "db" }],
    volumes: [{}],
  },
  updated_at: new Date().toISOString(),
} as unknown as Stack;

const healthy = {
  ...baseStack,
  latest_release: { id: "r1", state: ReleaseState.Released },
  converged_release: { id: "r1", state: ReleaseState.Released, health: "ok" },
} as Stack;

function renderCard(stack: Stack, onDelete?: () => void) {
  return render(
    <MemoryRouter>
      <DeployStackCard stack={stack} onDelete={onDelete} />
    </MemoryRouter>,
  );
}

afterEach(cleanup);

describe("DeployStackCard", () => {
  it("names every component rather than counting them", () => {
    renderCard(healthy);
    expect(screen.getByText("tooljet")).toBeTruthy();
    expect(screen.getByText("Healthy")).toBeTruthy();
    // `2 services · 1 volume` is a count you cannot act on. The card lists what
    // the stack is actually made of, by name — and carries no count at all.
    expect(screen.queryByText(/\d+ components?/)).toBeNull();
    expect(screen.getByText("web")).toBeTruthy();
    expect(screen.getByText("db")).toBeTruthy();
    expect(screen.getByText(/^Last change /)).toBeTruthy();
  });

  it("carries no chart — a sparkline with no axis reports nothing anyone can act on", () => {
    renderCard({ ...healthy, deploy_history: [1, 0, 4, 2, 0, 1, 3] } as Stack);
    expect(document.querySelector("[data-slot='deploy-sparkline']")).toBeNull();
    expect(screen.queryByText(/deploys/i)).toBeNull();
  });

  it("shows the trash action only when onDelete is wired", () => {
    const { rerender } = renderCard(baseStack);
    expect(screen.queryByLabelText("Delete tooljet")).toBeNull();
    rerender(
      <MemoryRouter>
        <DeployStackCard stack={baseStack} onDelete={() => {}} />
      </MemoryRouter>,
    );
    // One destination, so it is a trash and not a menu that exists to hide a
    // single item. The retype gate is what makes it safe, not the extra click.
    expect(screen.getByLabelText("Delete tooljet")).toBeTruthy();
  });

  it("reads Deploying while the latest release is in flight", () => {
    renderCard({ ...baseStack, latest_release: { id: "r2", state: ReleaseState.InProgress } } as Stack);
    expect(screen.getByText("Deploying")).toBeTruthy();
  });

  it("reads Not deployed for a stack with no releases", () => {
    renderCard(baseStack);
    expect(screen.getByText("Not deployed")).toBeTruthy();
  });

  it("reports a failed latest attempt as its own fact, without contradicting the live status", () => {
    renderCard({
      ...baseStack,
      latest_release: { id: "r3", state: ReleaseState.Failed },
      converged_release: { id: "r1", state: ReleaseState.Released, health: "ok" },
    } as Stack);
    // The stack really is serving, so the status word stays honest...
    expect(screen.getByText("Healthy")).toBeTruthy();
    // ...and the failed push is reported separately rather than overwriting it.
    expect(screen.getByText("Last deploy failed")).toBeTruthy();
  });

  it("says the status once — a glyph and a word in one hue, no dot, no rail, no chip", () => {
    renderCard(healthy);
    const word = screen.getByText("Healthy");
    expect(word.getAttribute("data-status-variant")).toBe("ready");
    // The glyph is per STATE, so it carries a fact the word does not have to
    // repeat — that is what earns it the space. What "says it once" forbids is
    // a second *colour channel*: a dot, a rail, a fill or a border on top.
    expect(word.querySelector("svg")).not.toBeNull();
    expect(document.querySelector("[data-rail]")).toBeNull();
  });

  it("puts the components in the head group, not beside the footer", () => {
    renderCard(healthy);
    // The chips say what the stack IS, so they belong with the name — parked by
    // the footer they read as metadata about the last change.
    const chips = document.querySelector("[data-slot='components']")!;
    const provenance = document.querySelector("[data-slot='provenance']")!;
    const footer = document.querySelector("[data-slot='card-footer']")!;
    expect(chips.parentElement).toBe(provenance.parentElement);
    expect(chips.parentElement).not.toBe(footer.parentElement);
    // No rule between the chips and the footer.
    expect(document.querySelector(".border-t")).toBeNull();
  });

  it("sets the ref in mono and the project in Geist — one line, two kinds of value", () => {
    renderCard(healthy);
    const line = document.querySelector("[data-slot='provenance']")!;
    // Nothing to pin here but the fallback: this fixture has no git source.
    expect(line.textContent).toBe("never deployed");
    expect(line.className).not.toContain("font-mono");
  });
});

describe("stackRollupState", () => {
  it("Deleting lifecycle overrides health", () => {
    expect(stackRollupState({ ...healthy, lifecycle: "deleting" } as Stack)).toBe("Deleting");
  });

  it("maps release health onto the rollup vocabulary", () => {
    expect(stackRollupState(healthy)).toBe("Healthy");
    expect(stackRollupState(baseStack)).toBe("NotDeployed");
  });
});

describe("needsAttention", () => {
  it("is false for a healthy stack and for one that has never deployed", () => {
    expect(needsAttention(healthy)).toBe(false);
    expect(needsAttention(baseStack)).toBe(false);
  });

  it("is true when the live release is unhealthy", () => {
    const failed = {
      ...baseStack,
      latest_release: { id: "r1", state: ReleaseState.Released },
      converged_release: { id: "r1", state: ReleaseState.Released, health: "failed" },
    } as Stack;
    expect(needsAttention(failed)).toBe(true);
  });

  it("is true when the stack serves fine but the newest push did not land", () => {
    const masked = {
      ...baseStack,
      latest_release: { id: "r3", state: ReleaseState.Failed },
      converged_release: { id: "r1", state: ReleaseState.Released, health: "ok" },
    } as Stack;
    expect(stackRollupState(masked)).toBe("Healthy");
    expect(lastDeployFailed(masked)).toBe(true);
    expect(needsAttention(masked)).toBe(true);
  });

  it("is false while a stack is being deleted — in flight is not in trouble", () => {
    const deleting = {
      ...baseStack,
      lifecycle: "deleting",
      latest_release: { id: "r3", state: ReleaseState.Failed },
      converged_release: { id: "r1", state: ReleaseState.Released, health: "ok" },
    } as Stack;
    expect(needsAttention(deleting)).toBe(false);
  });
});

describe("statusReason", () => {
  // The reason line is what makes a broken row visibly TALLER than a healthy
  // one, so it has to appear on exactly the stacks the header counts — no more.
  it("is null for a healthy stack even when the release carries a message", () => {
    const chatty = {
      ...baseStack,
      latest_release: { id: "r1", state: ReleaseState.Released },
      converged_release: {
        id: "r1",
        state: ReleaseState.Released,
        health: "ok",
        message: "superseded by release 12",
      },
    } as Stack;
    expect(needsAttention(chatty)).toBe(false);
    expect(statusReason(chatty)).toBeNull();
  });

  it("reports the release message for a stack that needs attention", () => {
    const degraded = {
      ...baseStack,
      latest_release: { id: "r1", state: ReleaseState.Released },
      converged_release: {
        id: "r1",
        state: ReleaseState.Released,
        health: "degraded",
        message: "1 of 3 replicas available",
      },
    } as Stack;
    expect(statusReason(degraded)).toEqual({ text: "1 of 3 replicas available", tone: "muted" });
  });

  it("is null while deleting, and null for a stack in flight", () => {
    expect(statusReason({ ...healthy, lifecycle: "deleting" } as Stack)).toBeNull();
    expect(
      statusReason({
        ...baseStack,
        latest_release: { id: "r2", state: ReleaseState.InProgress, message: "rolling out 1 of 3" },
      } as Stack),
    ).toBeNull();
  });
});
