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

  it("shows '+ Create Postgres addon' link when addons list is empty", async () => {
    const user = userEvent.setup();
    render(
      <EnvRow row={baseAddonRow({ addonId: "" }) as any} {...noopProps} addons={[]} />,
    );
    await user.click(screen.getByTestId("addon-picker-trigger"));
    const link = await screen.findByRole("link", { name: /create postgres addon/i });
    expect(link).toHaveAttribute("href", "/addons/create/postgres");
    expect(link).toHaveAttribute("target", "_blank");
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

  it("database picker is disabled when no addon is picked", () => {
    render(
      <EnvRow
        row={baseAddonRow({ addonId: "", database: undefined }) as any}
        {...noopProps}
        addons={[mkAddon()]}
      />,
    );
    expect(screen.getByTestId("database-picker-trigger")).toBeDisabled();
  });

  it("database picker lists addon's databases when an addon is picked", async () => {
    const user = userEvent.setup();
    render(<EnvRow row={baseAddonRow({ database: undefined }) as any} {...noopProps} />);
    await user.click(screen.getByTestId("database-picker-trigger"));
    expect(await screen.findByRole("option", { name: /tooljet/i })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: /analytics/i })).toBeInTheDocument();
  });

  it("does NOT show 'All databases' when addon does not enable superuser", async () => {
    const user = userEvent.setup();
    render(<EnvRow row={baseAddonRow({ database: undefined }) as any} {...noopProps} />);
    await user.click(screen.getByTestId("database-picker-trigger"));
    expect(screen.queryByRole("option", { name: /all databases/i })).not.toBeInTheDocument();
  });

  it("shows 'All databases' when addon enables superuser", async () => {
    const user = userEvent.setup();
    const su = mkAddon({
      spec: { ...mkAddon().spec, configuration: { enable_superuser_access: true } } as any,
    });
    render(
      <EnvRow row={baseAddonRow({ database: undefined }) as any} {...noopProps} addons={[su]} />,
    );
    await user.click(screen.getByTestId("database-picker-trigger"));
    expect(await screen.findByRole("option", { name: /all databases/i })).toBeInTheDocument();
  });

  it("calls onChangeAddon with superuser=true and database=null when 'All databases' is picked", async () => {
    const user = userEvent.setup();
    const onChangeAddon = vi.fn();
    const su = mkAddon({
      spec: { ...mkAddon().spec, configuration: { enable_superuser_access: true } } as any,
    });
    render(
      <EnvRow
        row={baseAddonRow({ database: undefined }) as any}
        {...noopProps}
        addons={[su]}
        onChangeAddon={onChangeAddon}
      />,
    );
    await user.click(screen.getByTestId("database-picker-trigger"));
    await user.click(await screen.findByRole("option", { name: /all databases/i }));
    expect(onChangeAddon).toHaveBeenCalledWith({ database: null, superuser: true });
  });

  it("field picker is disabled when no addon is picked", () => {
    render(
      <EnvRow
        row={baseAddonRow({ addonId: "", database: undefined, credField: undefined as any }) as any}
        {...noopProps}
        addons={[mkAddon()]}
      />,
    );
    expect(screen.getByTestId("field-picker-trigger")).toBeDisabled();
  });

  it("field picker lists all 8 CRED_FIELDS", async () => {
    const user = userEvent.setup();
    render(<EnvRow row={baseAddonRow() as any} {...noopProps} />);
    await user.click(screen.getByTestId("field-picker-trigger"));
    for (const f of [
      "host",
      "port",
      "username",
      "password",
      "database",
      "sslmode",
      "connectionString",
      "caCertificate",
    ]) {
      expect(await screen.findByRole("option", { name: new RegExp(f, "i") })).toBeInTheDocument();
    }
  });

  it("cluster-wide fields show 'cluster' badge in dropdown items", async () => {
    const user = userEvent.setup();
    render(<EnvRow row={baseAddonRow() as any} {...noopProps} />);
    await user.click(screen.getByTestId("field-picker-trigger"));
    const hostOption = await screen.findByRole("option", { name: /host/i });
    expect(hostOption).toHaveTextContent(/cluster/i);
    const userOption = screen.getByRole("option", { name: /^username/i });
    expect(userOption).not.toHaveTextContent(/cluster/i);
  });

  it("calls onChangeAddon with credField when a field is picked", async () => {
    const user = userEvent.setup();
    const onChangeAddon = vi.fn();
    render(<EnvRow row={baseAddonRow() as any} {...noopProps} onChangeAddon={onChangeAddon} />);
    await user.click(screen.getByTestId("field-picker-trigger"));
    await user.click(await screen.findByRole("option", { name: /port/i }));
    expect(onChangeAddon).toHaveBeenCalledWith({ credField: "port" });
  });

  it("renders no error styling without rowErrors", () => {
    render(<EnvRow row={baseAddonRow() as any} {...noopProps} />);
    expect(screen.getByTestId("addon-picker-trigger")).not.toHaveClass("border-destructive");
  });

  it("renders red border + message on addon picker when rowErrors.addonId set", () => {
    render(
      <EnvRow
        row={baseAddonRow({ addonId: "" }) as any}
        {...noopProps}
        rowErrors={{ addonId: "Pick an addon" }}
      />,
    );
    expect(screen.getByTestId("addon-picker-trigger")).toHaveClass("border-destructive");
    expect(screen.getByText("Pick an addon")).toBeInTheDocument();
  });

  it("renders red border + message on database picker when rowErrors.database set", () => {
    render(
      <EnvRow
        row={baseAddonRow({ database: undefined }) as any}
        {...noopProps}
        rowErrors={{ database: "Pick a database" }}
      />,
    );
    expect(screen.getByTestId("database-picker-trigger")).toHaveClass("border-destructive");
    expect(screen.getByText("Pick a database")).toBeInTheDocument();
  });

  it("renders red border + message on field picker when rowErrors.credField set", () => {
    render(
      <EnvRow
        row={baseAddonRow({ credField: undefined as any }) as any}
        {...noopProps}
        rowErrors={{ credField: "Pick a field" }}
      />,
    );
    expect(screen.getByTestId("field-picker-trigger")).toHaveClass("border-destructive");
    expect(screen.getByText("Pick a field")).toBeInTheDocument();
  });

  it("renders duplicate name error on the name input", () => {
    render(
      <EnvRow
        row={baseAddonRow() as any}
        {...noopProps}
        rowErrors={{ duplicate: 'Duplicate name "PG_HOST"' }}
      />,
    );
    expect(screen.getByPlaceholderText("KEY")).toHaveClass("border-destructive");
    expect(screen.getByText('Duplicate name "PG_HOST"')).toBeInTheDocument();
  });

  it("auto-selects the only database when picking an addon with one db and no superuser", async () => {
    const user = userEvent.setup();
    const onChangeAddon = vi.fn();
    const single = mkAddon({
      id: "addon-single",
      spec: { ...mkAddon().spec, databases: [{ name: "only-one" }] } as any,
    });
    render(
      <EnvRow
        row={baseAddonRow({ addonId: "", database: undefined }) as any}
        {...noopProps}
        addons={[single]}
        onChangeAddon={onChangeAddon}
      />,
    );
    await user.click(screen.getByTestId("addon-picker-trigger"));
    await user.click(await screen.findByRole("option", { name: /tooljet-db/i }));
    expect(onChangeAddon).toHaveBeenCalledWith({
      addonId: "addon-single",
      database: "only-one",
      superuser: false,
    });
  });
});
