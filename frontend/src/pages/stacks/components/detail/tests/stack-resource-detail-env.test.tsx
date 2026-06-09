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
  // A plain row positioned AFTER the addon rows: it must still render above the
  // dashed addon group, since all ungrouped rows render first.
  { from: "resource", name: "SMTP_DOMAIN", resourceName: "mail", output: "domain" },
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
  it("shows the plain stack row value with a Plain text From pill", () => {
    renderDetail();
    expect(screen.getByText("PLAIN")).toBeInTheDocument();
    expect(screen.getByText("plain-value")).toBeInTheDocument();
    // stack source renders as a "Plain text" From pill.
    expect(screen.getByText("Plain text")).toBeInTheDocument();
  });

  it("shows secret/resource/self source bindings with masked values + From pills", () => {
    renderDetail();
    // secret → masked value carries its key, plus a "Secret" From pill.
    expect(screen.getByText(/master_key|API_KEY/)).toBeInTheDocument();
    expect(screen.getByText("Secret")).toBeInTheDocument();
    // resource → resourceName · output ("web" also appears as the resource
    // title, so assert on the unique output token) + "Resource" pill.
    expect(screen.getByText(/web · url/)).toBeInTheDocument();
    expect(screen.getAllByText("Resource").length).toBeGreaterThan(0);
    // self → selfOutput + "Self" pill.
    expect(screen.getByText("endpoint")).toBeInTheDocument();
    expect(screen.getByText("Self")).toBeInTheDocument();
  });

  it("groups the two addon rows inside one dashed group with a notched legend", () => {
    renderDetail();
    const groups = screen.getAllByTestId("env-addon-group");
    expect(groups).toHaveLength(1);
    const group = groups[0];
    expect(group).toHaveClass("border-dashed");
    // notched legend header shows the addon name + Postgres + db.
    expect(group).toHaveTextContent("tooljet-db");
    expect(group).toHaveTextContent("Postgres");
    expect(group).toHaveTextContent("db: tooljet");
    // both addon rows live inside the group; value reads "<addon-name> · <field>"
    // (no masking dots) and the From pill reads just "Addon".
    expect(within(group).getByText("PG_HOST")).toBeInTheDocument();
    expect(within(group).getByText("PG_PORT")).toBeInTheDocument();
    expect(within(group).getByText("tooljet-db · host")).toBeInTheDocument();
    expect(within(group).getByText("tooljet-db · port")).toBeInTheDocument();
    expect(within(group).getAllByText("Addon").length).toBeGreaterThanOrEqual(2);
  });

  it("keeps the plain row outside the addon group", () => {
    renderDetail();
    const group = screen.getByTestId("env-addon-group");
    expect(group).not.toContainElement(screen.getByText("PLAIN"));
  });

  it("renders ungrouped rows (incl. one positioned after the addon rows) ABOVE the addon group", () => {
    renderDetail();
    const group = screen.getByTestId("env-addon-group");
    // SMTP_DOMAIN sits after the addon rows in the array but is ungrouped, so it
    // must render above the dashed addon group in document order.
    const smtp = screen.getByText("SMTP_DOMAIN");
    expect(group).not.toContainElement(smtp);
    expect(
      smtp.compareDocumentPosition(group) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
    // The plain bordered container also precedes the dashed group.
    const plain = screen.getByText("PLAIN");
    expect(
      plain.compareDocumentPosition(group) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });
});
