// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EnvRow } from "./env-row";
import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";
import type { PostgresAddon, PostgresAddonSpec } from "@/api/addons";

afterEach(cleanup);

// Radix Select uses pointer-capture APIs that jsdom doesn't implement.
// Stub them so userEvent.click on a SelectTrigger opens the popover.
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

type AddonRow = Extract<FormEnvVarData, { from: "addon" }>;

const mkAddon = (over: Partial<PostgresAddon> = {}): PostgresAddon => {
  const baseSpec: PostgresAddonSpec = {
    version: { major: 17 },
    storage: { size: "5Gi" },
    databases: [{ name: "tooljet" }, { name: "analytics" }],
    configuration: { enable_superuser_access: false },
  } as PostgresAddonSpec;
  return {
    id: "addon-1",
    name: "tooljet-db",
    status: { state: "Ready" },
    spec: baseSpec,
    ...over,
  } as PostgresAddon;
};

const baseAddonRow = (over: Partial<AddonRow> = {}): AddonRow => ({
  from: "addon",
  name: "PG_HOST",
  addonType: "postgres",
  addonId: "addon-1",
  database: "tooljet",
  superuser: false,
  credField: "host",
  ...over,
});

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
        row={baseAddonRow({ addonId: "" })}
        {...noopProps}
        addons={[
          mkAddon({ id: "a", name: "primary", status: { state: "Ready" } }),
          mkAddon({ id: "b", name: "secondary", status: { state: "Pending" } }),
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
      <EnvRow row={baseAddonRow({ addonId: "" })} {...noopProps} addons={[]} />,
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
        row={baseAddonRow({ addonId: "" })}
        {...noopProps}
        onChangeAddon={onChangeAddon}
        addons={[mkAddon({ id: "addon-x" })]}
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
        row={baseAddonRow({ addonId: "", database: undefined })}
        {...noopProps}
        addons={[mkAddon()]}
      />,
    );
    expect(screen.getByTestId("database-picker-trigger")).toBeDisabled();
  });

  it("database picker lists addon's databases when an addon is picked", async () => {
    const user = userEvent.setup();
    render(<EnvRow row={baseAddonRow({ database: undefined })} {...noopProps} />);
    await user.click(screen.getByTestId("database-picker-trigger"));
    expect(await screen.findByRole("option", { name: /tooljet/i })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: /analytics/i })).toBeInTheDocument();
  });

  it("does NOT show 'All databases' when addon does not enable superuser", async () => {
    const user = userEvent.setup();
    render(<EnvRow row={baseAddonRow({ database: undefined })} {...noopProps} />);
    await user.click(screen.getByTestId("database-picker-trigger"));
    expect(screen.queryByRole("option", { name: /all databases/i })).not.toBeInTheDocument();
  });

  it("shows 'All databases' when addon enables superuser", async () => {
    const user = userEvent.setup();
    const su = mkAddon({
      spec: {
        ...mkAddon().spec,
        configuration: { enable_superuser_access: true },
      } as PostgresAddonSpec,
    });
    render(
      <EnvRow row={baseAddonRow({ database: undefined })} {...noopProps} addons={[su]} />,
    );
    await user.click(screen.getByTestId("database-picker-trigger"));
    expect(await screen.findByRole("option", { name: /all databases/i })).toBeInTheDocument();
  });

  it("calls onChangeAddon with superuser=true and database=null when 'All databases' is picked", async () => {
    const user = userEvent.setup();
    const onChangeAddon = vi.fn();
    const su = mkAddon({
      spec: {
        ...mkAddon().spec,
        configuration: { enable_superuser_access: true },
      } as PostgresAddonSpec,
    });
    render(
      <EnvRow
        row={baseAddonRow({ database: undefined })}
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
        row={baseAddonRow({ addonId: "", database: undefined, credField: undefined })}
        {...noopProps}
        addons={[mkAddon()]}
      />,
    );
    expect(screen.getByTestId("field-picker-trigger")).toBeDisabled();
  });

  it("field picker lists all 8 CRED_FIELDS", async () => {
    const user = userEvent.setup();
    render(<EnvRow row={baseAddonRow()} {...noopProps} />);
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
    render(<EnvRow row={baseAddonRow()} {...noopProps} />);
    await user.click(screen.getByTestId("field-picker-trigger"));
    const hostOption = await screen.findByRole("option", { name: /host/i });
    expect(hostOption).toHaveTextContent(/cluster/i);
    const userOption = screen.getByRole("option", { name: /^username/i });
    expect(userOption).not.toHaveTextContent(/cluster/i);
  });

  it("calls onChangeAddon with credField when a field is picked", async () => {
    const user = userEvent.setup();
    const onChangeAddon = vi.fn();
    render(<EnvRow row={baseAddonRow()} {...noopProps} onChangeAddon={onChangeAddon} />);
    await user.click(screen.getByTestId("field-picker-trigger"));
    await user.click(await screen.findByRole("option", { name: /port/i }));
    expect(onChangeAddon).toHaveBeenCalledWith({ credField: "port" });
  });

  it("renders no error styling without rowErrors", () => {
    render(<EnvRow row={baseAddonRow()} {...noopProps} />);
    expect(screen.getByTestId("addon-picker-trigger")).not.toHaveClass("border-destructive");
  });

  it("renders red border + message on addon picker when rowErrors.addonId set", () => {
    render(
      <EnvRow
        row={baseAddonRow({ addonId: "" })}
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
        row={baseAddonRow({ database: undefined })}
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
        row={baseAddonRow({ credField: undefined })}
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
        row={baseAddonRow()}
        {...noopProps}
        rowErrors={{ duplicate: 'Duplicate name "PG_HOST"' }}
      />,
    );
    expect(screen.getByPlaceholderText("KEY")).toHaveClass("border-destructive");
    expect(screen.getByText('Duplicate name "PG_HOST"')).toBeInTheDocument();
  });

  it("orphan addon row renders read-only second line with warning, but From dropdown still works", () => {
    render(
      <EnvRow
        row={baseAddonRow({ addonId: "missing-addon", database: "tooljet", credField: "host" })}
        {...noopProps}
        addonNameById={new Map([["addon-1", "tooljet-db"]])}
        addons={[mkAddon()]}
      />,
    );
    expect(screen.getByText(/<missing addon>/i)).toBeInTheDocument();
    expect(
      screen.getByText(/Addon was deleted\. This variable won't resolve\. Remove to clean up\./i),
    ).toBeInTheDocument();
    expect(screen.queryByTestId("addon-picker-trigger")).not.toBeInTheDocument();
    expect(screen.queryByTestId("database-picker-trigger")).not.toBeInTheDocument();
    expect(screen.queryByTestId("field-picker-trigger")).not.toBeInTheDocument();
    const fromTrigger = screen
      .getAllByRole("combobox")
      .find((el) => el.textContent?.match(/Addon/));
    expect(fromTrigger).not.toBeDisabled();
  });

  it("auto-selects the only database when picking an addon with one db and no superuser", async () => {
    const user = userEvent.setup();
    const onChangeAddon = vi.fn();
    const single = mkAddon({
      id: "addon-single",
      spec: {
        ...mkAddon().spec,
        databases: [{ name: "only-one" }],
      } as PostgresAddonSpec,
    });
    render(
      <EnvRow
        row={baseAddonRow({ addonId: "", database: undefined })}
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
