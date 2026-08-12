// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, afterEach, beforeEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { CloudAlphaBanner, CLOUD_HOSTNAME, DISMISSED_KEY } from "../cloud-alpha-banner";

function setHostname(hostname: string) {
  Object.defineProperty(window, "location", {
    configurable: true,
    value: { ...window.location, hostname },
  });
}

beforeEach(() => {
  localStorage.clear();
});

afterEach(cleanup);

describe("CloudAlphaBanner", () => {
  it("stays hidden on a self-hosted install", () => {
    setHostname("stackdome.acme.internal");
    render(<CloudAlphaBanner />);
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });

  it("shows the alpha notice on the cloud host", () => {
    setHostname(CLOUD_HOSTNAME);
    render(<CloudAlphaBanner />);
    expect(screen.getByRole("status")).toHaveTextContent(
      /deleted 6 hours after they are created/i,
    );
  });

  it("stays hidden once dismissed", async () => {
    setHostname(CLOUD_HOSTNAME);
    render(<CloudAlphaBanner />);

    await userEvent.click(screen.getByRole("button", { name: "Dismiss" }));

    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(localStorage.getItem(DISMISSED_KEY)).toBe("true");

    cleanup();
    render(<CloudAlphaBanner />);
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });
});
