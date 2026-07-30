// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup, act } from "@testing-library/react";
import { EndpointInlineList, PublicEndpointRow } from "../public-endpoint-row";

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

describe("PublicEndpointRow multi-url overflow", () => {
  const multi = {
    service: "web",
    url: "https://web.acme.stackdome.app",
    port: 80,
    urls: [
      { url: "https://web.acme.stackdome.app", target_port: 80 },
      { url: "https://mqrkc2xw4t7b4dnz.web.acme.stackdome.app", target_port: 89 },
    ],
  };

  it("shows a +N tail only when an endpoint has more than one url", () => {
    render(<PublicEndpointRow endpoints={[multi, endpoints[1]]} />);
    expect(screen.getByRole("button", { name: "1 more endpoint for web" })).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: /more endpoint/ })).toHaveLength(1);
  });

  it("opens a popover listing only the OTHER endpoints — +N matches the row count", async () => {
    render(<PublicEndpointRow endpoints={[multi]} />);
    fireEvent.click(screen.getByRole("button", { name: "1 more endpoint for web" }));
    // Only the second endpoint is listed (the best one lives on the chip);
    // the PORT column shows the bare number.
    expect(await screen.findByText("89")).toBeInTheDocument();
    expect(screen.queryByText("80")).toBeNull();
    const rowLink = screen.getByRole("link", { name: "Go to https://mqrkc2xw4t7b4dnz.web.acme.stackdome.app" });
    expect(rowLink).toHaveAttribute("target", "_blank");
    expect(
      screen.getByRole("button", { name: "Copy https://mqrkc2xw4t7b4dnz.web.acme.stackdome.app" }),
    ).toBeInTheDocument();
  });

  it("single-url endpoints render exactly as before (no tail)", () => {
    render(<PublicEndpointRow endpoints={[endpoints[0]]} />);
    expect(screen.queryByRole("button", { name: /more endpoint/ })).toBeNull();
  });
});

describe("EndpointInlineList (drawer header)", () => {
  const urls = [
    { url: "https://web.acme.stackdome.app", target_port: 80 },
    { url: "https://mqrkc2xw4t7b4dnz.web.acme.stackdome.app", target_port: 89 },
  ];

  it("renders the first url plus the same '+N' popover trigger the chips use", () => {
    render(<EndpointInlineList service="web" urls={urls} />);
    expect(screen.getByText("web.acme.stackdome.app")).toBeInTheDocument();
    expect(screen.queryByText(/mqrkc2xw4t7b4dnz/)).toBeNull();
    expect(screen.getByRole("button", { name: "1 more endpoint for web" })).toBeInTheDocument();
  });

  it("popover lists every endpoint", async () => {
    render(<EndpointInlineList service="web" urls={urls} />);
    fireEvent.click(screen.getByRole("button", { name: "1 more endpoint for web" }));
    expect(
      await screen.findByRole("link", { name: "Go to https://mqrkc2xw4t7b4dnz.web.acme.stackdome.app" }),
    ).toBeInTheDocument();
  });

  it("single url renders inline with no trigger; empty renders nothing", () => {
    render(<EndpointInlineList service="web" urls={[urls[0]]} />);
    expect(screen.queryByRole("button", { name: /more endpoint/ })).toBeNull();
    const { container } = render(<EndpointInlineList service="web" urls={[]} />);
    expect(container).toBeEmptyDOMElement();
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
