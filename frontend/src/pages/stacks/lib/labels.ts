/** Design rule: labels are lowercase, whitespace becomes dashes. */
export function normalizeLabel(raw: string): string {
  return raw.trim().toLowerCase().replace(/\s+/g, "-");
}
