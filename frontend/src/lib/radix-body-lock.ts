/** How long to keep watching after a close, covering a layer's exit animation. */
const LOCK_WATCH_MS = 1000;

const MODAL_LAYER = "[role=dialog],[role=alertdialog]";

/**
 * Enforce the invariant Radix's pointer-events save/restore breaks: with no
 * modal layer open, <body> must be interactive.
 *
 * A dialog opened while another one is still closing captures that one's
 * `pointer-events: none` as the value to restore, because the exit animation
 * outlives the tick an open can be deferred by (radix-ui/primitives#1836).
 * Handing that stale value back on close leaves the whole page dead with
 * nothing on screen to explain why, and only a reload recovers.
 *
 * Deferring harder cannot fix it — an exit animation has no duration this code
 * can know — so watch until the layers have actually settled, then assert the
 * invariant. Watching stops at the first settled frame, or after
 * LOCK_WATCH_MS if some layer stays open.
 */
export function releaseStaleBodyLock(): void {
  const started = Date.now();
  const check = () => {
    if (document.querySelector(MODAL_LAYER)) {
      if (Date.now() - started < LOCK_WATCH_MS) requestAnimationFrame(check);
      return;
    }
    if (document.body.style.pointerEvents === "none") {
      document.body.style.pointerEvents = "";
    }
  };
  requestAnimationFrame(check);
}
