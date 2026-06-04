// @vitest-environment jsdom
import { describe, it, expect, afterEach, beforeAll, vi } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, within, fireEvent } from "@testing-library/react";
import { Accordion } from "@/components/ui/accordion";
import StackResourceDetail from "../stack-resource-detail";
import type { FormStackResourceData, FormEnvVarData } from "@/pages/stacks/schemas/form-schema";

afterEach(cleanup);

// Radix Tabs/Tooltip touch pointer-capture + scroll APIs that jsdom lacks.
beforeAll(() => {
  const stubs: Record<string, () => unknown> = {
    hasPointerCapture: () => false,
    releasePointerCapture: () => undefined,
    setPointerCapture: () => undefined,
    scrollIntoView: () => undefined,
  };
  for (const [name, impl] of Object.entries(stubs)) {
    const proto = Element.prototype as unknown as Record<string, unknown>;
    if (!proto[name]) proto[name] = vi.fn(impl);
  }
});

const envVars: FormEnvVarData[] = [
  { from: "stack", name: "PLAIN", value: "plain-value" },
  { from: "secret", name: "SECRET_VAR", secretId: "s1", secretKey: "API_KEY" },
  { from: "resource", name: "RES_VAR", resourceName: "web", output: "url" },
  { from: "self", name: "SELF_VAR", selfOutput: "endpoint" },
  { from: "addon", name: "PG_HOST", addonId: "a1", database: "tooljet", superuser: false, credField: "host" },
  { from: "addon", name: "PG_PORT", addonId: "a1", database: "tooljet", superuser: false, credField: "port" },
];

const resource: Partial<FormStackResourceData> = {
  name: "web",
  sourceType: "image",
  image_spec: { image: "nginx" },
  execution_config: { environment_variables: envVars },
} as unknown as Partial<FormStackResourceData>;

const renderDetail = () => {
  const utils = render(
    <Accordion type="single" collapsible defaultValue="0">
      <StackResourceDetail
        resource={resource}
        index={0}
        addonNameById={new Map([["a1", "tooljet-db"]])}
      />
    </Accordion>,
  );
  // Read-mode view defaults to the Configuration tab; activate Environment.
  fireEvent.mouseDown(screen.getByRole("tab", { name: /environment/i }));
  fireEvent.click(screen.getByRole("tab", { name: /environment/i }));
  return utils;
};

describe("StackResourceDetail read-mode environment tab", () => {
  it("shows the plain stack row value", () => {
    renderDetail();
    expect(screen.getByText("PLAIN")).toBeInTheDocument();
    expect(screen.getByText("plain-value")).toBeInTheDocument();
  });

  it("shows secret/resource/self source bindings instead of blank", () => {
    renderDetail();
    // secret → key shown
    expect(screen.getByText("API_KEY")).toBeInTheDocument();
    // resource → resourceName · output ("web" also appears as the resource
    // title, so assert on the unique output token).
    expect(screen.getByText("url")).toBeInTheDocument();
    // self → selfOutput
    expect(screen.getByText("endpoint")).toBeInTheDocument();
  });

  it("groups the two addon rows inside one dashed group titled by addon name", () => {
    renderDetail();
    const groups = screen.getAllByTestId("env-addon-group");
    expect(groups).toHaveLength(1);
    const group = groups[0];
    expect(group).toHaveClass("border-dashed");
    expect(group).toHaveTextContent("tooljet-db");
    expect(group).toHaveTextContent("db: tooljet");
    // both addon rows + their credFields live inside the group
    expect(within(group).getByText("PG_HOST")).toBeInTheDocument();
    expect(within(group).getByText("PG_PORT")).toBeInTheDocument();
    expect(within(group).getByText("host")).toBeInTheDocument();
    expect(within(group).getByText("port")).toBeInTheDocument();
  });

  it("keeps the plain row outside the addon group", () => {
    renderDetail();
    const group = screen.getByTestId("env-addon-group");
    expect(group).not.toContainElement(screen.getByText("PLAIN"));
  });
});
