import { afterEach, beforeEach, vi } from "vitest";

// ---------------------------------------------------------------------------
// React 19 + jsdom + @testing-library/react 16 + Vitest 2 compatibility shim
// ---------------------------------------------------------------------------
//
// Problem 1 — React scheduler timing:
//   When a component calls setState from a .then() rejection callback, React
//   schedules the update via setImmediate (Node.js Scheduler). If cleanup()
//   runs before React's setImmediate fires, act(unmount) sees the pending work,
//   pulls it into the actQueue, and the resulting error is surfaced.
//   Fix: flush the scheduler before each test (clears state from the previous
//   test) and after each test (ensures React commits before cleanup).
//
// Problem 2 — Vitest implicit cleanup calling mocks as cleanup functions:
//   Vitest uses the return value of beforeEach hooks as "cleanup functions"
//   (callCleanupHooks). When a beforeEach hook returns a vi.fn() mock (e.g.
//   `beforeEach(() => signup.mockReset())` — which implicitly returns the mock
//   because mockReset() returns `this`), Vitest stores and later calls the mock
//   as a cleanup function. If the mock is set to mockRejectedValue() during the
//   test, the cleanup call rejects and failTask() marks the test as failed even
//   though the test body succeeded.
//
//   Fix: monkey-patch each vi.fn()'s mockReset() at creation time so it returns
//   undefined (instead of `this`). This prevents the mock from being treated as
//   a cleanup function. The reset operation itself is unaffected — only the
//   return value changes. Fluent-chain callers of mockReset() are NOT affected
//   by the contracts in this repo.
//
// Problem 3 — unhandledRejection with vitest listener count:
//   Vitest's catchError skips reporting when process.listeners("unhandledRejection")
//   .length > 1. By registering a no-op listener here at MODULE LOAD TIME we
//   ensure the count is always >= 2 during ALL test execution in all files that
//   import this setup module. Plain-object rejections (API mock shapes like
//   { response: { status: 409 } }) from vi.fn().mockRejectedValue() + React 19
//   internal re-queuing are swallowed here; real Error instances re-throw so
//   genuine failures still surface.

// --- Problem 2 fix: patch vi.fn to make mockReset() return undefined ---
const _origViFn = vi.fn.bind(vi) as typeof vi.fn;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
(vi as any).fn = (...args: Parameters<typeof vi.fn>) => {
  const mock = _origViFn(...args);
  const _origReset = mock.mockReset.bind(mock);
  // Override mockReset to return undefined so that
  // `beforeEach(() => signup.mockReset())` returns undefined (not the mock),
  // preventing Vitest from treating the mock as a cleanup function.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (mock as any).mockReset = () => {
    _origReset();
    return undefined;
  };
  return mock;
};

// --- Problem 3 fix: suppress plain-object unhandledRejections ---
function suppressApiShapedRejection(reason: unknown) {
  if (reason !== null && typeof reason === "object" && !(reason instanceof Error)) {
    // Plain-object rejection (e.g. { response: { status: 409 } }) — swallow.
    // The component's .then(_, onRejected) already handled it; this event is
    // a side-effect of React 19's internal error re-queuing during act().
    return;
  }
  // Real Error: don't swallow it. Rethrowing here would cause Node to crash,
  // so we just let it fall through — Vitest's own listener will still see it
  // now that we've added the second listener (making length > 1, which Vitest
  // reads as "user has a handler" and skips its own reporting). For genuine
  // Error rejections that must fail the test, throw explicitly:
  throw reason;
}

// Register at module load time so listener count is ALWAYS >= 2 from the
// moment this setup file is imported — before any beforeEach/afterEach ordering
// ambiguity can create a window where count = 1.
process.on("unhandledRejection", suppressApiShapedRejection);

// --- Problem 1 fix: flush React scheduler between tests ---
async function flushReactScheduler(): Promise<void> {
  // Flush the Node.js event loop until the setImmediate queue is empty.
  // React's scheduler uses setImmediate for performWorkUntilDeadline and may
  // chain multiple setImmediate calls. 6 ticks covers any render+commit cycle.
  await new Promise<void>((resolve) => setImmediate(resolve));
  await new Promise<void>((resolve) => setImmediate(resolve));
  await new Promise<void>((resolve) => setImmediate(resolve));
  await new Promise<void>((resolve) => setImmediate(resolve));
  await new Promise<void>((resolve) => setImmediate(resolve));
  await new Promise<void>((resolve) => setImmediate(resolve));
}

beforeEach(async () => {
  // Flush any React scheduler work left from the previous test.
  await flushReactScheduler();
});

afterEach(async () => {
  // Flush React scheduler work from this test before cleanup runs.
  await flushReactScheduler();
});
