// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import { DeployStackCard } from "../stack-card";
import { PreviewEnvCard } from "../preview-env-card";
import type { Stack } from "@/pages/stacks/types";
import type { PreviewStack } from "@/api/preview-envs";

const baseStack = {
  id: "s1",
  name: "tooljet",
  namespace: "ns-tooljet",
  revision: 4,
  spec: { stack_resources: [{}, {}], volumes: [{}] },
  updated_at: new Date().toISOString(),
} as unknown as Stack;

afterEach(cleanup);

describe("DeployStackCard", () => {
  it("renders ready card with success rail, namespace chip, and footer meta", () => {
    render(
      <MemoryRouter>
        <DeployStackCard stack={{ ...baseStack, status: { state: "Ready" } } as Stack} />
      </MemoryRouter>,
    );
    expect(screen.getByText("tooljet")).toBeTruthy();
    expect(screen.getByText("ready")).toBeTruthy();
    expect(screen.getByText("ns-tooljet")).toBeTruthy();
    expect(screen.getByText("rev 4")).toBeTruthy();
    expect(screen.getByText("2 res")).toBeTruthy();
    expect(screen.getByText("1 vol")).toBeTruthy();
    expect(document.querySelector('[data-rail="success"]')).toBeTruthy();
  });

  it("renders animated deploying rail for in-flight states", () => {
    render(
      <MemoryRouter>
        <DeployStackCard stack={{ ...baseStack, status: { state: "Deploying" } } as Stack} />
      </MemoryRouter>,
    );
    expect(screen.getByText("deploying")).toBeTruthy();
    const rail = document.querySelector('[data-rail="deploying"]');
    expect(rail).toBeTruthy();
    expect(rail?.querySelector(".animate-rail-sweep")).toBeTruthy();
  });
});

describe("PreviewEnvCard", () => {
  const env = {
    id: "e1",
    stack_id: "s2",
    pr_number: "128",
    branch: "feat/checkout-redesign",
    commit: "895026c1234",
    config_id: "c1",
    status: {
      phase: "Ready",
      outputs: { urls: [{ resource: "web", url: "http://pr-128.example.test" }] },
    },
    updated_at: new Date().toISOString(),
  } as unknown as PreviewStack;

  it("renders brand rail, repo/branch rows, endpoint pill, and commit footer", () => {
    render(
      <MemoryRouter>
        <PreviewEnvCard env={env} configName="stackdome-preview-demo" />
      </MemoryRouter>,
    );
    expect(screen.getByText("PR #128")).toBeTruthy();
    expect(screen.getByText("ready")).toBeTruthy();
    expect(screen.getByText("stackdome-preview-demo")).toBeTruthy();
    expect(screen.getByText("feat/checkout-redesign")).toBeTruthy();
    const pill = screen.getByRole("link", { name: /web/ });
    expect(pill.getAttribute("href")).toBe("http://pr-128.example.test");
    expect(screen.getByText("895026c")).toBeTruthy();
    expect(document.querySelector('[data-rail="brand"]')).toBeTruthy();
  });
});
