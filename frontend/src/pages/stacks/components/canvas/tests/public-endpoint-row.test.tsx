// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup, act } from "@testing-library/react";
import { PublicEndpointRow } from "../PublicEndpointRow";

afterEach(cleanup);

const endpoints = [
  { service: "web", url: "https://web.acme.stackdome.app", port: 80 },
  { service: "api", url: "https://api.mycompany.com", port: 8080 },
];

describe("PublicEndpointRow", () => {
  beforeEach(() => {
    Object.assign(navigator, { clipboard: { writeText: vi.fn().mockResolvedValue(undefined) } });
  });

  it("renders one pill per endpoint with service chip and host", () => {
    render(<PublicEndpointRow endpoints={endpoints} />);
    expect(screen.getByText("PUBLIC")).toBeInTheDocument();
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("web.acme.stackdome.app")).toBeInTheDocument();
    expect(screen.getByText("api.mycompany.com")).toBeInTheDocument();
  });

  it("link opens in a new tab", () => {
    render(<PublicEndpointRow endpoints={endpoints} />);
    const link = screen.getByRole("link", { name: /web\.acme\.stackdome\.app/ });
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
});
