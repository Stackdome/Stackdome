import type { ScopeList } from "@/api/api-tokens";

// Actions that only observe. Anything outside this set (write, delete, exec,
// create) makes a token capable of change, so read-only must not grant it.
const READ_ONLY_ACTIONS = new Set(["read", "list"]);

export const READ_ONLY = "read-only";
export const FULL_ACCESS = "full-access";

export function readOnlyScopes(scopes: ScopeList): string[] {
  return (scopes.items ?? []).flatMap((entry) =>
    entry.resource
      ? (entry.actions ?? [])
        .filter((action) => READ_ONLY_ACTIONS.has(action))
        .map((action) => `${entry.resource}:${action}`)
      : [],
  );
}

// Hand-scoped tokens from the API also read as "Full access"; the row title
// carries the exact scopes.
export function accessLabel(scopes?: string[]): string {
  if (!scopes?.length) return "—";
  const readOnly = scopes.every((scope) => READ_ONLY_ACTIONS.has(scope.split(":")[1]));
  return readOnly ? "Read-only" : "Full access";
}
