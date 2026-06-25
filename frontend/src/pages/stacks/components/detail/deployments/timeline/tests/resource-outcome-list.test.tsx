// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { ResourceOutcomeList } from "../resource-outcome-list";
import type { ResourceRowVM } from "../resource-row";

afterEach(cleanup);

const rows: ResourceRowVM[] = [
  { name: "web", phase: "Ready", source: { kind: "image", label: "nginx:1.25" } },
  { name: "worker", phase: "Ready", source: { kind: "git", label: "github.com/acme/app" } },
];

describe("ResourceOutcomeList", () => {
  it("renders the header and a row per resource with its source", () => {
    render(<ResourceOutcomeList rows={rows} />);
    expect(screen.getByText("Resource outcome")).toBeInTheDocument();
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("nginx:1.25")).toBeInTheDocument();
    expect(screen.getByText("github.com/acme/app")).toBeInTheDocument();
  });

  it("renders nothing when there are no rows", () => {
    const { container } = render(<ResourceOutcomeList rows={[]} />);
    expect(container).toBeEmptyDOMElement();
  });
});
