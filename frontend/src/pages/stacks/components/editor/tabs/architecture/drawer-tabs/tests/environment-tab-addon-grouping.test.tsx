// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, within } from "@testing-library/react";
import { Tabs } from "@/components/ui/tabs";
import { StackResourceEnvironmentTab } from "../environment-tab";
import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import type { PostgresAddon } from "@/api/addons";

afterEach(cleanup);

// Radix Select uses pointer-capture APIs that jsdom doesn't implement.
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

const addons = [
  {
    id: "a1",
    name: "tooljet-db",
    spec: { databases: [{ name: "tooljet" }, { name: "app" }] },
  },
] as unknown as PostgresAddon[];

const renderTab = (envVars: FormEnvVarData[]) =>
  render(
    <Tabs defaultValue="environment">
      <StackResourceEnvironmentTab
        index={0}
        envVars={envVars}
        baselineEnvVars={undefined}
        errors={{}}
        resourceOptions={[]}
        selfOutputs={[]}
        secrets={[]}
        secretsLoading={false}
        addons={addons}
        addonNameById={new Map([["a1", "tooljet-db"]])}
        onChangeEnvVars={vi.fn()}
      />
    </Tabs>,
  );

describe("StackResourceEnvironmentTab addon grouping", () => {
  const envVars: FormEnvVarData[] = [
    { from: "addon", name: "PG_HOST", addonId: "a1", database: "tooljet", superuser: false, credField: "host" },
    { from: "addon", name: "PG_PORT", addonId: "a1", database: "tooljet", superuser: false, credField: "port" },
    { from: "stack", name: "PLAIN", value: "v" },
  ];

  it("renders the two addon rows inside one group container", () => {
    renderTab(envVars);
    const groups = screen.getAllByTestId("env-addon-group");
    expect(groups).toHaveLength(1);
    const group = groups[0];
    expect(group).toHaveClass("border-border-strong");
    // both addon rows live inside the group
    expect(within(group).getByTestId("env-row-0-0")).toBeInTheDocument();
    expect(within(group).getByTestId("env-row-0-1")).toBeInTheDocument();
  });

  it("shows the group-level addon name and a database control", () => {
    renderTab(envVars);
    const group = screen.getByTestId("env-addon-group");
    expect(within(group).getByTestId("addon-picker-trigger")).toHaveTextContent("tooljet-db");
    // addon has >1 database => database picker is rendered
    expect(within(group).getByTestId("database-picker-trigger")).toBeInTheDocument();
  });

  it("renders the plain row outside the addon group", () => {
    renderTab(envVars);
    const group = screen.getByTestId("env-addon-group");
    const plainRow = screen.getByTestId("env-row-0-2");
    expect(plainRow).toBeInTheDocument();
    expect(group).not.toContainElement(plainRow);
  });

  it("renders the plain row BEFORE the addon group even though it comes later in the array", () => {
    // PLAIN (idx 2) sits after the two addon rows in the env array, but plain
    // rows render first, so it must precede the dashed addon group in the DOM.
    renderTab(envVars);
    const plainRow = screen.getByTestId("env-row-0-2");
    const group = screen.getByTestId("env-addon-group");
    expect(
      plainRow.compareDocumentPosition(group) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });

  it("has an Add binding button inside the group", () => {
    renderTab(envVars);
    const group = screen.getByTestId("env-addon-group");
    expect(within(group).getByRole("button", { name: /add binding/i })).toBeInTheDocument();
  });

  it("renders the Add variable button before the addon group", () => {
    // Add variable sits directly below the plain rows, above the addon
    // fieldset, so a newly-added plain var appears at the click point.
    renderTab(envVars);
    const addVar = screen.getByRole("button", { name: /add variable/i });
    const group = screen.getByTestId("env-addon-group");
    expect(group).not.toContainElement(addVar);
    expect(
      addVar.compareDocumentPosition(group) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy();
  });
});
