// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";

afterEach(cleanup);
import { EnvRow } from "../src/pages/stacks/components/shared/env-row";
import type { FormEnvVarData } from "../src/pages/stacks/schemas/form-schema";
import type { PostgresAddon } from "../src/api/addons";

const mkAddon = (over: Partial<PostgresAddon> = {}): PostgresAddon => ({
  id: "addon-1",
  name: "tooljet-db",
  status: { state: "Ready" },
  spec: {
    version: { major: 17 },
    storage: { size: "5Gi" },
    databases: [{ name: "tooljet" }, { name: "analytics" }],
    configuration: { enable_superuser_access: false },
  } as any,
  ...(over as any),
});

const baseAddonRow = (
  over: Partial<Extract<FormEnvVarData, { from: "addon" }>> = {},
): FormEnvVarData =>
  ({
    from: "addon",
    name: "PG_HOST",
    addonType: "postgres",
    addonId: "addon-1",
    database: "tooljet",
    superuser: false,
    credField: "host",
    ...over,
  }) as FormEnvVarData;

const noopProps = {
  index: 0,
  resourceIndex: 0,
  secrets: [],
  secretsLoading: false,
  addons: [mkAddon()],
  addonNameById: new Map([["addon-1", "tooljet-db"]]),
  onChangeName: vi.fn(),
  onChangeValue: vi.fn(),
  onChangeFrom: vi.fn(),
  onChangeSecret: vi.fn(),
  onChangeAddon: vi.fn(),
  onBlur: vi.fn(),
  onRemove: vi.fn(),
};

describe("EnvRow (addon variant)", () => {
  it("makes name input editable on addon rows", () => {
    render(<EnvRow row={baseAddonRow()} {...noopProps} />);
    const nameInput = screen.getByPlaceholderText("KEY");
    expect(nameInput).not.toBeDisabled();
  });

  it("From dropdown is enabled and Addon item is enabled on addon rows", () => {
    render(<EnvRow row={baseAddonRow()} {...noopProps} />);
    const fromTrigger = screen
      .getAllByRole("combobox")
      .find((el) => el.textContent?.match(/Stack|Secret|Addon/));
    expect(fromTrigger).toBeDefined();
    expect(fromTrigger).not.toBeDisabled();
  });

  it("renders three picker triggers (Addon, Database, Field) on the second line of an addon row", () => {
    render(<EnvRow row={baseAddonRow()} {...noopProps} />);
    expect(screen.getByTestId("addon-picker-trigger")).toBeInTheDocument();
    expect(screen.getByTestId("database-picker-trigger")).toBeInTheDocument();
    expect(screen.getByTestId("field-picker-trigger")).toBeInTheDocument();
  });
});
