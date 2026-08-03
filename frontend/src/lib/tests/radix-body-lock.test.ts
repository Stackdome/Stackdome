// @vitest-environment jsdom
import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { releaseStaleBodyLock } from "../radix-body-lock";

/**
 * The lock is released only once the layers have settled, because a layer that
 * is still animating out is still entitled to hold it.
 *
 * Real timers: the watcher polls with requestAnimationFrame, and a watcher
 * still running from an earlier case survives vitest's fake-timer reset — it
 * then observes this case's DOM and reports on it. Waiting a few real frames
 * keeps each case reading only its own watcher.
 */
const frames = (n = 1) => new Promise((r) => setTimeout(r, n * 32));

beforeEach(() => {
  document.body.innerHTML = "";
  document.body.style.pointerEvents = "";
});

afterEach(() => {
  document.body.innerHTML = "";
  document.body.style.pointerEvents = "";
});

describe("releaseStaleBodyLock", () => {
  it("clears a lock left behind when no layer is open", async () => {
    document.body.style.pointerEvents = "none";
    releaseStaleBodyLock();
    await frames();
    expect(document.body.style.pointerEvents).toBe("");
  });

  it("leaves the lock alone while a layer is still open", async () => {
    document.body.innerHTML = '<div role="dialog"></div>';
    document.body.style.pointerEvents = "none";
    releaseStaleBodyLock();
    await frames(3);
    expect(document.body.style.pointerEvents).toBe("none");
  });

  /** The real sequence: the confirm settles, and only afterwards does the
   *  closing layer unmount and hand back its stale value. */
  it("clears a lock restored after the watch began, once the layer unmounts", async () => {
    document.body.innerHTML = '<div role="alertdialog"></div>';
    releaseStaleBodyLock();
    await frames(2);

    document.body.innerHTML = "";
    document.body.style.pointerEvents = "none";
    await frames(2);

    expect(document.body.style.pointerEvents).toBe("");
  });

  it("does not disturb an unlocked body", async () => {
    document.body.style.pointerEvents = "auto";
    releaseStaleBodyLock();
    await frames();
    expect(document.body.style.pointerEvents).toBe("auto");
  });
});
