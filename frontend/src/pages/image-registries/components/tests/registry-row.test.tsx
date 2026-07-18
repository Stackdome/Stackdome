// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { RegistryRow } from "../registry-row";
import { PURPOSE_PULL, PURPOSE_BOTH } from "../../lib/providers";
import type { RegistryCredential } from "@/api/registry-credentials";

afterEach(cleanup);

const credential: RegistryCredential = {
  id: "r1",
  host: "index.docker.io",
  username: "bob",
  purpose: PURPOSE_BOTH,
};

function renderRow(props: Partial<Parameters<typeof RegistryRow>[0]> = {}) {
  return render(
    <RegistryRow
      credential={credential}
      onVerify={vi.fn()}
      onUpdateCredentials={vi.fn()}
      onRemove={vi.fn()}
      {...props}
    />
  );
}

describe("RegistryRow", () => {
  it("displays the provider label and host", () => {
    renderRow();
    expect(screen.getByText("Docker Hub")).toBeInTheDocument();
    expect(screen.getByText("index.docker.io")).toBeInTheDocument();
    expect(screen.getByText("bob")).toBeInTheDocument();
  });

  it("labels a pull-and-push credential", () => {
    renderRow();
    expect(screen.getByText("Pull & push")).toBeInTheDocument();
  });

  it("labels a pull-only credential", () => {
    renderRow({ credential: { ...credential, purpose: PURPOSE_PULL } });
    expect(screen.getByText("Pull only")).toBeInTheDocument();
  });

  it("routes menu actions with the credential", async () => {
    const onVerify = vi.fn();
    const onUpdateCredentials = vi.fn();
    const onRemove = vi.fn();
    const user = userEvent.setup();
    renderRow({ onVerify, onUpdateCredentials, onRemove });

    await user.click(screen.getByRole("button", { name: /open row menu/i }));
    await user.click(await screen.findByRole("menuitem", { name: /update credentials/i }));
    await waitFor(() => expect(onUpdateCredentials).toHaveBeenCalledWith(credential));

    await user.click(screen.getByRole("button", { name: /open row menu/i }));
    await user.click(await screen.findByRole("menuitem", { name: /verify registry access/i }));
    await waitFor(() => expect(onVerify).toHaveBeenCalledWith(credential));

    await user.click(screen.getByRole("button", { name: /open row menu/i }));
    await user.click(await screen.findByRole("menuitem", { name: /remove registry/i }));
    await waitFor(() => expect(onRemove).toHaveBeenCalledWith(credential));
  });

});
