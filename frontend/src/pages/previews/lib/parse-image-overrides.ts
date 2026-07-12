/** "resource=registry/image:tag" lines → { resource: "registry/image:tag" };
    blank lines ignored, lines without "=" (or an empty key) are dropped. */
export function parseImageOverrides(text: string): Record<string, string> | undefined {
  const entries = text
    .split("\n")
    .map((l) => l.trim())
    .filter(Boolean)
    .map((l) => {
      const idx = l.indexOf("=");
      return idx > 0 ? ([l.slice(0, idx).trim(), l.slice(idx + 1).trim()] as const) : null;
    })
    .filter((e): e is readonly [string, string] => e != null);
  return entries.length > 0 ? Object.fromEntries(entries) : undefined;
}
