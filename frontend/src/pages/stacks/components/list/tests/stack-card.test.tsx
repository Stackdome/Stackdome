// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import { DeployStackCard, EndpointPills, headerStatus, relativeAge } from "../stack-card";
import { ReleaseState } from "@/pages/stacks/components/editor/tabs/deployments/release-states";
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
  it("renders a healthy card driven by the release rollup, with meta grid, footer status and no rail", () => {
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
    expect(screen.getByText("resources").nextElementSibling?.textContent).toBe("2");
    expect(screen.getByText("volumes").nextElementSibling?.textContent).toBe("1");
    expect(document.querySelector("[data-rail]")).toBeNull();
  });

  it("shows the kebab Delete action only when onDelete is wired", () => {
    const { rerender } = render(
      <MemoryRouter>
        <DeployStackCard stack={baseStack} />
      </MemoryRouter>,
    );
    expect(screen.queryByLabelText("Actions for tooljet")).toBeNull();
    rerender(
      <MemoryRouter>
        <DeployStackCard stack={baseStack} onDelete={() => {}} />
      </MemoryRouter>,
    );
    expect(screen.getByLabelText("Actions for tooljet")).toBeTruthy();
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

describe("relativeAge", () => {
  it("renders compact tokens", () => {
    const now = Date.now();
    expect(relativeAge(new Date(now - 10_000).toISOString())).toBe("just now");
    expect(relativeAge(new Date(now - 5 * 60_000).toISOString())).toBe("5m ago");
    expect(relativeAge(new Date(now - 5 * 3_600_000).toISOString())).toBe("5h ago");
    expect(relativeAge(new Date(now - 3 * 86_400_000).toISOString())).toBe("3d ago");
    expect(relativeAge(null)).toBeNull();
  });
});

describe("EndpointPills", () => {
  it("shows two pills and collapses the rest into a +N popover trigger", () => {
    const urls = [
      { resource: "web", url: "https://web.example.com" },
      { resource: "api", url: "https://api.example.com" },
      { resource: "docs", url: "https://docs.example.com" },
      { resource: "admin", url: "https://admin.example.com" },
    ];
    render(<EndpointPills urls={urls} />);
    expect(screen.getByRole("link", { name: /web/ })).toBeTruthy();
    expect(screen.getByRole("link", { name: /api/ })).toBeTruthy();
    expect(screen.queryByRole("link", { name: /docs/ })).toBeNull();
    expect(screen.getByRole("button", { name: "2 more endpoints" })).toBeTruthy();
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
