// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EnvRow } from "../env-row";
import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";

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
type SelfRow = Extract<FormEnvVarData, { from: "self" }>;

const baseAddonRow = (over: Partial<AddonRow> = {}): AddonRow => ({
  from: "addon",
  name: "PG_HOST",
  addonId: "addon-1",
  database: "tooljet",
  superuser: false,
  credField: "host",
  ...over,
});

const baseSelfRow = (over: Partial<SelfRow> = {}): SelfRow => ({
  from: "self",
  name: "MY_URL",
  selfOutput: "",
  ...over,
});

const noopProps = {
  index: 0,
  resourceIndex: 0,
  secrets: [],
  secretsLoading: false,
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

  it("renders the field picker on addon rows", () => {
    // After the inline-addon-picker refactor, only the credField picker lives
    // on each env row. Addon + database selection moved up to the addon-group
    // header in environment-tab.
    render(<EnvRow row={baseAddonRow()} {...noopProps} />);
    expect(screen.getByTestId("field-picker-trigger")).toBeInTheDocument();
    expect(screen.queryByTestId("addon-picker-trigger")).not.toBeInTheDocument();
    expect(screen.queryByTestId("database-picker-trigger")).not.toBeInTheDocument();
  });

  it("field picker is disabled when no addon is picked", () => {
    render(
      <EnvRow
        row={baseAddonRow({ addonId: "", database: undefined, credField: undefined })}
        {...noopProps}
      />,
    );
    expect(screen.getByTestId("field-picker-trigger")).toBeDisabled();
  });

  it("field picker lists all 8 ADDON_OUTPUT_FIELDS", async () => {
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
      "url",
      "ca_certificate",
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
    expect(screen.getByTestId("field-picker-trigger")).not.toHaveClass("border-danger");
    expect(screen.getByPlaceholderText("KEY")).not.toHaveClass("border-danger");
  });

  it("renders red border + message on field picker when rowErrors.credField set", () => {
    render(
      <EnvRow
        row={baseAddonRow({ credField: undefined })}
        {...noopProps}
        rowErrors={{ credField: "Pick a field" }}
      />,
    );
    expect(screen.getByTestId("field-picker-trigger")).toHaveClass("border-danger");
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
    expect(screen.getByPlaceholderText("KEY")).toHaveClass("border-danger");
    expect(screen.getByText('Duplicate name "PG_HOST"')).toBeInTheDocument();
  });

  it("orphan addon row renders read-only second line with warning, but From dropdown still works", () => {
    render(
      <EnvRow
        row={baseAddonRow({ addonId: "missing-addon", database: "tooljet", credField: "host" })}
        {...noopProps}
        addonNameById={new Map([["addon-1", "tooljet-db"]])}
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
});

describe("EnvRow (self variant)", () => {
  it("lists this resource's declared outputs and reports the chosen one", async () => {
    const user = userEvent.setup();
    const onChangeSelf = vi.fn();
    render(
      <EnvRow
        row={baseSelfRow()}
        {...noopProps}
        selfOutputs={["public.http.url", "host"]}
        onChangeSelf={onChangeSelf}
      />,
    );
    await user.click(screen.getByTestId("self-output-trigger"));
    expect(
      await screen.findByRole("option", { name: "public.http.url" }),
    ).toBeInTheDocument();
    await user.click(screen.getByRole("option", { name: "public.http.url" }));
    expect(onChangeSelf).toHaveBeenCalledWith("public.http.url");
  });
});
