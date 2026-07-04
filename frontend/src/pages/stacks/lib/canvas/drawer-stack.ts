/** One open panel in the floating drawer stack. */
export type DrawerEntry =
  | { kind: "resource"; index: number }
  | { kind: "volume"; name: string };

export function entryKey(e: DrawerEntry): string {
  return e.kind === "resource" ? `resource:${e.index}` : `volume:${e.name}`;
}

export function replaceStack(entry: DrawerEntry): DrawerEntry[] {
  return [entry];
}

/** Push, or truncate back to the entry if it is already open (no duplicates). */
export function pushEntry(stack: DrawerEntry[], entry: DrawerEntry): DrawerEntry[] {
  const existing = stack.findIndex((e) => entryKey(e) === entryKey(entry));
  return existing >= 0 ? stack.slice(0, existing + 1) : [...stack, entry];
}

export function truncateTo(stack: DrawerEntry[], depth: number): DrawerEntry[] {
  return stack.slice(0, depth + 1);
}

export function popEntry(stack: DrawerEntry[]): DrawerEntry[] {
  return stack.slice(0, -1);
}
