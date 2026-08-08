import {
  CircleCheck,
  CircleDashed,
  CircleOff,
  CircleX,
  Loader2,
  TriangleAlert,
  Trash2,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { statusVariant, type StatusDomain, type StatusVariant } from "./status-variant";

/**
 * Status, said once, as a word (§4, §7).
 *
 * **The colour is derived, never passed.** This component takes no `children`,
 * no `variant`, no `className` and no `style` — the only inputs are the domain
 * and the state string the backend actually sent, and both the word and the
 * colour fall out of them. That is the whole point of it.
 *
 * The Figma critique found the defect this replaces: `StatusPill` takes a
 * `variant` and its `children` separately, so nothing stops
 * `<StatusPill variant="error">Healthy</StatusPill>` — a red row that reads
 * "Healthy". A prop pair cannot be made safe by convention; it has to stop
 * being a pair. Callers that need a coloured word use this. `StatusPill`
 * remains for the bordered chip, where the word is genuinely fixed by the
 * caller.
 *
 * No dot. §7: a coloured dot at the left of a row AND the status word in a
 * column is saying it twice. The dot survives only where there is no room for
 * a word, which is not here.
 */
const TONE: Record<StatusVariant, string> = {
  ready: "text-success",
  pending: "text-warn",
  error: "text-danger",
  info: "text-info",
  neutral: "text-fg-muted",
};

/**
 * One glyph per **state** — and, like the colour, it is *derived* from domain +
 * state. It cannot be passed in, so an icon that disagrees with the word it
 * sits next to is as unbuildable as a colour that disagrees.
 *
 * **Why per state and not per family.** The first version had three glyphs for
 * three families, which meant `Degraded`, `Unavailable` and `Failed` all drew
 * the same triangle — and that is precisely the distinction that changes what
 * you do next. A glyph that cannot make the distinction the word makes is pure
 * chrome, which is why both call sites had switched icons off. Each state now
 * gets its own mark:
 *
 * | State | Glyph | Reads as |
 * |---|---|---|
 * | Healthy | `CircleCheck` | serving, nothing to do |
 * | Deploying | `Loader2`, spinning | in flight, wait |
 * | Degraded | `TriangleAlert` | serving, but not fully |
 * | Unavailable | `CircleOff` | not serving at all |
 * | Failed | `CircleX` | the deploy did not land |
 * | NotDeployed | `CircleDashed` | never run — a fact, not a fault |
 * | Deleting | `Trash2` | on its way out |
 *
 * Outline glyphs at the same stroke weight as every other icon in the product;
 * nothing here is filled. Colour still comes from the variant, so the glyph and
 * the word can never disagree about severity.
 */
const GLYPH: Partial<Record<StatusDomain, Record<string, LucideIcon>>> = {
  stack_rollup: {
    healthy: CircleCheck,
    deploying: Loader2,
    degraded: TriangleAlert,
    unavailable: CircleOff,
    failed: CircleX,
    notdeployed: CircleDashed,
    deleting: Trash2,
  },
};

/** States whose glyph turns. `motion-safe:` so reduced-motion gets a still mark
 *  rather than no mark — the shape still reports "in flight". */
const SPINS = new Set(["deploying"]);

/**
 * The word the backend sent, made readable — `InProgress` → `In progress`,
 * `image_pull_failed` → `Image pull failed`.
 *
 * The raw state is shown rather than a bucket label because the buckets are
 * lossy: `Degraded` and `Failed` are both `error`, and which one you are
 * looking at changes what you do next.
 */
function humanise(state: string): string {
  const words = state
    .trim()
    .replace(/[_-]+/g, " ")
    .replace(/([a-z0-9])([A-Z])/g, "$1 $2")
    .toLowerCase()
    .split(/\s+/)
    .filter(Boolean);
  if (!words.length) return "";
  return words[0].charAt(0).toUpperCase() + words[0].slice(1) + (words.length > 1 ? " " + words.slice(1).join(" ") : "");
}

export function StatusText({
  domain,
  state,
  /** Layout only — `text-meta` is the §3 size for row data. */
  size = "meta",
  /** Prefix the word with the glyph for its STATE, 6px apart. Still no way to
   *  choose which glyph — see `GLYPH`. A domain with no map renders the word
   *  alone rather than a stand-in. */
  icon = false,
}: {
  domain: StatusDomain;
  state?: string | null;
  size?: "meta" | "body";
  icon?: boolean;
}) {
  const variant = statusVariant(domain, state);
  const word = humanise(state ?? "") || "Unknown";
  const key = (state ?? "").trim().toLowerCase();
  const Glyph = icon ? GLYPH[domain]?.[key] : undefined;

  return (
    <span
      data-slot="status-text"
      data-status-variant={variant}
      className={cn(
        "font-medium whitespace-nowrap",
        size === "meta" ? "text-meta" : "text-body",
        TONE[variant],
      )}
    >
      {/* INLINE, never inline-flex. The card sets the name and the status word
          on a shared baseline (§8); an inline-flex box takes its baseline from
          its first flex item, so the glyph would drag the word off that line.
          An inline-block glyph with an optical nudge keeps the text baseline
          exactly where it was and centres the mark on the x-height. */}
      {Glyph && (
        <Glyph
          className={cn(
            "mr-1.5 inline-block align-[-0.22em]",
            size === "meta" ? "size-3.5" : "size-4",
            SPINS.has(key) && "motion-safe:animate-spin",
          )}
          aria-hidden
        />
      )}
      {word}
    </span>
  );
}
