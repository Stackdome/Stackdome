/**
 * Structural equality for stack data.
 *
 * `undefined`, `""`, `[]`, `{}` and objects whose values are all themselves
 * empty are one value: "nothing here". The form and the API disagree constantly
 * about which of those they use for an absent field, and none of those
 * disagreements is a change the user made.
 */
export function isStructurallyEmpty(v: unknown): boolean {
  if (v === null || v === undefined) return true;
  if (typeof v === "string") return v === "";
  if (Array.isArray(v)) return v.every(isStructurallyEmpty);
  if (typeof v === "object") {
    return Object.values(v as Record<string, unknown>).every(isStructurallyEmpty);
  }
  return false;
}

export function deepEqual(a: unknown, b: unknown): boolean {
  if (a === b) return true;
  if (isStructurallyEmpty(a) && isStructurallyEmpty(b)) return true;
  if (a === null || b === null) return a === b;
  if (typeof a !== typeof b) return false;
  if (typeof a !== "object") return false;

  if (Array.isArray(a) || Array.isArray(b)) {
    if (!Array.isArray(a) || !Array.isArray(b)) return false;
    if (a.length !== b.length) return false;
    return a.every((item, i) => deepEqual(item, (b as unknown[])[i]));
  }

  const ao = a as Record<string, unknown>;
  const bo = b as Record<string, unknown>;
  const aKeys = Object.keys(ao).filter((k) => !isStructurallyEmpty(ao[k]));
  const bKeys = Object.keys(bo).filter((k) => !isStructurallyEmpty(bo[k]));
  if (aKeys.length !== bKeys.length) return false;
  return aKeys.every((k) => deepEqual(ao[k], bo[k]));
}

/**
 * Greedily pair entries from two lists whose fingerprints match; each entry is
 * used at most once. A removed entry and an added entry with the same content
 * are one renamed entity, not two changes.
 */
export function pairByFingerprint<A, B>(
  as: A[],
  bs: B[],
  fpA: (a: A) => string,
  fpB: (b: B) => string,
): Array<[A, B]> {
  const pairs: Array<[A, B]> = [];
  const usedB = new Set<number>();
  for (const a of as) {
    const fp = fpA(a);
    const idx = bs.findIndex((b, i) => !usedB.has(i) && fpB(b) === fp);
    if (idx >= 0) {
      usedB.add(idx);
      pairs.push([a, bs[idx]]);
    }
  }
  return pairs;
}
