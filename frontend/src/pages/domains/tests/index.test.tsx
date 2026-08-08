// @vitest-environment jsdom
import { describe, it, expect, afterEach, vi } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { SheetHost } from "@/test-support/sheet-host";

import DomainsPage from "../index";
import type { Organization } from "@/api/organizations";

const getOrganizationMock = vi.fn();

vi.mock("@/api/organizations", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/api/organizations")>()),
  getOrganization: (...args: unknown[]) => getOrganizationMock(...args),
}));

vi.mock("@/lib/common", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/lib/common")>()),
  getCurrentOrganizationId: () => "org-1",
}));

const baseOrganization = {
  id: "org-1",
  name: "Acme",
  is_platform: true,
} as Organization;

afterEach(cleanup);

describe("DomainsPage", () => {
  it("disables Add Domain when one domain exists", async () => {
    getOrganizationMock.mockResolvedValue({
      ...baseOrganization,
      domains: [{ fqdn: "apps.acme.dev" }],
    });
    render(
      <MemoryRouter initialEntries={["/domains"]}>
        <SheetHost>
          <DomainsPage />
        </SheetHost>
      </MemoryRouter>,
    );
    expect(await screen.findByText("All Domains")).toBeInTheDocument();
    expect(screen.getByText("apps.acme.dev")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Add Domain/ })).toBeDisabled();
  });

  it("keeps Add Domain enabled when no domains exist", async () => {
    getOrganizationMock.mockResolvedValue({ ...baseOrganization, domains: [] });
    render(
      <MemoryRouter initialEntries={["/domains"]}>
        <SheetHost>
          <DomainsPage />
        </SheetHost>
      </MemoryRouter>,
    );
    expect(await screen.findByText("No domain configured")).toBeInTheDocument();
    for (const button of screen.getAllByRole("button", { name: /Add Domain/ })) {
      expect(button).toBeEnabled();
    }
  });
});
