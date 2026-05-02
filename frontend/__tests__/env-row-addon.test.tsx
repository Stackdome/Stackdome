// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

afterEach(cleanup);

// Radix Select uses pointer-capture APIs that jsdom doesn't implement.
// Stub them so userEvent.click on a SelectTrigger opens the popover.
beforeAll(() => {
  if (!(Element.prototype as any).hasPointerCapture) {
    (Element.prototype as any).hasPointerCapture = vi.fn(() => false);
  }
  if (!(Element.prototype as any).releasePointerCapture) {
    (Element.prototype as any).releasePointerCapture = vi.fn();
  }
  if (!(Element.prototype as any).setPointerCapture) {
    (Element.prototype as any).setPointerCapture = vi.fn();
  }
  if (!(Element.prototype as any).scrollIntoView) {
    (Element.prototype as any).scrollIntoView = vi.fn();
  }
});
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

  it("lists addons in the addon picker with name and state", async () => {
    const user = userEvent.setup();
    render(
      <EnvRow
        row={baseAddonRow({ addonId: "" }) as any}
        {...noopProps}
        addons={[
          mkAddon({ id: "a", name: "primary", status: { state: "Ready" } } as any),
          mkAddon({ id: "b", name: "secondary", status: { state: "Pending" } } as any),
        ]}
      />,
    );
    await user.click(screen.getByTestId("addon-picker-trigger"));
    expect(
      await screen.findByRole("option", { name: /primary.*Postgres.*Ready/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("option", { name: /secondary.*Postgres.*Pending/i }),
    ).toBeInTheDocument();
  });

  it("calls onChangeAddon when an addon is picked", async () => {
    const user = userEvent.setup();
    const onChangeAddon = vi.fn();
    render(
      <EnvRow
        row={baseAddonRow({ addonId: "" }) as any}
        {...noopProps}
        onChangeAddon={onChangeAddon}
        addons={[mkAddon({ id: "addon-x" } as any)]}
      />,
    );
    await user.click(screen.getByTestId("addon-picker-trigger"));
    await user.click(await screen.findByRole("option", { name: /tooljet-db/i }));
    expect(onChangeAddon).toHaveBeenCalledWith(
      expect.objectContaining({ addonId: "addon-x" }),
    );
  });
});
