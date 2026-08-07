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

// ponytail: the UI mints only these two levels, so "not all read/list" means
// full access. A token scoped by hand through the API would read as Full
// access too — the row's tooltip carries the exact scopes either way.
export function accessLabel(scopes?: string[]): string {
  if (!scopes?.length) return "—";
  const readOnly = scopes.every((scope) => READ_ONLY_ACTIONS.has(scope.split(":")[1]));
  return readOnly ? "Read-only" : "Full access";
}
