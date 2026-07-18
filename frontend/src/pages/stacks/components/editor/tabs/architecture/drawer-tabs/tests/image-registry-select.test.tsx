// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeAll, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ImageRegistrySelect } from "../image-registry-select";

// cmdk (used by the Command primitive) reads ResizeObserver on mount, and
// calls scrollIntoView when the highlighted item changes — neither of which
// jsdom implements.
beforeAll(() => {
  global.ResizeObserver =
    global.ResizeObserver ||
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    };
  Element.prototype.scrollIntoView = Element.prototype.scrollIntoView || (() => {});
});

afterEach(cleanup);

vi.mock("@/api/registry-credentials", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  listRegistryCredentials: vi.fn(),
}));
vi.mock("@/helpers/common", () => ({ getCurrentOrganizationId: () => "org-1" }));

import { listRegistryCredentials } from "@/api/registry-credentials";

const ghcr = { id: "cred-ghcr", host: "ghcr.io", username: "acme-pull-bot" };
const ecr = { id: "cred-ecr", host: "123.dkr.ecr.us-east-1.amazonaws.com", username: "aws" };

beforeEach(() => {
  vi.mocked(listRegistryCredentials).mockResolvedValue({ items: [ghcr, ecr] });
});

describe("ImageRegistrySelect", () => {
  it("preselects the credential matching the ref host", async () => {
    render(
      <ImageRegistrySelect id="reg" imageRef="ghcr.io/acme/api:1" registryCredentialsId="cred-ghcr" onChange={vi.fn()} />,
    );
    await waitFor(() => expect(screen.getByRole("combobox")).toHaveTextContent("ghcr.io"));
  });

  it("shows Public / Docker Hub for a bare ref", async () => {
    render(<ImageRegistrySelect id="reg" imageRef="nginx:latest" onChange={vi.fn()} />);
    await waitFor(() => expect(screen.getByRole("combobox")).toHaveTextContent(/public/i));
  });

  it("re-prefixes the ref and sets credentials on pick", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<ImageRegistrySelect id="reg" imageRef="acme/api:1" onChange={onChange} />);
    await user.click(screen.getByRole("combobox"));
    await user.click(await screen.findByText("ghcr.io"));
    expect(onChange).toHaveBeenCalledWith({
      ref: "ghcr.io/acme/api:1",
      registry_credentials_id: "cred-ghcr",
    });
  });

  it("strips the host and clears credentials when Public is picked", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <ImageRegistrySelect id="reg" imageRef="ghcr.io/acme/api:1" registryCredentialsId="cred-ghcr" onChange={onChange} />,
    );
    await user.click(screen.getByRole("combobox"));
    await user.click(await screen.findByText(/public \/ docker hub/i));
    expect(onChange).toHaveBeenCalledWith({
      ref: "acme/api:1",
      registry_credentials_id: undefined,
    });
  });

  it("accepts a custom host and clears credentials", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<ImageRegistrySelect id="reg" imageRef="acme/api:1" onChange={onChange} />);
    await user.click(screen.getByRole("combobox"));
    await user.type(screen.getByPlaceholderText(/registry host/i), "registry.example.com");
    await user.click(await screen.findByText(/use "registry\.example\.com"/i));
    expect(onChange).toHaveBeenCalledWith({
      ref: "registry.example.com/acme/api:1",
      registry_credentials_id: undefined,
    });
  });
});
