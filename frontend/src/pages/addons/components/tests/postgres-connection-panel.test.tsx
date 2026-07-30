// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { PostgresAddon, PostgresCredentials } from "@/api/addons";

vi.mock("@/api/addons", async () => {
  const actual = await vi.importActual<typeof import("@/api/addons")>("@/api/addons");
  return { ...actual, getPostgresCredentials: vi.fn() };
});

vi.mock("@/helpers/common", () => ({
  getCurrentOrganizationId: vi.fn(() => "org-1"),
}));

vi.mock("@/api/client", async () => {
  const actual = await vi.importActual<typeof import("@/api/client")>("@/api/client");
  return {
    ...actual,
    getErrorMessage: vi.fn((e: unknown) => (e as { message?: string })?.message ?? "error"),
  };
});

import { getPostgresCredentials } from "@/api/addons";
import { PostgresConnectionPanel } from "../postgres-connection-panel";

const mockedCredentials = vi.mocked(getPostgresCredentials);

const CREDENTIALS: PostgresCredentials = {
  database: "app",
  host: "pg-rw.addons.svc.cluster.local",
  port: 5432,
  username: "app",
  password: "s3cr3t-pw",
  sslMode: "require",
  connectionString: "postgresql://app:s3cr3t-pw@pg-rw.addons.svc.cluster.local:5432/app?sslmode=require",
  caCertificate: "-----BEGIN CERTIFICATE-----",
};

function mkAddon(overrides: Partial<PostgresAddon> = {}): PostgresAddon {
  return {
    id: "addon-1",
    name: "asdasdff12",
    spec: {
      version: { major: 17 },
      instances: { count: 1 },
      storage: { size: "10Gi" },
      configuration: { enable_superuser_access: false },
    },
    status: {
      state: "Ready",
      connection_info: {
        host: "pg-rw.addons.svc.cluster.local",
        port: 5432,
        databases: [{ name: "app", owner: "app" }],
        credentials: { app_user_secrets: { app: "addon-1-app" } },
      },
    },
    ...overrides,
  } as PostgresAddon;
}

beforeEach(() => {
  mockedCredentials.mockReset();
  Object.assign(navigator, { clipboard: { writeText: vi.fn().mockResolvedValue(undefined) } });
});
afterEach(cleanup);

describe("PostgresConnectionPanel", () => {
  it("shows the in-cluster endpoint and databases from the addon status", () => {
    render(<PostgresConnectionPanel addon={mkAddon()} projectName="default" />);

    expect(screen.getByText("pg-rw.addons.svc.cluster.local")).toBeInTheDocument();
    expect(screen.getByText(":5432")).toBeInTheDocument();
    expect(screen.getAllByText("app").length).toBeGreaterThan(0);
    expect(screen.getByText("addon-1-app")).toBeInTheDocument();
  });

  it("copies the endpoint as host:port", async () => {
    render(<PostgresConnectionPanel addon={mkAddon()} projectName="default" />);

    await userEvent.click(screen.getByRole("button", { name: /copy endpoint/i }));

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      "pg-rw.addons.svc.cluster.local:5432",
    );
  });

  it("renders an empty state when the addon has no connection info", () => {
    const addon = mkAddon({ status: { state: "Creating" } });
    render(<PostgresConnectionPanel addon={addon} projectName="default" />);

    expect(screen.getByText("No connection details yet")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /show credentials/i })).not.toBeInTheDocument();
  });

  it("fetches credentials on demand and masks the password until revealed", async () => {
    mockedCredentials.mockResolvedValue(CREDENTIALS);
    render(<PostgresConnectionPanel addon={mkAddon()} projectName="default" />);

    await userEvent.click(screen.getByRole("button", { name: /show credentials/i }));

    await waitFor(() =>
      expect(mockedCredentials).toHaveBeenCalledWith("org-1", "default", "addon-1", "app", false),
    );
    expect(await screen.findByText("require")).toBeInTheDocument();
    expect(screen.queryByText(CREDENTIALS.password!)).not.toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: /reveal password/i }));
    expect(screen.getByText(CREDENTIALS.password!)).toBeInTheDocument();
  });

  it("copies the real password even while it is masked", async () => {
    mockedCredentials.mockResolvedValue(CREDENTIALS);
    render(<PostgresConnectionPanel addon={mkAddon()} projectName="default" />);

    await userEvent.click(screen.getByRole("button", { name: /show credentials/i }));
    await userEvent.click(await screen.findByRole("button", { name: /copy password/i }));

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(CREDENTIALS.password);
  });

  it("re-reads credentials with the superuser scope when superuser access is on", async () => {
    mockedCredentials.mockResolvedValue(CREDENTIALS);
    const addon = mkAddon();
    addon.spec.configuration = { enable_superuser_access: true };
    render(<PostgresConnectionPanel addon={addon} projectName="default" />);

    await userEvent.click(screen.getByRole("button", { name: /show credentials/i }));
    await userEvent.click(await screen.findByRole("button", { name: "Superuser" }));

    await waitFor(() =>
      expect(mockedCredentials).toHaveBeenLastCalledWith("org-1", "default", "addon-1", "app", true),
    );
  });

  it("surfaces a fetch failure inline with a retry", async () => {
    mockedCredentials.mockRejectedValueOnce(new Error("cluster unreachable"));
    render(<PostgresConnectionPanel addon={mkAddon()} projectName="default" />);

    await userEvent.click(screen.getByRole("button", { name: /show credentials/i }));
    expect(await screen.findByText("cluster unreachable")).toBeInTheDocument();

    mockedCredentials.mockResolvedValue(CREDENTIALS);
    await userEvent.click(screen.getByRole("button", { name: /retry/i }));

    expect(await screen.findByText("require")).toBeInTheDocument();
  });
});
