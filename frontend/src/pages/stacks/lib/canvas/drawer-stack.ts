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

/**
 * Rebind open drawers onto another resource/volume list by name (names are
 * unique in a stack), remapping resource indexes — the two lists don't line
 * up. Entries with no same-named counterpart drop out.
 */
export function remapStackByName(
  stack: DrawerEntry[],
  from: { resources: { name?: string }[] },
  to: { resources: { name?: string }[]; volumeNames: ReadonlySet<string> },
): DrawerEntry[] {
  return stack.flatMap((e): DrawerEntry[] => {
    if (e.kind === "volume") return to.volumeNames.has(e.name) ? [e] : [];
    const name = from.resources[e.index]?.name;
    const index = name ? to.resources.findIndex((r) => r.name === name) : -1;
    return index >= 0 ? [{ kind: "resource", index }] : [];
  });
}
