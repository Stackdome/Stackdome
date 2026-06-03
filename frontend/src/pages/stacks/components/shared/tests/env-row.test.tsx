// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EnvRow } from "../env-row";
import type { FormEnvVarData } from "@/pages/stacks/schemas/form-schema";

afterEach(cleanup);

const stackRow = (over: Partial<Extract<FormEnvVarData, { from: "stack" }>> = {}): FormEnvVarData => ({
  from: "stack",
  name: "PORT",
  value: "80",
  ...over,
});

const noopProps = {
  index: 0,
  resourceIndex: 0,
  onChangeName: vi.fn(),
  onChangeValue: vi.fn(),
  onBlur: vi.fn(),
  onRemove: vi.fn(),
};

describe("EnvRow (literal variant)", () => {
  it("renders editable KEY and VALUE inputs", () => {
    render(<EnvRow row={stackRow()} {...noopProps} />);
    const key = screen.getByPlaceholderText("KEY") as HTMLInputElement;
    const value = screen.getByPlaceholderText("VALUE") as HTMLInputElement;
    expect(key.value).toBe("PORT");
    expect(value.value).toBe("80");
    expect(key).not.toBeDisabled();
    expect(value).not.toBeDisabled();
  });

  it("calls onChangeName when the KEY is edited", async () => {
    const user = userEvent.setup();
    const onChangeName = vi.fn();
    render(<EnvRow row={stackRow({ name: "" })} {...noopProps} onChangeName={onChangeName} />);
    await user.type(screen.getByPlaceholderText("KEY"), "A");
    expect(onChangeName).toHaveBeenCalledWith("A");
  });

  it("calls onChangeValue when the VALUE is edited", async () => {
    const user = userEvent.setup();
    const onChangeValue = vi.fn();
    render(<EnvRow row={stackRow({ value: "" })} {...noopProps} onChangeValue={onChangeValue} />);
    await user.type(screen.getByPlaceholderText("VALUE"), "x");
    expect(onChangeValue).toHaveBeenCalledWith("x");
  });

  it("calls onRemove when the remove button is clicked", async () => {
    const user = userEvent.setup();
    const onRemove = vi.fn();
    render(<EnvRow row={stackRow()} {...noopProps} onRemove={onRemove} />);
    await user.click(screen.getByRole("button", { name: /remove env var/i }));
    expect(onRemove).toHaveBeenCalled();
  });

  it("renders duplicate name error on the name input", () => {
    render(
      <EnvRow
        row={stackRow()}
        {...noopProps}
        rowErrors={{ duplicate: 'Duplicate name "PORT"' }}
      />,
    );
    expect(screen.getByPlaceholderText("KEY")).toHaveClass("border-danger");
    expect(screen.getByText('Duplicate name "PORT"')).toBeInTheDocument();
  });

  it("renders required name error on the name input", () => {
    render(<EnvRow row={stackRow({ name: "" })} {...noopProps} rowErrors={{ name: "Required" }} />);
    expect(screen.getByPlaceholderText("KEY")).toHaveClass("border-danger");
    expect(screen.getByText("Required")).toBeInTheDocument();
  });

  it("shows the reset affordance for a modified row and calls onReset", async () => {
    const user = userEvent.setup();
    const onReset = vi.fn();
    render(<EnvRow row={stackRow()} {...noopProps} status="modified" onReset={onReset} />);
    await user.click(screen.getByRole("button", { name: /reset env var to original value/i }));
    expect(onReset).toHaveBeenCalled();
  });
});
