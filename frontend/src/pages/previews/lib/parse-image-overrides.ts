/** "resource=registry/image:tag" lines → { resource: "registry/image:tag" };
    blank lines ignored, lines without "=" (or an empty key) are dropped. */

/** True when a single (already-trimmed, non-blank) line has the "key=value"
 *  shape this module accepts: a "=" with a non-empty key before it. Shared
 *  with form-schemas' per-line validation so the two rules can't drift —
 *  note this doesn't dedupe by key, so callers validating a whole block of
 *  text must check every line individually rather than comparing counts
 *  against parseImageOverrides' (key-deduped) output. */
export function isOverrideLine(line: string): boolean {
  return line.indexOf("=") > 0;
}

function toOverrideEntry(line: string): readonly [string, string] | null {
  if (!isOverrideLine(line)) return null;
  const idx = line.indexOf("=");
  return [line.slice(0, idx).trim(), line.slice(idx + 1).trim()] as const;
}

export function parseImageOverrides(text: string): Record<string, string> | undefined {
  const entries = text
    .split("\n")
    .map((l) => l.trim())
    .filter(Boolean)
    .map(toOverrideEntry)
    .filter((e): e is readonly [string, string] => e != null);
  return entries.length > 0 ? Object.fromEntries(entries) : undefined;
}
