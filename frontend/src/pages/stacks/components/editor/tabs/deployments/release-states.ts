import type { components } from "@/api/types/openapi";

/** The generated `StackReleaseState` is a types-only union; this is its runtime mirror. */
export type ReleaseStateValue = components["schemas"]["StackReleaseState"];

export const ReleaseState = {
  Pending: "Pending",
  InProgress: "InProgress",
  Released: "Released",
  Failed: "Failed",
  Superseded: "Superseded",
  Cancelled: "Cancelled",
} as const satisfies Record<string, ReleaseStateValue>;

/** Settled states — no further work will happen on the release. */
export const TERMINAL_STATES: ReadonlySet<string> = new Set([
  ReleaseState.Released,
  ReleaseState.Failed,
  ReleaseState.Superseded,
  ReleaseState.Cancelled,
]);

/** A release in one of these states is still being applied to the cluster. */
export const DEPLOYING_STATES: ReadonlySet<string> = new Set([
  ReleaseState.Pending,
  ReleaseState.InProgress,
]);

export function isTerminal(state?: string): boolean {
  return TERMINAL_STATES.has(state ?? "");
}

export function isDeploying(state?: string): boolean {
  return DEPLOYING_STATES.has(state ?? "");
}
