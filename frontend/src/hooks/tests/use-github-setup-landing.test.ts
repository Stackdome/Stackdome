// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook } from "@testing-library/react";
import {
  useGithubSetupLanding,
  GITHUB_APP_INSTALLED_MESSAGE,
  GITHUB_APP_SETUP_ERROR_MESSAGE,
} from "../use-github-setup-landing";

describe("useGithubSetupLanding", () => {
  const originalOpener = window.opener;
  let closeSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    closeSpy = vi.spyOn(window, "close").mockImplementation(() => {});
  });

  afterEach(() => {
    Object.defineProperty(window, "opener", { value: originalOpener, writable: true });
    window.history.replaceState({}, "", "/");
    closeSpy.mockRestore();
  });

  it("posts to opener and closes when landing from GitHub setup", () => {
    const postMessage = vi.fn();
    Object.defineProperty(window, "opener", { value: { postMessage }, writable: true });
    window.history.replaceState({}, "", "/?installation_id=123&setup_action=install");

    renderHook(() => useGithubSetupLanding());

    expect(postMessage).toHaveBeenCalledWith(
      { type: GITHUB_APP_INSTALLED_MESSAGE },
      window.location.origin,
    );
    expect(closeSpy).toHaveBeenCalled();
  });

  it("relays a setup_error to the opener and closes", () => {
    const postMessage = vi.fn();
    Object.defineProperty(window, "opener", { value: { postMessage }, writable: true });
    window.history.replaceState({}, "", "/?setup_error=the%20link%20expired");

    renderHook(() => useGithubSetupLanding());

    expect(postMessage).toHaveBeenCalledWith(
      { type: GITHUB_APP_SETUP_ERROR_MESSAGE, reason: "the link expired" },
      window.location.origin,
    );
    expect(closeSpy).toHaveBeenCalled();
  });

  it("does nothing without setup_action param", () => {
    const postMessage = vi.fn();
    Object.defineProperty(window, "opener", { value: { postMessage }, writable: true });
    window.history.replaceState({}, "", "/");

    renderHook(() => useGithubSetupLanding());

    expect(postMessage).not.toHaveBeenCalled();
    expect(closeSpy).not.toHaveBeenCalled();
  });

  it("does nothing when there is no opener", () => {
    Object.defineProperty(window, "opener", { value: null, writable: true });
    window.history.replaceState({}, "", "/?setup_action=install");

    renderHook(() => useGithubSetupLanding());

    expect(closeSpy).not.toHaveBeenCalled();
  });
});
