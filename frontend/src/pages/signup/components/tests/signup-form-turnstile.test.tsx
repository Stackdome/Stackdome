// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type {
  TurnstileAPI,
  TurnstileRenderOptions,
} from "@/components/auth/turnstile-widget";
import { SignupForm } from "@/pages/signup/components/signup-form";

const mocks = vi.hoisted(() => ({
  signup: vi.fn(),
  signupConfig: vi.fn(),
  navigate: vi.fn(),
}));

vi.mock("@/pages/signup/hooks/use-signup", () => ({
  useSignup: () => ({ signup: mocks.signup }),
}));
vi.mock("@/hooks/use-signup-config", () => ({
  useSignupConfig: () => mocks.signupConfig(),
}));
vi.mock("@/components/auth/github-sign-in-button", () => ({
  GitHubSignInButton: () => null,
}));
vi.mock("react-router-dom", async (importOriginal) => ({
  ...(await importOriginal<typeof import("react-router-dom")>()),
  useNavigate: () => mocks.navigate,
}));

describe("SignupForm Turnstile integration", () => {
  let renderOptions: TurnstileRenderOptions | undefined;
  let reset: ReturnType<typeof vi.fn>;
  let remove: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mocks.signup.mockReset();
    mocks.signupConfig.mockReset();
    mocks.navigate.mockReset();
    reset = vi.fn();
    remove = vi.fn();
    const turnstile: TurnstileAPI = {
      render: vi.fn((_container, options) => {
        renderOptions = options;
        return "widget-id";
      }),
      reset,
      remove,
    };
    window.turnstile = turnstile;
  });

  afterEach(() => {
    cleanup();
    delete window.turnstile;
  });

  it("does not render or submit Turnstile when disabled", async () => {
    mocks.signupConfig.mockReturnValue({
      turnstile: { enabled: false, site_key: "", action: "" },
      loading: false,
      error: null,
    });
    mocks.signup.mockResolvedValue({});
    const user = userEvent.setup();
    render(<SignupForm />);

    await fillSignupForm(user);
    await user.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() => expect(mocks.signup).toHaveBeenCalledTimes(1));
    expect(mocks.signup.mock.calls[0][0]).not.toHaveProperty("turnstile_token");
    expect(renderOptions).toBeUndefined();
  });

  it("submits the token returned by an enabled widget", async () => {
    enableTurnstileConfig();
    mocks.signup.mockResolvedValue({});
    const user = userEvent.setup();
    render(<SignupForm />);

    await waitFor(() => expect(renderOptions).toBeDefined());
    act(() => renderOptions?.callback("browser-token"));
    await fillSignupForm(user);
    await user.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() => expect(mocks.signup).toHaveBeenCalledTimes(1));
    expect(mocks.signup.mock.calls[0][0]).toMatchObject({
      email: "person@example.com",
      turnstile_token: "browser-token",
    });
    expect(renderOptions?.sitekey).toBe("public-site-key");
    expect(renderOptions?.action).toBe("turnstile-spin-v2");
    expect(renderOptions?.appearance).toBe("interaction-only");
  });

  it("resets the widget after the signup request fails", async () => {
    enableTurnstileConfig();
    mocks.signup.mockRejectedValue(new Error("rejected"));
    const user = userEvent.setup();
    render(<SignupForm />);

    await waitFor(() => expect(renderOptions).toBeDefined());
    act(() => renderOptions?.callback("browser-token"));
    await fillSignupForm(user);
    await user.click(screen.getByRole("button", { name: /create account/i }));

    await waitFor(() => expect(reset).toHaveBeenCalledWith("widget-id"));
  });

  function enableTurnstileConfig() {
    mocks.signupConfig.mockReturnValue({
      turnstile: {
        enabled: true,
        site_key: "public-site-key",
        action: "turnstile-spin-v2",
      },
      loading: false,
      error: null,
    });
  }
});

async function fillSignupForm(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText("Full name"), "Person");
  await user.type(screen.getByLabelText("Organization"), "Example");
  await user.type(screen.getByLabelText("Email"), "person@example.com");
  await user.type(screen.getByLabelText(/^Password/), "password123");
  await user.type(screen.getByLabelText(/^Confirm password/), "password123");
}
