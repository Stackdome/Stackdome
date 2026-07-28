// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup, act } from "@testing-library/react";
import { PublicEndpointRow } from "../public-endpoint-row";

afterEach(cleanup);

const endpoints = [
  { service: "web", url: "https://web.acme.stackdome.app", port: 80 },
  { service: "api", url: "https://api.mycompany.com", port: 8080 },
];

describe("PublicEndpointRow", () => {
  beforeEach(() => {
    Object.assign(navigator, { clipboard: { writeText: vi.fn().mockResolvedValue(undefined) } });
  });

  it("renders one pill per endpoint with service chip and hover-revealed host", () => {
    render(<PublicEndpointRow endpoints={endpoints} />);
    expect(screen.getByText("PUBLIC")).toBeInTheDocument();
    expect(screen.getByText("web")).toBeInTheDocument();
    // Host text is in the DOM (revealed on hover via CSS), inside the go-to link.
    expect(screen.getByText("web.acme.stackdome.app")).toBeInTheDocument();
    expect(screen.getByText("api.mycompany.com")).toBeInTheDocument();
  });

  it("go-to link opens in a new tab", () => {
    render(<PublicEndpointRow endpoints={endpoints} />);
    const link = screen.getByRole("link", { name: "Go to https://web.acme.stackdome.app" });
    expect(link).toHaveAttribute("href", "https://web.acme.stackdome.app");
    expect(link).toHaveAttribute("target", "_blank");
    expect(link).toHaveAttribute("rel", "noreferrer");
  });

  it("copy button writes the url and flashes a check", async () => {
    vi.useFakeTimers();
    render(<PublicEndpointRow endpoints={[endpoints[0]]} />);
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Copy https://web.acme.stackdome.app" }));
    });
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith("https://web.acme.stackdome.app");
    expect(screen.getByRole("button", { name: "Copied" })).toBeInTheDocument();
    act(() => vi.advanceTimersByTime(1400));
    expect(screen.getByRole("button", { name: "Copy https://web.acme.stackdome.app" })).toBeInTheDocument();
    vi.useRealTimers();
  });

  it("renders nothing without endpoints", () => {
    const { container } = render(<PublicEndpointRow endpoints={[]} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("colors each dot from its own endpoint's variant, not a shared one", () => {
    const { container } = render(
      <PublicEndpointRow
        endpoints={[
          { ...endpoints[0], variant: "ready" as const },
          { ...endpoints[1], variant: "error" as const },
        ]}
      />,
    );
    expect(container.querySelectorAll(".bg-success")).toHaveLength(1);
    expect(container.querySelectorAll(".bg-danger")).toHaveLength(1);
  });

  it("falls back to the neutral dot when an endpoint has no variant", () => {
    const { container } = render(<PublicEndpointRow endpoints={[endpoints[0]]} />);
    expect(container.querySelectorAll(".bg-fg-muted")).toHaveLength(1);
  });
});

describe("PublicEndpointRow compact", () => {
  it("renders label-less go-to chips; URL only in tooltip, no copy button", () => {
    render(<PublicEndpointRow compact endpoints={endpoints} />);
    expect(screen.queryByText("PUBLIC")).toBeNull();
    // Hostname never renders inline; the tooltip (closed at rest) carries it.
    expect(screen.queryByText("web.acme.stackdome.app")).toBeNull();
    expect(screen.getByRole("link", { name: "Go to https://web.acme.stackdome.app" }))
      .toHaveAttribute("target", "_blank");
    expect(screen.queryByRole("button", { name: /Copy/ })).toBeNull();
  });

});
