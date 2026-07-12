// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import { DeployStackCard } from "../stack-card";
import type { Stack } from "@/pages/stacks/types";

const baseStack = {
  id: "s1",
  name: "tooljet",
  namespace: "ns-tooljet",
  revision: 4,
  spec: {
    stack_resources: [
      { name: "web", status: { public_ingress: [{ url: "http://web.example.test" }] } },
      { name: "db" },
    ],
    volumes: [{}],
  },
  updated_at: new Date().toISOString(),
} as unknown as Stack;

afterEach(cleanup);

describe("DeployStackCard", () => {
  it("renders ready card with footer meta and no rail", () => {
    render(
      <MemoryRouter>
        <DeployStackCard stack={{ ...baseStack, status: { state: "Ready" } } as Stack} />
      </MemoryRouter>,
    );
    expect(screen.getByText("tooljet")).toBeTruthy();
    expect(screen.getByText("ready")).toBeTruthy();
    expect(screen.getByText("2 res")).toBeTruthy();
    expect(screen.getByText("1 vol")).toBeTruthy();
    const pill = screen.getByRole("link", { name: /web ↗/ });
    expect(pill.getAttribute("href")).toBe("http://web.example.test");
    expect(document.querySelector("[data-rail]")).toBeNull();
  });

  it("renders animated rail and pending word for in-flight states", () => {
    render(
      <MemoryRouter>
        <DeployStackCard stack={{ ...baseStack, status: { state: "Progressing" } } as Stack} />
      </MemoryRouter>,
    );
    expect(screen.getByText("pending")).toBeTruthy();
    const rail = document.querySelector('[data-rail="deploying"]');
    expect(rail).toBeTruthy();
    expect(rail?.querySelector(".animate-rail-sweep")).toBeTruthy();
  });
});
