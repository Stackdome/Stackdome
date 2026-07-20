// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { useState } from "react";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EnvVarsEditor, type EnvVarFormRow } from "../env-vars-editor";

afterEach(() => cleanup());

/** Wraps the editor with real state so keystrokes behave like they do when
 *  mounted in a form, and exposes every onChange call via the spy. */
function Harness({
  initial,
  onChange,
}: {
  initial: EnvVarFormRow[];
  onChange: (rows: EnvVarFormRow[]) => void;
}) {
  const [rows, setRows] = useState(initial);
  return (
    <EnvVarsEditor
      value={rows}
      onChange={(next) => {
        setRows(next);
        onChange(next);
      }}
    />
  );
}

describe("EnvVarsEditor", () => {
  it("adds a row when Add variable is clicked", async () => {
    const onChange = vi.fn();
    render(<Harness initial={[]} onChange={onChange} />);

    await userEvent.click(screen.getByRole("button", { name: /add variable/i }));

    expect(onChange).toHaveBeenCalledWith([{ name: "", value: "" }]);
  });

  it("edits a row's name and value", async () => {
    const onChange = vi.fn();
    render(<Harness initial={[{ name: "", value: "" }]} onChange={onChange} />);

    await userEvent.type(screen.getByLabelText(/^variable name$/i), "FOO");
    await userEvent.type(screen.getByLabelText(/^variable value$/i), "bar");

    expect(onChange).toHaveBeenLastCalledWith([{ name: "FOO", value: "bar" }]);
  });

  it("removes a row", async () => {
    const onChange = vi.fn();
    render(
      <Harness
        initial={[
          { name: "FOO", value: "1" },
          { name: "BAR", value: "2" },
        ]}
        onChange={onChange}
      />,
    );

    await userEvent.click(screen.getAllByRole("button", { name: /remove variable/i })[0]);

    expect(onChange).toHaveBeenLastCalledWith([{ name: "BAR", value: "2" }]);
  });
});
