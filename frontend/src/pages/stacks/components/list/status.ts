import { deriveHeaderHealth, latestDeployFailed, type ReleaseHealth } from "@/pages/stacks/components/editor/tabs/deployments/derive";
import { ReleaseState, isTerminal } from "@/pages/stacks/components/editor/tabs/deployments/release-states";
import type { components } from "@/api/types/openapi";
import type { Stack } from "@/api/stack-types";

type ReleaseSummary = components["schemas"]["ReleaseSummary"];

/** The rollup vocabulary — closed, and the only words the list ever shows.
 *  `StatusText` humanises these, so `NotDeployed` renders "Not deployed". */
export type StackRollupState =
  | "Deleting"
  | "NotDeployed"
  | "Healthy"
  | "Deploying"
  | "Degraded"
  | "Unavailable"
  | "Failed";

const HEALTH_STATE: Record<ReleaseHealth, StackRollupState> = {
  ok: "Healthy",
  progressing: "Deploying",
  degraded: "Degraded",
  unavailable: "Unavailable",
  failed: "Failed",
};

/**
 * One word for a stack, rolled up from lifecycle and release health.
 *
 * Returns the *state*, not a colour and not a label — `StatusText` derives both
 * from it, which is what makes a red row that reads "Healthy" unbuildable.
 *
 * "Deleting" beats health: a stack on its way out is not usefully described by
 * how healthy its last release was. No derivable health at all — never
 * deployed, or nothing but cancelled and superseded attempts — is
 * "NotDeployed", which is a fact rather than a failure and colours neutral.
 */
export function stackRollupState(stack: Stack): StackRollupState {
  if (stack.lifecycle === "deleting") return "Deleting";
  const health = deriveHeaderHealth(stack);
  if (!health) return "NotDeployed";
  return HEALTH_STATE[health];
}

/**
 * Stacks needing a human — the header fact and the default sort.
 *
 * Deleting is in flight, not in trouble; NotDeployed has simply not run yet.
 *
 * A stack whose live release is healthy but whose LATEST deploy attempt failed
 * counts too. The rollup word cannot report it — the stack really is serving,
 * so calling it "Failed" would be a lie — but somebody pushed something and it
 * did not land, which is exactly what this list exists to surface.
 */
export function needsAttention(stack: Stack): boolean {
  const state = stackRollupState(stack);
  if (state === "Deleting") return false;
  if (state === "Degraded" || state === "Unavailable" || state === "Failed") return true;
  return latestDeployFailed(stack);
}

/** True when the stack is live and healthy but its newest deploy attempt did
 *  not land — a second fact, not a second reading of the status word. */
export function lastDeployFailed(stack: Stack): boolean {
  return stackRollupState(stack) !== "Failed" && latestDeployFailed(stack);
}

/**
 * The release the rollup WORD came from — so the reason line and the word can
 * never describe two different releases.
 *
 * Deliberately mirrors `deriveHeaderHealth`'s branching. If that function's
 * rules change, this one has to change with it.
 */
function rollupRelease(stack: Stack): ReleaseSummary | undefined {
  const latest = stack.latest_release;
  if (!latest) return undefined;
  if (!isTerminal(latest.state)) return latest;
  if (stack.converged_release?.health) return stack.converged_release;
  return latest.state === ReleaseState.Failed ? latest : undefined;
}

export interface StatusReason {
  text: string;
  /** `danger` is a second, worse fact; `muted` elaborates the word above it. */
  tone: "danger" | "muted";
}

/**
 * Why the stack is in the state it is in — the line under the status word.
 *
 * **It returns null whenever nothing is wrong**, and that is the mechanic, not
 * an optimisation: a healthy row stays one line and a broken one grows to two,
 * so trouble is found by the SHAPE of the list before a word is read or a
 * colour registers. It therefore also works for anyone who cannot see the red.
 *
 * **The gate is `needsAttention`, and that is what makes the guarantee hold.**
 * It is not left to what the backend happens to write: `MarkCancelled` and
 * `MarkSuperseded` write messages too, so a stack that is serving perfectly
 * well could otherwise grow a second line and break the shape. Gating on the
 * predicate means the header's *"N need attention"*, the default sort order and
 * the set of two-line rows are all **the same set**, computed once.
 */
export function statusReason(stack: Stack): StatusReason | null {
  if (!needsAttention(stack)) return null;

  // A stack that is serving while its newest push did not land is the more
  // actionable fact, and it outranks whatever the live release last said.
  if (lastDeployFailed(stack)) return { text: "Last deploy failed", tone: "danger" };

  const message = rollupRelease(stack)?.message?.trim();
  return message ? { text: message, tone: "muted" } : null;
}

/**
 * When the stack entered its CURRENT state — not when the record was last
 * touched. `Failed 20m ago` and `Failed 6 days ago` are different situations:
 * one is a deploy going wrong right now, the other is something everyone has
 * stopped looking at. The old `updated_at` column could not tell them apart.
 */
export function stateChangedAt(stack: Stack): string | undefined {
  const release = rollupRelease(stack);
  return release?.completed_at ?? release?.created_at ?? stack.updated_at ?? stack.created_at;
}
