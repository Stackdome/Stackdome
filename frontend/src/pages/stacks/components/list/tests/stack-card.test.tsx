// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import { DeployStackCard, headerStatus } from "../stack-card";
import { ReleaseState } from "@/pages/stacks/components/detail/deployments/release-states";
import type { Stack } from "@/pages/stacks/types";

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

afterEach(cleanup);

describe("DeployStackCard", () => {
  it("renders a healthy card driven by the release rollup, with footer meta and no rail", () => {
    render(
      <MemoryRouter>
        <DeployStackCard
          stack={{
            ...baseStack,
            latest_release: { id: "r1", state: ReleaseState.Released },
            converged_release: { id: "r1", state: ReleaseState.Released, health: "ok" },
          } as Stack}
        />
      </MemoryRouter>,
    );
    expect(screen.getByText("tooljet")).toBeTruthy();
    expect(screen.getByText("ok")).toBeTruthy();
    expect(screen.getByText("2 res")).toBeTruthy();
    expect(screen.getByText("1 vol")).toBeTruthy();
    expect(document.querySelector("[data-rail]")).toBeNull();
  });

  it("renders animated rail and progressing word while the latest release is in flight", () => {
    render(
      <MemoryRouter>
        <DeployStackCard
          stack={{
            ...baseStack,
            latest_release: { id: "r2", state: ReleaseState.InProgress },
          } as Stack}
        />
      </MemoryRouter>,
    );
    expect(screen.getByText("progressing")).toBeTruthy();
    const rail = document.querySelector('[data-rail="deploying"]');
    expect(rail).toBeTruthy();
    expect(rail?.querySelector(".animate-rail-sweep")).toBeTruthy();
  });

  it("reads Not deployed for a stack with no releases", () => {
    render(
      <MemoryRouter>
        <DeployStackCard stack={baseStack} />
      </MemoryRouter>,
    );
    expect(screen.getByText("Not deployed")).toBeTruthy();
    expect(document.querySelector("[data-rail]")).toBeNull();
  });

  it("shows the deploy-failed hint when a live release masks a failed latest attempt", () => {
    render(
      <MemoryRouter>
        <DeployStackCard
          stack={{
            ...baseStack,
            latest_release: { id: "r3", state: ReleaseState.Failed },
            converged_release: { id: "r1", state: ReleaseState.Released, health: "ok" },
          } as Stack}
        />
      </MemoryRouter>,
    );
    expect(screen.getByText("ok")).toBeTruthy();
    expect(screen.getByLabelText("Latest deploy failed")).toBeTruthy();
  });
});

describe("headerStatus", () => {
  it("Deleting lifecycle overrides health", () => {
    const s = {
      ...baseStack,
      lifecycle: "deleting",
      converged_release: { id: "r1", state: ReleaseState.Released, health: "ok" },
      latest_release: { id: "r1", state: ReleaseState.Released },
    } as Stack;
    expect(headerStatus(s)).toEqual({ label: "Deleting", variant: "pending" });
  });
});
