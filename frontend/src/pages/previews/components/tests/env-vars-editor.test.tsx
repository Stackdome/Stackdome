// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { useState } from "react";
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { EnvVarsEditor, type EnvVarFormRow } from "../env-vars-editor";
import type { Secret } from "@/api/secrets";

vi.mock("@/pages/secrets/hooks/use-secrets", () => ({
  useSecrets: () => ({
    secrets: [
      { id: "s1", name: "api-token", type: "Generic" },
      { id: "s2", name: "tls-cert", type: "TLS" },
    ] as Secret[],
    loading: false,
    error: null,
    refetch: vi.fn(),
  }),
}));

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

  it("round-trips a {{ secret.NAME }} value into Secret mode with the secret selected", () => {
    render(<Harness initial={[{ name: "TOKEN", value: "{{ secret.api-token }}" }]} onChange={vi.fn()} />);

    expect(screen.getByRole("combobox", { name: "Value source" })).toHaveTextContent("Secret");
    expect(screen.getByRole("combobox", { name: "Secret" })).toHaveTextContent("api-token");
    expect(screen.queryByLabelText(/^variable value$/i)).not.toBeInTheDocument();
  });

  it("writes a secret reference once the source is Secret and a secret is picked", async () => {
    const onChange = vi.fn();
    render(<Harness initial={[{ name: "TOKEN", value: "plain" }]} onChange={onChange} />);

    await userEvent.click(screen.getByRole("combobox", { name: "Value source" }));
    await userEvent.click(screen.getByRole("option", { name: "Secret" }));
    expect(onChange).toHaveBeenLastCalledWith([{ name: "TOKEN", value: "{{ secret. }}" }]);

    await userEvent.click(screen.getByRole("combobox", { name: "Secret" }));
    await userEvent.click(screen.getByRole("option", { name: "api-token" }));
    expect(onChange).toHaveBeenLastCalledWith([{ name: "TOKEN", value: "{{ secret.api-token }}" }]);
  });

  it("keeps literal values editable as plain text", () => {
    render(<Harness initial={[{ name: "PORT", value: "8080" }]} onChange={vi.fn()} />);

    expect(screen.getByRole("combobox", { name: "Value source" })).toHaveTextContent("Plain text");
    expect(screen.getByLabelText(/^variable value$/i)).toHaveValue("8080");
  });
});
