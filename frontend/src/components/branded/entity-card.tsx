import { Fragment, type ReactNode } from "react";
import { formatDistanceToNowStrict } from "date-fns";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { type StatusTone } from "@/components/branded/status-variant";
import { cn } from "@/lib/utils";

/**
 * The parts a resource card is built from — the deploy rail, the status word,
 * the endpoint pills, the meta grid and footer. Stacks and preview environments
 * are different things that read the same way, so the pieces live here rather
 * than in whichever page happened to need them first.
 */

export type RailTone = StatusTone;

/**
 * 4px full-bleed rail across the card's top edge, shown only while a deploy
 * is in flight: a soft track with an animated amber segment (static partial
 * fill under prefers-reduced-motion). Settled states render no rail — the
 * colored status word carries the state.
 */
export function StatusRail({ tone }: { tone: RailTone }) {
  if (tone !== "deploying") return null;
  return (
    <div className="h-1 w-full overflow-hidden bg-warn-bg" role="presentation" data-rail="deploying">
      <div className="h-full w-[34%] bg-brand animate-rail-sweep" />
    </div>
  );
}

function toneTextClass(tone: RailTone | "neutral"): string {
  if (tone === "success") return "text-success";
  if (tone === "brand" || tone === "deploying") return "text-brand";
  if (tone === "danger") return "text-danger";
  return "text-fg-muted";
}

/** Mono status word, colored to match the rail. */
export function StatusWord({ tone, children }: { tone: RailTone | "neutral"; children: string }) {
  return (
    <span
      className={cn(
        "font-mono text-[10.5px] font-medium uppercase tracking-[1px] whitespace-nowrap",
        toneTextClass(tone),
      )}
    >
      {children}
    </span>
  );
}

export interface EndpointUrl {
  resource?: string;
  url?: string;
}

/** Pills stay on one row inside the fixed-height card; past this count the
 *  rest collapse into a "+N" popover. */
const MAX_VISIBLE_PILLS = 2;

const pillClass =
  "inline-flex items-center gap-1.5 rounded-sm border border-border px-2.5 py-1 font-mono text-xs text-fg-2 transition-colors duration-120 hover:text-brand hover:border-brand focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-brand/40";

function PillLink({ url }: { url: EndpointUrl }) {
  return (
    <a
      href={url.url}
      target="_blank"
      rel="noopener noreferrer"
      onClick={(e) => e.stopPropagation()}
      className={cn(pillClass, "min-w-0")}
    >
      <span className="truncate">{url.resource ?? "link"}</span>
      <span className="flex-none opacity-55">↗</span>
    </a>
  );
}

/** External-link pills for a card's public endpoints. */
export function EndpointPills({ urls }: { urls: EndpointUrl[] }) {
  const valid = urls.filter((u) => u.url);
  if (valid.length === 0) return null;
  const visible = valid.slice(0, MAX_VISIBLE_PILLS);
  const overflow = valid.slice(MAX_VISIBLE_PILLS);
  return (
    <div className="flex gap-1.5">
      {visible.map((u, i) => (
        <PillLink key={`${u.resource}-${u.url}-${i}`} url={u} />
      ))}
      {overflow.length > 0 && (
        <Popover>
          <PopoverTrigger asChild>
            <button
              type="button"
              aria-label={`${overflow.length} more endpoints`}
              onClick={(e) => e.stopPropagation()}
              className={cn(pillClass, "flex-none tabular-nums")}
            >
              +{overflow.length}
            </button>
          </PopoverTrigger>
          <PopoverContent
            align="end"
            className="flex w-[220px] flex-col gap-1 p-1.5"
            onClick={(e) => e.stopPropagation()}
          >
            {overflow.map((u, i) => (
              <PillLink key={`${u.resource}-${u.url}-${i}`} url={u} />
            ))}
          </PopoverContent>
        </Popover>
      )}
    </div>
  );
}

export interface CardMetaRow {
  label: string;
  value: ReactNode;
}

/** Label/value mono grid — the card grammar's META slot. */
export function CardMetaGrid({ rows }: { rows: Array<CardMetaRow | false | null | undefined> }) {
  const present = rows.filter((r): r is CardMetaRow => Boolean(r));
  if (present.length === 0) return null;
  return (
    <div className="grid grid-cols-[auto_minmax(0,1fr)] items-baseline gap-x-3 gap-y-1.5">
      {present.map((r) => (
        <Fragment key={r.label}>
          <span className="font-mono text-[9px] uppercase tracking-[1.2px] text-fg-muted">{r.label}</span>
          <span className="truncate font-mono text-[11px] text-fg-2">{r.value}</span>
        </Fragment>
      ))}
    </div>
  );
}

interface CardFooterMetaProps {
  /** Status tone driving the dot + word color. */
  tone: RailTone | "neutral";
  /** Status word at the left edge. */
  word: string;
  /** Relative age, uppercase at the right edge. */
  age?: string | null;
  /** Absolute timestamp revealed as a tooltip on the age. */
  ageTitle?: string | null;
  /** Optional alert glyph rendered beside the status word. */
  alert?: ReactNode;
}

/** Footer row — the card grammar's FOOTER slot: ● STATUS left, AGE right. */
export function CardFooterMeta({ tone, word, age, ageTitle, alert }: CardFooterMetaProps) {
  return (
    <div className="flex items-center justify-between gap-3 border-t border-border/60 pt-4 whitespace-nowrap">
      <span className={cn("inline-flex flex-none items-center gap-2", toneTextClass(tone))}>
        <span aria-hidden className="h-1.5 w-1.5 flex-none rounded-full bg-current" />
        <StatusWord tone={tone}>{word}</StatusWord>
        {alert}
      </span>
      {age && (
        <span
          title={ageTitle ?? undefined}
          className="flex-none font-mono text-[10.5px] uppercase tracking-[0.5px] text-fg-muted"
        >
          {age}
        </span>
      )}
    </div>
  );
}

const AGE_UNIT_ABBREV: Record<string, string> = {
  second: "s",
  minute: "m",
  hour: "h",
  day: "d",
  month: "mo",
  year: "y",
};

/** Compact relative age for card footers: "just now", "5m ago", "3d ago". */
export function relativeAge(timestamp?: string | null): string | null {
  if (!timestamp) return null;
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) return null;
  // Server clocks can run slightly ahead; a future timestamp reads as "just
  // now" rather than a nonsense token from the "in 5 minutes" phrasing.
  if (date.getTime() - Date.now() > -60_000) return "just now";
  const [value, unit] = formatDistanceToNowStrict(date, { roundingMethod: "floor" }).split(" ");
  return `${value}${AGE_UNIT_ABBREV[unit.replace(/s$/, "")]} ago`;
}

/** Absolute timestamp for the age tooltip. */
export function absoluteAge(timestamp?: string | null): string | null {
  if (!timestamp) return null;
  return new Date(timestamp).toLocaleString();
}
